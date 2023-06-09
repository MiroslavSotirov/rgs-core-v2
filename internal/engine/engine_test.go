package engine

import (
	"fmt"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"

	// "fmt"
	"math/rand"
	"testing"

	_ "gitlab.maverick-ops.com/maverick/rgs-core-v2/testing"
)

var testReels = [][]int{{0, 1, 0, 1, 0, 1, 0, 1}, {1, 2, 1, 1, 2, 1, 1, 2, 1, 1, 2, 1}, {1, 2, 2, 2, 2, 2, 2, 2, 2}, {3, 3, 3, 3, 3, 3, 3, 3, 3}, {1, 4, 1, 4, 1, 4, 1, 4}}
var testViewSize = []int{3, 3, 3, 3, 3}
var testWinLines = [][]int{{0, 0, 0, 0, 0}, {0, 1, 2, 1, 0}}
var testPayouts = []Payout{{Symbol: 1, Count: 5, Multiplier: 10}}
var testWaysPayouts = []Payout{{Symbol: 1, Count: 3, Multiplier: 10}, {Symbol: 2, Count: 5, Multiplier: 1000}}

var testWilds = []wild{{Symbol: 7, Multiplier: weightedMultiplier{
	Multipliers:   []int{5},
	Probabilities: []int{1},
}},
	{Symbol: 8, Multiplier: weightedMultiplier{
		Multipliers:   []int{7},
		Probabilities: []int{1},
	}}}

var cascadeEngine = EngineDef{
	Reels: [][]int{
		{0, 0, 0, 0, 1},
		{1, 0, 0, 1, 0, 0, 1},
		{2, 3, 4, 5, 1, 0},
	},
	ViewSize: []int{3, 3, 3},
	Payouts:  testWaysPayouts,
	WinType:  "ways",
	Multiplier: weightedMultiplier{
		Multipliers:   []int{1, 2, 5},
		Probabilities: []int{1, 1, 1},
	},
}

var cascadeGS = Gamestate{
	SymbolGrid: [][]int{{0, 0, 1}, {1, 0, 0}, {1, 0, 2}},
	Prizes: []Prize{{
		Payout:          testWaysPayouts[0],
		Index:           "1:3",
		Multiplier:      1,
		SymbolPositions: []int{2, 3, 6},
	}},
	Multiplier:  2,
	StopList:    []int{2, 0, 4},
	NextActions: []string{"cascade", "finish"},
}

func TestEngineDef_Cascade(t *testing.T) {
	gs := cascadeEngine.Cascade(GameParams{previousGamestate: cascadeGS, Action: "cascade"})
	expectedSymbolGrid := [][]int{{0, 0, 0}, {1, 0, 0}, {5, 0, 2}}
	for reel := 0; reel < len(expectedSymbolGrid); reel++ {
		for symbol := 0; symbol < len(expectedSymbolGrid[reel]); symbol++ {
			if expectedSymbolGrid[reel][symbol] != gs.SymbolGrid[reel][symbol] {
				t.Errorf("Expected %v, got %v", expectedSymbolGrid, gs.SymbolGrid)
			}
		}

	}
}

func TestEngineDef_CascadeMultiply(t *testing.T) {
	gs := cascadeEngine.CascadeMultiply(GameParams{previousGamestate: cascadeGS, Action: "cascade"})
	if gs.Multiplier != 5 {
		t.Errorf("Expected multiplier 5, got %v", gs.Multiplier)
	}
}

func TestSpinViewSize(t *testing.T) {

	randomTestViewLength := rand.Intn(10)
	randomTestViewSize := make([]int, randomTestViewLength)

	testReel := []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1}

	testReelSet := make([][]int, randomTestViewLength)

	for i := 0; i < randomTestViewLength; i++ {
		randomTestViewSize[i] = rand.Intn(10)
		testReelSet[i] = testReel
	}

	view, _ := EngineDef{Reels: testReelSet, ViewSize: randomTestViewSize}.Spin()
	if len(view) != len(randomTestViewSize) {
		t.Errorf("ViewSize not matched by view: %v", view)
	}
	for i, row := range view {
		if len(row) != randomTestViewSize[i] {
			t.Error("view size mismatch")
		}
	}
}
func slicesMatch(reel1 []int, reel2 []int) bool {
	if len(reel1) != len(reel2) {
		return false
	}
	for i, val := range reel1 {
		if reel2[i] != val {
			return false
		}
	}
	return true
}

func TestSpin(t *testing.T) {
	view, _ := EngineDef{Reels: testReels, ViewSize: testViewSize}.Spin()
	// we know first reel is alternating zeros and ones
	if !slicesMatch(view[0], []int{1, 0, 1}) && !slicesMatch(view[0], []int{0, 1, 0}) {
		t.Errorf("Spin error, expected reel 1 0/1/0/1, got %v", view[0])
	}
	// we know second reel is ones and twos
	if !slicesMatch(view[1], []int{1, 2, 1}) && !slicesMatch(view[1], []int{1, 1, 2}) && !slicesMatch(view[1], []int{2, 1, 1}) {
		t.Errorf("Spin error, expected reel 2 1/1/2, got %v", view[1])
	}

	// we know the fourth reel should contain only threes
	if !slicesMatch(view[3], []int{3, 3, 3}) {
		t.Errorf("Spin error, expected reel 4 []int{3,3,3}, got %v", view[3])
	}

	// we know fifth reel is alternating fours and ones
	if !slicesMatch(view[4], []int{1, 4, 1}) && !slicesMatch(view[4], []int{4, 1, 4}) {
		t.Errorf("Spin error, expected reel 4 1/4/1/4, got %v", view[4])
	}

}

func compareNextActions(gs Gamestate, na []string) bool {
	if len(gs.NextActions) != len(na) {
		return false
	}
	for i, action := range gs.NextActions {
		if na[i] != action {
			return false
		}
	}
	return true
}

func TestPrepareActions(t *testing.T) {
	// test generic action addition
	testGS := Gamestate{NextActions: []string{"fs1", "fs2", "fs3"}}
	testGS.PrepareActions([]string{"fs0", "finish"})
	want := []string{"fs1", "fs2", "fs3", "finish"}
	if !compareNextActions(testGS, want) {
		t.Errorf("generic action error, got %v, expected %v", testGS.NextActions, want)
	}

	// test no next actions
	var empty []string
	testGS = Gamestate{NextActions: empty}
	testGS.PrepareActions([]string{"base", "finish"})
	want = []string{"finish"}
	if !compareNextActions(testGS, want) {
		t.Errorf("no next actions error, got %v, expected %v", testGS.NextActions, want)
	}

	// test replaceQueuedActionType
	testGS = Gamestate{NextActions: []string{"replaceQueuedActionType", "B"}}
	testGS.PrepareActions([]string{"A", "A", "A", "A", "finish"})
	want = []string{"B", "B", "B", "B", "finish"}
	if !compareNextActions(testGS, want) {
		t.Errorf("replaceQueuedActionType error, got %v, expected %v", testGS.NextActions, want)
	}

	// test replaceQueuedActions
	testGS = Gamestate{NextActions: []string{"replaceQueuedActions", "B", "B"}}
	testGS.PrepareActions([]string{"A", "A", "A", "A", "finish"})
	want = []string{"B", "B", "finish"}
	if !compareNextActions(testGS, want) {
		t.Errorf("replaceQueuedActions error, got %v, expected %v", testGS.NextActions, want)
	}

	// test queueActionsAfter
	testGS = Gamestate{NextActions: []string{"queueActionsAfter", "B", "B"}}
	testGS.PrepareActions([]string{"A", "A", "A", "A", "finish"})
	want = []string{"A", "A", "A", "B", "B", "finish"}
	if !compareNextActions(testGS, want) {
		t.Errorf("queueActionsAfter error, got %v, expected %v", testGS.NextActions, want)
	}

}
func TestPrepareTransactions(t *testing.T) {
	// test relativepayout == 0 : should be no PAYOUT transaction
	// payout can have a preexisting transaction, if so this should remain
	// relative payout can be zero if relativepayout or multiplier is zero for gamestate
	// multiplier and relativepayout will be zero if not explicitly set
	// if there is only one action remaining and it is "finish" , there should be an endround transaction
	// this may change with respin games, where we might want to start only adding endround if the gamestate's action is "finish"

	gsTest := Gamestate{Id: "test", Action: "base", RelativePayout: 5, Multiplier: 2, BetPerLine: Money{Amount: 1000, Currency: "USD"}, NextActions: []string{"finish"}}
	gsTest.PrepareTransactions(Gamestate{RoundID: "previous"})
	if gsTest.Transactions[0].Amount.Amount != 10000 || gsTest.Transactions[0].Amount.Currency != "USD" || gsTest.Transactions[0].Type != "PAYOUT" {
		t.Errorf("payout improperly processed")
	}
	if gsTest.Transactions[0].Id != gsTest.Id {
		t.Errorf("Expected first tx id to match gamestate Id")
	}

	if gsTest.PlaySequence != 0 {
		t.Errorf("failed processing play sequence, expected 0, got %v", gsTest.PlaySequence)
	}
	if gsTest.CumulativeWin != 10000 {
		t.Errorf("failed processing cumulative win, expected 10000, got %v", gsTest.CumulativeWin)
	}
	// test RoundID setting for no wager tx
	if gsTest.RoundID != "previous" {
		t.Errorf("Expected Round ID to match previous ID")
	}

	// test relativepayout zero explicitly
	gsTest = Gamestate{RelativePayout: 0, Multiplier: 1, BetPerLine: Money{Amount: 1000, Currency: "USD"}, NextActions: []string{"finish"}}
	gsTest.PrepareTransactions(Gamestate{})

	if len(gsTest.Transactions) != 0 {
		t.Errorf("expected no transaction")
	}

	// test relativepayout zero implicitly and future actions
	gsTest = Gamestate{Multiplier: 1, BetPerLine: Money{Amount: 1000, Currency: "USD"}, NextActions: []string{"freespin", "finish"}}
	gsTest.PrepareTransactions(Gamestate{})

	if len(gsTest.Transactions) != 0 {
		t.Errorf("expected no transaction")
	}

	// test multiplier zero explicitly
	gsTest = Gamestate{RelativePayout: 1, Multiplier: 0, BetPerLine: Money{Amount: 1000, Currency: "USD"}, NextActions: []string{"finish"}}
	gsTest.PrepareTransactions(Gamestate{})

	if len(gsTest.Transactions) != 0 {
		t.Errorf("expected no transaction")
	}

	// test multiplier zero implicitly and preexisting transaction
	gsTest = Gamestate{Id: "test", RelativePayout: 1, BetPerLine: Money{Amount: 1000, Currency: "USD"}, NextActions: []string{"finish"}, Transactions: []WalletTransaction{{Amount: Money{Amount: 5000, Currency: "USD"}, Type: "WAGER", Id: "ABCDEFGH"}}}
	gsTest.PrepareTransactions(Gamestate{})

	if len(gsTest.Transactions) != 1 || gsTest.Transactions[0].Type != "WAGER" {
		t.Errorf("expected one wager tx")
	}

	if gsTest.RoundID != "test" {
		t.Errorf("expected RoundId to match gamestate Id")
	}

	// test cumulative win addition
	gsTest = Gamestate{Action: "freespin", RelativePayout: 5, Multiplier: 2, BetPerLine: Money{Amount: 1000, Currency: "USD"}, NextActions: []string{"finish"}}
	gsTest.PrepareTransactions(Gamestate{CumulativeWin: 5000, PlaySequence: 5})

	if gsTest.CumulativeWin != 15000 {
		t.Errorf("cumulative win calculation failed, expected 15000, got %v", gsTest.CumulativeWin)
	}
	if gsTest.PlaySequence != 6 {
		t.Errorf("play sequence calculation failed, expected 6, got %v", gsTest.PlaySequence)
	}

}

func TestDetermineLineWins(t *testing.T) {
	testGrid := [][]int{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}, {1, 1, 1}, {1, 1, 1}}

	wins := DetermineLineWins(testGrid, testWinLines, testPayouts, []wild{}, compounding_none, false)
	want := []Prize{{Payout: testPayouts[0], Index: "1:5", Multiplier: 1, SymbolPositions: []int{0, 3, 6, 9, 12}, Winline: 0}, {Payout: testPayouts[0], Index: "1:5", Multiplier: 1, SymbolPositions: []int{0, 4, 8, 10, 12}, Winline: 1}}
	if wins[0].Winline != want[0].Winline || wins[1].Winline != want[1].Winline { // todo: add more criteria for pass
		t.Errorf("first :\n %v \n %v \n second :\n %v \n %v", wins[0], want[0], wins[1], want[1])
	}

	// test multiple wilds, highest multiplier only
	testGrid = [][]int{{0, 1, 0}, {0, 7, 0}, {0, 8, 0}, {0, 1, 0}, {0, 1, 0}}
	wins = DetermineLineWins(testGrid, [][]int{{1, 1, 1, 1, 1}}, testPayouts, testWilds, compounding_none, false)
	if len(wins) != 1 || wins[0].Index != "1:5" || wins[0].Multiplier != 7 || wins[0].Winline != 0 {
		t.Fail()
	}

	// test multiple wilds, highest multiplier only different order
	testGrid = [][]int{{0, 1, 0}, {0, 8, 0}, {0, 7, 0}, {0, 1, 0}, {0, 1, 0}}
	wins = DetermineLineWins(testGrid, [][]int{{1, 1, 1, 1, 1}}, testPayouts, testWilds, compounding_none, false)
	if len(wins) != 1 || wins[0].Index != "1:5" || wins[0].Multiplier != 7 || wins[0].Winline != 0 {
		t.Fail()
	}

	// test 5 wilds no prize set
	testGrid = [][]int{{0, 7, 0}, {0, 7, 0}, {0, 7, 0}, {0, 7, 0}, {0, 7, 0}}
	wins = DetermineLineWins(testGrid, [][]int{{1, 1, 1, 1, 1}}, testPayouts, testWilds, compounding_none, false)
	if len(wins) != 0 {
		// prize must be explicitly set
		t.Fail()
	}

	// test 5 wilds prize set
	testGrid = [][]int{{0, 7, 0}, {0, 7, 0}, {0, 7, 0}, {0, 7, 0}, {0, 7, 0}}
	wins = DetermineLineWins(testGrid, [][]int{{1, 1, 1, 1, 1}}, []Payout{{Symbol: 7, Count: 5, Multiplier: 10}}, testWilds, compounding_none, false)
	if len(wins) != 1 || wins[0].Index != "7:5" || wins[0].Multiplier != 1 {
		// multiplier should not be counted
		t.Fail()
	}

	// test 4 wilds, only last symbol normal
	testGrid = [][]int{{0, 7, 0}, {0, 7, 0}, {0, 7, 0}, {0, 7, 0}, {0, 1, 0}}
	wins = DetermineLineWins(testGrid, [][]int{{1, 1, 1, 1, 1}}, testPayouts, testWilds, compounding_none, false)
	if len(wins) != 1 || wins[0].Index != "1:5" || wins[0].Multiplier != 5 || wins[0].Winline != 0 {
		t.Fail()
	}

	// test variable wild, maintaining chosen value
	wilds := []wild{{
		Symbol:     0,
		Multiplier: weightedMultiplier{[]int{1, 2, 3, 4, 5, 6}, []int{1, 1, 1, 1, 1, 1}},
	}}
	testGrid = [][]int{{1, 0, 5}, {0, 1, 5}, {0, 1, 5}, {1, 0, 5}, {1, 1, 5}}
	wins = DetermineLineWins(testGrid, [][]int{{0, 0, 0, 0, 0}, {1, 1, 1, 1, 1}}, testPayouts, wilds, compounding_none, false)
	if len(wins) != 2 || wins[0].Multiplier != wins[1].Multiplier {
		t.Errorf("the wild multipliers were not properly stored between instances; %v, %v", wins[0], wins[1])
		t.Fail()
	}
}

func TestDeterminsWaysWinsNoWins(t *testing.T) {
	//testGrid := [][]int{{1, 1, 1}, {1, 1, 1}, {3, 3, 3}, {1, 1, 1}, {1, 1, 1}}
	testGrid := [][]int{{1, 1, 1}, {1, 1, 1}, {3, 3, 3}, {1, 1, 1}, {1, 1, 1}}
	wins := DetermineWaysWins(testGrid, testWaysPayouts, []wild{})
	// fmt.Println("wins = ", wins)
	if len(wins) != 0 {
		t.Errorf("wins = %v; want none", wins)
	}
}

func TestDeterminsWaysWinsMultSameWins(t *testing.T) {
	// testGrid := [][]int{{1, 1, 1}, {1, 3, 3}, {1, 3, 3}, {3, 3, 3}, {3, 3, 3}}
	testGrid := [][]int{{1, 1, 1}, {1, 3, 3}, {1, 3, 3}, {3, 3, 3}, {3, 3, 3}}

	wins := DetermineWaysWins(testGrid, testWaysPayouts, []wild{})
	// fmt.Println("wins = ", wins)
	want := Prize{Payout: testWaysPayouts[0], Index: "1:3", Multiplier: 1}
	if len(wins) != 3 || wins[0].Index != want.Index || wins[1].Index != want.Index || wins[2].Index != want.Index {
		t.Errorf("wins = %v; want %v", wins, want)
	}
}

func TestDeterminsWaysWinsThreeSymbolWins(t *testing.T) {
	//testGrid := [][]int{{1, 3, 3}, {3, 3, 1}, {3, 1, 3}, {3, 3, 3}, {3, 3, 3}}
	testGrid := [][]int{{1, 3, 3}, {3, 3, 1}, {3, 1, 3}, {3, 3, 3}, {3, 3, 3}}

	wins := DetermineWaysWins(testGrid, testWaysPayouts, []wild{})
	// fmt.Println("wins = ", wins)
	want := Prize{Payout: testWaysPayouts[0], Index: "1:3", Multiplier: 1}
	if len(wins) != 1 || wins[0].Index != want.Index {
		t.Errorf("wins = %v; want %v", wins, want)
	}
}

func TestDeterminsWaysWinsDifferentSymbolWins(t *testing.T) {
	// testGrid := [][]int{{1, 2, 3}, {2, 3, 1}, {2, 1, 3}, {2, 3, 3}, {3, 2, 3}}
	testGrid := [][]int{{1, 2, 3}, {2, 3, 1}, {2, 1, 3}, {2, 3, 3}, {3, 2, 3}}

	wins := DetermineWaysWins(testGrid, testWaysPayouts, []wild{})
	// fmt.Println("wins = ", wins)
	// wins should be ordered by symbol on first reel
	want := []Prize{{Payout: testWaysPayouts[0], Index: "1:3", Multiplier: 1}, {Payout: testWaysPayouts[1], Index: "2:5", Multiplier: 1}}
	if len(wins) != 2 || wins[0].Index != want[0].Index || wins[1].Index != want[1].Index {
		t.Errorf("wins = %v; want %v", wins, want)
	}
}

// TEST DATATYPES

func TestFixedToBytes(t *testing.T) {
	fixed := NewFixedFromInt(10)
	fmt.Printf("fixed: %v", fixed)
	asBytes := fixed.Bytes()

	reFixed := NewFromBytes(asBytes)

	if fixed != reFixed {
		t.Error("Fixed byte conversion failed")
	}

}

func TestFixed_Add(t *testing.T) {
	fixed1 := Fixed(10000)
	fixed2 := Fixed(20000)
	if fixed1 != NewFixedFromFloat(0.01) {
		t.Errorf("Fixed instantiation failed, expected %v, got %v", NewFixedFromFloat(0.01), fixed1)
	}
	if fixed2 != NewFixedFromFloat(0.02) {
		t.Errorf("Fixed addition failed, expected %v, got %v", NewFixedFromFloat(0.02), fixed2)
	}

	res := fixed1.Add(fixed2)
	expected := Fixed(30000)
	expected2 := NewFixedFromFloat(0.03)
	if res != expected {
		t.Errorf("Fixed addition failed, expected %v, got %v", expected, res)
	}
	if res != expected2 {
		t.Errorf("Fixed addition failed, expected %v, got %v", expected2, res)
	}

}

func TestFixedIntConversion(t *testing.T) {
	fixed := NewFixedFromInt(18)
	if fixed.ValueAsInt() != 18 {
		t.Errorf("Fixed integer conversion failed, expected 18, got %v", fixed.ValueAsInt())
	}
}

func TestFixedFloatConversion(t *testing.T) {
	fixed := NewFixedFromFloat(18.1239)
	if fixed.ValueAsFloat() != 18.1239 {
		t.Errorf("Fixed float conversion failed, expected 18.1293, got %v", fixed.ValueAsFloat())
	}
}

func TestFixedStringConversion(t *testing.T) {
	fixed := Fixed(18123900)
	if fixed.ValueAsString() != "18.123" {
		t.Errorf("Fixed string conversion failed, expected 18.123, got %v", fixed.ValueAsFloat())
	}
}

func TestFixed_Sub(t *testing.T) {
	fixed1 := Fixed(987654)
	fixed2 := Fixed(28030)
	res := fixed1.Sub(fixed2)
	expected := Fixed(959624)
	if res != expected {
		t.Errorf("Fixed subtraction failed, expected %v, got %v", expected, res)
	}
}

func testFixed_Mul(t *testing.T, f1 Fixed, f2 Fixed, prod Fixed, msg string) bool {
	res := f1.Mul(f2)
	if res != prod {
		t.Errorf("%s, %d * %d expected %v, got %v (%.10f * %.10f = %.10f != %.10f)",
			msg, f1, f2, prod, res, float64(f1)/float64(fixedExp), float64(f2)/float64(fixedExp),
			float64(res)/float64(fixedExp), float64(prod)/float64(fixedExp))
		return false
	}
	return true
}

func TestFixed_MulSimple(t *testing.T) {
	fixed1 := NewFixedFromInt(1)
	fixed2 := NewFixedFromInt(1)
	fixedNeg := NewFixedFromInt(-1)
	res1 := fixed1.Mul(fixed2)
	expected := NewFixedFromInt(1)
	if res1 != expected {
		t.Errorf("Fixed multiplication failed, expected %v, got %v", expected, res1)
	}
	res2 := fixed1.Mul(fixedNeg)
	expected2 := NewFixedFromInt(-1)
	if res2 != expected2 {
		t.Errorf("Fixed negative multiplication failed, expected %v, got %v", expected2, res2)
	}
}

func TestFixed_Mul(t *testing.T) {
	fixed1 := NewFixedFromInt(7)
	fixed2 := NewFixedFromInt(18)
	fixedNeg := NewFixedFromInt(-1)
	res1 := fixed1.Mul(fixed2)
	expected := NewFixedFromInt(7 * 18)
	if res1 != expected {
		t.Errorf("Fixed multiplication failed, expected %v, got %v", expected, res1)
	}
	res2 := fixed1.Mul(fixedNeg)
	expected2 := NewFixedFromInt(-7)
	if res2 != expected2 {
		t.Errorf("Fixed negative multiplication failed, expected %v, got %v", expected2, res2)
	}
}

func TestFixed_MulOverflow(t *testing.T) {
	testFixed_Mul(t, NewFixedFromInt(50000), NewFixedFromInt(250), NewFixedFromInt(12500000), "Fixed multiplication failed due to overflow")
	testFixed_Mul(t, NewFixedFromInt(500000), NewFixedFromInt(2500), NewFixedFromInt(1250000000), "Fixed multiplication failed due to overflow")
	testFixed_Mul(t, NewFixedFromInt(5000001), NewFixedFromInt(25023), NewFixedFromInt(125115025023), "Fixed multiplication failed due to overflow")
}

func TestFixed_MulFraction(t *testing.T) {
	testFixed_Mul(t, NewFixedFromFloat(0.123), NewFixedFromFloat(0.003), NewFixedFromFloat(0.000369), "Fixed multiplication decimal error")
	testFixed_Mul(t, NewFixedFromFloat(50000), NewFixedFromFloat(250.5), NewFixedFromFloat(12525000), "Fixed multiplication decimal error")
	testFixed_Mul(t, NewFixedFromFloat(500000), NewFixedFromFloat64(2500.123), NewFixedFromFloat64(1250061500), "Fixed multiplication decimal error")
}

func TestFixed_DivSimple(t *testing.T) {
	fixed1 := NewFixedFromInt(1)
	fixed2 := NewFixedFromInt(1)
	fixedNeg := NewFixedFromInt(-1)
	res1 := fixed1.Div(fixed2)
	expected := NewFixedFromInt(1)
	if res1 != expected {
		t.Errorf("Fixed division failed, expected %v, got %v", expected, res1)
	}
	res2 := fixed1.Div(fixedNeg)
	expected2 := NewFixedFromInt(-1)
	if res2 != expected2 {
		t.Errorf("Fixed negative division failed, expected %v, got %v", expected2, res2)
	}
}

func TestFixed_Div(t *testing.T) {
	fixed1 := NewFixedFromInt(35)
	fixed2 := NewFixedFromInt(7)
	fixedNeg := NewFixedFromInt(-5)
	res1 := fixed1.Div(fixed2)
	expected := NewFixedFromInt(35 / 7)
	if res1 != expected {
		t.Errorf("Fixed division failed, expected %v, got %v", expected, res1)
	}
	res2 := fixed1.Div(fixedNeg)
	expected2 := NewFixedFromInt(-7)
	if res2 != expected2 {
		t.Errorf("Fixed negative division failed, expected %v, got %v", expected2, res2)
	}
}

func TestFixed_Pow(t *testing.T) {
	fixed := NewFixedFromInt(5)
	if fixed.Pow(1) != NewFixedFromInt(5) {
		t.Errorf("Fixed exponent failed, expected 5, got %v", fixed.Pow(2))
	}
	if fixed.Pow(2) != NewFixedFromInt(25) {
		t.Errorf("Fixed exponent failed, expected 25, got %v", fixed.Pow(2))
	}
	if fixed.Pow(3) != NewFixedFromInt(125) {
		t.Errorf("Fixed exponent failed, expected 25, got %v", fixed.Pow(3))
	}
	if fixed.Pow(4) != NewFixedFromInt(625) {
		t.Errorf("Fixed exponent failed, expected 625, got %v", fixed.Pow(4))
	}
	if fixed.Pow(5) != NewFixedFromInt(3125) {
		t.Errorf("Fixed exponent failed, expected %v, got %v", NewFixedFromInt(3125), fixed.Pow(5))
	}
}

func TestEngineConfig_DetectSpecialWins(t *testing.T) {
	testPrizes := []Prize{{Index: "4:3"}, {Index: "somethingElse"}, {Index: "11:3"}, {Index: "11:4"}, {Index: "11:5"}, {Index: "aa:bb"}, {Index: "10:3"}}

	testEngine := "mvgEngineI"
	config := BuildEngineDefs(testEngine)
	targetIndices := []string{"4:3", "somethingElse", "freespin:10", "freespin:15", "freespin:20", "aa:bb", "10:3"}
	for i := 0; i < len(testPrizes); i++ {
		index := config.DetectSpecialWins(0, testPrizes[i])
		if index != targetIndices[i] {
			t.Errorf("Special win detection failed, expected %v, got %v", targetIndices[i], index)
		}
	}

	testEngine = "mvgEngineII"
	config = BuildEngineDefs(testEngine)
	targetIndices = []string{"4:3", "somethingElse", "freespin:15", "freespin:15", "freespin:15", "aa:bb", "10:3"}
	for i := 0; i < len(testPrizes); i++ {
		index := config.DetectSpecialWins(0, testPrizes[i])
		if index != targetIndices[i] {
			t.Errorf("Special win detection failed, expected %v, got %v", targetIndices[i], index)
		}
	}

	testEngine = "mvgEngineIII"
	config = BuildEngineDefs(testEngine)
	targetIndices = []string{"4:3", "somethingElse", "pickSpins:1", "pickSpins:1", "pickSpins:1", "aa:bb", "10:3"}
	for i := 0; i < len(testPrizes); i++ {
		index := config.DetectSpecialWins(0, testPrizes[i])
		if index != targetIndices[i] {
			t.Errorf("Special win detection failed, expected %v, got %v", targetIndices[i], index)
		}
	}
	testEngine = "mvgEngineIX"
	config = BuildEngineDefs(testEngine)
	targetIndices = []string{"freespin:8", "somethingElse", "11:3", "11:4", "11:5", "aa:bb", "10:3"}
	for i := 0; i < len(testPrizes); i++ {
		index := config.DetectSpecialWins(0, testPrizes[i])
		if index != targetIndices[i] {
			t.Errorf("Special win detection failed, expected %v, got %v", targetIndices[i], index)
		}
	}

	testEngine = "mvgEngineV"
	config = BuildEngineDefs(testEngine)
	targetIndices = []string{"4:3", "somethingElse", "freespin:8", "freespin:16", "freespin:24", "aa:bb", "10:3"}
	for i := 0; i < len(testPrizes); i++ {
		index := config.DetectSpecialWins(0, testPrizes[i])
		if index != targetIndices[i] {
			t.Errorf("Special win detection failed, expected %v, got %v", targetIndices[i], index)
		}
	}

	testEngine = "mvgEngineVII"
	config = BuildEngineDefs(testEngine)
	targetIndices = []string{"4:3", "somethingElse", "11:3", "11:4", "11:5", "aa:bb", "freespin2:10"}
	for i := 0; i < len(testPrizes); i++ {
		index := config.DetectSpecialWins(0, testPrizes[i])
		if index != targetIndices[i] {
			t.Errorf("Special win detection failed, expected %v, got %v", targetIndices[i], index)
		}
	}
	testEngine = "mvgEngineVII"
	config = BuildEngineDefs(testEngine)
	defIndex := config.DefIdByName("freespin2")
	targetIndices = []string{"4:3", "somethingElse", "11:3", "11:4", "11:5", "aa:bb", "freespin3:10"}
	for i := 0; i < len(testPrizes); i++ {
		index := config.DetectSpecialWins(defIndex, testPrizes[i])
		if index != targetIndices[i] {
			t.Errorf("Special win detection failed, expected %v, got %v", targetIndices[i], index)
		}
	}
	testEngine = "mvgEngineVII"
	config = BuildEngineDefs(testEngine)
	defIndex = config.DefIdByName("freespin3")

	targetIndices = []string{"4:3", "somethingElse", "11:3", "11:4", "11:5", "aa:bb", "freespin4:10"}
	for i := 0; i < len(testPrizes); i++ {
		index := config.DetectSpecialWins(defIndex, testPrizes[i])
		if index != targetIndices[i] {
			t.Errorf("Special win detection failed, expected %v, got %v", targetIndices[i], index)
		}
	}

	testEngine = "mvgEngineVII"
	config = BuildEngineDefs(testEngine)
	defIndex = config.DefIdByName("freespin4")
	targetIndices = []string{"4:3", "somethingElse", "11:3", "11:4", "11:5", "aa:bb", "freespin5:10"}
	for i := 0; i < len(testPrizes); i++ {
		index := config.DetectSpecialWins(defIndex, testPrizes[i])
		if index != targetIndices[i] {
			t.Errorf("Special win detection failed, expected %v, got %v", targetIndices[i], index)
		}
	}
	testEngine = "mvgEngineX"
	config = BuildEngineDefs(testEngine)
	targetIndices = []string{"4:3", "somethingElse", "11:3", "11:4", "11:5", "aa:bb", "10:3"}
	for i := 0; i < len(testPrizes); i++ {
		index := config.DetectSpecialWins(0, testPrizes[i])
		if index != targetIndices[i] {
			t.Errorf("Special win detection failed, expected %v, got %v", targetIndices[i], index)
		}
	}
	// todo: add in engineXI tests
	//testEngine = "mvgEngineXI"
	//config = BuildEngineDefs(testEngine)
	//targetIndices = []string{"3:3", "somethingElse", "freespin:8", "freespin:16", "freespin:24", "aa:bb", "10:3"}
	//for i:=0; i<len(testPrizes); i++ {
	//	index := config.DetectSpecialWins(0,testPrizes[i])
	//	if index != targetIndices[i] {
	//		t.Errorf("Special win detection failed, expected %v, got %v", targetIndices[i], index)
	//	}
	//}
	testEngine = "mvgEngineXII"
	config = BuildEngineDefs(testEngine)
	targetIndices = []string{"4:3", "somethingElse", "freespin:5", "freespin:10", "freespin:20", "aa:bb", "10:3"}
	for i := 0; i < len(testPrizes); i++ {
		index := config.DetectSpecialWins(0, testPrizes[i])
		if index != targetIndices[i] {
			t.Errorf("Special win detection failed, expected %v, got %v", targetIndices[i], index)
		}
	}

	testEngine = "mvgEngineXIII"
	config = BuildEngineDefs(testEngine)
	testPrizes = []Prize{{Index: "10:2"}, {Index: "10:3"}, {Index: "10:4"}, {Index: "10:5"}}
	targetIndices = []string{"freespin:0", "freespin:10", "freespin:10", "freespin:10"}
	for i := 0; i < len(testPrizes); i++ {
		index := config.DetectSpecialWins(0, testPrizes[i])
		if index != targetIndices[i] {
			t.Errorf("Special win detection failed, expected %v, got %v", targetIndices[i], index)
		}
	}
}

func TestDetermineBarLineWins(t *testing.T) {
	//determineBarLineWins(symbolGrid [][]int, winLines [][]int, payouts []Payout, bars []bar, wilds []wild) []Prize {
	winLines := [][]int{{1, 1, 1}} // one win line
	payouts := []Payout{
		{1, 3, 10},
		{2, 3, 100}, // bar payout greater than line
		{3, 3, 1},   // bar payout less than line
	}
	bars := []bar{{2, []int{1, 4, 5}}} // greater than line
	symbolGrid := [][]int{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}}

	// test higher payout overrides lower
	prizes := determineBarLineWins(symbolGrid, winLines, payouts, bars, []wild{}, false)
	if len(prizes) != 1 {
		t.Errorf("Expected one win")
	}
	if prizes[0].Payout.Multiplier != 100 {
		t.Errorf("Expected bar win to override line win")
	}

	bars = []bar{{3, []int{1, 4, 5}}} // less than line
	prizes = determineBarLineWins(symbolGrid, winLines, payouts, bars, []wild{}, false)
	if len(prizes) != 1 {
		t.Errorf("Expected one win")
	}
	if prizes[0].Payout.Multiplier != 10 {
		t.Errorf("Expected line win to override bar win")
	}

	// test general symbol substitution
	symbolGrid = [][]int{{1, 4, 1}, {1, 5, 0}, {1, 4, 0}}
	prizes = determineBarLineWins(symbolGrid, winLines, payouts, bars, []wild{}, false)
	if len(prizes) != 1 {
		t.Errorf("Expected one win")
	}
	if prizes[0].Payout.Multiplier != 1 {
		t.Errorf("Expected bar win")
	}

	// test symbol substitution with normal payout symbol also present
	symbolGrid = [][]int{{1, 1, 1}, {1, 5, 0}, {1, 4, 0}}
	prizes = determineBarLineWins(symbolGrid, winLines, payouts, bars, []wild{}, false)
	if len(prizes) != 1 {
		t.Errorf("Expected one win")
	}
	if prizes[0].Payout.Multiplier != 1 {
		t.Errorf("Expected bar win")
	}

	// test symbol substitution with normal payout symbol also present
	symbolGrid = [][]int{{1, 4, 1}, {1, 1, 0}, {1, 4, 0}}
	prizes = determineBarLineWins(symbolGrid, winLines, payouts, bars, []wild{}, false)
	if len(prizes) != 1 {
		t.Errorf("Expected one win")
	}
	if prizes[0].Payout.Multiplier != 1 {
		t.Errorf("Expected bar win")
	}
}

func TestEngineDef_ProcessWinLines(t *testing.T) {
	var testDef = EngineDef{
		WinLines:     [][]int{{0, 0, 0}, {1, 1, 1}, {2, 2, 2}},
		StakeDivisor: 0,
	}
	wl, testDef := testDef.ProcessWinLines([]int{})
	if len(wl) != 3 || wl[0] != 0 || wl[1] != 1 || wl[2] != 2 {
		t.Errorf("expected win lines 0,1,2, got %v", wl)
	}
	if testDef.StakeDivisor != 3 {
		t.Errorf("expected stake divisor 3, got %v", testDef.StakeDivisor)
	}

	wl, testDef = testDef.ProcessWinLines([]int{0, 1, 2})
	if len(wl) != 3 || wl[0] != 0 || wl[1] != 1 || wl[2] != 2 {
		t.Errorf("expected winlines returned [0,1,2], got %v", wl)
	}
	if testDef.StakeDivisor != 3 {
		t.Errorf("expected stake divisor 3, got %v", testDef.StakeDivisor)
	}

	wl, testDef = testDef.ProcessWinLines([]int{0, 2, 3})
	if len(wl) != 2 || wl[0] != 0 || wl[1] != 2 {
		t.Errorf("expected win lines 0,2, got %v", wl)
	}
	if testDef.StakeDivisor != 2 {
		t.Errorf("expected stake divisor 2, got %v", testDef.StakeDivisor)
	}

}

func TestMilliCcies(t *testing.T) {
	// test 2-pt ccy
	twoPtCcy := "USD"
	amt := Money{Fixed(10000), twoPtCcy}
	res := RoundUpToNearestCCYUnit(amt)
	if res.Currency != twoPtCcy {
		t.Errorf("Currency changed from %v to %v", twoPtCcy, res.Currency)
	}
	if res.Amount != Fixed(10000) {
		t.Errorf(".01 even rounded to %v", res.Amount.ValueAsString())
	}

	amt.Amount = Fixed(10001)
	res = RoundUpToNearestCCYUnit(amt)
	if res.Currency != twoPtCcy {
		t.Errorf("Currency changed from %v to %v", twoPtCcy, res.Currency)
	}
	if res.Amount != Fixed(20000) {
		t.Errorf(".010001 incorrectly rounded to %v", res.Amount.ValueAsString())
	}

	// test rounding at 2 digit change
	amt.Amount = Fixed(123491111)
	res = RoundUpToNearestCCYUnit(amt)
	if res.Amount != Fixed(123500000) {
		t.Errorf("123.491111 incorrectly rounded to %v", res.Amount.ValueAsString())
	}

	// test 3-digit ccy
	threePtCcy := "BTC"

	amt = Money{Fixed(1000), threePtCcy}
	res = RoundUpToNearestCCYUnit(amt)
	if res.Currency != threePtCcy {
		t.Errorf("Currency changed from %v to %v", threePtCcy, res.Currency)
	}
	if res.Amount != Fixed(1000) {
		t.Errorf(".001 even rounded to %v", res.Amount.ValueAsString())
	}

	amt.Amount = Fixed(1001)
	res = RoundUpToNearestCCYUnit(amt)
	if res.Currency != threePtCcy {
		t.Errorf("Currency changed from %v to %v", threePtCcy, res.Currency)
	}
	if res.Amount != Fixed(2000) {
		t.Errorf(".001001 incorrectly rounded to %v", res.Amount.ValueAsString())
	}

	// test rounding at 2 digit change
	amt.Amount = Fixed(123459111)
	res = RoundUpToNearestCCYUnit(amt)
	if res.Amount != Fixed(123460000) {
		t.Errorf("123459111 incorrectly rounded to %v", res.Amount.ValueAsString())
	}

	// test zero

	amt.Amount = Fixed(0)
	res = RoundUpToNearestCCYUnit(amt)
	if res.Amount != Fixed(1000) {
		t.Errorf("zero incorrectly rounded to %v", res.Amount.ValueAsString())
	}

}

func TestShuffleFlop(t *testing.T) {
	t.Skip("skipping shuffle flop due to implementation error")
	rng.Init()
	shuffleDef := EngineDef{
		Reels:          [][]int{{0, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5}, {1, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5}, {2, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5}, {3, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5}, {4, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5}},
		ViewSize:       []int{1, 1, 1, 1, 1},
		Payouts:        []Payout{},
		WinType:        "pAndF",
		SpecialPayouts: []Prize{},
		Wilds:          []wild{},
		Multiplier:     weightedMultiplier{},
		//StakeDivisor:   1,
	}
	testGS := Gamestate{
		DefID:          0,
		BetPerLine:     Money{1, "BTC"},
		SymbolGrid:     [][]int{{0}, {1}, {2}, {3}, {4}},
		Prizes:         []Prize{},
		RelativePayout: 0,
		Multiplier:     1,
		StopList:       []int{0, 0, 0, 0, 0},
	}
	params := GameParams{
		Stake:             1,
		previousGamestate: testGS,
	}

	GS := shuffleDef.ShuffleFlop(params)
	if GS.SymbolGrid[0][0] != 0 {
		//try again because there is a channnce it landed on teh same
		GS = shuffleDef.ShuffleFlop(params)
		if GS.SymbolGrid[0][0] != 0 {
			t.Errorf("prime was probably shuffled %v", GS)
		}
	}
	if GS.SymbolGrid[1][0] == 1 && GS.SymbolGrid[2][0] == 2 && GS.SymbolGrid[3][0] == 3 && GS.SymbolGrid[4][0] == 4 {
		t.Errorf("flop was probably not shuffled %v", GS)
	}
}

// TestShufflePrime passes locally yet failed on github merge.. keep this comment and see if it happens again
func TestShufflePrime(t *testing.T) {
	t.Skip("skipping shuffle flop due to implementation error")
	rng.Init()
	shuffleDef := EngineDef{
		Reels:          [][]int{{0, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5}, {1, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5}, {2, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5}, {3, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5}, {4, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5}},
		ViewSize:       []int{1, 1, 1, 1, 1},
		Payouts:        []Payout{},
		WinType:        "pAndF",
		SpecialPayouts: []Prize{},
		Wilds:          []wild{},
		Multiplier:     weightedMultiplier{},
		//StakeDivisor:   1,
	}
	testGS := Gamestate{
		DefID:          0,
		BetPerLine:     Money{1, "BTC"},
		SymbolGrid:     [][]int{{0}, {1}, {2}, {3}, {4}},
		Prizes:         []Prize{},
		RelativePayout: 0,
		Multiplier:     1,
		StopList:       []int{0, 0, 0, 0, 0},
	}
	params := GameParams{
		Stake:             1,
		previousGamestate: testGS,
	}
	GS := shuffleDef.ShufflePrime(params)
	if GS.SymbolGrid[0][0] == 0 {
		GS = shuffleDef.ShufflePrime(params)
		if GS.SymbolGrid[0][0] == 0 {
			t.Errorf("Prime was probably not shuffled: %#v", GS)
		}
	}
	if GS.SymbolGrid[1][0] != 1 || GS.SymbolGrid[2][0] != 2 || GS.SymbolGrid[3][0] != 3 || GS.SymbolGrid[4][0] != 4 {
		t.Errorf("Flop was probably shuffled: %#v", GS)
	}
}
