package features

type InstaWinData struct {
	Type     string `json:"type"`
	SourceId int32  `json:"sourceid"`
	Amount   int    `json:"amount"`
	Payouts  []int  `json:"payouts,omitempty"`
}

type InstaWin struct {
	FeatureDef
	Data InstaWinData `json:"data"`
}

func (f *InstaWin) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *InstaWin) DataPtr() interface{} {
	return &f.Data
}

func (f *InstaWin) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f InstaWin) Trigger(state *FeatureState, params FeatureParams) {
	multiplier := params.GetInt("InstaWinAmount")
	var payouts []int
	if params.HasKey("Payouts") {
		payouts = params.GetIntSlice("Payouts")
	}
	state.Features = append(state.Features,
		&InstaWin{
			FeatureDef: *f.DefPtr(),
			Data: InstaWinData{
				Type:     params.GetString("InstaWinType"),
				SourceId: params.GetInt32("InstaWinSourceId"),
				Amount:   multiplier,
				Payouts:  payouts,
			},
		})
	state.Wins = append(state.Wins,
		FeatureWin{
			Multiplier:      multiplier,
			Symbols:         []int{params.GetInt("TileId")},
			SymbolPositions: params.GetIntSlice("Positions"),
		})
}

func (f *InstaWin) Serialize() ([]byte, error) {
	return serializeFeatureToBytes(f)
}

func (f *InstaWin) Deserialize(data []byte) (err error) {
	return deserializeFeatureFromBytes(f, data)
}
