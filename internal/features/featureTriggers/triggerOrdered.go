package featureTriggers

import (
	"fmt"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

const (
	FEATURE_ID_TRIGGER_ORDERED = "TriggerOrdered"

	PARAM_ID_TRIGGER_ORDERED_TRIGGERS = "Triggers"
	PARAM_ID_TRIGGER_ORDERED_ORDER    = "Order"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_ORDERED, func() feature.Feature { return new(TriggerOrdered) })

type TriggerOrdered struct {
	feature.Base
}

func (f TriggerOrdered) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	if params.HasKey(PARAM_ID_TRIGGER_ORDERED_ORDER) {
		triggers := params.GetStringSlice(PARAM_ID_TRIGGER_ORDERED_TRIGGERS)
		order := params.GetStringSlice(PARAM_ID_TRIGGER_ORDERED_ORDER)

		logger.Debugf("trigger in order: %v out of %v", order, triggers)

		for _, o := range order {

			idx := -1
			for i := range triggers {
				if triggers[i] == o {
					idx = i
					break
				}
			}
			if idx < 0 {
				panic(fmt.Sprintf("unknown ordered trigger: %s", o))
			} else {
				logger.Debugf("trigger ordered index %d", idx)
			}

			matchidx := func(i int, d feature.FeatureDef, s *feature.FeatureState, p feature.FeatureParams) bool {
				return i == idx
			}
			feature.ActivateFilteredFeatures(f.FeatureDef, state, params, matchidx)
		}
	}
}

func (f *TriggerOrdered) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerOrdered) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
