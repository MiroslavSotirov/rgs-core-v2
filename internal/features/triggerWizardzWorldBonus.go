package features

import (
	"fmt"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type TriggerWizardzWorldBonus struct {
	FeatureDef
}

func (f *TriggerWizardzWorldBonus) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerWizardzWorldBonus) DataPtr() interface{} {
	return nil
}

func (f *TriggerWizardzWorldBonus) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f TriggerWizardzWorldBonus) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}
	tileIds := params.GetIntSlice("TileIds")
	limits := params.GetIntSlice("Limits")
	counters := make([]int, len(limits))

	counterName := func(idx int) string {
		return fmt.Sprintf("counter%d", idx+1)
	}
	var statefulMap FeatureParams = FeatureParams{}
	stake := fmt.Sprintf("%.3f", state.TotalStake)
	if state.Stateful != nil {
		sf := FindFeature("StatefulMap", state.Stateful.Features)
		if sf != nil {
			sfmap := sf.(*StatefulMap)
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
	logger.Debugf("counters: %v incremented: %v amounts: %v limited: %v", counters, incremented, amounts, limited)

	if len(limited) > 1 {
		inudge := rng.RandFromRange(len(limited))
		tileId := tileIds[inudge]
		logger.Infof("Nudge on %s (symbol %d) due to both counters reaching limit", counterName(inudge), tileId)
		logger.Debugf("SourceGrid: %v", state.SourceGrid)
		logger.Debugf("SymbolGrid: %v", state.SourceGrid)
		for x, r := range state.SymbolGrid {
			full := true
			for _, s := range r {
				if s != tileId {
					full = false
					break
				}
			}
			if full {
				ofs := []int{-2, -1, 1, 2}[rng.RandFromRange(4)]
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
		logger.Debugf("SourceGrid: %v", state.SourceGrid)
		logger.Debugf("SymbolGrid: %v", state.SourceGrid)
	}

	for i, c := range incremented {
		counters[c] += amounts[i]
		logger.Debugf("%s is %d after an increase of %d", counterName(c), counters[c], amounts[i])
	}

	if len(limited) > 0 {
		c := limited[0]
		logger.Debugf("activating %s with tileId %d", counterName(c), tileIds[c])
		params["TileId"] = tileIds[c]
		activateFeatures(f.FeatureDef, state, params)
	}

	stakeMap := make(map[string]interface{}, len(counters))
	for i := range counters {
		stakeMap[counterName(i)] = counters[i]
		stakeMap[fmt.Sprintf("min%d", i+1)] = 0
		stakeMap[fmt.Sprintf("max%d", i+1)] = limits[i]
	}
	statefulMap[stake] = stakeMap
	params["StatefulMap"] = statefulMap
	logger.Debugf("statefulMap= %#v", statefulMap)

	return
}

func (f TriggerWizardzWorldBonus) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *TriggerWizardzWorldBonus) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerWizardzWorldBonus) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
