package featureTriggers

import "gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"

const (
	FEATURE_ID_TRIGGER_SWORD_KING = "TriggerSwordKing"

	PARAM_ID_TRIGGER_SWORD_KING_RUN_WILDS         = "RunWilds"
	PARAM_ID_TRIGGER_SWORD_KING_RUN_RESPIN        = "RunRespin"
	PARAM_ID_TRIGGER_SWORD_KING_RUN_BONUS_SCATTER = "RunBonusScatter"
	PARAM_ID_TRIGGER_SWORD_KING_RUN_BONUS         = "RunBonus"
	PARAM_ID_TRIGGER_SWORD_KING_RUN_FSSCATTER     = "RunFSScatter"

	STATEFUL_ID_TRIGGER_SWORD_KING_COUNTER = "counter"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_SWORD_KING, func() feature.Feature { return new(TriggerSwordKing) })

type TriggerSwordKing struct {
	feature.Base
}

func (f TriggerSwordKing) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	counter := 0
	statefulStake := feature.GetStatefulStakeMap(*state)
	if statefulStake.HasKey(STATEFUL_ID_TRIGGER_SWORD_KING_COUNTER) {
		counter = statefulStake.GetInt(STATEFUL_ID_TRIGGER_SWORD_KING_COUNTER)
	}

	feature.SetStatefulStakeMap(*state, feature.FeatureParams{STATEFUL_ID_TRIGGER_SWORD_KING_COUNTER: counter},
		params)

	feature.ActivateFeatures(f.FeatureDef, state, params)
	return
}

func (f *TriggerSwordKing) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerSwordKing) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
