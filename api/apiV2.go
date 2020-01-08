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
	switch wallet {
	case "demo":
		if authToken == "" {
			authToken = rng.RandStringRunes(6)
		}
		latestGamestate, player, err = store.InitPlayerGS(authToken, authToken, gameSlug, operator, currency)

	case "dashur":
		//we don't use this right now so hack:
		return GameInitResponseV2{}, rgserror.ErrInvalidCredentials
	}
	if err != nil {
		return GameInitResponseV2{}, err
	}
	giResp := fillGameInitPreviousGameplay(latestGamestate, store.BalanceStore{Balance: player.Balance, Token: player.Token}, gameSlug)
	giResp.FillEngineInfo(engineConfig)
	logger.Debugf("reel response: %v", giResp.ReelSets)

	// set stakevalues, links,
	links := make(map[string]string, 1)
	links["new-game"] = fmt.Sprintf("%s%s/%s/play2/%s", GetURLScheme(request), request.Host, APIVersion, gameSlug)
	giResp.Links = links
	stakeValues, defaultBet, err := parameterSelector.GetGameplayParameters(latestGamestate.BetPerLine.Amount, player, gameSlug)
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
	logger.Debugf("Auth string: %v", authHeader)
	memID := strings.Trim(strings.Split(authHeader, " ")[1], "\"")
	logger.Debugf("memID: %v", memID)
	player, previousGamestateStore, err := store.Serv.PlayerByToken(store.Token(memID), store.ModeDemo, gameSlug)
	if err != nil {
		logger.Errorf("Error retrieving player: %v", err)
		// may also be spin sequence mismatch
		return GameplayResponseV2{}, rgserror.ErrInvalidCredentials
	}

	previousGamestate := engine.Gamestate{}
	if len(previousGamestateStore.GameState) == 0 {
		// if no previous gamestate, this is the first gameplay by this player, make sham previous
		logger.Warnf("No previous gameplay, initializing with sham gamestate")
		previousGamestate = engine.Gamestate{Id: player.PlayerId + gameSlug + "GSinit", GameID: fmt.Sprintf("%v:%v", gameSlug, 0), NextActions: []string{"finish"}, NextGamestate: rng.RandStringRunes(8), Transactions: []engine.WalletTransaction{{Type: "WAGER", Id: player.PlayerId + gameSlug + "GSinit", Amount: engine.Money{0, player.Balance.Currency}}}, Gamification: &engine.GamestatePB_Gamification{Level: 0, Stage: 0, RemainingSpins: 0}}
	} else {
		previousGamestate = store.DeserializeGamestateFromBytes(previousGamestateStore.GameState)
		// compare session.LatestGamestate with previous gamestate to see if player has been playing in another window
	}

	// get parameters from post form
	decoder := json.NewDecoder(request.Body)
	var data engine.GameParams
	decoderror := decoder.Decode(&data)
	logger.Debugf("request body: %#v", request.Body)
	if decoderror != nil {
		panic(decoderror)
	}
	if data.Action == "" {
		data.Action = "base"
	}

	gamestate, _ := engine.Play(previousGamestate, data.Stake, player.Balance.Currency, data)
	//log.Printf("Previous Gamestate: %v \n New Gamestate: %v", previousGamestate, gamestate)
	if config.GlobalConfig.DevMode == true {
		forcedGamestate, err := forceTool.GetForceValues(previousGamestate, gameSlug, player.PlayerId)
		if err == nil {
			logger.Warnf("Forcing gamestate: %v", forcedGamestate)
			gamestate = forcedGamestate
		} else {
			//assume error is of memcache.ErrCacheMiss variety
			logger.Warnf("No force value found for this player")
		}
	}

	// settle transactions
	var balance store.BalanceStore
	gamestate.PreviousGamestate = previousGamestate.Id
	for _, transaction := range gamestate.Transactions {
		balance, err = store.Serv.Transaction(player.Token, store.ModeDemo, store.TransactionStore{
			TransactionId:       transaction.Id,
			Token:               player.Token,
			Mode:                store.ModeDemo,
			Category:            store.Category(transaction.Type),
			RoundStatus:         "OPEN",
			PlayerId:            player.PlayerId,
			GameId:              gameSlug,
			RoundId:             gamestate.Id,
			Amount:              transaction.Amount,
			ParentTransactionId: "",
			TxTime:              time.Now(),
			GameState:           store.SerializeGamestateToBytes(gamestate),
		})
	}

	if err != nil {
		logger.Errorf("Error in transactions: %v", err)
		return GameplayResponseV2{}, rgserror.ErrDasInsufficientFundError
	}

	return fillGamestateResponseV2(gamestate, balance), nil
}
