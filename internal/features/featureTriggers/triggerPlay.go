package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
)

const (
	FEATURE_ID_TRIGGER_PLAY = "TriggerPlay"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_PLAY,
	func() feature.Feature { return new(TriggerPlay) })

type TriggerPlay struct {
	feature.Base
}

func (f TriggerPlay) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	if state.NextReplay == nil {
		feature.ActivateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f *TriggerPlay) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerPlay) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
