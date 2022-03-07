package features

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type TriggerWeightedPayout struct {
	FeatureDef
}

func (f *TriggerWeightedPayout) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerWeightedPayout) DataPtr() interface{} {
	return nil
}

func (f *TriggerWeightedPayout) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f TriggerWeightedPayout) Trigger(state *FeatureState, params FeatureParams) {
	var weights []int
	var payouts []int = params.GetIntSlice("Payouts")
	if params.HasKey("Weights") {
		weights = params.GetIntSlice("Weights")
	} else {
		weights = make([]int, len(payouts))
		for i := range weights {
			weights[i] = 1
		}
	}
	logger.Debugf("weights: %v payouts: %v", weights, payouts)
	idx := WeightedRandomIndex(weights)
	logger.Debugf("Weighted payout index: %d", idx)
	params["InstaWinType"] = "bonus"
	params["InstaWinSourceId"] = f.FeatureDef.Id
	params["InstaWinAmount"] = payouts[idx]
	params["Positions"] = []int{}

	activateFeatures(f.FeatureDef, state, params)
}

func (f TriggerWeightedPayout) ForceTrigger(state *FeatureState, params FeatureParams) {
}

func (f *TriggerWeightedPayout) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerWeightedPayout) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
