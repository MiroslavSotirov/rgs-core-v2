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
	operator := request.FormValue("operator")
	mode := request.FormValue("mode")
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
	giResp := fillGameInitPreviousGameplay(latestGamestate, store.BalanceStore{Balance: player.Balance, Token: player.Token}, gameSlug)
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
	authHeader := request.Header.Get("Authorization")
	gameSlug := chi.URLParam(request, "gameSlug")
	wallet := request.FormValue("wallet")
	clientID := request.FormValue("lastGS")
	logger.Debugf("Auth string: %v", authHeader)
	memID := strings.Trim(strings.Split(authHeader, " ")[1], "\"")
	logger.Debugf("memID: %v", memID)
	var txStore store.TransactionStore
	var previousGamestate engine.Gamestate
	var err *store.Error
	switch wallet {
	case "demo":
		txStore, err = store.ServLocal.TransactionByGameId(store.Token(memID), store.ModeDemo, gameSlug)
	case "dashur":
		txStore, err = store.Serv.TransactionByGameId(store.Token(memID), store.ModeReal, gameSlug)
	default:
		logger.Errorf("No such wallet '%v': %#v", wallet, request)
	}
	logger.Infof("txstore: %v, err: %v", txStore, err)
	if err != nil {
		// if there is an error retrieving last tx, it may be that this is the first gameplay on this game for this player. if so ,handle appropriately
		if err.Code == store.ErrorCodeEntityNotFound {
			// this is first gameplay
			txStore, previousGamestate = getInitPlayValues(request, clientID, memID, gameSlug)
		} else {
			//this is a real error
			// todo: get rgserr from store
			return GameplayResponseV2{}, rgserror.ErrGenericWalletErr
		}
	} else {
		previousGamestate = store.DeserializeGamestateFromBytes(txStore.GameState)
	}
	// check if the gsID passed in by the client matches that retrieved from the wallet
	if clientID != previousGamestate.Id {
		return GameplayResponseV2{}, rgserror.ErrSpinSequence
	}


	// get parameters from post form
	decoder := json.NewDecoder(request.Body)
	var data engine.GameParams
	decoderror := decoder.Decode(&data)
	logger.Debugf("request body: %#v", request.Body)
	if decoderror != nil {
		logger.Errorf("Unable to decode request body: %s", decoderror.Error())
		return GameplayResponseV2{}, rgserror.ErrGamestateStore
	}

	if data.Action == "" {
		data.Action = "base"
	}
	minBet, data, betValidationErr := validateBet(data, txStore, previousGamestate, gameSlug)
	if betValidationErr != nil {
		return GameplayResponseV2{}, betValidationErr
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
	gamestate, _ := engine.Play(previousGamestate, data.Stake, previousGamestate.BetPerLine.Currency, data)
	if config.GlobalConfig.DevMode == true {
		forcedGamestate, err := forceTool.GetForceValues(data.Stake, previousGamestate, gameSlug, txStore.PlayerId)
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

	// settle transactions
	var balance store.BalanceStore
	token := txStore.Token
	for _, transaction := range gamestate.Transactions {
		gs := store.SerializeGamestateToBytes(gamestate)
		txStore := store.TransactionStore{
			TransactionId:       transaction.Id,
			Token:               token,
			Category:            store.Category(transaction.Type),
			RoundStatus:         store.RoundStatusOpen,
			PlayerId:            txStore.PlayerId,
			GameId:              gameSlug,
			RoundId:             gamestate.RoundID,
			Amount:              transaction.Amount,
			ParentTransactionId: "",
			TxTime:              time.Now(),
			GameState:           gs,
			BetLimitSettingCode: txStore.BetLimitSettingCode,
			FreeGames: 			 store.FreeGamesStore{NoOfFreeSpins:0, CampaignRef:freeGameRef},
		}
		switch wallet {
		case "demo":
			txStore.Mode = store.ModeDemo
			balance, err = store.ServLocal.Transaction(token, store.ModeDemo, txStore)
		case "dashur":
			txStore.Mode = store.ModeReal
			balance, err = store.Serv.Transaction(token, store.ModeReal, txStore)
		}

		if err != nil {
			if err.Code == store.ErrorCodeNotEnoughBalance {
				return GameplayResponseV2{}, rgserror.ErrInsufficientFundError
			}
			return GameplayResponseV2{}, rgserror.ErrGenericWalletErr
		}
		token = balance.Token
	}
	//player := store.PlayerStore{Token: token, PlayerId: txStore.PlayerId}
	// todo: need to pass token into response
	//balanceResponse := BalanceResponse{
	//	Amount:   balance.Balance.Amount.ValueAsString(),
	//	Currency: balance.Balance.Currency,
	//	FreeGames: balance.FreeGames.NoOfFreeSpins,
	//}

	return fillGamestateResponseV2(gamestate, balance), nil
}
