package featureTriggers

import (
	"fmt"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

const (
	FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_BONUS = "TriggerLawOfGilgameshTowerBonus"

	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_BONUS_TILE_ID       = "TileId"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_BONUS_WINS_LEVELS   = "WinsLevels"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_BONUS_PROB_LEVELS   = "ProbabilitiesLevels"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_BONUS_THRESHOLD     = "BonusThreshold"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_BONUS_TRIGGER_TOWER = "TriggerTower"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_BONUS, func() feature.Feature { return new(TriggerLawOfGilgameshTowerBonus) })

type TriggerLawOfGilgameshTowerBonus struct {
	feature.Base
}

func (f TriggerLawOfGilgameshTowerBonus) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	tileId := params.GetInt(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_BONUS_TILE_ID)
	bonusThreshold := params.GetInt(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_BONUS_THRESHOLD)

	gridh := len(state.SourceGrid[0])
	positions := []int{}
	for reel, r := range state.SymbolGrid {
		for row, s := range r {
			if s == tileId {
				positions = append(positions, reel*gridh+row)
			}
		}
	}

	if len(positions) >= bonusThreshold {

		winsLevels := params.GetSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_BONUS_WINS_LEVELS)
		probLevels := params.GetSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_BONUS_PROB_LEVELS)

		amount, payouts := f.towerBonus(winsLevels, probLevels)

		if amount > 0 {

			if state.Multiplier > 1 {
				panic(fmt.Sprintf("tower bonus with multiplier %d and amount %d", state.Multiplier, amount))
			}
			symbols := make([]int, len(positions))
			for s := range positions {
				symbols[s] = tileId
			}
			params[featureProducts.PARAM_ID_INSTA_WIN_AMOUNT] = amount
			params[featureProducts.PARAM_ID_INSTA_WIN_PAYOUTS] = payouts
			params[featureProducts.PARAM_ID_INSTA_WIN_TYPE] = "tower"
			params[featureProducts.PARAM_ID_INSTA_WIN_SOURCE_ID] = 4
			params[featureProducts.PARAM_ID_INSTA_WIN_TILE_ID] = tileId
			params[featureProducts.PARAM_ID_INSTA_WIN_POSITIONS] = positions
			//			params[featureProducts.PARAM_ID_INSTA_WIN_INDEX] = "finish:1"
			params[featureProducts.PARAM_ID_INSTA_WIN_INDEX] = ""
			params[featureProducts.PARAM_ID_INSTA_WIN_SYMBOLS] = symbols
			params[PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_TOWER_BONUS_TRIGGER_TOWER] = true
			feature.ActivateFeatures(f.FeatureDef, state, params)
			delete(params, PARAM_ID_TRIGGER_WINS_PAYOUTS)
		}
	}

	return
}

func (f TriggerLawOfGilgameshTowerBonus) towerBonus(winsLevels []interface{}, probLevels []interface{}) (int, []int) {
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

func (f TriggerLawOfGilgameshTowerBonus) testBonusProbabilites(winsLevels []interface{}, probLevels []interface{}) {
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

func (f *TriggerLawOfGilgameshTowerBonus) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerLawOfGilgameshTowerBonus) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
