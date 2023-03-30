package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

const (
	FEATURE_ID_TRIGGER_SPIRIT_HUNTERS_BONUS = "TriggerSpiritHuntersBonus"

	PARAM_ID_TRIGGER_SPIRIT_HUNTERS_BONUS_TILE_ID = "TileId"
	PARAM_ID_TRIGGER_SPIRIT_HUNTERS_BONUS_PRIZES  = "Prizes"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_SPIRIT_HUNTERS_BONUS, func() feature.Feature { return new(TriggerSpiritHuntersBonus) })

type TriggerSpiritHuntersBonus struct {
	feature.Base
}

func (f TriggerSpiritHuntersBonus) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	tileid := params.GetInt(PARAM_ID_TRIGGER_SPIRIT_HUNTERS_BONUS_TILE_ID)
	gridw := len(state.SymbolGrid)
	index := 0
	positions := []int{}
	for x := 0; x < gridw; x++ {
		gridh := len(state.SymbolGrid[x])
		for y := 0; y < gridh; y++ {
			if state.SymbolGrid[x][y] == tileid {
				positions = append(positions, index+y)
			}
		}
		index += gridh
	}
	if len(positions) >= 3 {
		prizes := params.GetIntSlice(PARAM_ID_TRIGGER_SPIRIT_HUNTERS_BONUS_PRIZES)
		ran := rng.RandFromRangePool(len(prizes))
		params[featureProducts.PARAM_ID_INSTA_WIN_TYPE] = featureProducts.PARAM_VALUE_INSTA_WIN_BONUS
		params[featureProducts.PARAM_ID_INSTA_WIN_SOURCE_ID] = f.FeatureDef.Id
		params[featureProducts.PARAM_ID_INSTA_WIN_AMOUNT] = prizes[ran]
		params[featureProducts.PARAM_ID_INSTA_WIN_POSITIONS] = positions
		feature.ActivateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f *TriggerSpiritHuntersBonus) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerSpiritHuntersBonus) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
