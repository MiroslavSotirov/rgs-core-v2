package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

const (
	FEATURE_ID_TRIGGER_BATTLE_OF_MYTHS_TIGER = "TriggerBattleOfMythsTiger"

	PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_TIGER_SIZES              = "Sizes"
	PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_TIGER_SIZE_PROBABILITIES = "SizeProbabilities"
	PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_TIGER_REEL_PROBABILITIES = "ReelProbabilities"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_BATTLE_OF_MYTHS_TIGER, func() feature.Feature { return new(TriggerBattleOfMythsTiger) })

type TriggerBattleOfMythsTiger struct {
	feature.Base
}

func (f TriggerBattleOfMythsTiger) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	if state.Action == "cascade" {
		return
	}

	sizes := params.GetIntSlice(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_TIGER_SIZES)
	sizeProbs := params.GetIntSlice(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_TIGER_SIZE_PROBABILITIES)
	reelProbs := params.GetSlice(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_TIGER_REEL_PROBABILITIES)

	sizeidx := feature.WeightedRandomIndex(sizeProbs)
	reelidx := feature.WeightedRandomIndex(feature.ConvertIntSlice(reelProbs[sizeidx]))
	rowidx := 0
	if sizes[sizeidx] < len(state.SymbolGrid[0]) {
		rowidx = rng.RandFromRangePool(len(state.SymbolGrid[0]) - sizes[sizeidx])
	}

	params[featureProducts.PARAM_ID_FAT_TILE_W] = sizes[sizeidx]
	params[featureProducts.PARAM_ID_FAT_TILE_H] = sizes[sizeidx]
	params[featureProducts.PARAM_ID_FAT_TILE_X] = reelidx
	params[featureProducts.PARAM_ID_FAT_TILE_Y] = rowidx

	feature.ActivateFeatures(f.FeatureDef, state, params)
	return
}

func (f *TriggerBattleOfMythsTiger) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerBattleOfMythsTiger) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
