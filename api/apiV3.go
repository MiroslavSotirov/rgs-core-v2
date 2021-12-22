package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/parameterSelector"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type IGameV3 interface {
	//	init(params []byte, player store.PlayerStore) IGameState
	//	play(params []byte, state IGameState) IGameState  // prevStore store.TransactionStore)
	//	close(params []byte, state IGameState) IGameState // prevStore store.TransactionStore)

	//	initState(data paramsV3) IGameState
	//	initState(state IGameState, ...)
	//	processState(state IGameState, ...) IGamePlayResponse

	//	fillInitResponse() IGameInitResponseV3
	//	fillPlayResponse() IGamePlayResponseV3

	//	Close()
	Base() *GameV3

	InitState() IGameState
	SerializeState(IGameState) []byte
	DeserializeState([]byte) (IGameState, rgse.RGSErr)
}

type GameV3 struct {
	Game       string
	EngineId   string
	Wallet     string
	Currency   string
	Token      store.Token
	EngineConf engine.EngineConfig
}

func (g *GameV3) Base() *GameV3 {
	return g
}

func (g GameV3) InitState() IGameState {
	return nil
}

func (g GameV3) SerializeState(_ IGameState) []byte {
	return []byte{}
}

func (g GameV3) DeserializeState(_ []byte) (IGameState, rgse.RGSErr) {
	return nil, nil
}

func (g *GameV3) Init(token store.Token, wallet string, currency string) {
	g.Token = token
	g.Wallet = wallet
	g.Currency = currency
	g.EngineConf = engine.BuildEngineDefs(g.EngineId)
}

func CreateGameV3FromEngine(engineId string) (IGameV3, rgse.RGSErr) {
	switch engineId {
	case "mvgEngineRoulette1":
		return &GameRouletteV3{
			GameV3: GameV3{
				EngineId: engineId,
			},
		}, nil
	}
	return nil, rgse.Create(rgse.EngineNotFoundError)
}

func CreateGameV3(game string) (IGameV3, rgse.RGSErr) {
	engineId, rgserr := config.GetEngineFromGame(game)
	if rgserr != nil {
		return nil, rgserr
	}
	gameV3, rgserr := CreateGameV3FromEngine(engineId)
	if rgserr != nil {
		return nil, rgserr
	}
	gameV3.Base().Game = game
	return gameV3, nil
}

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

type IGameInitResponseV3 interface {
	Base() *GameInitResponseV3
	Render(http.ResponseWriter, *http.Request) error
}

type GameInitResponseV3 struct {
	Name        string         `json:"name"`
	Version     string         `json:"version"`
	Wallet      string         `json:"wallet"`
	StakeValues []engine.Fixed `json:"stakeValues"`
	DefaultBet  engine.Fixed   `json:"defaultBet"`
}

func (resp *GameInitResponseV3) Base() *GameInitResponseV3 {
	return resp
}

/* // in rendererV2.go
func (resp GameInitResponseV3) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
*/

type IGamePlayResponseV3 interface {
	Base() GamePlayResponseV3
	Render(http.ResponseWriter, *http.Request) error
}

type GamePlayResponseV3 struct {
	Token    store.Token        `json:"token`
	StateId  string             `json:"stateId"`
	RoundId  string             `json:"roundId"`
	Bet      engine.Fixed       `json:"bet"`
	Win      engine.Fixed       `json:"win"`
	Balance  BalanceResponseV3  `json:"balance"`
	Closed   bool               `json:"closed"`
	Features []features.Feature `json:"features,omitempty"`
}

func (resp GamePlayResponseV3) Base() GamePlayResponseV3 {
	return resp
}

/* // in rendererV2.go
func (resp GamePlayResponseV3) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
*/

type IGameState interface {
	Serialize() []byte
	GetTtl() int64
	Base() *GameStateV3
}

type GameStateV3 struct {
	Id                string                     `json:"id"`
	Game              string                     `json:"game"`
	Currency          string                     `json:"ccy"`
	Transactions      []engine.WalletTransaction `json:"transactions"`
	PreviousGamestate string                     `json:"prevGamestate"`
	NextGamestate     string                     `json:"nextGamestate"`
	Closed            bool                       `json:"closed"`
	RoundId           string                     `json:"roundId"`
	Features          []features.Feature         `json:"features"`
}

/*
func (s *GameStateV3) Base() *GameStateV3 {
	return s
}
*/
func (s GameStateV3) Serialize() []byte {
	b, _ := json.Marshal(s)
	logger.Debugf("GameStateV3.Serialize %s", string(b))
	return b
}

func (s GameStateV3) GetTtl() int64 {
	return 3600
}

type BalanceResponseV3 struct {
	Amount engine.Money `json:"amount"`
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

	logger.Debugf("initV3 %s", string(body))

	var authToken string
	authToken, rgserr = getAuth(request)
	if rgserr != nil {
		return
	}
	token := store.Token(authToken)

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

	player, state, paserr := getPlayerAndState(token, wallet, data.Game)
	if paserr != nil && paserr.(*rgse.RGSError).ErrCode != rgse.NoSuchPlayer {
		rgserr = paserr
		logger.Debugf("error in getPlayerAndState %s\n", rgserr.Error())
		return
	}
	if len(state.GameState) == 0 {
		logger.Debugf("initV3 gamestate is length")
		if wallet == "demo" {
			logger.Debugf("initV3 wallet is demo, save a player")
			var balance engine.Money
			var ctFS int
			var waFS engine.Fixed
			balance, ctFS, waFS, rgserr = parameterSelector.GetDemoWalletDefaults(data.Ccy, data.Game, "", authToken)
			if rgserr != nil {
				return
			}

			player = store.PlayerStore{
				PlayerId:            authToken,
				Token:               token,
				Mode:                store.ModeDemo,
				Username:            "",
				Balance:             balance,
				BetLimitSettingCode: "",
				FreeGames: store.FreeGamesStore{
					NoOfFreeSpins: ctFS,
					CampaignRef:   authToken,
					TotalWagerAmt: waFS,
				},
			}
			player, rgserr = store.ServLocal.PlayerSave(token, store.ModeDemo, player)
		}
	}
	response, err = initGameV3(player, engineId, wallet, body, engineConfig, token, state.GameState)

	return
}

// build initial gamestate

func initGameV3(player store.PlayerStore, engineId string, wallet string, body []byte, engineConf engine.EngineConfig, token store.Token, state []byte) (
	response IGameInitResponseV3, rgserr rgse.RGSErr) {
	// use engine config to call dynamic init method?
	switch engineId {
	case "mvgEngineRoulette1":
		return initRoulette(player, engineId, wallet, body, engineConf, token, state)
	default:
		logger.Errorf("v3 api has no support for engineId %s", engineId)
		break
	}
	return nil, rgse.Create(rgse.EngineNotFoundError)
}

func playV3(request *http.Request) (response IGamePlayResponseV3, rgserr rgse.RGSErr) {
	var token store.Token
	token, rgserr = handleAuth(request)
	if rgserr != nil {
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		logger.Errorf("request read error")
		return nil, rgse.Create(rgse.JsonError)
	}
	logger.Debugf("playV3 data: %s\n", string(body))
	var data playParamsV3
	if rgserr = data.deserialize(body); rgserr != nil {
		return
	}
	logger.Debugf("playV3 params: %#v\n", data)

	rgserr = data.validate()
	if rgserr != nil {
		return
	}

	bfirst := strings.Contains(data.PreviousID, "GSinit")

	//	var fngetstate func(store.Token, store.Mode, string) (store.PlayerStore, store.GameStateStore, rgse.RGSErr) //:= store.Serv.PlayerByToken

	var player store.PlayerStore
	var txStore store.TransactionStore
	//	var prevStateStore store.GameStateStore
	var prevIState IGameState // GameStateRoulette

	switch data.Wallet {
	case "dashur":
		logger.Debugf("wallet is dashur")
		if bfirst {
			//			player, prevStateStore, err = store.Serv.PlayerByToken(token, store.ModeReal, data.Game)
			logger.Debugf("store.Serv.PlayerByToken token=%s, mode=%v, game=%s", string(token), store.ModeReal, data.Game)
			player, _, rgserr = store.Serv.PlayerByToken(token, store.ModeReal, data.Game)
			logger.Debugf("store.Serv.PlayerByToken done. player=%#v", player)
		} else {
			logger.Debugf("store.Serv.TransactionByGameId token=%s, mode=%v, game=%s", string(token), store.ModeReal, data.Game)
			txStore, rgserr = store.Serv.TransactionByGameId(token, store.ModeReal, data.Game)
			logger.Debugf("store.Serv.TransactionByGameId done. txStore=%#v", txStore)
		}
		break
	case "demo":
		logger.Debugf("wallet is demo")
		if bfirst {
			//			player, prevStateStore, err = store.ServLocal.PlayerByToken(token, store.ModeDemo, data.Game)
			logger.Debugf("store.ServLocal.PlayerByToken token=%s, mode=%v, game=%s", string(token), store.ModeReal, data.Game)
			player, _, rgserr = store.ServLocal.PlayerByToken(token, store.ModeDemo, data.Game)
			logger.Debugf("store.ServLocal.PlayerByToken done. player=%#v", player)
		} else {
			logger.Debugf("store.ServLocal.TransactionByGameId token=%s, mode=%v, game=%s", string(token), store.ModeReal, data.Game)
			txStore, rgserr = store.ServLocal.TransactionByGameId(token, store.ModeDemo, data.Game)
			logger.Debugf("store.ServLocal.TransactionByGameId done. txStore=%#v", txStore)
		}
		break
	default:
		logger.Errorf("unknown wallet\n")
		rgserr = rgse.Create(rgse.InvalidWallet)
		return
	}

	if rgserr != nil {
		if bfirst && rgserr.(*rgse.RGSError).ErrCode == rgse.NoSuchPlayer {
			rgserr = nil
		} else {
			logger.Debugf("rgserr = %s\n", rgserr.Error())
			return
		}
	}

	var gameV3 IGameV3
	gameV3, rgserr = CreateGameV3(data.Game)
	gameV3.Base().Init(token, data.Wallet, player.Balance.Currency)
	if rgserr != nil {
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
		logger.Debugf("first gameplay. transaction: %#v", txStore)

		prevIState = gameV3.InitState()
	} else {
		logger.Debugf("not first gameplay")

		// GameStateRoulette
		//		prevState := store.DeserializeGamestateFromBytes(txStore.GameState)
		//		err = json.Unmarshal(txStore.GameState, &prevState)
		//		if err != nil {
		//			return nil, rgse.Create(rgse.JsonError)
		//		}
		prevIState, rgserr = gameV3.DeserializeState(txStore.GameState)
		if rgserr != nil {
			return
		}
		var prevState *GameStateV3 = prevIState.Base()
		if txStore.Amount.Currency == "" {
			logger.Debugf("previous transaction has no currency, using prev gamestate setting: %s", prevState.Currency)
			txStore.Amount.Currency = prevState.Currency
		}
		switch txStore.WalletStatus {
		case 0:
			// this tx is pending in wallet, quit and force reload
			rgserr = rgse.Create(rgse.PeviousTXPendingError)
			return
		case -1:
			// the next tx failed, retrying it will cause a duplicate tx id error, so add a suffix
			prevState.NextGamestate = prevState.NextGamestate + rng.RandStringRunes(4)
			logger.Debugf("adding suffix to next tx to avoid duplication error, resulting id: %v", prevState.NextGamestate)
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

	return playGameV3(engineId, data.Wallet, body, txStore)
}

func validateState(state IGameState) rgse.RGSErr {
	return nil
}

func playGameV3(engineId string, wallet string, body []byte, txStore store.TransactionStore) (response IGamePlayResponseV3, rgserr rgse.RGSErr) {

	switch engineId {
	case "mvgEngineRoulette1":
		return playRoulette(engineId, wallet, body, txStore)
	default:
		break
	}
	return nil, rgse.Create(rgse.EngineNotFoundError)
}

func closeV3(request *http.Request) rgse.RGSErr {
	token, rgserr := handleAuth(request)
	if rgserr != nil {
		return rgserr
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		logger.Errorf("request read error")
		return rgse.Create(rgse.JsonError)
	}
	logger.Debugf("closeV3 data = %s\n", string(body))
	var data CloseRoundParams
	if rgserr = data.deserialize(body); rgserr != nil {
		return rgserr
	}

	gameV3, rgserr := CreateGameV3(data.Game)
	if rgserr != nil {
		return rgserr
	}

	var txStore store.TransactionStore
	txStore, rgserr = TransactionByWalletAndGame(token, data.Wallet, data.Game)
	if rgserr != nil {
		return rgserr
	}
	if txStore.WalletStatus != 1 {
		// if this is zero, the tx is pending and shouldn't be resent, if it is -1, the tx is failed and an error should be sent to reload the client
		logger.Debugf("INTERNAL STATUS: %v", txStore.WalletStatus)
		return rgse.Create(rgse.PeviousTXPendingError)
	}
	istate, rgserr := gameV3.DeserializeState(txStore.GameState)
	if rgserr != nil {
		return rgserr
	}
	var state *GameStateV3 = istate.Base()

	logger.Debugf("serialized state: %s", string(txStore.GameState))
	logger.Debugf("deserialized state: %#v", state)
	if state.RoundId != data.RoundID {
		logger.Debugf("state round id %s != data round id %s", state.RoundId, data.RoundID)
		return rgse.Create(rgse.SpinSequenceError)
	}
	state.Closed = true
	roundId := state.RoundId
	if roundId == "" {
		roundId = state.Id
	}
	serializedState := istate.Serialize()

	CloseByWallet(token, data.Wallet, data.Game, roundId, serializedState)

	return nil
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

func (i *playParamsV3) decode(request *http.Request) rgse.RGSErr {
	return decodeParams(i, request)
}

func (i playParamsV3) validate() rgse.RGSErr {
	return nil
}

func (i *playParamsV3) deserialize(b []byte) rgse.RGSErr {
	return deserializeParams(i, b)
}

func (i CloseRoundParams) validate() rgse.RGSErr {
	return nil
}

func (i *CloseRoundParams) deserialize(b []byte) rgse.RGSErr {
	return deserializeParams(i, b)
}

func getPlayerAndState(token store.Token, wallet string, game string) (player store.PlayerStore, state store.GameStateStore, rgserr rgse.RGSErr) {
	logger.Debugf("getPlayerAndState token=%s, wallet=%s, game=%s", string(token), wallet, game)
	switch wallet {
	case "dashur":
		player, state, rgserr = store.Serv.PlayerByToken(token, store.ModeReal, game)
	case "demo":
		player, state, rgserr = store.ServLocal.PlayerByToken(token, store.ModeDemo, game)
	default:
		rgserr = rgse.Create(rgse.GenericWalletError)
	}
	logger.Debugf("getPlayerAndState done. player=%#v", player)
	return
}

func TransactionByWallet(token store.Token, wallet string, tx store.TransactionStore) (balance store.BalanceStore, err rgse.RGSErr) {
	logger.Debugf("TransactionByWallet token:%s, wallet:%s transactionId:%s", token, wallet, tx.TransactionId)
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
	logger.Debugf("TransactionByWallet done. balance=%#v", balance)
	return
}

func TransactionByWalletAndGame(token store.Token, wallet string, game string) (txStore store.TransactionStore, rgserr rgse.RGSErr) {
	logger.Debugf("TransactionByWalletAndGame token=%s, wallet=%s, game=%s", string(token), wallet, game)
	switch wallet {
	case "demo":
		txStore, rgserr = store.ServLocal.TransactionByGameId(token, store.ModeDemo, game)
	case "dashur":
		txStore, rgserr = store.Serv.TransactionByGameId(token, store.ModeReal, game)
	default:
		rgserr = rgse.Create(rgse.InvalidWallet)
	}
	logger.Debugf("TransactionByWalletAndGame done.")
	return
}

func CloseByWallet(token store.Token, wallet string, game string, roundId string, serializedState []byte) (rgserr rgse.RGSErr) {
	logger.Debugf("CloseByWallet token=%s, wallet=%s, game=%s, serializedState=%s", string(token), wallet, game, string(serializedState))
	switch wallet {
	case "demo":
		_, rgserr = store.ServLocal.CloseRound(token, store.ModeDemo, game, roundId, serializedState, 3600)
	case "dashur":
		_, rgserr = store.Serv.CloseRound(token, store.ModeReal, game, roundId, serializedState, 3600)
	}
	logger.Debugf("CloseByWallet done")
	return
}
