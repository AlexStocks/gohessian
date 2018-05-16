/******************************************************
# DESC    : decode hessian response
# AUTHOR  : Alex Stocks
# VERSION : 1.0
# LICENCE : Apache License 2.0
# EMAIL   : alexstocks@foxmail.com
# MOD     : 2017-10-17 11:12
# FILE    : response.go
******************************************************/

package hessian

import (
	"encoding/binary"
	"reflect"
)

import (
	jerrors "github.com/juju/errors"
)

const (
	Response_OK                byte = 20
	Response_CLIENT_TIMEOUT    byte = 30
	Response_SERVER_TIMEOUT    byte = 31
	Response_BAD_REQUEST       byte = 40
	Response_BAD_RESPONSE      byte = 50
	Response_SERVICE_NOT_FOUND byte = 60
	Response_SERVICE_ERROR     byte = 70
	Response_SERVER_ERROR      byte = 80
	Response_CLIENT_ERROR      byte = 90

	RESPONSE_WITH_EXCEPTION int32 = 0
	RESPONSE_VALUE          int32 = 1
	RESPONSE_NULL_VALUE     int32 = 2
)

var (
	ErrIllegalPackage = jerrors.Errorf("illegal pacakge!")
)

// hessian decode respone
func UnpackResponse(buf []byte) (interface{}, error) {
	length := len(buf)
	if length < HEADER_LENGTH || (buf[0] != byte(MAGIC_HIGH) && buf[1] != byte(MAGIC_LOW)) {
		return nil, ErrIllegalPackage
	}

	// Header{serialization id(5 bit), event, two way, req/response}
	var serialID = buf[2] & SERIALIZATION_MASK
	if serialID == byte(0x00) {
		return nil, jerrors.Errorf("serialization ID:%v", serialID)
	}
	//var eventFlag byte = buf[2] & FLAG_EVENT
	//if eventFlag == byte(0x00) {
	//	return nil, jerrors.Errorf("event flag:%v", eventFlag)
	//}
	//var twoWayFlag byte = buf[2] & FLAG_TWOWAY
	//if twoWayFlag == byte(0x00) {
	//	return nil, jerrors.Errorf("twoway flag:%v", twoWayFlag)
	//}
	var rspFlag = buf[2] & FLAG_REQUEST
	if rspFlag != byte(0x00) {
		return nil, jerrors.Errorf("response flag:%v", rspFlag)
	}

	// Header{status}
	if buf[3] != Response_OK {
		return nil, jerrors.Errorf("Response not OK, java exception:%s", string(buf[18:length-1]))
	}

	// Header{req id}
	//var ID int64 = int64(binary.BigEndian.Uint64(buf[4:]))
	//fmt.Printf("response package id:%#X\n", ID)

	// Header{body len}
	var bodyLen = int32(binary.BigEndian.Uint32(buf[12:]))
	if int(bodyLen+HEADER_LENGTH) != length {
		return nil, ErrIllegalPackage
	}

	// body
	decoder := NewDecoder(buf[16:length])
	rspObj, _ := decoder.Decode()
	switch rspObj {
	case RESPONSE_WITH_EXCEPTION:
		return decoder.Decode()
	case RESPONSE_VALUE:
		return decoder.Decode()
	case RESPONSE_NULL_VALUE:
		return nil, jerrors.New("Received null")
	}

	return nil, nil
}

func cpSlice(in, out interface{}) {
	inValue := reflect.ValueOf(in)
	if inValue.IsNil() {
		panic("@in is nil")
		return
	}

	outValue := reflect.ValueOf(out)
	if outValue.Kind() == reflect.Ptr {
		outValue = outValue.Elem()
	}

	outValue.Set(reflect.MakeSlice(outValue.Type(), inValue.Len(), inValue.Len()))
	outElemKind := outValue.Type().Elem().Kind()
	for i := 0; i < outValue.Len(); i++ {
		valueI := inValue.Index(i)
		if outValue.Index(i).Kind() == reflect.Struct {
			if valueI.Kind() == reflect.Ptr && valueI.Kind() != outElemKind {
				valueI = valueI.Elem()
			}
			valueI = valueI.Interface().(reflect.Value)
		}
		if valueI.Type().AssignableTo(outValue.Index(i).Type()) {
			outValue.Index(i).Set(valueI)
		} else {
			panic(jerrors.Errorf("in element type %s can not assign to out element type %s",
				valueI.Type().Name(), outValue.Type().Name()))
		}
	}
}

func cpMap(in, out interface{}) {
	inValue := reflect.ValueOf(in)
	if inValue.IsNil() {
		panic("@in is nil")
		return
	}

	outValue := reflect.ValueOf(out)
	if outValue.Kind() == reflect.Ptr {
		outValue = outValue.Elem()
	}

	outValue.Set(reflect.MakeMap(outValue.Type()))
	// TODO: get map key kind
	// outKeyKind := outValue.Type().Elem().Kind()
	// TODO: get map value kind
	// outValueKind := outValue.Type().Elem().Kind()
	for _, key := range inValue.MapKeys() {
		mapKey := inValue.MapIndex(key)
		mapValue := inValue.MapIndex(mapKey)
		if mapKey.Kind() == reflect.Ptr {
			mapKey = mapKey.Elem()
		}
		if mapKey.Kind() == reflect.Struct {
			mapKey = mapKey.Interface().(reflect.Value)
		}

		if mapValue.Kind() == reflect.Ptr {
			mapValue = mapValue.Elem()
		}
		if mapValue.Kind() == reflect.Struct {
			mapValue = mapValue.Interface().(reflect.Value)
		}

		outValue.SetMapIndex(mapKey, mapValue)
	}
}

// reflect return value
func ReflectResponse(in interface{}, out interface{}) error {
	if in == nil {
		return jerrors.Errorf("@in is nil")
	}

	if out == nil {
		return jerrors.Errorf("@out is nil")
	}
	if reflect.TypeOf(out).Kind() != reflect.Ptr {
		return jerrors.Errorf("@out should be a pointer")
	}

	inType := reflect.TypeOf(in)
	switch inType.Kind() {
	case reflect.Bool:
		reflect.ValueOf(out).Elem().Set(reflect.ValueOf(in.(bool)))
	case reflect.Int8:
		reflect.ValueOf(out).Elem().Set(reflect.ValueOf(in.(int8)))
	case reflect.Int16:
		reflect.ValueOf(out).Elem().Set(reflect.ValueOf(in.(int16)))
	case reflect.Int32:
		reflect.ValueOf(out).Elem().Set(reflect.ValueOf(in.(int32)))
	case reflect.Int64:
		reflect.ValueOf(out).Elem().Set(reflect.ValueOf(in.(int64)))
	case reflect.Float32:
		reflect.ValueOf(out).Elem().Set(reflect.ValueOf(in.(float32)))
	case reflect.Float64:
		reflect.ValueOf(out).Elem().Set(reflect.ValueOf(in.(float64)))
	case reflect.String:
		reflect.ValueOf(out).Elem().Set(reflect.ValueOf(in.(string)))
	case reflect.Ptr:
		reflect.ValueOf(out).Elem().Set(reflect.ValueOf(in.(reflect.Value).Elem().Interface()))
	case reflect.Struct:
		reflect.ValueOf(out).Elem().Set(in.(reflect.Value)) // reflect.ValueOf(in.(reflect.Value)))
	case reflect.Slice, reflect.Array:
		cpSlice(in, out)
	case reflect.Map:
		cpMap(in, out)
	}

	return nil
}
