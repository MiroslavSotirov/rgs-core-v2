package store

import (
	"testing"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
)

// test init gs creation
func TestCreateInitGS(t *testing.T) {
	gameName := "the-year-of-zhu"
	initGS := CreateInitGS(PlayerStore{PlayerId: "test", Balance: engine.Money{Amount: 0, Currency: "USD"}}, gameName)
	if initGS.Game != "the-year-of-zhu" || initGS.DefID != 0 || initGS.Id != "test"+gameName+"GSinit" || initGS.BetPerLine.Currency != "USD" {
		t.Errorf("Error initializing gamestate: %v", initGS)
	}
}
