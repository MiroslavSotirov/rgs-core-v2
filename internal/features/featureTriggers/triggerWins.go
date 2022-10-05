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
	var payouts []feature.FeaturePayout = nil
	if params.HasValue("Payouts") {
		payoutParams := params.GetParamsSlice("Payouts")
		payouts = make([]feature.FeaturePayout, len(payoutParams))
		for i, p := range payoutParams {
			payouts[i].Symbol = p.GetInt("Symbol")
			payouts[i].Count = p.GetInt("Count")
			payouts[i].Multiplier = p.GetInt("Multiplier")
		}
		logger.Debugf("TriggerWins. Payouts %#v", payouts)
	}

	if len(state.CalculateWins(state.SymbolGrid, payouts)) > 0 {
		logger.Debugf("Trigger on Wins (use engine payouts: %t)", len(payouts) > 0)
		feature.ActivateFeatures(f.FeatureDef, state, params)
	}
}

func (f *TriggerWins) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerWins) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
