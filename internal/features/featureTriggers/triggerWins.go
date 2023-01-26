package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

const (
	FEATURE_ID_TRIGGER_WINS = "TriggerWins"

	PARAM_ID_TRIGGER_WINS_CONDITION = "Condition"
	PARAM_ID_TRIGGER_WINS_PAYOUTS   = "Payouts"
	PARAM_VALUE_TRIGGER_WINS_ANY    = "any"
	PARAM_VALUE_TRIGGER_WINS_NONE   = "none"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_WINS, func() feature.Feature { return new(TriggerWins) })

type TriggerWins struct {
	feature.Base
}

func (f TriggerWins) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	var payouts []feature.FeaturePayout = nil
	if params.HasValue(PARAM_ID_TRIGGER_WINS_PAYOUTS) {
		payoutParams := params.GetParamsSlice(PARAM_ID_TRIGGER_WINS_PAYOUTS)
		payouts = make([]feature.FeaturePayout, len(payoutParams))
		for i, p := range payoutParams {
			payouts[i].Symbol = p.GetInt("Symbol")
			payouts[i].Count = p.GetInt("Count")
			payouts[i].Multiplier = p.GetInt("Multiplier")
		}
		logger.Debugf("TriggerWins. Payouts %#v", payouts)
	}

	condition := PARAM_VALUE_TRIGGER_WINS_ANY
	if params.HasKey(PARAM_ID_TRIGGER_WINS_CONDITION) {
		condition = params.GetString(PARAM_ID_TRIGGER_WINS_CONDITION)
	}

	wins := state.CalculateWins(state.SymbolGrid, payouts)
	switch condition {
	case PARAM_VALUE_TRIGGER_WINS_ANY:
		if len(wins) > 0 {
			logger.Debugf("Trigger on Wins (use engine payouts: %t)", len(payouts) > 0)
			feature.ActivateFeatures(f.FeatureDef, state, params)
		}
	case PARAM_VALUE_TRIGGER_WINS_NONE:
		if len(wins) == 0 {
			logger.Debugf("Trigger on no wins (use engine payouts: %t)", len(payouts) > 0)
			feature.ActivateFeatures(f.FeatureDef, state, params)
		}
	default:
		panic("unknown trigger wins condition")
	}
}

func (f *TriggerWins) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerWins) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
