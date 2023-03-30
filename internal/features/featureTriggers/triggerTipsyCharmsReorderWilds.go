package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

const (
	FEATURE_ID_TRIGGER_TIPSY_CHARMS_REORDER_WILDS = "TriggerTipsyCharmsReorderWilds"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_TIPSY_CHARMS_REORDER_WILDS,
	func() feature.Feature { return new(TriggerTipsyCharmsReorderWilds) })

type TriggerTipsyCharmsReorderWilds struct {
	feature.Base
}

func (f TriggerTipsyCharmsReorderWilds) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	gridh := len(state.SymbolGrid[0])
	scatterIds := params.GetIntSlice(PARAM_ID_TRIGGER_TIPSY_CHARMS_WILD_IDS)

	positions := []int{}
	for ireel, r := range state.SymbolGrid {
		for irow, s := range r {
			if func() bool {
				for _, scatter := range scatterIds {
					if scatter == s {
						return true
					}
				}
				return false
			}() {
				positions = append(positions, ireel*gridh+irow)
			}
		}
	}

	candidates := []int{}
	for ireel, r := range state.SymbolGrid {
		for irow := range r {
			candidates = append(candidates, ireel*gridh+irow)
		}
	}

	destinations := make([]int, len(positions))
	for i := range positions {
		nc := len(candidates)
		if nc == 0 {
			panic("no candidate destination")
		}
		ic := rng.RandFromRangePool(nc)
		destinations[i] = candidates[ic]
		candidates = append(candidates[:ic], candidates[ic+1:]...)
	}

	if len(positions) > 0 {

		params[featureProducts.PARAM_ID_MOVE_TILES_POSITIONS] = positions
		params[featureProducts.PARAM_ID_MOVE_TILES_DESTINATIONS] = destinations

		feature.ActivateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f *TriggerTipsyCharmsReorderWilds) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerTipsyCharmsReorderWilds) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
