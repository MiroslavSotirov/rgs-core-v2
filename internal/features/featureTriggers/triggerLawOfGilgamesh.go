package featureTriggers

import (
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

	if state.Action == "init" {
		state.Features = append(state.Features,
			&TriggerLawOfGilgamesh{
				Base: feature.Base{FeatureDef: *f.DefPtr()},
				Data: TriggerLawOfGilgameshData{
					Levels: levels,
				},
			})
		return
	} else if state.Action == "base" || state.Action == "freespin" {

		bonuses := feature.RandomPermutation([]int{1, 2, 3})
		stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_ORDER] = encodeOrd(bonuses[0], bonuses[1], bonuses[2])
		stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_LEVEL] = 0

		if state.Action == "base" {
			stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_COUNTER] = 0
		} else if state.Action == "freespin" {
			stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_COUNTER] = levels[0]
		}
	}

	triggers := params.GetStringSlice(PARAM_ID_TRIGGER_ORDERED_TRIGGERS)

	counter := stateless.GetInt(STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_COUNTER)
	ord := stateless.GetInt(STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_ORDER)
	level := stateless.GetInt(STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_LEVEL)

	logger.Debugf("gilgamesh counter %d ord %d level %d", counter, ord, level)

	wins := state.CalculateWins(state.SourceGrid, nil)
	for _, w := range wins {
		counter += len(w.SymbolPositions)
		logger.Debugf("increased counter by %d to %d", len(w.SymbolPositions), counter)
	}

	b1 := ord % 4
	b2 := (ord / 4) % 4
	b3 := ord / 16
	logger.Debugf("ord: %d b1: %d b2: %d b3: %d", ord, b1, b2, b3)
	order := []string{}
	if counter >= levels[0] && level < 1 {
		order = append(order, triggers[b1-1])
	}
	if counter >= levels[1] && level < 2 {
		order = append(order, triggers[b2-1])
	}
	if counter >= levels[2] && level < 3 {
		order = append(order, triggers[b3-1])
	}
	if len(order) > 0 {
		params[PARAM_ID_TRIGGER_ORDERED_ORDER] = order
		logger.Debugf("activation order is %v", order)
	}

	stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_COUNTER] = counter
	stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_ORDER] = ord
	stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_LEVEL] = level

	feature.SetStatelessMap(stateless, params)
	feature.ActivateFeatures(f.FeatureDef, state, params)

	wins = state.CalculateWins(state.SymbolGrid, nil)
	if len(wins) > 0 {

		for _, w := range wins {
			counter += len(w.SymbolPositions)
			logger.Debugf("increased counter by %d to %d", len(w.SymbolPositions), counter)
		}
		stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_COUNTER] = counter
		feature.SetStatelessMap(stateless, params)
	}

	//	state.Features = append(state.Features{
	statelessMap := feature.MakeFeature(feature.FEATURE_ID_STATELESS_MAP)
	statelessMap.Init(feature.FeatureDef{Type: "StatelessMap"})
	statelessMap.Trigger(state, params)

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

func incLawOfGilgameshLevel(state *feature.FeatureState, params feature.FeatureParams) {
	stateless := feature.GetParamStatelessMap(params)
	stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_LEVEL] = stateless.GetInt(STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_LEVEL) + 1
	feature.SetStatelessMap(stateless, params)
}
