package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
)

const (
	FEATURE_ID_TRIGGER_FAT_TILE = "TriggerFatTile"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_FAT_TILE, func() feature.Feature { return new(TriggerFatTile) })

type TriggerFatTile struct {
	feature.Base
}

func (f TriggerFatTile) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	tileid := params.GetInt(featureProducts.PARAM_ID_FAT_TILE_TILE_ID)
	height := params.GetInt(featureProducts.PARAM_ID_FAT_TILE_H)
	for x, r := range state.SymbolGrid {
		yf, c := 0, 0
		for y, s := range r {
			if s == tileid {
				if c == 0 {
					yf = y
				}
				c++
			}
		}
		if c > 0 {
			params[featureProducts.PARAM_ID_FAT_TILE_X] = x
			if yf > 0 {
				params[featureProducts.PARAM_ID_FAT_TILE_Y] = yf
			} else {
				params[featureProducts.PARAM_ID_FAT_TILE_Y] = c - height
			}
			feature.ActivateFeatures(f.FeatureDef, state, params)
		}

	}
	return
}

func (f *TriggerFatTile) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerFatTile) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
