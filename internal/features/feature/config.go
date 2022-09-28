package feature

const (
	FEATURE_ID_CONFIG = "Config"
)

var _ Factory = RegisterFeature(FEATURE_ID_CONFIG, func() Feature { return new(Config) })

type Config struct {
	Base
	Data FeatureParams `json:"data"`
}

func (f *Config) DataPtr() interface{} {
	return &f.Data
}

func (f *Config) Serialize() ([]byte, error) {
	return SerializeFeatureToBytes(f)
}

func (f *Config) Deserialize(data []byte) (err error) {
	return DeserializeFeatureFromBytes(f, data)
}
