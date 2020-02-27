package store

import "testing"

// test init gs creation
func TestCreateInitGS(t *testing.T) {
	gameName := "the-year-of-zhu"
	initGS := CreateInitGS("test", gameName)
	if initGS.GameID != "the-year-of-zhu:0" || initGS.Id != "test" + gameName + "GSinit" {
		t.Errorf("Error initializing gamestate: %v", initGS)
	}
}

