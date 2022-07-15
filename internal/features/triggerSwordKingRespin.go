package features

import (
	"fmt"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type TriggerSwordKingRespin struct {
	FeatureDef
}

func (f *TriggerSwordKingRespin) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerSwordKingRespin) DataPtr() interface{} {
	return nil
}

func (f *TriggerSwordKingRespin) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *TriggerSwordKingRespin) OnInit(state *FeatureState) {
}

func (f TriggerSwordKingRespin) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}

	if state.PureWins {
		logger.Debugf("skipping respin due to wins")
		return
	} else if params.HasKey("RunWilds") && params.GetBool("RunWilds") {
		logger.Debugf("skipping respin due to random wilds")
		return
	}

	Probability := params.GetInt("Probability")
	if rng.RandFromRange(10000) < Probability {

		fstype := params.GetString("FSType")
		numFreespins := params.GetInt("NumFreespins")

		logger.Debugf("Respin trigger %d freespins of type %s", numFreespins, fstype)

		state.Wins = append(state.Wins, FeatureWin{
			Index: fmt.Sprintf("%s:%d", fstype, numFreespins),
		})

		params["RunRespin"] = true
		activateFeatures(f.FeatureDef, state, params)
	}

	return
}

func (f TriggerSwordKingRespin) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *TriggerSwordKingRespin) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerSwordKingRespin) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
