package featureTriggers

import "gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"

const (
	FEATURE_ID_TRIGGER_WEIGHTED_RANDOM = "TriggerWeightedRandom"

	PARAM_ID_TRIGGER_WEIGHTED_RANDOM_WEIGHTS = "Weights"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_WEIGHTED_RANDOM, func() feature.Feature { return new(TriggerWeightedRandom) })

type TriggerWeightedRandom struct {
	feature.Base
}

func (f TriggerWeightedRandom) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	var weights []int
	if params.HasKey(PARAM_ID_TRIGGER_WEIGHTED_RANDOM_WEIGHTS) {
		weights = params.GetIntSlice(PARAM_ID_TRIGGER_WEIGHTED_RANDOM_WEIGHTS)
	} else {
		num := len(f.Features)
		weights = make([]int, num)
		for i := range weights {
			weights[i] = 1
		}
	}
	idx := feature.WeightedRandomIndex(weights)
	matchidx := func(i int, d feature.FeatureDef, s *feature.FeatureState, p feature.FeatureParams) bool {
		return i == idx
	}
	feature.ActivateFilteredFeatures(f.FeatureDef, state, params, matchidx)
}

func (f *TriggerWeightedRandom) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerWeightedRandom) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
