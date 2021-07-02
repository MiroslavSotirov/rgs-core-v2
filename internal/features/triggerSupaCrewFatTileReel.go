package features

type TriggerSupaCrewFatTileReel struct {
	Def FeatureDef `json:"def"`
}

func (f *TriggerSupaCrewFatTileReel) DefPtr() *FeatureDef {
	return &f.Def
}

func (f *TriggerSupaCrewFatTileReel) DataPtr() interface{} {
	return nil
}

func (f *TriggerSupaCrewFatTileReel) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f TriggerSupaCrewFatTileReel) Trigger(state FeatureState, params FeatureParams) []Feature {
	features := []Feature{}
	gridw, gridh := len(state.SymbolGrid), len(state.SymbolGrid[0])
	tilew, tileh := params["W"].(int), params["H"].(int)
	tileid := params["TileId"].(int)

	for x := 0; x < gridw-tilew+1; x++ {
		for y := 0; y < gridh-tileh+1; y++ {
			found := func() bool {
				for r := x; r < x+tilew; r++ {
					for s := y; s < y+tileh; s++ {
						if state.SymbolGrid[r][s] != tileid {
							return false
						}
					}
				}
				return true
			}()
			if found {
				params["X"] = x
				params["Y"] = y
				features = append(features, activateFeatures(f.Def, state, params)...)
			}
		}
	}
	return features
}

func (f *TriggerSupaCrewFatTileReel) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerSupaCrewFatTileReel) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
