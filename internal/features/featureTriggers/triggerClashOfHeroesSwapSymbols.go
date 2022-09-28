package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

const (
	FEATURE_ID_TRIGGER_CLASH_OF_HEROES_SWAP_SYMBOLS = "TriggerClashOfHeroesSwapSymbols"

	PARAM_ID_TRIGGER_CLASH_OF_HEROES_SWAP_SYMBOLS_REPLACE_IDS      = "ReplaceIds"
	PARAM_ID_TRIGGER_CLASH_OF_HEROES_SWAP_SYMBOLS_REPLACE_WITH_IDS = "ReplaceWithIds"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_CLASH_OF_HEROES_SWAP_SYMBOLS, func() feature.Feature { return new(TriggerClashOfHeroesSwapSymbols) })

type TriggerClashOfHeroesSwapSymbols struct {
	feature.Base
}

func (f TriggerClashOfHeroesSwapSymbols) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	replaceIds := params.GetIntSlice(PARAM_ID_TRIGGER_CLASH_OF_HEROES_SWAP_SYMBOLS_REPLACE_IDS)
	replaceWithIds := params.GetIntSlice(PARAM_ID_TRIGGER_CLASH_OF_HEROES_SWAP_SYMBOLS_REPLACE_WITH_IDS)

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
			params[featureProducts.PARAM_ID_REPLACE_TILE_POSITIONS] = positions
			params[featureProducts.PARAM_ID_REPLACE_TILE_REPLACE_WITH_ID] = seniors[rng.RandFromRange(len(seniors))]
			params[featureProducts.PARAM_ID_REPLACE_TILE_TILE_ID] = 10
			feature.ActivateFeatures(f.FeatureDef, state, params)
			return
		}

	}
	return
}

func (f *TriggerClashOfHeroesSwapSymbols) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerClashOfHeroesSwapSymbols) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
