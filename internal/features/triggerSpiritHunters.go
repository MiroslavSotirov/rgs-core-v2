package features

type TriggerSpiritHunters struct {
	FeatureDef
}

func (f *TriggerSpiritHunters) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerSpiritHunters) DataPtr() interface{} {
	return nil
}

func (f *TriggerSpiritHunters) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f TriggerSpiritHunters) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}
	//	params["Random"] = rng.RandFromRange(f.FeatureDef.Params["RandomRange"].(int))
	activateFeatures(f.FeatureDef, state, params)
	return
}

func (f TriggerSpiritHunters) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *TriggerSpiritHunters) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerSpiritHunters) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
