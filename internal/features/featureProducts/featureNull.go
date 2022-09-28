package featureProducts

import "gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"

const (
	FEATURE_ID_FEATURE_NULL = "FeatureNull"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_FEATURE_NULL, func() feature.Feature { return new(FeatureNull) })

type FeatureNull struct {
	feature.Base
	Data feature.FeatureParams `json:"data"`
}

func (f *FeatureNull) DataPtr() interface{} {
	return &f.Data
}

func (f FeatureNull) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
}

func (f *FeatureNull) Serialize() ([]byte, error) {
	return feature.SerializeFeatureToBytes(f)
}

func (f *FeatureNull) Deserialize(data []byte) (err error) {
	return feature.DeserializeFeatureFromBytes(f, data)
}
