package featureTriggers

import (
	"fmt"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

const (
	FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH = "TriggerLawOfGilgamesh"

	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_LEVELS = "Levels"
	PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_BONUS  = "Bonus"

	STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_COUNTER = "counter"
	STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_ORDER   = "order"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH, func() feature.Feature { return new(TriggerLawOfGilgamesh) })

type TriggerLawOfGilgamesh struct {
	feature.Base
}

func (f TriggerLawOfGilgamesh) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	if f.ForceTrigger(state, params) {
		logger.Debugf("force %s was applied so no more features will be executed", params.GetString("force"))
		return
	}

	stateless := feature.FeatureParams{
		STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_COUNTER: 0,
		STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_ORDER:   0,
	}
	for k, v := range feature.GetStatelessMap(*state) {
		stateless[k] = v
	}

	encodeOrd := func(l1 int, l2 int, l3 int) int {
		return (l1 % 4) + ((l2%4)+(l3%4)*4)*4
	}

	if state.Action == "base" {
		bonuses := feature.RandomPermutation([]int{1, 2, 3})

		stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_COUNTER] = 0
		stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_ORDER] = encodeOrd(bonuses[0], bonuses[1], bonuses[2])
	}

	wins := state.CalculateWins(state.SourceGrid, nil)
	counter := stateless.GetInt(STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_COUNTER)
	for _, w := range wins {
		counter += len(w.SymbolPositions)
		logger.Debugf("increased counter by %d to %d", len(w.SymbolPositions), counter)
	}

	ord := stateless.GetInt(STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_ORDER)
	levels := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_LEVELS)
	bonus := 0
	if len(wins) == 0 {
		//		counter = 20
		b1 := ord % 4
		b2 := (ord / 4) % 4
		b3 := ord / 16
		logger.Debugf("ord: %d b1: %d b2: %d b3: %d", ord, b1, b2, b3)
		if counter >= levels[0] && b1 > 0 {
			bonus, b1 = b1, 0
		} else if counter >= levels[1] && b2 > 0 {
			bonus, b2 = b2, 0
		} else if counter >= levels[2] && b3 > 0 {
			bonus, b3 = b3, 0
		}
		ord = encodeOrd(b1, b2, b3)
	}
	logger.Debugf("bonus is %d", bonus)
	stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_COUNTER] = counter
	stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_ORDER] = ord
	if bonus > 0 {
		params[PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_BONUS] = fmt.Sprintf("%d", bonus)
	}

	feature.SetStatelessMap(stateless, params)

	feature.ActivateFeatures(f.FeatureDef, state, params)
	return
}

func (f TriggerLawOfGilgamesh) ForceTrigger(state *feature.FeatureState, params feature.FeatureParams) bool {
	return false
}

func (f *TriggerLawOfGilgamesh) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerLawOfGilgamesh) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
