package features

type InstaWinData struct {
	Coin int `json:"coin"`
}

type InstaWin struct {
	Def  FeatureDef   `json:"def"`
	Data InstaWinData `json:"data"`
}

func (f *InstaWin) DefPtr() *FeatureDef {
	return &f.Def
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
			Def: *f.DefPtr(),
			Data: InstaWinData{
				Coin: 100,
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
