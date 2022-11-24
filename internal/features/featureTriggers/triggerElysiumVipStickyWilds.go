package featureTriggers

import (
	"sort"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

const (
	FEATURE_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS = "TriggerElysiumVipStickyWilds"

	PARAM_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS_TILE_ID                  = "TileId"
	PARAM_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS_NUM_WILDS_LEVELS         = "NumWildsLevels"
	PARAM_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS_NUM_PROBABILITIES_LEVELS = "NumProbabilitiesLevels"
	//	PARAM_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS_REEL_PROBABILITIES_LEVELS = "ReelProbabilitiesLevels"
	PARAM_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS_RETRY_FACTOR       = "RetryFactor"
	PARAM_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS_PROBABILITY_LEVELS = "ProbabilityLevels"
	PARAM_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS_REEL_PROBABILITIES = "ReelProbabilities"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS, func() feature.Feature { return new(TriggerElysiumVipStickyWilds) })

type TriggerElysiumVipStickyWilds struct {
	feature.Base
}

func (f TriggerElysiumVipStickyWilds) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	level := 0
	inserts := []int{}
	originals := []int{0, 1, 2}
	stakeMap := feature.GetParamStakeMap(*state, params)
	if stakeMap.HasKey(STATEFUL_ID_TRIGGER_ELYSIUM_VIP_LEVEL) {
		level = stakeMap.GetInt(STATEFUL_ID_TRIGGER_ELYSIUM_VIP_LEVEL)
	}
	if stakeMap.HasKey(STATEFUL_ID_TRIGGER_ELYSIUM_VIP_INSERTS) {
		inserts = stakeMap.GetIntSlice(STATEFUL_ID_TRIGGER_ELYSIUM_VIP_INSERTS)
	}
	if stakeMap.HasKey(STATEFUL_ID_TRIGGER_ELYSIUM_VIP_ORIGINALS) {
		originals = stakeMap.GetIntSlice(STATEFUL_ID_TRIGGER_ELYSIUM_VIP_ORIGINALS)
	}
	probabilityLevels := params.GetIntSlice(PARAM_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS_PROBABILITY_LEVELS)
	if rng.RandFromRange(10000) < probabilityLevels[level] {

		gridw := len(state.SymbolGrid)
		gridh := len(state.SymbolGrid[0])
		wildId := params.GetInt(PARAM_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS_TILE_ID)

		numWildsLevels := params.GetSlice(
			PARAM_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS_NUM_WILDS_LEVELS)
		numWildsLevel := feature.ConvertIntSlice(numWildsLevels[level])
		numProbabilitiesLevel := feature.ConvertIntSlice(params.GetSlice(
			PARAM_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS_NUM_PROBABILITIES_LEVELS)[level])
		reelProbabilities := params.GetIntSlice(PARAM_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS_REEL_PROBABILITIES)

		numWilds := numWildsLevel[feature.WeightedRandomIndex(numProbabilitiesLevel)]
		numTries := params.GetInt(PARAM_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS_RETRY_FACTOR) * numWilds
		positions := []int{}

		isWild := func(reel int, row int) bool {
			if state.SymbolGrid[reel][row] == wildId {
				return true
			}
			pos := reel*gridh + row
			for _, p := range positions {
				if p == pos {
					return true
				}
			}
			return false
		}

		isOriginal := func(reel int) bool {
			for _, o := range originals {
				if o == reel {
					return true
				}
			}
			return false
		}

		for try := 0; len(positions) < numWilds && try < numTries+1; try++ {
			var reelidx int
			if level == 0 {
				reelidx = feature.WeightedRandomIndex(reelProbabilities)
			} else {
				reelidx = rng.RandFromRange(gridw)
			}
			rowidx := rng.RandFromRange(3)
			pos := reelidx*gridh + rowidx
			if !isWild(reelidx, rowidx) {
				positions = append(positions, pos)
			}
		}

		if len(positions) > 0 {

			inserts = []int{}
			isInserted := func(ireel int) bool {
				for _, ins := range inserts {
					if ins == ireel {
						return true
					}
				}
				return false
			}

			for _, p := range positions {
				reelidx := p / gridh
				rowidx := p % gridh
				if isOriginal(reelidx) {
					if reelidx > 0 && isOriginal(reelidx-1) && isWild(reelidx-1, rowidx) && !isInserted(reelidx) {
						inserts = append(inserts, reelidx)
					}
					if reelidx < gridw-1 && isOriginal(reelidx+1) && isWild(reelidx+1, rowidx) && !isInserted(reelidx+1) {
						inserts = append(inserts, reelidx+1)
					}
				}
			}

			params[featureProducts.PARAM_ID_RESPIN_AMOUNT] = 1
			numreels := gridw + len(inserts)
			action := ""
			switch numreels {
			case 3:
				action = "respinall1"
			case 4:
				action = "respinall2"
			case 5:
				action = "respinall3"
			}
			params[featureProducts.PARAM_ID_RESPIN_ACTION] = action

			level++
			if level >= len(numWildsLevels) {
				level = len(numWildsLevels) - 1
			}
			if len(inserts) > 0 {
				sort.Slice(inserts, func(i, j int) bool { return i < j })

				for im, m := range originals {
					originals[im] = func() int {
						c := 0
						for _, i := range inserts {
							if i <= m {
								c++
							}
						}
						return m + c
					}()
				}
			}

			params[featureProducts.PARAM_ID_REPLACE_TILE_POSITIONS] = positions

			feature.SetStatefulStakeMap(*state, feature.FeatureParams{
				STATEFUL_ID_TRIGGER_ELYSIUM_VIP_LEVEL:     level,
				STATEFUL_ID_TRIGGER_ELYSIUM_VIP_INSERTS:   inserts,
				STATEFUL_ID_TRIGGER_ELYSIUM_VIP_ORIGINALS: originals},
				params)

			feature.ActivateFeatures(f.FeatureDef, state, params)
		}
	}
	return
}

func (f *TriggerElysiumVipStickyWilds) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerElysiumVipStickyWilds) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
