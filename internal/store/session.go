package store

import (
	uuid "github.com/satori/go.uuid"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/parameterSelector"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"strings"
)

func InitPlayerGS(refreshToken string, playerID string, gameName string, currency string, wallet string) (engine.Gamestate, PlayerStore, rgse.RGSErr) {
	logger.Debugf("init game %v for player %v", gameName, playerID)
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
		logger.Debugf("latest gamestate had length zero")
		if wallet == "demo" {
			// todo : allow setting of betlimitsettignscode
			balance, ctFS, waFS, err := parameterSelector.GetDemoWalletDefaults(currency, gameName, "", playerID)
			// todo: get this per currency
			if err != nil {
				return engine.Gamestate{}, PlayerStore{}, err
			}
			logger.Debugf("balance: %v, freespins: %v, wageramt: %v", balance, ctFS, waFS)

			newPlayer = PlayerStore{playerID, Token(refreshToken), ModeDemo, playerID, balance, "", FreeGamesStore{
				NoOfFreeSpins: ctFS,
				CampaignRef:   playerID,
				WagerAmt:      waFS,
			}}
			newPlayer, err = ServLocal.PlayerSave(newPlayer.Token, ModeDemo, newPlayer)
		}
		latestGamestate = CreateInitGS(newPlayer, gameName)

	} else {
		latestGamestate = DeserializeGamestateFromBytes(latestGamestateStore.GameState)
	}

	return latestGamestate, newPlayer, nil
}

func CreateInitGS(player PlayerStore, gameName string) (latestGamestate engine.Gamestate) {
	// from player we use balance currency and id
	logger.Debugf("First %v gameplay for player %v, creating sham gamestate", gameName, player)

	gsID := player.PlayerId + gameName + "GSinit"
	latestGamestate = engine.Gamestate{Game: gameName, DefID: 0, Id: gsID, BetPerLine: engine.Money{0, player.Balance.Currency}, NextActions: []string{"finish"}, Action: "init", Gamification: &engine.GamestatePB_Gamification{}, SymbolGrid: engine.GetDefaultView(gameName), NextGamestate: uuid.NewV4().String(), Closed: true}
	if strings.Contains(gameName, "seasons") {
		latestGamestate.SelectedWinLines = []int{0, 1, 2}
	}
	return
}

func PlayerBalance(token, wallet string) (BalanceStore, rgse.RGSErr) {
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
		return balance, nil
	case "demo":
		balance, err = ServLocal.BalanceByToken(Token(token), ModeDemo)
		if err != nil {
			logger.Debugf("PlayerBalance Error: %v", err)
			return BalanceStore{}, err
		}
		return balance, nil
	default:
		logger.Debugf("PlayerBalance Error: %v", "No wallet specified")
		return BalanceStore{}, err
	}
}
