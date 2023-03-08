package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
)

const (
	FEATURE_ID_TRIGGER_TIPSY_CHARMS_BEER = "TriggerTipsyCharmsBeer"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_TIPSY_CHARMS_BEER,
	func() feature.Feature { return new(TriggerTipsyCharmsBeer) })

type TriggerTipsyCharmsBeer struct {
	feature.Base
}

func (f TriggerTipsyCharmsBeer) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	feature.ActivateFeatures(f.FeatureDef, state, params)
	return
}

func (f *TriggerTipsyCharmsBeer) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerTipsyCharmsBeer) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
