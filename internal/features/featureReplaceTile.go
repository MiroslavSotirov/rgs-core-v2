package features

import (
	"encoding/json"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type ReplaceTileData struct {
	Positions     []int `json:"positions"`
	TileId        int   `json:"tileid"`
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

// remove this as soon as the duplicates problem has been tracked down
func (f ReplaceTile) Validate() {
	duplicates := make(map[int]bool)
	for _, p := range f.Data.Positions {
		_, ok := duplicates[p]
		if ok {
			b, _ := json.Marshal(f)
			logger.Debugf("broken ReplaceTile: %s", string(b))
			panic("ReplaceTile feature validation failed")
		}
		duplicates[p] = true
	}
}

func (f *ReplaceTile) Serialize() ([]byte, error) {
	//	f.Validate()
	return serializeFeatureToBytes(f)
}

func (f *ReplaceTile) Deserialize(data []byte) (err error) {
	err = deserializeFeatureFromBytes(f, data)
	//	f.Validate()
	return
}
