package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
)

const (
	FEATURE_ID_TRIGGER_SWORD_KING_BONUS = "TriggerSwordKingBonus"

	PARAM_ID_TRIGGER_SWORD_KING_BONUS_NUM_REELS          = "NumReels"
	PARAM_ID_TRIGGER_SWORD_KING_BONUS_NUM_PROBABILITIES  = "NumProbabilities"
	PARAM_ID_TRIGGER_SWORD_KING_BONUS_REEL_PROBABILITIES = "ReelProbabilities"
	PARAM_ID_TRIGGER_SWORD_KING_BONUS_REEL_POSITIONS     = "ReelPositions"
	PARAM_ID_TRIGGER_SWORD_KING_BONUS_RESET_COUNTER      = "ResetCounter"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_SWORD_KING_BONUS, func() feature.Feature { return new(TriggerSwordKingBonus) })

type TriggerSwordKingBonus struct {
	feature.Base
}

func (f *TriggerSwordKingBonus) DataPtr() interface{} {
	return nil
}

func (f TriggerSwordKingBonus) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	number := params.GetIntSlice(PARAM_ID_TRIGGER_SWORD_KING_BONUS_NUM_REELS)
	numberProbs := params.GetIntSlice(PARAM_ID_TRIGGER_SWORD_KING_BONUS_NUM_PROBABILITIES)
	numIdx := feature.WeightedRandomIndex(numberProbs)
	num := number[numIdx]

	reelProbs := feature.ConvertIntSlice(params.GetSlice(PARAM_ID_TRIGGER_SWORD_KING_BONUS_REEL_PROBABILITIES)[numIdx])
	reelsIdx := feature.WeightedRandomIndex(reelProbs)

	reels := feature.ConvertIntSlice(params.GetSlice(PARAM_ID_TRIGGER_SWORD_KING_BONUS_REEL_POSITIONS)[numIdx].([]interface{})[reelsIdx])

	gridh := len(state.SymbolGrid[0])

	positions := []int{}
	for i := 0; i < num; i++ {
		x := reels[i]
		for y := 0; y < 4; y++ {
			positions = append(positions, x*gridh+y)
		}
	}

	if params.GetBool(PARAM_ID_TRIGGER_SWORD_KING_BONUS_RESET_COUNTER) {
		feature.SetStatefulStakeMap(*state, feature.FeatureParams{
			STATEFUL_ID_TRIGGER_SWORD_KING_COUNTER: 0,
		}, params)
	}

	if len(positions) > 0 {
		params[featureProducts.PARAM_ID_REPLACE_TILE_POSITIONS] = positions
		feature.ActivateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f *TriggerSwordKingBonus) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerSwordKingBonus) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
