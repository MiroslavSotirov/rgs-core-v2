package store

import "testing"

// test init gs creation
var testPlayer = PlayerStore{PlayerId: "test"}
func TestCreateInitGS(t *testing.T) {
	gameName := "the-year-of-zhu"
	initGS := CreateInitGS(testPlayer, gameName)
	if initGS.GameID != "the-year-of-zhu:0" || initGS.Id != testPlayer.PlayerId + gameName + "GSinit" {
		t.Errorf("Error initializing gamestate: %v", initGS)
	}
}

