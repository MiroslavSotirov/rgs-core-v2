package featureTriggers

import "gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"

const (
	FEATURE_ID_TRIGGER_SPIRIT_HUNTERS = "TriggerSpiritHunters"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_SPIRIT_HUNTERS, func() feature.Feature { return new(TriggerSpiritHunters) })

type TriggerSpiritHunters struct {
	feature.Base
}

func (f TriggerSpiritHunters) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	feature.ActivateFeatures(f.FeatureDef, state, params)
	return
}

func (f *TriggerSpiritHunters) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerSpiritHunters) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
