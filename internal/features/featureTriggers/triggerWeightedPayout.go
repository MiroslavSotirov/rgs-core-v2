package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
)

const (
	FEATURE_ID_TRIGGER_WEIGHTED_PAYOUT = "TriggerWeightedPayout"

	PARAM_ID_TRIGGER_WEIGHTED_PAYOUT_PAYOUTS = "Payouts"
	PARAM_ID_TRIGGER_WEIGHTED_PAYOUT_WEIGHTS = "Weights"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_WEIGHTED_PAYOUT, func() feature.Feature { return new(TriggerWeightedPayout) })

type TriggerWeightedPayout struct {
	feature.Base
}

func (f TriggerWeightedPayout) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	var weights []int
	var payouts []int = params.GetIntSlice(PARAM_ID_TRIGGER_WEIGHTED_PAYOUT_PAYOUTS)
	if params.HasKey(PARAM_ID_TRIGGER_WEIGHTED_PAYOUT_WEIGHTS) {
		weights = params.GetIntSlice(PARAM_ID_TRIGGER_WEIGHTED_PAYOUT_WEIGHTS)
	} else {
		weights = make([]int, len(payouts))
		for i := range weights {
			weights[i] = 1
		}
	}
	idx := feature.WeightedRandomIndex(weights)
	params[featureProducts.PARAM_ID_INSTA_WIN_TYPE] = featureProducts.PARAM_VALUE_INSTA_WIN_BONUS
	params[featureProducts.PARAM_ID_INSTA_WIN_SOURCE_ID] = f.FeatureDef.Id
	params[featureProducts.PARAM_ID_INSTA_WIN_AMOUNT] = payouts[idx]
	params[featureProducts.PARAM_ID_INSTA_WIN_POSITIONS] = []int{}

	feature.ActivateFeatures(f.FeatureDef, state, params)
}

func (f *TriggerWeightedPayout) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerWeightedPayout) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
