package store

import (
	rgserror "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"strings"
	"time"
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
	var balance BalanceStore
	if len(latestGamestateStore.GameState) == 0 {
		// assume this is first gameplay
		if wallet == "demo" {
			newPlayer = PlayerStore{playerID, Token(refreshToken), ModeDemo, playerID, engine.Money{5000000000, currency}, host, 0, "www.google.com", "www.maverickslots.com"}
			newPlayer, err = ServLocal.PlayerSave(Token(refreshToken), ModeDemo, newPlayer)
		}
		gsID := newPlayer.PlayerId + gameName + "GSinit"
		latestGamestate = engine.Gamestate{Transactions: []engine.WalletTransaction{{
			Id:     gsID,
			Amount: engine.Money{0, currency},
			Type:   "WAGER",
		}}, GameID: gameName + ":0", Id: gsID, NextActions: []string{"finish"}, Action: "init", Gamification: &engine.GamestatePB_Gamification{}, SymbolGrid: [][]int{{0, 0, 0}, {0, 0, 0}}, NextGamestate: rng.RandStringRunes(8)}
		if strings.Contains(gameName, "seasons") {
			latestGamestate.SelectedWinLines = []int{0, 1, 2}
		}
		shamTx := TransactionStore{
			TransactionId:       latestGamestate.Id,
			Token:               newPlayer.Token,
			Category:            CategoryWager,
			RoundStatus:         RoundStatusClose,
			PlayerId:            newPlayer.PlayerId,
			GameId:              gameName,
			RoundId:             latestGamestate.Id,
			Amount:              engine.Money{0, currency},
			ParentTransactionId: "",
			TxTime:              time.Now(),
			GameState:           SerializeGamestateToBytes(latestGamestate),
		}
		switch wallet {
		case "demo":
			shamTx.Mode = ModeDemo
			balance, err = ServLocal.Transaction(newPlayer.Token, ModeDemo, shamTx)
		case "dashur":
			shamTx.Mode = ModeReal
			balance, err = Serv.Transaction(newPlayer.Token, ModeReal, shamTx)
		}
		newPlayer.Balance = balance.Balance
		newPlayer.Token = balance.Token

	} else {
		latestGamestate = DeserializeGamestateFromBytes(latestGamestateStore.GameState)
	}
	logger.Infof("end of INIT, balance: %v", newPlayer.Balance)
	return latestGamestate, newPlayer, nil
}
