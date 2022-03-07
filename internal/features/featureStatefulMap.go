package features

type StatefulMapData struct {
	Map FeatureParams `json:"map"`
}

type StatefulMap struct {
	FeatureDef
	Data StatefulMapData `json:"data"`
}

func (f *StatefulMap) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *StatefulMap) DataPtr() interface{} {
	return &f.Data
}

func (f *StatefulMap) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f StatefulMap) Trigger(state *FeatureState, params FeatureParams) {
	sfmap := params.GetParams("StatefulMap")
	state.Features = append(state.Features,
		&StatefulMap{
			FeatureDef: *f.DefPtr(),
			Data:       StatefulMapData{Map: sfmap},
		})
}

func (f *StatefulMap) Serialize() ([]byte, error) {
	return serializeFeatureToBytes(f)
}

func (f *StatefulMap) Deserialize(data []byte) (err error) {
	return deserializeFeatureFromBytes(f, data)
}
