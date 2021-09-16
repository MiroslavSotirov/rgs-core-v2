package features

import (
	"strings"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
)

type TriggerFoxTailWild struct {
	FeatureDef
}

func (f *TriggerFoxTailWild) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerFoxTailWild) DataPtr() interface{} {
	return nil
}

func (f *TriggerFoxTailWild) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f TriggerFoxTailWild) Trigger(state *FeatureState, params FeatureParams) {
	if config.GlobalConfig.DevMode && params.HasKey("force") && strings.Contains(params.GetString("force"), "actionsymbol") {
		f.ForceTrigger(state, params)
	}

	random := params.GetInt("Random")
	tileid := params.GetInt("TileId")
	engine := params.GetString("Engine")
	expand := random < params.GetInt("Limit") || engine == "freespin"

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
					params["Positions"] = positions
					activateFeatures(f.FeatureDef, state, params)
					break
				}
			}
			index += gridh
		}

	}
	return
}

func (f TriggerFoxTailWild) ForceTrigger(state *FeatureState, params FeatureParams) {
}

func (f *TriggerFoxTailWild) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerFoxTailWild) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
