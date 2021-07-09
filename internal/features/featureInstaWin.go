package features

type InstaWinData struct {
	Type     string `json:"type"`
	SourceId int32  `json:"sourceid"`
	Amount   int64  `json:"amount"`
}

type InstaWin struct {
	FeatureDef
	Data InstaWinData `json:"data"`
}

func (f *InstaWin) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *InstaWin) DataPtr() interface{} {
	return &f.Data
}

func (f *InstaWin) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f InstaWin) Trigger(featurestate FeatureState, params FeatureParams) []Feature {
	return []Feature{
		&InstaWin{
			FeatureDef: *f.DefPtr(),
			Data: InstaWinData{
				Type:     params.GetString("InstaWinType"),
				SourceId: params.GetInt32("InstaWinSourceId"),
				Amount:   params.GetInt64("InstaWinAmount"),
			},
		},
	}
}

func (f *InstaWin) Serialize() ([]byte, error) {
	return serializeFeatureToBytes(f)
}

func (f *InstaWin) Deserialize(data []byte) (err error) {
	return deserializeFeatureFromBytes(f, data)
}
