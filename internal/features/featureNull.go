package features

type FeatureNull struct {
	FeatureDef
	Data FeatureParams `json:"data"`
}

func (f *FeatureNull) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *FeatureNull) DataPtr() interface{} {
	return &f.Data
}

func (f *FeatureNull) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *FeatureNull) OnInit(state *FeatureState) {
}

func (f FeatureNull) Trigger(state *FeatureState, params FeatureParams) {
}

func (f *FeatureNull) Serialize() ([]byte, error) {
	return serializeFeatureToBytes(f)
}

func (f *FeatureNull) Deserialize(data []byte) (err error) {
	return deserializeFeatureFromBytes(f, data)
}
