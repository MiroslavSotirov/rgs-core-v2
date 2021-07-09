package features

import "gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"

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
	random := params.GetInt("Random")
	ran15 := rng.RandFromRange(15)
	if random/9 < 20 {
		h := []int{1, 2, 3 - 2, -1}[ran15%5]
		x := ran15 / 5
		y := 0
		if h < 0 {
			h = -h
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
