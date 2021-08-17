package features

type InstaWinData struct {
	Type     string `json:"type"`
	SourceId int32  `json:"sourceid"`
	Amount   int    `json:"amount"`
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

func (f InstaWin) Trigger(featurestate *FeatureState, params FeatureParams) {
	multiplier := params.GetInt("InstaWinAmount")
	featurestate.Features = append(featurestate.Features,
		&InstaWin{
			FeatureDef: *f.DefPtr(),
			Data: InstaWinData{
				Type:     params.GetString("InstaWinType"),
				SourceId: params.GetInt32("InstaWinSourceId"),
				Amount:   multiplier,
			},
		})
	gridh := len(featurestate.SymbolGrid[0])
	x := params.GetInt("X")
	y := params.GetInt("Y")
	tileid := params.GetInt("TileId")
	symbolpositions := []int{x*gridh + y, (x+1)*gridh + y, x*gridh + y + 1, (x+1)*gridh + y + 1}
	featurestate.Wins = append(featurestate.Wins,
		FeatureWin{
			Multiplier:      multiplier,
			Symbols:         []int{tileid},
			SymbolPositions: symbolpositions,
		})
}

func (f *InstaWin) Serialize() ([]byte, error) {
	return serializeFeatureToBytes(f)
}

func (f *InstaWin) Deserialize(data []byte) (err error) {
	return deserializeFeatureFromBytes(f, data)
}
