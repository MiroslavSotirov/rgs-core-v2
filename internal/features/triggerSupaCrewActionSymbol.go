package features

type TriggerSupaCrewActionSymbol struct {
	FeatureDef
}

func (f *TriggerSupaCrewActionSymbol) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerSupaCrewActionSymbol) DataPtr() interface{} {
	return nil
}

func (f *TriggerSupaCrewActionSymbol) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f TriggerSupaCrewActionSymbol) Trigger(state *FeatureState, params FeatureParams) {
	random := params.GetInt("Random")
	tileid := params.GetInt("TileId")
	replaceid := random % 9
	params["ReplaceWithId"] = replaceid
	gridw, gridh := len(state.SymbolGrid), len(state.SymbolGrid[0])

	for x := 0; x < gridw; x++ {
		for y := 0; y < gridh; y++ {
			if state.SymbolGrid[x][y] == tileid {
				params["X"] = x
				params["Y"] = y
				activateFeatures(f.FeatureDef, state, params)
			}
		}
	}
	return
}

func (f *TriggerSupaCrewActionSymbol) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerSupaCrewActionSymbol) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
