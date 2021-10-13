package features

type ExpandingWildData struct {
	Positions []int `json:"positions"`
	TileId    int   `json:"tileid"`
	From      int   `json:"from"`
}

type ExpandingWild struct {
	FeatureDef
	Data ExpandingWildData `json:"data"`
}

func (f *ExpandingWild) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *ExpandingWild) DataPtr() interface{} {
	return &f.Data
}

func (f *ExpandingWild) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f ExpandingWild) forceActivateFeature(featurestate *FeatureState) {
	featurestate.SymbolGrid[0][0] = f.FeatureDef.Params.GetInt("TileId")
}

func (f ExpandingWild) Trigger(state *FeatureState, params FeatureParams) {
	tileid := params.GetInt("TileId")
	position := params.GetInt("Position")
	positions := params.GetIntSlice("Positions")
	gridh := len(state.SymbolGrid[0])
	for _, p := range positions {
		x := p / gridh
		y := p - (x * gridh)
		state.SymbolGrid[x][y] = tileid
	}
	state.Features = append(state.Features,
		&ExpandingWild{
			FeatureDef: *f.DefPtr(),
			Data: ExpandingWildData{
				Positions: positions,
				TileId:    tileid,
				From:      position,
			},
		})
}

func (f *ExpandingWild) Serialize() ([]byte, error) {
	return serializeFeatureToBytes(f)
}

func (f *ExpandingWild) Deserialize(data []byte) (err error) {
	err = deserializeFeatureFromBytes(f, data)
	return
}
