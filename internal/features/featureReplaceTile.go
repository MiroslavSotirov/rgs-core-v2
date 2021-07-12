package features

type ReplaceTileData struct {
	X             int `json:"x"`
	Y             int `json:"y"`
	TileId        int `json:"titleid"`
	ReplaceWithId int `json:"replacewithid"`
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

func (f ReplaceTile) Trigger(featurestate *FeatureState, params FeatureParams) {
	x := params.GetInt("X")
	y := params.GetInt("Y")
	replaceid := params.GetInt("ReplaceWithId")
	featurestate.SymbolGrid[x][y] = replaceid
	featurestate.Features = append(featurestate.Features,
		&ReplaceTile{
			FeatureDef: *f.DefPtr(),
			Data: ReplaceTileData{
				X:             x,
				Y:             y,
				TileId:        params.GetInt("TileId"),
				ReplaceWithId: replaceid,
			},
		})
}

func (f *ReplaceTile) Serialize() ([]byte, error) {
	return serializeFeatureToBytes(f)
}

func (f *ReplaceTile) Deserialize(data []byte) (err error) {
	return deserializeFeatureFromBytes(f, data)
}
