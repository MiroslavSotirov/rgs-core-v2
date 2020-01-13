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
	// 1310 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x56, 0xdb, 0x92, 0xda, 0x46,
	0x10, 0x35, 0xcb, 0x7d, 0xb8, 0x6c, 0xef, 0xf8, 0xa6, 0xac, 0x63, 0x87, 0xe0, 0x4a, 0x79, 0x93,
	0x4a, 0xb6, 0x62, 0x6c, 0x57, 0xc5, 0x0f, 0x79, 0x18, 0xd0, 0x00, 0xf2, 0x0a, 0x09, 0x8f, 0x24,
	0x63, 0xfc, 0x32, 0xa5, 0x85, 0x59, 0xa2, 0x0a, 0x08, 0x82, 0xc4, 0x3a, 0xe4, 0x07, 0xf2, 0x0b,
	0xc9, 0x73, 0x7e, 0x26, 0x5f, 0x90, 0xef, 0x49, 0xcd, 0x08, 0xad, 0x59, 0x5f, 0x2a, 0x79, 0xeb,
	0x3e, 0xd3, 0x3d, 0xd3, 0xe7, 0x74, 0xb7, 0x00, 0xe1, 0xa9, 0x1f, 0xfb, 0xf1, 0x76, 0x25, 0x22,
	0x7e, 0xf9, 0xf8, 0x74, 0xb5, 0x5e, 0xc6, 0x4b, 0x5c, 0x10, 0xe1, 0x2c, 0x08, 0x45, 0xf3, 0x35,
	0x2a, 0x0d, 0xfd, 0xed, 0x72, 0x13, 0x0f, 0xdb, 0xf8, 0x0e, 0x2a, 0x44, 0xdb, 0xc5, 0xf9, 0x72,
	0xae, 0x65, 0x1a, 0x99, 0x93, 0x3c, 0xdb, 0x79, 0xf8, 0x16, 0xca, 0x4f, 0x96, 0x9b, 0x30, 0xd6,
	0x0e, 0x14, 0x9c, 0x38, 0xf8, 0x01, 0x42, 0x8b, 0xcd, 0x3c, 0x0e, 0x56, 0xf3, 0x40, 0xac, 0xb5,
	0xac, 0x3a, 0xda, 0x43, 0x9a, 0x7f, 0x64, 0x50, 0x71, 0xb8, 0x0e, 0x7e, 0x13, 0xc3, 0x36, 0x3e,
	0x41, 0x85, 0x95, 0x7a, 0x45, 0xdd, 0x5c, 0x69, 0xc1, 0x69, 0xf2, 0xfc, 0x69, 0xfa, 0x36, 0xdb,
	0x9d, 0xbf, 0x77, 0xeb, 0xc1, 0xfb, 0xb7, 0xe2, 0xaf, 0x11, 0x24, 0x55, 0xf1, 0xd5, 0x32, 0x0a,
	0xe2, 0x60, 0x19, 0x46, 0x5a, 0xb6, 0x91, 0x3d, 0xc9, 0xb3, 0xc3, 0x04, 0x1f, 0xa6, 0x30, 0xd6,
	0x50, 0xf1, 0x6d, 0x10, 0xce, 0x83, 0x50, 0x68, 0x39, 0x75, 0x4f, 0xea, 0x36, 0xff, 0xc9, 0xa0,
	0x9b, 0x23, 0x7f, 0x3e, 0x17, 0xb1, 0xbb, 0xf6, 0xc3, 0xc8, 0x9f, 0xc8, 0x84, 0x61, 0x1b, 0xd7,
	0xd1, 0x41, 0x30, 0x55, 0x25, 0x96, 0xd9, 0x41, 0x30, 0x95, 0x82, 0xf8, 0x8b, 0x2b, 0xe6, 0x59,
	0xb6, 0xf3, 0xf0, 0x53, 0x94, 0x93, 0x72, 0x2a, 0xd2, 0xf5, 0x56, 0x23, 0x25, 0xf3, 0x91, 0x2b,
	0x4f, 0xdd, 0xed, 0x4a, 0x30, 0x15, 0x8d, 0x1f, 0xa1, 0xd2, 0x64, 0xb3, 0x5e, 0x8b, 0x70, 0xb2,
	0x55, 0x05, 0xd5, 0x5b, 0x95, 0x34, 0xb3, 0x33, 0xd9, 0xb2, 0xab, 0xc3, 0xe6, 0x0f, 0x28, 0x27,
	0xd3, 0x70, 0x05, 0x15, 0x75, 0xda, 0x25, 0x9e, 0xe9, 0xc2, 0x0d, 0x5c, 0x46, 0xf9, 0x11, 0xe9,
	0x51, 0x06, 0x19, 0x8c, 0x50, 0x61, 0x48, 0xc6, 0xb6, 0xe7, 0xc2, 0x01, 0xae, 0xa2, 0x12, 0xb5,
	0x74, 0x66, 0x7b, 0x96, 0x0e, 0xd9, 0xe6, 0xdf, 0x35, 0x54, 0xe9, 0xf9, 0x0b, 0x11, 0xc5, 0x7e,
	0x2c, 0x75, 0x7f, 0x82, 0x8a, 0x33, 0x7f, 0x21, 0xf8, 0x8e, 0x55, 0xbd, 0x75, 0x9c, 0xbe, 0xb8,
	0x17, 0xa5, 0x6c, 0x43, 0x67, 0x05, 0x19, 0x6a, 0x4c, 0xf1, 0x7d, 0x84, 0x92, 0x20, 0x3e, 0x15,
	0x17, 0xbb, 0x16, 0x94, 0x13, 0x44, 0x17, 0x17, 0xb8, 0x81, 0xaa, 0xe7, 0x22, 0xe6, 0x2b, 0xb1,
	0xe6, 0x4a, 0xdb, 0xac, 0x92, 0x06, 0x9d, 0x8b, 0x78, 0x28, 0xd6, 0x66, 0x10, 0xfe, 0x7f, 0xa2,
	0xf8, 0x3b, 0x84, 0x57, 0x6b, 0x71, 0x19, 0x2c, 0x37, 0x11, 0x9f, 0xa5, 0x05, 0x69, 0xf9, 0x46,
	0xe6, 0xa4, 0xca, 0x8e, 0xd2, 0x93, 0xab, 0x4a, 0xf1, 0x57, 0xa8, 0x1e, 0x8a, 0x5f, 0xe3, 0xbd,
	0xd0, 0x82, 0x0a, 0xad, 0x49, 0xf4, 0x5d, 0x58, 0x0b, 0x15, 0x12, 0xf9, 0xb5, 0xe2, 0xa7, 0x39,
	0x13, 0x15, 0xc1, 0x76, 0x91, 0xf8, 0x39, 0xaa, 0xec, 0xc6, 0x6a, 0xb6, 0x0e, 0xa6, 0x5a, 0xa9,
	0x91, 0x3d, 0xa9, 0xb4, 0xb4, 0x8f, 0x25, 0x32, 0x21, 0xe6, 0x0c, 0x25, 0xc1, 0xbd, 0x75, 0x30,
	0xc5, 0x8f, 0x50, 0x61, 0x25, 0xc7, 0x3c, 0xd2, 0xca, 0x2a, 0xeb, 0xf0, 0x6a, 0xb6, 0x93, 0xe1,
	0x67, 0xbb, 0x63, 0xfc, 0x2d, 0xc2, 0x91, 0x98, 0x8b, 0x49, 0x2c, 0xa6, 0xfc, 0x6d, 0x10, 0x2a,
	0xf5, 0x22, 0x0d, 0xa9, 0xe1, 0x85, 0xf4, 0x64, 0x14, 0x84, 0x52, 0xc3, 0x08, 0x3f, 0x42, 0x87,
	0x6b, 0x31, 0xf7, 0xe3, 0xe0, 0x52, 0xf0, 0xdd, 0xee, 0x54, 0x54, 0x2b, 0xea, 0x29, 0x3c, 0xfc,
	0xd8, 0xc6, 0x54, 0x3f, 0xd8, 0x98, 0x7b, 0xa8, 0x1c, 0xc5, 0xcb, 0x15, 0x9f, 0x07, 0x51, 0xac,
	0xd5, 0xd4, 0x6b, 0x25, 0x09, 0x98, 0x41, 0x14, 0xe3, 0x1f, 0x51, 0x55, 0x49, 0x9a, 0xc8, 0x10,
	0x69, 0xf5, 0x46, 0xf6, 0x3f, 0x14, 0xab, 0xc8, 0xf8, 0xc4, 0x8e, 0xb0, 0x8e, 0xaa, 0x33, 0x7f,
	0x11, 0x5c, 0x04, 0x13, 0x5f, 0x09, 0x7e, 0xa8, 0xb6, 0xbb, 0xf1, 0x89, 0x21, 0xbb, 0x8a, 0x63,
	0xd7, 0xb2, 0x64, 0x5f, 0x27, 0x9b, 0xc5, 0x66, 0x47, 0xf6, 0x6d, 0x10, 0x6a, 0xa0, 0x66, 0xaa,
	0xf6, 0x0e, 0x1d, 0x05, 0x21, 0x7e, 0x88, 0x6a, 0xab, 0xb9, 0xbf, 0xe5, 0x91, 0xf8, 0x65, 0x23,
	0xc2, 0x89, 0xd0, 0x8e, 0x14, 0xd7, 0xaa, 0x04, 0x9d, 0x1d, 0x76, 0xdc, 0x40, 0x39, 0xd9, 0x21,
	0xb9, 0xfc, 0x49, 0x8f, 0x22, 0x2d, 0xa3, 0x38, 0xa7, 0xee, 0xf1, 0x04, 0x55, 0xf7, 0x6b, 0x91,
	0x5f, 0xb7, 0xb9, 0xb8, 0x14, 0xe9, 0x47, 0x2f, 0x71, 0x24, 0x1a, 0xc5, 0xfe, 0x4c, 0xa4, 0xdf,
	0x3c, 0xe5, 0x24, 0x4d, 0x59, 0xf8, 0x41, 0x18, 0x84, 0x33, 0x1e, 0xad, 0x02, 0xf5, 0xf1, 0xd9,
	0x35, 0x65, 0x07, 0x3b, 0x12, 0x6d, 0xfe, 0x99, 0x47, 0x85, 0x64, 0xad, 0xae, 0x6f, 0xf1, 0x4d,
	0x74, 0xe8, 0xf6, 0x29, 0x1f, 0x53, 0xc2, 0xb8, 0xdd, 0xe5, 0x6f, 0xfa, 0x5e, 0xb2, 0xcf, 0x6f,
	0x6c, 0xdd, 0x20, 0x1d, 0x38, 0xc0, 0x35, 0x54, 0xee, 0x10, 0x97, 0xbb, 0x7d, 0x83, 0x76, 0x21,
	0x8b, 0x6f, 0xa3, 0x23, 0xb7, 0xcf, 0x28, 0xe5, 0x67, 0x86, 0xd5, 0xd3, 0xed, 0x01, 0x77, 0xfa,
	0x1e, 0xe4, 0x3e, 0x84, 0x47, 0xd4, 0x80, 0x3c, 0xbe, 0x85, 0xe0, 0x3d, 0xd8, 0x83, 0x02, 0xbe,
	0x83, 0x70, 0x87, 0x19, 0x03, 0xc7, 0xb6, 0xf8, 0x80, 0x38, 0x2f, 0x3d, 0xca, 0x88, 0x4e, 0xa1,
	0x88, 0x0f, 0x51, 0xe5, 0x85, 0x67, 0xf5, 0x4c, 0xca, 0x1d, 0xd2, 0x23, 0x50, 0x92, 0x40, 0x87,
	0x58, 0xfa, 0x98, 0xf7, 0x0c, 0x66, 0x3a, 0x50, 0x96, 0xf7, 0x8d, 0xbc, 0x33, 0xdb, 0xea, 0x71,
	0x97, 0x51, 0xe2, 0x78, 0x8c, 0x3a, 0x80, 0x30, 0xa0, 0xaa, 0xe3, 0x32, 0x4a, 0x5d, 0xce, 0x48,
	0x87, 0x32, 0xa8, 0x48, 0xa4, 0x4d, 0xda, 0xe4, 0xcc, 0xe1, 0x2f, 0x3d, 0xea, 0xb8, 0x50, 0xc5,
	0x75, 0x84, 0x88, 0xe3, 0x32, 0x9b, 0xf7, 0xe8, 0xc0, 0x81, 0x9a, 0xfc, 0x7a, 0x0d, 0x89, 0xa5,
	0x13, 0xa8, 0xcb, 0x4b, 0x65, 0x79, 0x92, 0x7e, 0x8f, 0x0c, 0xda, 0x26, 0x65, 0x0e, 0x1c, 0x4a,
	0x61, 0x08, 0x73, 0x25, 0x28, 0xf5, 0xe9, 0x1a, 0x8e, 0x0b, 0x80, 0x31, 0xaa, 0x3b, 0x94, 0x38,
	0xb6, 0xe5, 0xf0, 0x91, 0x61, 0xb9, 0x94, 0xc1, 0xd1, 0x3e, 0xe6, 0x0c, 0x99, 0x61, 0xf5, 0x00,
	0x5f, 0xc3, 0xbc, 0xc1, 0x80, 0x32, 0xb8, 0xb9, 0x8f, 0x11, 0xcf, 0xf5, 0x06, 0x16, 0xdc, 0x92,
	0xad, 0xd8, 0x61, 0x70, 0x5b, 0x16, 0xdd, 0x65, 0x9e, 0xe1, 0x8e, 0xf9, 0x2b, 0xca, 0x1c, 0x0a,
	0x77, 0x24, 0x42, 0x78, 0x97, 0x18, 0x6c, 0xcc, 0x5d, 0x62, 0x52, 0xb8, 0xab, 0xaa, 0xe2, 0x7d,
	0x43, 0xd7, 0xa9, 0xc5, 0xbb, 0x36, 0x93, 0xdc, 0x34, 0xc9, 0xad, 0x4d, 0x5c, 0xd7, 0xa4, 0x03,
	0xda, 0xe9, 0xc3, 0x67, 0x92, 0x10, 0xe1, 0x03, 0xc3, 0x34, 0x0d, 0xdb, 0xe2, 0xa6, 0xd1, 0xeb,
	0xbb, 0x0e, 0x1c, 0xcb, 0xa6, 0xb6, 0x0d, 0x29, 0x01, 0xdc, 0x93, 0x76, 0xc7, 0xb4, 0x3d, 0xfd,
	0x39, 0x7c, 0x2e, 0xa3, 0x4d, 0x22, 0xb9, 0x58, 0xbc, 0x4b, 0x1d, 0xd7, 0x78, 0x45, 0x4c, 0xb8,
	0x8f, 0xef, 0xa1, 0xbb, 0x84, 0xef, 0x89, 0xcf, 0x3b, 0x7d, 0x66, 0x38, 0xee, 0x80, 0x38, 0xf0,
	0x40, 0x3e, 0xe8, 0x9c, 0x8d, 0xf9, 0x0b, 0x3a, 0xa2, 0xa6, 0x03, 0x5f, 0xc8, 0x3a, 0x87, 0x94,
	0x30, 0x53, 0xca, 0xd4, 0xa7, 0x0c, 0x1a, 0xb8, 0x84, 0x72, 0x3d, 0x9b, 0x98, 0xf0, 0xa5, 0x9a,
	0x36, 0x32, 0x76, 0x6d, 0x8b, 0x40, 0x33, 0xa9, 0x2c, 0x9d, 0x35, 0x93, 0xd8, 0x72, 0x78, 0x1e,
	0x36, 0xff, 0xca, 0xa0, 0x42, 0xb2, 0xc0, 0x32, 0xef, 0xdc, 0x8f, 0x04, 0xdc, 0x90, 0x25, 0x5e,
	0x04, 0x61, 0x10, 0xfd, 0x04, 0x19, 0xf9, 0x9b, 0x72, 0xb1, 0x16, 0x42, 0xce, 0x77, 0x32, 0x91,
	0xab, 0x60, 0xf2, 0xb3, 0x9a, 0x6b, 0xc8, 0xca, 0xc0, 0x75, 0x72, 0x94, 0x93, 0x47, 0x69, 0x60,
	0x0b, 0xf2, 0xfb, 0xee, 0x13, 0x28, 0xec, 0xbb, 0x4f, 0xa1, 0xb8, 0xef, 0x3e, 0x83, 0x92, 0x24,
	0x95, 0xba, 0x8f, 0xbf, 0x87, 0xf2, 0xbe, 0xdf, 0x7a, 0x06, 0xe8, 0x9b, 0xdf, 0x33, 0x28, 0xdb,
	0x99, 0x6c, 0xaf, 0xaf, 0x4f, 0x11, 0x65, 0x3b, 0xd6, 0x18, 0x32, 0xd2, 0xf0, 0x1c, 0x1d, 0x0e,
	0xa4, 0xd1, 0x6b, 0x0f, 0x21, 0x2b, 0x0d, 0xea, 0x31, 0xc8, 0x49, 0xe3, 0xc5, 0x70, 0x0c, 0x05,
	0x69, 0xb8, 0xfd, 0x36, 0x14, 0xa5, 0x31, 0x18, 0x33, 0x28, 0x49, 0xe3, 0x95, 0xa5, 0x43, 0x59,
	0x1a, 0x67, 0x6c, 0x04, 0x48, 0x1a, 0x86, 0x2e, 0x47, 0xb8, 0x88, 0xb2, 0x6f, 0x08, 0x83, 0xaa,
	0x34, 0x5e, 0xb7, 0x5d, 0xa8, 0xa9, 0x74, 0x36, 0x86, 0xfa, 0x79, 0x41, 0xfd, 0x63, 0x7a, 0xf2,
	0x6f, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xc1, 0x02, 0x8c, 0x47, 0x09, 0x00, 0x00,
}
