package features

import "fmt"

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

func convertInt64(in interface{}) int64 {
	val, ok := deserializeInt(in)
	if !ok {
		panic("not an int type")
	}
	return val
}

func convertUint64(in interface{}) uint64 {
	val, ok := deserializeUint(in)
	if !ok {
		panic("not an uint type")
	}
	return val
}

func convertFloat(in interface{}) float64 {
	val, ok := deserializeFloat(in)
	if !ok {
		panic("not a float type")
	}
	return val
}

func convertComplex(in interface{}) complex128 {
	val, ok := deserializeComplex(in)
	if !ok {
		panic("not a complex type")
	}
	return val
}

func convertBool(in interface{}) bool {
	val, ok := in.(bool)
	if !ok {
		panic("not a bool type")
	}
	return val
}

func convertString(in interface{}) string {
	val, ok := in.(string)
	if !ok {
		panic("not a string type")
	}
	return val
}

func convertInt32(in interface{}) int32 {
	return int32(convertInt64(in))
}

func convertUint32(in interface{}) uint32 {
	return uint32(convertUint64(in))
}

func convertInt16(in interface{}) int16 {
	return int16(convertInt64(in))
}

func convertUint16(in interface{}) uint16 {
	return uint16(convertUint64(in))
}

func convertInt8(in interface{}) int8 {
	return int8(convertInt64(in))
}

func convertUint8(in interface{}) uint8 {
	return uint8(convertUint64(in))
}

func convertInt(in interface{}) int {
	return int(convertInt64(in))
}

func convertUint(in interface{}) uint {
	return uint(convertUint64(in))
}

func convertFloat32(in interface{}) float32 {
	return float32(convertInt64(in))
}

func convertFloat64(in interface{}) float64 {
	return float64(convertUint64(in))
}

func convertComplex64(in interface{}) complex64 {
	return complex64(convertComplex64(in))
}

func convertComplex128(in interface{}) complex128 {
	return complex128(convertComplex128(in))
}

func convertIntSlice(in interface{}) []int {
	val, ok := in.([]int)
	if !ok {
		panic("not an int slize")
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

func (p FeatureParams) Get(name string) interface{} {
	val, ok := p[name]
	if !ok {
		panic("not set")
	}
	return val
}

func (p FeatureParams) GetInt64(name string) int64 {
	defer paramconvertpanic(name)
	return convertInt64(p.Get(name))
}

func (p FeatureParams) GetUint64(name string) uint64 {
	defer paramconvertpanic(name)
	return convertUint64(p.Get(name))
}

func (p FeatureParams) GetInt32(name string) int32 {
	defer paramconvertpanic(name)
	return convertInt32(p.Get(name))
}

func (p FeatureParams) GetUint32(name string) uint32 {
	defer paramconvertpanic(name)
	return convertUint32(p.Get(name))
}

func (p FeatureParams) GetInt16(name string) int16 {
	defer paramconvertpanic(name)
	return convertInt16(p.Get(name))
}

func (p FeatureParams) GetUint16(name string) uint16 {
	defer paramconvertpanic(name)
	return convertUint16(p.Get(name))
}

func (p FeatureParams) GetInt8(name string) int8 {
	defer paramconvertpanic(name)
	return convertInt8(p.Get(name))
}

func (p FeatureParams) GetUint8(name string) uint8 {
	defer paramconvertpanic(name)
	return convertUint8(p.Get(name))
}

func (p FeatureParams) GetInt(name string) int {
	defer paramconvertpanic(name)
	return convertInt(p.Get(name))
}

func (p FeatureParams) GetUint(name string) uint {
	defer paramconvertpanic(name)
	return convertUint(p.Get(name))
}

func (p FeatureParams) GetFloat32(name string) float32 {
	defer paramconvertpanic(name)
	return convertFloat32(p.Get(name))
}

func (p FeatureParams) GetFloat64(name string) float64 {
	defer paramconvertpanic(name)
	return convertFloat64(p.Get(name))
}

func (p FeatureParams) GetComplex64(name string) complex64 {
	defer paramconvertpanic(name)
	return convertComplex64(p.Get(name))
}

func (p FeatureParams) GetComplex128(name string) complex128 {
	defer paramconvertpanic(name)
	return convertComplex128(p.Get(name))
}

func (p FeatureParams) GetString(name string) string {
	defer paramconvertpanic(name)
	return convertString(p.Get(name))
}

func (p FeatureParams) GetBool(name string) bool {
	defer paramconvertpanic(name)
	return convertBool(p.Get(name))
}

func (p FeatureParams) GetIntSlice(name string) []int {
	defer paramconvertpanic(name)
	return convertIntSlice(p.Get(name))
}
