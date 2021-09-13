package features

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

type TriggerFoxTail struct {
	FeatureDef
}

func (f *TriggerFoxTail) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerFoxTail) DataPtr() interface{} {
	return nil
}

func (f *TriggerFoxTail) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f TriggerFoxTail) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}
	params["Random"] = rng.RandFromRange(f.FeatureDef.Params["RandomRange"].(int))
	activateFeatures(f.FeatureDef, state, params)
	return
}

func (f TriggerFoxTail) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *TriggerFoxTail) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerFoxTail) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
