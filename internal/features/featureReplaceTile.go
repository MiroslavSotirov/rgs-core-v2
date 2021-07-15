package features

type ReplaceTileData struct {
	Positions     []int `json:"positions"`
	TileId        int   `json:"titleid"`
	ReplaceWithId int   `json:"replacewithid"`
}

type ReplaceTile struct {
	FeatureDef
	Data ReplaceTileData `json:"data"`
}

func (f *ReplaceTile) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *ReplaceTile) DataPtr() interface{} {
	return &f.Data
}

func (f *ReplaceTile) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f ReplaceTile) forceActivateFeature(featurestate *FeatureState) {
	featurestate.SymbolGrid[0][0] = f.FeatureDef.Params.GetInt("TileId")
}

func (f ReplaceTile) Trigger(state *FeatureState, params FeatureParams) {
	replaceid := params.GetInt("ReplaceWithId")
	//	featurestate.SymbolGrid[x][y] = replaceid
	positions := params.GetIntSlice("Positions")
	gridh := len(state.SymbolGrid[0])
	for _, p := range positions {
		x := p / gridh
		y := p - (x * gridh)
		state.SymbolGrid[x][y] = replaceid
	}
	state.Features = append(state.Features,
		&ReplaceTile{
			FeatureDef: *f.DefPtr(),
			Data: ReplaceTileData{
				Positions:     positions,
				TileId:        params.GetInt("TileId"),
				ReplaceWithId: params.GetInt("ReplaceWithId"),
			},
		})
}

func (f *ReplaceTile) Serialize() ([]byte, error) {
	return serializeFeatureToBytes(f)
}

func (f *ReplaceTile) Deserialize(data []byte) (err error) {
	return deserializeFeatureFromBytes(f, data)
}
