package engine

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"testing"
)

var baseGS = Gamestate{
	Id:            "testbase",
	Game:          "test-respin",
	DefID:         0,
	BetPerLine:    Money{Fixed(1), "BTC"},
	Transactions:  nil,
	NextGamestate: "testbase2",
	Action:        "base",
	SymbolGrid:    [][]int{{1, 1, 1}, {1, 2, 2}, {1, 2, 2}, {2, 2, 2}, {2, 2, 2}},
	Multiplier:    1,
	StopList:      []int{0, 0, 0, 0, 0},
	NextActions:   []string{"finish"},
	Closed:        false,
	RoundID:       "test",
}

var almostFeatureGS = Gamestate{
	Id:            "testbase",
	Game:          "test-respin",
	DefID:         0,
	BetPerLine:    Money{Fixed(1), "BTC"},
	Transactions:  nil,
	NextGamestate: "testbase2",
	Action:        "base",
	SymbolGrid:    [][]int{{1, 1, 1}, {1, 2, 2}, {2, 2, 0}, {2, 2, 0}, {2, 2, 2}},
	Multiplier:    1,
	StopList:      []int{0, 0, 7, 7, 0},
	NextActions:   []string{"finish"},
	Closed:        false,
	RoundID:       "test",
}

func TestRespin(t *testing.T) {
	rng.InitPool()

	//params := GameParams{
	//	Stake:             1,
	//	previousGamestate: baseGS,
	//	Action: "respin",
	//	RespinReel: 0,
	//}

	// reel 0 should give same results as existing gs, all results are equivalent
	if baseGS.ExpectedReelValue(0) != Fixed(300) || baseGS.ExpectedReelValue(1) != Fixed(300) {
		t.Errorf("Expected payout of 300 for any combination got %v", baseGS.ExpectedReelValue(0))
	}

	if baseGS.ExpectedReelValue(2) != Fixed(90) {
		t.Errorf("expected payout should be 90, got %v", baseGS.ExpectedReelValue(2))
	}

	if almostFeatureGS.ExpectedReelValue(4) != Fixed(1503) {
		t.Errorf("expected payout 150, got %v", almostFeatureGS.ExpectedReelValue(4))

	}

	//GS := shuffleDef.ShuffleFlop(params)
	//if GS.SymbolGrid[0][0] !=0 {
	//	//try again because there is a channnce it landed on teh same
	//	GS = shuffleDef.ShuffleFlop(params)
	//	if GS.SymbolGrid[0][0] !=0 {
	//		t.Errorf("prime was probably shuffled %v", GS)
	//	}
	//}
	//if GS.SymbolGrid[1][0] == 1 && GS.SymbolGrid[2][0] ==2 && GS.SymbolGrid[3][0] == 3 && GS.SymbolGrid[4][0] ==4 {
	//	t.Errorf("flop was probably not shuffled %v", GS)
	//}
}
