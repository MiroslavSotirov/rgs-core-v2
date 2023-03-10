package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
)

const (
	FEATURE_ID_TRIGGER_TIPSY_CHARMS_LEPRECHAUN = "TriggerTipsyCharmsLeprechaun"

	PARAM_ID_TRIGGER_TIPSY_CHARMS_LEPRECHAUN_INCREMENTS               = "Increments"
	PARAM_ID_TRIGGER_TIPSY_CHARMS_LEPRECHAUN_INCREMENTS_PROBABILITIES = "IncrementsProbabilities"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_TIPSY_CHARMS_LEPRECHAUN,
	func() feature.Feature { return new(TriggerTipsyCharmsLeprechaun) })

type TriggerTipsyCharmsLeprechaun struct {
	feature.Base
}

func (f TriggerTipsyCharmsLeprechaun) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	gridh := len(state.SymbolGrid[0])
	scatterIds := params.GetIntSlice(PARAM_ID_TRIGGER_TIPSY_CHARMS_WILD_IDS)

	if func() bool {
		wins := state.CalculateWins(state.SymbolGrid, nil)
		for _, win := range wins {
			for _, pos := range win.SymbolPositions {
				r := pos / gridh
				s := pos % gridh
				for _, scatter := range scatterIds {
					if state.SymbolGrid[r][s] == scatter {
						return true
					}
				}
			}
		}
		return false
	}() {

		increments := params.GetIntSlice(PARAM_ID_TRIGGER_TIPSY_CHARMS_LEPRECHAUN_INCREMENTS)
		incrementsProbs := params.GetIntSlice(PARAM_ID_TRIGGER_TIPSY_CHARMS_LEPRECHAUN_INCREMENTS_PROBABILITIES)

		positions := []int{}
		replaceIds := []int{}

		for ireel, r := range state.SymbolGrid {
			for isymbol, s := range r {
				for iscatter, scatter := range scatterIds {
					if s == scatter {
						increment := increments[feature.WeightedRandomIndex(incrementsProbs)]
						iscatter += increment
						if iscatter >= len(scatterIds) {
							iscatter = len(scatterIds) - 1
						}
						positions = append(positions, ireel*gridh+isymbol)
						replaceIds = append(replaceIds, iscatter)
						break
					}
				}
			}
		}

		if len(positions) > 0 {

			params[featureProducts.PARAM_ID_REPLACE_TILE_POSITIONS] = positions
			params[featureProducts.PARAM_ID_REPLACE_TILE_REPLACE_WITH_IDS] = replaceIds

			feature.ActivateFeatures(f.FeatureDef, state, params)
		}
	}
	return
}

func (f *TriggerTipsyCharmsLeprechaun) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerTipsyCharmsLeprechaun) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
