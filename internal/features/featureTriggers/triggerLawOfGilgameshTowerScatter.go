package featureTriggers

import (
	"strings"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

const (
	FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER = "TriggerLawOfGilgameshTowerScatter"

	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_TILE_ID            = "TileId"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_KEEP_IDS           = "KeepIds"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_UNTRIGGER_IDS      = "UntriggerIds"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_RETRY_FACTOR       = "RetryFactor"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_NUM_SCATTERS       = "NumScatters"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_NUM_PROBABILITIES  = "NumProbabilities"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_REEL_PROBABILITIES = "ReelProbabilities"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_ROW_PROBABILITIES  = "RowProbabilities"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_BONUS_THRESHOLD    = "BonusThreshold"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_TRIGGER_TOWER      = "TriggerTower"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER, func() feature.Feature { return new(TriggerLawOfGilgameshTowerScatter) })

type TriggerLawOfGilgameshTowerScatter struct {
	feature.Base
}

func (f TriggerLawOfGilgameshTowerScatter) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	tileId := params.GetInt(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_TILE_ID)
	retryFactor := params.GetInt(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_RETRY_FACTOR)
	bonusThreshold := params.GetInt(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_BONUS_THRESHOLD)
	var untriggerIds []int
	if params.HasKey(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_UNTRIGGER_IDS) {
		untriggerIds = params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_UNTRIGGER_IDS)
	}
	isRespin := strings.Contains(state.Action, "cascade")

	gridh := len(state.SourceGrid[0])
	positions := []int{}
	for reel, r := range state.SymbolGrid {
		for row, s := range r {
			if s == tileId {
				positions = append(positions, reel*gridh+row)
			}
			if isRespin {
				for _, u := range untriggerIds {
					if s == u {
						logger.Debugf("untrigger tower scatter due to presence of symbols %v", untriggerIds)
						return
					}
				}
			}
		}
	}

	if len(positions) >= bonusThreshold {
		logger.Debugf("skipping tower scatter due to already %d placed", bonusThreshold)
		return
	}

	newPositions := []int{}
	numScatters := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_NUM_SCATTERS)
	numProbs := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_NUM_PROBABILITIES)
	ns := numScatters[feature.WeightedRandomIndex(numProbs)]

	if isRespin {

		candidates := state.GetCandidatePositions()
		if len(positions) < bonusThreshold {
			for i := 0; i < ns && len(candidates) > 0; i++ {
				ic := rng.RandFromRange(len(candidates))
				p := candidates[ic]
				candidates = append(candidates[:ic], candidates[ic+1:]...)

				reel := p / gridh
				row := p % gridh
				//				state.SourceGrid[reel][row] = tileId
				state.SymbolGrid[reel][row] = tileId
				positions = append(positions, p)
				newPositions = append(newPositions, p)
			}
		}

	} else {

		keepIds := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_KEEP_IDS)
		reelProbs := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_REEL_PROBABILITIES)
		rowProbs := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_ROW_PROBABILITIES)

		tries := ns * retryFactor
		for i := 0; i < tries && ns > 0; i++ {
			reel := feature.WeightedRandomIndex(reelProbs)
			row := feature.WeightedRandomIndex(rowProbs)
			if func(sym int) bool {
				for _, s := range keepIds {
					if s == sym {
						return false
					}
				}
				return true
			}(state.SymbolGrid[reel][row]) {
				//				state.SourceGrid[reel][row] = tileId
				pos := reel*gridh + row
				state.SymbolGrid[reel][row] = tileId
				positions = append(positions, pos)
				newPositions = append(newPositions, pos)
				ns--
			}
		}
	}

	if len(newPositions) > 0 {

		params[featureProducts.PARAM_ID_REPLACE_TILE_POSITIONS] = newPositions

		if len(positions) >= bonusThreshold {

			params[PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_TRIGGER_TOWER] = true

		}

		feature.ActivateFeatures(f.FeatureDef, state, params)

	}

	return
}

func (f *TriggerLawOfGilgameshTowerScatter) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerLawOfGilgameshTowerScatter) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
