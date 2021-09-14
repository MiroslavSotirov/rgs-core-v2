package features

import (
	"strings"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

type TriggerFoxTailBonus struct {
	FeatureDef
}

func (f *TriggerFoxTailBonus) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerFoxTailBonus) DataPtr() interface{} {
	return nil
}

func (f *TriggerFoxTailBonus) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f TriggerFoxTailBonus) Trigger(state *FeatureState, params FeatureParams) {
	if config.GlobalConfig.DevMode && params.HasKey("force") && strings.Contains(params.GetString("force"), "actionsymbol") {
		f.ForceTrigger(state, params)
	}

	tileid := params.GetInt("TileId")
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
		ran8 := rng.RandFromRange(8)
		params["InstaWinType"] = "bonus"
		params["InstaWinSourceId"] = f.FeatureDef.Id
		params["InstaWinAmount"] = []int{
			15, 18, 21, 24, 27, 30, 35, 40,
		}[ran8]
		params["Positions"] = positions
		activateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f TriggerFoxTailBonus) ForceTrigger(state *FeatureState, params FeatureParams) {
}

func (f *TriggerFoxTailBonus) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerFoxTailBonus) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
