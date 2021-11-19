package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
)

type playParamsV3 struct {
	Game       string        `json:"game"`
	Wallet     string        `json:"wallet"`
	PreviousID string        `json:"previousID"`
	Bets       []betRoulette `json:"bets"`
}

type closeParamsV3 struct {
}

type betRoulette struct {
	Stake   engine.Fixed
	Symbols []int32
}

type GameInitResponseV3 struct {
	Name        string             `json:"name"`
	Version     string             `json:"version"`
	Wallet      string             `json:"wallet"`
	StakeValues []engine.Fixed     `json:"stakeValues"`
	DefaultBet  engine.Fixed       `json:"defaultBet"`
	LastRound   GamePlayResponseV3 `json:"lastRound"`
}

type GamePlayResponseV3 struct {
	SessionID store.Token
	StateID   string `json:"stateID"`
	//	RoundID   string `json:"roundID"`
	Stake    engine.Fixed       `json:"totalStake"`
	Win      engine.Fixed       `json:"win"`
	Balance  BalanceResponseV3  `json:"balance"`
	Number   int32              `json:"number"`
	Prizes   []PrizeRoulette    `json:"wins"`
	Closed   bool               `json:"closed"`
	Features []features.Feature `json:"features,omitempty"`
}

type GameStateV3 struct {
	GameId            engine.GamestatePB_GameID
	EngineDef         int32
	Currency          engine.Ccy
	Transactions      []*engine.WalletTransactionPB
	PreviousGamestate []byte
	NextGamestate     []byte
	Closed            bool
	RoundId           string
	Features          []*engine.FeaturePB
}

type GameStateRoulette struct {
	GameStateV3

	Position int32
	Symbol   int32
	Prizes   []*PrizeRoulette
}

type PrizeRoulette struct {
	Amount  engine.Fixed
	Symbols []int32
}

type BalanceResponseV3 struct {
	Amount engine.Money `json:"amount"`
}

func initV3(request *http.Request) (response GameInitResponseV3, err rgse.RGSErr) {
	var data initParams
	if err = data.decode(request); err != nil {
		return
	}
	var authToken string
	authToken, err = getAuth(request)
	if err != nil {
		return
	}

	engineId, err := config.GetEngineFromGame(data.Game)
	if err != nil {
		return
	}

	//	engineConfig := engine.BuildEngineDefs(engineId)
	_ = engine.BuildEngineDefs(engineId)

	var wallet string
	wallet, err = config.GetWalletFromOperatorAndMode(data.Operator, data.Mode)
	if err != nil {
		return
	}

	gameState := initRouletteGS()
	playerID := ""
	stateId := playerID + data.Game + "GSinit"
	playResponse := fillRoulettePlayResponse(authToken, stateId, gameState)

	response.Wallet = wallet
	//	response.SessionID = authToken
	//	response.Prizes = gameState.Prizes
	response.LastRound = playResponse

	return
}

func playV3(request *http.Request) (response GamePlayResponseV3, err rgse.RGSErr) {
	var data playParamsV3
	if err = data.decode(request); err != nil {
		return
	}

	token, autherr := handleAuth(request)
	if autherr != nil {
		err = autherr
		return
	}

	err = data.validate()
	if err != nil {
		return
	}

	bfirst := strings.Contains(data.PreviousID, "GSinit")

	var fngetstate func(store.Token, store.Mode, string) (store.PlayerStore, store.GameStateStore, rgse.RGSErr) //:= store.Serv.PlayerByToken

	switch data.Wallet {
	case "dashur":
		break
	case "demo":
		break
	}

	_, _, err = fngetstate(token, store.ModeDemo, data.Game)
	if err != nil {
		return
	}
	fmt.Printf("%v %v", bfirst)

	return
}

func (i *playParamsV3) decode(request *http.Request) rgse.RGSErr {
	decoder := json.NewDecoder(request.Body)
	decoderror := decoder.Decode(i)

	if decoderror != nil {
		return rgse.Create(rgse.JsonError)
	}
	return nil
}

func (i playParamsV3) validate() rgse.RGSErr {
	return nil
}

func initRouletteGS() GameStateRoulette {
	gameState := GameStateRoulette{
		GameStateV3: GameStateV3{
			//			GameId:
		},
		Position: 0,
		Symbol:   0,
		Prizes:   []*PrizeRoulette{},
	}
	return gameState
}

func fillRoulettePlayResponse(token string, stateId string, gameState GameStateRoulette) GamePlayResponseV3 {

	prizes := []PrizeRoulette{}
	for _, p := range gameState.Prizes {
		prizes = append(prizes, *p)
	}

	playResponse := GamePlayResponseV3{
		SessionID: store.Token(token),
		StateID:   stateId,
		Prizes:    prizes,
	}

	return playResponse
}
