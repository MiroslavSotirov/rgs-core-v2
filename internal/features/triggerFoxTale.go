package features

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

type TriggerFoxTale struct {
	FeatureDef
}

func (f *TriggerFoxTale) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerFoxTale) DataPtr() interface{} {
	return nil
}

func (f *TriggerFoxTale) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *TriggerFoxTale) OnInit(state *FeatureState) {
}

func (f TriggerFoxTale) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}
	params["Random"] = rng.RandFromRange(f.FeatureDef.Params["RandomRange"].(int))
	activateFeatures(f.FeatureDef, state, params)
	return
}

func (f TriggerFoxTale) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *TriggerFoxTale) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerFoxTale) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
