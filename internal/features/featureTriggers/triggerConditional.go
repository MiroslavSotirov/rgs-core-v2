package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

const (
	FEATURE_ID_TRIGGER_CONDITIONAL = "TriggerConditional"

	PARAM_ID_TRIGGER_CONDITIONAL_CONDITIONAL_FLAG = "ConditionalFlag"
	PARAM_ID_TRIGGER_CONDITIONAL_CONDITION        = "Condition"

	PARAM_VALUE_TRIGGER_CONDITIONAL_SET   = "Set"
	PARAM_VALUE_TRIGGER_CONDITIONAL_UNSET = "Unset"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_CONDITIONAL, func() feature.Feature { return new(TriggerConditional) })

type TriggerConditional struct {
	feature.Base
}

func (f TriggerConditional) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	conditionalFlag := params.GetString(PARAM_ID_TRIGGER_CONDITIONAL_CONDITIONAL_FLAG)
	condition := ""
	if params.HasKey(PARAM_ID_TRIGGER_CONDITIONAL_CONDITION) {
		condition = params.GetString(PARAM_ID_TRIGGER_CONDITIONAL_CONDITION)
	}

	activate := false
	switch {
	case condition == PARAM_VALUE_TRIGGER_CONDITIONAL_SET || condition == "":
		activate = params.HasKey(conditionalFlag)
	case condition == PARAM_VALUE_TRIGGER_CONDITIONAL_UNSET:
		activate = !params.HasKey(conditionalFlag)
	}
	if activate {
		logger.Debugf("%s is %s", conditionalFlag, condition)
		feature.ActivateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f *TriggerConditional) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerConditional) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
