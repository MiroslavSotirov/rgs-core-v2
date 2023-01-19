package store

import (
	"strings"
	"time"

	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/parameterSelector"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

func InitPlayerGS(refreshToken string, playerID string, gameName string, currency string, wallet string) (latestGamestate engine.Gamestate, newPlayer PlayerStore, err rgse.RGSErr) {
	logger.Debugf("init game %v for player %v", gameName, playerID)

	err = ValidateWallet(wallet)
	if err != nil {
		return
	}

	var latestGamestateStore GameStateStore
	mode := WalletMode(wallet)
	newPlayer, latestGamestateStore, err = GetService(mode).PlayerByToken(Token(refreshToken), mode, gameName)
	if err != nil && err.(*rgse.RGSError).ErrCode != rgse.NoSuchPlayer {
		return
	}

	if len(latestGamestateStore.GameState) == 0 {
		logger.Debugf("latest gamestate had length zero")
		if wallet == "demo" {
			// todo : allow setting of betlimitsettignscode
			balance, ctFS, waFS, err := parameterSelector.GetDemoWalletDefaults(currency, gameName, "", playerID, newPlayer.CompanyId)

			if err != nil {
				return engine.Gamestate{}, PlayerStore{}, err
			}
			logger.Debugf("balance: %v, freespins: %v, wageramt: %v", balance, ctFS, waFS)

			newPlayer = PlayerStore{
				PlayerId:            playerID,
				Token:               Token(refreshToken),
				Mode:                ModeDemo,
				Username:            playerID,
				Balance:             balance,
				BetLimitSettingCode: "",
				CompanyId:           newPlayer.CompanyId,
				FreeGames: FreeGamesStore{
					NoOfFreeSpins: ctFS,
					CampaignRef:   playerID,
					TotalWagerAmt: waFS,
				}}
			newPlayer, err = ServLocal.PlayerSave(newPlayer.Token, ModeDemo, newPlayer)
		}
		latestGamestate = CreateInitGS(newPlayer, gameName)
		txamount := engine.Money{0, newPlayer.Balance.Currency}
		txtype := CategoryWager
		latestGamestate.Transactions = []engine.WalletTransaction{
			engine.WalletTransaction{
				Id:     latestGamestate.Id,
				Amount: txamount,
				Type:   string(txtype),
			},
		}
		transaction := TransactionStore{
			TransactionId: latestGamestate.Id,
			Token:         newPlayer.Token,
			Mode:          mode,
			Category:      txtype,
			RoundStatus:   RoundStatusClose,
			PlayerId:      playerID,
			GameId:        gameName,
			RoundId:       latestGamestate.Id,
			Amount:        txamount,
			TxTime:        time.Now(),
			GameState:     SerializeGamestateToBytes(latestGamestate),
			FreeGames:     newPlayer.FreeGames,
			Ttl:           latestGamestate.GetTtl(),
		}
		logger.Debugf("initial gamestate: %#v", latestGamestate)

		// store the initial gamestate
		var balanceStore BalanceStore
		balanceStore, err = GetService(mode).Transaction(newPlayer.Token, mode, transaction)

		if err != nil {
			logger.Debugf("initial gamestate transaction failed")
			return
		}
		newPlayer.Token = balanceStore.Token
		newPlayer.Balance = balanceStore.Balance

	} else {
		latestGamestate = DeserializeGamestateFromBytes(latestGamestateStore.GameState)
		//		_, initFeatures, _, _ = engine.GetDefaultView(gameName)
	}

	return
}

func CreateInitGS(player PlayerStore, gameName string) (latestGamestate engine.Gamestate) {
	// from player we use balance currency and id
	logger.Debugf("Creating sham gamestate. First %v gameplay for player %#v, ", gameName, player)

	//	gsID := player.PlayerId + gameName + "GSinit"
	gsID := string(rng.Uuid())
	view, features, defId, reelsetId := engine.GetDefaultView(gameName)
	latestGamestate = engine.Gamestate{
		Game:          gameName,
		DefID:         defId,
		ReelsetID:     reelsetId,
		Id:            gsID,
		BetPerLine:    engine.Money{0, player.Balance.Currency},
		NextActions:   []string{"finish"},
		Action:        "base",
		Gamification:  &engine.GamestatePB_Gamification{},
		SymbolGrid:    view,
		Features:      features,
		NextGamestate: rng.Uuid(),
		Closed:        true,
	}
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

func SetPlayerBalance(token string, wallet string, balance engine.Money) rgse.RGSErr {
	switch wallet {
	case "demo":
		err := ServLocal.SetBalance(Token(token), balance)
		if err != nil {
			return err
		}
	default:
		logger.Errorf("SetPlayerBalance is only available in demo mode")
		return rgse.Create(rgse.InvalidWallet)
	}
	return nil
}
