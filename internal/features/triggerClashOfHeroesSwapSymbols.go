package features

import "gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"

type TriggerClashOfHeroesSwapSymbols struct {
	FeatureDef
}

func (f *TriggerClashOfHeroesSwapSymbols) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerClashOfHeroesSwapSymbols) DataPtr() interface{} {
	return nil
}

func (f *TriggerClashOfHeroesSwapSymbols) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *TriggerClashOfHeroesSwapSymbols) OnInit(state *FeatureState) {
}

func (f TriggerClashOfHeroesSwapSymbols) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}

	replaceIds := params.GetIntSlice("ReplaceIds")
	replaceWithIds := params.GetIntSlice("ReplaceWithIds")

	seniors := []int{}
	for _, r := range state.SymbolGrid {
		for _, s := range r {
			if func() bool {
				for _, rs := range replaceWithIds {
					if rs == s {
						return true
					}
				}
				return false
			}() {
				seniors = append(seniors, s)
			}
		}
	}

	if len(seniors) > 0 {

		gridh := len(state.SymbolGrid[0])
		positions := []int{}

		for reel, r := range state.SymbolGrid {
			for symbol, s := range r {
				if func() bool {
					for _, rs := range replaceIds {
						if rs == s {
							return true
						}
					}
					return false
				}() {
					positions = append(positions, reel*gridh+symbol)
				}
			}
		}

		if len(positions) > 0 {
			params["Positions"] = positions
			params["ReplaceWithId"] = seniors[rng.RandFromRange(len(seniors))]
			params["TileId"] = 10
			activateFeatures(f.FeatureDef, state, params)
			return
		}

	}
	return
}

func (f TriggerClashOfHeroesSwapSymbols) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *TriggerClashOfHeroesSwapSymbols) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerClashOfHeroesSwapSymbols) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
