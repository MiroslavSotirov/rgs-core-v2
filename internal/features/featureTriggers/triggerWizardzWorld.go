package featureTriggers

import "gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"

const (
	FEATURE_ID_TRIGGER_WIZARDZ_WORLD = "TriggerWizardzWorld"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_WIZARDZ_WORLD, func() feature.Feature { return new(TriggerWizardzWorld) })

type TriggerWizardzWorld struct {
	feature.Base
}

func (f TriggerWizardzWorld) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	feature.ActivateFeatures(f.FeatureDef, state, params)
	return
}

func (f *TriggerWizardzWorld) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerWizardzWorld) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
