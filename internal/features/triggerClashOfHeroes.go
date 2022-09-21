package features

type TriggerClashOfHeroes struct {
	FeatureDef
}

func (f *TriggerClashOfHeroes) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerClashOfHeroes) DataPtr() interface{} {
	return nil
}

func (f *TriggerClashOfHeroes) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *TriggerClashOfHeroes) OnInit(state *FeatureState) {
}

func (f TriggerClashOfHeroes) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}
	activateFeatures(f.FeatureDef, state, params)
	return
}

func (f TriggerClashOfHeroes) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *TriggerClashOfHeroes) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerClashOfHeroes) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
