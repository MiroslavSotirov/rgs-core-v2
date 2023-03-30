package featureTriggers

import (
	"fmt"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

const (
	FEATURE_ID_TRIGGER_SWORD_KING_RESPIN = "TriggerSwordKingRespin"

	PARAM_ID_TRIGGER_SWORD_KING_RESPIN_PROBABILITY   = "Probability"
	PARAM_ID_TRIGGER_SWORD_KING_RESPIN_FSTYPE        = "FSType"
	PARAM_ID_TRIGGER_SWORD_KING_RESPIN_NUM_FREESPINS = "NumFreespins"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_SWORD_KING_RESPIN, func() feature.Feature { return new(TriggerSwordKingRespin) })

type TriggerSwordKingRespin struct {
	feature.Base
}

func (f TriggerSwordKingRespin) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	if params.HasKey("PureWins") {
		// logger.Debugf("skipping respin due to wins")
		return
	} else if params.HasKey(PARAM_ID_TRIGGER_SWORD_KING_RUN_WILDS) && params.GetBool(PARAM_ID_TRIGGER_SWORD_KING_RUN_WILDS) {
		// logger.Debugf("skipping respin due to random wilds")
		return
	}

	Probability := params.GetInt(PARAM_ID_TRIGGER_SWORD_KING_RESPIN_PROBABILITY)
	if rng.RandFromRangePool(10000) < Probability {

		fstype := params.GetString(PARAM_ID_TRIGGER_SWORD_KING_RESPIN_FSTYPE)
		numFreespins := params.GetInt(PARAM_ID_TRIGGER_SWORD_KING_RESPIN_NUM_FREESPINS)

		// logger.Debugf("Respin trigger %d freespins of type %s", numFreespins, fstype)

		state.Wins = append(state.Wins, feature.FeatureWin{
			Index: fmt.Sprintf("%s:%d", fstype, numFreespins),
		})

		params[PARAM_ID_TRIGGER_SWORD_KING_RUN_RESPIN] = true
		feature.ActivateFeatures(f.FeatureDef, state, params)
	}

	return
}

func (f *TriggerSwordKingRespin) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerSwordKingRespin) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
