package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

const (
	FEATURE_ID_TRIGGER_WINS = "TriggerWins"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_WINS, func() feature.Feature { return new(TriggerWins) })

type TriggerWins struct {
	feature.Base
}

func (f TriggerWins) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	if state.AnyWins(state.SymbolGrid) {
		logger.Debugf("Trigger on Wins")
		feature.ActivateFeatures(f.FeatureDef, state, params)
	}
}

func (f *TriggerWins) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerWins) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
