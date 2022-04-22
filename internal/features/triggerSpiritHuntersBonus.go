package features

import (
	"strings"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type TriggerSpiritHuntersBonus struct {
	FeatureDef
}

func (f *TriggerSpiritHuntersBonus) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerSpiritHuntersBonus) DataPtr() interface{} {
	return nil
}

func (f *TriggerSpiritHuntersBonus) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *TriggerSpiritHuntersBonus) OnInit(state *FeatureState) {
}

func (f TriggerSpiritHuntersBonus) Trigger(state *FeatureState, params FeatureParams) {
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
		logger.Debugf("params= %#v", params)
		prizes := params.GetIntSlice("Prizes")
		ran := rng.RandFromRange(len(prizes))
		params["InstaWinType"] = "bonus"
		params["InstaWinSourceId"] = f.FeatureDef.Id
		params["InstaWinAmount"] = prizes[ran]
		params["Positions"] = positions
		activateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f TriggerSpiritHuntersBonus) ForceTrigger(state *FeatureState, params FeatureParams) {
}

func (f *TriggerSpiritHuntersBonus) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerSpiritHuntersBonus) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
