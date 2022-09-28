package featureProducts

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

const (
	FEATURE_ID_SET_REELS = "SetReels"

	PARAM_ID_SET_REELS_ACTION     = "Action"
	PARAM_ID_SET_REELS_REELS      = "Reels"
	PARAM_ID_SET_REELS_REELSET_ID = "ReelsetId"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_SET_REELS, func() feature.Feature { return new(SetReels) })

type SetReels struct {
	feature.Base
	Data feature.FeatureParams `json:"data"`
}

func (f *SetReels) DataPtr() interface{} {
	return &f.Data
}

func (f SetReels) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	if params.HasKey("Action") {
		if state.Action != params.GetString(PARAM_ID_SET_REELS_ACTION) {
			return
		}
	}
	paramreels := params.GetSlice(PARAM_ID_SET_REELS_REELS)
	if params.HasKey(PARAM_ID_SET_REELS_REELSET_ID) {
		state.ReelsetId = params.AsString(PARAM_ID_SET_REELS_REELSET_ID)
		logger.Debugf("Set reels using ReelsetId %s", state.ReelsetId)
	}
	reels := make([][]int, len(paramreels))
	for i, r := range paramreels {
		reels[i] = feature.ConvertIntSlice(r)
	}
	state.Reels = reels
}

func (f *SetReels) Serialize() ([]byte, error) {
	return feature.SerializeFeatureToBytes(f)
}

func (f *SetReels) Deserialize(data []byte) (err error) {
	return feature.DeserializeFeatureFromBytes(f, data)
}
