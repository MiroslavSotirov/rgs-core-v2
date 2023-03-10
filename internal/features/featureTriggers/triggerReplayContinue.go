package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
)

const (
	FEATURE_ID_TRIGGER_REPLAY_CONTINUE = "TriggerReplayContinue"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_REPLAY_CONTINUE,
	func() feature.Feature { return new(TriggerReplayContinue) })

type TriggerReplayContinue struct {
	feature.Base
}

func (f TriggerReplayContinue) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	if state.NextReplay != nil && state.Replay {
		feature.ActivateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f *TriggerReplayContinue) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerReplayContinue) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
