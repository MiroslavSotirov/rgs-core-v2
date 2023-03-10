package parameterSelector

import (
	"fmt"
	"testing"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	_ "gitlab.maverick-ops.com/maverick/rgs-core-v2/testing"
)

var testGameID = "the-year-of-zhu"
var testBetLimitCode = "maverick"
var testCcy = "USD"
var testCompany = "2625894"

func TestLowLastBet(t *testing.T) {
	_, ds, _, _, _ := GetGameplayParameters(engine.Money{engine.Fixed(0), testCcy}, testBetLimitCode, testGameID, "")
	if ds == engine.Fixed(0) {
		t.Error(fmt.Sprintf("Expected last bet to be overridden by default. defaultStake: %v", ds))
	}
	_, ds, _, _, _ = GetGameplayParameters(engine.Money{engine.Fixed(10000), testCcy}, testBetLimitCode, testGameID, "")
	if ds != engine.Fixed(10000) {
		t.Error(fmt.Sprintf("Expected last bet to be maintained. defaultStake: %v", ds))
	}

}

func TestHighLastBet(t *testing.T) {
	_, ds, _, _, _ := GetGameplayParameters(engine.Money{engine.Fixed(100000000), testCcy}, testBetLimitCode, testGameID, "")
	if ds == engine.Fixed(100000000) {
		t.Error(fmt.Sprintf("Expected last bet to be overridden by default. defaultStake: %v", ds))
	}
	sv, ds, _, _, _ := GetGameplayParameters(engine.Money{engine.Fixed(500000), testCcy}, testBetLimitCode, testGameID, "")
	if ds != engine.Fixed(500000) {
		t.Error(fmt.Sprintf("Expected last bet to be maintained. defaultStake: %v; stakeValues: %v", ds, sv))
	}
}

func TestEngineXSetting(t *testing.T) {
	sv, ds, _, _, err := GetGameplayParameters(engine.Money{engine.Fixed(0), testCcy}, testBetLimitCode, "seasons", "")
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
	sv, ds, _, _, err := GetGameplayParameters(engine.Money{engine.Fixed(0), "NIL"}, testBetLimitCode, testGameID, "")
	if err == nil {
		t.Error(fmt.Sprintf("Should have gotten error for nil currency. sv: %v; ds: %v", sv, ds))
	}
}

func TestBetLimit(t *testing.T) {
	game := "battle-of-myths"
	sv, ds, _, mx, err := GetGameplayParameters(engine.Money{engine.Fixed(0), testCcy}, testBetLimitCode, game, testCompany)
	if err != nil {
		t.Error(err.Error())
	}
	if len(sv) < 4 {
		t.Errorf("Bet limit generated too few valid stakes")
	}
	found := false
	for _, s := range sv {
		if s == ds {
			found = true
		}
		if mx > 0 && s > mx {
			t.Errorf("Bet limit allowed bet %f to exeed maxbet %f", s.ValueAsFloat(), mx.ValueAsFloat())
			break
		}
	}
	if !found {
		t.Errorf("Bet limit did not allow the default bet")
	}

	testBetLimitNumStakes("battle-of-myths", "189931772", 14, t)
}

func testBetLimitNumStakes(game string, company string, numStakes int, t *testing.T) {
	sv, _, _, _, err := GetGameplayParameters(engine.Money{engine.Fixed(0), testCcy}, testBetLimitCode, game, company)
	if err != nil {
		t.Error(err.Error())
	}
	if len(sv) != numStakes {
		t.Errorf("Bet limit for company %s did not allow the correct number of stakes %d", company, numStakes)
	}
}
