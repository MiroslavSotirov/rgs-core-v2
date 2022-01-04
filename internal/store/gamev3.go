package store

import (
	"encoding/json"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
)

type IGameV3 interface {
	Base() *GameV3
	InitState() engine.IGameStateV3
	SerializeState(engine.IGameStateV3) []byte
	DeserializeState([]byte) (engine.IGameStateV3, rgse.RGSErr)
}

type GameV3 struct {
	Game       string
	EngineId   string
	Wallet     string
	Currency   string
	Token      Token
	EngineConf engine.EngineConfig
}

func (g *GameV3) Base() *GameV3 {
	return g
}

func (g GameV3) InitState() engine.IGameStateV3 {
	return nil
}

func (g GameV3) SerializeState(_ engine.IGameStateV3) []byte {
	return []byte{}
}

func (g GameV3) DeserializeState(serialized []byte) (state engine.IGameStateV3, rgserr rgse.RGSErr) {
	var stateV3 engine.GameStateV3
	err := json.Unmarshal(serialized, &stateV3)
	if err != nil {
		rgserr = rgse.Create(rgse.GamestateByteDeserializerError)
		return
	}
	state = &stateV3
	return
	//	return nil, nil
}

func (g *GameV3) Init(token Token, wallet string, currency string) {
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
