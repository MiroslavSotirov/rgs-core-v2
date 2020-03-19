package api

import (
	"encoding/json"
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
	"strconv"
	"strings"
	"time"
)

func initGame(request *http.Request) (store.PlayerStore, engine.EngineConfig, engine.Gamestate, rgserror.IRGSError) {
	// get refresh token from auth header if applicable
	gameSlug := chi.URLParam(request, "gameSlug")
	currency := request.FormValue("currency")
	engineID, err := config.GetEngineFromGame(gameSlug)
	if err != nil {
		logger.Errorf("InitGame Error EngineID: %s - %s", gameSlug+"-engine", err)
		return store.PlayerStore{}, engine.EngineConfig{}, engine.Gamestate{}, rgserror.ErrEngineNotFound
	}
	engineConfig := engine.BuildEngineDefs(engineID)
	authToken, err := processAuthorization(request)
	if err != nil {
		return store.PlayerStore{}, engine.EngineConfig{}, engine.Gamestate{}, err
	}
	wallet := chi.URLParam(request, "wallet")
	latestGamestate, player, err := store.InitPlayerGS(authToken, authToken, gameSlug, currency, wallet)
	return player, engineConfig, latestGamestate, err

}

func renderNextGamestate(request *http.Request) (GameplayResponse, rgserror.IRGSError) {
	gamestate, player, balance, engineConf, err := play(request)
	if err != nil {
		return GameplayResponse{}, err
	}
	return renderGamestate(request, gamestate, balance, engineConf, player), nil
}

func validateParams(data engine.GameParams) engine.GameParams {
	// VALIDATE PARAMETERS
	// legacy for old client
	if data.Action == "spin" || data.Action == "" {
		data.Action = "base"
	} else if data.Action == "feature_select" {
		data.Action = "pickSpins"
		switch data.Selection {
		case "freeSpins5":
			data.Selection = "freespin5:5"
		case "freeSpins10":
			data.Selection = "freespin10:10"
		case "freeSpins25":
			data.Selection = "freespin25:25"
		}

	}
	return data
}
// Play function for engines
func getInitPlayValues(request *http.Request, clientID string, memID string, gameSlug string) (txStore store.TransactionStore, previousGamestate engine.Gamestate) {
	// we can expect error here on first gameplay, if error is entity not found then we can assume this is the first round
	logger.Warnf("First gameplay for this player")
	// because txstore is nil, we need to be smart about choosing currency
	ccy := request.FormValue("ccy")
	playerID := request.FormValue("playerId")

	previousGamestate = store.CreateInitGS(store.PlayerStore{PlayerId:playerID, Balance:engine.Money{0,ccy}}, gameSlug)
	previousGamestate.Id = clientID
	txStore.RoundStatus = store.RoundStatusClose
	txStore.Token = store.Token(memID)
	txStore.Amount.Currency = ccy
	txStore.PlayerId = playerID
	campaign := request.FormValue("campaign")
	if campaign != "" {
		numFG, err := strconv.Atoi(request.FormValue("numFG"))
		if err != nil {
			logger.Warnf("error decoding free game info: campaign - %v, err -- %v", campaign, err)
		} else {
			txStore.FreeGames.NoOfFreeSpins = numFG
			txStore.FreeGames.CampaignRef = campaign
		}
	}
	txStore.BetLimitSettingCode = request.FormValue("betLimitCode")
	txStore.WalletStatus = 1
	return
}

func validateBet(data engine.GameParams, txStore store.TransactionStore, previousGamestate engine.Gamestate, gameSlug string) (bool, engine.GameParams, *rgserror.RGSError) {
	minBet := false
	if data.Action != "base" {
		// stake value must be zero
		// todo: handle respin
		logger.Debugf("setting zero stake value for %v round", data.Action)
		data.Stake = 0
	} else {
		// check that previous TX was endround
		if txStore.RoundStatus != store.RoundStatusClose {
			logger.Warnf("last TX: %#v", txStore)
			return false, data, rgserror.ErrSpinSequence
		}

		stakeValues, _, err := parameterSelector.GetGameplayParameters(previousGamestate.BetPerLine, txStore.BetLimitSettingCode, gameSlug)
		if err != nil {
			logger.Warnf("Error: %v", err)
			return false, data, rgserror.ErrInvalidStake
		}
		valid := false
		for i := 0; i < len(stakeValues); i++ {
			if data.Stake == stakeValues[i] {
				valid = true
				if i == len(stakeValues)-1 && data.Action == "base" {
					// pass on when max bet is played, only if no action is passed already
					data.Action = "maxBase"
				}
				if i == 0 {
					minBet = true
				}
				break
			}
		}
		if valid == false {
			logger.Warnf("invalid stake: %v (options: %v)", data.Stake, stakeValues)
			return false, data, rgserror.ErrInvalidStake
		}
	}
	return minBet, data, nil
}

func play(request *http.Request) (engine.Gamestate, store.PlayerStore, BalanceResponse, engine.EngineConfig, rgserror.IRGSError) {
	authHeader := request.Header.Get("Authorization")
	gameSlug := chi.URLParam(request, "gameSlug")
	wallet := chi.URLParam(request, "wallet")
	memID := strings.Split(authHeader, "\"")[1]
	clientID := chi.URLParam(request, "gamestateID")

	engineID, engineErr := config.GetEngineFromGame(gameSlug)
	if engineErr != nil {
		logger.Errorf("Play Error: %s - %s", gameSlug+"-engine", engineErr)
		return engine.Gamestate{}, store.PlayerStore{}, BalanceResponse{}, engine.EngineConfig{}, rgserror.ErrEngineNotFound
	}

	var txStore store.TransactionStore
	var err *store.Error
	var previousGamestate engine.Gamestate
	switch wallet {
	case "demo":
		txStore, err = store.ServLocal.TransactionByGameId(store.Token(memID), store.ModeDemo, gameSlug)
	case "dashur":
		txStore, err = store.Serv.TransactionByGameId(store.Token(memID), store.ModeReal, gameSlug)
	}
	if err != nil {
		if err.Code == store.ErrorCodeEntityNotFound {
			// this is first gameplay
			txStore, previousGamestate = getInitPlayValues(request, clientID, memID, gameSlug)
		} else {
			//otherwise this is an actual error
			return previousGamestate, store.PlayerStore{}, BalanceResponse{}, engine.EngineConfig{}, rgserror.ErrInvalidCredentials
		}
	} else {
		previousGamestate = store.DeserializeGamestateFromBytes(txStore.GameState)
		// there is a rare case where a player launched a game before we had the proper handling for init cases, here we can detect this by checking if the last tx was an endround with an incomplete gamestate
		if previousGamestate.PreviousGamestate == "" {
			logger.Warnf("Solving Previous Gamestate Issue")
			txStore, previousGamestate = getInitPlayValues(request, clientID, memID, gameSlug)
		}
	}
	// check that previous gamestate matches what the client expects
	logger.Debugf("Previous id: %v, requested id: %v", previousGamestate.Id, clientID)
	if clientID != previousGamestate.Id {
		return engine.Gamestate{}, store.PlayerStore{}, BalanceResponse{}, engine.EngineConfig{}, rgserror.ErrSpinSequence
	}
	logger.Debugf("Previous Gamestate: %v", previousGamestate)
	// get parameters from post form (perhaps this should be handled POST func)
	decoder := json.NewDecoder(request.Body)
	var data engine.GameParams
	decodeerr := decoder.Decode(&data)
	if decodeerr != nil {
		logger.Errorf("Unable to decode request body: %s", decodeerr.Error())
		return engine.Gamestate{}, store.PlayerStore{}, BalanceResponse{}, engine.EngineConfig{}, rgserror.ErrGamestateStore
	}
	data = validateParams(data)

	// bugfix for engine xiii (this should really be fixed in the client)
	// until the clients are fixed, this must stay
	if engineID == "mvgEngineXIII"  && len(data.SelectedWinLines) == 49 {
		swl := make([]int, 50)
		for i := 0; i < 50; i++ {
			swl[i] = i
			data.SelectedWinLines = swl
		}
	}
	minBet, data, betValidationErr := validateBet(data, txStore, previousGamestate, gameSlug)
	if betValidationErr != nil {
		return engine.Gamestate{}, store.PlayerStore{}, BalanceResponse{}, engine.EngineConfig{}, betValidationErr
	}

	// add suffix to gamestate in case this is a retry attempt
	switch txStore.WalletStatus {
	case 0:
		// this tx is pending in wallet, quit and force reload
		return engine.Gamestate{}, store.PlayerStore{}, BalanceResponse{}, engine.EngineConfig{}, rgserror.ErrPreviousTXPending
	case -1:
		// the next tx failed, retrying it will cause a duplicate tx id error, so add a suffix
		previousGamestate.NextGamestate = previousGamestate.NextGamestate + rng.RandStringRunes(4)
		logger.Debugf("adding suffix to next tx to avoid duplication error, resulting id: %v", previousGamestate.NextGamestate)
	case 1:
		// business as usual
	default:
		// it should always be one of the above three
		logger.Infof("Wallet Status is unexpectedly %v", txStore.WalletStatus)
		return engine.Gamestate{}, store.PlayerStore{}, BalanceResponse{}, engine.EngineConfig{}, rgserror.ErrGenericWalletErr
	}
	gamestate, engineConf := engine.Play(previousGamestate, data.Stake, previousGamestate.BetPerLine.Currency, data)
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
		status := store.RoundStatusOpen
		txStore := store.TransactionStore{
			TransactionId:       transaction.Id,
			Token:               token,
			Category:            store.Category(transaction.Type),
			RoundStatus:         status,
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
				return engine.Gamestate{}, store.PlayerStore{}, BalanceResponse{}, engine.EngineConfig{}, rgserror.ErrInsufficientFundError
			}
			logger.Infof("error: %v", err)
			return engine.Gamestate{}, store.PlayerStore{}, BalanceResponse{}, engine.EngineConfig{}, rgserror.ErrGenericWalletErr
		}
		token = balance.Token
	}
	player := store.PlayerStore{Token: token, PlayerId: txStore.PlayerId}

	balanceResponse := BalanceResponse{
		Amount:   balance.Balance.Amount,
		Currency: balance.Balance.Currency,
		FreeGames: balance.FreeGames.NoOfFreeSpins,
	}

	return gamestate, player, balanceResponse, engineConf, nil
}
