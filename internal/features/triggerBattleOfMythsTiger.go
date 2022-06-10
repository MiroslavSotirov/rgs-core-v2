package features

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type TriggerBattleOfMythsTiger struct {
	FeatureDef
}

func (f *TriggerBattleOfMythsTiger) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerBattleOfMythsTiger) DataPtr() interface{} {
	return nil
}

func (f *TriggerBattleOfMythsTiger) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *TriggerBattleOfMythsTiger) OnInit(state *FeatureState) {
}

func (f TriggerBattleOfMythsTiger) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}
	if state.Action == "cascade" {
		return
	}

	sizes := params.GetIntSlice("Sizes")
	sizeProbs := params.GetIntSlice("SizeProbabilities")
	reelProbs := params.GetSlice("ReelProbabilities")

	sizeidx := WeightedRandomIndex(sizeProbs)
	reelidx := WeightedRandomIndex(convertIntSlice(reelProbs[sizeidx]))
	rowidx := 0
	if sizes[sizeidx] < len(state.SymbolGrid[0]) {
		rowidx = rng.RandFromRange(len(state.SymbolGrid[0]) - sizes[sizeidx])
	}

	params["W"] = sizes[sizeidx]
	params["H"] = sizes[sizeidx]
	params["X"] = reelidx
	params["Y"] = rowidx

	logger.Debugf("Activate tiger feature. w: %d h: %d x: %d y: %d",
		params.GetInt("W"), params.GetInt("H"), params.GetInt("X"), params.GetInt("Y"))

	activateFeatures(f.FeatureDef, state, params)
	return
}

func (f TriggerBattleOfMythsTiger) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *TriggerBattleOfMythsTiger) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerBattleOfMythsTiger) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
