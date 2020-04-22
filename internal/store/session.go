package store

import (
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"strings"
)

func InitPlayerGS(refreshToken string, playerID string, gameName string, currency string, wallet string) (engine.Gamestate, PlayerStore, rgse.RGSErr) {
	var newPlayer PlayerStore
	var latestGamestateStore GameStateStore
	var err rgse.RGSErr

	switch wallet {
	case "dashur":
		newPlayer, latestGamestateStore, err = Serv.PlayerByToken(Token(refreshToken), ModeReal, gameName)
	case "demo":
		newPlayer, latestGamestateStore, err = ServLocal.PlayerByToken(Token(refreshToken), ModeDemo, gameName)
	}
	if err != nil && err.(*rgse.RGSError).ErrCode != rgse.NoSuchPlayer {
		return engine.Gamestate{}, PlayerStore{}, err
	}
	var latestGamestate engine.Gamestate

	if len(latestGamestateStore.GameState) == 0 {
		if wallet == "demo" {
			// todo: get this per currency
			balance := engine.NewFixedFromInt(5000)
			freeGames := FreeGamesStore{
				NoOfFreeSpins: 0,
				CampaignRef:   "",
			}
			// solution for testing low balance
			if playerID == "lowbalance" {
				balance = 0
			} else if playerID == "" {
				playerID = rng.RandStringRunes(8)
			} else if strings.Contains(playerID, "campaign") {
				freeGames.NoOfFreeSpins = 10
				freeGames.CampaignRef = playerID
			}
			newPlayer = PlayerStore{playerID, Token(refreshToken), ModeDemo, playerID, engine.Money{balance, currency}, "", freeGames}
			newPlayer, err = ServLocal.PlayerSave(newPlayer.Token, ModeDemo, newPlayer)
		}
		latestGamestate = CreateInitGS(newPlayer, gameName)

	} else {
		latestGamestate = DeserializeGamestateFromBytes(latestGamestateStore.GameState)
	}
	logger.Debugf("latestgs: %v, player %v")
	return latestGamestate, newPlayer, nil
}

func CreateInitGS(player PlayerStore, gameName string) (latestGamestate engine.Gamestate) {
	logger.Debugf("First gameplay for player %v, creating sham gamestate", player)

	gsID := player.PlayerId + gameName + "GSinit"
	//todo: initialize gamification properly
	latestGamestate = engine.Gamestate{GameID: gameName + ":0", Id: gsID, BetPerLine: engine.Money{0, player.Balance.Currency}, NextActions: []string{"finish"}, Action: "init", Gamification: &engine.GamestatePB_Gamification{}, SymbolGrid: engine.GetDefaultView(gameName), NextGamestate: rng.RandStringRunes(16), Closed: true}
	if strings.Contains(gameName, "seasons") {
		latestGamestate.SelectedWinLines = []int{0, 1, 2}
	}
	return
}

func PlayerBalance(token, wallet string) (BalanceStore , rgse.RGSErr) {
	logger.Debugf("Token [%s] Wallet [%s]", token, wallet)
	var balance BalanceStore
	var err rgse.RGSErr
	switch wallet {
	case "dashur":
		balance, err = Serv.BalanceByToken(Token(token), ModeReal)
		if err != nil {
			logger.Debugf("PlayerBalance Error: %v", &err)
			return BalanceStore{}, err
		}
		return balance,  nil
	case "demo":
		balance, err = ServLocal.BalanceByToken(Token(token), ModeDemo)
		if err != nil {
			logger.Debugf("PlayerBalance Error: %v", err)
			return BalanceStore{}, err
		}
		return balance,  nil
	default:
		logger.Debugf("PlayerBalance Error: %v", "No wallet specified")
		return BalanceStore{}, err
	}
}