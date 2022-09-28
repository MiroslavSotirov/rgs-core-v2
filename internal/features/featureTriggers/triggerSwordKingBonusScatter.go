package featureTriggers

import (
	"fmt"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

const (
	FEATURE_ID_TRIGGER_SWORD_KING_BONUS_SCATTER = "TriggerSwordKingBonusScatter"

	PARAM_ID_TRIGGER_SWORD_KING_BONUS_SCATTER_PROBABILITY       = "Probability"
	PARAM_ID_TRIGGER_SWORD_KING_BONUS_SCATTER_NUM_SCATTERS      = "NumScatters"
	PARAM_ID_TRIGGER_SWORD_KING_BONUS_SCATTER_NUM_PROBABILITIES = "NumProbabilities"
	PARAM_ID_TRIGGER_SWORD_KING_BONUS_SCATTER_COUNTER_MAX       = "CounterMax"
	PARAM_ID_TRIGGER_SWORD_KING_BONUS_SCATTER_FSTYPE            = "FSType"
	PARAM_ID_TRIGGER_SWORD_KING_BONUS_SCATTER_NUM_FREESPINS     = "NumFreespins"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_SWORD_KING_BONUS_SCATTER, func() feature.Feature { return new(TriggerSwordKingBonusScatter) })

type TriggerSwordKingBonusScatter struct {
	feature.Base
}

func (f TriggerSwordKingBonusScatter) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	if state.PureWins ||
		(params.HasKey(PARAM_ID_TRIGGER_SWORD_KING_RUN_WILDS) && params.GetBool(PARAM_ID_TRIGGER_SWORD_KING_RUN_WILDS)) ||
		(params.HasKey(PARAM_ID_TRIGGER_SWORD_KING_RUN_RESPIN) && params.GetBool(PARAM_ID_TRIGGER_SWORD_KING_RUN_RESPIN)) {
		return
	}

	Probability := params.GetInt(PARAM_ID_TRIGGER_SWORD_KING_BONUS_SCATTER_PROBABILITY)
	if rng.RandFromRange(10000) < Probability {

		activate := false
		positions := []int{}
		if !params.HasKey(PARAM_ID_TRIGGER_SWORD_KING_BONUS_SCATTER_NUM_SCATTERS) {
			activate = true
		} else {
			numScatters := params.GetIntSlice(PARAM_ID_TRIGGER_SWORD_KING_BONUS_SCATTER_NUM_SCATTERS)[feature.WeightedRandomIndex(
				params.GetIntSlice(PARAM_ID_TRIGGER_SWORD_KING_BONUS_SCATTER_NUM_PROBABILITIES))]
			gridh := len(state.SymbolGrid[0])

			reels := feature.RandomPermutation([]int{0, 1, 2, 3, 4})

			for s := 0; s < numScatters; s++ {
				reel := reels[s]
				row := rng.RandFromRange(4)
				pos := reel*gridh + row
				positions = append(positions, pos)
			}

			if len(positions) > 0 {
				params[PARAM_ID_TRIGGER_SWORD_KING_RUN_BONUS_SCATTER] = true
				params[featureProducts.PARAM_ID_REPLACE_TILE_POSITIONS] = positions
				feature.ActivateFeatures(f.FeatureDef, state, params)

				var counter int
				statefulStake := feature.GetStatefulStakeMap(*state)
				if statefulStake.HasKey(STATEFUL_ID_TRIGGER_SWORD_KING_COUNTER) {
					counter = statefulStake.GetInt(STATEFUL_ID_TRIGGER_SWORD_KING_COUNTER)
				}
				counter += len(positions)

				counterMax := params.GetInt(PARAM_ID_TRIGGER_SWORD_KING_BONUS_SCATTER_COUNTER_MAX)
				activate = counter >= counterMax
				feature.SetStatefulStakeMap(*state, feature.FeatureParams{
					STATEFUL_ID_TRIGGER_SWORD_KING_COUNTER: counter,
				}, params)
			}
		}

		if activate {
			params[PARAM_ID_TRIGGER_SWORD_KING_RUN_BONUS] = true
			fstype := params.GetString(PARAM_ID_TRIGGER_SWORD_KING_BONUS_SCATTER_FSTYPE)
			numFreespins := params.GetInt(PARAM_ID_TRIGGER_SWORD_KING_BONUS_SCATTER_NUM_FREESPINS)

			state.Wins = append(state.Wins, feature.FeatureWin{
				Index:           fmt.Sprintf("%s:%d", fstype, numFreespins),
				SymbolPositions: positions,
			})
		}
	}

	return
}

func (f *TriggerSwordKingBonusScatter) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerSwordKingBonusScatter) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
