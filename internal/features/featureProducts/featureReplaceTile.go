package featureProducts

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
)

const (
	FEATURE_ID_REPLACE_TILE = "ReplaceTile"

	PARAM_ID_REPLACE_TILE_REPLACE_WITH_ID  = "ReplaceWithId"
	PARAM_ID_REPLACE_TILE_REPLACE_WITH_IDS = "ReplaceWithIds"
	PARAM_ID_REPLACE_TILE_POSITIONS        = "Positions"
	PARAM_ID_REPLACE_TILE_TILE_ID          = "TileId"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_REPLACE_TILE, func() feature.Feature { return new(ReplaceTile) })

type ReplaceTileData struct {
	Positions      []int `json:"positions"`
	TileId         int   `json:"tileid"`
	ReplaceWithId  int   `json:"replacewithid"`
	ReplaceWithIds []int `json:"replacewithids,omitempty"`
}

type ReplaceTile struct {
	feature.Base
	Data ReplaceTileData `json:"data"`
}

func (f *ReplaceTile) DataPtr() interface{} {
	return &f.Data
}

func (f ReplaceTile) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	var tileid int
	var replacewithid int
	var replacewithids []int
	if params.HasKey(PARAM_ID_REPLACE_TILE_TILE_ID) {
		tileid = params.GetInt(PARAM_ID_REPLACE_TILE_TILE_ID)
	}
	positions := params.GetIntSlice(PARAM_ID_REPLACE_TILE_POSITIONS)
	if params.HasKey(PARAM_ID_REPLACE_TILE_REPLACE_WITH_ID) {
		replacewithid = params.GetInt(PARAM_ID_REPLACE_TILE_REPLACE_WITH_ID)

		gridh := len(state.SymbolGrid[0])
		for _, p := range positions {
			x := p / gridh
			y := p - (x * gridh)
			state.SymbolGrid[x][y] = replacewithid
		}

	} else {
		replacewithids = params.GetIntSlice(PARAM_ID_REPLACE_TILE_REPLACE_WITH_IDS)

		gridh := len(state.SymbolGrid[0])
		for i, p := range positions {
			x := p / gridh
			y := p - (x * gridh)
			state.SymbolGrid[x][y] = replacewithids[i]
		}

	}

	//	replaceid := params.GetInt(PARAM_ID_REPLACE_TILE_REPLACE_WITH_ID)
	state.Features = append(state.Features,
		&ReplaceTile{
			Base: feature.Base{FeatureDef: *f.DefPtr()},
			Data: ReplaceTileData{
				Positions:      positions,
				TileId:         tileid,
				ReplaceWithId:  replacewithid,
				ReplaceWithIds: replacewithids,
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
