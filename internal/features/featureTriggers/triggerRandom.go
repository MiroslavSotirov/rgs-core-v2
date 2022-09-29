package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

const (
	FEATURE_ID_TRIGGER_RANDOM = "TriggerRandom"

	PARAM_ID_TRIGGER_RANDOM_PROBABILITY = "Probability"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_RANDOM, func() feature.Feature { return new(TriggerRandom) })

type TriggerRandom struct {
	feature.Base
}

func (f TriggerRandom) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	probability := params.GetInt(PARAM_ID_TRIGGER_RANDOM_PROBABILITY)
	rand := rng.RandFromRange(10000)
	if rand < probability {
		feature.ActivateFeatures(f.FeatureDef, state, params)
	}
}

func (f *TriggerRandom) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerRandom) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
