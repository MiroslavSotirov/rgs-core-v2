package features

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

type TriggerSupaCrew struct {
	FeatureDef
}

func (f *TriggerSupaCrew) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerSupaCrew) DataPtr() interface{} {
	return nil
}

func (f *TriggerSupaCrew) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f TriggerSupaCrew) Trigger(state *FeatureState, params FeatureParams) {
	params["Random"] = rng.RandFromRange(f.FeatureDef.Params["RandomRange"].(int))
	activateFeatures(f.FeatureDef, state, params)
	return
}

func (f *TriggerSupaCrew) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerSupaCrew) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
