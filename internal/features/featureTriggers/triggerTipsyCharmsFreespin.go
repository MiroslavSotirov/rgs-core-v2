package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

const (
	FEATURE_ID_TRIGGER_TIPSY_CHARMS_FREESPIN = "TriggerTipsyCharmsFreespin"

	PARAM_ID_TRIGGER_TIPSY_CHARMS_FREESPIN_SCATTER_ID                        = "ScatterId"
	PARAM_ID_TRIGGER_TIPSY_CHARMS_FREESPIN_NUM_SCATTERS_LEVELS               = "NumScattersLevels"
	PARAM_ID_TRIGGER_TIPSY_CHARMS_FREESPIN_NUM_SCATTERS_PROBABILITIES_LEVELS = "NumScattersProbabilitiesLevels"
	PARAM_ID_TRIGGER_TIPSY_CHARMS_FREESPIN_AMOUNTS                           = "Amounts"
	PARAM_ID_TRIGGER_TIPSY_CHARMS_FREESPIN_AMOUNTS_PROBABILITIES             = "AmountsProbabilities"

	PARAM_VALUE_TRIGGER_TIPSY_CHARMS_FREESPIN_WILD_RESPIN = "WildRespin"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_TIPSY_CHARMS_FREESPIN,
	func() feature.Feature { return new(TriggerTipsyCharmsFreespin) })

type TriggerTipsyCharmsFreespin struct {
	feature.Base
}

func (f TriggerTipsyCharmsFreespin) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	level := 0

	numScattersLevels := feature.ConvertIntSlice(
		params.GetSlice(PARAM_ID_TRIGGER_TIPSY_CHARMS_FREESPIN_NUM_SCATTERS_LEVELS)[level])
	numScattersProbabilitiesLevels := feature.ConvertIntSlice(
		params.GetSlice(PARAM_ID_TRIGGER_TIPSY_CHARMS_FREESPIN_NUM_SCATTERS_PROBABILITIES_LEVELS)[level])
	numScatters := numScattersLevels[feature.WeightedRandomIndex(numScattersProbabilitiesLevels)]

	scatterId := params.GetInt(PARAM_ID_TRIGGER_TIPSY_CHARMS_FREESPIN_SCATTER_ID)
	gridw, gridh := len(state.SymbolGrid), len(state.SymbolGrid[0])
	tries := numScatters * 3
	positions := []int{}

	ns := countSymbols(scatterId, state.SymbolGrid)

	for i := 0; i < tries && len(positions) < numScatters; i++ {
		reel := rng.RandFromRange(gridw)
		row := rng.RandFromRange(gridh)
		pos := reel*gridh + row
		if func() bool {
			if state.SymbolGrid[reel][row] == scatterId {
				return false
			}
			for p := range positions {
				if positions[p] == pos {
					return false
				}
			}
			return true
		}() {
			positions = append(positions, pos)
		}
	}

	if len(positions) > 0 {
		if len(positions)+ns >= 3 {
			// activate freespins
			numFreespins := params.GetIntSlice(PARAM_ID_TRIGGER_TIPSY_CHARMS_FREESPIN_AMOUNTS)[feature.WeightedRandomIndex(
				params.GetIntSlice(PARAM_ID_TRIGGER_TIPSY_CHARMS_FREESPIN_AMOUNTS_PROBABILITIES))]

			logger.Debugf("activate %d freespins", numFreespins)
			params[featureProducts.PARAM_ID_RESPIN_AMOUNT] = numFreespins
			params[PARAM_VALUE_TRIGGER_TIPSY_CHARMS_FREESPIN_WILD_RESPIN] = true
		}

		params[featureProducts.PARAM_ID_REPLACE_TILE_REPLACE_WITH_ID] = scatterId
		params[featureProducts.PARAM_ID_REPLACE_TILE_POSITIONS] = positions

		feature.ActivateFeatures(f.FeatureDef, state, params)
	}

	return
}

func (f *TriggerTipsyCharmsFreespin) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerTipsyCharmsFreespin) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
