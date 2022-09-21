package features

type SetConditional struct {
	FeatureDef
}

func (f *SetConditional) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *SetConditional) DataPtr() interface{} {
	return nil
}

func (f *SetConditional) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *SetConditional) OnInit(state *FeatureState) {
}

func (f SetConditional) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}

	conditionalFlag := params.GetString("ConditionalFlag")
	conditionalType := ""
	conditionalValue := ""
	if params.HasKey("ConditionalType") {
		conditionalType = params.GetString("ConditionalType")
	}
	if params.HasKey("ConditionalValue") {
		conditionalValue = params.GetString("ConditionalValue")
	}

	switch {
	case conditionalType == "bool" || conditionalType == "":
		params[conditionalFlag] = conditionalValue == "true" || conditionalType == ""
	default:
		panic("SetConditional. unknown conditional type")
	}
	return
}

func (f SetConditional) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *SetConditional) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *SetConditional) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
