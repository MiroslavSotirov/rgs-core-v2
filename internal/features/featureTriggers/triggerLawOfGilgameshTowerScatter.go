package featureTriggers

import (
	"fmt"
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
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_WINS_LEVELS        = "WinsLevels"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_PROB_LEVELS        = "ProbabilitiesLevels"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_BONUS_THRESHOLD    = "BonusThreshold"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_TRIGGER_TOWER      = "TriggerTower"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_TOWER_SCATTERS     = "TowerScatters"
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
		params[PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_TOWER_SCATTERS] = true
		params[featureProducts.PARAM_ID_REPLACE_TILE_POSITIONS] = newPositions
	}

	if len(positions) >= bonusThreshold {

		winsLevels := params.GetSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_WINS_LEVELS)
		probLevels := params.GetSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_PROB_LEVELS)

		amount, payouts := f.towerBonus(winsLevels, probLevels)

		if amount > 0 {

			if state.Multiplier > 1 {
				panic(fmt.Sprintf("tower bonus with multiplier %d and amount %d", state.Multiplier, amount))
			}

			params[featureProducts.PARAM_ID_INSTA_WIN_AMOUNT] = amount
			params[featureProducts.PARAM_ID_INSTA_WIN_PAYOUTS] = payouts
			params[featureProducts.PARAM_ID_INSTA_WIN_TYPE] = "tower"
			params[featureProducts.PARAM_ID_INSTA_WIN_SOURCE_ID] = 4
			params[featureProducts.PARAM_ID_INSTA_WIN_TILE_ID] = tileId
			params[featureProducts.PARAM_ID_INSTA_WIN_POSITIONS] = positions
			params[featureProducts.PARAM_ID_INSTA_WIN_INDEX] = "finish:1"
			params[PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_SCATTER_TRIGGER_TOWER] = true
			feature.ActivateFeatures(f.FeatureDef, state, params)
			delete(params, PARAM_ID_TRIGGER_WINS_PAYOUTS)
		}
	} else if len(newPositions) > 0 {

		feature.ActivateFeatures(f.FeatureDef, state, params)
	}

	return
}

func (f TriggerLawOfGilgameshTowerScatter) towerBonus(winsLevels []interface{}, probLevels []interface{}) (int, []int) {
	level := 0
	amount := 0
	payouts := []int{}
	for level < len(winsLevels) {
		win := feature.WeightedRandomIndex(feature.ConvertIntSlice(probLevels[level]))
		amount = feature.ConvertIntSlice(winsLevels[level])[win]
		if amount < 0 {
			payouts = append(payouts, 0)
			level++
		} else {
			payouts = append(payouts, amount)
			break
		}
	}
	return amount, payouts
}

func (f TriggerLawOfGilgameshTowerScatter) testBonusProbabilites(winsLevels []interface{}, probLevels []interface{}) {
	stats := make(map[int]int)
	num := 100000
	tot := 0
	for i := 0; i < num; i++ {
		a, _ := f.towerBonus(winsLevels, probLevels)
		tot += a
		n, ok := stats[a]
		if !ok {
			n = 0
		}
		stats[a] = n + 1
	}
	logger.Debugf("tower payout probabilities")
	logger.Debugf("--------------------------")
	for k, v := range stats {
		logger.Debugf("%d: %f", k, float32(v)/float32(num))
	}
	logger.Debugf("mean: %f", float32(tot)/float32(num))
}

func (f *TriggerLawOfGilgameshTowerScatter) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerLawOfGilgameshTowerScatter) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
