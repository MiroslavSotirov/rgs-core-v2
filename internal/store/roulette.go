package store

import (
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
	b := state.Serialize()
	return CompressState(b, COMPRESSION_LZW)
	//	return b
}

func (g GameRouletteV3) DeserializeState(serialized []byte) (engine.IGameStateV3, rgse.RGSErr) {
	state, err := g.DeserializeStateRoulette(serialized)
	return &state, err
}

func (g GameRouletteV3) DeserializeStateRoulette(serialized []byte) (state engine.GameStateRoulette, rgserr rgse.RGSErr) {
	var uncompressed []byte
	uncompressed, rgserr = DecompressState(serialized)
	if rgserr != nil {
		return
	}
	//	rgserr = state.Deserialize(serialized)
	rgserr = state.Deserialize(uncompressed)
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
