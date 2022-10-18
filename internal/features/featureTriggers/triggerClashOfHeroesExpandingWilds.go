package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
)

const (
	FEATURE_ID_TRIGGER_CLASH_OF_HEROES_EXPANDING_WILDS = "TriggerClashOfHeroesExpandingWilds"

	PARAM_ID_TRIGGER_CLASH_OF_HEROES_EXPANDING_WILDS_TILE_ID       = "TileId"
	PARAM_ID_TRIGGER_CLASH_OF_HEROES_EXPANDING_WILDS_PATTERNS      = "Patterns"
	PARAM_ID_TRIGGER_CLASH_OF_HEROES_EXPANDING_WILDS_PROBABILITIES = "Probabilities"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_CLASH_OF_HEROES_EXPANDING_WILDS, func() feature.Feature { return new(TriggerClashOfHeroesExpandingWilds) })

type TriggerClashOfHeroesExpandingWilds struct {
	feature.Base
}

func (f TriggerClashOfHeroesExpandingWilds) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	wildId := params.GetInt(PARAM_ID_TRIGGER_CLASH_OF_HEROES_EXPANDING_WILDS_TILE_ID)
	patterns := params.GetSlice(PARAM_ID_TRIGGER_CLASH_OF_HEROES_EXPANDING_WILDS_PATTERNS)
	probabilities := params.GetSlice(PARAM_ID_TRIGGER_CLASH_OF_HEROES_EXPANDING_WILDS_PROBABILITIES)

	_, gridh := len(state.SourceGrid), len(state.SourceGrid[0])

	for idx, pat := range patterns {
		patternpositions := feature.ConvertIntSlice(pat)
		match := func() bool {
			for _, pos := range patternpositions {
				if state.SourceGrid[pos/gridh][pos%gridh] != wildId {
					return false
				}
			}
			return true
		}()
		if match {
			pos := patternpositions[feature.WeightedRandomIndex(feature.ConvertIntSlice(probabilities[idx]))]
			positions := make([]int, 9)
			pidx := 0
			for x := -1; x <= 1; x++ {
				for y := -1; y <= 1; y, pidx = y+1, pidx+1 {
					positions[pidx] = pos + y + x*gridh
				}
			}
			params[featureProducts.PARAM_ID_EXPANDING_WILD_POSITION] = pos
			params[featureProducts.PARAM_ID_EXPANDING_WILD_POSITIONS] = positions
			feature.ActivateFeatures(f.FeatureDef, state, params)
			return
		}
	}

	return
}

func (f *TriggerClashOfHeroesExpandingWilds) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerClashOfHeroesExpandingWilds) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
