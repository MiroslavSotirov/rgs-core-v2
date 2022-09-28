package feature

import (
	"fmt"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

const (
	FEATURE_ID_STATEFUL_MAP = "StatefulMap"
)

var _ Factory = RegisterFeature(FEATURE_ID_STATEFUL_MAP, func() Feature { return new(StatefulMap) })

type StatefulMapData struct {
	Map FeatureParams `json:"map"`
}

type StatefulMap struct {
	Base
	Data StatefulMapData `json:"data"`
}

func (f *StatefulMap) DataPtr() interface{} {
	return &f.Data
}

func (f StatefulMap) Trigger(state *FeatureState, params FeatureParams) {
	if !params.HasKey(FEATURE_ID_STATEFUL_MAP) {
		panic(FEATURE_ID_STATEFUL_MAP + " param must be set")
	}
	sfmap := params.GetParams(FEATURE_ID_STATEFUL_MAP)
	state.Features = append(state.Features,
		&StatefulMap{
			Base: Base{FeatureDef: *f.DefPtr()},
			Data: StatefulMapData{Map: sfmap},
		})
}

func (f *StatefulMap) Serialize() ([]byte, error) {
	return SerializeFeatureToBytes(f)
}

func (f *StatefulMap) Deserialize(data []byte) (err error) {
	return DeserializeFeatureFromBytes(f, data)
}

// help functions to get and set data related to the selected stake
func GetStatefulMap(state FeatureState) (statefulMap FeatureParams) {
	if state.Stateful != nil {
		sf := FindFeature(FEATURE_ID_STATEFUL_MAP, state.Stateful.Features)
		if sf != nil {
			sfmap := sf.(*StatefulMap)
			if sfmap != nil {
				statefulMap = FeatureParams{}
				for k, v := range sfmap.Data.Map {
					statefulMap[k] = v
				}
			} else {
				panic(FEATURE_ID_STATEFUL_MAP + " has wrong type")
			}
		} else {
			logger.Debugf("no " + FEATURE_ID_STATEFUL_MAP + " in previous gamestate")
		}
	} else {
		panic("feature state is not stateful")
	}
	return
}

func SetStatefulMap(statefulMap FeatureParams, params FeatureParams) {
	params["StatefulMap"] = statefulMap
}

func GetStatefulStakeMap(state FeatureState) (statefulStakeMap FeatureParams) {
	statefulMap := GetStatefulMap(state)
	stake := fmt.Sprintf("%.3f", state.TotalStake)
	statefulStakeMap = FeatureParams{}
	if statefulMap.HasKey(stake) {
		for k, v := range statefulMap.GetParams(stake) {
			statefulStakeMap[k] = v
		}
	}
	return
}

func SetStatefulStakeMap(state FeatureState, statefulStakeMap FeatureParams, params FeatureParams) {
	statefulMap := FeatureParams{}
	for k, v := range GetStatefulMap(state) {
		statefulMap[k] = v
	}
	stake := fmt.Sprintf("%.3f", state.TotalStake)
	statefulMap[stake] = statefulStakeMap
	SetStatefulMap(statefulMap, params)
}
