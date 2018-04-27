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
	"errors"
	"fmt"
	"reflect"
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
	ErrIllegalPackage = fmt.Errorf("illegal pacakge!")
)

// hessian decode respone
func UnpackResponse(buf []byte) (interface{}, error) {
	length := len(buf)
	if length < HEADER_LENGTH || (buf[0] != byte(MAGIC_HIGH) && buf[1] != byte(MAGIC_LOW)) {
		return nil, ErrIllegalPackage
	}

	// Header{serialization id(5 bit), event, two way, req/response}
	var serialID byte = buf[2] & SERIALIZATION_MASK
	if serialID == byte(0x00) {
		return nil, fmt.Errorf("serialization ID:%v", serialID)
	}
	//var eventFlag byte = buf[2] & FLAG_EVENT
	//if eventFlag == byte(0x00) {
	//	return nil, fmt.Errorf("event flag:%v", eventFlag)
	//}
	//var twoWayFlag byte = buf[2] & FLAG_TWOWAY
	//if twoWayFlag == byte(0x00) {
	//	return nil, fmt.Errorf("twoway flag:%v", twoWayFlag)
	//}
	var rspFlag byte = buf[2] & FLAG_REQUEST
	if rspFlag != byte(0x00) {
		return nil, fmt.Errorf("response flag:%v", rspFlag)
	}

	// Header{status}
	if buf[3] != Response_OK {
		return nil, fmt.Errorf("Response not OK, java exception:%s", string(buf[18:length-1]))
	}

	// Header{req id}
	//var ID int64 = int64(binary.BigEndian.Uint64(buf[4:]))
	//fmt.Printf("response package id:%#X\n", ID)

	// Header{body len}
	var bodyLen int32 = int32(binary.BigEndian.Uint32(buf[12:]))
	//fmt.Printf("response package body length:%d\n", bodyLen)
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
		return nil, errors.New("Received null")
	}

	return nil, nil
}

// reflect return value
func ReflectResponse(rsp interface{}, rspType reflect.Type) interface{} {
	if rsp == nil {
		return nil
	}

	switch rspType.Kind() {
	case reflect.Bool:
		return rsp.(bool)
	case reflect.Int8:
		return rsp.(int8)
	case reflect.Int16:
		return rsp.(int16)
	case reflect.Int32:
		return rsp.(int32)
	case reflect.Int64:
		return rsp.(int64)
	//case reflect.Uint8: // of no usage, just list here
	//	return rsp.(uint8)
	case reflect.Uint16:
		return rsp.(uint16)
	//case reflect.Uint32:
	//	return rsp.(uint32)
	//case reflect.Uint64:
	//	return rsp.(uint64)
	case reflect.Float32:
		return rsp.(float32)
	case reflect.Float64:
		return rsp.(float64)
	case reflect.String:
		return rsp.(string)
	case reflect.Ptr:
		return rsp.(reflect.Value).Elem().Interface()
	case reflect.Slice, reflect.Array:
		array := rsp.([]interface{})
		//retArray := reflect.MakeSlice(reflect.SliceOf(rspType), len(ind), len(ind)).Interface().([][]interface{})
		//for i:=0;i<len(ind);i++{
		//	retArray[0][i]=ind[i].(reflect.Value).Elem().Interface()
		//}
		var retArray []interface{} = make([]interface{}, len(array))
		for i := 0; i < len(array); i++ {
			retArray[i] = array[i].(reflect.Value).Elem().Interface()
		}
		return retArray
	case reflect.Map:
		m := rsp.(map[interface{}]interface{})
		var retMap map[interface{}]interface{} = make(map[interface{}]interface{}, len(m))
		for k, v := range m {
			vv := v.(reflect.Value).Elem().Interface()
			retMap[k] = vv
		}
		return retMap
	case reflect.Struct:
		return rsp.(reflect.Value).Elem().Interface()
	}

	return nil
}
