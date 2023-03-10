package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

const (
	FEATURE_ID_TRIGGER_TIPSY_CHARMS = "TriggerTipsyCharms"

	PARAM_ID_TRIGGER_TIPSY_CHARMS_WILD_IDS = "WildIds"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_TIPSY_CHARMS,
	func() feature.Feature { return new(TriggerTipsyCharms) })

type TriggerTipsyCharms struct {
	feature.Base
}

func (f TriggerTipsyCharms) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	wildIds := params.GetIntSlice(PARAM_ID_TRIGGER_TIPSY_CHARMS_WILD_IDS)

	if state.Stateless != nil {
		if len(state.SymbolGrid) > len(state.Stateless.SymbolGrid) {
			logger.Debugf("expanding number of reels from 4 to 6. Old symbol grid: %#v", state.Stateless.SymbolGrid)

			for i := 0; i < 4; i++ {
				for j, s := range state.Stateless.SymbolGrid[i] {
					state.SymbolGrid[i+2][j] = s
				}
			}
		} else {
			logger.Debugf("copying wilds from previous spin")

			for i, r := range state.Stateless.SymbolGrid {
				for j, s := range r {
					for _, w := range wildIds {
						if s == w {
							state.SymbolGrid[i][j] = w
							break
						}
					}
				}
			}
		}
	}

	feature.ActivateFeatures(f.FeatureDef, state, params)
	return
}

func (f *TriggerTipsyCharms) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerTipsyCharms) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
