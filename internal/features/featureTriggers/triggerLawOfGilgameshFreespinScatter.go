package featureTriggers

import (
	"fmt"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
)

const (
	FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER = "TriggerLawOfGilgameshFreespinScatter"

	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER_TILE_ID            = "TileId"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER_KEEP_IDS           = "KeepIds"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER_RETRY_FACTOR       = "RetryFactor"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER_NUM_SCATTERS       = "NumScatters"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER_NUM_PROBABILITIES  = "NumProbabilities"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER_REEL_PROBABILITIES = "ReelProbabilities"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER_ROW_PROBABILITIES  = "RowProbabilities"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER_WINS_LEVELS        = "WinsLevels"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER_PROB_LEVELS        = "ProbabilitiesLevels"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER_BONUS_THRESHOLD    = "BonusThreshold"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER_FREESPINS          = "Freespins"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER_ADDITIONAL         = "Additional"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER, func() feature.Feature { return new(TriggerLawOfGilgameshFreespinScatter) })

type TriggerLawOfGilgameshFreespinScatter struct {
	feature.Base
}

func (f TriggerLawOfGilgameshFreespinScatter) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	tileId := params.GetInt(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER_TILE_ID)
	retryFactor := params.GetInt(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER_RETRY_FACTOR)
	bonusThreshold := params.GetInt(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER_BONUS_THRESHOLD)

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
		keepIds := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER_KEEP_IDS)
		numScatters := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER_NUM_SCATTERS)
		numProbs := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER_NUM_PROBABILITIES)
		reelProbs := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER_REEL_PROBABILITIES)
		rowProbs := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER_ROW_PROBABILITIES)

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

		freespins := params.GetInt(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER_FREESPINS)
		additional := params.GetInt(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_FREESPIN_SCATTER_ADDITIONAL)

		numFreespins := freespins
		if len(positions) > bonusThreshold {
			numFreespins += len(positions) * additional
		}

		if numFreespins > 0 {
			state.Wins = append(state.Wins, feature.FeatureWin{
				Index:           fmt.Sprintf("%s:%d", "freespin", numFreespins),
				SymbolPositions: positions,
			})
		}
	}
	return
}

func (f *TriggerLawOfGilgameshFreespinScatter) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerLawOfGilgameshFreespinScatter) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
