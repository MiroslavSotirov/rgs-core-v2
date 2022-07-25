package features

type TriggerSwordKing struct {
	FeatureDef
}

func (f *TriggerSwordKing) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerSwordKing) DataPtr() interface{} {
	return nil
}

func (f *TriggerSwordKing) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *TriggerSwordKing) OnInit(state *FeatureState) {
}

func (f TriggerSwordKing) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}

	counter := 0
	statefulStake := GetStatefulStakeMap(*state)
	if statefulStake.HasKey("counter") {
		counter = statefulStake.GetInt("counter")
	}

	SetStatefulStakeMap(*state, FeatureParams{"counter": counter},
		params)

	activateFeatures(f.FeatureDef, state, params)
	return
}

func (f TriggerSwordKing) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *TriggerSwordKing) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerSwordKing) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
