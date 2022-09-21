package features

import "gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"

type TriggerClashOfHeroesRandomWilds struct {
	FeatureDef
}

func (f *TriggerClashOfHeroesRandomWilds) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerClashOfHeroesRandomWilds) DataPtr() interface{} {
	return nil
}

func (f *TriggerClashOfHeroesRandomWilds) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *TriggerClashOfHeroesRandomWilds) OnInit(state *FeatureState) {
}

func (f TriggerClashOfHeroesRandomWilds) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}

	gridh := len(state.SourceGrid[0])
	wildId := params.GetInt("TileId")
	numWilds := params.GetIntSlice("NumWilds")[WeightedRandomIndex(params.GetIntSlice("NumProbabilities"))]
	numTries := params.GetInt("RetryFactor") * numWilds
	positions := []int{}
	for try := 0; len(positions) < numWilds && try < numTries+1; try++ {
		reelidx := WeightedRandomIndex(params.GetIntSlice("ReelProbabilities"))
		rowidx := rng.RandFromRange(3)
		pos := reelidx*gridh + rowidx
		if func() bool {
			if state.SourceGrid[reelidx][rowidx] == wildId {
				return false
			}
			for _, p := range positions {
				if p == pos {
					return false
				}
			}
			return true
		}() {
			positions = append(positions, pos)
		}
	}

	if len(positions) > 0 {
		params["Positions"] = positions
		activateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f TriggerClashOfHeroesRandomWilds) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *TriggerClashOfHeroesRandomWilds) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerClashOfHeroesRandomWilds) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
