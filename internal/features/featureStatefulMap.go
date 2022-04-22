package features

import (
	"fmt"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type StatefulMapData struct {
	Map FeatureParams `json:"map"`
}

type StatefulMap struct {
	FeatureDef
	Data StatefulMapData `json:"data"`
}

func (f *StatefulMap) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *StatefulMap) DataPtr() interface{} {
	return &f.Data
}

func (f *StatefulMap) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *StatefulMap) OnInit(state *FeatureState) {
}

func (f StatefulMap) Trigger(state *FeatureState, params FeatureParams) {
	if !params.HasKey("StatefulMap") {
		panic("StatefulMap param must be set")
	}
	sfmap := params.GetParams("StatefulMap")
	state.Features = append(state.Features,
		&StatefulMap{
			FeatureDef: *f.DefPtr(),
			Data:       StatefulMapData{Map: sfmap},
		})
	for i, f := range state.Features {
		logger.Debugf("feature array [%d] %#v", i, f)
	}
}

func (f *StatefulMap) Serialize() ([]byte, error) {
	return serializeFeatureToBytes(f)
}

func (f *StatefulMap) Deserialize(data []byte) (err error) {
	return deserializeFeatureFromBytes(f, data)
}

// help functions to get and set data related to the selected stake
func GetStatefulMap(state FeatureState) (statefulMap FeatureParams) {
	if state.Stateful != nil {
		sf := FindFeature("StatefulMap", state.Stateful.Features)
		if sf != nil {
			sfmap := sf.(*StatefulMap)
			if sfmap != nil {
				statefulMap = FeatureParams{}
				for k, v := range sfmap.Data.Map {
					statefulMap[k] = v
				}
			} else {
				panic("StatefulMap has wrong type")
			}
		} else {
			logger.Debugf("no StatefulMap in previous gamestate")
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
