package features

import "gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"

type TriggerFatTile struct {
	FeatureDef
}

func (f *TriggerFatTile) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerFatTile) DataPtr() interface{} {
	return nil
}

func (f *TriggerFatTile) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *TriggerFatTile) OnInit(state *FeatureState) {
}

func (f TriggerFatTile) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}
	tileid := params.GetInt("TileId")
	height := params.GetInt("H")
	logger.Debugf("TriggerFatTile tileId: %d h: %d", tileid, height)
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
			logger.Debugf("  found at x: %d y: %d c: %d", x, yf, c)
			params["X"] = x
			if yf > 0 {
				params["Y"] = yf
			} else {
				params["Y"] = c - height
			}
			activateFeatures(f.FeatureDef, state, params)
		}

	}
	return
}

func (f TriggerFatTile) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *TriggerFatTile) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerFatTile) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
