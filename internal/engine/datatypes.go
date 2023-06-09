package engine

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/getsentry/sentry-go"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	rgserror "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"gopkg.in/yaml.v3"
)

// Constants

// Datatypes
type weightedMultiplier struct {
	Multipliers   []int `yaml:"multipliers,flow"`
	Probabilities []int `yaml:"probabilities,flow"`
}

// wilds
type wild struct {
	Symbol     int                `yaml:"symbol"`
	Multiplier weightedMultiplier `yaml:"multiplier"`
	Sticky     bool               `yaml:"sticky"`
}

// bar symbols
type bar struct {
	PayoutID int   `yaml:"payoutId"`
	Symbols  []int `yaml:"symbols"`
}

const (
	compounding_none = iota
	compounding_multiplication
	compounding_addition
)

const (
	winconf_none                 = "none"
	winconf_anchor_left          = "anchor_left"
	winconf_anchor_right         = "anchor_right"
	winconf_anchor_left_or_right = "anchor_left_or_right"
	winconf_compounding_addition = "compounding_addition"
)

const (
	inheritance_none  = "none"  // initialize enginedef to zero value
	inheritance_first = "first" // initialize enginedef from first defined def
	inheritance_prev  = "prev"  // initialize enginedef from previously defined def
)

type WinConfiguration struct {
	Flags string `yaml:"Flags"` // concatenation of winconf_ flag strings, if empty string then configuration is ignored(set to none then)
}

type RoulettePayout struct {
	Multiplier int   `yaml:"Multiplier"`
	Symbols    []int `yaml:"Symbols"`
}

// EngineDef ...
type EngineDef struct {
	ID          string `yaml:"name"`
	Inheritance string `yaml:"inheritance"` // inherit for previous def if true, first def if false
	Index       int
	Function    string   `yaml:"function"`                // determines what funciton is run with this enginedef
	ReelsetId   string   `yaml:"ReelsetId"`               // unique identifier for the reels for certification
	Reels       [][]int  `yaml:"Reels,flow" json:"reels"` // reel contents
	ViewSize    []int    `yaml:"ViewSize"`                // the number and shape of symbols to display in the client
	Payouts     []Payout `yaml:"Payouts"`                 // the payouts for line wins (can be nil for ways games)
	WinType     string   `yaml:"WinType"`                 // ways, lines, or barLines (specifying lines insteadof barLines saves comp. power)
	// The string represents the method to be run. should be ordered by precedence
	SpecialPayouts        []Prize                   `yaml:"SpecialPayouts"`
	WinLines              [][]int                   `yaml:"WinLines,flow"`
	WinConfig             WinConfiguration          `yaml:"WinConfig"`
	Wilds                 []wild                    `yaml:"wilds"`
	Bars                  []bar                     `yaml:"bars"`
	Multiplier            weightedMultiplier        `yaml:"multiplier"`
	StakeDivisor          int                       `yaml:"StakeDivisor"`
	Probability           int                       `yaml:"Probability"`      // the probability of this engine being selected if it shares id with other engines
	ExpectedPayout        Fixed                     `yaml:"expectedPayout"`   // the expected payout of one round of this engineDef
	RTP                   float32                   `yaml:"RTP"`              // the expected payout of one round of this engineDef
	RespinAction          string                    `yaml:"RespinAction"`     // the action to add to nextActions
	RespinAllowed         bool                      `yaml:"respin"`           // must be explicitly enabled on each def
	VariableWL            bool                      `yaml:"variableWinLines"` // will be false by default
	Compounding           bool                      `yaml:"compoundingWilds"` // will be false by default
	force                 []int                     // may not be set via yaml
	Features              []feature.FeatureDef      `yaml:"Features"`
	RoulettePayouts       map[string]RoulettePayout `yaml:"RoulettePayouts"`
	NextMultiplierActions []string                  `yaml:"NextMultiplierActions"` // actions that selects the next multiplier, default ["cascade"]
	HoldMultiplierActions []string                  `yaml:"HoldMultiplierActions"` // actions that keeps the current multiplier, default ["freespin"]
	FeatureStages         []string                  `yaml:"FeatureStages"`         // execution stages ("reelupdate")
}

func (engine EngineDef) SetForce(force []int) (forcedengine EngineDef, err rgserror.RGSErr) {
	forcedengine = engine
	if config.GlobalConfig.DevMode == true {
		forcedengine.force = force
		return
	}
	err = rgserror.Create(rgserror.ForceProhibited)
	return
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
	var s float32
	var err error
	if b[0] == '"' {
		var str string
		if err = json.Unmarshal(b, &str); err != nil {
			*f = Fixed(0) // hardcode empty string to be zero stake
			return nil
		}
		var d float64
		if d, err = strconv.ParseFloat(str, 64); err != nil {
			*f = Fixed(0) // hardcode empty string to be zero stake
			return nil
		}
		s = float32(d)
	} else {
		if err = json.Unmarshal(b, &s); err != nil {
			return err
		}
	}
	*f = NewFixedFromFloat(s)
	return nil
}

func (f Fixed) MarshalJSON() ([]byte, error) {
	s := f.ValueAsString()
	return json.Marshal(&s)
}

// UnmarshalJSON implements the yaml.Unmarshaler interface, which
// allows us to ingest values of any yaml type as a string and run our custom conversion

func (f *Fixed) UnmarshalYAML(b []byte) error {
	// todo this function is never called ?
	if b[0] == '"' {
		*f = Fixed(0) // hardcode empty string to be zero stake
		return nil
	}
	var s float32
	if err := yaml.Unmarshal(b, &s); err != nil {
		return err
	}
	logger.Warnf("float value %v", s)
	*f = NewFixedFromFloat(s)
	return nil
}

func (f *Fixed) MarshalYAML() ([]byte, error) {
	s := f.ValueAsString()
	return yaml.Marshal(&s)
}

func (f Fixed) String() string {
	return fmt.Sprintf("%f", f.ValueAsFloat())
}

type Money struct {
	Amount   Fixed  `json:"amount"`
	Currency string `json:"currency"`
}

type Payout struct {
	Symbol     int `json:"symbol" yaml:"Symbol"`
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
	Game              string                    `json:"engine,omitempty"`
	DefID             int                       `json:"reelset,omitempty"`
	ReelsetID         string                    `json:"reelset_id,omitempty"`
	BetPerLine        Money                     `json:"bet_per_line,omitempty"`
	Transactions      []WalletTransaction       `json:"transactions,omitempty"`
	PreviousGamestate string                    `json:"previous_gamestate,omitempty"`
	NextGamestate     string                    `json:"next_gamestate,omitempty"`
	Action            string                    `json:"action,omitempty"`
	SymbolGrid        [][]int                   `json:"symbol_grid,omitempty"`
	RecoveryGrid      [][]int                   `json:"recovery_grid,omitempty"`
	Prizes            []Prize                   `json:"prizes,omitempty"`
	SelectedWinLines  []int                     `json:"selected_win_lines,omitempty"`
	RelativePayout    int                       `json:"relative_payout,omitempty"`
	Multiplier        int                       `json:"multiplier,omitempty"`
	StopList          []int                     `json:"stop_list,omitempty"`
	NextActions       []string                  `json:"next_actions,omitempty"`
	Gamification      *GamestatePB_Gamification `json:"gamification,omitempty"`
	CumulativeWin     Fixed                     `json:"cumulative_win,omitempty"`
	SpinWin           Fixed                     `json:"spin_win,omitempty"`
	CampaignWin       Fixed                     `json:"campaign_win,omitempty"`
	CampaignRef       string                    `json:"campaign_ref,omitempty"`
	PlaySequence      int                       `json:"play_sequence,omitempty"`
	Closed            bool                      `json:"closed"`
	RoundID           string                    `json:"round_id"`
	Features          []feature.Feature         `json:"features,omitempty"`
	FeatureView       [][]int                   `json:"feature_view,omitempty"`
	Replay            bool
	ReplayParams      feature.FeatureParams
}

func (gamestate Gamestate) Engine() (engine EngineConfig, err rgserror.RGSErr) {
	var engineID string
	engineID, err = config.GetEngineFromGame(gamestate.Game)
	if err != nil {
		logger.Errorf("error parsing game name: %v", gamestate.Game)
		return
	}
	engine = BuildEngineDefs(engineID)
	return
}

func (gamestate Gamestate) EngineDef() (engineDef EngineDef, err rgserror.RGSErr) {
	EC, err := gamestate.Engine()
	if err != nil {
		return
	}
	if gamestate.DefID > len(EC.EngineDefs) {
		err = rgserror.Create(rgserror.BadConfigError)
		return
	}
	engineDef = EC.EngineDefs[gamestate.DefID]
	return
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

func convertPrizesFromPB(unconverted []*PrizePB, betPerLine Fixed, gameID string, rS int32) []Prize {
	engineConf, err := GetEngineDefFromGame(gameID)
	if err != nil {
		logger.Errorf("error convering gamestate: %v", err)
		return []Prize{}
	}

	converted := make([]Prize, len(unconverted))
	for i, prize := range unconverted {
		convertedP := prize.Convert(betPerLine)
		convertedP.Index = engineConf.DetectSpecialWins(int(rS), convertedP)
		converted[i] = convertedP
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

func convertFeaturesFromPB(unconverted []*FeaturePB) []feature.Feature {
	converted := make([]feature.Feature, len(unconverted))
	for i, featurepb := range unconverted {
		ft := feature.MakeFeature(featurepb.Type)
		if ft != nil {
			ft.DefPtr().Id = featurepb.Id
			ft.DefPtr().Type = featurepb.Type
			err := ft.Deserialize(featurepb.Data)
			if err != nil {
				logger.Errorf(err.Error())
				return []feature.Feature{}
			} else {
				converted[i] = ft
			}
		} else {
			logger.Errorf("could not deserialize unknown feature %s"+
				", check engine.EnabledFeatureSet", featurepb.Type)
			return []feature.Feature{}
		}
	}
	return converted
}

func convertFeaturesToPB(unconverted []feature.Feature) []*FeaturePB {
	converted := make([]*FeaturePB, len(unconverted))
	for i, feature := range unconverted {
		var featurepb FeaturePB
		var err error
		featurepb.Id = feature.DefPtr().Id
		featurepb.Type = feature.DefPtr().Type
		featurepb.Data, err = feature.Serialize()
		if err != nil {
			logger.Errorf("could not serialize feature %s"+
				", error: %s", feature.DefPtr().Type, err.Error())
			return []*FeaturePB{}
		}
		converted[i] = &featurepb
	}
	return converted
}

func (gamestatePB GamestatePB) Convert() Gamestate {
	nextActions := make([]string, len(gamestatePB.NextActions))
	for i, action := range gamestatePB.NextActions {
		nextActions[i] = action.String()
	}
	// every set of transactions should begin with a WAGER. this ID is also the gamestate ID
	if len(gamestatePB.Transactions) == 0 {
		logger.Errorf("NO TX associated")
		return Gamestate{}
	}
	gameID := GetGameIDFromPB(gamestatePB.GameId.String())
	return Gamestate{
		Id:                gamestatePB.Transactions[0].Id,
		Game:              gameID,
		DefID:             int(gamestatePB.EngineDef),
		BetPerLine:        Money{Amount: Fixed(gamestatePB.BetPerLine), Currency: gamestatePB.Currency.String()},
		Transactions:      convertTransactionsFromPB(gamestatePB.Transactions),
		PreviousGamestate: string(gamestatePB.PreviousGamestate),
		NextGamestate:     string(gamestatePB.NextGamestate),
		Action:            gamestatePB.Action.String(),
		SymbolGrid:        convertSymbolGridFromPB(gamestatePB.SymbolGrid),
		Prizes:            convertPrizesFromPB(gamestatePB.Prizes, Fixed(gamestatePB.BetPerLine), gameID, gamestatePB.EngineDef),
		SelectedWinLines:  convertInt32Int(gamestatePB.SelectedWinLines),
		RelativePayout:    int(gamestatePB.RelativePayout),
		Multiplier:        int(gamestatePB.Multiplier),
		StopList:          convertInt32Int(gamestatePB.StopList),
		NextActions:       nextActions,
		Gamification:      gamestatePB.Gamification,
		CumulativeWin:     Fixed(gamestatePB.CumulativeWin),
		SpinWin:           Fixed(gamestatePB.SpinWin),
		CampaignWin:       Fixed(gamestatePB.CampaignWin),
		CampaignRef:       string(gamestatePB.CampaignRef),
		PlaySequence:      int(gamestatePB.PlaySequence),
		Closed:            gamestatePB.Closed,
		RoundID:           gamestatePB.RoundId,
		RecoveryGrid:      convertSymbolGridFromPB(gamestatePB.RecoveryGrid),
		Features:          convertFeaturesFromPB(gamestatePB.Features),
		FeatureView:       convertSymbolGridFromPB(gamestatePB.FeatureView),
		ReelsetID:         gamestatePB.ReelsetId,
	}
}

func (gamestatePB GamestatePB) ConvertLegacy(transactions []*WalletTransactionPB) Gamestate {
	nextActions := make([]string, len(gamestatePB.NextActions))
	for i, action := range gamestatePB.NextActions {
		nextActions[i] = action.String()
	}
	// every set of transactions should begin with a WAGER. this ID is also the gamestate ID

	// get Game ID
	gameID := GetGameIDFromPB(gamestatePB.GameId.String())
	return Gamestate{
		Id:                transactions[0].Id,
		Game:              gameID,
		DefID:             int(gamestatePB.EngineDef),
		BetPerLine:        Money{Amount: Fixed(gamestatePB.BetPerLine), Currency: gamestatePB.Currency.String()},
		Transactions:      convertTransactionsFromPB(transactions),
		PreviousGamestate: string(gamestatePB.PreviousGamestate),
		NextGamestate:     string(gamestatePB.NextGamestate),
		Action:            gamestatePB.Action.String(),
		SymbolGrid:        convertSymbolGridFromPB(gamestatePB.SymbolGrid),
		Prizes:            convertPrizesFromPB(gamestatePB.Prizes, Fixed(gamestatePB.BetPerLine), gameID, gamestatePB.EngineDef),
		SelectedWinLines:  convertInt32Int(gamestatePB.SelectedWinLines),
		RelativePayout:    int(gamestatePB.RelativePayout),
		Multiplier:        int(gamestatePB.Multiplier),
		StopList:          convertInt32Int(gamestatePB.StopList),
		NextActions:       nextActions,
		Gamification:      gamestatePB.Gamification,
		CumulativeWin:     Fixed(gamestatePB.CumulativeWin),
		PlaySequence:      int(gamestatePB.PlaySequence),
	}
}

func GetEngineDefFromGame(gameID string) (EngineConfig, rgserror.RGSErr) {
	engineID, err := config.GetEngineFromGame(gameID)
	if err != nil {
		return EngineConfig{}, err
	}
	return BuildEngineDefs(engineID), nil
}

func (gamestate Gamestate) Convert() GamestatePB {
	nextActions := make([]GamestatePB_Action, len(gamestate.NextActions))
	for i, action := range gamestate.NextActions {
		nextActions[i] = GamestatePB_Action(GamestatePB_Action_value[action])
	}
	// every set of transactions should begin with a WAGER. this ID is also the gamestate ID
	return GamestatePB{
		GameId:            GamestatePB_GameID(GamestatePB_GameID_value[GetPBFromGameID(gamestate.Game)]),
		EngineDef:         int32(gamestate.DefID),
		BetPerLine:        gamestate.BetPerLine.Amount.ValueRaw(),
		Currency:          Ccy(Ccy_value[gamestate.BetPerLine.Currency]),
		PreviousGamestate: []byte(gamestate.PreviousGamestate),
		NextGamestate:     []byte(gamestate.NextGamestate),
		Action:            GamestatePB_Action(GamestatePB_Action_value[gamestate.Action]),
		SymbolGrid:        convertSymbolGridToPB(gamestate.SymbolGrid),
		RecoveryGrid:      convertSymbolGridToPB(gamestate.RecoveryGrid),
		Prizes:            convertPrizesToPB(gamestate.Prizes),
		SelectedWinLines:  convertIntInt32(gamestate.SelectedWinLines),
		RelativePayout:    int32(gamestate.RelativePayout),
		Multiplier:        int32(gamestate.Multiplier),
		StopList:          convertIntInt32(gamestate.StopList),
		NextActions:       nextActions,
		Gamification:      gamestate.Gamification,
		CumulativeWin:     gamestate.CumulativeWin.ValueRaw(),
		SpinWin:           gamestate.SpinWin.ValueRaw(),
		CampaignWin:       gamestate.CampaignWin.ValueRaw(),
		CampaignRef:       gamestate.CampaignRef,
		PlaySequence:      int32(gamestate.PlaySequence),
		Transactions:      convertTransactionsToPB(gamestate.Transactions),
		Closed:            gamestate.Closed,
		RoundId:           gamestate.RoundID,
		Features:          convertFeaturesToPB(gamestate.Features),
		FeatureView:       convertSymbolGridToPB(gamestate.FeatureView),
		ReelsetId:         gamestate.ReelsetID,
	}
}

func (gamestate Gamestate) ConvertLegacy() (GamestatePB, []*WalletTransactionPB) {
	// todo: deprecate this in 6 months as all txs fitting this will have been removed
	sentry.CaptureMessage("Performing legacy gamestate conversion")
	nextActions := make([]GamestatePB_Action, len(gamestate.NextActions))
	for i, action := range gamestate.NextActions {
		nextActions[i] = GamestatePB_Action(GamestatePB_Action_value[action])
	}
	// every set of transactions should begin with a WAGER. this ID is also the gamestate ID
	return GamestatePB{
		GameId:            GamestatePB_GameID(GamestatePB_GameID_value[GetPBFromGameID(gamestate.Game)]),
		EngineDef:         int32(gamestate.DefID),
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
		PlaySequence:      int32(gamestate.PlaySequence),
	}, convertTransactionsToPB(gamestate.Transactions)
}

type IGameStateV3 interface {
	Serialize() []byte
	Deserialize([]byte) rgse.RGSErr
	GetTtl() int64
	Base() *GameStateV3
}

type GameStateV3 struct {
	Id                string              `json:"id"`
	Game              string              `json:"game"`
	Version           string              `json:"version"`
	Currency          string              `json:"ccy"`
	Transactions      []WalletTransaction `json:"transactions"`
	PreviousGamestate string              `json:"prevGamestate"`
	NextGamestate     string              `json:"nextGamestate"`
	Closed            bool                `json:"closed"`
	RoundId           string              `json:"roundId"`
	Features          []feature.Feature   `json:"features"`
}

func (s *GameStateV3) Base() *GameStateV3 {
	return s
}

func (s GameStateV3) Serialize() []byte {
	b, _ := json.Marshal(s)
	logger.Debugf("GameStateV3.Serialize %s", string(b))
	return b
}

func (s *GameStateV3) Deserialize(serialized []byte) rgse.RGSErr {
	err := json.Unmarshal(serialized, s)
	if err != nil {
		logger.Debugf("unmarshal json failed with error %s", err.Error())
		return rgse.Create(rgse.GamestateByteDeserializerError)
	}
	return nil
}

func (s GameStateV3) GetTtl() int64 {
	return 3600
}
