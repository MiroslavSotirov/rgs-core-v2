package store

import (
	"encoding/json"

	uuid "github.com/satori/go.uuid"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type GameRouletteV3 struct {
	GameV3
}

func (g *GameRouletteV3) Base() *GameV3 {
	return &g.GameV3
}

func (g GameRouletteV3) InitState() engine.IGameStateV3 {
	rouletteState := InitStateRoulette(g.GameV3.Game, g.GameV3.Currency)
	return &rouletteState
}

func (g GameRouletteV3) SerializeState(state engine.IGameStateV3) []byte {
	return state.Serialize()
}

func (g GameRouletteV3) DeserializeState(serializedState []byte) (state engine.IGameStateV3, rgserr rgse.RGSErr) {
	var rouletteState engine.GameStateRoulette
	err := json.Unmarshal(serializedState, &rouletteState)
	if err != nil {
		rgserr = rgse.Create(rgse.GamestateByteDeserializerError)
		return
	}
	state = &rouletteState
	return
}

func InitStateRoulette(game string, currency string) engine.GameStateRoulette {
	id := uuid.NewV4().String()
	nextid := uuid.NewV4().String()
	gameState := engine.GameStateRoulette{
		GameStateV3: engine.GameStateV3{
			Id:            id,
			NextGamestate: nextid,
			Game:          game,
			Version:       "3",
			Currency:      currency,
		},
		Position: 0,
		Symbol:   0,
		Prizes:   []engine.PrizeRoulette{},
	}
	logger.Debugf("init state roulette %#v", gameState)
	return gameState
}
