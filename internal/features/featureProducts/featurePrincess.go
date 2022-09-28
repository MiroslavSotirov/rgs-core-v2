package featureProducts

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
)

const (
	FEATURE_ID_PRINCESS = "Princess"

	PARAM_ID_PRINCESS_REPLACE_WITH_ID = "ReplaceWithId"
	PARAM_ID_PRINCESS_POSIIONS        = "Positions"
	PARAM_ID_PRINCESS_START_POS       = "StartPos"
	PARAM_ID_PRINCESS_END_POS         = "EndPos"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_PRINCESS, func() feature.Feature { return new(Princess) })

type PrincessData struct {
	Positions     []int `json:"positions"`
	StartPos      int   `json:"startpos"`
	EndPos        int   `json:"endpos"`
	ReplaceWithId int   `json:"replacewithid"`
}

type Princess struct {
	feature.Base
	Data PrincessData `json:"data"`
}

func (f *Princess) DataPtr() interface{} {
	return &f.Data
}

func (f Princess) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	replaceid := params.GetInt(PARAM_ID_PRINCESS_REPLACE_WITH_ID)
	positions := params.GetIntSlice(PARAM_ID_PRINCESS_POSIIONS)
	gridh := len(state.SymbolGrid[0])
	for _, p := range positions {
		x := p / gridh
		y := p - (x * gridh)
		state.SymbolGrid[x][y] = replaceid
	}
	state.Features = append(state.Features,
		&Princess{
			Base: feature.Base{FeatureDef: *f.DefPtr()},
			Data: PrincessData{
				Positions:     positions,
				StartPos:      params.GetInt(PARAM_ID_PRINCESS_START_POS),
				EndPos:        params.GetInt(PARAM_ID_PRINCESS_END_POS),
				ReplaceWithId: params.GetInt(PARAM_ID_PRINCESS_REPLACE_WITH_ID),
			},
		})
}

func (f *Princess) Serialize() ([]byte, error) {
	return feature.SerializeFeatureToBytes(f)
}

func (f *Princess) Deserialize(data []byte) (err error) {
	err = feature.DeserializeFeatureFromBytes(f, data)
	return
}
