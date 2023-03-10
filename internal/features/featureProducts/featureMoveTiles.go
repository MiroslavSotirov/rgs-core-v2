package featureProducts

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
)

const (
	FEATURE_ID_MOVE_TILES = "MoveTiles"

	PARAM_ID_MOVE_TILES_POSITIONS    = "Positions"
	PARAM_ID_MOVE_TILES_DESTINATIONS = "Destinations"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_MOVE_TILES, func() feature.Feature { return new(MoveTiles) })

type MoveTilesData struct {
	Positions    []int `json:"positions"`
	Destinations []int `json:"destinations"`
}

type MoveTiles struct {
	feature.Base
	Data MoveTilesData `json:"data"`
}

func (f *MoveTiles) DataPtr() interface{} {
	return &f.Data
}

func (f MoveTiles) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	positions := params.GetIntSlice(PARAM_ID_MOVE_TILES_POSITIONS)
	destinations := params.GetIntSlice(PARAM_ID_MOVE_TILES_DESTINATIONS)

	gridh := len(state.SymbolGrid[0])
	for i, p := range positions {
		sx := p / gridh
		sy := p % gridh
		dx := destinations[i] / gridh
		dy := destinations[i] % gridh
		state.SymbolGrid[dx][dy] = state.SymbolGrid[sx][sy]
	}

	//	replaceid := params.GetInt(PARAM_ID_REPLACE_TILE_REPLACE_WITH_ID)
	state.Features = append(state.Features,
		&MoveTiles{
			Base: feature.Base{FeatureDef: *f.DefPtr()},
			Data: MoveTilesData{
				Positions:    positions,
				Destinations: destinations,
			},
		})
}

func (f *MoveTiles) Serialize() ([]byte, error) {
	return feature.SerializeFeatureToBytes(f)
}

func (f *MoveTiles) Deserialize(data []byte) (err error) {
	err = feature.DeserializeFeatureFromBytes(f, data)
	return
}
