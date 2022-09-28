package featureTriggers

import "gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"

const (
	FEATURE_ID_TRIGGER_CLASH_OF_HEROES = "TriggerClashOfHeroes"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_CLASH_OF_HEROES, func() feature.Feature { return new(TriggerClashOfHeroes) })

type TriggerClashOfHeroes struct {
	feature.Base
}

func (f TriggerClashOfHeroes) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	feature.ActivateFeatures(f.FeatureDef, state, params)
	return
}

func (f *TriggerClashOfHeroes) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerClashOfHeroes) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
