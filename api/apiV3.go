package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type paramsV3 interface {
	validate() rgse.RGSErr
	decode(*http.Request) rgse.RGSErr
	deserialize([]byte) rgse.RGSErr
}

type initParamsV3 struct {
	Game     string `json:"game"`
	Operator string `json:"operator"`
	Mode     string `json:"mode"`
	Ccy      string `json:"currency"`
}

type playParamsV3 struct {
	Game       string `json:"game"`
	Wallet     string `json:"wallet"`
	PreviousID string `json:"previousID"`
}

type closeParamsV3 struct {
}

type betRoulette struct {
	Stake   engine.Fixed
	Symbols []int32
}

type IGameInitResponseV3 interface {
	Base() GameInitResponseV3
	Render(http.ResponseWriter, *http.Request) error
}

type GameInitResponseV3 struct {
	Name        string         `json:"name"`
	Version     string         `json:"version"`
	Wallet      string         `json:"wallet"`
	StakeValues []engine.Fixed `json:"stakeValues"`
	DefaultBet  engine.Fixed   `json:"defaultBet"`
}

func (resp GameInitResponseV3) Base() GameInitResponseV3 {
	return resp
}

/* // in rendererV2.go
func (resp GameInitResponseV3) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
*/

type GameInitResponseRoulette struct {
	GameInitResponseV3
	LastRound IGamePlayResponseV3 `json:"lastRound"`
	Reel      []int               `json:"reel"`
}

func (resp GameInitResponseRoulette) base() GameInitResponseV3 {
	return resp.GameInitResponseV3
}

func (resp GameInitResponseRoulette) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type IGamePlayResponseV3 interface {
	Base() GamePlayResponseV3
	Render(http.ResponseWriter, *http.Request) error
}

type GamePlayResponseV3 struct {
	SessionID store.Token        `json:"host/verified-token"`
	StateID   string             `json:"stateID"`
	RoundID   string             `json:"roundID"`
	Stake     engine.Fixed       `json:"totalStake"`
	Win       engine.Fixed       `json:"win"`
	Balance   BalanceResponseV3  `json:"balance"`
	Closed    bool               `json:"closed"`
	Features  []features.Feature `json:"features,omitempty"`
}

func (resp GamePlayResponseV3) Base() GamePlayResponseV3 {
	return resp
}

/* // in rendererV2.go
func (resp GamePlayResponseV3) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
*/

type GamePlayResponseRoulette struct {
	GamePlayResponseV3

	Number int32           `json:"number"`
	Prizes []PrizeRoulette `json:"wins"`
}

func (resp GamePlayResponseRoulette) Base() GamePlayResponseV3 {
	return resp.GamePlayResponseV3
}

func (resp GamePlayResponseRoulette) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type GameState interface {
	Serialize() []byte
	GetTtl() int64
}

type GameStateV3 struct {
	Id                string
	Game              string
	EngineDef         int32
	Currency          engine.Ccy
	Transactions      []engine.WalletTransaction
	PreviousGamestate string
	NextGamestate     string
	Closed            bool
	RoundId           string
	Features          []features.Feature
}

func (s GameStateV3) Serialize() []byte {
	b, _ := json.Marshal(s)
	return b
}

type playParamsRoulette struct {
	playParamsV3

	Bets []betRoulette `json:"bets"`
}

type GameStateRoulette struct {
	GameStateV3

	Position int
	Symbol   int
	Prizes   []*PrizeRoulette
}

func (s GameStateRoulette) GetTtl() int64 {
	return 3600
}

type PrizeRoulette struct {
	Amount  engine.Fixed `json:"amount"`
	Symbols []int32      `json:"symbols"`
}

type BalanceResponseV3 struct {
	Amount engine.Money `json:"amount"`
}

type initParamsRoulette struct {
	initParamsV3
	Bets string `json:"bets"`
}

func initV3(request *http.Request) (response IGameInitResponseV3, rgserr rgse.RGSErr) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		logger.Errorf("request read error")
		return nil, rgse.Create(rgse.JsonError)
	}

	var data initParamsV3
	if rgserr = data.deserialize(body); rgserr != nil {
		return
	}

	var authToken string
	authToken, rgserr = getAuth(request)
	if rgserr != nil {
		return
	}

	var engineId string
	engineId, rgserr = config.GetEngineFromGame(data.Game)
	if rgserr != nil {
		return
	}

	engineConfig := engine.BuildEngineDefs(engineId)

	var wallet string
	wallet, rgserr = config.GetWalletFromOperatorAndMode(data.Operator, data.Mode)
	if rgserr != nil {
		return
	}

	return initGameV3(engineId, wallet, body, engineConfig, store.Token(authToken))
}

func initGameV3(engineId string, wallet string, body []byte, engineConf engine.EngineConfig, token store.Token) (response IGameInitResponseV3, rgserr rgse.RGSErr) {
	// use engine config to call dynamic init method?
	switch engineId {
	case "roulette":
		return initRoulette(engineId, wallet, body, engineConf, token)
	default:
		logger.Errorf("v3 api has no support for engineId %s", engineId)
		break
	}
	return nil, rgse.Create(rgse.EngineNotFoundError)
}

func initRoulette(engineId string, wallet string, body []byte, engineConf engine.EngineConfig, token store.Token) (response IGameInitResponseV3, rgserr rgse.RGSErr) {

	var data initParamsRoulette
	if rgserr = data.deserialize(body); rgserr != nil {
		return nil, rgse.Create(rgse.JsonError)
	}

	gameState := initRouletteGS(data)
	playerID := ""
	gameState.Id = playerID + data.Game + "GSinit"

	balance := store.BalanceStore{
		Token: token,
	}

	playResponse := fillRoulettePlayResponse(gameState, balance)

	response = GameInitResponseRoulette{
		GameInitResponseV3: GameInitResponseV3{
			Name:   gameState.Game,
			Wallet: wallet,
		},
		LastRound: playResponse,
		Reel:      []int{0},
	}
	return
}

func playV3(request *http.Request) (response IGamePlayResponseV3, rgserr rgse.RGSErr) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		logger.Errorf("request read error")
		return nil, rgse.Create(rgse.JsonError)
	}
	var data playParamsV3
	if rgserr = data.decode(request); rgserr != nil {
		return
	}

	var token store.Token
	token, rgserr = handleAuth(request)
	if rgserr != nil {
		return
	}

	rgserr = data.validate()
	if rgserr != nil {
		return
	}

	bfirst := strings.Contains(data.PreviousID, "GSinit")

	//	var fngetstate func(store.Token, store.Mode, string) (store.PlayerStore, store.GameStateStore, rgse.RGSErr) //:= store.Serv.PlayerByToken

	var player store.PlayerStore
	var txStore store.TransactionStore
	var latestStateStore store.GameStateStore
	var latestState GameStateRoulette

	switch data.Wallet {
	case "dashur":
		if bfirst {
			player, latestStateStore, err = store.Serv.PlayerByToken(token, store.ModeReal, data.Game)
		} else {
			txStore, err = store.Serv.TransactionByGameId(token, store.ModeReal, data.Game)
		}
		break
	case "demo":
		if bfirst {
			player, latestStateStore, err = store.ServLocal.PlayerByToken(token, store.ModeDemo, data.Game)
		} else {
			txStore, err = store.ServLocal.TransactionByGameId(token, store.ModeDemo, data.Game)
		}
		break
	default:
		err = rgse.Create(rgse.InvalidWallet)
		return
	}

	latestStateStore = latestStateStore

	if err != nil {
		return
	}

	if bfirst {
		txStore = store.TransactionStore{
			RoundStatus:         store.RoundStatusClose,
			BetLimitSettingCode: player.BetLimitSettingCode,
			PlayerId:            player.PlayerId,
			FreeGames:           player.FreeGames,
			Token:               player.Token,
			Amount:              engine.Money{0, player.Balance.Currency},
			Ttl:                 3600,
		}

		initParams := initParamsRoulette{
			initParamsV3: initParamsV3{
				Game: data.Game,
			},
		}
		latestState = initRouletteGS(initParams)

		//		fmt.Print("%v %v", latestStateStore, txStore)
	} else {
		// GameStateRoulette
		//		latestState := store.DeserializeGamestateFromBytes(txStore.GameState)
		err = json.Unmarshal(txStore.GameState, &latestState)
		if err != nil {
			return nil, rgse.Create(rgse.JsonError)
		}

		switch txStore.WalletStatus {
		case 0:
			// this tx is pending in wallet, quit and force reload
			rgserr = rgse.Create(rgse.PeviousTXPendingError)
			return
		case -1:
			// the next tx failed, retrying it will cause a duplicate tx id error, so add a suffix
			latestState.NextGamestate = latestState.NextGamestate + rng.RandStringRunes(4)
			logger.Debugf("adding suffix to next tx to avoid duplication error, resulting id: %v", latestState.NextGamestate)
		case 1:
			// business as usual
		default:
			// it should always be one of the above three
			logger.Debugf("Wallet status not 1, 0, or -1: %v", txStore)
			rgserr = rgse.Create(rgse.UnexpectedWalletStatus)
			return
		}
	}

	if rgserr != nil {
		return
	}

	engineId, rgserr := config.GetEngineFromGame(data.Game)
	if rgserr != nil {
		return
	}

	playGameV3(engineId, data.Wallet, body, &txStore)

	// look up engine and use config to call dynamic play method

	response, err = getRouletteResults(data, latestState, txStore)

	return
}

func playGameV3(engineId string, wallet string, body []byte, txStore *store.TransactionStore) (response IGamePlayResponseV3, rgserr rgse.RGSErr) {

	switch engineId {
	case "roulette":
		return playRoulette(engineId, wallet, body, txStore)
	default:
		break
	}
	return nil, rgse.Create(rgse.EngineNotFoundError)
}

func playRoulette(engineId string, wallet string, body []byte, txStore *store.TransactionStore) (response IGamePlayResponseV3, rgserr rgse.RGSErr) {
	//	fmt.Printf("%v\n", txStore)
	return
}

func decodeParams(p paramsV3, request *http.Request) rgse.RGSErr {
	decoder := json.NewDecoder(request.Body)
	decoderror := decoder.Decode(p)

	if decoderror != nil {
		return rgse.Create(rgse.JsonError)
	}
	return nil
}

func deserializeParams(p paramsV3, b []byte) rgse.RGSErr {
	err := json.Unmarshal(b, p)
	if err != nil {
		return rgse.Create(rgse.JsonError)
	}
	return nil
}

func (i *initParamsV3) decode(request *http.Request) rgse.RGSErr {
	return decodeParams(i, request)
}

func (i initParamsV3) validate() rgse.RGSErr {
	return nil
}

func (i *initParamsV3) deserialize(b []byte) rgse.RGSErr {
	return deserializeParams(i, b)
}

func (i *initParamsRoulette) decode(request *http.Request) rgse.RGSErr {
	return decodeParams(i, request)
}

func (i initParamsRoulette) validate() rgse.RGSErr {
	return nil
}

func (i *initParamsRoulette) deserialize(b []byte) rgse.RGSErr {
	return deserializeParams(i, b)
}

func (i *playParamsV3) decode(request *http.Request) rgse.RGSErr {
	return decodeParams(i, request)
}

func (i playParamsV3) validate() rgse.RGSErr {
	return nil
}

func (i playParamsV3) deserialize(b []byte) rgse.RGSErr {
	return deserializeParams(&i, b)
}

func initRouletteGS(data initParamsRoulette) GameStateRoulette {
	gameState := GameStateRoulette{
		GameStateV3: GameStateV3{
			Game: data.Game,
		},
		Position: 0,
		Symbol:   0,
		Prizes:   []*PrizeRoulette{},
	}
	return gameState
}

func getRouletteResults(data playParamsV3, latestState GameStateRoulette, txStore store.TransactionStore) (response GamePlayResponseRoulette, err rgse.RGSErr) {

	reel := []int{0, 32, 16, 13, 21, 6, 19, 2, 27, 17, 36, 4, 25, 15, 34, 11, 28, 8, 23, 12, 5, 22, 18, 31, 00, 20, 14, 33, 7, 24, 16, 29, 9, 30, 10, 35, 1, 26}

	position := rng.RandFromRange(len(reel))
	symbol := reel[position]

	gameState := GameStateRoulette{
		GameStateV3: GameStateV3{
			PreviousGamestate: data.PreviousID,
			//			NextGamestate:     string(token),
		},
		Position: position,
		Symbol:   symbol,
	}

	var balance store.BalanceStore
	var freeGameRef string = "" // check apiV2 for how to determine if this is a free game, and set
	/*
		= store.BalanceStore{
			PlayerId: txStore.PlayerId,
			Token:    txStore.Token,
		}
	*/
	stateBytes := gameState.Serialize()
	token := txStore.Token
	for _, t := range gameState.Transactions {
		tx := store.TransactionStore{
			TransactionId:       t.Id,
			Token:               token,
			Category:            store.Category(t.Type),
			RoundStatus:         store.RoundStatusOpen,
			PlayerId:            txStore.PlayerId,
			GameId:              data.Game,
			RoundId:             gameState.RoundId,
			Amount:              t.Amount,
			ParentTransactionId: "",
			TxTime:              time.Now(),
			GameState:           stateBytes,
			BetLimitSettingCode: txStore.BetLimitSettingCode,
			FreeGames:           store.FreeGamesStore{NoOfFreeSpins: 0, CampaignRef: freeGameRef},
			Ttl:                 gameState.GetTtl(),
		}
		balance, err = TransactionByWallet(token, data.Wallet, tx)
		if err != nil {
			return
		}
		token = balance.Token
	}

	response = fillRoulettePlayResponse(gameState, balance)

	return
}

func TransactionByWallet(token store.Token, wallet string, tx store.TransactionStore) (balance store.BalanceStore, err rgse.RGSErr) {
	switch wallet {
	case "demo":
		tx.Mode = store.ModeDemo
		balance, err = store.ServLocal.Transaction(token, store.ModeDemo, tx)
	case "dashur":
		tx.Mode = store.ModeReal
		balance, err = store.Serv.Transaction(token, store.ModeReal, tx)
	default:
		err = rgse.Create(rgse.InvalidWallet)
	}
	return
}

func fillRoulettePlayResponse(gameState GameStateRoulette, balance store.BalanceStore) GamePlayResponseRoulette {

	prizes := []PrizeRoulette{}
	for _, p := range gameState.Prizes {
		prizes = append(prizes, *p)
	}
	/*

		rouletteResponse := GamePlayResponseRoulette{
			GamePlayResponseV3: GamePlayResponseV3{
				SessionID: balance.Token,
				StateID:   gameState.Id,
			},
			Prizes: prizes,
		}
		return &playResponse
	*/

	return GamePlayResponseRoulette{
		GamePlayResponseV3: GamePlayResponseV3{
			SessionID: balance.Token,
			StateID:   gameState.Id,
		},
		Prizes: prizes,
	}
}
