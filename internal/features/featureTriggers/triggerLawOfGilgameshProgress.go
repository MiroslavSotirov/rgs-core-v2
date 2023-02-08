package featureTriggers

import (
	"strconv"
	"strings"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

const (
	FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH_PROGRESS = "TriggerLawOfGilgameshProgress"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_LAW_OF_GILGAMESH_PROGRESS, func() feature.Feature { return new(TriggerLawOfGilgameshProgress) })

type TriggerLawOfGilgameshProgress struct {
	feature.Base
}

func (f TriggerLawOfGilgameshProgress) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	stateless := feature.GetParamStatelessMap(params)

	wins := state.CalculateWins(state.SymbolGrid, nil)
	if len(wins) > 0 {

		stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_COUNTER] = stateless.GetInt(STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_COUNTER) +
			countLawOfGilgameshWin(state)
		feature.SetStatelessMap(stateless, params)
	} else {
		feature.ActivateFeatures(f.FeatureDef, state, params)
	}

	return
}

func (f *TriggerLawOfGilgameshProgress) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerLawOfGilgameshProgress) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}

func countLawOfGilgameshWin(state *feature.FeatureState) int {
	counter := 0
	wins := state.CalculateWins(state.SymbolGrid, nil)
	for _, w := range wins {
		num, err := strconv.ParseInt(strings.Split(w.Index, ":")[1], 10, 64)
		if err != nil {
			panic("could not parse win for cluser size determination")
		}
		logger.Debugf("cluster size %d from win %#v", num, w)
		counter += int(num)                                                           // len(w.SymbolPositions)
		logger.Debugf("increased counter after activation by %d to %d", num, counter) //len(w.SymbolPositions), counter)
	}
	return counter
}

func incLawOfGilgameshLevel(state *feature.FeatureState, params feature.FeatureParams) {
	stateless := feature.GetParamStatelessMap(params)
	stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_LEVEL] = stateless.GetInt(STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_LEVEL) + 1
	stateless[STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_COUNTER] = stateless.GetInt(STATELESS_ID_TRIGGER_LAW_OF_GILGAMESH_COUNTER) +
		countLawOfGilgameshWin(state)
	feature.SetStatelessMap(stateless, params)
}
