package features

type TriggerWizardzWorld struct {
	FeatureDef
}

func (f *TriggerWizardzWorld) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerWizardzWorld) DataPtr() interface{} {
	return nil
}

func (f *TriggerWizardzWorld) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f TriggerWizardzWorld) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}
	activateFeatures(f.FeatureDef, state, params)
	return
}

func (f TriggerWizardzWorld) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *TriggerWizardzWorld) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerWizardzWorld) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
