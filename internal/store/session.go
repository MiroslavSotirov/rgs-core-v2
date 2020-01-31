package store

import (
	rgserror "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"strings"
)

func InitPlayerGS(refreshToken string, playerID string, gameName string, host string, currency string, wallet string) (engine.Gamestate, PlayerStore, rgserror.IRGSError) {
	var newPlayer PlayerStore
	var latestGamestateStore GameStateStore
	var err *Error

	switch wallet {
	case "dashur":
		newPlayer, latestGamestateStore, err = Serv.PlayerByToken(Token(refreshToken), ModeReal, gameName)
	case "demo":
		newPlayer, latestGamestateStore, err = ServLocal.PlayerByToken(Token(refreshToken), ModeDemo, gameName)
	}
	logger.Debugf("newPlayer: %v, latestGS: %v", newPlayer, latestGamestateStore)
	if err != nil {
		logger.Errorf("got err : %v from player retrieval", err)
		return engine.Gamestate{}, PlayerStore{}, rgserror.ErrBalanceStoreError
	}
	var latestGamestate engine.Gamestate

	if len(latestGamestateStore.GameState) == 0 {
		if wallet == "demo" {
			newPlayer = PlayerStore{playerID, Token(refreshToken), ModeDemo, playerID, engine.Money{5000000000, currency}, host, 0}
			newPlayer, err = ServLocal.PlayerSave(newPlayer.Token, ModeDemo, newPlayer)
		}
		latestGamestate = CreateInitGS(newPlayer, gameName)

	} else {
		latestGamestate = DeserializeGamestateFromBytes(latestGamestateStore.GameState)
	}
	logger.Infof("end of INIT, balance: %v", newPlayer.Balance)
	return latestGamestate, newPlayer, nil
}

func CreateInitGS(newPlayer PlayerStore, gameName string) (latestGamestate engine.Gamestate) {
	logger.Debugf("First gameplay for player %v, creating sham gamestate", newPlayer)
	gsID := newPlayer.PlayerId + gameName + "GSinit"
	//todo: initialize gamification properly
	latestGamestate = engine.Gamestate{GameID: gameName + ":0", Id: gsID, NextActions: []string{"finish"}, Action: "init", Gamification: &engine.GamestatePB_Gamification{}, SymbolGrid: [][]int{{0, 0, 0}, {0, 0, 0}}, NextGamestate: rng.RandStringRunes(8)}
	if strings.Contains(gameName, "seasons") {
		latestGamestate.SelectedWinLines = []int{0, 1, 2}
	}
	return
}