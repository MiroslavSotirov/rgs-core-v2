package featureProducts

import "gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"

const (
	FEATURE_ID_SET_CONDITIONAL = "SetConditional"

	PARAM_ID_SET_CONDITIONAL_CONDITIONAL_FLAG  = "ConditionalFlag"
	PARAM_ID_SET_CONDITIONAL_CONDITIONAL_TYPE  = "ConditionalType"
	PARAM_ID_SET_CONDITIONAL_CONDITIONAL_VALUE = "ConditionalValue"

	PARAM_VALUE_SET_CONDITIONAL_BOOL = "bool"
	PARAM_VALUE_SET_CONDITIONAL_TRUE = "true"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_SET_CONDITIONAL, func() feature.Feature { return new(SetConditional) })

type SetConditional struct {
	feature.Base
}

func (f SetConditional) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	conditionalFlag := params.GetString(PARAM_ID_SET_CONDITIONAL_CONDITIONAL_FLAG)
	conditionalType := ""
	conditionalValue := ""
	if params.HasKey(PARAM_ID_SET_CONDITIONAL_CONDITIONAL_TYPE) {
		conditionalType = params.GetString(PARAM_ID_SET_CONDITIONAL_CONDITIONAL_TYPE)
	}
	if params.HasKey(PARAM_ID_SET_CONDITIONAL_CONDITIONAL_VALUE) {
		conditionalValue = params.GetString(PARAM_ID_SET_CONDITIONAL_CONDITIONAL_VALUE)
	}

	switch {
	case conditionalType == PARAM_VALUE_SET_CONDITIONAL_BOOL || conditionalType == "":
		params[conditionalFlag] = conditionalValue == PARAM_VALUE_SET_CONDITIONAL_TRUE || conditionalType == ""
	default:
		panic(FEATURE_ID_SET_CONDITIONAL + " unknown conditional type")
	}
	return
}

func (f *SetConditional) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *SetConditional) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
