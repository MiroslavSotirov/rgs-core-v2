package featureProducts

import "gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"

const (
	FEATURE_ID_EXPANDING_WILD = "ExpandingWild"

	PARAM_ID_EXPANDING_WILD_TILE_ID   = "TileId"
	PARAM_ID_EXPANDING_WILD_POSITION  = "Position"
	PARAM_ID_EXPANDING_WILD_POSITIONS = "Positions"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_EXPANDING_WILD, func() feature.Feature { return new(ExpandingWild) })

type ExpandingWildData struct {
	Positions []int `json:"positions"`
	TileId    int   `json:"tileid"`
	From      int   `json:"from"`
}

type ExpandingWild struct {
	feature.Base
	Data ExpandingWildData `json:"data"`
}

func (f *ExpandingWild) DataPtr() interface{} {
	return &f.Data
}

func (f ExpandingWild) forceActivateFeature(featurestate *feature.FeatureState) {
	featurestate.SymbolGrid[0][0] = f.FeatureDef.Params.GetInt("TileId")
}

func (f ExpandingWild) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	tileid := params.GetInt(PARAM_ID_EXPANDING_WILD_TILE_ID)
	position := params.GetInt(PARAM_ID_EXPANDING_WILD_POSITION)
	positions := params.GetIntSlice(PARAM_ID_EXPANDING_WILD_POSITIONS)
	gridh := len(state.SymbolGrid[0])
	for _, p := range positions {
		x := p / gridh
		y := p - (x * gridh)
		state.SymbolGrid[x][y] = tileid
	}
	state.Features = append(state.Features,
		&ExpandingWild{
			Base: feature.Base{FeatureDef: *f.DefPtr()},
			Data: ExpandingWildData{
				Positions: positions,
				TileId:    tileid,
				From:      position,
			},
		})
}

func (f *ExpandingWild) Serialize() ([]byte, error) {
	return feature.SerializeFeatureToBytes(f)
}

func (f *ExpandingWild) Deserialize(data []byte) (err error) {
	err = feature.DeserializeFeatureFromBytes(f, data)
	return
}
