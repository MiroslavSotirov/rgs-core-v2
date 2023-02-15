package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

const (
	FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_BONUS = "TriggerLawOfGilgameshFreespinBonus"

	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_BONUS_TILE_ID    = "TileId"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_BONUS_THRESHOLD  = "BonusThreshold"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_BONUS_FREESPINS  = "Freespins"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_BONUS_ADDITIONAL = "Additional"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_BONUS, func() feature.Feature { return new(TriggerLawOfGilgameshFreespinBonus) })

type TriggerLawOfGilgameshFreespinBonus struct {
	feature.Base
}

func (f TriggerLawOfGilgameshFreespinBonus) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	tileId := 10
	bonusThreshold := params.GetInt(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_BONUS_THRESHOLD)

	gridh := len(state.SourceGrid[0])
	positions := []int{}
	for reel, r := range state.SymbolGrid {
		for row, s := range r {
			if s == tileId {
				positions = append(positions, reel*gridh+row)
			}
		}
	}

	numFreespins := 0
	numScatters := len(positions)
	if numScatters >= bonusThreshold {
		freespins := params.GetInt(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_BONUS_FREESPINS)
		additional := params.GetInt(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_BONUS_ADDITIONAL)

		numFreespins = freespins
		if len(positions) > bonusThreshold {
			numFreespins += (numScatters - bonusThreshold) * additional
		}
	}

	if numFreespins > 0 {
		logger.Debugf("award %d freespins", numFreespins)
		params[featureProducts.PARAM_ID_RESPIN_AMOUNT] = numFreespins
		feature.ActivateFeatures(f.FeatureDef, state, params)
	}

	return
}

func (f *TriggerLawOfGilgameshFreespinBonus) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerLawOfGilgameshFreespinBonus) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
