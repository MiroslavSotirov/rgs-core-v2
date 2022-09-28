package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

const (
	FEATURE_ID_TRIGGER_SWORD_KING_RANDOM_WILDS = "TriggerSwordKingRandomWilds"

	PARAM_ID_TRIGGER_SWORD_KING_RANDOM_WILDS_PROBABILITY        = "Probability"
	PARAM_ID_TRIGGER_SWORD_KING_RANDOM_WILDS_NUM_WILDS          = "NumWilds"
	PARAM_ID_TRIGGER_SWORD_KING_RANDOM_WILDS_NUM_PROBABILITIES  = "NumProbabilities"
	PARAM_ID_TRIGGER_SWORD_KING_RANDOM_WILDS_RETRY_FACTOR       = "RetryFactor"
	PARAM_ID_TRIGGER_SWORD_KING_RANDOM_WILDS_REEL_PROBABILITIES = "ReelProbabilities"
	PARAM_ID_TRIGGER_SWORD_KING_RANDOM_WILDS_ROW_PROBABILITIES  = "RowProbabilities"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_SWORD_KING_RANDOM_WILDS, func() feature.Feature { return new(TriggerSwordKingRandomWilds) })

type TriggerSwordKingRandomWilds struct {
	feature.Base
}

func (f *TriggerSwordKingRandomWilds) DataPtr() interface{} {
	return nil
}

func (f TriggerSwordKingRandomWilds) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	Probability := params.GetInt(PARAM_ID_TRIGGER_SWORD_KING_RANDOM_WILDS_PROBABILITY)
	if rng.RandFromRange(10000) < Probability {

		numWilds := params.GetIntSlice(PARAM_ID_TRIGGER_SWORD_KING_RANDOM_WILDS_NUM_WILDS)[feature.WeightedRandomIndex(
			params.GetIntSlice(PARAM_ID_TRIGGER_SWORD_KING_RANDOM_WILDS_NUM_PROBABILITIES))]
		positions := []int{}
		gridh := len(state.SymbolGrid[0])

		for tries := numWilds * params.GetInt(PARAM_ID_TRIGGER_SWORD_KING_RANDOM_WILDS_RETRY_FACTOR); numWilds > 0 && tries > 0; tries-- {
			reel := feature.WeightedRandomIndex(params.GetIntSlice(PARAM_ID_TRIGGER_SWORD_KING_RANDOM_WILDS_REEL_PROBABILITIES))
			row := feature.WeightedRandomIndex(params.GetIntSlice(PARAM_ID_TRIGGER_SWORD_KING_RANDOM_WILDS_ROW_PROBABILITIES))
			pos := reel*gridh + row
			if func() bool {
				for _, p := range positions {
					if p == pos {
						return true
					}
				}
				return false
			}() {
				continue
			}
			numWilds--
			positions = append(positions, pos)
		}

		if len(positions) > 0 {
			params[PARAM_ID_TRIGGER_SWORD_KING_RUN_WILDS] = true
			params[featureProducts.PARAM_ID_REPLACE_TILE_POSITIONS] = positions
			feature.ActivateFeatures(f.FeatureDef, state, params)
		}
	}

	return
}

func (f *TriggerSwordKingRandomWilds) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerSwordKingRandomWilds) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
