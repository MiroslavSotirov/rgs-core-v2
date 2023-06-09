package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/forceTool"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/parameterSelector"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type initParams struct {
	Game     string `json:"game"`
	Operator string `json:"operator"`
	Mode     string `json:"mode"`
	Ccy      string `json:"currency"`
}

type CloseRoundParams struct {
	Game    string `json:"game"`
	Wallet  string `json:"wallet"`
	RoundID string `json:"round"`
}

type FeedParams struct {
	Game      string `json:"game"`
	Wallet    string `json:"wallet"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	PageSize  int    `json:"page_size"`
	Page      int    `json:"page"`
}

type FeedRoundParams struct {
	Game    string `json:"game"`
	Wallet  string `json:"wallet"`
	RoundId int64  `json:"round_id"`
}

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

func getGameHashes(request *http.Request) (GameHashResponse, rgse.RGSErr) {
	response := GameHashResponse{}
	params := request.URL.Query()
	ccys := strings.Split(params.Get("currencies"), ",")
	companyId := params.Get("companyId")
	for _, c := range config.GlobalGameConfig {
		cfg := c.EngineID + ".yml"
		h, ok := config.GlobalHashes[cfg]
		if ok {

			EC := engine.BuildEngineDefs(c.EngineID)

			for _, g := range c.Games {

				var stakes map[string][]engine.Fixed = nil
				if len(ccys) > 0 {
					stakes = make(map[string][]engine.Fixed, len(ccys))
					for _, ccy := range ccys {
						stakeValues, _, _, _, err := parameterSelector.GetGameplayParameters(engine.Money{0, ccy}, "", g.Name, companyId)
						if err == nil {
							for i, _ := range stakeValues {
								stakeValues[i] = stakeValues[i].Mul(engine.NewFixedFromInt(EC.EngineDefs[0].StakeDivisor))
							}
							stakes[ccy] = stakeValues
						}
					}
				}

				category := c.Category
				if category == "" {
					category = "slot"
				}

				title := g.Title
				if title == "" {
					title = DefaultTitleName(g.Name)
				}

				response = append(response, GameHashInfo{
					ItemId:   g.Item,
					Name:     g.Name,
					Title:    title,
					Config:   cfg,
					Md5:      h.MD5Digest,
					Sha1:     h.SHA1Digest,
					Category: category,
					Flags:    g.Flags,
					Stakes:   stakes,
				})
			}
		}
	}
	return response, nil
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

	//	logger.Debugf("Game: %v; operator: %v; mode: %v; request: %#v", data.Game, data.Operator, data.Mode, request)

	engineID, err := config.GetEngineFromGame(data.Game)
	if err != nil {
		return GameInitResponseV2{}, err
	}
	engineConfig := engine.BuildEngineDefs(engineID)
	var authToken string
	authToken, err = getAuth(request)
	if err != nil {
		return GameInitResponseV2{}, err
	}

	// get wallet from operator config
	wallet, err := config.GetWalletFromOperatorAndMode(data.Operator, data.Mode)
	if err != nil {
		return GameInitResponseV2{}, err
	}
	var player store.PlayerStore
	var latestGamestate engine.Gamestate

	latestGamestate, player, err = store.InitPlayerGS(authToken, authToken, data.Game, data.Ccy, wallet)

	if err != nil {
		logger.Debugf("error: %v", err)
		return GameInitResponseV2{}, err
	}

	giResp := fillGameInitPreviousGameplay(latestGamestate, store.BalanceStore{Balance: player.Balance, Token: player.Token, FreeGames: player.FreeGames})
	logger.Debugf("fillGameInitPreviousGameplay")
	giResp.FillEngineInfo(engineConfig)
	logger.Debugf("fillEngineInfo")
	//logger.Debugf("reel response: %v", giResp.ReelSets)
	giResp.Wallet = wallet
	// set stakevalues, links,

	stakeValues, defaultBet, _, _, err := parameterSelector.GetGameplayParameters(latestGamestate.BetPerLine, player.BetLimitSettingCode, data.Game, player.BetSettingId)
	if err != nil {
		logger.Debugf("error: %v", err)
		return GameInitResponseV2{}, err
	}
	sd := engine.NewFixedFromInt(engineConfig.EngineDefs[0].StakeDivisor)
	giResp.DefaultBet = defaultBet
	giResp.DefaultTotal = defaultBet.Mul(sd)
	giResp.StakeValues = stakeValues
	giResp.TotalStakes = make([]engine.Fixed, len(stakeValues))
	giResp.CurrencyDecimals, err = parameterSelector.GetCurrencyMinorUnit(latestGamestate.BetPerLine.Currency)
	if err != nil {
		logger.Debugf("could not get currency minor unit: %v", err)
		return GameInitResponseV2{}, err
	}
	for i := range giResp.TotalStakes {
		giResp.TotalStakes[i] = stakeValues[i].Mul(sd)
	}
	for i := 0; i < len(stakeValues); i++ {
		if player.FreeGames.NoOfFreeSpins > 0 && stakeValues[i].Mul(engine.NewFixedFromInt(engineConfig.EngineDefs[0].StakeDivisor)) == player.FreeGames.TotalWagerAmt {
			giResp.DefaultBet = stakeValues[i]
			logger.Debugf("setting defaultbet to %v for freegames", defaultBet)
			break
		}
	}

	logger.Debugf("initV2 done")
	return giResp, nil
}

func playV2(request *http.Request) (GameplayResponseV2, rgse.RGSErr) {
	var data engine.GameParams
	if err := data.Decode(request); err != nil {
		return GameplayResponseV2{}, err
	}
	/*
		if strings.Contains(data.PreviousID, "GSinit") {
			return playFirst(request, data)
		}
	*/
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

	logger.Debugf("txstore: {%v}, err: %v", txStore, err)
	if err != nil {
		return GameplayResponseV2{}, err
	}
	previousGamestate = store.DeserializeGamestateFromBytes(txStore.GameState)
	// check if the gsID passed in by the client matches that retrieved from the wallet
	if data.PreviousID != previousGamestate.Id {
		logger.Debugf("Previous ID doesn't match: %v, %v", data.PreviousID, previousGamestate.Id)
		return GameplayResponseV2{}, rgse.Create(rgse.SpinSequenceError)
	}
	logger.Debugf("Previous gamestate: %#v", previousGamestate)

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

	var res GameplayResponseV2
	res, err = getRoundResults(data, previousGamestate, txStore)
	return res, err
}

func lastTransaction(token store.Token, wallet string, game string) (store.TransactionStore, rgse.RGSErr) {
	switch wallet {
	case "demo":
		return store.ServLocal.TransactionByGameId(token, store.ModeDemo, game)
	case "dashur":
		return store.Serv.TransactionByGameId(token, store.ModeReal, game)
	}
	logger.Debugf("No such wallet '%v'", wallet)
	return store.TransactionStore{}, rgse.Create(rgse.InvalidWallet)
}

func playRound(request *http.Request) (RoundResponse, rgse.RGSErr) {
	token, autherr := handleAuth(request)
	if autherr != nil {
		return RoundResponse{}, autherr
	}

	var data engine.GameParams
	if err := data.Decode(request); err != nil {
		return RoundResponse{}, err
	}

	var mode store.Mode
	switch data.Wallet {
	case "demo":
		mode = store.ModeDemo
	case "dashur":
		mode = store.ModeReal
	default:
		return RoundResponse{}, rgse.Create(rgse.InvalidWallet)
	}

	txStore, err := lastTransaction(token, data.Wallet, data.Game)
	if err != nil {
		return RoundResponse{}, err
	}

	lastGamestate := store.DeserializeGamestateFromBytes(txStore.GameState)
	previousGamestate := lastGamestate
	// check if the gsID passed in by the client matches that retrieved from the wallet
	if data.PreviousID != previousGamestate.Id {
		logger.Debugf("Previous ID doesn't match: %v, %v", data.PreviousID, previousGamestate.Id)
		return RoundResponse{}, rgse.Create(rgse.SpinSequenceError)
	}

	data, betValidationErr := validateBet(data, txStore, data.Game)
	if betValidationErr != nil {
		return RoundResponse{}, betValidationErr
	}

	if txStore.Amount.Currency == "" {
		txStore.Amount.Currency = previousGamestate.BetPerLine.Currency
	}

	nextAction := "base"
	if len(previousGamestate.NextActions) > 0 && previousGamestate.NextActions[0] != "finish" {
		logger.Debugf("completing unfinished round actions [%#v]", previousGamestate.NextActions)
		nextAction = previousGamestate.NextActions[0]
	}

	var stake, roundWin, campaignWin, lineBet engine.Fixed
	var roundId, freeGameRef string
	var stakeDivisor int
	var ttl int64

	if txStore.FreeGames.NoOfFreeSpins > 0 && data.Stake.Mul(engine.NewFixedFromInt(stakeDivisor)) == txStore.FreeGames.TotalWagerAmt {
		freeGameRef = txStore.FreeGames.CampaignRef
		if txStore.FreeGames.CampaignRef == previousGamestate.CampaignRef {
			campaignWin = previousGamestate.CampaignWin
		}
	}

	type ReplayPoint struct {
		NumSpins        int
		NumTransactions int
		NumTries        int
		Params          feature.FeatureParams
	}

	spins := []SpinResponse{}
	states := []engine.Gamestate{}
	transactions := []store.TransactionStore{}
	replayPoints := []ReplayPoint{}
	var gamestate engine.Gamestate

	for nextAction != "finish" {

		data.Action = nextAction
		data.Replay = nil

		logger.Debugf("PREVIOUS GAMESTATE: %#v", previousGamestate)
		gamestate, _, err = engine.Play(previousGamestate, data.Stake, previousGamestate.BetPerLine.Currency, data)
		if err != nil {
			return RoundResponse{}, err
		}

		if gamestate.Replay {
			replayPoints = append(replayPoints, ReplayPoint{
				NumSpins:        len(spins),
				NumTransactions: len(transactions),
				NumTries:        1,
				Params:          gamestate.ReplayParams,
			})
		}

		if gamestate.NextActions[0] == "finish" {
			if len(replayPoints) > 0 {
				point := replayPoints[len(replayPoints)-1]
				logger.Debugf("Replaying at point %#v", point)
				if point.NumSpins > 0 {
					previousGamestate = states[point.NumSpins-1]
				} else {
					previousGamestate = lastGamestate
				}

				var replaystate engine.Gamestate
				data.Action = previousGamestate.NextActions[0]
				if data.Action == "finish" {
					logger.Debugf("Replaying from first spin in round")
					data.Action = "base"
				}
				data.Replay = states[point.NumSpins:]
				data.ReplayTries = point.NumTries
				data.ReplayParams = point.Params
				if point.Params != nil {
					logger.Debugf("replay params %#v", point.Params)
				}
				replaystate, _, err = engine.Play(previousGamestate, data.Stake, previousGamestate.BetPerLine.Currency, data)

				if replaystate.Replay {
					logger.Debugf("replaying state %d in a sequence of length %d", point.NumSpins, len(states))
					gamestate = replaystate
					spins = spins[point.NumSpins:]
					states = states[point.NumSpins:]
					transactions = transactions[point.NumTransactions:]

					nextAction = previousGamestate.NextActions[0]
				} else {
					logger.Debugf("replaying state %d was completed at replay point %d", point.NumSpins, len(replayPoints)-1)
					replayPoints = replayPoints[1:]
				}
			}
		}

		states = append(states, gamestate)

		/*	TODO: remove the campaign accumulator that required this
			for p := 0; p < len(gamestate.Prizes); p++ {
				gamestate.Prizes[p].Win = engine.NewFixedFromInt(gamestate.Prizes[p].Payout.Multiplier * gamestate.Prizes[p].Multiplier * gamestate.Multiplier).Mul(gamestate.BetPerLine.Amount)
			}
		*/
		if len(transactions) == 0 {
			lineBet = gamestate.BetPerLine.Amount
			roundId = gamestate.RoundID
			ttl = gamestate.GetTtl()
			ED, _ := gamestate.EngineDef()
			stakeDivisor = ED.StakeDivisor
		}

		var win engine.Fixed
		for _, transaction := range gamestate.Transactions {
			AppendHistory(&txStore, transaction)

			switch transaction.Type {
			case "WAGER":
				stake += transaction.Amount.Amount
			case "PAYOUT":
				win += transaction.Amount.Amount
				roundWin += transaction.Amount.Amount
				if freeGameRef != "" {
					campaignWin += transaction.Amount.Amount
				}
			}

			gamestate.CampaignRef = freeGameRef
			gamestate.CampaignWin = campaignWin

			gs := store.SerializeGamestateToBytes(gamestate)
			tx := store.TransactionStore{
				TransactionId:       transaction.Id,
				Token:               token,
				Mode:                mode,
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
				Ttl:                 gamestate.GetTtl(),
				History:             txStore.History,
			}
			transactions = append(transactions, tx)
		}

		spins = append(spins, SpinResponse{
			Action:           gamestate.Action,
			StateID:          gamestate.Id,
			DefID:            gamestate.DefID,
			ReelsetID:        gamestate.ReelsetID,
			Win:              win,
			Freespins:        countFreespinsRemaining(gamestate),
			View:             gamestate.SymbolGrid,
			Prizes:           adjustPrizes(gamestate), // gamestate.Prizes),
			Multiplier:       gamestate.Multiplier,
			CascadePositions: getCascadePositions(gamestate),
			Features:         gamestate.Features,
			FeatureView:      gamestate.FeatureView,
		})

		nextAction = gamestate.NextActions[0]
		previousGamestate = gamestate
	}

	//	gamestate.NextActions = []string{"base"}
	gamestate.Closed = true
	transactions = append(transactions, store.TransactionStore{
		TransactionId: rng.Uuid(),
		Token:         token,
		Mode:          mode,
		Category:      store.CategoryClose,
		RoundStatus:   store.RoundStatusClose,
		PlayerId:      txStore.PlayerId,
		GameId:        data.Game,
		RoundId:       roundId,
		Amount: engine.Money{
			Currency: txStore.Amount.Currency,
			Amount:   0,
		},
		ParentTransactionId: "",
		TxTime:              time.Now(),
		GameState:           store.SerializeGamestateToBytes(gamestate),
		BetLimitSettingCode: txStore.BetLimitSettingCode,
		FreeGames:           store.FreeGamesStore{NoOfFreeSpins: 0, CampaignRef: freeGameRef},
		Ttl:                 ttl,
		History:             txStore.History,
	})

	//	logger.Debugf("playround spins: %#v", spins)

	balance, err := store.GetService(mode).MultiTransaction(token, mode, transactions)

	if err != nil {
		return RoundResponse{}, err
	}

	var fsresp FreespinResponse
	if balance.FreeGames.NoOfFreeSpins > 0 {
		fsresp.CtRemaining = balance.FreeGames.NoOfFreeSpins
		if stakeDivisor != 0 {
			fsresp.WagerAmt = balance.FreeGames.TotalWagerAmt.Div(engine.NewFixedFromInt(stakeDivisor))
		}
	}
	fsresp.TotalWin = campaignWin

	return RoundResponse{
		MetaData:  MetaResponse{},
		SessionID: balance.Token,
		RoundID:   roundId,
		Stake:     stake,
		LineBet:   lineBet,
		Win:       roundWin,
		Balance: BalanceResponseV2{
			Amount:       balance.Balance,
			FreeGames:    balance.FreeGames.NoOfFreeSpins,
			FreeSpinInfo: &fsresp,
		},
		Spins: spins,
	}, nil
}

func getAuth(r *http.Request) (token string, err rgse.RGSErr) {
	authHeader := r.Header.Get("Authorization")
	tokenInfo := strings.Split(authHeader, " ")
	if tokenInfo[0] != "Maverick-Host-Token" || len(tokenInfo) < 2 {
		err = rgse.Create(rgse.InvalidCredentials)
		return
	}
	token = tokenInfo[1]
	return
}

func handleAuth(r *http.Request) (token store.Token, err rgse.RGSErr) {
	var t string
	t, err = getAuth(r)
	if err == nil {
		token = store.Token(t)
	}
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
	var gamestate engine.Gamestate
	var EC engine.EngineConfig
	gamestate, EC, err = engine.Play(previousGamestate, data.Stake, previousGamestate.BetPerLine.Currency, data)
	if err != nil {
		return
	}
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

	gamestate.CampaignWin = engine.Fixed(0)
	gamestate.CampaignRef = ""

	if txStore.FreeGames.NoOfFreeSpins > 0 && data.Stake.Mul(engine.NewFixedFromInt(EC.EngineDefs[0].StakeDivisor)) == txStore.FreeGames.TotalWagerAmt {
		// this game qualifies as a free game!
		freeGameRef = txStore.FreeGames.CampaignRef
		spinWin := gamestate.GetPrizeAmount()
		if freeGameRef == previousGamestate.CampaignRef {
			gamestate.CampaignWin = previousGamestate.CampaignWin + spinWin
		} else {
			gamestate.CampaignWin = spinWin
		}
		gamestate.CampaignRef = freeGameRef
		logger.Infof("Free game campaign %v total win %s", freeGameRef, gamestate.CampaignWin.ValueAsString())
	} else if previousGamestate.RoundID == gamestate.RoundID && gamestate.Transactions[0].Type != "WAGER" {
		// if the game is a continuation of a round propogate the previous campaign ref to all txs linked to this round
		// except if there is a tx on this state that is a wager and it is not the first wager of  the round
		freeGameRef = txStore.FreeGames.CampaignRef
		if freeGameRef != "" {
			gamestate.CampaignWin = previousGamestate.CampaignWin + gamestate.GetPrizeAmount()
			gamestate.CampaignRef = freeGameRef
			logger.Infof("Campaign %v continues total win %s", freeGameRef, gamestate.CampaignWin.ValueAsString())
		}
	}

	autoClose := false
	if data.AutoClose && len(gamestate.NextActions) > 0 && gamestate.NextActions[0] == "finish" {
		autoClose = true
		gamestate.Closed = true
	}

	// settle transactions
	var balance store.BalanceStore
	token := txStore.Token
	logger.Debugf("%v txs", len(gamestate.Transactions))
	roundStatus := store.RoundStatusOpen
	for txIdx, transaction := range gamestate.Transactions {
		logger.Debugf("%#v", transaction)
		AppendHistory(&txStore, transaction)
		if autoClose && txIdx+1 == len(gamestate.Transactions) {
			logger.Debugf("last transaction in the last spin of the round, set RoundStatusClose")
			roundStatus = store.RoundStatusClose
		}
		gs := store.SerializeGamestateToBytes(gamestate)
		tx := store.TransactionStore{
			TransactionId:       transaction.Id,
			Token:               token,
			Category:            store.Category(transaction.Type),
			RoundStatus:         roundStatus, // store.RoundStatusOpen,
			PlayerId:            txStore.PlayerId,
			GameId:              data.Game,
			RoundId:             gamestate.RoundID,
			Amount:              transaction.Amount,
			ParentTransactionId: "",
			TxTime:              time.Now(),
			GameState:           gs,
			BetLimitSettingCode: txStore.BetLimitSettingCode,
			FreeGames:           store.FreeGamesStore{NoOfFreeSpins: 0, CampaignRef: freeGameRef},
			Ttl:                 gamestate.GetTtl(),
			History:             txStore.History,
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
	logger.Debugf("data= %#v token= %#v", data, token)
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
	state := store.SerializeGamestateToBytes(gamestateUnmarshalled)
	ttl := gamestateUnmarshalled.GetTtl()
	switch data.Wallet {
	case "demo":
		_, err = store.ServLocal.CloseRound(token, store.ModeDemo, data.Game, roundId, "", state, ttl, &txStore.History)
	case "dashur":
		_, err = store.Serv.CloseRound(token, store.ModeReal, data.Game, roundId, txStore.FreeGames.CampaignRef, state, ttl, nil)
	}
	return
}

func (i *FeedParams) decode(request *http.Request) rgse.RGSErr {
	decoder := json.NewDecoder(request.Body)
	decoderror := decoder.Decode(i)

	if decoderror != nil {
		return rgse.Create(rgse.JsonError)
	}
	return nil
}

func (i *FeedRoundParams) decode(request *http.Request) rgse.RGSErr {
	decoder := json.NewDecoder(request.Body)
	decoderror := decoder.Decode(i)

	if decoderror != nil {
		return rgse.Create(rgse.JsonError)
	}
	return nil
}

func Feed(r *http.Request) (FeedResponse, rgse.RGSErr) {
	token, autherr := handleAuth(r)
	if autherr != nil {
		return FeedResponse{}, autherr
	}

	var err rgse.RGSErr
	var data FeedParams
	if err := data.decode(r); err != nil {
		return FeedResponse{}, err
	}

	if data.StartTime == "" {
		// Dashur has a limit of 10 day history
		data.StartTime = time.Now().AddDate(0, 0, -10).UTC().Format("2006-01-02 15:04:05.000")
		logger.Debugf("feed round start time set to %v", data.StartTime)
	}
	if data.EndTime == "" {
		data.EndTime = time.Now().UTC().Format("2006-01-02 15:04:05.000")
	}
	if data.PageSize == 0 {
		data.PageSize = 1
	}

	var nextPage int
	var rounds []store.FeedRound
	switch data.Wallet {
	case "demo":
		rounds, nextPage, err = store.ServLocal.Feed(token, store.ModeDemo, data.Game, data.StartTime, data.EndTime, data.PageSize, data.Page)
	case "dashur":
		rounds, nextPage, err = store.Serv.Feed(token, store.ModeReal, data.Game, data.StartTime, data.EndTime, data.PageSize, data.Page)
	default:
		return FeedResponse{}, rgse.Create(rgse.InvalidWallet)
	}

	if err != nil {
		return FeedResponse{}, err
	}

	return FeedResponse{
		Rounds:   rounds,
		NextPage: nextPage,
	}, nil
}

func FeedRound(r *http.Request) (FeedRoundResponse, rgse.RGSErr) {
	token, autherr := handleAuth(r)
	if autherr != nil {
		return FeedRoundResponse{}, autherr
	}

	var err rgse.RGSErr
	var data FeedRoundParams
	if err := data.decode(r); err != nil {
		return FeedRoundResponse{}, err
	}

	var transactions []store.FeedTransaction
	switch data.Wallet {
	case "demo":
		transactions, err = store.ServLocal.FeedRound(token, store.ModeDemo, data.Game, data.RoundId)
	case "dashur":
		transactions, err = store.Serv.FeedRound(token, store.ModeReal, data.Game, data.RoundId)
	default:
		return FeedRoundResponse{}, rgse.Create(rgse.InvalidWallet)
	}

	if err != nil {
		return FeedRoundResponse{}, err
	}

	return FeedRoundResponse{
		Feeds: transactions,
	}, nil
}
