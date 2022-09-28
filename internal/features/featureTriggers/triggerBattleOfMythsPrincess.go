package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
)

const (
	FEATURE_ID_TRIGGER_BATTLE_OF_MYTHS_PRINCESS = "TriggerBattleOfMythsPrincess"

	PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_PRINCESS_TILE_ID = "TileId"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_BATTLE_OF_MYTHS_PRINCESS, func() feature.Feature { return new(TriggerBattleOfMythsPrincess) })

type TriggerBattleOfMythsPrincess struct {
	feature.Base
}

func (f TriggerBattleOfMythsPrincess) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	tileId := params.GetInt(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_PRINCESS_TILE_ID)

	positions := []int{}
	gridh := len(state.SymbolGrid[0])
	for x, r := range state.SymbolGrid {
		fy := -1
		for y, s := range r {
			if fy < 0 {
				if s == tileId {
					fy = y
				}
			} else {
				if s == tileId {
					for ry := fy + 1; ry < y; ry++ {
						positions = append(positions, x*gridh+ry)
					}
					more := false
					for yc := y + 1; !more && yc < len(r); yc++ {
						if r[yc] == tileId {
							more = true
						}
					}
					if len(positions) > 0 {
						params[featureProducts.PARAM_ID_PRINCESS_POSIIONS] = positions
						params[featureProducts.PARAM_ID_PRINCESS_START_POS] = x*gridh + fy
						params[featureProducts.PARAM_ID_PRINCESS_END_POS] = x*gridh + y
						feature.ActivateFeatures(f.FeatureDef, state, params)
					}
					if more {
						fy = y
					} else {
						break
					}
				}
			}
		}
	}
	gridw := len(state.SymbolGrid)
	for y := 0; y < gridh; y++ {
		fx := -1
		for x := 0; x < gridw; x++ {
			s := state.SymbolGrid[x][y]
			if fx < 0 {
				if s == tileId {
					fx = x
				}
			} else {
				if s == tileId {
					for rx := fx + 1; rx < x; rx++ {
						positions = append(positions, rx*gridh+y)
					}
					more := false
					for xc := x + 1; !more && xc < gridw; xc++ {
						if state.SymbolGrid[xc][y] == tileId {
							more = true
						}
					}
					if len(positions) > 0 {
						params[featureProducts.PARAM_ID_PRINCESS_POSIIONS] = positions
						params[featureProducts.PARAM_ID_PRINCESS_START_POS] = fx*gridh + y
						params[featureProducts.PARAM_ID_PRINCESS_END_POS] = x*gridh + y
						feature.ActivateFeatures(f.FeatureDef, state, params)
					}
					if more {
						fx = x
					} else {
						break
					}
				}
			}
		}
	}
	return
}

func (f *TriggerBattleOfMythsPrincess) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerBattleOfMythsPrincess) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
