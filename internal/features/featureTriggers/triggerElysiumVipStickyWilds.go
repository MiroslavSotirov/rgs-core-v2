package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

const (
	FEATURE_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS = "TriggerElysiumVipStickyWilds"

	PARAM_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS_TILE_ID                   = "TileId"
	PARAM_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS_NUM_WILDS_LEVELS          = "NumWildsLevels"
	PARAM_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS_NUM_PROBABILITIES_LEVELS  = "NumProbabilitiesLevels"
	PARAM_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS_REEL_PROBABILITIES_LEVELS = "ReelProbabilitiesLevels"
	PARAM_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS_RETRY_FACTOR              = "RetryFactor"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS, func() feature.Feature { return new(TriggerElysiumVipStickyWilds) })

type TriggerElysiumVipStickyWilds struct {
	feature.Base
}

func (f TriggerElysiumVipStickyWilds) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	level := 0
	inserts := []int{}
	statefulStake := feature.GetStatefulStakeMap(*state)
	if statefulStake.HasKey(STATEFUL_ID_TRIGGER_ELYSIUM_VIP_LEVEL) {
		level = statefulStake.GetInt(STATEFUL_ID_TRIGGER_ELYSIUM_VIP_LEVEL)
	}
	if statefulStake.HasKey(STATEFUL_ID_TRIGGER_ELYSIUM_VIP_INSERTS) {
		inserts = statefulStake.GetIntSlice(STATEFUL_ID_TRIGGER_ELYSIUM_VIP_INSERTS)
	}

	gridw := len(state.SymbolGrid)
	gridh := len(state.SymbolGrid[0])
	wildId := params.GetInt(PARAM_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS_TILE_ID)

	/*
		if strings.Contains(state.Action, "respinall") {

		}
	*/
	numWildsLevel := feature.ConvertIntSlice(params.GetSlice(
		PARAM_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS_NUM_WILDS_LEVELS)[level])
	numProbabilitiesLevel := feature.ConvertIntSlice(params.GetSlice(
		PARAM_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS_NUM_PROBABILITIES_LEVELS)[level])
	//	reelProbabilitiesLevel := feature.ConvertIntSlice(params.GetSlice(
	//		PARAM_ID_TRIGGER_ELYSIUM_VIP_STICKY_WILDS_REEL_PROBABILITIES_LEVELS)[level])

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

	for try := 0; len(positions) < numWilds && try < numTries+1; try++ {
		//		reelidx := reelConfig[feature.WeightedRandomIndex(reelProbabilitiesLevel)]
		reelidx := rng.RandFromRange(gridw)
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
			if reelidx > 0 && isWild(reelidx-1, rowidx) && !isInserted(reelidx) {
				inserts = append(inserts, reelidx)
			}
			if reelidx < gridw-1 && isWild(reelidx+1, rowidx) && !isInserted(reelidx+1) {
				inserts = append(inserts, reelidx+1)
			}
		}

		params[featureProducts.PARAM_ID_RESPIN_AMOUNT] = 1
		numreels := gridw + len(inserts)
		if numreels > 5 {
			numreels = 5
		}
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
		if level >= len(numWildsLevel) {
			level = len(numWildsLevel) - 1
		}
		feature.SetStatefulStakeMap(*state, feature.FeatureParams{
			STATEFUL_ID_TRIGGER_ELYSIUM_VIP_LEVEL:   level,
			STATEFUL_ID_TRIGGER_ELYSIUM_VIP_INSERTS: inserts},
			params)

		params[featureProducts.PARAM_ID_REPLACE_TILE_POSITIONS] = positions
		feature.ActivateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f *TriggerElysiumVipStickyWilds) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerElysiumVipStickyWilds) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
