package features

import "gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"

type ReplaceTileData struct {
	TileId        int `json:"titleid"`
	ReplaceWithId int `json:"replacewithid"`
}

type ReplaceTile struct {
	Def  FeatureDef      `json:"def"`
	Data ReplaceTileData `json:"data"`
}

func (f *ReplaceTile) DefPtr() *FeatureDef {
	return &f.Def
}

func (f *ReplaceTile) DataPtr() interface{} {
	return &f.Data
}

func (f *ReplaceTile) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f ReplaceTile) forceActivateFeature(featurestate *FeatureState) {
	featurestate.SymbolGrid[0][0] = f.Def.Params["TileId"].(int)
}

func (f ReplaceTile) Trigger(featurestate FeatureState, params FeatureParams) []Feature {
	logger.Debugf("ReplaceTime params %v\n", params)
	return []Feature{
		&ReplaceTile{
			Def: *f.DefPtr(),
			Data: ReplaceTileData{
				TileId:        params["TileId"].(int),
				ReplaceWithId: params["ReplaceWithId"].(int),
			},
		},
	}
}

func (f *ReplaceTile) Serialize() ([]byte, error) {
	return serializeFeatureToBytes(f)
}

func (f *ReplaceTile) Deserialize(data []byte) (err error) {
	return deserializeFeatureFromBytes(f, data)
}
