package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
)

const (
	FEATURE_ID_TRIGGER_FOX_TALE_WILD = "TriggerFoxTaleWild"

	PARAM_ID_TRIGGER_FOX_TALE_WILD_RANDOM  = "Random"
	PARAM_ID_TRIGGER_FOX_TALE_WILD_TILE_ID = "TileId"
	PARAM_ID_TRIGGER_FOX_TALE_WILD_ENGINE  = "Engine"
	PARAM_ID_TRIGGER_FOX_TALE_WILD_LIMIT   = "Limit"

	PARAM_VALUE_TRIGGER_FOX_TALE_WILD_FREESPIN = "freespin"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_FOX_TALE_WILD, func() feature.Feature { return new(TriggerFoxTaleWild) })

type TriggerFoxTaleWild struct {
	feature.Base
}

func (f TriggerFoxTaleWild) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	random := params.GetInt(PARAM_ID_TRIGGER_FOX_TALE_WILD_RANDOM)
	tileid := params.GetInt(PARAM_ID_TRIGGER_FOX_TALE_WILD_TILE_ID)
	engine := params.GetString(PARAM_ID_TRIGGER_FOX_TALE_WILD_ENGINE)
	expand := random < params.GetInt(PARAM_ID_TRIGGER_FOX_TALE_WILD_LIMIT) || engine == PARAM_VALUE_TRIGGER_FOX_TALE_WILD_FREESPIN

	if expand {
		index := 0
		gridw := len(state.SymbolGrid)
		for x := 0; x < gridw; x++ {
			gridh := len(state.SymbolGrid[x])
			for y := 0; y < gridh; y++ {
				if state.SymbolGrid[x][y] == tileid {
					positions := []int{}
					for i := 0; i < gridh; i++ {
						positions = append(positions, index+i)
					}
					params[featureProducts.PARAM_ID_EXPANDING_WILD_POSITION] = index + y
					params[featureProducts.PARAM_ID_EXPANDING_WILD_POSITIONS] = positions
					feature.ActivateFeatures(f.FeatureDef, state, params)
					break
				}
			}
			index += gridh
		}

	}
	return
}

func (f *TriggerFoxTaleWild) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerFoxTaleWild) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
