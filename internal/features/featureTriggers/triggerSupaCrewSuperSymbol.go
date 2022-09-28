package featureTriggers

import (
	"strings"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

const (
	FEATURE_ID_TRIGGER_SUPA_CREW_SUPER_SYMBOL = "TriggerSupaCrewSuperSymbol"

	PARAM_ID_TRIGGER_SUPA_CREW_SUPER_SYMBOL_FORCE  = "force"
	PARAM_ID_TRIGGER_SUPA_CREW_SUPER_SYMBOL_RANDOM = "Random"

	PARAM_VALUE_TRIGGER_SUPA_CREW_SUPER_SYMBOL_SUPER_SYMBOL = "supersymbol"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_SUPA_CREW_SUPER_SYMBOL, func() feature.Feature { return new(TriggerSupaCrewSuperSymbol) })

type TriggerSupaCrewSuperSymbol struct {
	feature.Base
}

func (f TriggerSupaCrewSuperSymbol) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	if params.HasKey(PARAM_ID_TRIGGER_SUPA_CREW_SUPER_SYMBOL_FORCE) &&
		strings.Contains(params.GetString(PARAM_ID_TRIGGER_SUPA_CREW_SUPER_SYMBOL_FORCE),
			PARAM_VALUE_TRIGGER_SUPA_CREW_SUPER_SYMBOL_SUPER_SYMBOL) {
		f.ForceTrigger(state, params)
		return
	}

	random := params.GetInt(PARAM_ID_TRIGGER_SUPA_CREW_SUPER_SYMBOL_RANDOM)
	ran15 := rng.RandFromRange(15)
	ran9 := random / 9
	if ran9 >= 30 && ran9 <= 39 {
		x := ran15 / 5
		y := []int{-2, -1, 0, 1, 2}[ran15%5]
		params[featureProducts.PARAM_ID_FAT_TILE_W] = 3
		params[featureProducts.PARAM_ID_FAT_TILE_H] = 3
		params[featureProducts.PARAM_ID_FAT_TILE_X] = x
		params[featureProducts.PARAM_ID_FAT_TILE_Y] = y
		params[featureProducts.PARAM_ID_FAT_TILE_TILE_ID] = random % 9

		feature.ActivateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f TriggerSupaCrewSuperSymbol) ForceTrigger(state *feature.FeatureState, params feature.FeatureParams) {
	params[featureProducts.PARAM_ID_FAT_TILE_X] = rng.RandFromRange(3)
	params[featureProducts.PARAM_ID_FAT_TILE_Y] = rng.RandFromRange(5) - 2
	params[featureProducts.PARAM_ID_FAT_TILE_W] = 3
	params[featureProducts.PARAM_ID_FAT_TILE_H] = 3
	params[featureProducts.PARAM_ID_FAT_TILE_TILE_ID] = rng.RandFromRange(9)

	feature.ActivateFeatures(f.FeatureDef, state, params)
}

func (f *TriggerSupaCrewSuperSymbol) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerSupaCrewSuperSymbol) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
