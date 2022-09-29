package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

const (
	FEATURE_ID_TRIGGER_CLASH_OF_HEROES_RANDOM_WILDS = "TriggerClashOfHeroesRandomWilds"

	PARAM_ID_TRIGGER_CLASH_OF_HEROES_RANDOM_WILDS_TILE_ID            = "TileId"
	PARAM_ID_TRIGGER_CLASH_OF_HEROES_RANDOM_WILDS_NUM_WILDS          = "NumWilds"
	PARAM_ID_TRIGGER_CLASH_OF_HEROES_RANDOM_WILDS_NUM_PROBABILITIES  = "NumProbabilities"
	PARAM_ID_TRIGGER_CLASH_OF_HEROES_RANDOM_WILDS_RETRY_FACTOR       = "RetryFactor"
	PARAM_ID_TRIGGER_CLASH_OF_HEROES_RANDOM_WILDS_REEL_PROBABILITIES = "ReelProbabilities"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_CLASH_OF_HEROES_RANDOM_WILDS, func() feature.Feature { return new(TriggerClashOfHeroesRandomWilds) })

type TriggerClashOfHeroesRandomWilds struct {
	feature.Base
}

func (f TriggerClashOfHeroesRandomWilds) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	gridh := len(state.SymbolGrid[0])
	wildId := params.GetInt(PARAM_ID_TRIGGER_CLASH_OF_HEROES_RANDOM_WILDS_TILE_ID)
	numWilds := params.GetIntSlice(PARAM_ID_TRIGGER_CLASH_OF_HEROES_RANDOM_WILDS_NUM_WILDS)[feature.WeightedRandomIndex(
		params.GetIntSlice(PARAM_ID_TRIGGER_CLASH_OF_HEROES_RANDOM_WILDS_NUM_PROBABILITIES))]
	numTries := params.GetInt(PARAM_ID_TRIGGER_CLASH_OF_HEROES_RANDOM_WILDS_RETRY_FACTOR) * numWilds
	positions := []int{}
	for try := 0; len(positions) < numWilds && try < numTries+1; try++ {
		reelidx := feature.WeightedRandomIndex(
			params.GetIntSlice(PARAM_ID_TRIGGER_CLASH_OF_HEROES_RANDOM_WILDS_REEL_PROBABILITIES))
		rowidx := rng.RandFromRange(3)
		pos := reelidx*gridh + rowidx
		if func() bool {
			if state.SymbolGrid[reelidx][rowidx] == wildId {
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
		params[featureProducts.PARAM_ID_REPLACE_TILE_POSITIONS] = positions
		feature.ActivateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f *TriggerClashOfHeroesRandomWilds) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerClashOfHeroesRandomWilds) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
