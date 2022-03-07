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

	if state.Stateful != nil {
		sf := FindFeature("StatefulMap", state.Stateful.Features)
		if sf != nil {
			sfmap := sf.(*StatefulMap)
			if sfmap != nil {
				reset := false
				for i := range counters {
					counters[i] = sfmap.Data.Map.GetInt(counterName(i))
					if counters[i] >= limits[i] {
						reset = true
					}
				}
				if reset {
					for i := range counters {
						counters[i] = 0
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
				incremented = append(incremented, i)
			}
		}
	}

	if len(incremented) > 1 {
		inudge := rng.RandFromRange(len(incremented))
		tileId := tileIds[inudge]
		for _, r := range state.SymbolGrid {
			full := true
			for _, s := range r {
				if s != tileId {
					full = false
					break
				}
			}
			if full {
				logger.Warnf("STUB: Nudge when both counters reach limit is not implemented")
			}
		}
	}

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
				counters[i]++
			}
		}
		logger.Debugf("%s is %d", counterName(i), counters[i])
	}

	actcounter := -1
	for i := range counters {
		if counters[i] >= limits[i] {
			if actcounter < 0 {
				actcounter = i
			} else {

			}
		}
	}

	if actcounter >= 0 {
		logger.Debugf("activating counter %d with tileId %d", actcounter, tileIds[actcounter])
		params["TileId"] = tileIds[actcounter]
		activateFeatures(f.FeatureDef, state, params)
	}

	statefulMap := make(map[string]interface{}, len(counters))
	for i := range counters {
		statefulMap[counterName(i)] = counters[i]
	}
	params["StatefulMap"] = statefulMap

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
