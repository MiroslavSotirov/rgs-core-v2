package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

const (
	FEATURE_ID_TRIGGER_TIPSY_CHARMS_WILDS = "TriggerTipsyCharmsWilds"

	PARAM_ID_TRIGGER_TIPSY_CHARMS_WILDS_PROBABILITY_LEVELS        = "ProbabilitiesLevels"
	PARAM_ID_TRIGGER_TIPSY_CHARMS_WILDS_NUM_WILDS_LEVELS          = "NumWildsLevels"
	PARAM_ID_TRIGGER_TIPSY_CHARMS_WILDS_NUM_PROBABILITIES_LEVELS  = "NumProbabilitiesLevels"
	PARAM_ID_TRIGGER_TIPSY_CHARMS_WILDS_WILD_LEVELS               = "WildLevels"
	PARAM_ID_TRIGGER_TIPSY_CHARMS_WILDS_WILD_PROBABILITIES_LEVELS = "WildProbabilitiesLevels"
	PARAM_ID_TRIGGER_TIPSY_CHARMS_WILDS_REEL_PROBABILITIES_LEVELS = "ReelProbabilitiesLevels"
	PARAM_ID_TRIGGER_TIPSY_CHARMS_WILDS_RETRY_FACTOR              = "RetryFactor"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_TIPSY_CHARMS_WILDS,
	func() feature.Feature { return new(TriggerTipsyCharmsWilds) })

type TriggerTipsyCharmsWilds struct {
	feature.Base
}

func (f TriggerTipsyCharmsWilds) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	probabilityLevels := params.GetIntSlice(PARAM_ID_TRIGGER_TIPSY_CHARMS_WILDS_PROBABILITY_LEVELS)
	level := 0

	if probabilityLevels[0] < rng.RandFromRange(10000) {

		numWildsLevels := feature.ConvertIntSlice(params.GetSlice(PARAM_ID_TRIGGER_TIPSY_CHARMS_WILDS_NUM_WILDS_LEVELS)[level])
		numProbabilitiesLevels := feature.ConvertIntSlice(params.GetSlice(PARAM_ID_TRIGGER_TIPSY_CHARMS_WILDS_NUM_PROBABILITIES_LEVELS)[level])
		wilds := feature.ConvertIntSlice(params.GetSlice(PARAM_ID_TRIGGER_TIPSY_CHARMS_WILDS_WILD_LEVELS)[level])
		wildProbabilities := feature.ConvertIntSlice(params.GetSlice(PARAM_ID_TRIGGER_TIPSY_CHARMS_WILDS_WILD_PROBABILITIES_LEVELS)[level])
		reelProbabilities := feature.ConvertIntSlice(params.GetSlice(PARAM_ID_TRIGGER_TIPSY_CHARMS_WILDS_REEL_PROBABILITIES_LEVELS)[level])

		numWilds := numWildsLevels[feature.WeightedRandomIndex(numProbabilitiesLevels)]
		tries := numWilds * params.GetInt(PARAM_ID_TRIGGER_TIPSY_CHARMS_WILDS_RETRY_FACTOR)
		gridh := len(state.SymbolGrid[0])
		positions := []int{}
		replaceids := []int{}
		for i := 0; i < tries && len(positions) < numWilds; i++ {
			wild := wilds[feature.WeightedRandomIndex(wildProbabilities)]
			reel := feature.WeightedRandomIndex(reelProbabilities)
			row := rng.RandFromRange(gridh)
			positions = append(positions, reel*gridh+row)
			replaceids = append(replaceids, wild)
		}

		params[featureProducts.PARAM_ID_REPLACE_TILE_POSITIONS] = positions
		params[featureProducts.PARAM_ID_REPLACE_TILE_REPLACE_WITH_IDS] = replaceids

		feature.ActivateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f *TriggerTipsyCharmsWilds) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerTipsyCharmsWilds) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
