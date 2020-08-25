package store

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"testing"
	"time"
)

func TestLocalServiceImpl_Player(t *testing.T) {
	logger.NewLogger(logger.Configuration{})

	serv := NewLocal()
	token := Token("test-token")
	player := PlayerStore{
		Token:    token,
		PlayerId: "id-1",
		Username: "id-1-user",
		Balance: engine.Money{
			Currency: "USD",
			Amount:   100,
		},
		BetLimitSettingCode: "DEFAULT",
		FreeGames:           FreeGamesStore{0, "", engine.Fixed(10000)},
	}
	player, _ = serv.PlayerSave(token, ModeReal, player)
	serv2 := New(&config.Config{
		DevMode: true,
	})
	player2, _, _ := serv2.PlayerByToken(player.Token, ModeReal, "gameId-1")

	if player.Token == player2.Token {
		t.Errorf("Value should not match [%v] <> [%v]", player.Token, player2.Token)
	}

	if player.PlayerId != player2.PlayerId {
		t.Errorf("Value not match [%v] <> [%v]", player.PlayerId, player2.PlayerId)
	}

	if player.Username != player2.Username {
		t.Errorf("Value not match [%v] <> [%v]", player.Username, player2.Username)
	}

	if player.Username != player2.Username {
		t.Errorf("Value not match [%v] <> [%v]", player.Username, player2.Username)
	}

	if player.Balance.Currency != player2.Balance.Currency {
		t.Errorf("Value not match [%v] <> [%v]", player.Balance.Currency, player2.Balance.Currency)
	}

	if player.Balance.Amount != player2.Balance.Amount {
		t.Errorf("Value not match [%v] <> [%v]", player.Balance.Amount, player2.Balance.Amount)
	}

	if player.FreeGames.NoOfFreeSpins != player2.FreeGames.NoOfFreeSpins {
		t.Errorf("Value not match [%v] <> [%v]", player.FreeGames.NoOfFreeSpins, player2.FreeGames.NoOfFreeSpins)
	}

	if player.BetLimitSettingCode != player2.BetLimitSettingCode {
		t.Errorf("Value not match [%v] <> [%v]", player.BetLimitSettingCode, player2.BetLimitSettingCode)
	}

	if player.FreeGames.TotalWagerAmt != player2.FreeGames.TotalWagerAmt {
		t.Errorf("Value not match [%v] <> [%v]", player.FreeGames.TotalWagerAmt, player2.FreeGames.TotalWagerAmt)

	}

	logger.Infof("Value of player  => [%v]", player)
	logger.Infof("Value of player2 => [%v]", player2)

	balanceA1, _ := serv.BalanceByToken(token, ModeReal)
	balanceA2, _ := serv2.BalanceByToken(token, ModeReal)

	if balanceA1.Balance.Currency != balanceA2.Balance.Currency {
		t.Errorf("Value not match [%v] <> [%v]", balanceA1.Balance.Currency, balanceA2.Balance.Currency)
	}

	if balanceA1.Balance.Amount != balanceA2.Balance.Amount {
		t.Errorf("Value not match [%v] <> [%v]", balanceA1.Balance.Amount, balanceA2.Balance.Amount)
	}
}

func TestLocalServiceImpl_Transaction(t *testing.T) {
	logger.NewLogger(logger.Configuration{})

	serv := NewLocal()
	token := Token("test-token-3")
	player := PlayerStore{
		Token:    token,
		PlayerId: "id-3",
		Username: "id-3-user",
		Balance: engine.Money{
			Currency: "USD",
			Amount:   100,
		},
		BetLimitSettingCode: "DEFAULT",
		FreeGames:           FreeGamesStore{0, "", 0},
	}
	player, _ = serv.PlayerSave(token, ModeReal, player)

	serv2 := New(&config.Config{
		DevMode: true,
	})

	tx1 := TransactionStore{
		TransactionId: "tx-1",
		Token:         player.Token,
		Mode:          ModeReal,
		Category:      CategoryWager,
		RoundStatus:   RoundStatusOpen,
		PlayerId:      player.PlayerId,
		GameId:        "1",
		RoundId:       "1",
		Amount: engine.Money{
			Currency: player.Balance.Currency,
			Amount:   10,
		},
		ParentTransactionId: "",
		TxTime:              time.Now(),
		GameState:           nil,
	}
	balance, _ := serv2.Transaction(player.Token, ModeReal, tx1)

	if balance.Balance.Amount != 90 {
		t.Errorf("Value not match [%v]", balance.Balance.Amount)
	}

	if player.Token == balance.Token {
		t.Errorf("Value should not match [%v] <> [%v]", player.Token, balance.Token)
	}

	tx2 := TransactionStore{
		TransactionId: "tx-2",
		Token:         player.Token,
		Mode:          ModeReal,
		Category:      CategoryPayout,
		RoundStatus:   RoundStatusClose,
		PlayerId:      player.PlayerId,
		GameId:        "1",
		RoundId:       "1",
		Amount: engine.Money{
			Currency: player.Balance.Currency,
			Amount:   10,
		},
		ParentTransactionId: "",
		TxTime:              time.Now(),
		GameState:           nil,
	}
	balance2, _ := serv2.Transaction(balance.Token, ModeReal, tx2)

	if balance2.Balance.Amount != 100 {
		t.Errorf("Value not match [%v] - ", balance.Balance.Amount)
	}

	if balance.Token == balance2.Token {
		t.Errorf("Value should not match [%v] <> [%v]", balance.Token, balance2.Token)
	}
}

func TestLocalServiceImpl_BalanceByToken(t *testing.T) {
	logger.NewLogger(logger.Configuration{})

	serv := NewLocal()
	token := Token("test-token-4")
	money := engine.Money{
		Currency: "USD",
		Amount:   1000000,
	}

	player := PlayerStore{
		Token:               token,
		PlayerId:            "id-4",
		Username:            "id-4-user",
		Balance:             money,
		BetLimitSettingCode: "DEFAULT",
		FreeGames:           FreeGamesStore{0, "", 0},
	}

	player, err := serv.PlayerSave(token, ModeReal, player)
	if err != nil {
		t.Errorf("Error should not be displayed [%v]", err)
	}

	token = player.Token

	balance, err := serv.BalanceByToken(token, ModeReal)

	if err != nil {
		t.Errorf("Error should not be displayed [%v]", err)
	}

	if balance.Balance.Amount != money.Amount {
		t.Errorf("Value not match [%v] <> [%v]- ", money.Amount, balance.Balance.Amount)
	}

	if balance.Balance.Currency != money.Currency {
		t.Errorf("Value not match [%v] <> [%v] - ", money.Currency, balance.Balance.Currency)
	}

	if balance.Token == token {
		t.Errorf("Value should not match [%v] <> [%v]", token, balance.Token)
	}
}
