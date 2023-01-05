package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
)

const (
	FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER = "TriggerLawOfGilgameshTowerScatter"

	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_TILE_ID            = "TileId"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_KEEP_IDS           = "KeepIds"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_RETRY_FACTOR       = "RetryFactor"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_NUM_SCATTERS       = "NumScatters"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_NUM_PROBABILITIES  = "NumProbabilities"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_REEL_PROBABILITIES = "ReelProbabilities"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_ROW_PROBABILITIES  = "RowProbabilities"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_WINS_LEVELS        = "WinsLevels"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_PROB_LEVELS        = "ProbabilitiesLevels"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_BONUS_THRESHOLD    = "BonusThreshold"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER, func() feature.Feature { return new(TriggerLawOfGilgameshTowerScatter) })

type TriggerLawOfGilgameshTowerScatter struct {
	feature.Base
}

func (f TriggerLawOfGilgameshTowerScatter) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	tileId := params.GetInt(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_TILE_ID)
	retryFactor := params.GetInt(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_RETRY_FACTOR)
	bonusThreshold := params.GetInt(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_BONUS_THRESHOLD)

	gridh := len(state.SourceGrid[0])
	positions := []int{}
	for reel, r := range state.SymbolGrid {
		for row, s := range r {
			if s == tileId {
				positions = append(positions, reel*gridh+row)
			}
		}
	}

	if len(positions) < bonusThreshold {
		keepIds := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_KEEP_IDS)
		numScatters := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_NUM_SCATTERS)
		numProbs := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_NUM_PROBABILITIES)
		reelProbs := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_REEL_PROBABILITIES)
		rowProbs := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_ROW_PROBABILITIES)

		ns := numScatters[feature.WeightedRandomIndex(numProbs)]
		tries := ns * retryFactor
		for i := 0; i < tries && ns > 0; i++ {
			reel := feature.WeightedRandomIndex(reelProbs)
			row := feature.WeightedRandomIndex(rowProbs)
			if func(sym int) bool {
				for s := range keepIds {
					if s == sym {
						return false
					}
				}
				return true
			}(state.SourceGrid[reel][row]) {
				state.SourceGrid[reel][row] = tileId
				positions = append(positions, reel*gridh+row)
			}
		}
	}

	if len(positions) >= bonusThreshold {
		winsLevels := params.GetSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_WINS_LEVELS)
		probLevels := params.GetSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_PROB_LEVELS)

		level := 0
		amount := 0
		payouts := []int{}
		for level < len(winsLevels) {
			win := feature.WeightedRandomIndex(feature.ConvertIntSlice(probLevels[level]))
			amount := feature.ConvertIntSlice(winsLevels[level])[win]
			payouts = append(payouts, amount)
			if amount < 0 {
				level++
			} else {
				break
			}
		}

		if amount > 0 {
			params[featureProducts.PARAM_ID_INSTA_WIN_AMOUNT] = amount
			params[featureProducts.PARAM_ID_INSTA_WIN_PAYOUTS] = payouts
			params[featureProducts.PARAM_ID_INSTA_WIN_TYPE] = "tower"
			params[featureProducts.PARAM_ID_INSTA_WIN_SOURCE_ID] = 4
			params[featureProducts.PARAM_ID_INSTA_WIN_TILE_ID] = tileId
			params[featureProducts.PARAM_ID_INSTA_WIN_POSITIONS] = positions
			feature.ActivateFeatures(f.FeatureDef, state, params)
		}
	}
	return
}

func (f *TriggerLawOfGilgameshTowerScatter) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerLawOfGilgameshTowerScatter) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
