package featureTriggers

import (
	"fmt"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

const (
	FEATURE_ID_TRIGGER_ELYSIUM_VIP = "TriggerElysiumVip"

	PARAM_ID_TRIGGER_ELYSIUM_VIP_WILD_ID = "WildId"

	STATEFUL_ID_TRIGGER_ELYSIUM_VIP_LEVEL     = "level"
	STATEFUL_ID_TRIGGER_ELYSIUM_VIP_INSERTS   = "inserts"
	STATEFUL_ID_TRIGGER_ELYSIUM_VIP_ORIGINALS = "originals"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_ELYSIUM_VIP, func() feature.Feature { return new(TriggerElysiumVip) })

type TriggerElysiumVip struct {
	feature.Base
}

func (f TriggerElysiumVip) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	level := 0
	inserts := []int{}
	originals := []int{0, 1, 2}
	if state.Action != "base" {
		statefulStake := feature.GetStatefulStakeMap(*state)
		logger.Debugf("statefulStake: %#v", statefulStake)
		if statefulStake.HasKey(STATEFUL_ID_TRIGGER_ELYSIUM_VIP_LEVEL) {
			level = statefulStake.GetInt(STATEFUL_ID_TRIGGER_ELYSIUM_VIP_LEVEL)
		}
		if statefulStake.HasKey(STATEFUL_ID_TRIGGER_ELYSIUM_VIP_INSERTS) {
			inserts = feature.ConvertIntSlice(statefulStake.GetSlice(STATEFUL_ID_TRIGGER_ELYSIUM_VIP_INSERTS))
		}
		if statefulStake.HasKey(STATEFUL_ID_TRIGGER_ELYSIUM_VIP_ORIGINALS) {
			originals = statefulStake.GetIntSlice(STATEFUL_ID_TRIGGER_ELYSIUM_VIP_ORIGINALS)
		}
		if len(state.SymbolGrid) != len(state.Stateful.SymbolGrid)+len(inserts) {
			panic(fmt.Sprintf("number of reels %d is not last spin num %d plus num inserts %d",
				len(state.SymbolGrid), len(state.Stateful.SymbolGrid), len(inserts)))
		}

		logger.Debugf("copying wilds from last spin")
		wildId := params.GetInt(PARAM_ID_TRIGGER_ELYSIUM_VIP_WILD_ID)
		ireel := 0
		for ilreel, lreel := range state.Stateful.SymbolGrid {
			if func() bool {
				for _, ins := range inserts {
					if ins == ilreel {
						return true
					}
				}
				return false
			}() {
				ireel++
			}
			for isym, lsym := range lreel {
				if lsym == wildId {
					logger.Debugf("setting reel %d row %d to %d", ireel, isym, wildId)
					state.SymbolGrid[ireel][isym] = wildId
				}
			}
			ireel++
		}
		logger.Debugf("%v", state.SymbolGrid)
		//		inserts := []int{}
	}

	feature.SetStatefulStakeMap(*state, feature.FeatureParams{
		STATEFUL_ID_TRIGGER_ELYSIUM_VIP_LEVEL:     level,
		STATEFUL_ID_TRIGGER_ELYSIUM_VIP_INSERTS:   inserts,
		STATEFUL_ID_TRIGGER_ELYSIUM_VIP_ORIGINALS: originals},
		params)

	feature.ActivateFeatures(f.FeatureDef, state, params)
	return
}

func (f *TriggerElysiumVip) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerElysiumVip) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
