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

	juniors := []int{}
	for _, r := range state.SymbolGrid {
		for _, s := range r {
			if containsInt(replaceIds, s) && !containsInt(juniors, s) {
				juniors = append(juniors, s)
			}
		}
	}

	if len(juniors) > 0 {

		junior := juniors[rng.RandFromRange(len(juniors))]

		seniors := []int{}
		for _, r := range state.SymbolGrid {
			for _, s := range r {
				if containsInt(replaceWithIds, s) && !containsInt(seniors, s) {
					seniors = append(seniors, s)
				}
			}
		}

		if len(seniors) > 0 {

			gridh := len(state.SymbolGrid[0])
			positions := []int{}

			for reel, r := range state.SymbolGrid {
				for symbol, s := range r {
					if s == junior {
						positions = append(positions, reel*gridh+symbol)
					}
				}
			}

			senior := seniors[rng.RandFromRange(len(seniors))]
			params[featureProducts.PARAM_ID_REPLACE_TILE_POSITIONS] = positions
			params[featureProducts.PARAM_ID_REPLACE_TILE_REPLACE_WITH_ID] = senior
			params[featureProducts.PARAM_ID_REPLACE_TILE_TILE_ID] = junior
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
