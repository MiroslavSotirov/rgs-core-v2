package api

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgserror "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
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

func initV2(request *http.Request) (GameInitResponseV2, rgserror.IRGSError) {
	//
	gameSlug := request.FormValue("game")
	operator := request.FormValue("operator") // required
	mode := request.FormValue("mode") // required
	params := request.URL.Query()
	currency := params.Get("currency")
	logger.Debugf("Game: %v; operator: %v; mode: %v; request: %#v", gameSlug, operator, mode, request)

	engineID, err := config.GetEngineFromGame(gameSlug)
	if err != nil {
		logger.Errorf("InitGame Error EngineID: %s - %s", gameSlug+"-engine", err)
		return GameInitResponseV2{}, rgserror.ErrEngineNotFound
	}
	engineConfig := engine.BuildEngineDefs(engineID)
	authToken := strings.Split(request.Header.Get("Authorization"), " ")[1] // assume format MAVERICK-Host-Token aaa-1234-aaa is passed in Auth header

	// get wallet from operator config
	wallet, err := config.GetWalletFromOperatorAndMode(operator, mode)
	if err != nil {
		return GameInitResponseV2{}, err
	}
	var player store.PlayerStore
	var latestGamestate engine.Gamestate

	latestGamestate, player, err = store.InitPlayerGS(authToken, authToken, gameSlug, currency, wallet)

	if err != nil {
		return GameInitResponseV2{}, err
	}
	giResp := fillGameInitPreviousGameplay(latestGamestate, store.BalanceStore{Balance: player.Balance, Token: player.Token})
	giResp.FillEngineInfo(engineConfig)
	logger.Debugf("reel response: %v", giResp.ReelSets)

	// set stakevalues, links,
	links := make(map[string]string, 1)
	newGameHref := fmt.Sprintf("%s%s/%s/play2/%s?wallet=%v", GetURLScheme(request), request.Host, APIVersion, gameSlug, wallet)
	// handle initial gamestate
	if len(latestGamestate.Transactions) == 0 {
		newGameHref += fmt.Sprintf("&playerId=%v&ccy=%v&betLimitCode=%v&campaign=%v", player.PlayerId, player.Balance.Currency, player.BetLimitSettingCode, player.FreeGames.CampaignRef)
		//if player.FreeGames.NoOfFreeSpins > 0 {
		//	newGameHref += fmt.Sprintf("c", player.FreeGames.CampaignRef, player.FreeGames.NoOfFreeSpins)
		//}
		logger.Debugf("Rendering sham init gamestate: %v", latestGamestate.Id)
	}
	logger.Warnf("link new game: %v", newGameHref)
	links["new-game"] = newGameHref
	giResp.Links = links
	stakeValues, defaultBet, err := parameterSelector.GetGameplayParameters(latestGamestate.BetPerLine, player.BetLimitSettingCode, gameSlug)
	if err != nil {
		return GameInitResponseV2{}, err
	}
	giResp.StakeValues = stakeValues
	giResp.DefaultBet = defaultBet
	return giResp, nil
}

func playV2(request *http.Request) (GameplayResponseV2, rgserror.IRGSError) {
	clientID := chi.URLParam(request, "lastID")
	if clientID == "" {
		return playFirst(request)
	}
	authHeader := request.Header.Get("Authorization")
	// get parameters from post form

	decoder := json.NewDecoder(request.Body)
	var data engine.GameParams
	decoderror := decoder.Decode(&data)
	logger.Debugf("request body: %#v", request.Body)
	if decoderror != nil {
		logger.Errorf("Unable to decode request body: %s", decoderror.Error())
		return GameplayResponseV2{}, rgserror.ErrGamestateStore
	}

	memID := strings.Trim(strings.Split(authHeader, " ")[1], "\"")

	var txStore store.TransactionStore
	var previousGamestate engine.Gamestate
	var err *store.Error
	switch data.Wallet {
	case "demo":
		txStore, err = store.ServLocal.TransactionByGameId(store.Token(memID), store.ModeDemo, data.Game)
	case "dashur":
		txStore, err = store.Serv.TransactionByGameId(store.Token(memID), store.ModeReal, data.Game)
	default:
		logger.Errorf("No such wallet '%v': %#v", data.Wallet, request)
	}
	logger.Infof("txstore: %v, err: %v", txStore, err)
	if err != nil {
		// there would be an error on the first gameplay, but first gameplay should be routed to playFirst
		// todo: get rgserr from store
		return GameplayResponseV2{}, rgserror.ErrGenericWalletErr
	}
	previousGamestate = store.DeserializeGamestateFromBytes(txStore.GameState)
	// check if the gsID passed in by the client matches that retrieved from the wallet
	if clientID != previousGamestate.Id {
		return GameplayResponseV2{}, rgserror.ErrSpinSequence
	}

	// add suffix to gamestate in case this is a retry attempt
	switch txStore.WalletStatus {
	case 0:
		// this tx is pending in wallet, quit and force reload
		return GameplayResponseV2{}, rgserror.ErrPreviousTXPending
	case -1:
		// the next tx failed, retrying it will cause a duplicate tx id error, so add a suffix
		previousGamestate.NextGamestate = previousGamestate.NextGamestate + rng.RandStringRunes(4)
		logger.Debugf("adding suffix to next tx to avoid duplication error, resulting id: %v", previousGamestate.NextGamestate)
	case 1:
		// business as usual
	default:
		// it should always be one of the above three
		return GameplayResponseV2{}, rgserror.ErrGenericWalletErr
	}

 	return getRoundResults(data, previousGamestate, txStore)
}

func playFirst(request *http.Request) (GameplayResponseV2, rgserror.IRGSError) {
	authHeader := request.Header.Get("Authorization")

	// get parameters from post form
	decoder := json.NewDecoder(request.Body)
	var data engine.GameParams
	decoderror := decoder.Decode(&data)

	if decoderror != nil {
		logger.Debugf("Unable to decode request body: %#v, %v", request.Body, decoderror.Error())
		return GameplayResponseV2{}, rgserror.ErrBadConfig
	}
	errV := data.Validate()
	if errV != nil {
		return GameplayResponseV2{}, errV
	}

	memID := strings.Trim(strings.Split(authHeader, " ")[1], "\"")
	var player store.PlayerStore
	var latestGamestateStore store.GameStateStore
	var err *store.Error
	var initGS engine.Gamestate

	switch data.Wallet {
	case "dashur":
		player, latestGamestateStore, err = store.Serv.PlayerByToken(store.Token(memID), store.ModeReal, data.Game)
	case "demo":
		player, latestGamestateStore, err = store.ServLocal.PlayerByToken(store.Token(memID), store.ModeDemo, data.Game)
	default:
		return GameplayResponseV2{}, rgserror.ErrBadConfig
	}
	if err != nil {
		// todo: get exact rgserr from store
		return GameplayResponseV2{}, rgserror.ErrGenericWalletErr
	}

	// because this is the first round, there should be no previous gamestate
	if len(latestGamestateStore.GameState) != 0 {
		return GameplayResponseV2{}, rgserror.ErrSpinSequence
	}

	// this is first gameplay
	initGS = store.CreateInitGS(player, data.Game)
	txStoreInit := store.TransactionStore{
		RoundStatus:         store.RoundStatusClose,
		BetLimitSettingCode: player.BetLimitSettingCode,
		PlayerId: player.PlayerId,
		FreeGames: player.FreeGames,
		Token: player.Token,
	}

	// don't need to worry about the wallet status as the GS ID is randomly generated on reload

	return getRoundResults(data, initGS, txStoreInit)
}


func getRoundResults(data engine.GameParams, previousGamestate engine.Gamestate, txStore store.TransactionStore) (GameplayResponseV2, rgserror.IRGSError) {


	minBet, data, betValidationErr := validateBet(data, txStore, data.Game)
	if betValidationErr != nil {
		return GameplayResponseV2{}, betValidationErr
	}

	gamestate, _ := engine.Play(previousGamestate, data.Stake, previousGamestate.BetPerLine.Currency, data)
	if config.GlobalConfig.DevMode == true {
		forcedGamestate, err := forceTool.GetForceValues(data.Stake, previousGamestate, data.Game, txStore.PlayerId)
		if err == nil {
			logger.Warnf("Forcing gamestate: %v", forcedGamestate)
			gamestate = forcedGamestate
		} else {
			//assume error is of memcache.ErrCacheMiss variety
			logger.Warnf("No force value found for player %v", txStore.PlayerId)
		}
	}

	var freeGameRef string
	if txStore.FreeGames.NoOfFreeSpins > 0 && minBet == true {
		// this game qualifies as a free game!
		freeGameRef = txStore.FreeGames.CampaignRef
		logger.Warnf("Free game campaign %v", freeGameRef)
	}
	var storeErr *store.Error

	// settle transactions
	var balance store.BalanceStore
	token := txStore.Token
	for _, transaction := range gamestate.Transactions {
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
			FreeGames: 			 store.FreeGamesStore{NoOfFreeSpins:0, CampaignRef:freeGameRef},
		}
		switch data.Wallet {
		case "demo":
			tx.Mode = store.ModeDemo
			balance, storeErr = store.ServLocal.Transaction(token, store.ModeDemo, tx)
		case "dashur":
			tx.Mode = store.ModeReal
			balance, storeErr = store.Serv.Transaction(token, store.ModeReal, tx)
		}

		if storeErr != nil {
			if storeErr.Code == store.ErrorCodeNotEnoughBalance {
				return GameplayResponseV2{}, rgserror.ErrInsufficientFundError
			}
			return GameplayResponseV2{}, rgserror.ErrGenericWalletErr
		}
		token = balance.Token
	}
	return fillGamestateResponseV2(gamestate, balance), nil
}

