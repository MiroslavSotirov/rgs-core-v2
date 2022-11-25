package featureProducts

import (
	"fmt"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

const (
	FEATURE_ID_RESPIN = "Respin"

	PARAM_ID_RESPIN_ACTION = "Action"
	PARAM_ID_RESPIN_AMOUNT = "Amount"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_RESPIN, func() feature.Feature { return new(Respin) })

type Respin struct {
	feature.Base
	Data feature.FeatureParams `json:"data"`
}

func (f *Respin) DataPtr() interface{} {
	return &f.Data
}

func (f Respin) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	if params.HasKey(PARAM_ID_RESPIN_ACTION) {
		action := params.GetString(PARAM_ID_RESPIN_ACTION)
		amount := 1
		if params.HasKey(PARAM_ID_RESPIN_AMOUNT) {
			amount = params.GetInt(PARAM_ID_RESPIN_AMOUNT)
		}
		state.Wins = append(state.Wins, feature.FeatureWin{
			Index: fmt.Sprintf("%s:%d", action, amount),
		})
		feature.ActivateFeatures(f.FeatureDef, state, params)
	} else {
		logger.Debugf("skip respin due to no action")
	}
}

func (f *Respin) Serialize() ([]byte, error) {
	return feature.SerializeFeatureToBytes(f)
}

func (f *Respin) Deserialize(data []byte) (err error) {
	return feature.DeserializeFeatureFromBytes(f, data)
}
