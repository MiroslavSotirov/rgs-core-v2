package features

import (
	"strings"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

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
	if params.HasKey("force") && strings.Contains(params.GetString("force"), "actionsymbol") {
		f.ForceTrigger(state, params)
		return
	}

	random := params.GetInt("Random")
	tileid := params.GetInt("TileId")
	replaceid := random % 9
	params["ReplaceWithId"] = replaceid
	gridw, gridh := len(state.SymbolGrid), len(state.SymbolGrid[0])

	positions := []int{}

	for x := 0; x < gridw; x++ {
		for y := 0; y < gridh; y++ {
			if state.SymbolGrid[x][y] == tileid {
				positions = append(positions, x*gridh+y)
			}
		}
	}
	if len(positions) > 0 {
		params["Positions"] = positions
		activateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f TriggerSupaCrewActionSymbol) ForceTrigger(state *FeatureState, params FeatureParams) {
	num := rng.RandFromRange(15) + 1
	tileid := params.GetInt("TileId")
	replaceid := rng.RandFromRange(9)
	params["ReplaceWithId"] = replaceid
	gridh := len(state.SymbolGrid[0])
	positions := make([]int, num)
	for i := 0; i < num; i++ {
		x := rng.RandFromRange(5)
		y := rng.RandFromRange(3)
		positions[i] = x*gridh + y
		state.SymbolGrid[x][y] = tileid
	}
	params["Positions"] = positions
	activateFeatures(f.FeatureDef, state, params)
}

func (f *TriggerSupaCrewActionSymbol) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerSupaCrewActionSymbol) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
