package store

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"testing"
)

// test init gs creation
func TestCreateInitGS(t *testing.T) {
	gameName := "the-year-of-zhu"
	initGS := CreateInitGS(PlayerStore{PlayerId:"test", Balance:engine.Money{0, "USD"}}, gameName)
	if initGS.Game != "the-year-of-zhu" || initGS.DefID != 0 || initGS.Id != "test" + gameName + "GSinit" || initGS.BetPerLine.Currency != "USD"{
		t.Errorf("Error initializing gamestate: %v", initGS)
	}
}

