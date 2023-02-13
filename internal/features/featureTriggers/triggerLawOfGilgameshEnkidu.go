package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
)

const (
	FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH_ENKIDU = "TriggerLawOfGilgameshEnkidu"

	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_ENKIDU_POSITIONS = "RemovePositions"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_ENKIDU_KEEP_IDS  = "KeepIds"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH_ENKIDU, func() feature.Feature { return new(TriggerLawOfGilgameshEnkidu) })

type TriggerLawOfGilgameshEnkidu struct {
	feature.Base
}

func (f TriggerLawOfGilgameshEnkidu) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	gridh := len(state.SymbolGrid[0])
	remove := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_ENKIDU_POSITIONS)
	keep := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_ENKIDU_KEEP_IDS)
	positions := []int{}
	for _, pos := range remove {
		reel := pos / gridh
		slot := pos % gridh
		if func(sym int) bool {
			for _, k := range keep {
				if k == sym {
					return false
				}
			}
			return true
		}(state.SymbolGrid[reel][slot]) {
			positions = append(positions, pos)
		}
	}

	params[featureProducts.PARAM_ID_RESPIN_POSITIONS] = positions
	feature.ActivateFeatures(f.FeatureDef, state, params)

	incLawOfGilgameshLevel(state, params)

	return
}

func (f *TriggerLawOfGilgameshEnkidu) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerLawOfGilgameshEnkidu) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
