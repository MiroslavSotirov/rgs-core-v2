package features

import (
	"strings"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

type TriggerSupaCrewMultiSymbol struct {
	FeatureDef
}

func (f *TriggerSupaCrewMultiSymbol) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerSupaCrewMultiSymbol) DataPtr() interface{} {
	return nil
}

func (f *TriggerSupaCrewMultiSymbol) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f TriggerSupaCrewMultiSymbol) Trigger(state *FeatureState, params FeatureParams) {
	if params.HasKey("force") && strings.Contains(params.GetString("force"), "multisymbol") {
		f.ForceTrigger(state, params)
		return
	}

	random := params.GetInt("Random")
	randiv := random / 9
	if randiv >= 20 && randiv <= 24 {
		ran8 := rng.RandFromRange(8)
		y := ran8 / 4
		x := ran8 % 4

		params["X"] = x
		params["Y"] = y

		ran12 := rng.RandFromRange(12)
		params["InstaWinType"] = "spinningcoin"
		params["InstaWinSourceId"] = f.FeatureDef.Id
		params["InstaWinAmount"] = []int{
			7, 8, 10, 12, 14, 16, 18, 20, 22, 25, 28, 30,
		}[ran12]

		activateFeatures(f.FeatureDef, state, params)
	}
}

func (f TriggerSupaCrewMultiSymbol) ForceTrigger(state *FeatureState, params FeatureParams) {
	params["X"] = rng.RandFromRange(4)
	params["Y"] = rng.RandFromRange(2)
	ran12 := rng.RandFromRange(12)
	params["InstaWinType"] = "spinningcoin"
	params["InstaWinSourceId"] = f.FeatureDef.Id
	params["InstaWinAmount"] = []int{
		7, 8, 10, 12, 14, 16, 18, 20, 22, 25, 28, 30,
	}[ran12]

	activateFeatures(f.FeatureDef, state, params)
}

func (f *TriggerSupaCrewMultiSymbol) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerSupaCrewMultiSymbol) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
