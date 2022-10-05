package feature

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type FeatureParams map[string]interface{}

// help functions to read parameters with type safety checks

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
		var fval float64
		fval, ok = deserializeFloat(in)
		val = int64(fval)
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
		var fval float64
		fval, ok = deserializeFloat(in)
		val = uint64(fval)
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

func ConvertInt64(in interface{}) int64 {
	val, ok := deserializeInt(in)
	if !ok {
		panic("not an int type")
	}
	return val
}

func ConvertUint64(in interface{}) uint64 {
	val, ok := deserializeUint(in)
	if !ok {
		panic("not an uint type")
	}
	return val
}

func ConvertFloat(in interface{}) float64 {
	val, ok := deserializeFloat(in)
	if !ok {
		panic("not a float type")
	}
	return val
}

func ConvertComplex(in interface{}) complex128 {
	val, ok := deserializeComplex(in)
	if !ok {
		panic("not a complex type")
	}
	return val
}

func ConvertBool(in interface{}) bool {
	val, ok := in.(bool)
	if !ok {
		panic("not a bool type")
	}
	return val
}

func ConvertString(in interface{}) string {
	val, ok := in.(string)
	if !ok {
		panic("not a string type")
	}
	return val
}

func ConvertInt32(in interface{}) int32 {
	return int32(ConvertInt64(in))
}

func ConvertUint32(in interface{}) uint32 {
	return uint32(ConvertUint64(in))
}

func ConvertInt16(in interface{}) int16 {
	return int16(ConvertInt64(in))
}

func ConvertUint16(in interface{}) uint16 {
	return uint16(ConvertUint64(in))
}

func ConvertInt8(in interface{}) int8 {
	return int8(ConvertInt64(in))
}

func ConvertUint8(in interface{}) uint8 {
	return uint8(ConvertUint64(in))
}

func ConvertInt(in interface{}) int {
	return int(ConvertInt64(in))
}

func ConvertUint(in interface{}) uint {
	return uint(ConvertUint64(in))
}

func ConvertFloat32(in interface{}) float32 {
	return float32(ConvertInt64(in))
}

func ConvertFloat64(in interface{}) float64 {
	return float64(ConvertUint64(in))
}

func ConvertComplex64(in interface{}) complex64 {
	return complex64(ConvertComplex64(in))
}

func ConvertComplex128(in interface{}) complex128 {
	return complex128(ConvertComplex128(in))
}

func ConvertIntSlice(in interface{}) []int {
	val, ok := in.([]int)
	if !ok {
		var val2 []interface{}
		val2, ok = in.([]interface{})
		if !ok {
			panic("not a slize")
		}
		val = make([]int, len(val2))
		for i, av := range val2 {
			var v int
			v, ok = av.(int)
			if !ok {
				panic("not an int slize")
			}
			val[i] = v
		}
	}
	return val
}

func ConvertSlice(in interface{}) []interface{} {
	val, ok := in.([]interface{})
	if !ok {
		panic("not a slice")
	}
	return val
}

func ConvertParams(in interface{}) FeatureParams {
	val, ok := in.(FeatureParams)
	if !ok {
		logger.Debugf("type: %s", reflect.TypeOf(in).Name())
		panic("not a map[string]interface{} type")
	}
	return val
}

func ConvertParamsSlice(in interface{}) []FeatureParams {
	val, ok := in.([]FeatureParams)
	if !ok {
		var val2 []interface{}
		val2, ok = in.([]interface{})
		if !ok {
			panic("not a slize")
		}
		val = make([]FeatureParams, len(val2))
		for i, av := range val2 {
			var vm map[interface{}]interface{}
			vm, ok = av.(map[interface{}]interface{})
			if !ok {
				panic("not a map slize")
			}
			var fp FeatureParams = make(FeatureParams)
			for k, v := range vm {
				var kstr string
				kstr, ok = k.(string)
				if !ok {
					panic("not a params slize. contanins a non string key")
				}
				fp[kstr] = v
			}
			val[i] = fp
		}
	}
	return val

}

func paramconvertpanic(name string) {
	if x := recover(); x != nil {
		panic(fmt.Errorf("param %s is %s", name, x))
	}
}

func (p FeatureParams) HasKey(name string) bool {
	_, ok := p[name]
	return ok
}

func (p FeatureParams) HasValue(name string) bool {
	val, ok := p[name]
	return ok && val != nil
}

func (p FeatureParams) Get(name string) interface{} {
	val, ok := p[name]
	if !ok {
		panic("not set")
	}
	return val
}

func (p FeatureParams) GetInt64(name string) int64 {
	defer paramconvertpanic(name)
	return ConvertInt64(p.Get(name))
}

func (p FeatureParams) GetUint64(name string) uint64 {
	defer paramconvertpanic(name)
	return ConvertUint64(p.Get(name))
}

func (p FeatureParams) GetInt32(name string) int32 {
	defer paramconvertpanic(name)
	return ConvertInt32(p.Get(name))
}

func (p FeatureParams) GetUint32(name string) uint32 {
	defer paramconvertpanic(name)
	return ConvertUint32(p.Get(name))
}

func (p FeatureParams) GetInt16(name string) int16 {
	defer paramconvertpanic(name)
	return ConvertInt16(p.Get(name))
}

func (p FeatureParams) GetUint16(name string) uint16 {
	defer paramconvertpanic(name)
	return ConvertUint16(p.Get(name))
}

func (p FeatureParams) GetInt8(name string) int8 {
	defer paramconvertpanic(name)
	return ConvertInt8(p.Get(name))
}

func (p FeatureParams) GetUint8(name string) uint8 {
	defer paramconvertpanic(name)
	return ConvertUint8(p.Get(name))
}

func (p FeatureParams) GetInt(name string) int {
	defer paramconvertpanic(name)
	return ConvertInt(p.Get(name))
}

func (p FeatureParams) GetUint(name string) uint {
	defer paramconvertpanic(name)
	return ConvertUint(p.Get(name))
}

func (p FeatureParams) GetFloat32(name string) float32 {
	defer paramconvertpanic(name)
	return ConvertFloat32(p.Get(name))
}

func (p FeatureParams) GetFloat64(name string) float64 {
	defer paramconvertpanic(name)
	return ConvertFloat64(p.Get(name))
}

func (p FeatureParams) GetComplex64(name string) complex64 {
	defer paramconvertpanic(name)
	return ConvertComplex64(p.Get(name))
}

func (p FeatureParams) GetComplex128(name string) complex128 {
	defer paramconvertpanic(name)
	return ConvertComplex128(p.Get(name))
}

func (p FeatureParams) GetString(name string) string {
	defer paramconvertpanic(name)
	return ConvertString(p.Get(name))
}

func (p FeatureParams) AsString(name string) string {
	defer paramconvertpanic(name)
	v := p.Get(name)
	var s string
	switch v.(type) {
	case string:
		s = v.(string)
	case int:
		s = fmt.Sprintf("%d", v.(int))
	case int8:
		s = fmt.Sprintf("%d", v.(int8))
	case int16:
		s = fmt.Sprintf("%d", v.(int16))
	case int32:
		s = fmt.Sprintf("%d", v.(int32))
	case int64:
		s = fmt.Sprintf("%d", v.(int64))
	case uint:
		s = fmt.Sprintf("%d", v.(uint))
	case uint8:
		s = fmt.Sprintf("%d", v.(uint8))
	case uint16:
		s = fmt.Sprintf("%d", v.(uint16))
	case uint32:
		s = fmt.Sprintf("%d", v.(uint32))
	case uint64:
		s = fmt.Sprintf("%d", v.(uint64))
	default:
	}
	return s
}

func (p FeatureParams) GetBool(name string) bool {
	defer paramconvertpanic(name)
	return ConvertBool(p.Get(name))
}

func (p FeatureParams) GetIntSlice(name string) []int {
	defer paramconvertpanic(name)
	return ConvertIntSlice(p.Get(name))
}

func (p FeatureParams) GetSlice(name string) []interface{} {
	defer paramconvertpanic(name)
	return ConvertSlice(p.Get(name))
}

func (p FeatureParams) GetParams(name string) FeatureParams {
	defer paramconvertpanic(name)
	return ConvertParams(p.Get(name))
}

func (p FeatureParams) GetParamsSlice(name string) []FeatureParams {
	defer paramconvertpanic(name)
	return ConvertParamsSlice(p.Get(name))
}

// get a specific force from a list of space separated forces [force:value] or [forceflag]
func (p FeatureParams) GetForce(force string) string {
	if config.GlobalConfig.DevMode && p.HasKey("force") {
		forces := strings.Split(p.GetString("force"), " ")
		for _, f := range forces {
			if strings.Contains(f, force) {
				var val string
				if strings.Contains(f, ":") {
					parts := strings.Split(f, ":")
					val, _ = url.PathUnescape(parts[1])
				} else {
					val = f
				}
				return val
			}
		}
	}
	return ""
}

func (p FeatureParams) GetForceFloat64(force string) (v float64, ok bool) {
	ok = false
	f := p.GetForce(force)
	if f != "" {
		val, err := strconv.ParseFloat(f, 64)
		if err != nil {
			return
		}
		ok = true
		v = val
	}
	return
}
