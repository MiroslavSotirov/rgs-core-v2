package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

const (
	FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH_ISHTAR = "TriggerLawOfGilgameshIshtar"

	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_ISHTAR_REMOVE_IDS = "RemoveIds"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH_ISHTAR, func() feature.Feature { return new(TriggerLawOfGilgameshIshtar) })

type TriggerLawOfGilgameshIshtar struct {
	feature.Base
}

func (f TriggerLawOfGilgameshIshtar) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	if len(state.CalculateWins(state.SymbolGrid, nil)) != 0 {
		logger.Debugf("postponing feature due to wins")
	}

	removeIds := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_ISHTAR_REMOVE_IDS)

	gridh := len(state.SymbolGrid[0])
	positions := []int{}

	for ireel, r := range state.SymbolGrid {
		for irow, s := range r {
			pos := ireel*gridh + irow
			if func() bool {
				for _, r := range removeIds {
					if r == s {
						return true
					}
				}
				return false
			}() {
				positions = append(positions, pos)
			}
		}
	}

	if len(positions) > 0 {
		incLawOfGilgameshLevel(state, params)
		params[featureProducts.PARAM_ID_RESPIN_POSITIONS] = positions
		feature.ActivateFeatures(f.FeatureDef, state, params)
		/*
			state.Wins = append(state.Wins, feature.FeatureWin{
				Index:           "cascade1:1",
				SymbolPositions: positions,
			})
		*/
	}

	return
}

func (f *TriggerLawOfGilgameshIshtar) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerLawOfGilgameshIshtar) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
