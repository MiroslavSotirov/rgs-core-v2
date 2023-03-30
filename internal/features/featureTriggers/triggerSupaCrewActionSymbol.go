package featureTriggers

import (
	"strings"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

const (
	FEATURE_ID_TRIGGER_SUPA_CREW_ACTION_SYMBOL = "TriggerSupaCrewActionSymbol"

	PARAM_ID_TRIGGER_SUPA_CREW_ACTION_SYMBOL_FORCE   = "force"
	PARAM_ID_TRIGGER_SUPA_CREW_ACTION_SYMBOL_RANDOM  = "Random"
	PARAM_ID_TRIGGER_SUPA_CREW_ACTION_SYMBOL_TILE_ID = "TileId"

	PARAM_VALUE_TRIGGER_SUPA_CREW_ACTION_SYMBOL_ACTION_SYMBOL = "actionsymbol"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_SUPA_CREW_ACTION_SYMBOL, func() feature.Feature { return new(TriggerSupaCrewActionSymbol) })

type TriggerSupaCrewActionSymbol struct {
	feature.Base
}

func (f TriggerSupaCrewActionSymbol) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	if config.GlobalConfig.DevMode && params.HasKey(PARAM_ID_TRIGGER_SUPA_CREW_ACTION_SYMBOL_FORCE) &&
		strings.Contains(params.GetString(PARAM_ID_TRIGGER_SUPA_CREW_ACTION_SYMBOL_FORCE),
			PARAM_VALUE_TRIGGER_SUPA_CREW_ACTION_SYMBOL_ACTION_SYMBOL) {
		f.ForceTrigger(state, params)
	}

	random := params.GetInt(PARAM_ID_TRIGGER_SUPA_CREW_ACTION_SYMBOL_RANDOM)
	tileid := params.GetInt(PARAM_ID_TRIGGER_SUPA_CREW_ACTION_SYMBOL_TILE_ID)
	replaceid := random % 9
	params[featureProducts.PARAM_ID_REPLACE_TILE_REPLACE_WITH_ID] = replaceid
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
		params[featureProducts.PARAM_ID_REPLACE_TILE_POSITIONS] = positions
		feature.ActivateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f TriggerSupaCrewActionSymbol) ForceTrigger(state *feature.FeatureState, params feature.FeatureParams) {
	gridw, gridh := len(state.SymbolGrid), len(state.SymbolGrid[0])
	for x := 0; x < gridw; x++ {
		for y := 0; y < gridh; y++ {
			state.SymbolGrid[x][y] = state.SourceGrid[x][y]
		}
	}
	state.Features = []feature.Feature{}
	num := rng.RandFromRangePool(15) + 1
	tileid := params.GetInt(PARAM_ID_TRIGGER_SUPA_CREW_ACTION_SYMBOL_TILE_ID)
	for i := 0; i < num; i++ {
		x := rng.RandFromRangePool(5)
		y := rng.RandFromRangePool(3)
		state.SourceGrid[x][y] = tileid
		state.SymbolGrid[x][y] = tileid
	}
}

func (f *TriggerSupaCrewActionSymbol) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerSupaCrewActionSymbol) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
