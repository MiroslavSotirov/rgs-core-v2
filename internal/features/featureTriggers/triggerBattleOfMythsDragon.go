package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

const (
	FEATURE_ID_TRIGGER_BATTLE_OF_MYTHS_DRAGON = "TriggerBattleOfMythsDragon"

	PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_DRAGON_NUMBER               = "Number"
	PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_DRAGON_NUMBER_PROBABILITIES = "NumberProbabilities"
	PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_DRAGON_REEL_PROBABILITIES   = "ReelProbabilities"
	PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_DRAGON_REEL_POSITIONS       = "ReelPositions"
	PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_DRAGON_PATTERNS             = "Patterns"
	PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_DRAGON_POSITIONS            = "Positions"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_BATTLE_OF_MYTHS_DRAGON, func() feature.Feature { return new(TriggerBattleOfMythsDragon) })

type TriggerBattleOfMythsDragon struct {
	feature.Base
}

func (f *TriggerBattleOfMythsDragon) DataPtr() interface{} {
	return nil
}

func (f TriggerBattleOfMythsDragon) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	if state.Action == "cascade" {
		return
	}

	number := params.GetIntSlice(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_DRAGON_NUMBER)
	numberProbs := params.GetIntSlice(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_DRAGON_NUMBER_PROBABILITIES)
	numIdx := feature.WeightedRandomIndex(numberProbs)
	num := number[numIdx]

	reelProbs := feature.ConvertIntSlice(params.GetSlice(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_DRAGON_REEL_PROBABILITIES)[numIdx])
	reelsIdx := feature.WeightedRandomIndex(reelProbs)

	reels := feature.ConvertIntSlice(params.GetSlice(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_DRAGON_REEL_POSITIONS)[numIdx].([]interface{})[reelsIdx])

	patterns := params.GetSlice(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_DRAGON_PATTERNS)
	patternIdx := rng.RandFromRange(len(patterns))
	pattern := feature.ConvertIntSlice(patterns[patternIdx])

	gridh := len(state.SymbolGrid[0])

	positions := []int{}
	for i := 0; i < num; i++ {
		x := reels[i]
		y := pattern[x]
		positions = append(positions, x*gridh+y)
	}

	if len(positions) > 0 {
		params[PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_DRAGON_POSITIONS] = positions
		feature.ActivateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f *TriggerBattleOfMythsDragon) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerBattleOfMythsDragon) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
