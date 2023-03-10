package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
)

const (
	FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH_PROGRESS = "TriggerLawOfGilgameshProgress"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH_PROGRESS, func() feature.Feature { return new(TriggerLawOfGilgameshProgress) })

type TriggerLawOfGilgameshProgress struct {
	feature.Base
}

func (f TriggerLawOfGilgameshProgress) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	stateless := feature.GetParamStatelessMap(params)

	wins := state.CalculateWins(state.SymbolGrid, nil)
	if len(wins) > 0 {

		stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_COUNTER] = stateless.GetInt(STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_COUNTER) +
			countLawOfGilgameshWin(state)
		feature.SetStatelessMap(stateless, params)
	} else {
		feature.ActivateFeatures(f.FeatureDef, state, params)
	}

	return
}

func (f *TriggerLawOfGilgameshProgress) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerLawOfGilgameshProgress) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}

func countLawOfGilgameshWin(state *feature.FeatureState) int {
	counter := 0
	wins := state.CalculateWins(state.SymbolGrid, nil)
	for _, w := range wins {
		num := len(w.SymbolPositions)
		counter += num
	}
	return counter
}

func incLawOfGilgameshLevel(state *feature.FeatureState, params feature.FeatureParams) {
	stateless := feature.GetParamStatelessMap(params)
	stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_LEVEL] = stateless.GetInt(STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_LEVEL) + 1
	stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_COUNTER] = stateless.GetInt(STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_COUNTER) +
		countLawOfGilgameshWin(state)
	feature.SetStatelessMap(stateless, params)
}
