### Engine Package Documentation

Engine configParser:
- Engine must have at least one engineDef configured, called "base"
- Engine may have as many additional configs as desires
- Any values left blank on subsequent engine defs will take values from base
- If it is desired to override these values to be blank, a sham non-zero value must be added (i.e. to override specialPayouts from base engine to not allow special payouts in feature, an empty list will not suffice. rather, add a special payout that will never occur)
- Special prize indices should follow the format "action:count", where action corresponds to the name of the engineDef that should be triggered, and count is the number of times this engine def should be run before completion of the round
- Prizes from regular payouts will be given indices based on the method indicated



Gamestate Spec
==============
```

type GamestatePB struct {
	Id                   string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Engine               string                 `protobuf:"bytes,2,opt,name=engine,proto3" json:"engine,omitempty"`
	BetPerLine           int64                  `protobuf:"varint,3,opt,name=bet_per_line,json=betPerLine,proto3" json:"bet_per_line,omitempty"`
	Currency             string                 `protobuf:"bytes,4,opt,name=currency,proto3" json:"currency,omitempty"`
	Transactions         []*WalletTransactionPB `protobuf:"bytes,5,rep,name=transactions,proto3" json:"transactions,omitempty"`
	PreviousGamestate    string                 `protobuf:"bytes,6,opt,name=previous_gamestate,json=previousGamestate,proto3" json:"previous_gamestate,omitempty"`
	NextGamestate        string                 `protobuf:"bytes,7,opt,name=next_gamestate,json=nextGamestate,proto3" json:"next_gamestate,omitempty"`
	Action               string                 `protobuf:"bytes,8,opt,name=action,proto3" json:"action,omitempty"`
	SymbolGrid           []*GamestatePB_Reel    `protobuf:"bytes,9,rep,name=symbol_grid,json=symbolGrid,proto3" json:"symbol_grid,omitempty"`
	Prizes               []*PrizePB             `protobuf:"bytes,10,rep,name=prizes,proto3" json:"prizes,omitempty"`
	SelectedWinLines     []int32                `protobuf:"varint,11,rep,packed,name=selected_win_lines,json=selectedWinLines,proto3" json:"selected_win_lines,omitempty"`
	RelativePayout       int32                  `protobuf:"varint,12,opt,name=relative_payout,json=relativePayout,proto3" json:"relative_payout,omitempty"`
	Multiplier           int32                  `protobuf:"varint,13,opt,name=multiplier,proto3" json:"multiplier,omitempty"`
	StopList             []int32                `protobuf:"varint,14,rep,packed,name=stop_list,json=stopList,proto3" json:"stop_list,omitempty"`
	NextActions          []string               `protobuf:"bytes,15,rep,name=next_actions,json=nextActions,proto3" json:"next_actions,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}
```

#### Generate protobuf
```
cd internal/engine
protoc --go_out=. *.proto
```