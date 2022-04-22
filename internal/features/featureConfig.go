package features

type Config struct {
	FeatureDef
	Data FeatureParams `json:"data"`
}

func (f *Config) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *Config) DataPtr() interface{} {
	return &f.Data
}

func (f *Config) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *Config) OnInit(state *FeatureState) {
}

func (f Config) Trigger(state *FeatureState, params FeatureParams) {
}

func (f *Config) Serialize() ([]byte, error) {
	return serializeFeatureToBytes(f)
}

func (f *Config) Deserialize(data []byte) (err error) {
	return deserializeFeatureFromBytes(f, data)
}
