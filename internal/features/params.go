package features

import "fmt"

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
	fmt.Printf("convertString in=%v\n", in)
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

func paramconvertpanic(paramname string) {
	if x := recover(); x != nil {
		panic(fmt.Errorf("param %s is %s", paramname, x))
	}
}

func param(params FeatureParams, paramname string) interface{} {
	val, ok := params[paramname]
	if !ok {
		panic("not set")
	}
	return val
}

func paramInt64(params FeatureParams, paramname string) int64 {
	defer paramconvertpanic(paramname)
	return convertInt64(param(params, paramname))
}

func paramUint64(params FeatureParams, paramname string) uint64 {
	defer paramconvertpanic(paramname)
	return convertUint64(param(params, paramname))
}

func paramInt32(params FeatureParams, paramname string) int32 {
	defer paramconvertpanic(paramname)
	return convertInt32(param(params, paramname))
}

func paramUint32(params FeatureParams, paramname string) uint32 {
	defer paramconvertpanic(paramname)
	return convertUint32(param(params, paramname))
}

func paramInt16(params FeatureParams, paramname string) int16 {
	defer paramconvertpanic(paramname)
	return convertInt16(param(params, paramname))
}

func paramUint16(params FeatureParams, paramname string) uint16 {
	defer paramconvertpanic(paramname)
	return convertUint16(param(params, paramname))
}

func paramInt8(params FeatureParams, paramname string) int8 {
	defer paramconvertpanic(paramname)
	return convertInt8(param(params, paramname))
}

func paramUint8(params FeatureParams, paramname string) uint8 {
	defer paramconvertpanic(paramname)
	return convertUint8(param(params, paramname))
}

func paramInt(params FeatureParams, paramname string) int {
	defer paramconvertpanic(paramname)
	return convertInt(param(params, paramname))
}

func paramUint(params FeatureParams, paramname string) uint {
	defer paramconvertpanic(paramname)
	return convertUint(param(params, paramname))
}

func paramFloat32(params FeatureParams, paramname string) float32 {
	defer paramconvertpanic(paramname)
	return convertFloat32(param(params, paramname))
}

func paramFloat64(params FeatureParams, paramname string) float64 {
	defer paramconvertpanic(paramname)
	return convertFloat64(param(params, paramname))
}

func paramComplex64(params FeatureParams, paramname string) complex64 {
	defer paramconvertpanic(paramname)
	return convertComplex64(param(params, paramname))
}

func paramComplex128(params FeatureParams, paramname string) complex128 {
	defer paramconvertpanic(paramname)
	return convertComplex128(param(params, paramname))
}

func paramString(params FeatureParams, paramname string) string {
	defer paramconvertpanic(paramname)
	return convertString(param(params, paramname))
}

func paramBool(params FeatureParams, paramname string) bool {
	defer paramconvertpanic(paramname)
	return convertBool(param(params, paramname))
}
