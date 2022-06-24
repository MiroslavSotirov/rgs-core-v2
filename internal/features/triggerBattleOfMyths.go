package features

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type TriggerBattleOfMyths struct {
	FeatureDef
}

func (f *TriggerBattleOfMyths) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerBattleOfMyths) DataPtr() interface{} {
	return nil
}

func (f *TriggerBattleOfMyths) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *TriggerBattleOfMyths) OnInit(state *FeatureState) {
}

func (f TriggerBattleOfMyths) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}

	runPrincess := false
	for _, f := range state.Features {
		if f.DefPtr().Type == "Princess" {
			logger.Infof("Princess on board. Do not run tiger or dragon")
			runPrincess = true
		}
	}

	params["RunPrincess"] = runPrincess
	if !runPrincess {
		featureProb := params.GetInt("FeatureProbability")
		random := rng.RandFromRange(10000)
		if random < featureProb {
			random = rng.RandFromRange(2)
			if random == 0 {
				params["RunTiger"] = true
				logger.Debugf("RunTiger is true")
			} else {
				params["RunDragon"] = true
				logger.Debugf("RunDragon is true")
			}
		}
	}

	activateFeatures(f.FeatureDef, state, params)
	return
}

func (f TriggerBattleOfMyths) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *TriggerBattleOfMyths) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerBattleOfMyths) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
