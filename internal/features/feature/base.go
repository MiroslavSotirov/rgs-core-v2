package feature

type Base struct {
	FeatureDef
}

func (f *Base) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *Base) DataPtr() interface{} {
	return nil
}

func (f *Base) Init(def FeatureDef) error {
	f.FeatureDef.Id = def.Id
	f.FeatureDef.Type = def.Type
	f.FeatureDef.Params = def.Params
	f.FeatureDef.Features = def.Features
	return nil
}

func (f *Base) OnInit(state *FeatureState) {
}

func (f Base) Trigger(state *FeatureState, params FeatureParams) {
	panic("feature.Base Trigger")
}
