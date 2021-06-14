package api

import (
	"encoding/json"
	"fmt"
	"github.com/getsentry/sentry-go"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/forceTool"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/parameterSelector"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"net/http"
	"strings"
	"time"
)

func getGameLink(request *http.Request) GameLinkResponse {
	params := request.URL.Query()
	game := params.Get("game")
	wallet := params.Get("interface")
	currency := params.Get("ccy")
	link := fmt.Sprintf("%s%s/%s/rgs/init/%s/%s?currency=%s", GetURLScheme(request), request.Host, APIVersion, game, wallet, currency)
	resultJSON := []LinkResponse{{ID: link}}
	response := GameLinkResponse{Results: resultJSON}

	return response
}

type initParams struct {
	Game     string `json:"game"`
	Operator string `json:"operator"`
	Mode     string `json:"mode"`
	Ccy      string `json:"currency"`
}

func (i *initParams) decode(request *http.Request) rgse.RGSErr {
	decoder := json.NewDecoder(request.Body)
	decoderror := decoder.Decode(i)

	if decoderror != nil {
		return rgse.Create(rgse.JsonError)
	}
	return nil
}

func initV2(request *http.Request) (GameInitResponseV2, rgse.RGSErr) {
	var data initParams
	if err := data.decode(request); err != nil {
		return GameInitResponseV2{}, err
	}

	logger.Debugf("Game: %v; operator: %v; mode: %v; request: %#v", data.Game, data.Operator, data.Mode, request)

	engineID, err := config.GetEngineFromGame(data.Game)
	if err != nil {
		return GameInitResponseV2{}, err
	}
	engineConfig := engine.BuildEngineDefs(engineID)
	authToken := strings.Split(request.Header.Get("Authorization"), " ")[1] // assume format MAVERICK-Host-Token aaa-1234-aaa is passed in Auth header

	// get wallet from operator config
	wallet, err := config.GetWalletFromOperatorAndMode(data.Operator, data.Mode)
	if err != nil {
		return GameInitResponseV2{}, err
	}
	var player store.PlayerStore
	var latestGamestate engine.Gamestate

	latestGamestate, player, err = store.InitPlayerGS(authToken, authToken, data.Game, data.Ccy, wallet)

	if err != nil {
		return GameInitResponseV2{}, err
	}

	giResp := fillGameInitPreviousGameplay(latestGamestate, store.BalanceStore{Balance: player.Balance, Token: player.Token, FreeGames: player.FreeGames})
	giResp.FillEngineInfo(engineConfig)
	//logger.Debugf("reel response: %v", giResp.ReelSets)
	giResp.Wallet = wallet
	// set stakevalues, links,

	stakeValues, defaultBet, err := parameterSelector.GetGameplayParameters(latestGamestate.BetPerLine, player.BetLimitSettingCode, data.Game)
	if err != nil {
		return GameInitResponseV2{}, err
	}
	giResp.StakeValues = stakeValues
	for i:=0; i<len(stakeValues); i++ {
		if player.FreeGames.NoOfFreeSpins > 0 && stakeValues[i].Mul(engine.NewFixedFromInt(engineConfig.EngineDefs[0].StakeDivisor)) == player.FreeGames.TotalWagerAmt {
			defaultBet = stakeValues[i]
			logger.Debugf("setting defaultbet to %v for freegames", defaultBet)
			break
		}
	}

	giResp.DefaultBet = defaultBet
	return giResp, nil
}

func playV2(request *http.Request) (GameplayResponseV2, rgse.RGSErr) {
	var data engine.GameParams
	if err := data.Decode(request); err != nil {
		return GameplayResponseV2{}, err
	}

	if strings.Contains(data.PreviousID, "GSinit") {
		return playFirst(request, data)
	}
	token, autherr := handleAuth(request)
	if autherr != nil {
		return GameplayResponseV2{}, autherr
	}

	var txStore store.TransactionStore
	var previousGamestate engine.Gamestate
	var err rgse.RGSErr
	switch data.Wallet {
	case "demo":
		txStore, err = store.ServLocal.TransactionByGameId(token, store.ModeDemo, data.Game)
	case "dashur":
		txStore, err = store.Serv.TransactionByGameId(token, store.ModeReal, data.Game)
	default:
		logger.Debugf("No such wallet '%v': %#v", data.Wallet, request)
		return GameplayResponseV2{}, rgse.Create(rgse.InvalidWallet)
	}

	logger.Debugf("txstore: %#v, err: %v", txStore, err)
	if err != nil {
		return GameplayResponseV2{}, err
	}
	previousGamestate = store.DeserializeGamestateFromBytes(txStore.GameState)
	// check if the gsID passed in by the client matches that retrieved from the wallet
	if data.PreviousID != previousGamestate.Id {
		logger.Debugf("Previous ID doesn't match: %v, %v", data.PreviousID, previousGamestate.Id)
		return GameplayResponseV2{}, rgse.Create(rgse.SpinSequenceError)
	}

	// add suffix to gamestate in case this is a retry attempt
	switch txStore.WalletStatus {
	case 0:
		// this tx is pending in wallet, quit and force reload
		return GameplayResponseV2{}, rgse.Create(rgse.PeviousTXPendingError)
	case -1:
		// the next tx failed, retrying it will cause a duplicate tx id error, so add a suffix
		previousGamestate.NextGamestate = previousGamestate.NextGamestate + rng.RandStringRunes(4)
		logger.Debugf("adding suffix to next tx to avoid duplication error, resulting id: %v", previousGamestate.NextGamestate)
	case 1:
		// business as usual
	default:
		// it should always be one of the above three
		logger.Debugf("Wallet status not 1, 0, or -1: %v", txStore)
		return GameplayResponseV2{}, rgse.Create(rgse.UnexpectedWalletStatus)
	}

	return getRoundResults(data, previousGamestate, txStore)
}

func handleAuth(r *http.Request) (token store.Token, err rgse.RGSErr) {
	authHeader := r.Header.Get("Authorization")
	tokenInfo := strings.Split(authHeader, " ")
	if tokenInfo[0] != "Maverick-Host-Token" || len(tokenInfo) < 2 {
		err = rgse.Create(rgse.InvalidCredentials)
		return
	}
	token = store.Token(tokenInfo[1])
	return
}
func playFirst(request *http.Request, data engine.GameParams) (GameplayResponseV2, rgse.RGSErr) {
	logger.Debugf("First gameplay for player")
	token, autherr := handleAuth(request)
	if autherr != nil {
		return GameplayResponseV2{}, autherr
	}

	if errV := data.Validate(); errV != nil {
		return GameplayResponseV2{}, errV
	}

	var player store.PlayerStore
	var latestGamestateStore store.GameStateStore
	var err rgse.RGSErr
	var initGS engine.Gamestate

	switch data.Wallet {
	case "dashur":
		player, latestGamestateStore, err = store.Serv.PlayerByToken(token, store.ModeReal, data.Game)
	case "demo":
		player, latestGamestateStore, err = store.ServLocal.PlayerByToken(token, store.ModeDemo, data.Game)
	default:
		return GameplayResponseV2{}, rgse.Create(rgse.InvalidWallet)
	}
	if err != nil {
		return GameplayResponseV2{}, err
	}

	// because this is the first round, there should be no previous gamestate
	if len(latestGamestateStore.GameState) != 0 {
		logger.Debugf("previous gamestate %v", latestGamestateStore)
		return GameplayResponseV2{}, rgse.Create(rgse.SpinSequenceError)
	}

	// this is first gameplay
	initGS = store.CreateInitGS(player, data.Game)
	txStoreInit := store.TransactionStore{
		RoundStatus:         store.RoundStatusClose,
		BetLimitSettingCode: player.BetLimitSettingCode,
		PlayerId:            player.PlayerId,
		FreeGames:           player.FreeGames,
		Token:               player.Token,
		Amount:              engine.Money{0, player.Balance.Currency},
		Ttl:                 3600,
	}

	// don't need to worry about the wallet status as the GS ID is randomly generated on reload

	return getRoundResults(data, initGS, txStoreInit)
}

func getRoundResults(data engine.GameParams, previousGamestate engine.Gamestate, txStore store.TransactionStore) (gameplay GameplayResponseV2, err rgse.RGSErr) {
	if txStore.Amount.Currency == "" {
		txStore.Amount.Currency = previousGamestate.BetPerLine.Currency
	}
	data, betValidationErr := validateBet(data, txStore, data.Game)
	if betValidationErr != nil {
		return GameplayResponseV2{}, betValidationErr
	}

	gamestate, EC := engine.Play(previousGamestate, data.Stake, previousGamestate.BetPerLine.Currency, data)
	if config.GlobalConfig.DevMode == true {
		forcedGamestate, err := forceTool.GetForceValues(data, previousGamestate, txStore.PlayerId)
		if err == nil {
			logger.Warnf("Forcing gamestate: %v", forcedGamestate)
			sentry.CaptureMessage("Forcing gamestate")
			gamestate = forcedGamestate
		} else {
			// continue play, assume no force was stored
			logger.Debugf("Error retrieving force for player %v: %v", txStore.PlayerId, err.Error())
		}
	}
	var freeGameRef string

	if txStore.FreeGames.NoOfFreeSpins > 0 && data.Stake.Mul(engine.NewFixedFromInt(EC.EngineDefs[0].StakeDivisor)) == txStore.FreeGames.TotalWagerAmt {
		// this game qualifies as a free game!
		freeGameRef = txStore.FreeGames.CampaignRef
		logger.Infof("Free game campaign %v", freeGameRef)
	} else if previousGamestate.RoundID == gamestate.RoundID && gamestate.Transactions[0].Type != "WAGER" {
		// if the game is a continuation of a round propogate the previous campaign ref to all txs linked to this round
		// except if there is a tx on this state that is a wager and it is not the first wager of  the round
		freeGameRef = txStore.FreeGames.CampaignRef
		logger.Infof("Campaign %v continues", freeGameRef)
	}

	// settle transactions
	var balance store.BalanceStore
	token := txStore.Token
	logger.Debugf("%v txs", len(gamestate.Transactions))
	for _, transaction := range gamestate.Transactions {
		logger.Debugf("%#v", transaction)
		gs := store.SerializeGamestateToBytes(gamestate)
		tx := store.TransactionStore{
			TransactionId:       transaction.Id,
			Token:               token,
			Category:            store.Category(transaction.Type),
			RoundStatus:         store.RoundStatusOpen,
			PlayerId:            txStore.PlayerId,
			GameId:              data.Game,
			RoundId:             gamestate.RoundID,
			Amount:              transaction.Amount,
			ParentTransactionId: "",
			TxTime:              time.Now(),
			GameState:           gs,
			BetLimitSettingCode: txStore.BetLimitSettingCode,
			FreeGames:           store.FreeGamesStore{NoOfFreeSpins: 0, CampaignRef: freeGameRef},
			Ttl:				 gamestate.GetTtl(),
		}
		switch data.Wallet {
		case "demo":
			tx.Mode = store.ModeDemo
			balance, err = store.ServLocal.Transaction(token, store.ModeDemo, tx)
		case "dashur":
			tx.Mode = store.ModeReal
			balance, err = store.Serv.Transaction(token, store.ModeReal, tx)
		default:
			err = rgse.Create(rgse.InvalidWallet)
			return
		}

		if err != nil {
			return
		}
		token = balance.Token
	}
	return fillGamestateResponseV2(gamestate, balance), nil
}

type CloseRoundParams struct {
	Game    string `json:"game"`
	Wallet  string `json:"wallet"`
	RoundID string `json:"round"`
}

func (i *CloseRoundParams) decode(request *http.Request) rgse.RGSErr {
	decoder := json.NewDecoder(request.Body)
	decoderror := decoder.Decode(i)

	if decoderror != nil {
		return rgse.Create(rgse.JsonError)
	}
	return nil
}

func CloseGS(r *http.Request) (err rgse.RGSErr) {
	token, autherr := handleAuth(r)
	if autherr != nil {
		return autherr
	}
	var data CloseRoundParams
	if err := data.decode(r); err != nil {
		return err
	}
	logger.Debugf("#v", data)
	logger.Debugf("#v", token)
	var txStore store.TransactionStore
	switch data.Wallet {
	case "demo":
		txStore, err = store.ServLocal.TransactionByGameId(token, store.ModeDemo, data.Game)
	case "dashur":
		txStore, err = store.Serv.TransactionByGameId(token, store.ModeReal, data.Game)
	default:
		return rgse.Create(rgse.InvalidWallet)
	}

	if err != nil {
		return
	}
	if txStore.WalletStatus != 1 {
		// if this is zero, the tx is pending and shouldn't be resent, if it is -1, the tx is failed and an error should be sent to reload the client
		logger.Debugf("INTERNAL STATUS: %v", txStore.WalletStatus)
		err = rgse.Create(rgse.PeviousTXPendingError)
		return
	}
	gamestateUnmarshalled := store.DeserializeGamestateFromBytes(txStore.GameState)
	if gamestateUnmarshalled.RoundID != data.RoundID {
		err = rgse.Create(rgse.SpinSequenceError)
		return
	}
	if len(gamestateUnmarshalled.NextActions) > 1 {
		// we should not be closing a gameround if the last gamestate has more actions to be completed
		err = rgse.Create(rgse.IncompleteRoundError)
		return
	}
	gamestateUnmarshalled.Closed = true
	roundId := gamestateUnmarshalled.RoundID
	if roundId == "" {
		roundId = gamestateUnmarshalled.Id
	}
	switch data.Wallet {
	case "demo":
		_, err = store.ServLocal.CloseRound(token, store.ModeDemo, data.Game, roundId, store.SerializeGamestateToBytes(gamestateUnmarshalled))
	case "dashur":
		_, err = store.Serv.CloseRound(token, store.ModeReal, data.Game, roundId, store.SerializeGamestateToBytes(gamestateUnmarshalled))
	}
	return
}
