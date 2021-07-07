package features

type TriggerSupaCrewFatTileChance struct {
	FeatureDef
}

func (f *TriggerSupaCrewFatTileChance) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerSupaCrewFatTileChance) DataPtr() interface{} {
	return nil
}

func (f *TriggerSupaCrewFatTileChance) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f TriggerSupaCrewFatTileChance) Trigger(state FeatureState, params FeatureParams) []Feature {
	gridh := len(state.SymbolGrid[0])
	random := paramInt(params, "Random")
	if random/9 < 10 {
		ran15 := random % 15
		h := (ran15 % 3) + 1
		x := ran15 / 5
		y := 0
		bottom := (ran15/3)%2 > 0
		if bottom {
			y = gridh - h
		}
		params["W"] = 3
		params["H"] = h
		params["X"] = x
		params["Y"] = y

		return activateFeatures(f.FeatureDef, state, params)
	}
	return []Feature{}
}

func (f *TriggerSupaCrewFatTileChance) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerSupaCrewFatTileChance) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
