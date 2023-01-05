package featureProducts

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
)

const (
	FEATURE_ID_ACTIVATION = "Activation"

	PARAM_ID_ACTIVATION_PARAMS = "ActivationParams"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_ACTIVATION, func() feature.Feature { return new(Activation) })

type Activation struct {
	feature.Base
	Data feature.FeatureParams `json:"data"`
}

func (f *Activation) DataPtr() interface{} {
	return &f.Data
}

func (f Activation) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	activationParams := feature.FeatureParams{}
	if params.HasKey(PARAM_ID_ACTIVATION_PARAMS) {
		activationParams = params.GetParams(PARAM_ID_ACTIVATION_PARAMS)
	}
	state.Features = append(state.Features,
		&Activation{
			Base: feature.Base{FeatureDef: *f.DefPtr()},
			Data: activationParams,
		})
}

func (f *Activation) Serialize() ([]byte, error) {
	return feature.SerializeFeatureToBytes(f)
}

func (f *Activation) Deserialize(data []byte) (err error) {
	err = feature.DeserializeFeatureFromBytes(f, data)
	return
}
