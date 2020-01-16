package api

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgserror "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/forceTool"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/parameterSelector"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"net/http"
	"strings"
	"time"
)

func initGame(request *http.Request) (store.PlayerStore, engine.EngineConfig, engine.Gamestate, rgserror.IRGSError) {
	// get refresh token from auth header if applicable
	gameSlug := chi.URLParam(request, "gameSlug")
	logger.Debugf("request: %#v", request)
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
	latestGamestate, player, err := store.InitPlayerGS(authToken, authToken, gameSlug, "maverick", currency, wallet)
	return player, engineConfig, latestGamestate, err

}

func renderNextGamestate(request *http.Request) (GameplayResponse, rgserror.IRGSError) {
	gamestate, player, balance, engineConf, err := play(request)
	if err != nil {
		return GameplayResponse{}, err
	}
	logger.Warnf("balance: %#v", balance)
	return renderGamestate(request, gamestate, balance, engineConf, player), nil
}

// Play function for engines

func play(request *http.Request) (engine.Gamestate, store.PlayerStore, BalanceResponse, engine.EngineConfig, rgserror.IRGSError) {
	authHeader := request.Header.Get("Authorization")
	gameSlug := chi.URLParam(request, "gameSlug")
	wallet := chi.URLParam(request, "wallet")
	memID := strings.Split(authHeader, "\"")[1]
	logger.Debugf("request: %v", request)
	var player store.PlayerStore
	var previousGamestateStore store.GameStateStore
	var txStore store.TransactionStore
	var err *store.Error
	switch wallet {
	case "demo":
		player, previousGamestateStore, err = store.ServLocal.PlayerByToken(store.Token(memID), store.ModeDemo, gameSlug)
		txStore, err = store.ServLocal.TransactionByGameId(store.Token(memID), store.ModeDemo, gameSlug)
	case "dashur":
		player, previousGamestateStore, err = store.Serv.PlayerByToken(store.Token(memID), store.ModeReal, gameSlug)
		txStore, err = store.Serv.TransactionByGameId(store.Token(memID), store.ModeReal, gameSlug)
	}
	var previousGamestate engine.Gamestate

	if err != nil {
		// no player with that token
		return previousGamestate, player, BalanceResponse{}, engine.EngineConfig{}, rgserror.ErrInvalidCredentials
	}
	if len(previousGamestateStore.GameState) == 0 {
		logger.Warnf("No previous gameplay")
		// this should never happen as on first round init a sham gamestate is stored
		return previousGamestate, player, BalanceResponse{}, engine.EngineConfig{}, rgserror.ErrInvalidCredentials
	}

	previousGamestate = store.DeserializeGamestateFromBytes(previousGamestateStore.GameState)

	// check that previous gamestate matches what the client expects
	clientID := chi.URLParam(request, "gamestateID")
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
	// bugfix for skyjewels
	if gameSlug == "sky-jewels" || gameSlug == "goal" || gameSlug == "cookoff-champion" && len(data.SelectedWinLines) == 49 {
		swl := make([]int, 50)
		for i := 0; i < 50; i++ {
			swl[i] = i
			data.SelectedWinLines = swl
		}
	}

	if data.Action != "base" {
		// stake value must be zero
		logger.Debugf("setting zero stake value for %v round", data.Action)
		data.Stake = 0
	} else {

		// check that previous TX was endround
		if txStore.RoundStatus != store.RoundStatusClose {
			return engine.Gamestate{}, store.PlayerStore{}, BalanceResponse{}, engine.EngineConfig{}, rgserror.ErrSpinSequence
		}

		stakeValues, _, err := parameterSelector.GetGameplayParameters(0, player, gameSlug)
		if err != nil {
			logger.Warnf("Error: %v", err)
			return engine.Gamestate{}, store.PlayerStore{}, BalanceResponse{}, engine.EngineConfig{}, rgserror.ErrInvalidStake
		}
		valid := false
		for i := 0; i < len(stakeValues); i++ {
			if data.Stake == stakeValues[i] {
				valid = true
				if i == len(stakeValues)-1 && data.Action == "base" {
					// pass on when max bet is played, only if no action is passed already
					data.Action = "maxBase"
				}
				break
			}
		}
		if valid == false {
			logger.Warnf("invalid stake: %v (options: %v)", data.Stake, stakeValues)
			return engine.Gamestate{}, store.PlayerStore{}, BalanceResponse{}, engine.EngineConfig{}, rgserror.ErrInvalidStake
		}
	}

	gamestate, engineConf := engine.Play(previousGamestate, data.Stake, player.Balance.Currency, data)
	if config.GlobalConfig.DevMode == true {
		forcedGamestate, err := forceTool.GetForceValues(previousGamestate, gameSlug, player.PlayerId)
		if err == nil {
			logger.Warnf("Forcing gamestate: %v", forcedGamestate)
			gamestate = forcedGamestate
		} else {
			//assume error is of memcache.ErrCacheMiss variety
			logger.Debugf("No force value found for player %v", player.PlayerId)
		}
	}
	gamestate.PreviousGamestate = previousGamestate.Id

	// settle transactions
	var balance store.BalanceStore
	token := player.Token
	for _, transaction := range gamestate.Transactions {
		gs := store.SerializeGamestateToBytes(gamestate)
		status := store.RoundStatusOpen
		txStore := store.TransactionStore{
			TransactionId:       transaction.Id,
			Token:               token,
			Category:            store.Category(transaction.Type),
			RoundStatus:         status,
			PlayerId:            player.PlayerId,
			GameId:              gameSlug,
			RoundId:             gamestate.Id,
			Amount:              transaction.Amount,
			ParentTransactionId: "",
			TxTime:              time.Now(),
			GameState:           gs,
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
			return engine.Gamestate{}, store.PlayerStore{}, BalanceResponse{}, engine.EngineConfig{}, rgserror.ErrDasInsufficientFundError
		}
		token = balance.Token
	}
	player.Token = token
	player.Balance = balance.Balance
	balanceResponse := BalanceResponse{
		Amount:   balance.Balance.Amount.ValueAsString(),
		Currency: balance.Balance.Currency,
	}
	logger.Debugf("end of PLAY, balance: %v", player.Balance)

	return gamestate, player, balanceResponse, engineConf, nil
}
