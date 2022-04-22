package features

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

type TriggerSupaCrew struct {
	FeatureDef
}

func (f *TriggerSupaCrew) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerSupaCrew) DataPtr() interface{} {
	return nil
}

func (f *TriggerSupaCrew) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *TriggerSupaCrew) OnInit(state *FeatureState) {
}

func (f TriggerSupaCrew) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}
	params["Random"] = rng.RandFromRange(f.FeatureDef.Params["RandomRange"].(int))
	activateFeatures(f.FeatureDef, state, params)
	return
}

func (f TriggerSupaCrew) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	forceGrids := map[string][][]int{
		"addbalance": {
			{0, 1, 2},
			{3, 4, 5},
			{6, 7, 8},
			{0, 1, 2},
			{3, 4, 5},
		},
		"bigwin": {
			{0, 0, 0},
			{0, 0, 0},
			{0, 0, 0},
			{0, 0, 0},
			{0, 0, 0},
		},
		"superwin": {
			{3, 3, 3},
			{3, 3, 3},
			{3, 3, 3},
			{3, 3, 3},
			{3, 3, 3},
		},
		"megawin": {
			{8, 8, 8},
			{8, 8, 8},
			{8, 8, 8},
			{8, 8, 8},
			{8, 8, 8},
		},
		"linesymbol0": {
			{0, 1, 2},
			{0, 4, 5},
			{0, 7, 8},
			{0, 1, 2},
			{0, 4, 5},
		},
		"linesymbol1": {
			{0, 1, 2},
			{3, 1, 5},
			{6, 1, 8},
			{0, 1, 2},
			{3, 1, 5},
		},
		"linesymbol2": {
			{0, 1, 2},
			{3, 4, 2},
			{6, 7, 2},
			{0, 1, 2},
			{3, 4, 2},
		},
		"linesymbol3": {
			{3, 1, 2},
			{3, 4, 5},
			{3, 7, 8},
			{3, 1, 2},
			{3, 4, 5},
		},
		"linesymbol4": {
			{0, 4, 2},
			{3, 4, 5},
			{6, 4, 8},
			{0, 4, 2},
			{3, 4, 5},
		},
		"linesymbol5": {
			{0, 1, 5},
			{3, 4, 5},
			{6, 7, 5},
			{0, 1, 5},
			{3, 4, 5},
		},
		"linesymbol6": {
			{6, 1, 2},
			{6, 4, 5},
			{6, 7, 8},
			{6, 1, 2},
			{6, 4, 5},
		},
		"linesymbol7": {
			{0, 7, 2},
			{3, 7, 5},
			{6, 7, 8},
			{0, 7, 2},
			{3, 7, 5},
		},
		"linesymbol8": {
			{0, 1, 8},
			{3, 4, 8},
			{6, 7, 8},
			{0, 1, 8},
			{3, 4, 8},
		},
	}

	forceConfigs := []string{
		"addbalance", "bigwin", "superwin", "megawin", "linesymbol0", "linesymbol1", "linesymbol2",
		"linesymbol3", "linesymbol4", "linesymbol5", "linesymbol6", "linesymbol7", "linesymbol8"}
	for _, f := range forceConfigs {
		if params.GetForce(f) != "" {
			symbols, ok := forceGrids[f]
			if ok {
				state.SymbolGrid = symbols
				for i := range state.SourceGrid {
					for j := range state.SourceGrid[i] {
						state.SourceGrid[i][j] = symbols[i][j]
					}
				}
				return true
			}
		}
	}
	return false
}

func (f *TriggerSupaCrew) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerSupaCrew) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
