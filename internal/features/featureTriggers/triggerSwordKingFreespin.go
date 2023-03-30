package featureTriggers

import (
	"fmt"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

const (
	FEATURE_ID_TRIGGER_SWORD_KING_FREESPIN = "TriggerSwordKingFreespin"

	PARAM_ID_TRIGGER_SWORD_KING_FREESPIN_PROBABILITY       = "Probability"
	PARAM_ID_TRIGGER_SWORD_KING_FREESPIN_NUM_PROBABILITIES = "NumProbabilities"
	PARAM_ID_TRIGGER_SWORD_KING_FREESPIN_NUM_SCATTERS      = "NumScatters"
	PARAM_ID_TRIGGER_SWORD_KING_FREESPIN_NUM_FREESPINS     = "NumFreespins"
	PARAM_ID_TRIGGER_SWORD_KING_FREESPIN_FSTYPE            = "FSType"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_SWORD_KING_FREESPIN, func() feature.Feature { return new(TriggerSwordKingFreespin) })

type TriggerSwordKingFreespin struct {
	feature.Base
}

func (f *TriggerSwordKingFreespin) DataPtr() interface{} {
	return nil
}

func (f TriggerSwordKingFreespin) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	if params.HasKey("PureWins") ||
		(params.HasKey(PARAM_ID_TRIGGER_SWORD_KING_RUN_WILDS) && params.GetBool(PARAM_ID_TRIGGER_SWORD_KING_RUN_WILDS)) ||
		(params.HasKey(PARAM_ID_TRIGGER_SWORD_KING_RUN_RESPIN) && params.GetBool(PARAM_ID_TRIGGER_SWORD_KING_RUN_RESPIN)) ||
		(params.HasKey(PARAM_ID_TRIGGER_SWORD_KING_RUN_BONUS_SCATTER) && params.GetBool(PARAM_ID_TRIGGER_SWORD_KING_RUN_BONUS_SCATTER)) ||
		(params.HasKey(PARAM_ID_TRIGGER_SWORD_KING_RUN_BONUS) && params.GetBool(PARAM_ID_TRIGGER_SWORD_KING_RUN_BONUS)) {
		return
	}

	Probability := params.GetInt(PARAM_ID_TRIGGER_SWORD_KING_FREESPIN_PROBABILITY)
	if rng.RandFromRangePool(10000) < Probability {
		numIdx := feature.WeightedRandomIndex(params.GetIntSlice(PARAM_ID_TRIGGER_SWORD_KING_FREESPIN_NUM_PROBABILITIES))
		numScatters := params.GetIntSlice(PARAM_ID_TRIGGER_SWORD_KING_FREESPIN_NUM_SCATTERS)[numIdx]
		positions := []int{}
		gridh := len(state.SymbolGrid[0])

		reels := feature.RandomPermutation([]int{0, 1, 2, 3, 4})

		for s := 0; s < numScatters; s++ {
			reel := reels[s]
			row := rng.RandFromRangePool(4)
			pos := reel*gridh + row
			positions = append(positions, pos)
		}

		if len(positions) > 0 {
			params[PARAM_ID_TRIGGER_SWORD_KING_RUN_FSSCATTER] = true
			params[featureProducts.PARAM_ID_REPLACE_TILE_POSITIONS] = positions
			feature.ActivateFeatures(f.FeatureDef, state, params)

			numFreespins := params.GetIntSlice(PARAM_ID_TRIGGER_SWORD_KING_FREESPIN_NUM_FREESPINS)[numIdx]
			if numFreespins > 0 {
				fstype := params.GetString(PARAM_ID_TRIGGER_SWORD_KING_FREESPIN_FSTYPE)

				state.Wins = append(state.Wins, feature.FeatureWin{
					Index:           fmt.Sprintf("%s:%d", fstype, numFreespins),
					SymbolPositions: positions,
				})
			}
		}
	}

	return
}

func (f *TriggerSwordKingFreespin) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerSwordKingFreespin) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
