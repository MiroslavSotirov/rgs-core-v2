package features

import "gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"

type TriggerConditional struct {
	FeatureDef
}

func (f *TriggerConditional) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerConditional) DataPtr() interface{} {
	return nil
}

func (f *TriggerConditional) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *TriggerConditional) OnInit(state *FeatureState) {
}

func (f TriggerConditional) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}

	conditionalFlag := params.GetString("ConditionalFlag")
	condition := ""
	if params.HasKey("Condition") {
		condition = params.GetString("Condition")
	}

	activate := false
	switch {
	case condition == "Set" || condition == "":
		activate = params.HasKey(conditionalFlag)
	case condition == "Unset":
		activate = !params.HasKey(conditionalFlag)
	}
	if activate {
		logger.Debugf("Trigger %s with features %#v", conditionalFlag, f.FeatureDef)
		activateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f TriggerConditional) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *TriggerConditional) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerConditional) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
