package features

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type FeatureDef struct {
	Id       int32         `yaml:"Id" json:"id"`
	Type     string        `yaml:"Type" json:"type"`
	Params   FeatureParams `yaml:"Params" json:"-"`
	Features []FeatureDef  `yaml:"Features" json:"-"`
}

type FeatureState struct {
	SymbolGrid [][]int
	Features   []Feature
}

type Feature interface {
	DefPtr() *FeatureDef
	DataPtr() interface{}
	Init(FeatureDef) error
	Trigger(FeatureState, FeatureParams) []Feature
	Serialize() ([]byte, error)
	Deserialize([]byte) error
}

// features must be included here to make them deserializable by the engine
type EnabledFeatureSet struct {
	_ FatTile
	_ InstaWin
	_ ReplaceTile
	_ TriggerSupaCrew
	_ TriggerSupaCrewActionSymbol
	_ TriggerSupaCrewFatTileChance
	_ TriggerSupaCrewFatTileReel
}

func MakeFeature(typename string) Feature {
	featuretype, ok := enabledFeatureMap[typename]
	if !ok {
		return nil
	}
	feature, ok := reflect.New(featuretype).Interface().(Feature)
	if !ok {
		return nil
	}
	return feature
}

//func GetTypeName(f Feature) string {
//	return reflect.TypeOf(f).Name()
//}

var enabledFeatureMap map[string]reflect.Type = buildFeatureMap(EnabledFeatureSet{})

func buildFeatureMap(featureset EnabledFeatureSet) map[string]reflect.Type {
	featuresetval := reflect.ValueOf(featureset)
	numfields := featuresetval.NumField()
	featuremap := make(map[string]reflect.Type)
	for i := 0; i < numfields; i++ {
		featureval := featuresetval.Field(i)
		featuretype := featureval.Type()
		featureptrtype := reflect.PtrTo(featuretype)
		if featureptrtype.Implements(reflect.TypeOf((*Feature)(nil)).Elem()) {
			featuremap[featuretype.Name()] = featuretype
		} else {
			panic(fmt.Sprintf("EnabledFeatureSet contains %s that doesn't implement Feature",
				featuretype.Name()))
		}
	}
	return featuremap
}

func mergeParams(p1 FeatureParams, p2 FeatureParams) (p FeatureParams) {
	p = make(FeatureParams, len(p1)+len(p2))
	for k, v := range p1 {
		p[k] = v
	}
	for k, v := range p2 {
		p[k] = v
	}
	return
}

func activateFeatures(def FeatureDef, state FeatureState, params FeatureParams) []Feature {
	activated := []Feature{}
	for _, featuredef := range def.Features {
		feature := MakeFeature(featuredef.Type)
		feature.Init(featuredef)
		activated = append(activated, feature.Trigger(state, mergeParams(featuredef.Params, params))...)
		// if mode, ok := params["mode"]; ok == true && mode.(string) == "replace" {
		//   state.Features = activated
		// } else {
		//   state.Features = append(state.Features, [...]activated)
		// }

	}
	return activated
}

func deserializeFeatureDef(f Feature, def FeatureDef) error {
	//	f.SetId(def.Id)
	//	f.SetType(def.Type)

	f.DefPtr().Id = def.Id
	f.DefPtr().Type = def.Type
	//	err := DeserializeStruct(f.DataPtr(), def.Params)
	//	if err != nil {
	//		logger.Errorf("err: %s", err.Error())
	//		return err
	//	}
	//	var dataptr interface{} = f.DataPtr()
	//	logger.Debugf("FeatureDef created: %v from params %v\n", dataptr, def.Params)
	// return err
	f.DefPtr().Params = def.Params
	f.DefPtr().Features = def.Features
	return nil
}

func deserializeFeatureFromBytes(f Feature, b []byte) error {
	return json.Unmarshal(b, f)
}

func serializeFeatureToBytes(f Feature) ([]byte, error) {
	return json.Marshal(f)
}

func deserializeTriggerFromBytes(f Feature, b []byte) error {
	return fmt.Errorf("trying to deserialize feature trigger %s from bytes", f.DefPtr().Type)
}

func serializeTriggerToBytes(f Feature) ([]byte, error) {
	return []byte{}, fmt.Errorf("trying to serialize feature trigger %s to bytes", f.DefPtr().Type)
}
