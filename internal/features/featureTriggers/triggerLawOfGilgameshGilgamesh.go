package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
)

const (
	FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH_GILGAMESH = "TriggerLawOfGilgameshGilgamesh"

	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_GILGAMESH_WILD_ID            = "WildId"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_GILGAMESH_NUM_WILDS          = "NumWilds"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_GILGAMESH_NUM_PROBABILITIES  = "NumProbabilities"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_GILGAMESH_RETRY_FACTOR       = "RetryFactor"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_GILGAMESH_REEL_PROBABILITIES = "ReelProbabilities"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_GILGAMESH_ROW_PROBABILITIES  = "RowProbabilities"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_GILGAMESH_KEEP_IDS           = "KeepIds"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH_GILGAMESH, func() feature.Feature { return new(TriggerLawOfGilgameshGilgamesh) })

type TriggerLawOfGilgameshGilgamesh struct {
	feature.Base
	Data feature.FeatureParams `j́son:"data"`
}

func (f *TriggerLawOfGilgameshGilgamesh) DataPtr() interface{} {
	return &f.Data
}
func (f TriggerLawOfGilgameshGilgamesh) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	wildId := params.GetInt(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_GILGAMESH_WILD_ID)
	numWilds := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_GILGAMESH_NUM_WILDS)
	numProbs := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_GILGAMESH_NUM_PROBABILITIES)
	retryFactor := params.GetInt(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_GILGAMESH_RETRY_FACTOR)
	reelProbs := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_GILGAMESH_REEL_PROBABILITIES)
	rowProbs := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_GILGAMESH_ROW_PROBABILITIES)
	keepIds := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_GILGAMESH_KEEP_IDS)

	positions := []int{}
	gridh := len(state.SymbolGrid[0])
	nw := numWilds[feature.WeightedRandomIndex(numProbs)]
	for i := 0; i < nw*retryFactor && len(positions) < nw; i++ {
		reel := feature.WeightedRandomIndex(reelProbs)
		row := feature.WeightedRandomIndex(rowProbs)
		pos := reel*gridh + row
		if func(sym int) bool {
			for _, s := range keepIds {
				if s == sym {
					return false
				}
			}
			for _, p := range positions {
				if p == pos {
					return false
				}
			}
			return true
		}(state.SymbolGrid[reel][row]) {
			positions = append(positions, pos)
			state.SourceGrid[reel][row] = wildId
		}
	}

	if len(positions) > 0 {
		params[featureProducts.PARAM_ID_REPLACE_TILE_TILE_ID] = wildId
		params[featureProducts.PARAM_ID_REPLACE_TILE_REPLACE_WITH_ID] = wildId
		params[featureProducts.PARAM_ID_REPLACE_TILE_POSITIONS] = positions
		feature.ActivateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f *TriggerLawOfGilgameshGilgamesh) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerLawOfGilgameshGilgamesh) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
