package featureProducts

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
)

const (
	FEATURE_ID_REPLACE_TILE = "ReplaceTile"

	PARAM_ID_REPLACE_TILE_REPLACE_WITH_ID = "ReplaceWithId"
	PARAM_ID_REPLACE_TILE_POSITIONS       = "Positions"
	PARAM_ID_REPLACE_TILE_TILE_ID         = "TileId"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_REPLACE_TILE, func() feature.Feature { return new(ReplaceTile) })

type ReplaceTileData struct {
	Positions     []int `json:"positions"`
	TileId        int   `json:"tileid"`
	ReplaceWithId int   `json:"replacewithid"`
}

type ReplaceTile struct {
	feature.Base
	Data ReplaceTileData `json:"data"`
}

func (f *ReplaceTile) DataPtr() interface{} {
	return &f.Data
}

func (f ReplaceTile) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	replaceid := params.GetInt(PARAM_ID_REPLACE_TILE_REPLACE_WITH_ID)
	positions := params.GetIntSlice(PARAM_ID_REPLACE_TILE_POSITIONS)
	gridh := len(state.SymbolGrid[0])
	for _, p := range positions {
		x := p / gridh
		y := p - (x * gridh)
		state.SymbolGrid[x][y] = replaceid
	}
	state.Features = append(state.Features,
		&ReplaceTile{
			Base: feature.Base{FeatureDef: *f.DefPtr()},
			Data: ReplaceTileData{
				Positions:     positions,
				TileId:        params.GetInt(PARAM_ID_REPLACE_TILE_TILE_ID),
				ReplaceWithId: params.GetInt(PARAM_ID_REPLACE_TILE_REPLACE_WITH_ID),
			},
		})
}

func (f *ReplaceTile) Serialize() ([]byte, error) {
	return feature.SerializeFeatureToBytes(f)
}

func (f *ReplaceTile) Deserialize(data []byte) (err error) {
	err = feature.DeserializeFeatureFromBytes(f, data)
	return
}
