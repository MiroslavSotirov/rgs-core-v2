package feature

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

const (
	FEATURE_ID_STATELESS_MAP = "StatelessMap"
)

var _ Factory = RegisterFeature(FEATURE_ID_STATELESS_MAP, func() Feature { return new(StatelessMap) })

/*
type StatelessMapData struct {
	Map FeatureParams `json:"map"`
}
*/
type StatelessMap struct {
	Base
	Data FeatureParams `json:"data"`
}

func (f *StatelessMap) DataPtr() interface{} {
	return &f.Data
}

func (f StatelessMap) Trigger(state *FeatureState, params FeatureParams) {
	if !params.HasKey(FEATURE_ID_STATELESS_MAP) {
		panic(FEATURE_ID_STATELESS_MAP + " param must be set")
	}
	sfmap := params.GetParams(FEATURE_ID_STATELESS_MAP)
	state.Features = append(state.Features,
		&StatelessMap{
			Base: Base{FeatureDef: *f.DefPtr()},
			Data: sfmap,
		})
}

func (f *StatelessMap) Serialize() ([]byte, error) {
	return SerializeFeatureToBytes(f)
}

func (f *StatelessMap) Deserialize(data []byte) (err error) {
	return DeserializeFeatureFromBytes(f, data)
}

// help functions to get and set data related to the selected stake
func GetStatelessMap(state FeatureState) (statelessMap FeatureParams) {
	if state.Stateless != nil {
		sf := FindFeature(FEATURE_ID_STATELESS_MAP, state.Stateless.Features)
		if sf != nil {
			sfmap := sf.(*StatelessMap)
			if sfmap != nil {
				statelessMap = FeatureParams{}
				for k, v := range sfmap.Data {
					statelessMap[k] = v
				}
			} else {
				panic(FEATURE_ID_STATELESS_MAP + " has wrong type")
			}
		} else {
			logger.Debugf("no " + FEATURE_ID_STATELESS_MAP + " in previous gamestate")
			statelessMap = FeatureParams{}
		}
	} else {
		logger.Debugf("no features from previous spin. use empty statelessMap")
		statelessMap = FeatureParams{}
	}
	return
}

func GetParamStatelessMap(params FeatureParams) (statelessMap FeatureParams) {
	if params.HasKey(FEATURE_ID_STATELESS_MAP) {
		statelessMap = params.GetParams(FEATURE_ID_STATELESS_MAP)
	}
	return
}

func SetStatelessMap(statelessMap FeatureParams, params FeatureParams) {
	logger.Debugf("stateless map: %#v", statelessMap)
	params["StatelessMap"] = statelessMap
}
