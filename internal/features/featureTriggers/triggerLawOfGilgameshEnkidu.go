package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
)

const (
	FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH_ENKIDU = "TriggerLawOfGilgameshEnkidu"

	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_ENKIDU_POSITIONS = "RemovePositions"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH_ENKIDU, func() feature.Feature { return new(TriggerLawOfGilgameshEnkidu) })

type TriggerLawOfGilgameshEnkidu struct {
	feature.Base
}

func (f TriggerLawOfGilgameshEnkidu) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	positions := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_ENKIDU_POSITIONS)
	if len(positions) > 0 {
		state.Wins = append(state.Wins, feature.FeatureWin{
			Index:           "cascade:1", // fmt.Sprintf("enkido", len(positions)),
			SymbolPositions: positions,
		})
	}

	feature.ActivateFeatures(f.FeatureDef, state, params)
	return
}

func (f *TriggerLawOfGilgameshEnkidu) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerLawOfGilgameshEnkidu) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
