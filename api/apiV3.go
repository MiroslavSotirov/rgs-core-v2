package api

import (
	"encoding/json"
	"fmt"
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
	init(params []byte, player store.PlayerStore)
	play(params []byte, prevState store.TransactionStore)
}

type GameRouletteV3 struct {
	EngineId   string
	Wallet     string
	Token      store.Token
	EngineConf engine.EngineConfig
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

	logger.Debugf("initV3 data = %s\n", string(body))

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
		logger.Debugf("initV3 gamestate is length 0\n")
		if wallet == "demo" {
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
			//			state, err = json.Marshal()
		}
	}
	response, err = initGameV3(player, engineId, wallet, body, engineConfig, token, state.GameState)

	stakeValues, defaultBet, err := parameterSelector.GetGameplayParameters(engine.Money{Currency: data.Ccy}, player.BetLimitSettingCode, data.Game)

	fmt.Printf("initV3 stakeValues = %#v defaultBet = %#v\n", stakeValues, defaultBet)

	response.Base().StakeValues = stakeValues
	response.Base().DefaultBet = defaultBet

	return
}

func getPlayerAndState(token store.Token, wallet string, game string) (player store.PlayerStore, state store.GameStateStore, rgserr rgse.RGSErr) {
	switch wallet {
	case "dashur":
		player, state, rgserr = store.Serv.PlayerByToken(token, store.ModeReal, game)
	case "demo":
		player, state, rgserr = store.ServLocal.PlayerByToken(token, store.ModeDemo, game)
	default:
		rgserr = rgse.Create(rgse.GenericWalletError)
	}
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
	logger.Debugf("playV3 data = %s\n", string(body))
	var data playParamsV3
	if rgserr = data.deserialize(body); rgserr != nil {
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
	//	var latestStateStore store.GameStateStore
	var latestState GameStateRoulette

	switch data.Wallet {
	case "dashur":
		if bfirst {
			//			player, latestStateStore, err = store.Serv.PlayerByToken(token, store.ModeReal, data.Game)
			player, _, rgserr = store.Serv.PlayerByToken(token, store.ModeReal, data.Game)
		} else {
			txStore, rgserr = store.Serv.TransactionByGameId(token, store.ModeReal, data.Game)
		}
		break
	case "demo":
		if bfirst {
			//			player, latestStateStore, err = store.ServLocal.PlayerByToken(token, store.ModeDemo, data.Game)
			player, _, rgserr = store.ServLocal.PlayerByToken(token, store.ModeDemo, data.Game)
			/*
				var balance store.BalanceStore
				var ctFS int
				var waFS engine.Fixed

				balance, ctFS, waFS, rgserr = parameterSelector.GetDemoWalletDefaults(currency, gameName, "", playerID)
				if rgserr != nil {
					return
				}

				player = store.PlayerStore{
					PlayerId: string(token),
					Token: token,
					Mode: store.ModeDemo,
					Username: "",
					Balance: engine.Money{},
					BetLimitSettingCode: "",
					FreeGames: store.FreeGamesStore{
						NoOfFreeSpins: ctFS,
						CampaignRef: string(token),
						TotalWagerAmt: waFS,
					},
				}
			*/

		} else {
			txStore, rgserr = store.ServLocal.TransactionByGameId(token, store.ModeDemo, data.Game)
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

	if bfirst {
		logger.Debugf("first gameplay")
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
		logger.Debugf("not first gameplay")

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
	var state GameStateV3
	err = json.Unmarshal(txStore.GameState, &state)
	if err != nil {
		return rgse.Create(rgse.GamestateStringDeserializerError)
	}
	logger.Debugf("serialized state: %s", string(txStore.GameState))
	logger.Debugf("deserialized state: %#v", state)
	if state.RoundId != data.RoundID {
		logger.Debugf("state round id %s != data round id %s", state.RoundId, data.RoundID)
		return rgse.Create(rgse.SpinSequenceError)
	}
	/*
		if len(state.NextActions) > 1 {
			// we should not be closing a gameround if the last gamestate has more actions to be completed
			err = rgse.Create(rgse.IncompleteRoundError)
			return
		}
	*/
	state.Closed = true
	roundId := state.RoundId
	if roundId == "" {
		roundId = state.Id
	}
	var serializedState []byte
	serializedState, err = json.Marshal(state)
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

func TransactionByWalletAndGame(token store.Token, wallet string, game string) (txStore store.TransactionStore, rgserr rgse.RGSErr) {
	switch wallet {
	case "demo":
		txStore, rgserr = store.ServLocal.TransactionByGameId(token, store.ModeDemo, game)
	case "dashur":
		txStore, rgserr = store.Serv.TransactionByGameId(token, store.ModeReal, game)
	default:
		rgserr = rgse.Create(rgse.InvalidWallet)
	}
	return
}

func CloseByWallet(token store.Token, wallet string, game string, roundId string, serializedState []byte) (rgserr rgse.RGSErr) {
	switch wallet {
	case "demo":
		_, rgserr = store.ServLocal.CloseRound(token, store.ModeDemo, game, roundId, serializedState, 3600)
	case "dashur":
		_, rgserr = store.Serv.CloseRound(token, store.ModeReal, game, roundId, serializedState, 3600)
	}
	return
}