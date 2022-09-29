package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
)

const (
	FEATURE_ID_TRIGGER_SYMBOL_COUNT = "TriggerSymbolCount"

	PARAM_ID_TRIGGER_SYMBOL_COUNT_TILE_ID = "TileId"
	PARAM_ID_TRIGGER_SYMBOL_COUNT_NUMBERS = "Numbers"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_SYMBOL_COUNT, func() feature.Feature { return new(TriggerSymbolCount) })

type TriggerSymbolCount struct {
	feature.Base
}

func (f TriggerSymbolCount) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	tileId := params.GetInt(PARAM_ID_TRIGGER_SYMBOL_COUNT_TILE_ID)
	numbers := params.GetIntSlice(PARAM_ID_TRIGGER_SYMBOL_COUNT_NUMBERS)
	count := 0
	for _, r := range state.SymbolGrid {
		for _, c := range r {
			if c == tileId {
				count++
			}
		}
	}
	for _, n := range numbers {
		if n == count {
			feature.ActivateFeatures(f.FeatureDef, state, params)
			return
		}
	}

	return
}

func (f *TriggerSymbolCount) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerSymbolCount) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
