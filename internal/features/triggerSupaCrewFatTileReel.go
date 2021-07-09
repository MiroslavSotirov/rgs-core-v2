package features

type TriggerSupaCrewFatTileReel struct {
	FeatureDef
}

func (f *TriggerSupaCrewFatTileReel) DefPtr() *FeatureDef {
	return &f.FeatureDef
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
	tilew, tileh := params.GetInt("W"), params.GetInt("H")
	tileid := params.GetInt("TileId")

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

				params["InstaWinType"] = "spinningcoin"
				params["InstaWinSourceId"] = f.FeatureDef.Id
				params["InstaWinAmount"] = 100

				features = append(features, activateFeatures(f.FeatureDef, state, params)...)
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
