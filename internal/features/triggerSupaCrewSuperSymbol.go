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
	random := params.GetInt("Random")
	ran15 := rng.RandFromRange(15)
	if random/9 < 20 {
		x := ran15 / 5
		y := []int{-2, -1, 0, 1, 2}[ran15%5]
		params["W"] = 3
		params["H"] = 3
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
