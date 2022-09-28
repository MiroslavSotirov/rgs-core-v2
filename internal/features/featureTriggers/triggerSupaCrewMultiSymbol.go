package featureTriggers

import (
	"strings"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

const (
	FEATURE_ID_TRIGGER_SUPA_CREW_MULTI_SYMBOL = "TriggerSupaCrewMultiSymbol"

	PARAM_ID_TRIGGER_SUPA_CREW_MULTI_SYMBOL_FORCE  = "force"
	PARAM_ID_TRIGGER_SUPA_CREW_MULTI_SYMBOL_RANDOM = "Random"

	PARAM_VALUE_TRIGGER_SUPA_CREW_MULTI_SYMBOL_MULTI_SYMBOL  = "multisymbol"
	PARAM_VALUE_TRIGGER_SUPA_CREW_MULTI_SYMBOL_SPINNING_COIN = "spinningcoin"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_SUPA_CREW_MULTI_SYMBOL, func() feature.Feature { return new(TriggerSupaCrewMultiSymbol) })

type TriggerSupaCrewMultiSymbol struct {
	feature.Base
}

func (f TriggerSupaCrewMultiSymbol) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	if params.HasKey(PARAM_ID_TRIGGER_SUPA_CREW_MULTI_SYMBOL_FORCE) &&
		strings.Contains(params.GetString(PARAM_ID_TRIGGER_SUPA_CREW_MULTI_SYMBOL_FORCE),
			PARAM_VALUE_TRIGGER_SUPA_CREW_MULTI_SYMBOL_MULTI_SYMBOL) {
		f.ForceTrigger(state, params)
		return
	}

	random := params.GetInt(PARAM_ID_TRIGGER_SUPA_CREW_MULTI_SYMBOL_RANDOM)
	randiv := random / 9
	if randiv < 30 {
		ran8 := rng.RandFromRange(8)
		y := ran8 / 4
		x := ran8 % 4

		params[featureProducts.PARAM_ID_FAT_TILE_X] = x
		params[featureProducts.PARAM_ID_FAT_TILE_Y] = y

		ran12 := rng.RandFromRange(12)
		params[featureProducts.PARAM_ID_INSTA_WIN_TYPE] = PARAM_VALUE_TRIGGER_SUPA_CREW_MULTI_SYMBOL_SPINNING_COIN
		params[featureProducts.PARAM_ID_INSTA_WIN_SOURCE_ID] = f.FeatureDef.Id
		params[featureProducts.PARAM_ID_INSTA_WIN_AMOUNT] = []int{
			7, 8, 10, 12, 14, 16, 18, 20, 22, 25, 28, 30,
		}[ran12]
		gridh := len(state.SymbolGrid[0])
		params[featureProducts.PARAM_ID_INSTA_WIN_POSITIONS] = []int{x*gridh + y, (x+1)*gridh + y, x*gridh + y + 1, (x+1)*gridh + y + 1}

		feature.ActivateFeatures(f.FeatureDef, state, params)
	}
}

func (f TriggerSupaCrewMultiSymbol) ForceTrigger(state *feature.FeatureState, params feature.FeatureParams) {
	params[featureProducts.PARAM_ID_FAT_TILE_X] = rng.RandFromRange(4)
	params[featureProducts.PARAM_ID_FAT_TILE_Y] = rng.RandFromRange(2)
	ran12 := rng.RandFromRange(12)
	params[featureProducts.PARAM_ID_INSTA_WIN_TYPE] = PARAM_VALUE_TRIGGER_SUPA_CREW_MULTI_SYMBOL_SPINNING_COIN
	params[featureProducts.PARAM_ID_INSTA_WIN_SOURCE_ID] = f.FeatureDef.Id
	params[featureProducts.PARAM_ID_INSTA_WIN_AMOUNT] = []int{
		7, 8, 10, 12, 14, 16, 18, 20, 22, 25, 28, 30,
	}[ran12]

	feature.ActivateFeatures(f.FeatureDef, state, params)
}

func (f *TriggerSupaCrewMultiSymbol) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerSupaCrewMultiSymbol) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
