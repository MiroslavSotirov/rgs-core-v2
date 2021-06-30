package features

import (
	"encoding/json"
	"fmt"
	"reflect"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type FeatureDef struct {
	Id     int32                  `yaml:"Id"`
	Type   string                 `yaml:"Type"`
	Params map[string]interface{} `yaml:"Params"`
}

type FeatureState struct {
	SymbolGrid [][]int
	Features   []Feature
}

type Feature interface {
	GetId() int32
	GetType() string
	SetId(int32)
	SetType(string)
	DataPtr() interface{}
	Init(FeatureDef) error
	Trigger(FeatureState) []Feature
	Serialize() ([]byte, error)
	Deserialize([]byte) error
}

// features must be included here to make them deserializable by the engine
type EnabledFeatureSet struct {
	_ FatTileReel
	_ FatTileChance
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

func deserializeFeatureDef(f Feature, def FeatureDef) error {
	f.SetId(def.Id)
	f.SetType(def.Type)
	err := DeserializeStruct(f.DataPtr(), def.Params)
	if err != nil {
		logger.Errorf("err: %s", err.Error())
		return err
	}
	var dataptr interface{} = f.DataPtr()
	logger.Debugf("FeatureDef created: %v from params %v\n", dataptr, def.Params)

	return err
}

func deserializeFeatureFromBytes(f Feature, b []byte) error {
	return json.Unmarshal(b, f)
}

func serializeFeatureToBytes(f Feature) ([]byte, error) {
	return json.Marshal(f)
}

func DeserializeStruct(data interface{}, params map[string]interface{}) error {
	datainterptr := reflect.ValueOf(data)
	datainterval := datainterptr.Elem()
	dataintertype := datainterval.Type()
	numfields := datainterval.NumField()
	for i := 0; i < numfields; i++ {
		dataval := datainterval.Field(i)
		datatype := dataval.Type()
		datakind := datatype.Kind()
		dataname := dataintertype.Field(i).Name
		param, ok := params[dataname]
		if !ok {
			//			return fmt.Errorf("DeserializeStruct could not find param %s in map", dataname)
			continue
		}
		paramval := reflect.ValueOf(param)
		//		paramtype := paramval.Type()
		switch datakind {
		case reflect.Bool:
			v, ok := param.(bool)
			if !ok {
				return fmt.Errorf("DeserializeStruct could not set %s as a bool", dataname)
			}
			dataval.SetBool(v)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v, ok := deserializeInt(param)
			if !ok {
				return fmt.Errorf("DeserializeStruct could not set %s as an int", dataname)
			}
			dataval.SetInt(v)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			v, ok := deserializeUint(param)
			if !ok {
				return fmt.Errorf("DeserializeStruct could not set %s as an uint", dataname)
			}
			dataval.SetUint(v)
		case reflect.Float32, reflect.Float64:
			v, ok := deserializeFloat(param)
			if !ok {
				return fmt.Errorf("DeserializeStruct could not set %s as a float", dataname)
			}
			dataval.SetFloat(v)
		case reflect.Complex64, reflect.Complex128:
			v, ok := deserializeComplex(param)
			if !ok {
				return fmt.Errorf("DeserializeStruct could not set %s as a complex", dataname)
			}
			dataval.SetComplex(v)
		case reflect.String:
			v, ok := params[dataname].(string)
			if !ok {
				return fmt.Errorf("DeserializeStruct could not set %s as a string", dataname)
			}
			dataval.SetString(v)
			//		case reflect.Struct:
			//			if paramkind == reflect.Map {
			//				parammap, ok := param.(map[string]interface{})
			//				if !ok {
			//					return fmt.Errorf("DeserializeStruct could not assert map type of param %s", dataname)
			//				}
			//				err := DeserializeStruct(dataval.Addr().Interface(), parammap)
			//				if err != nil {
			//					return err
			//				}
			//			} else {
			//				dataval.Set(paramval)
			//			}
		default:
			dataval.Set(paramval)
		}
	}
	return nil
}

func SerializeStruct(data interface{}) map[string]interface{} {
	datainterval := reflect.ValueOf(data)
	dataintertype := datainterval.Type()
	numfields := datainterval.NumField()
	params := make(map[string]interface{})
	for i := 0; i < numfields; i++ {
		dataval := datainterval.Field(i)
		datatype := dataval.Type()
		datakind := datatype.Kind()
		dataname := dataintertype.Field(i).Name
		switch datakind {
		case reflect.Bool:
			params[dataname] = dataval.Bool()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			params[dataname] = dataval.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			params[dataname] = dataval.Uint()
		case reflect.Float32, reflect.Float64:
			params[dataname] = dataval.Float()
		case reflect.Complex64, reflect.Complex128:
			params[dataname] = dataval.Complex()
		case reflect.String:
			params[dataname] = dataval.String()
			//    	case reflect.Struct:
		default: //reflect.Interface, reflect.Map, reflect.Array, reflect.Slize:
			params[dataname] = dataval.Interface()
		}
	}
	return params
}

func deserializeInt(in interface{}) (val int64, ok bool) {
	switch in.(type) {
	case int:
		val, ok = int64(in.(int)), true
	case int8:
		val, ok = int64(in.(int8)), true
	case int16:
		val, ok = int64(in.(int16)), true
	case int32:
		val, ok = int64(in.(int32)), true
	case int64:
		val, ok = in.(int64), true
	default:
		val, ok = 0, false
	}
	return
}

func deserializeUint(in interface{}) (val uint64, ok bool) {
	switch in.(type) {
	case uint:
		val, ok = uint64(in.(uint)), true
	case uint8:
		val, ok = uint64(in.(uint8)), true
	case uint16:
		val, ok = uint64(in.(uint16)), true
	case uint32:
		val, ok = uint64(in.(uint32)), true
	case uint64:
		val, ok = in.(uint64), true
	default:
		val, ok = 0, false
	}
	return
}

func deserializeFloat(in interface{}) (val float64, ok bool) {
	switch in.(type) {
	case float32:
		val, ok = float64(in.(float32)), true
	case float64:
		val, ok = float64(in.(float64)), true
	default:
		val, ok = 0, false
	}
	return
}

func deserializeComplex(in interface{}) (val complex128, ok bool) {
	switch in.(type) {
	case complex64:
		val, ok = complex128(in.(complex64)), true
	case complex128:
		val, ok = complex128(in.(complex128)), true
	default:
		val, ok = 0, false
	}
	return
}
