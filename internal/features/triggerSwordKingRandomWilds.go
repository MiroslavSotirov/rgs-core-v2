package features

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

type TriggerSwordKingRandomWilds struct {
	FeatureDef
}

func (f *TriggerSwordKingRandomWilds) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerSwordKingRandomWilds) DataPtr() interface{} {
	return nil
}

func (f *TriggerSwordKingRandomWilds) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *TriggerSwordKingRandomWilds) OnInit(state *FeatureState) {
}

func (f TriggerSwordKingRandomWilds) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}

	Probability := params.GetInt("Probability")
	if rng.RandFromRange(10000) < Probability {

		numWilds := params.GetIntSlice("NumWilds")[WeightedRandomIndex(params.GetIntSlice("NumProbabilities"))]
		positions := []int{}
		gridh := len(state.SymbolGrid[0])

		for tries := numWilds * params.GetInt("RetryFactor"); numWilds > 0 && tries > 0; tries-- {
			reel := WeightedRandomIndex(params.GetIntSlice("ReelProbabilities"))
			row := WeightedRandomIndex(params.GetIntSlice("RowProbabilities"))
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
			params["RunWilds"] = true
			params["Positions"] = positions
			activateFeatures(f.FeatureDef, state, params)
		}
	}

	return
}

func (f TriggerSwordKingRandomWilds) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *TriggerSwordKingRandomWilds) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerSwordKingRandomWilds) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
