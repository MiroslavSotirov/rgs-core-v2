package featureTriggers

import (
	"fmt"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

const (
	FEATURE_ID_TRIGGER_WIZARDZ_WORLD_BONUS = "TriggerWizardzWorldBonus"

	PARAM_ID_TRIGGER_WIZARDZ_WORLD_BONUS_TILE_IDS = "TileIds"
	PARAM_ID_TRIGGER_WIZARDZ_WORLD_BONUS_LIMITS   = "Limits"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_WIZARDZ_WORLD_BONUS, func() feature.Feature { return new(TriggerWizardzWorldBonus) })

type TriggerWizardzWorldBonus struct {
	feature.Base
}

func (f TriggerWizardzWorldBonus) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	tileIds := params.GetIntSlice(PARAM_ID_TRIGGER_WIZARDZ_WORLD_BONUS_TILE_IDS)
	limits := params.GetIntSlice(PARAM_ID_TRIGGER_WIZARDZ_WORLD_BONUS_LIMITS)
	counters := make([]int, len(limits))

	counterName := func(idx int) string {
		return fmt.Sprintf("counter%d", idx+1)
	}
	var statefulMap feature.FeatureParams = feature.FeatureParams{}
	stake := fmt.Sprintf("%.3f", state.TotalStake)
	if state.Stateful != nil {
		sf := feature.FindFeature(feature.FEATURE_ID_STATEFUL_MAP, state.Stateful.Features)
		if sf != nil {
			sfmap := sf.(*feature.StatefulMap)
			if sfmap != nil {
				for k, v := range sfmap.Data.Map {
					statefulMap[k] = v
				}
				if sfmap.Data.Map.HasKey(stake) {
					reset := false
					for i := range counters {
						counters[i] = sfmap.Data.Map.GetParams(stake).GetInt(counterName(i))
						if counters[i] >= limits[i] {
							reset = true
						}
					}
					if reset {
						for i := range counters {
							counters[i] = 0
						}
					}
				}
			} else {
				logger.Errorf("StatefulMap has wrong type")
			}
		} else {
			logger.Warnf("no StatefulMap in previous gamestate")
		}
	} else {
		logger.Warnf("feature state is not stateful")
	}

	incremented := []int{}
	amounts := []int{}
	for i := range counters {
		tileId := tileIds[i]
		for _, r := range state.SymbolGrid {
			full := true
			for _, s := range r {
				if s != tileId {
					full = false
					break
				}
			}
			if full {
				if len(incremented) > 0 && incremented[len(incremented)-1] == i {
					amounts[len(incremented)-1]++
				} else {
					incremented = append(incremented, i)
					amounts = append(amounts, 1)
				}
			}
		}
	}

	limited := []int{}
	for i, c := range incremented {
		if counters[c]+amounts[i] >= limits[i] {
			limited = append(limited, c)
		}
	}

	if len(limited) > 1 {
		inudge := rng.RandFromRangePool(len(limited))
		tileId := tileIds[inudge]
		logger.Infof("Nudge on %s (symbol %d) due to both counters reaching limit", counterName(inudge), tileId)
		for x, r := range state.SymbolGrid {
			full := true
			for _, s := range r {
				if s != tileId {
					full = false
					break
				}
			}
			if full {
				ofs := []int{-2, -1, 1, 2}[rng.RandFromRangePool(4)]
				num := len(state.Reels[x])
				if ofs < 0 {
					ofs += len(state.Reels[x])
				}
				for y := range r {
					s := state.Reels[x][(state.StopList[x]+ofs+y)%num]
					state.SourceGrid[x][y] = s
					state.SymbolGrid[x][y] = s

				}
			}
		}
		idx := 0
		for i, v := range incremented {
			if v == inudge {
				idx = i
				break
			}
		}
		incremented = append(incremented[:idx], incremented[idx+1:]...)
		amounts = append(amounts[:idx], amounts[idx+1:]...)
		for i, v := range limited {
			if v == inudge {
				idx = i
				break
			}
		}
		limited = append(limited[:idx], limited[idx+1:]...)
		logger.Debugf("after nudge. counters: %v incremented: %v amounts: %v limited: %v", counters, incremented, amounts, limited)
	}

	for i, c := range incremented {
		counters[c] += amounts[i]
	}

	if len(limited) > 0 {
		c := limited[0]
		params[featureProducts.PARAM_ID_INSTA_WIN_TILE_ID] = tileIds[c]
		feature.ActivateFeatures(f.FeatureDef, state, params)
	}

	stakeMap := make(map[string]interface{}, len(counters))
	for i := range counters {
		stakeMap[counterName(i)] = counters[i]
		stakeMap[fmt.Sprintf("min%d", i+1)] = 0
		stakeMap[fmt.Sprintf("max%d", i+1)] = limits[i]
	}
	statefulMap[stake] = stakeMap
	params[feature.FEATURE_ID_STATEFUL_MAP] = statefulMap

	return
}

func (f *TriggerWizardzWorldBonus) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerWizardzWorldBonus) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
