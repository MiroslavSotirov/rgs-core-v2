package features

type TriggerClashOfHeroesExpandingWilds struct {
	FeatureDef
}

func (f *TriggerClashOfHeroesExpandingWilds) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerClashOfHeroesExpandingWilds) DataPtr() interface{} {
	return nil
}

func (f *TriggerClashOfHeroesExpandingWilds) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *TriggerClashOfHeroesExpandingWilds) OnInit(state *FeatureState) {
}

func (f TriggerClashOfHeroesExpandingWilds) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}

	wildId := params.GetInt("TileId")
	patterns := params.GetSlice("Patterns")
	probabilities := params.GetSlice("Probabilities")

	_, gridh := len(state.SourceGrid), len(state.SourceGrid[0])

	for idx, pat := range patterns {
		positions := convertIntSlice(pat)
		match := func() bool {
			for _, pos := range positions {
				if state.SourceGrid[pos/gridh][pos%gridh] != wildId {
					return false
				}
			}
			return true
		}()
		if match {
			pos := positions[WeightedRandomIndex(convertIntSlice(probabilities[idx]))]
			positions := make([]int, 9)
			pidx := 0
			for x := -1; x <= 1; x++ {
				for y := -1; y <= 1; y, pidx = y+1, pidx+1 {
					positions[pidx] = pos + y + x*gridh
				}
			}
			params["Position"] = pos
			params["Positions"] = positions
			activateFeatures(f.FeatureDef, state, params)
			return
		}
	}

	return
}

func (f TriggerClashOfHeroesExpandingWilds) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *TriggerClashOfHeroesExpandingWilds) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerClashOfHeroesExpandingWilds) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
