// Code generated by protoc-gen-go. DO NOT EDIT.
// source: datatypes_v1.proto

package engine

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Ccy int32

const (
	Ccy_DEFAULT Ccy = 0
	Ccy_CNY     Ccy = 1
	Ccy_USD     Ccy = 2
	Ccy_GBP     Ccy = 3
	Ccy_EUR     Ccy = 4
	Ccy_JPY     Ccy = 6
	Ccy_THB     Ccy = 7
	Ccy_MYR     Ccy = 8
	Ccy_VND     Ccy = 9
	Ccy_KRW     Ccy = 10
	Ccy_IDR     Ccy = 11
	Ccy_ZAR     Ccy = 12
	Ccy_XBT     Ccy = 13
	Ccy_TRY     Ccy = 14
)

var Ccy_name = map[int32]string{
	0:  "DEFAULT",
	1:  "CNY",
	2:  "USD",
	3:  "GBP",
	4:  "EUR",
	6:  "JPY",
	7:  "THB",
	8:  "MYR",
	9:  "VND",
	10: "KRW",
	11: "IDR",
	12: "ZAR",
	13: "XBT",
	14: "TRY",
}

var Ccy_value = map[string]int32{
	"DEFAULT": 0,
	"CNY":     1,
	"USD":     2,
	"GBP":     3,
	"EUR":     4,
	"JPY":     6,
	"THB":     7,
	"MYR":     8,
	"VND":     9,
	"KRW":     10,
	"IDR":     11,
	"ZAR":     12,
	"XBT":     13,
	"TRY":     14,
}

func (x Ccy) String() string {
	return proto.EnumName(Ccy_name, int32(x))
}

func (Ccy) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_4563ec715222b270, []int{0}
}

type WalletTransactionPB_Type int32

const (
	WalletTransactionPB_DEFAULT  WalletTransactionPB_Type = 0
	WalletTransactionPB_WAGER    WalletTransactionPB_Type = 1
	WalletTransactionPB_PAYOUT   WalletTransactionPB_Type = 2
	WalletTransactionPB_ENDROUND WalletTransactionPB_Type = 3
)

var WalletTransactionPB_Type_name = map[int32]string{
	0: "DEFAULT",
	1: "WAGER",
	2: "PAYOUT",
	3: "ENDROUND",
}

var WalletTransactionPB_Type_value = map[string]int32{
	"DEFAULT":  0,
	"WAGER":    1,
	"PAYOUT":   2,
	"ENDROUND": 3,
}

func (x WalletTransactionPB_Type) String() string {
	return proto.EnumName(WalletTransactionPB_Type_name, int32(x))
}

func (WalletTransactionPB_Type) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_4563ec715222b270, []int{2, 0}
}

type GamestatePB_GameID int32

const (
	GamestatePB_DEFAULT                 GamestatePB_GameID = 0
	GamestatePB_THE_YEAR_OF_ZHU         GamestatePB_GameID = 1
	GamestatePB_ZODIAC                  GamestatePB_GameID = 2
	GamestatePB_CAT_THIEF               GamestatePB_GameID = 3
	GamestatePB_THREE_KINGDOM_SHU       GamestatePB_GameID = 4
	GamestatePB_THREE_KINGDOM_WEI       GamestatePB_GameID = 5
	GamestatePB_THREE_KINGDOM_WU        GamestatePB_GameID = 6
	GamestatePB_CRIMSON_MASQUERADE      GamestatePB_GameID = 7
	GamestatePB_JUNGLE_SAGA             GamestatePB_GameID = 8
	GamestatePB_CANDY_GIRLS             GamestatePB_GameID = 9
	GamestatePB_WUKONG_TREASURES        GamestatePB_GameID = 10
	GamestatePB_STREET_RACER            GamestatePB_GameID = 11
	GamestatePB_BABAKS_QUEST            GamestatePB_GameID = 12
	GamestatePB_ASTRO_GEMS              GamestatePB_GameID = 13
	GamestatePB_PANDA                   GamestatePB_GameID = 14
	GamestatePB_KING_OF_GAMBLERS        GamestatePB_GameID = 15
	GamestatePB_ART_OF_THE_FIST         GamestatePB_GameID = 16
	GamestatePB_SEASONS_WINTER          GamestatePB_GameID = 17
	GamestatePB_SEASONS_SPRING          GamestatePB_GameID = 18
	GamestatePB_SEASONS_SUMMER          GamestatePB_GameID = 19
	GamestatePB_SEASONS_AUTUMN          GamestatePB_GameID = 20
	GamestatePB_SEASONS                 GamestatePB_GameID = 21
	GamestatePB_FRUITY_VERSE            GamestatePB_GameID = 22
	GamestatePB_A_FAIRY_TALE            GamestatePB_GameID = 23
	GamestatePB_A_HIDDEN_FOREST         GamestatePB_GameID = 24
	GamestatePB_BATTLEMECH              GamestatePB_GameID = 25
	GamestatePB_A_MILLION_LIGHTS        GamestatePB_GameID = 26
	GamestatePB_BISTRO                  GamestatePB_GameID = 27
	GamestatePB_CLOUD9                  GamestatePB_GameID = 28
	GamestatePB_LANTERN_FESTIVAL        GamestatePB_GameID = 29
	GamestatePB_A_CANDY_GIRLS_CHRISTMAS GamestatePB_GameID = 30
	GamestatePB_SKY_JEWELS              GamestatePB_GameID = 31
	GamestatePB_PEARL_FISHER            GamestatePB_GameID = 32
	GamestatePB_GOAL                    GamestatePB_GameID = 33
	GamestatePB_DAYTONA                 GamestatePB_GameID = 34
	GamestatePB_A_YEAR_OF_LAOSHU        GamestatePB_GameID = 35
	GamestatePB_DRAGON_MYST             GamestatePB_GameID = 36
	GamestatePB_COOKOFF_CHAMPION        GamestatePB_GameID = 37
	GamestatePB_CANDY_SMASH             GamestatePB_GameID = 38
	GamestatePB_VALLEY_OF_KINGS         GamestatePB_GameID = 39
)

var GamestatePB_GameID_name = map[int32]string{
	0:  "DEFAULT",
	1:  "THE_YEAR_OF_ZHU",
	2:  "ZODIAC",
	3:  "CAT_THIEF",
	4:  "THREE_KINGDOM_SHU",
	5:  "THREE_KINGDOM_WEI",
	6:  "THREE_KINGDOM_WU",
	7:  "CRIMSON_MASQUERADE",
	8:  "JUNGLE_SAGA",
	9:  "CANDY_GIRLS",
	10: "WUKONG_TREASURES",
	11: "STREET_RACER",
	12: "BABAKS_QUEST",
	13: "ASTRO_GEMS",
	14: "PANDA",
	15: "KING_OF_GAMBLERS",
	16: "ART_OF_THE_FIST",
	17: "SEASONS_WINTER",
	18: "SEASONS_SPRING",
	19: "SEASONS_SUMMER",
	20: "SEASONS_AUTUMN",
	21: "SEASONS",
	22: "FRUITY_VERSE",
	23: "A_FAIRY_TALE",
	24: "A_HIDDEN_FOREST",
	25: "BATTLEMECH",
	26: "A_MILLION_LIGHTS",
	27: "BISTRO",
	28: "CLOUD9",
	29: "LANTERN_FESTIVAL",
	30: "A_CANDY_GIRLS_CHRISTMAS",
	31: "SKY_JEWELS",
	32: "PEARL_FISHER",
	33: "GOAL",
	34: "DAYTONA",
	35: "A_YEAR_OF_LAOSHU",
	36: "DRAGON_MYST",
	37: "COOKOFF_CHAMPION",
	38: "CANDY_SMASH",
	39: "VALLEY_OF_KINGS",
}

var GamestatePB_GameID_value = map[string]int32{
	"DEFAULT":                 0,
	"THE_YEAR_OF_ZHU":         1,
	"ZODIAC":                  2,
	"CAT_THIEF":               3,
	"THREE_KINGDOM_SHU":       4,
	"THREE_KINGDOM_WEI":       5,
	"THREE_KINGDOM_WU":        6,
	"CRIMSON_MASQUERADE":      7,
	"JUNGLE_SAGA":             8,
	"CANDY_GIRLS":             9,
	"WUKONG_TREASURES":        10,
	"STREET_RACER":            11,
	"BABAKS_QUEST":            12,
	"ASTRO_GEMS":              13,
	"PANDA":                   14,
	"KING_OF_GAMBLERS":        15,
	"ART_OF_THE_FIST":         16,
	"SEASONS_WINTER":          17,
	"SEASONS_SPRING":          18,
	"SEASONS_SUMMER":          19,
	"SEASONS_AUTUMN":          20,
	"SEASONS":                 21,
	"FRUITY_VERSE":            22,
	"A_FAIRY_TALE":            23,
	"A_HIDDEN_FOREST":         24,
	"BATTLEMECH":              25,
	"A_MILLION_LIGHTS":        26,
	"BISTRO":                  27,
	"CLOUD9":                  28,
	"LANTERN_FESTIVAL":        29,
	"A_CANDY_GIRLS_CHRISTMAS": 30,
	"SKY_JEWELS":              31,
	"PEARL_FISHER":            32,
	"GOAL":                    33,
	"DAYTONA":                 34,
	"A_YEAR_OF_LAOSHU":        35,
	"DRAGON_MYST":             36,
	"COOKOFF_CHAMPION":        37,
	"CANDY_SMASH":             38,
	"VALLEY_OF_KINGS":         39,
}

func (x GamestatePB_GameID) String() string {
	return proto.EnumName(GamestatePB_GameID_name, int32(x))
}

func (GamestatePB_GameID) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_4563ec715222b270, []int{3, 0}
}

type GamestatePB_Action int32

const (
	GamestatePB_base       GamestatePB_Action = 0
	GamestatePB_finish     GamestatePB_Action = 1
	GamestatePB_freespin   GamestatePB_Action = 2
	GamestatePB_pickSpins  GamestatePB_Action = 3
	GamestatePB_respin     GamestatePB_Action = 4
	GamestatePB_freespin2  GamestatePB_Action = 5
	GamestatePB_freespin3  GamestatePB_Action = 6
	GamestatePB_freespin4  GamestatePB_Action = 7
	GamestatePB_freespin5  GamestatePB_Action = 8
	GamestatePB_freespin10 GamestatePB_Action = 9
	GamestatePB_freespin25 GamestatePB_Action = 10
	GamestatePB_cascade    GamestatePB_Action = 11
)

var GamestatePB_Action_name = map[int32]string{
	0:  "base",
	1:  "finish",
	2:  "freespin",
	3:  "pickSpins",
	4:  "respin",
	5:  "freespin2",
	6:  "freespin3",
	7:  "freespin4",
	8:  "freespin5",
	9:  "freespin10",
	10: "freespin25",
	11: "cascade",
}

var GamestatePB_Action_value = map[string]int32{
	"base":       0,
	"finish":     1,
	"freespin":   2,
	"pickSpins":  3,
	"respin":     4,
	"freespin2":  5,
	"freespin3":  6,
	"freespin4":  7,
	"freespin5":  8,
	"freespin10": 9,
	"freespin25": 10,
	"cascade":    11,
}

func (x GamestatePB_Action) String() string {
	return proto.EnumName(GamestatePB_Action_name, int32(x))
}

func (GamestatePB_Action) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_4563ec715222b270, []int{3, 1}
}

type PayoutPB struct {
	Symbol               int32    `protobuf:"varint,1,opt,name=symbol,proto3" json:"symbol,omitempty"`
	Count                int32    `protobuf:"varint,2,opt,name=count,proto3" json:"count,omitempty"`
	Multiplier           int32    `protobuf:"varint,3,opt,name=multiplier,proto3" json:"multiplier,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PayoutPB) Reset()         { *m = PayoutPB{} }
func (m *PayoutPB) String() string { return proto.CompactTextString(m) }
func (*PayoutPB) ProtoMessage()    {}
func (*PayoutPB) Descriptor() ([]byte, []int) {
	return fileDescriptor_4563ec715222b270, []int{0}
}

func (m *PayoutPB) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PayoutPB.Unmarshal(m, b)
}
func (m *PayoutPB) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PayoutPB.Marshal(b, m, deterministic)
}
func (m *PayoutPB) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PayoutPB.Merge(m, src)
}
func (m *PayoutPB) XXX_Size() int {
	return xxx_messageInfo_PayoutPB.Size(m)
}
func (m *PayoutPB) XXX_DiscardUnknown() {
	xxx_messageInfo_PayoutPB.DiscardUnknown(m)
}

var xxx_messageInfo_PayoutPB proto.InternalMessageInfo

func (m *PayoutPB) GetSymbol() int32 {
	if m != nil {
		return m.Symbol
	}
	return 0
}

func (m *PayoutPB) GetCount() int32 {
	if m != nil {
		return m.Count
	}
	return 0
}

func (m *PayoutPB) GetMultiplier() int32 {
	if m != nil {
		return m.Multiplier
	}
	return 0
}

type PrizePB struct {
	Payout               *PayoutPB `protobuf:"bytes,1,opt,name=payout,proto3" json:"payout,omitempty"`
	Multiplier           int32     `protobuf:"varint,2,opt,name=multiplier,proto3" json:"multiplier,omitempty"`
	SymbolPositions      []int32   `protobuf:"varint,3,rep,packed,name=symbol_positions,json=symbolPositions,proto3" json:"symbol_positions,omitempty"`
	Winline              int32     `protobuf:"varint,4,opt,name=winline,proto3" json:"winline,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *PrizePB) Reset()         { *m = PrizePB{} }
func (m *PrizePB) String() string { return proto.CompactTextString(m) }
func (*PrizePB) ProtoMessage()    {}
func (*PrizePB) Descriptor() ([]byte, []int) {
	return fileDescriptor_4563ec715222b270, []int{1}
}

func (m *PrizePB) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PrizePB.Unmarshal(m, b)
}
func (m *PrizePB) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PrizePB.Marshal(b, m, deterministic)
}
func (m *PrizePB) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PrizePB.Merge(m, src)
}
func (m *PrizePB) XXX_Size() int {
	return xxx_messageInfo_PrizePB.Size(m)
}
func (m *PrizePB) XXX_DiscardUnknown() {
	xxx_messageInfo_PrizePB.DiscardUnknown(m)
}

var xxx_messageInfo_PrizePB proto.InternalMessageInfo

func (m *PrizePB) GetPayout() *PayoutPB {
	if m != nil {
		return m.Payout
	}
	return nil
}

func (m *PrizePB) GetMultiplier() int32 {
	if m != nil {
		return m.Multiplier
	}
	return 0
}

func (m *PrizePB) GetSymbolPositions() []int32 {
	if m != nil {
		return m.SymbolPositions
	}
	return nil
}

func (m *PrizePB) GetWinline() int32 {
	if m != nil {
		return m.Winline
	}
	return 0
}

type WalletTransactionPB struct {
	Id                   string                   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Amount               int64                    `protobuf:"varint,2,opt,name=amount,proto3" json:"amount,omitempty"`
	Type                 WalletTransactionPB_Type `protobuf:"varint,3,opt,name=type,proto3,enum=engine.WalletTransactionPB_Type" json:"type,omitempty"`
	Currency             Ccy                      `protobuf:"varint,4,opt,name=currency,proto3,enum=engine.Ccy" json:"currency,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                 `json:"-"`
	XXX_unrecognized     []byte                   `json:"-"`
	XXX_sizecache        int32                    `json:"-"`
}

func (m *WalletTransactionPB) Reset()         { *m = WalletTransactionPB{} }
func (m *WalletTransactionPB) String() string { return proto.CompactTextString(m) }
func (*WalletTransactionPB) ProtoMessage()    {}
func (*WalletTransactionPB) Descriptor() ([]byte, []int) {
	return fileDescriptor_4563ec715222b270, []int{2}
}

func (m *WalletTransactionPB) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_WalletTransactionPB.Unmarshal(m, b)
}
func (m *WalletTransactionPB) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_WalletTransactionPB.Marshal(b, m, deterministic)
}
func (m *WalletTransactionPB) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WalletTransactionPB.Merge(m, src)
}
func (m *WalletTransactionPB) XXX_Size() int {
	return xxx_messageInfo_WalletTransactionPB.Size(m)
}
func (m *WalletTransactionPB) XXX_DiscardUnknown() {
	xxx_messageInfo_WalletTransactionPB.DiscardUnknown(m)
}

var xxx_messageInfo_WalletTransactionPB proto.InternalMessageInfo

func (m *WalletTransactionPB) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *WalletTransactionPB) GetAmount() int64 {
	if m != nil {
		return m.Amount
	}
	return 0
}

func (m *WalletTransactionPB) GetType() WalletTransactionPB_Type {
	if m != nil {
		return m.Type
	}
	return WalletTransactionPB_DEFAULT
}

func (m *WalletTransactionPB) GetCurrency() Ccy {
	if m != nil {
		return m.Currency
	}
	return Ccy_DEFAULT
}

type GamestatePB struct {
	GameId               GamestatePB_GameID        `protobuf:"varint,1,opt,name=game_id,json=gameId,proto3,enum=engine.GamestatePB_GameID" json:"game_id,omitempty"`
	EngineDef            int32                     `protobuf:"varint,2,opt,name=engine_def,json=engineDef,proto3" json:"engine_def,omitempty"`
	BetPerLine           int64                     `protobuf:"varint,3,opt,name=bet_per_line,json=betPerLine,proto3" json:"bet_per_line,omitempty"`
	Currency             Ccy                       `protobuf:"varint,4,opt,name=currency,proto3,enum=engine.Ccy" json:"currency,omitempty"`
	PreviousGamestate    []byte                    `protobuf:"bytes,5,opt,name=previous_gamestate,json=previousGamestate,proto3" json:"previous_gamestate,omitempty"`
	NextGamestate        []byte                    `protobuf:"bytes,6,opt,name=next_gamestate,json=nextGamestate,proto3" json:"next_gamestate,omitempty"`
	Action               GamestatePB_Action        `protobuf:"varint,7,opt,name=action,proto3,enum=engine.GamestatePB_Action" json:"action,omitempty"`
	SymbolGrid           []*GamestatePB_Reel       `protobuf:"bytes,8,rep,name=symbol_grid,json=symbolGrid,proto3" json:"symbol_grid,omitempty"`
	Prizes               []*PrizePB                `protobuf:"bytes,9,rep,name=prizes,proto3" json:"prizes,omitempty"`
	SelectedWinLines     []int32                   `protobuf:"varint,10,rep,packed,name=selected_win_lines,json=selectedWinLines,proto3" json:"selected_win_lines,omitempty"`
	RelativePayout       int32                     `protobuf:"varint,11,opt,name=relative_payout,json=relativePayout,proto3" json:"relative_payout,omitempty"`
	Multiplier           int32                     `protobuf:"varint,12,opt,name=multiplier,proto3" json:"multiplier,omitempty"`
	StopList             []int32                   `protobuf:"varint,13,rep,packed,name=stop_list,json=stopList,proto3" json:"stop_list,omitempty"`
	NextActions          []GamestatePB_Action      `protobuf:"varint,14,rep,packed,name=next_actions,json=nextActions,proto3,enum=engine.GamestatePB_Action" json:"next_actions,omitempty"`
	Gamification         *GamestatePB_Gamification `protobuf:"bytes,15,opt,name=gamification,proto3" json:"gamification,omitempty"`
	CumulativeWin        int64                     `protobuf:"varint,16,opt,name=cumulative_win,json=cumulativeWin,proto3" json:"cumulative_win,omitempty"`
	PlaySequence         int32                     `protobuf:"varint,17,opt,name=play_sequence,json=playSequence,proto3" json:"play_sequence,omitempty"`
	Transactions         []*WalletTransactionPB    `protobuf:"bytes,18,rep,name=transactions,proto3" json:"transactions,omitempty"`
	Closed               bool                      `protobuf:"varint,19,opt,name=closed,proto3" json:"closed,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                  `json:"-"`
	XXX_unrecognized     []byte                    `json:"-"`
	XXX_sizecache        int32                     `json:"-"`
}

func (m *GamestatePB) Reset()         { *m = GamestatePB{} }
func (m *GamestatePB) String() string { return proto.CompactTextString(m) }
func (*GamestatePB) ProtoMessage()    {}
func (*GamestatePB) Descriptor() ([]byte, []int) {
	return fileDescriptor_4563ec715222b270, []int{3}
}

func (m *GamestatePB) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GamestatePB.Unmarshal(m, b)
}
func (m *GamestatePB) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GamestatePB.Marshal(b, m, deterministic)
}
func (m *GamestatePB) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GamestatePB.Merge(m, src)
}
func (m *GamestatePB) XXX_Size() int {
	return xxx_messageInfo_GamestatePB.Size(m)
}
func (m *GamestatePB) XXX_DiscardUnknown() {
	xxx_messageInfo_GamestatePB.DiscardUnknown(m)
}

var xxx_messageInfo_GamestatePB proto.InternalMessageInfo

func (m *GamestatePB) GetGameId() GamestatePB_GameID {
	if m != nil {
		return m.GameId
	}
	return GamestatePB_DEFAULT
}

func (m *GamestatePB) GetEngineDef() int32 {
	if m != nil {
		return m.EngineDef
	}
	return 0
}

func (m *GamestatePB) GetBetPerLine() int64 {
	if m != nil {
		return m.BetPerLine
	}
	return 0
}

func (m *GamestatePB) GetCurrency() Ccy {
	if m != nil {
		return m.Currency
	}
	return Ccy_DEFAULT
}

func (m *GamestatePB) GetPreviousGamestate() []byte {
	if m != nil {
		return m.PreviousGamestate
	}
	return nil
}

func (m *GamestatePB) GetNextGamestate() []byte {
	if m != nil {
		return m.NextGamestate
	}
	return nil
}

func (m *GamestatePB) GetAction() GamestatePB_Action {
	if m != nil {
		return m.Action
	}
	return GamestatePB_base
}

func (m *GamestatePB) GetSymbolGrid() []*GamestatePB_Reel {
	if m != nil {
		return m.SymbolGrid
	}
	return nil
}

func (m *GamestatePB) GetPrizes() []*PrizePB {
	if m != nil {
		return m.Prizes
	}
	return nil
}

func (m *GamestatePB) GetSelectedWinLines() []int32 {
	if m != nil {
		return m.SelectedWinLines
	}
	return nil
}

func (m *GamestatePB) GetRelativePayout() int32 {
	if m != nil {
		return m.RelativePayout
	}
	return 0
}

func (m *GamestatePB) GetMultiplier() int32 {
	if m != nil {
		return m.Multiplier
	}
	return 0
}

func (m *GamestatePB) GetStopList() []int32 {
	if m != nil {
		return m.StopList
	}
	return nil
}

func (m *GamestatePB) GetNextActions() []GamestatePB_Action {
	if m != nil {
		return m.NextActions
	}
	return nil
}

func (m *GamestatePB) GetGamification() *GamestatePB_Gamification {
	if m != nil {
		return m.Gamification
	}
	return nil
}

func (m *GamestatePB) GetCumulativeWin() int64 {
	if m != nil {
		return m.CumulativeWin
	}
	return 0
}

func (m *GamestatePB) GetPlaySequence() int32 {
	if m != nil {
		return m.PlaySequence
	}
	return 0
}

func (m *GamestatePB) GetTransactions() []*WalletTransactionPB {
	if m != nil {
		return m.Transactions
	}
	return nil
}

func (m *GamestatePB) GetClosed() bool {
	if m != nil {
		return m.Closed
	}
	return false
}

type GamestatePB_Reel struct {
	Symbols              []int32  `protobuf:"varint,1,rep,packed,name=symbols,proto3" json:"symbols,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GamestatePB_Reel) Reset()         { *m = GamestatePB_Reel{} }
func (m *GamestatePB_Reel) String() string { return proto.CompactTextString(m) }
func (*GamestatePB_Reel) ProtoMessage()    {}
func (*GamestatePB_Reel) Descriptor() ([]byte, []int) {
	return fileDescriptor_4563ec715222b270, []int{3, 0}
}

func (m *GamestatePB_Reel) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GamestatePB_Reel.Unmarshal(m, b)
}
func (m *GamestatePB_Reel) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GamestatePB_Reel.Marshal(b, m, deterministic)
}
func (m *GamestatePB_Reel) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GamestatePB_Reel.Merge(m, src)
}
func (m *GamestatePB_Reel) XXX_Size() int {
	return xxx_messageInfo_GamestatePB_Reel.Size(m)
}
func (m *GamestatePB_Reel) XXX_DiscardUnknown() {
	xxx_messageInfo_GamestatePB_Reel.DiscardUnknown(m)
}

var xxx_messageInfo_GamestatePB_Reel proto.InternalMessageInfo

func (m *GamestatePB_Reel) GetSymbols() []int32 {
	if m != nil {
		return m.Symbols
	}
	return nil
}

type GamestatePB_Gamification struct {
	Level                int32    `protobuf:"varint,1,opt,name=level,proto3" json:"level,omitempty"`
	Stage                int32    `protobuf:"varint,2,opt,name=stage,proto3" json:"stage,omitempty"`
	RemainingSpins       int32    `protobuf:"varint,3,opt,name=remaining_spins,json=remainingSpins,proto3" json:"remaining_spins,omitempty"`
	SpinsToStageUp       int32    `protobuf:"varint,4,opt,name=spins_to_stage_up,json=spinsToStageUp,proto3" json:"spins_to_stage_up,omitempty"`
	TotalSpins           int32    `protobuf:"varint,5,opt,name=total_spins,json=totalSpins,proto3" json:"total_spins,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GamestatePB_Gamification) Reset()         { *m = GamestatePB_Gamification{} }
func (m *GamestatePB_Gamification) String() string { return proto.CompactTextString(m) }
func (*GamestatePB_Gamification) ProtoMessage()    {}
func (*GamestatePB_Gamification) Descriptor() ([]byte, []int) {
	return fileDescriptor_4563ec715222b270, []int{3, 1}
}

func (m *GamestatePB_Gamification) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GamestatePB_Gamification.Unmarshal(m, b)
}
func (m *GamestatePB_Gamification) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GamestatePB_Gamification.Marshal(b, m, deterministic)
}
func (m *GamestatePB_Gamification) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GamestatePB_Gamification.Merge(m, src)
}
func (m *GamestatePB_Gamification) XXX_Size() int {
	return xxx_messageInfo_GamestatePB_Gamification.Size(m)
}
func (m *GamestatePB_Gamification) XXX_DiscardUnknown() {
	xxx_messageInfo_GamestatePB_Gamification.DiscardUnknown(m)
}

var xxx_messageInfo_GamestatePB_Gamification proto.InternalMessageInfo

func (m *GamestatePB_Gamification) GetLevel() int32 {
	if m != nil {
		return m.Level
	}
	return 0
}

func (m *GamestatePB_Gamification) GetStage() int32 {
	if m != nil {
		return m.Stage
	}
	return 0
}

func (m *GamestatePB_Gamification) GetRemainingSpins() int32 {
	if m != nil {
		return m.RemainingSpins
	}
	return 0
}

func (m *GamestatePB_Gamification) GetSpinsToStageUp() int32 {
	if m != nil {
		return m.SpinsToStageUp
	}
	return 0
}

func (m *GamestatePB_Gamification) GetTotalSpins() int32 {
	if m != nil {
		return m.TotalSpins
	}
	return 0
}

func init() {
	proto.RegisterEnum("engine.Ccy", Ccy_name, Ccy_value)
	proto.RegisterEnum("engine.WalletTransactionPB_Type", WalletTransactionPB_Type_name, WalletTransactionPB_Type_value)
	proto.RegisterEnum("engine.GamestatePB_GameID", GamestatePB_GameID_name, GamestatePB_GameID_value)
	proto.RegisterEnum("engine.GamestatePB_Action", GamestatePB_Action_name, GamestatePB_Action_value)
	proto.RegisterType((*PayoutPB)(nil), "engine.PayoutPB")
	proto.RegisterType((*PrizePB)(nil), "engine.PrizePB")
	proto.RegisterType((*WalletTransactionPB)(nil), "engine.WalletTransactionPB")
	proto.RegisterType((*GamestatePB)(nil), "engine.GamestatePB")
	proto.RegisterType((*GamestatePB_Reel)(nil), "engine.GamestatePB.Reel")
	proto.RegisterType((*GamestatePB_Gamification)(nil), "engine.GamestatePB.Gamification")
}

func init() { proto.RegisterFile("datatypes_v1.proto", fileDescriptor_4563ec715222b270) }

var fileDescriptor_4563ec715222b270 = []byte{
	// 1447 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x56, 0xef, 0x72, 0xdb, 0xc6,
	0x11, 0x0f, 0x45, 0xf1, 0xdf, 0x91, 0xa2, 0x56, 0x67, 0xc7, 0x41, 0xed, 0x26, 0x61, 0x95, 0xa6,
	0x56, 0x3a, 0xad, 0xa7, 0x91, 0x93, 0x99, 0xe6, 0x43, 0xa7, 0x73, 0x24, 0x8e, 0x24, 0x2c, 0x10,
	0x60, 0xee, 0x00, 0x33, 0xf4, 0x97, 0x1b, 0x88, 0x3c, 0xa9, 0x98, 0x52, 0x20, 0x4b, 0x80, 0x76,
	0xd5, 0x17, 0xe8, 0x2b, 0xf4, 0x2d, 0xda, 0xc7, 0xe8, 0x13, 0xf4, 0x11, 0xfa, 0x1c, 0x9d, 0x3d,
	0x00, 0x32, 0xe5, 0xd8, 0x6d, 0xbf, 0xed, 0xfe, 0x76, 0x6f, 0xb1, 0xfb, 0xdb, 0x3f, 0x24, 0xa1,
	0xcb, 0x28, 0x8b, 0xb2, 0xdb, 0x8d, 0x4e, 0xd5, 0xeb, 0xaf, 0x9f, 0x6d, 0xb6, 0xeb, 0x6c, 0x4d,
	0xeb, 0x3a, 0xb9, 0x8e, 0x13, 0x7d, 0xfa, 0x03, 0x69, 0x4e, 0xa3, 0xdb, 0xf5, 0x2e, 0x9b, 0xf6,
	0xe9, 0x23, 0x52, 0x4f, 0x6f, 0x6f, 0x2e, 0xd7, 0x2b, 0xab, 0xd2, 0xab, 0x9c, 0xd5, 0x44, 0xa1,
	0xd1, 0x87, 0xa4, 0xb6, 0x58, 0xef, 0x92, 0xcc, 0x3a, 0x30, 0x70, 0xae, 0xd0, 0xcf, 0x08, 0xb9,
	0xd9, 0xad, 0xb2, 0x78, 0xb3, 0x8a, 0xf5, 0xd6, 0xaa, 0x1a, 0xd3, 0x1e, 0x72, 0xfa, 0xb7, 0x0a,
	0x69, 0x4c, 0xb7, 0xf1, 0x5f, 0xf4, 0xb4, 0x4f, 0xcf, 0x48, 0x7d, 0x63, 0xbe, 0x62, 0x22, 0xb7,
	0xcf, 0xe1, 0x59, 0xfe, 0xf9, 0x67, 0xe5, 0xb7, 0x45, 0x61, 0x7f, 0x27, 0xea, 0xc1, 0xbb, 0x51,
	0xe9, 0x57, 0x04, 0xf2, 0xac, 0xd4, 0x66, 0x9d, 0xc6, 0x59, 0xbc, 0x4e, 0x52, 0xab, 0xda, 0xab,
	0x9e, 0xd5, 0xc4, 0x71, 0x8e, 0x4f, 0x4b, 0x98, 0x5a, 0xa4, 0xf1, 0x26, 0x4e, 0x56, 0x71, 0xa2,
	0xad, 0x43, 0x13, 0xa7, 0x54, 0x4f, 0xff, 0x55, 0x21, 0x0f, 0x66, 0xd1, 0x6a, 0xa5, 0xb3, 0x60,
	0x1b, 0x25, 0x69, 0xb4, 0xc0, 0x07, 0xd3, 0x3e, 0xed, 0x92, 0x83, 0x78, 0x69, 0x52, 0x6c, 0x89,
	0x83, 0x78, 0x89, 0x84, 0x44, 0x37, 0x77, 0x95, 0x57, 0x45, 0xa1, 0xd1, 0x6f, 0xc8, 0x21, 0xd2,
	0x69, 0x8a, 0xee, 0x9e, 0xf7, 0xca, 0x62, 0xde, 0x13, 0xf2, 0x59, 0x70, 0xbb, 0xd1, 0xc2, 0x78,
	0xd3, 0xa7, 0xa4, 0xb9, 0xd8, 0x6d, 0xb7, 0x3a, 0x59, 0xdc, 0x9a, 0x84, 0xba, 0xe7, 0xed, 0xf2,
	0xe5, 0x60, 0x71, 0x2b, 0xee, 0x8c, 0xa7, 0xbf, 0x25, 0x87, 0xf8, 0x8c, 0xb6, 0x49, 0xc3, 0xe6,
	0x43, 0x16, 0xba, 0x01, 0x7c, 0x44, 0x5b, 0xa4, 0x36, 0x63, 0x23, 0x2e, 0xa0, 0x42, 0x09, 0xa9,
	0x4f, 0xd9, 0xdc, 0x0f, 0x03, 0x38, 0xa0, 0x1d, 0xd2, 0xe4, 0x9e, 0x2d, 0xfc, 0xd0, 0xb3, 0xa1,
	0x7a, 0xfa, 0xcf, 0x63, 0xd2, 0x1e, 0x45, 0x37, 0x3a, 0xcd, 0xa2, 0x0c, 0x79, 0x7f, 0x4e, 0x1a,
	0xd7, 0xd1, 0x8d, 0x56, 0x45, 0x55, 0xdd, 0xf3, 0xc7, 0xe5, 0x17, 0xf7, 0xbc, 0x8c, 0xec, 0xd8,
	0xa2, 0x8e, 0xae, 0xce, 0x92, 0x7e, 0x4a, 0x48, 0xee, 0xa4, 0x96, 0xfa, 0xaa, 0x68, 0x41, 0x2b,
	0x47, 0x6c, 0x7d, 0x45, 0x7b, 0xa4, 0x73, 0xa9, 0x33, 0xb5, 0xd1, 0x5b, 0x65, 0xb8, 0xad, 0x1a,
	0x6a, 0xc8, 0xa5, 0xce, 0xa6, 0x7a, 0xeb, 0xc6, 0xc9, 0xff, 0x5f, 0x28, 0xfd, 0x35, 0xa1, 0x9b,
	0xad, 0x7e, 0x1d, 0xaf, 0x77, 0xa9, 0xba, 0x2e, 0x13, 0xb2, 0x6a, 0xbd, 0xca, 0x59, 0x47, 0x9c,
	0x94, 0x96, 0xbb, 0x4c, 0xe9, 0x97, 0xa4, 0x9b, 0xe8, 0x3f, 0x67, 0x7b, 0xae, 0x75, 0xe3, 0x7a,
	0x84, 0xe8, 0x5b, 0xb7, 0x73, 0x52, 0xcf, 0xe9, 0xb7, 0x1a, 0x1f, 0xae, 0x99, 0x19, 0x0f, 0x51,
	0x78, 0xd2, 0xef, 0x48, 0xbb, 0x18, 0xab, 0xeb, 0x6d, 0xbc, 0xb4, 0x9a, 0xbd, 0xea, 0x59, 0xfb,
	0xdc, 0x7a, 0xdf, 0x43, 0xa1, 0xf5, 0x4a, 0x90, 0xdc, 0x79, 0xb4, 0x8d, 0x97, 0xf4, 0x29, 0xa9,
	0x6f, 0x70, 0xcc, 0x53, 0xab, 0x65, 0x5e, 0x1d, 0xdf, 0xcd, 0x76, 0x3e, 0xfc, 0xa2, 0x30, 0xd3,
	0x5f, 0x11, 0x9a, 0xea, 0x95, 0x5e, 0x64, 0x7a, 0xa9, 0xde, 0xc4, 0x89, 0x61, 0x2f, 0xb5, 0x88,
	0x19, 0x5e, 0x28, 0x2d, 0xb3, 0x38, 0x41, 0x0e, 0x53, 0xfa, 0x94, 0x1c, 0x6f, 0xf5, 0x2a, 0xca,
	0xe2, 0xd7, 0x5a, 0x15, 0xbb, 0xd3, 0x36, 0xad, 0xe8, 0x96, 0xf0, 0xf4, 0x7d, 0x1b, 0xd3, 0xf9,
	0xd1, 0xc6, 0x3c, 0x21, 0xad, 0x34, 0x5b, 0x6f, 0xd4, 0x2a, 0x4e, 0x33, 0xeb, 0xc8, 0x7c, 0xad,
	0x89, 0x80, 0x1b, 0xa7, 0x19, 0xfd, 0x1d, 0xe9, 0x18, 0x4a, 0x73, 0x1a, 0x52, 0xab, 0xdb, 0xab,
	0xfe, 0x0f, 0xc6, 0xda, 0xe8, 0x9f, 0xcb, 0x29, 0xb5, 0x49, 0xe7, 0x3a, 0xba, 0x89, 0xaf, 0xe2,
	0x45, 0x64, 0x08, 0x3f, 0x36, 0xdb, 0xdd, 0xfb, 0xc0, 0x90, 0xdd, 0xf9, 0x89, 0x7b, 0xaf, 0xb0,
	0xaf, 0x8b, 0xdd, 0xcd, 0xae, 0x28, 0xf6, 0x4d, 0x9c, 0x58, 0x60, 0x66, 0xea, 0xe8, 0x2d, 0x3a,
	0x8b, 0x13, 0xfa, 0x05, 0x39, 0xda, 0xac, 0xa2, 0x5b, 0x95, 0xea, 0x3f, 0xed, 0x74, 0xb2, 0xd0,
	0xd6, 0x89, 0xa9, 0xb5, 0x83, 0xa0, 0x2c, 0x30, 0xfa, 0x7b, 0xd2, 0xc9, 0xde, 0x2e, 0x60, 0x6a,
	0x51, 0xd3, 0x93, 0x27, 0xff, 0x65, 0x45, 0xc5, 0xbd, 0x07, 0xb8, 0xf3, 0x8b, 0xd5, 0x3a, 0xd5,
	0x4b, 0xeb, 0x41, 0xaf, 0x72, 0xd6, 0x14, 0x85, 0xf6, 0xb8, 0x47, 0x0e, 0xb1, 0xf5, 0x78, 0x55,
	0xf2, 0xe6, 0xa7, 0x56, 0xc5, 0x90, 0x59, 0xaa, 0x8f, 0xff, 0x5e, 0x21, 0x9d, 0xfd, 0x2a, 0xf1,
	0x6e, 0xae, 0xf4, 0x6b, 0x5d, 0x9e, 0xd3, 0x5c, 0x41, 0x34, 0xcd, 0xa2, 0x6b, 0x5d, 0x5e, 0x53,
	0xa3, 0xe4, 0xed, 0xbe, 0x89, 0xe2, 0x24, 0x4e, 0xae, 0x55, 0xba, 0x89, 0xcd, 0x59, 0x2b, 0xda,
	0x5d, 0xc0, 0x12, 0x51, 0xfa, 0x15, 0x39, 0x31, 0x66, 0x95, 0xad, 0x95, 0x79, 0xaa, 0x76, 0x9b,
	0xe2, 0xbe, 0x75, 0x8d, 0x21, 0x58, 0x4b, 0x84, 0xc3, 0x0d, 0xfd, 0x9c, 0xb4, 0xb3, 0x75, 0x16,
	0xad, 0x8a, 0x78, 0xb5, 0x7c, 0x34, 0x0c, 0x64, 0x62, 0x9d, 0xfe, 0xbb, 0x46, 0xea, 0xf9, 0xf2,
	0xdf, 0xbf, 0x35, 0x0f, 0xc8, 0x71, 0x30, 0xe6, 0x6a, 0xce, 0x99, 0x50, 0xfe, 0x50, 0xbd, 0x1a,
	0x87, 0xf9, 0xd5, 0x79, 0xe5, 0xdb, 0x0e, 0x1b, 0xc0, 0x01, 0x3d, 0x22, 0xad, 0x01, 0x0b, 0x54,
	0x30, 0x76, 0xf8, 0x10, 0xaa, 0xf4, 0x63, 0x72, 0x12, 0x8c, 0x05, 0xe7, 0xea, 0xc2, 0xf1, 0x46,
	0xb6, 0x3f, 0x51, 0x72, 0x1c, 0xc2, 0xe1, 0x8f, 0xe1, 0x19, 0x77, 0xa0, 0x46, 0x1f, 0x12, 0x78,
	0x07, 0x0e, 0xa1, 0x4e, 0x1f, 0x11, 0x3a, 0x10, 0xce, 0x44, 0xfa, 0x9e, 0x9a, 0x30, 0xf9, 0x7d,
	0xc8, 0x05, 0xb3, 0x39, 0x34, 0xe8, 0x31, 0x69, 0xbf, 0x08, 0xbd, 0x91, 0xcb, 0x95, 0x64, 0x23,
	0x06, 0x4d, 0x04, 0x06, 0xcc, 0xb3, 0xe7, 0x6a, 0xe4, 0x08, 0x57, 0x42, 0x0b, 0xe3, 0xcd, 0xc2,
	0x0b, 0xdf, 0x1b, 0xa9, 0x40, 0x70, 0x26, 0x43, 0xc1, 0x25, 0x10, 0x0a, 0xa4, 0x23, 0x03, 0xc1,
	0x79, 0xa0, 0x04, 0x1b, 0x70, 0x01, 0x6d, 0x44, 0xfa, 0xac, 0xcf, 0x2e, 0xa4, 0xfa, 0x3e, 0xe4,
	0x32, 0x80, 0x0e, 0xed, 0x12, 0xc2, 0x64, 0x20, 0x7c, 0x35, 0xe2, 0x13, 0x09, 0x47, 0x78, 0x63,
	0xa7, 0xcc, 0xb3, 0x19, 0x74, 0x31, 0x28, 0xa6, 0x87, 0xe5, 0x8f, 0xd8, 0xa4, 0xef, 0x72, 0x21,
	0xe1, 0x18, 0x89, 0x61, 0x22, 0x40, 0x10, 0xf9, 0x19, 0x3a, 0x32, 0x00, 0xa0, 0x94, 0x74, 0x25,
	0x67, 0xd2, 0xf7, 0xa4, 0x9a, 0x39, 0x5e, 0xc0, 0x05, 0x9c, 0xec, 0x63, 0x72, 0x2a, 0x1c, 0x6f,
	0x04, 0xf4, 0x1e, 0x16, 0x4e, 0x26, 0x5c, 0xc0, 0x83, 0x7d, 0x8c, 0x85, 0x41, 0x38, 0xf1, 0xe0,
	0x21, 0xb6, 0xa2, 0xc0, 0xe0, 0x63, 0x4c, 0x7a, 0x28, 0x42, 0x27, 0x98, 0xab, 0x97, 0x5c, 0x48,
	0x0e, 0x8f, 0x10, 0x61, 0x6a, 0xc8, 0x1c, 0x31, 0x57, 0x01, 0x73, 0x39, 0x7c, 0x62, 0xb2, 0x52,
	0x63, 0xc7, 0xb6, 0xb9, 0xa7, 0x86, 0xbe, 0xc0, 0xda, 0x2c, 0xac, 0xad, 0xcf, 0x82, 0xc0, 0xe5,
	0x13, 0x3e, 0x18, 0xc3, 0x4f, 0xb0, 0x20, 0xa6, 0x26, 0x8e, 0xeb, 0x3a, 0xbe, 0xa7, 0x5c, 0x67,
	0x34, 0x0e, 0x24, 0x3c, 0xc6, 0xa6, 0xf6, 0x1d, 0xa4, 0x00, 0x9e, 0xa0, 0x3c, 0x70, 0xfd, 0xd0,
	0xfe, 0x0e, 0x7e, 0x8a, 0xde, 0x2e, 0xc3, 0x5a, 0x3c, 0x35, 0xe4, 0x32, 0x70, 0x5e, 0x32, 0x17,
	0x3e, 0xa5, 0x4f, 0xc8, 0x27, 0x4c, 0xed, 0x91, 0xaf, 0x06, 0x63, 0xe1, 0xc8, 0x60, 0xc2, 0x24,
	0x7c, 0x86, 0x1f, 0x94, 0x17, 0x73, 0xf5, 0x82, 0xcf, 0xb8, 0x2b, 0xe1, 0x73, 0xcc, 0x73, 0xca,
	0x99, 0x70, 0x91, 0xa6, 0x31, 0x17, 0xd0, 0xa3, 0x4d, 0x72, 0x38, 0xf2, 0x99, 0x0b, 0x3f, 0x33,
	0xd3, 0xc6, 0xe6, 0x81, 0xef, 0x31, 0x38, 0xcd, 0x33, 0x2b, 0x67, 0xcd, 0x65, 0x3e, 0x0e, 0xcf,
	0x17, 0xd8, 0x66, 0x5b, 0xb0, 0x11, 0x8e, 0xc3, 0x5c, 0x06, 0xf0, 0x73, 0x74, 0x1b, 0xf8, 0xfe,
	0x85, 0x3f, 0x1c, 0xaa, 0xc1, 0x98, 0x4d, 0xa6, 0x8e, 0xef, 0xc1, 0x97, 0x6f, 0xa7, 0x41, 0x4e,
	0x98, 0x1c, 0xc3, 0x2f, 0x90, 0x8c, 0x97, 0xcc, 0x75, 0xf9, 0x1c, 0xa3, 0x61, 0x0b, 0x25, 0x3c,
	0x3d, 0xfd, 0x47, 0x85, 0xd4, 0xf3, 0x9b, 0x85, 0x49, 0x5c, 0x46, 0xa9, 0x86, 0x8f, 0xb0, 0xde,
	0xab, 0x38, 0x89, 0xd3, 0x3f, 0x40, 0x05, 0x7f, 0x46, 0xaf, 0xb6, 0x5a, 0xe3, 0xa2, 0xe4, 0xe3,
	0xbd, 0x89, 0x17, 0x7f, 0x34, 0x4b, 0x02, 0x55, 0x74, 0xdc, 0xe6, 0xa6, 0x43, 0x34, 0x95, 0x8e,
	0xe7, 0x50, 0xdb, 0x57, 0x9f, 0x43, 0x7d, 0x5f, 0xfd, 0x06, 0x1a, 0xfb, 0xea, 0xb7, 0xd0, 0x44,
	0x86, 0x4a, 0xf5, 0xeb, 0xdf, 0x40, 0x6b, 0x5f, 0x3f, 0xff, 0x16, 0x08, 0xb2, 0xb2, 0x88, 0xd2,
	0x45, 0xb4, 0xd4, 0xd0, 0xfe, 0xe5, 0x5f, 0x2b, 0xa4, 0x3a, 0x58, 0xdc, 0xde, 0x5f, 0xcc, 0x06,
	0xa9, 0x0e, 0xbc, 0x39, 0x54, 0x50, 0x08, 0xa5, 0x0d, 0x07, 0x28, 0x8c, 0xfa, 0x53, 0xa8, 0xa2,
	0xc0, 0x43, 0x01, 0x87, 0x28, 0xbc, 0x98, 0xce, 0xa1, 0x8e, 0x42, 0x30, 0xee, 0x43, 0x03, 0x85,
	0xc9, 0x5c, 0x40, 0x13, 0x85, 0x97, 0x9e, 0x0d, 0x2d, 0x14, 0x2e, 0xc4, 0x0c, 0x08, 0x0a, 0x8e,
	0x8d, 0xcb, 0xd1, 0x20, 0xd5, 0x57, 0x4c, 0x40, 0x07, 0x85, 0x1f, 0xfa, 0x01, 0x1c, 0x99, 0xe7,
	0x62, 0x0e, 0xdd, 0xcb, 0xba, 0xf9, 0xc7, 0xf8, 0xfc, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x3a,
	0x0d, 0x25, 0xc9, 0x47, 0x0a, 0x00, 0x00,
}
