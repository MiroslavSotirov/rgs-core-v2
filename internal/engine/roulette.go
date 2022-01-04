package engine

import (
	"encoding/json"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type GameStateRoulette struct {
	GameStateV3

	Position int             `json:"positions"`
	Symbol   int             `json:"symbol"`
	Prizes   []PrizeRoulette `json:"prizes"`
	Bet      Fixed           `json:"bet"`
	Win      Fixed           `json:"win"`
}

func (g *GameStateRoulette) Base() *GameStateV3 {
	return &g.GameStateV3
}

func (s GameStateRoulette) Serialize() []byte {
	b, _ := json.Marshal(s)
	logger.Debugf("GameStateRoulette.Serialize %s", string(b))
	return b
}

func (s GameStateRoulette) GetTtl() int64 {
	return 3600
}

type PrizeRoulette struct {
	Index  string `json:"index"`
	Amount Fixed  `json:"amount"`
}
