package engine

import (
	"encoding/json"
	"fmt"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"gopkg.in/yaml.v3"
	"strconv"
	"strings"
)

// Datatypes
type weightedMultiplier struct {
	Multipliers   []int `yaml:"multipliers,flow"`
	Probabilities []int `yaml:"probabilities,flow"`
}

// wilds
type wild struct {
	Symbol     int                `yaml:"symbol"`
	Multiplier weightedMultiplier `yaml:"multiplier"`
}

// bar symbols
type bar struct {
	PayoutID int   `yaml:"payoutId"`
	Symbols  []int `yaml:"symbols"`
}

// EngineDef ...
type EngineDef struct {
	ID       string `yaml:"name"`
	Index    int
	Function string   `yaml:"function"`                // determines what funciton is run with this enginedef
	Reels    [][]int  `yaml:"Reels,flow" json:"reels"` // reel contents
	ViewSize []int    `yaml:"ViewSize"`                // the number and shape of symbols to display in the client
	Payouts  []Payout `yaml:"Payouts"`                 // the payouts for line wins (can be nil for ways games)
	WinType  string   `yaml:"WinType"`                 // ways, lines, or barLines (specifying lines insteadof barLines saves comp. power)
	// The string represents the method to be run. should be ordered by precedence
	SpecialPayouts []Prize            `yaml:"SpecialPayouts"`
	WinLines       [][]int            `yaml:"WinLines,flow"`
	Wilds          []wild             `yaml:"wilds"`
	Bars           []bar              `yaml:"bars"`
	Multiplier     weightedMultiplier `yaml:"multiplier"`
	StakeDivisor   int                `yaml:"StakeDivisor"`
	Probability    int                `yaml:"Probability"`    // the probability of this engine being selected if it shares id with other engines
	ExpectedPayout Fixed              `yaml:"expectedPayout"` // the expected payout of one round of this engineDef
	RTP            float32            `yaml:"RTP"`            // the expected payout of one round of this engineDef
}

type Fixed int64

const fixedExp Fixed = 1000000

// A StakeFloat is a string that can be unmarshalled from a JSON field
// that has either a number or a string value.
// E.g. if the json field contains an string "42", the
// StakeFloat value will be "42".
//type StakeFloat engine.Fixed

// UnmarshalJSON implements the json.Unmarshaler interface, which
// allows us to ingest values of any json type as a string and run our custom conversion

func (f *Fixed) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		*f = Fixed(0) // hardcode empty string to be zero stake
		return nil
	}
	var s float32
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*f = NewFixedFromFloat(s)
	return nil
}

func (f *Fixed) MarshalJSON() ([]byte, error) {
	s := f.ValueAsFloat()
	return json.Marshal(&s)
}

// UnmarshalJSON implements the yaml.Unmarshaler interface, which
// allows us to ingest values of any yaml type as a string and run our custom conversion

func (f *Fixed) UnmarshalYAML(b []byte) error {
	if b[0] == '"' {
		*f = Fixed(0) // hardcode empty string to be zero stake
		return nil
	}
	var s float32
	if err := yaml.Unmarshal(b, &s); err != nil {
		return err
	}
	*f = NewFixedFromFloat(s)
	return nil
}

func (f *Fixed) MarshalYAML() ([]byte, error) {
	s := f.ValueAsFloat()
	return yaml.Marshal(&s)
}

type Money struct {
	Amount   Fixed  `json:"amount"`
	Currency string `json:"currency"`
}

type Payout struct {
	Symbol     int `json:"symbol,omitempty" yaml:"Symbol"`
	Count      int `json:"count,omitempty" yaml:"Count"`
	Multiplier int `json:"multiplier,omitempty" yaml:"Multiplier"`
}

type Prize struct {
	Payout          Payout `json:"payout,omitempty" yaml:"Payout,flow"`
	Index           string `json:"index,omitempty" yaml:"Index"`
	Multiplier      int    `json:"multiplier,omitempty" yaml:"Multiplier"`
	SymbolPositions []int  `json:"symbol_positions,omitempty" yaml:"SymbolPositions"`
	Winline         int    `json:"winline,omitempty"`
	Win             Fixed  `json:"win,omitempty"`
}

type Gamestate struct {
	// internal representation of GamestatePB
	Id                string                    `json:"id,omitempty"`
	GameID            string                    `json:"engine,omitempty"`
	BetPerLine        Money                     `json:"bet_per_line,omitempty"`
	Transactions      []WalletTransaction       `json:"transactions,omitempty"`
	PreviousGamestate string                    `json:"previous_gamestate,omitempty"`
	NextGamestate     string                    `json:"next_gamestate,omitempty"`
	Action            string                    `json:"action,omitempty"`
	SymbolGrid        [][]int                   `json:"symbol_grid,omitempty"`
	Prizes            []Prize                   `json:"prizes,omitempty"`
	SelectedWinLines  []int                     `json:"selected_win_lines,omitempty"`
	RelativePayout    int                       `json:"relative_payout,omitempty"`
	Multiplier        int                       `json:"multiplier,omitempty"`
	StopList          []int                     `json:"stop_list,omitempty"`
	NextActions       []string                  `json:"next_actions,omitempty"`
	Gamification      *GamestatePB_Gamification `json:"gamification,omitempty"`
	CumulativeWin	  Fixed						`json:"cumulative_win,omitempty"`
	PlaySequence	  int						`json:"play_sequence,omitempty"`
}

func (gamestate Gamestate) Engine() string {
	engine, err := config.GetEngineFromGame(strings.Split(gamestate.GameID, ":")[0])
	if err != nil {
		logger.Errorf("error parsing game name: %v", gamestate.GameID)
	}
	return engine
}

type WalletTransaction struct {
	// internal representation of WalletTransactionPB
	Id     string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Amount Money  `protobuf:"bytes,2,opt,name=amount,proto3" json:"amount,omitempty"`
	Type   string `protobuf:"bytes,3,opt,name=type,proto3" json:"type,omitempty"`
}

func (transactionPB WalletTransactionPB) Convert() WalletTransaction {
	return WalletTransaction{
		Id:     transactionPB.Id,
		Amount: Money{Amount: Fixed(transactionPB.Amount), Currency: transactionPB.Currency.String()},
		Type:   transactionPB.Type.String(),
	}
}

func (transaction WalletTransaction) Convert() *WalletTransactionPB {
	return &WalletTransactionPB{
		Id:       transaction.Id,
		Amount:   transaction.Amount.Amount.ValueRaw(),
		Currency: Ccy(Ccy_value[transaction.Amount.Currency]),
		Type:     WalletTransactionPB_Type(WalletTransactionPB_Type_value[transaction.Type]),
	}
}

func convertSymbolGridFromPB(symbolGrid []*GamestatePB_Reel) [][]int {
	converted := make([][]int, len(symbolGrid))
	for i, reel := range symbolGrid {
		convertedReel := make([]int, len(reel.Symbols))
		for j, symbol := range reel.Symbols {
			convertedReel[j] = int(symbol)
		}
		converted[i] = convertedReel
	}
	return converted
}

func convertSymbolGridToPB(symbolGrid [][]int) []*GamestatePB_Reel {
	converted := make([]*GamestatePB_Reel, len(symbolGrid))
	for i, reel := range symbolGrid {
		convertedReel := make([]int32, len(reel))
		for j, symbol := range reel {
			convertedReel[j] = int32(symbol)
		}
		gsReel := GamestatePB_Reel{Symbols: convertedReel}
		converted[i] = &gsReel
	}
	return converted
}

func convertTransactionsFromPB(unconverted []*WalletTransactionPB) []WalletTransaction {
	converted := make([]WalletTransaction, len(unconverted))
	for i, transaction := range unconverted {
		converted[i] = transaction.Convert()
	}
	return converted
}
func convertTransactionsToPB(unconverted []WalletTransaction) []*WalletTransactionPB {
	converted := make([]*WalletTransactionPB, len(unconverted))
	for i, transaction := range unconverted {

		converted[i] = transaction.Convert()
	}
	return converted
}

func (payoutPB PayoutPB) Convert() Payout {
	return Payout{
		Symbol:     int(payoutPB.Symbol),
		Count:      int(payoutPB.Count),
		Multiplier: int(payoutPB.Multiplier),
	}
}

func (payout Payout) Convert() *PayoutPB {
	return &PayoutPB{
		Symbol:     int32(payout.Symbol),
		Count:      int32(payout.Count),
		Multiplier: int32(payout.Multiplier),
	}
}

func convertInt32Int(in []int32) []int {
	out := make([]int, len(in))
	for i, val := range in {
		out[i] = int(val)
	}
	return out
}

func convertIntInt32(in []int) []int32 {
	out := make([]int32, len(in))
	for i, val := range in {
		out[i] = int32(val)
	}
	return out
}

func (prizePB PrizePB) Convert(betPerLine Fixed) Prize {
	prize := Prize{
		Payout:          prizePB.Payout.Convert(),
		Index:           fmt.Sprintf("%v:%v", prizePB.Payout.Symbol, prizePB.Payout.Count),
		Multiplier:      int(prizePB.Multiplier),
		SymbolPositions: convertInt32Int(prizePB.SymbolPositions),
		Winline:         int(prizePB.Winline),
	}
	prize.Win = NewFixedFromInt(prize.Payout.Multiplier * prize.Multiplier).Mul(betPerLine)
	return prize
}

func (prize Prize) Convert() PrizePB {
	return PrizePB{
		Payout:          prize.Payout.Convert(),
		Multiplier:      int32(prize.Multiplier),
		SymbolPositions: convertIntInt32(prize.SymbolPositions),
		Winline:         int32(prize.Winline),
	}
}

func convertPrizesFromPB(unconverted []*PrizePB, betPerLine Fixed) []Prize {
	converted := make([]Prize, len(unconverted))
	for i, prize := range unconverted {
		converted[i] = prize.Convert(betPerLine)
	}
	return converted
}

func convertPrizesToPB(unconverted []Prize) []*PrizePB {
	converted := make([]*PrizePB, len(unconverted))
	for i, prize := range unconverted {
		prizePB := prize.Convert()
		converted[i] = &prizePB
	}
	return converted
}

func (gamestatePB GamestatePB) Convert(transactions []*WalletTransactionPB) Gamestate {
	nextActions := make([]string, len(gamestatePB.NextActions))
	for i, action := range gamestatePB.NextActions {
		nextActions[i] = action.String()
	}
	// every set of transactions should begin with a WAGER. this ID is also the gamestate ID

	// get Game ID
	return Gamestate{
		Id:                transactions[0].Id,
		GameID:            fmt.Sprintf("%v:%v", GetGameIDFromPB(gamestatePB.GameId.String()), gamestatePB.EngineDef),
		BetPerLine:        Money{Amount: Fixed(gamestatePB.BetPerLine), Currency: gamestatePB.Currency.String()},
		Transactions:      convertTransactionsFromPB(transactions),
		PreviousGamestate: string(gamestatePB.PreviousGamestate),
		NextGamestate:     string(gamestatePB.NextGamestate),
		Action:            gamestatePB.Action.String(),
		SymbolGrid:        convertSymbolGridFromPB(gamestatePB.SymbolGrid),
		Prizes:            convertPrizesFromPB(gamestatePB.Prizes, Fixed(gamestatePB.BetPerLine)),
		SelectedWinLines:  convertInt32Int(gamestatePB.SelectedWinLines),
		RelativePayout:    int(gamestatePB.RelativePayout),
		Multiplier:        int(gamestatePB.Multiplier),
		StopList:          convertInt32Int(gamestatePB.StopList),
		NextActions:       nextActions,
		Gamification:      gamestatePB.Gamification,
		CumulativeWin:	   Fixed(gamestatePB.CumulativeWin),
		PlaySequence: 	   int(gamestatePB.PlaySequence),
	}
}

func GetGameIDAndReelset(gameID string) (string, int) {
	gameInfo := strings.Split(gameID, ":")
	if len(gameInfo) != 2 {
		panic(fmt.Errorf("EngineInfo not correct format: %v", gameInfo))
	}
	engineDef, err := strconv.Atoi(gameInfo[1])
	if err != nil {
		logger.Warnf("Could not get engineDef integer from %v, setting zero", gameInfo)
		engineDef = 0
	}
	//logger.Debugf("Engine: %v; DefID: %v", engineInfo[0], engineDef)
	return gameInfo[0], engineDef
}

func GetEngineDefFromGame(gameID string) (EngineConfig, int, error) {
	gameID, rsID := GetGameIDAndReelset(gameID)
	engineID, err := config.GetEngineFromGame(gameID)
	if err != nil {
		return EngineConfig{}, 0, err
	}
	return BuildEngineDefs(engineID), rsID, nil
}

func (gamestate Gamestate) Convert() (GamestatePB, []*WalletTransactionPB) {
	nextActions := make([]GamestatePB_Action, len(gamestate.NextActions))
	for i, action := range gamestate.NextActions {
		nextActions[i] = GamestatePB_Action(GamestatePB_Action_value[action])
	}
	// every set of transactions should begin with a WAGER. this ID is also the gamestate ID
	gameID, engineDef := GetGameIDAndReelset(gamestate.GameID)
	return GamestatePB{
		GameId:            GamestatePB_GameID(GamestatePB_GameID_value[GetPBFromGameID(gameID)]),
		EngineDef:         int32(engineDef),
		BetPerLine:        gamestate.BetPerLine.Amount.ValueRaw(),
		Currency:          Ccy(Ccy_value[gamestate.BetPerLine.Currency]),
		PreviousGamestate: []byte(gamestate.PreviousGamestate),
		NextGamestate:     []byte(gamestate.NextGamestate),
		Action:            GamestatePB_Action(GamestatePB_Action_value[gamestate.Action]),
		SymbolGrid:        convertSymbolGridToPB(gamestate.SymbolGrid),
		Prizes:            convertPrizesToPB(gamestate.Prizes),
		SelectedWinLines:  convertIntInt32(gamestate.SelectedWinLines),
		RelativePayout:    int32(gamestate.RelativePayout),
		Multiplier:        int32(gamestate.Multiplier),
		StopList:          convertIntInt32(gamestate.StopList),
		NextActions:       nextActions,
		Gamification:      gamestate.Gamification,
		CumulativeWin:     gamestate.CumulativeWin.ValueRaw(),
		PlaySequence: 	   int32(gamestate.PlaySequence),
	}, convertTransactionsToPB(gamestate.Transactions)
}
