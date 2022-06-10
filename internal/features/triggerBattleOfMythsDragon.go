package features

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type TriggerBattleOfMythsDragon struct {
	FeatureDef
}

func (f *TriggerBattleOfMythsDragon) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerBattleOfMythsDragon) DataPtr() interface{} {
	return nil
}

func (f *TriggerBattleOfMythsDragon) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *TriggerBattleOfMythsDragon) OnInit(state *FeatureState) {
}

func (f TriggerBattleOfMythsDragon) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}
	if state.Action == "cascade" {
		return
	}

	number := params.GetIntSlice("Number")
	numberProbs := params.GetIntSlice("NumberProbabilities")
	numIdx := WeightedRandomIndex(numberProbs)
	num := number[numIdx]

	reelProbs := convertIntSlice(params.GetSlice("ReelProbabilities")[numIdx])
	reelsIdx := WeightedRandomIndex(reelProbs)

	logger.Debugf("num: %d reelsIdx: %d ReelPositions: %v", num, reelsIdx, params.GetSlice("ReelPositions"))
	reels := convertIntSlice(params.GetSlice("ReelPositions")[numIdx].([]interface{})[reelsIdx])

	patterns := params.GetSlice("Patterns")
	patternIdx := rng.RandFromRange(len(patterns))
	pattern := convertIntSlice(patterns[patternIdx])

	gridh := len(state.SymbolGrid[0])

	positions := []int{}
	for i := 0; i < num; i++ {
		x := reels[i]
		y := pattern[x]
		positions = append(positions, x*gridh+y)
	}

	if len(positions) > 0 {
		params["Positions"] = positions

		logger.Debugf("activating dragon num: %d reels: %v pattern: %v position: %v0", num, reels, pattern, positions)
		activateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f TriggerBattleOfMythsDragon) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *TriggerBattleOfMythsDragon) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerBattleOfMythsDragon) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
