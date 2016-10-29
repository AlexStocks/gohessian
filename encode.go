/******************************************************
# DESC    : hessian encode
# AUTHOR  : Alex Stocks
# EMAIL   : alexstocks@foxmail.com
# MOD     : 2016-10-22 20:24
# FILE    : encode.go
******************************************************/

// refers to https://github.com/xjing521/gohessian/blob/master/src/gohessian/encode.go

package hessian

import (
	"bytes"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"
)

import (
	"fmt"
	"github.com/AlexStocks/goext/strings"
	log "github.com/AlexStocks/log4go"
)

// interface{} 的别名
type Any interface{}

/*
nil bool int8 int32 int64 float64 time.Time
string []byte []interface{} map[interface{}]interface{}
array object struct
*/

type Encoder struct {
}

const (
	CHUNK_SIZE    = 0x8000
	ENCODER_DEBUG = false
)

// If @v can not be encoded, the return value is nil. At present only struct may can not be encoded.
func Encode(v interface{}, b []byte) []byte {
	switch v.(type) {
	case nil:
		return encNull(b)

	case bool:
		b = encBool(v.(bool), b)

	case int:
		// if v.(int) >= -2147483648 && v.(int) <= 2147483647 {
		// 	b = encInt32(int32(v.(int)), b)
		// } else {
		// 	b = encInt64(int64(v.(int)), b)
		// }
		// 把int统一按照int64处理，这样才不会导致decode的时候出现" reflect: Call using int32 as type int64 [recovered]"这种panic
		b = encInt64(int64(v.(int)), b)

	case int32:
		b = encInt32(v.(int32), b)

	case int64:
		b = encInt64(v.(int64), b)

	case time.Time:
		b = encDate(v.(time.Time), b)

	case float64:
		b = encFloat(v.(float64), b)

	case string:
		b = encString(v.(string), b)

	case []byte:
		b = encBinary(v.([]byte), b)

	case []Any:
		b = encList(v.([]Any), b)

	case map[Any]Any:
		b = encMap(v.(map[Any]Any), b)

	default:
		t := reflect.TypeOf(v)
		if reflect.Ptr == t.Kind() {
			// tmp := reflect.ValueOf(v).Elem()
			// t = reflect.TypeOf(tmp)
			t = reflect.TypeOf(reflect.ValueOf(v).Elem())
		}
		switch t.Kind() {
		case reflect.Struct:
			b = encStruct(v, b)
		case reflect.Slice, reflect.Array:
			b = encList(v.([]Any), b)
		case reflect.Map: // 进入这个case，就说明map可能是map[string]int这种类型
			// b = encMap(v, b)
			b = encMapByReflect(v, b)
		default:
			log.Debug("type not Support! %s", t.Kind().String())
			panic("unknow type")
		}
	}

	if ENCODER_DEBUG {
		log.Debug(SprintHex(b))
	}

	return b
}

//=====================================
//对各种数据类型的编码
//=====================================

// null
func encNull(b []byte) []byte {
	return append(b, 'N')
}

// boolean
func encBool(v bool, b []byte) []byte {
	var c byte = 'F'
	if v == true {
		c = 'T'
	}

	return append(b, c)
}

// int
func encInt32(v int32, b []byte) []byte {
	b = append(b, 'I')
	// return PackInt32(v, b)
	return append(b, PackInt32(v)...)
}

// long
func encInt64(v int64, b []byte) []byte {
	b = append(b, 'L')
	// return PackInt64(v, b)
	return append(b, PackInt64(v)...)
}

// date
func encDate(v time.Time, b []byte) []byte {
	b = append(b, 'd')
	// return PackInt64(v.UnixNano()/1e6, b)
	return append(b, PackInt64(v.UnixNano()/1e6)...)
}

// double
func encFloat(v float64, b []byte) []byte {
	b = append(b, 'D')
	// return PackFloat64(v, b)
	return append(b, PackFloat64(v)...)
}

// string
func encString(v string, b []byte) []byte {
	var (
		vBuf = *bytes.NewBufferString(v)
		vLen = utf8.RuneCountInString(v)

		vChunk = func(length int) {
			for i := 0; i < length; i++ {
				if r, s, err := vBuf.ReadRune(); s > 0 && err == nil {
					// b = append(b, []byte(string(r))...)
					b = append(b, gxstrings.Slice(string(r))...) // 直接基于r的内存空间把它转换为[]byte
				}
			}
		}
	)

	if v == "" {
		b = append(b, 'S')
		// b = PackUint16(uint16(vLen), b)
		b = append(b, PackUint16(uint16(vLen))...)
		b = append(b, []byte{}...)
		return b
	}

	for {
		vLen = utf8.RuneCount(vBuf.Bytes())
		if vLen == 0 {
			break
		}
		if vLen > CHUNK_SIZE {
			b = append(b, 's')
			// b = PackUint16(uint16(CHUNK_SIZE), b)
			b = append(b, PackUint16(uint16(CHUNK_SIZE))...)
			vChunk(CHUNK_SIZE)
		} else {
			b = append(b, 'S')
			// b = PackUint16(uint16(vLen), b)
			b = append(b, PackUint16(uint16(vLen))...)
			vChunk(vLen)
		}
	}

	return b
}

// binary
func encBinary(v []byte, b []byte) []byte {
	var (
		tag     byte
		length  uint16
		vLength int
	)

	if len(v) == 0 {
		b = append(b, 'B')
		// b = PackUint16(0, b)
		b = append(b, PackUint16(0)...)
		return b
	}

	// vBuf := *bytes.NewBuffer(v)
	// for vBuf.Len() > 0 {
	vLength = len(v)
	for vLength > 0 {
		// if vBuf.Len() > CHUNK_SIZE {
		if len(v) > CHUNK_SIZE {
			tag = 'b'
			length = uint16(CHUNK_SIZE)
		} else {
			tag = 'B'
			// length = uint16(vBuf.Len())
			length = uint16(len(v))
		}

		b = append(b, tag)
		// b = PackUint16(length, b)
		b = append(b, PackUint16(length)...)
		// b = append(b, vBuf.Next(length)...)
		b = append(b, v[:length]...)
		v = v[length:]
		vLength = len(v)
	}

	return b
}

// list
func encList(v []Any, b []byte) []byte {
	b = append(b, 'V')

	b = append(b, 'l')
	// b = PackInt32(int32(len(v)), b)
	b = append(b, PackInt32(int32(len(v)))...)

	for _, a := range v {
		b = Encode(a, b)
	}

	b = append(b, 'z')

	return b
}

// map
func encMap(m map[Any]Any, b []byte) []byte {
	if len(m) == 0 {
		return b
	}

	b = append(b, 'M')

	for k, v := range m {
		b = Encode(k, b)
		b = Encode(v, b)
	}

	b = append(b, 'z')

	return b
}

func buildMapKey(key reflect.Value, typ reflect.Type) interface{} {
	switch typ.Kind() {
	case reflect.String:
		return key.String()
	case reflect.Bool:
		return key.Bool()
	case reflect.Int:
		return int32(key.Int())
	case reflect.Int8:
		return int8(key.Int())
	case reflect.Int16:
	case reflect.Int32:
		return int32(key.Int())
	case reflect.Int64:
		return key.Int()
	case reflect.Uint8:
		return byte(key.Uint())
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return key.Uint()
	}

	return nil
	// return newCodecError("unsuport key kind " + typ.Kind().String())
}

func encMapByReflect(m interface{}, b []byte) []byte {
	var (
		buf   []byte // 如果map encode失败，也不会影响b中已有的内容
		typ   reflect.Type
		value reflect.Value
		keys  []reflect.Value
	)

	buf = append(buf, 'M')
	value = reflect.ValueOf(m)
	typ = reflect.TypeOf(m).Key()
	keys = value.MapKeys()
	if len(keys) == 0 {
		return b
	}
	for i := 0; i < len(keys); i++ {
		k := buildMapKey(keys[i], typ)
		if k == nil {
			return b
		}
		buf = Encode(k, buf)
		buf = Encode(value.MapIndex(keys[i]).Interface(), buf)
	}
	buf = append(buf, 'z')

	return append(b, buf...)
}

// encode struct
// attention list:
// @v should have method "GetType" which return @v struct name
// @v should have method "Get..." to get its member value
func encStruct(v Any, b []byte) []byte {
	var (
		i          int
		l          int
		length     int
		str        string
		buf        *bytes.Buffer
		vT         reflect.Type
		vV         reflect.Value
		methodType reflect.Value
		typeName   reflect.Value
		method     reflect.Method
		rvArray    []reflect.Value
	)

	// check Type exists
	// mast contains Type Field to convert to object
	vV = reflect.ValueOf(v)
	methodType = vV.MethodByName("GetType")
	if !methodType.IsValid() {
		log.Error("Don'T contains GetType !")
		return nil
	}

	b = append(b, 'M')
	//encode type Name
	b = append(b, 't')
	// encode struct name
	typeName = methodType.Call([]reflect.Value{})[0] //call return [string,]
	buf = bytes.NewBufferString(typeName.String())
	length = utf8.RuneCount(buf.Bytes())
	b = append(b, PackUint16(uint16(length))...)
	for i = 0; i < length; i++ {
		if r, s, err := buf.ReadRune(); s > 0 && err == nil {
			// b = append(b, []byte(string(r))...)
			b = append(b, gxstrings.Slice(string(r))...) // 直接基于r的内存空间把它转换为[]byte
		}
	}

	//encode the Fields
	vT = reflect.TypeOf(v)
	for i = 0; i < vT.NumMethod(); i++ {
		method = vT.Method(i)
		if !strings.HasPrefix(method.Name, "Get") {
			continue
		}
		if strings.EqualFold(method.Name, "GetType") {
			continue //jump type Field
		}

		//name change GetXaa to xaa
		if method.Name[3] < 'a' {
			str = string(method.Name[3] + 32)
		} else {
			str = string(method.Name[3])
		}
		// key
		l = len(b)
		b = encString(str+method.Name[4:], b)
		length = len(b)

		// value
		rvArray = vV.Method(i).Call([]reflect.Value{}) //return [] reflect.Value
		b = Encode(rvArray[0].Interface(), b)          //GetXXX returns [string,]
		// 如果值为空就不向b里面填充key了
		if len(b) == length {
			fmt.Printf("key:%s, rvArray:%#v, %v, %v, %v\n", str+method.Name[4:], rvArray, rvArray == nil, len(rvArray), rvArray[0])
			b = b[:l]
			continue
		}
	} //end of for

	return append(b, 'z')
}
