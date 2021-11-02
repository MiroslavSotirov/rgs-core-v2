package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/forceTool"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/parameterSelector"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

func initGame(request *http.Request) (store.PlayerStore, engine.EngineConfig, engine.Gamestate, rgse.RGSErr) {
	// get refresh token from auth header if applicable
	gameSlug := chi.URLParam(request, "gameSlug")
	currency := request.FormValue("currency")
	engineID, err := config.GetEngineFromGame(gameSlug)
	if err != nil {
		logger.Errorf("InitGame Error EngineID: %s - %s", gameSlug+"-engine", err)
		return store.PlayerStore{}, engine.EngineConfig{}, engine.Gamestate{}, rgse.Create(rgse.EngineNotFoundError)
	}
	engineConfig := engine.BuildEngineDefs(engineID)
	authToken, err := processAuthorization(request)
	if err != nil {
		return store.PlayerStore{}, engine.EngineConfig{}, engine.Gamestate{}, err
	}
	wallet := chi.URLParam(request, "wallet")
	latestGamestate, player, err := store.InitPlayerGS(authToken, authToken, gameSlug, currency, wallet)
	if err != nil {
		return store.PlayerStore{}, engine.EngineConfig{}, engine.Gamestate{}, err
	}

	return player, engineConfig, latestGamestate, nil

}

func renderNextGamestate(request *http.Request) (GameplayResponse, rgse.RGSErr) {

	decoder := json.NewDecoder(request.Body)
	var data engine.GameParams
	decodeerr := decoder.Decode(&data)
	if decodeerr != nil {
		logger.Errorf("Unable to decode request body: %s", decodeerr.Error())
		return GameplayResponse{}, rgse.Create(rgse.JsonError)
	}
	gamestate, player, balance, engineConf, err := play(request, data)
	if err != nil {
		return GameplayResponse{}, err
	}
	if gamestate.Action == "pickSpins" {
		// hack for engine III
		request.Header.Set("Authorization", "\""+string(player.Token))
		logger.Debugf("engine3, calculating next round: %#v", request.Body)
		gamestate, player, balance, engineConf, err = play(request, data)
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
// todo: deprecate this once v1 api is no longer used
func getInitPlayValues(request *http.Request, clientID string, memID string, gameSlug string) (txStore store.TransactionStore, previousGamestate engine.Gamestate) {
	// we can expect error here on first gameplay, if error is entity not found then we can assume this is the first round
	logger.Warnf("First gameplay for this player")
	// because txstore is nil, we need to be smart about choosing currency
	ccy := request.FormValue("ccy")
	playerID := request.FormValue("playerId")

	previousGamestate = store.CreateInitGS(store.PlayerStore{PlayerId: playerID, Balance: engine.Money{0, ccy}}, gameSlug)
	previousGamestate.Id = clientID
	txStore.RoundStatus = store.RoundStatusClose
	txStore.Token = store.Token(memID)
	txStore.Amount.Currency = ccy
	txStore.PlayerId = playerID
	txStore.BetLimitSettingCode = request.FormValue("betLimitCode")
	txStore.WalletStatus = 1
	campaign := request.FormValue("campaign")
	if campaign != "" {
		wagerAmt, err := strconv.Atoi(request.FormValue("wagerAmt"))
		if err != nil {
			return
		}
		txStore.FreeGames.CampaignRef = campaign
		txStore.FreeGames.TotalWagerAmt = engine.Fixed(wagerAmt)
	}

	return
}

func validateBet(data engine.GameParams, txStore store.TransactionStore, game string) (engine.GameParams, rgse.RGSErr) {
	if data.Action != "base" {
		// stake value must be zero
		// check that the round is open
		if txStore.RoundStatus != store.RoundStatusOpen {
			logger.Warnf("last TX should be open: %#v", txStore)
			return data, rgse.Create(rgse.SpinSequenceError)
		}
		logger.Debugf("setting zero stake value for %v round", data.Action)
		if data.Action != "respin" && data.Action != "gamble" {
			data.Stake = 0
		}
	} else {
		// check that previous TX was endround
		if txStore.RoundStatus != store.RoundStatusClose {
			logger.Warnf("last TX: %#v", txStore)
			return data, rgse.Create(rgse.SpinSequenceError)
		}

		stakeValues, _, err := parameterSelector.GetGameplayParameters(engine.Money{0, txStore.Amount.Currency}, txStore.BetLimitSettingCode, game)
		if err != nil {
			return data, err
		}

		valid := false
		for i := 0; i < len(stakeValues); i++ {
			if data.Stake == stakeValues[i] {
				valid = true
				if i == len(stakeValues)-1 && data.Action == "base" {
					// pass on when max bet is played, only if no action is passed already and game allows it
					EC, err := engine.Gamestate{Game: game}.Engine()
					if err == nil {
						maxDef := EC.DefIdByName("maxBase")
						if maxDef >= 0 && len(data.SelectedWinLines) == len(EC.EngineDefs[maxDef].WinLines) {
							// if the selected winlines is also maximum and there is a special config for this
							data.Action = "maxBase"
							logger.Debugf("setting action to maxbase")
						}

					}
				}
				break
			}
		}
		if valid == false {
			logger.Debugf("invalid stake: %v (options: %v)", data.Stake, stakeValues)
			return data, rgse.Create(rgse.InvalidStakeError)
		}
	}
	return data, nil
}

func play(request *http.Request, data engine.GameParams) (engine.Gamestate, store.PlayerStore, BalanceResponse, engine.EngineConfig, rgse.RGSErr) {
	authHeader := request.Header.Get("Authorization")
	gameSlug := chi.URLParam(request, "gameSlug")
	wallet := chi.URLParam(request, "wallet")
	memID := strings.Split(authHeader, "\"")[1]
	clientID := chi.URLParam(request, "gamestateID")

	var txStore store.TransactionStore
	var err rgse.RGSErr
	var previousGamestate engine.Gamestate
	switch wallet {
	case "demo":
		txStore, err = store.ServLocal.TransactionByGameId(store.Token(memID), store.ModeDemo, gameSlug)
	case "dashur":
		txStore, err = store.Serv.TransactionByGameId(store.Token(memID), store.ModeReal, gameSlug)
	}
	if err != nil {
		if err.(*rgse.RGSError).ErrCode == rgse.EntityNotFound {
			// this is first gameplay
			txStore, previousGamestate = getInitPlayValues(request, clientID, memID, gameSlug)
		} else {
			//otherwise this is an actual error
			return previousGamestate, store.PlayerStore{}, BalanceResponse{}, engine.EngineConfig{}, err
		}
	} else {
		previousGamestate = store.DeserializeGamestateFromBytes(txStore.GameState)
		// there is a rare case where a player launched a game before we had the proper handling for init cases, here we can detect this by checking if the last tx was an endround with an incomplete gamestate
		if previousGamestate.PreviousGamestate == "" {
			logger.Warnf("Solving Previous Gamestate Issue")
			sentry.CaptureMessage("solving previous gamestate issue")
			txStore, previousGamestate = getInitPlayValues(request, clientID, memID, gameSlug)
		}
	}
	// check that previous gamestate matches what the client expects
	logger.Debugf("Previous id: %v, requested id: %v", previousGamestate.Id, clientID)
	if clientID != previousGamestate.Id {
		// make an exception for engine iii, where on pickSpins the clientID should match the previous previous ID
		if previousGamestate.Action != "pickSpins" || clientID != previousGamestate.PreviousGamestate {
			return engine.Gamestate{}, store.PlayerStore{}, BalanceResponse{}, engine.EngineConfig{}, rgse.Create(rgse.SpinSequenceError)
		}
	}
	logger.Debugf("Previous Gamestate: %v", previousGamestate)
	// get parameters from post form (perhaps this should be handled POST func)

	data = validateParams(data)

	// bugfix for engine xiii (this should really be fixed in the client)
	if gameSlug == "sky-jewels" || gameSlug == "goal" || gameSlug == "cookoff-champion" && len(data.SelectedWinLines) == 49 {
		swl := make([]int, 50)
		for i := 0; i < 50; i++ {
			swl[i] = i
			data.SelectedWinLines = swl
		}
	}
	if txStore.Amount.Currency == "" {
		txStore.Amount.Currency = previousGamestate.BetPerLine.Currency
	}
	data, betValidationErr := validateBet(data, txStore, gameSlug)
	if betValidationErr != nil {
		return engine.Gamestate{}, store.PlayerStore{}, BalanceResponse{}, engine.EngineConfig{}, betValidationErr
	}

	// add suffix to gamestate in case this is a retry attempt
	switch txStore.WalletStatus {
	case 0:
		// this tx is pending in wallet, quit and force reload
		return engine.Gamestate{}, store.PlayerStore{}, BalanceResponse{}, engine.EngineConfig{}, rgse.Create(rgse.PeviousTXPendingError)
	case -1:
		// the next tx failed, retrying it will cause a duplicate tx id error, so add a suffix
		previousGamestate.NextGamestate = previousGamestate.NextGamestate + rng.RandStringRunes(4)
		sentry.CaptureMessage("Adding random suffix to tx id to avoid duplication error")
		logger.Debugf("adding suffix to next tx to avoid duplication error, resulting id: %v", previousGamestate.NextGamestate)
	case 1:
		// business as usual
	default:
		// it should always be one of the above three
		logger.Infof("Wallet Status is unexpectedly %v", txStore.WalletStatus)
		return engine.Gamestate{}, store.PlayerStore{}, BalanceResponse{}, engine.EngineConfig{}, rgse.Create(rgse.UnexpectedWalletStatus)
	}
	gamestate, engineConf := engine.Play(previousGamestate, data.Stake, previousGamestate.BetPerLine.Currency, data)
	if config.GlobalConfig.DevMode == true {
		forcedGamestate, err := forceTool.GetForceValues(data, previousGamestate, txStore.PlayerId)
		if err == nil {
			logger.Warnf("Forcing gamestate: %v", forcedGamestate)
			gamestate = forcedGamestate
		} else {
			//assume error is of memcache.ErrCacheMiss variety
			logger.Infof("No force value found for player %v", txStore.PlayerId)
		}
	}

	var freeGameRef string
	if data.Stake.Mul(engine.NewFixedFromInt(engineConf.EngineDefs[0].StakeDivisor)) == txStore.FreeGames.TotalWagerAmt { // this will also be true if the stake is 2x the wager but half the max lines are selected, but we will have to assume that's ok for the operators as the total exposure is even less
		// this game qualifies as a free game!
		freeGameRef = txStore.FreeGames.CampaignRef
		logger.Debugf("Free game campaign %v", freeGameRef)
	}

	// settle transactions
	var balance store.BalanceStore
	token := txStore.Token
	for _, transaction := range gamestate.Transactions {
		gs := store.SerializeGamestateToBytes(gamestate)
		status := store.RoundStatusOpen
		tx := store.TransactionStore{
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
			FreeGames:           store.FreeGamesStore{NoOfFreeSpins: 0, CampaignRef: freeGameRef},
			Ttl:                 gamestate.GetTtl(),
		}
		switch wallet {
		case "demo":
			tx.Mode = store.ModeDemo
			balance, err = store.ServLocal.Transaction(token, store.ModeDemo, tx)
		case "dashur":
			tx.Mode = store.ModeReal
			balance, err = store.Serv.Transaction(token, store.ModeReal, tx)
		default:
			return engine.Gamestate{}, store.PlayerStore{}, BalanceResponse{}, engine.EngineConfig{}, rgse.Create(rgse.InvalidWallet)
		}

		if err != nil {
			return engine.Gamestate{}, store.PlayerStore{}, BalanceResponse{}, engine.EngineConfig{}, err
		}
		token = balance.Token
	}
	player := store.PlayerStore{Token: token, PlayerId: txStore.PlayerId, Balance: balance.Balance}

	balanceResponse := BalanceResponse{
		Amount:    balance.Balance.Amount,
		Currency:  balance.Balance.Currency,
		FreeGames: balance.FreeGames.NoOfFreeSpins,
	}

	return gamestate, player, balanceResponse, engineConf, nil
}
