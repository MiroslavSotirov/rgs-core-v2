package features

import (
	"strings"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

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
	if params.HasKey("force") && strings.Contains(params.GetString("force"), "supersymbol") {
		f.ForceTrigger(state, params)
		return
	}

	random := params.GetInt("Random")
	ran15 := rng.RandFromRange(15)
	//	if random/9 < 20 {
	// test version
	ran9 := random / 9
	if ran9 >= 30 && ran9 <= 39 {
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

func (f TriggerSupaCrewSuperSymbol) ForceTrigger(state *FeatureState, params FeatureParams) {
	params["X"] = rng.RandFromRange(5)
	params["Y"] = rng.RandFromRange(5) - 2
	params["W"] = 3
	params["H"] = 3
	params["TileId"] = rng.RandFromRange(9)

	activateFeatures(f.FeatureDef, state, params)
}

func (f *TriggerSupaCrewSuperSymbol) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerSupaCrewSuperSymbol) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
