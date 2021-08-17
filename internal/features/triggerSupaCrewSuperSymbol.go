package features

import "gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"

type TriggerSupaCrewSuperSymbol struct {
	FeatureDef
}

func (f *TriggerSupaCrewSuperSymbol) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerSupaCrewSuperSymbol) DataPtr() interface{} {
	return nil
}

func (f *TriggerSupaCrewSuperSymbol) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f TriggerSupaCrewSuperSymbol) Trigger(state *FeatureState, params FeatureParams) {
	gridh := len(state.SymbolGrid[0])
	random := params.GetInt("Random")
	ran15 := rng.RandFromRange(15)
	if random/9 < 20 {
		h := []int{1, 2, 3, -2, -1}[ran15%5]
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
		params["TileId"] = random % 9

		activateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f *TriggerSupaCrewSuperSymbol) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerSupaCrewSuperSymbol) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
