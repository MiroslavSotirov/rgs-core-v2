package features

import "gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"

type SetReels struct {
	FeatureDef
	Data FeatureParams `json:"data"`
}

func (f *SetReels) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *SetReels) DataPtr() interface{} {
	return &f.Data
}

func (f *SetReels) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *SetReels) OnInit(state *FeatureState) {
}

func (f SetReels) Trigger(state *FeatureState, params FeatureParams) {
	if params.HasKey("Action") {
		if state.Action != params.GetString("Action") {
			return
		}
	}
	paramreels := params.GetSlice("Reels")
	if params.HasKey("ReelsetId") {
		state.ReelsetId = params.AsString("ReelsetId")
		logger.Debugf("Set reels using ReelsetId %s", state.ReelsetId)
	}
	reels := make([][]int, len(paramreels))
	for i, r := range paramreels {
		reels[i] = convertIntSlice(r)
	}
	state.Reels = reels
}

func (f *SetReels) Serialize() ([]byte, error) {
	return serializeFeatureToBytes(f)
}

func (f *SetReels) Deserialize(data []byte) (err error) {
	return deserializeFeatureFromBytes(f, data)
}
