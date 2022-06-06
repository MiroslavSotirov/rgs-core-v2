package features

import (
	"fmt"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type TriggerBattleOfMythsFreespin struct {
	FeatureDef
}

func (f *TriggerBattleOfMythsFreespin) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerBattleOfMythsFreespin) DataPtr() interface{} {
	return nil
}

func (f *TriggerBattleOfMythsFreespin) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *TriggerBattleOfMythsFreespin) OnInit(state *FeatureState) {
	state.Features = append(state.Features,
		&Config{
			FeatureDef: *f.DefPtr(),
			Data:       f.DefPtr().Params,
		})
}

func (f TriggerBattleOfMythsFreespin) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}

	var counter int
	var fstype string
	statefulStake := GetStatefulStakeMap(*state)
	if statefulStake.HasKey("counter") {
		counter = statefulStake.GetInt("counter")
		fstype = statefulStake.GetString("fstype")
	}

	scatterInc := params.GetInt("ScatterInc")
	scatterDec := params.GetInt("ScatterDec")
	scatterMin := params.GetInt("ScatterMin")
	scatterMax := params.GetInt("ScatterMax")
	numFreespins := params.GetInt("NumFreespins")

	if counter >= scatterMax || counter <= scatterMin {
		counter = 0
	}

	positions := []int{}
	for x, r := range state.SymbolGrid {
		for y, s := range r {
			any := false
			if s == scatterInc {
				counter++
				any = true
			}
			if s == scatterDec {
				counter--
				any = true
			}
			if any {
				positions = append(positions, x*len(r)+y)
			}
		}
	}

	if counter >= scatterMax || counter <= scatterMin {
		// add freespin win to state.Wins

		if numFreespins > 0 {
			fstype = "freespinE1"
			if counter >= scatterMax {
				fstype = "freespinE2"
			}

			state.Wins = append(state.Wins, FeatureWin{
				Index:           fmt.Sprintf("%s:%d", fstype, numFreespins),
				SymbolPositions: positions,
			})
			logger.Debugf("Trigger %d freespins of type %s", numFreespins, fstype)
		}
	}

	SetStatefulStakeMap(*state, FeatureParams{
		"counter": counter,
		"fstype":  fstype,
	}, params)

	activateFeatures(f.FeatureDef, state, params)
	return
}

func (f TriggerBattleOfMythsFreespin) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *TriggerBattleOfMythsFreespin) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerBattleOfMythsFreespin) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
