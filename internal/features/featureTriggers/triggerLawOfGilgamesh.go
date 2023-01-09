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
	STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_LEVEL   = "level"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH, func() feature.Feature { return new(TriggerLawOfGilgamesh) })

type TriggerLawOfGilgameshData struct {
	Levels []int `json:"levels"`
}

type TriggerLawOfGilgamesh struct {
	feature.Base
	Data TriggerLawOfGilgameshData `json:"lawOfGilgamesh"`
}

func (f *TriggerLawOfGilgamesh) DataPtr() interface{} {
	return &f.Data
}

func (f TriggerLawOfGilgamesh) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	if f.ForceTrigger(state, params) {
		logger.Debugf("force %s was applied so no more features will be executed", params.GetString("force"))
		return
	}

	stateless := feature.FeatureParams{
		STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_COUNTER: 0,
		STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_ORDER:   0,
		STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_LEVEL:   0,
	}
	for k, v := range feature.GetStatelessMap(*state) {
		stateless[k] = v
	}
	logger.Debugf("state action = %#v", state.Action)

	encodeOrd := func(l1 int, l2 int, l3 int) int {
		return (l1 % 4) + ((l2%4)+(l3%4)*4)*4
	}

	levels := params.GetIntSlice(PARAM_ID_TRIGGER_LAW_OF_GILGAMESH_LEVELS)

	if state.Action == "base" {
		bonuses := feature.RandomPermutation([]int{1, 2, 3})

		stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_COUNTER] = 0
		stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_ORDER] = encodeOrd(bonuses[0], bonuses[1], bonuses[2])
		stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_LEVEL] = 0
	} else if state.Action == "init" {
		state.Features = append(state.Features,
			&TriggerLawOfGilgamesh{
				Base: feature.Base{FeatureDef: *f.DefPtr()},
				Data: TriggerLawOfGilgameshData{
					Levels: levels,
				},
			})
	}

	counter := stateless.GetInt(STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_COUNTER)
	ord := stateless.GetInt(STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_ORDER)
	level := stateless.GetInt(STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_LEVEL)

	wins := state.CalculateWins(state.SourceGrid, nil)
	for _, w := range wins {
		counter += len(w.SymbolPositions)
		logger.Debugf("increased counter by %d to %d", len(w.SymbolPositions), counter)
	}

	bonus := 0
	if len(wins) == 0 {
		//		counter = 20
		b1 := ord % 4
		b2 := (ord / 4) % 4
		b3 := ord / 16
		logger.Debugf("ord: %d b1: %d b2: %d b3: %d", ord, b1, b2, b3)
		if counter >= levels[0] && level == 0 {
			bonus, level = b1, 1
		} else if counter >= levels[1] && level == 1 {
			bonus, level = b2, 2
		} else if counter >= levels[2] && level == 2 {
			bonus, level = b3, 3
		}
	}
	logger.Debugf("bonus is %d", bonus)
	stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_COUNTER] = counter
	stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_ORDER] = ord
	stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_LEVEL] = level
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
