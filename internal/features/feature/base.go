package feature

import "gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"

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
	logger.Debugf("feature.Base init")
	f.FeatureDef.Id = def.Id
	f.FeatureDef.Type = def.Type
	f.FeatureDef.Params = def.Params
	f.FeatureDef.Features = def.Features
	return nil
}

func (f *Base) OnInit(state *FeatureState) {
	logger.Debugf("feature.Base OnInit")
}

func (f Base) Trigger(state *FeatureState, params FeatureParams) {
	panic("feature.Base Trigger")
}
