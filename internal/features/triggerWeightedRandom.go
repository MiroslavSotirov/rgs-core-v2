package features

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

type TriggerWeightedRandom struct {
	FeatureDef
}

func (f *TriggerWeightedRandom) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerWeightedRandom) DataPtr() interface{} {
	return nil
}

func (f *TriggerWeightedRandom) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *TriggerWeightedRandom) OnInit(state *FeatureState) {
}

func (f TriggerWeightedRandom) Trigger(state *FeatureState, params FeatureParams) {
	var weights []int
	if params.HasKey("Weights") {
		weights = params.GetIntSlice("Weights")
	} else {
		num := len(f.Features)
		weights = make([]int, num)
		for i := range weights {
			weights[i] = 1
		}
	}
	idx := WeightedRandomIndex(weights)
	matchidx := func(i int, d FeatureDef, s *FeatureState, p FeatureParams) bool { return i == idx }
	activateFilteredFeatures(f.FeatureDef, state, params, matchidx)
}

func (f TriggerWeightedRandom) ForceTrigger(state *FeatureState, params FeatureParams) {
}

func (f *TriggerWeightedRandom) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerWeightedRandom) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}

func WeightedRandomIndex(weights []int) int {
	var sum, i, w int
	for _, w = range weights {
		sum += w
	}
	r := rng.RandFromRange(sum)
	sum = 0
	for i, w = range weights {
		sum += w
		if r < sum {
			break
		}
	}
	return i
}
