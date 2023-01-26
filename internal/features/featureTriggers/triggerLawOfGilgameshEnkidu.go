package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
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

	if len(state.CalculateWins(state.SymbolGrid, nil)) != 0 {
		logger.Debugf("postponing feature due to wins")
	}

	positions := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_ENKIDU_POSITIONS)
	if len(positions) > 0 {
		incLawOfGilgameshLevel(state, params)
		feature.ActivateFeatures(f.FeatureDef, state, params)
		/*
			state.Wins = append(state.Wins, feature.FeatureWin{
				Index:           "cascade:1", // fmt.Sprintf("enkido", len(positions)),
				SymbolPositions: positions,
			})
		*/
	}

	return
}

func (f *TriggerLawOfGilgameshEnkidu) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerLawOfGilgameshEnkidu) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
