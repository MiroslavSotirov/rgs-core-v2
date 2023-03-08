package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

const (
	FEATURE_ID_TRIGGER_MIN_PAYOUT = "TriggerMinPayout"

	PARAM_ID_TRIGGER_MIN_PAYOUT_PAYOUT_FACTOR = "PayoutFactor"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_MIN_PAYOUT,
	func() feature.Feature { return new(TriggerMinPayout) })

type TriggerMinPayout struct {
	feature.Base
}

func (f TriggerMinPayout) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	payoutFactor := params.GetInt(PARAM_ID_TRIGGER_MIN_PAYOUT_PAYOUT_FACTOR)

	totalWin := 0
	replay := state.NextReplay
	for replay != nil {

		for _, w := range replay.Wins {
			totalWin += w.Multiplier
		}
		replay = replay.NextReplay
	}

	if totalWin >= payoutFactor {
		logger.Debugf("total win %d satisfies the min payout factor %d", totalWin, payoutFactor)
		feature.ActivateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f *TriggerMinPayout) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerMinPayout) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
