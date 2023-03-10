package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
)

const (
	FEATURE_ID_TRIGGER_REPLAY = "TriggerReplay"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_REPLAY,
	func() feature.Feature { return new(TriggerReplay) })

type TriggerReplay struct {
	feature.Base
}

func (f TriggerReplay) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	if state.NextReplay != nil {
		feature.ActivateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f *TriggerReplay) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerReplay) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
