package features

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

type TriggerSupaCrew struct {
	Def FeatureDef `json:"def"`
}

func (f *TriggerSupaCrew) DefPtr() *FeatureDef {
	return &f.Def
}

func (f *TriggerSupaCrew) DataPtr() interface{} {
	return nil
}

func (f *TriggerSupaCrew) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f TriggerSupaCrew) Trigger(state FeatureState, params FeatureParams) []Feature {
	params["Random"] = rng.RandFromRange(f.Def.Params["RandomRange"].(int))
	return activateFeatures(f.Def, state, params)
}

func (f *TriggerSupaCrew) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerSupaCrew) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
