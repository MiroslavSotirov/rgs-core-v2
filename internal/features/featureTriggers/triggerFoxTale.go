package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

const (
	FEATURE_ID_TRIGGER_FOX_TALE = "TriggerFoxTale"

	PARAM_ID_TRIGGER_FOX_TALE_RANDOM_RANGE = "RandomRange"
	PARAM_ID_TRIGGER_FOX_TALE_RANDOM       = "Random"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_FOX_TALE, func() feature.Feature { return new(TriggerFoxTale) })

type TriggerFoxTale struct {
	feature.Base
}

func (f TriggerFoxTale) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	params[PARAM_ID_TRIGGER_FOX_TALE_RANDOM] = rng.RandFromRange(f.FeatureDef.Params[PARAM_ID_TRIGGER_FOX_TALE_RANDOM_RANGE].(int))
	feature.ActivateFeatures(f.FeatureDef, state, params)
	return
}

func (f *TriggerFoxTale) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerFoxTale) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
