package features

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type TriggerRandom struct {
	FeatureDef
}

func (f *TriggerRandom) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerRandom) DataPtr() interface{} {
	return nil
}

func (f *TriggerRandom) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *TriggerRandom) OnInit(state *FeatureState) {
}

func (f TriggerRandom) Trigger(state *FeatureState, params FeatureParams) {
	probability := params.GetInt("Probability")
	rand := rng.RandFromRange(10000)
	if rand < probability {
		logger.Debugf("TriggerRandom activate %d < %d", rand, probability)
		activateFeatures(f.FeatureDef, state, params)
	}
}

func (f TriggerRandom) ForceTrigger(state *FeatureState, params FeatureParams) {
}

func (f *TriggerRandom) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerRandom) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
