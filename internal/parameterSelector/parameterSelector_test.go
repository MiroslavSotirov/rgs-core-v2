package parameterSelector

import (
	"fmt"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	_ "gitlab.maverick-ops.com/maverick/rgs-core-v2/testing"
	"testing"
)

var testPlayer store.PlayerStore = store.PlayerStore{BetLimitSettingCode: "maverick", Balance: engine.Money{Currency: "USD"}}
var testGameID string = "the-year-of-zhu"

func TestLowLastBet(t *testing.T) {
	_, ds, _ := GetGameplayParameters(engine.Fixed(0), testPlayer, testGameID)
	if ds == engine.Fixed(0) {
		t.Error(fmt.Sprintf("Expected last bet to be overridden by default. defaultStake: %v", ds))
	}
	_, ds, _ = GetGameplayParameters(engine.Fixed(10000), testPlayer, testGameID)
	if ds != engine.Fixed(10000) {
		t.Error(fmt.Sprintf("Expected last bet to be maintained. defaultStake: %v", ds))
	}

}

func TestHighLastBet(t *testing.T) {
	_, ds, _ := GetGameplayParameters(engine.Fixed(100000000), testPlayer, testGameID)
	if ds == engine.Fixed(100000000) {
		t.Error(fmt.Sprintf("Expected last bet to be overridden by default. defaultStake: %v", ds))
	}
	sv, ds, _ := GetGameplayParameters(engine.Fixed(500000), testPlayer, testGameID)
	if ds != engine.Fixed(500000) {
		t.Error(fmt.Sprintf("Expected last bet to be maintained. defaultStake: %v; stakeValues: %v", ds, sv))
	}
}

func TestEngineXSetting(t *testing.T) {
	sv, ds, err := GetGameplayParameters(engine.Fixed(0), testPlayer, "seasons")
	// expect sv to be 0.01, 0.02, 0.03, ds to be max of these
	if err != nil {
		t.Error(err.Error())
	}
	if len(sv) != 3 || sv[1] != sv[0].Mul(engine.NewFixedFromInt(2)) || sv[2] != sv[0].Mul(engine.NewFixedFromInt(3)) {
		t.Errorf("Did not get expected 1x,2x,3x stake values for engineX, got %v", sv)
	}
	if ds != sv[2] {
		t.Errorf("Default stake not set to max stakeValues, set to %v", ds)
	}
}

func TestBadCcy(t *testing.T) {
	testPlayer.Balance.Currency = "NIL"
	sv, ds, err := GetGameplayParameters(engine.Fixed(0), testPlayer, testGameID)
	if err == nil {
		t.Error(fmt.Sprintf("Should have gotten error for nil currency. sv: %v; ds: %v", sv, ds))
	}
}
