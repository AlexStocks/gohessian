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
	"math"
	"reflect"
	"time"
	"unicode/utf8"
)

import (
	"github.com/AlexStocks/goext/strings"
	log "github.com/AlexStocks/log4go"
)

// nil bool int8 int32 int64 float64 time.Time
// string []byte []interface{} map[interface{}]interface{}
// array object struct

const (
	ENCODER_DEBUG = false
)

// If @v can not be encoded, the return value is nil. At present only struct may can not be encoded.
func Encode(v interface{}, b []byte) []byte {
	if v == nil {
		return encNull(b)
	}
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
		b = encDateInMs(v.(time.Time), b)

	case float32:
		b = encFloat(float64(v.(float32)), b)

	case float64:
		b = encFloat(v.(float64), b)

	case string:
		b = encString(v.(string), b)

	case []byte:
		b = encBinary(v.([]byte), b)

	case map[interface{}]interface{}:
		b = encUntypedMap(v.(map[interface{}]interface{}), b)

	default:
		t := reflect.TypeOf(v)
		if reflect.Ptr == t.Kind() {
			// tmp := reflect.ValueOf(v).Elem()
			// t = reflect.TypeOf(tmp)
			t = reflect.TypeOf(reflect.ValueOf(v).Elem())
		}
		switch t.Kind() {
		case reflect.Struct:
			if p, ok := v.(POJO); ok {
				b = encStruct(p, b)
			} else {
				log.Warn("struct type not Support! %s is not a instance of POJO", t.Kind().String())
				panic("unknow struct type, not instance of POJO")
			}
		case reflect.Slice, reflect.Array:
			b = encUntypedList(v, b)
		case reflect.Map: // 进入这个case，就说明map可能是map[string]int这种类型
			b = encMap(v, b)
		default:
			log.Warn("type not Support! %s", t.Kind().String())
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

func encBT(b []byte, t ...byte) []byte {
	return append(b, t...)
}

/////////////////////////////////////////
// Null
/////////////////////////////////////////
func encNull(b []byte) []byte {
	return append(b, 'N')
}

/////////////////////////////////////////
// Bool
/////////////////////////////////////////

// # boolean true/false
// ::= 'T'
// ::= 'F'
func encBool(v bool, b []byte) []byte {
	var c byte = BC_FALSE
	if v == true {
		c = BC_TRUE
	}

	return append(b, c)
}

/////////////////////////////////////////
// Int32
/////////////////////////////////////////

// # 32-bit signed integer
// ::= 'I' b3 b2 b1 b0
// ::= [x80-xbf]             # -x10 to x3f
// ::= [xc0-xcf] b0          # -x800 to x7ff
// ::= [xd0-xd7] b1 b0       # -x40000 to x3ffff
func encInt32(v int32, b []byte) []byte {
	if int32(INT_DIRECT_MIN) <= v && v <= int32(INT_DIRECT_MAX) {
		return encBT(b, byte(v+int32(BC_INT_ZERO)))
	} else if int32(INT_BYTE_MIN) <= v && v <= int32(INT_BYTE_MAX) {
		return encBT(b, byte(int32(BC_INT_BYTE_ZERO)+v>>8), byte(v))
	} else if int32(INT_SHORT_MIN) <= v && v <= int32(INT_SHORT_MAX) {
		return encBT(b, byte(v>>16+int32(BC_INT_SHORT_ZERO)), byte(v>>8), byte(v))
	}

	return encBT(b, byte('I'), byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

/////////////////////////////////////////
// Int64
/////////////////////////////////////////

// # 64-bit signed long integer
// ::= 'L' b7 b6 b5 b4 b3 b2 b1 b0
// ::= [xd8-xef]             # -x08 to x0f
// ::= [xf0-xff] b0          # -x800 to x7ff
// ::= [x38-x3f] b1 b0       # -x40000 to x3ffff
// ::= x59 b3 b2 b1 b0       # 32-bit integer cast to long
func encInt64(v int64, b []byte) []byte {
	if int64(LONG_DIRECT_MIN) <= v && v <= int64(LONG_DIRECT_MAX) {
		return encBT(b, byte(v-int64(BC_LONG_ZERO)))
	} else if int64(LONG_BYTE_MIN) <= v && v <= int64(LONG_BYTE_MAX) {
		return encBT(b, byte(int64(BC_LONG_BYTE_ZERO)+(v>>8)), byte(v))
	} else if int64(LONG_SHORT_MIN) <= v && v <= int64(LONG_SHORT_MAX) {
		return encBT(b, byte(int64(BC_LONG_SHORT_ZERO)+(v>>16)), byte(v>>8), byte(v))
	} else if 0x80000000 <= v && v <= 0x7fffffff {
		return encBT(b, BC_LONG_INT, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
	}

	return encBT(b, 'L', byte(v>>56), byte(v>>48), byte(v>>40), byte(v>>32), byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

/////////////////////////////////////////
// Date
/////////////////////////////////////////

// # time in UTC encoded as 64-bit long milliseconds since epoch
// ::= x4a b7 b6 b5 b4 b3 b2 b1 b0
// ::= x4b b3 b2 b1 b0       # minutes since epoch
func encDateInMs(v time.Time, b []byte) []byte {
	b = append(b, BC_DATE)
	return append(b, PackInt64(v.UnixNano()/1e6)...)
}

func encDateInMimute(v time.Time, b []byte) []byte {
	b = append(b, BC_DATE_MINUTE)
	return append(b, PackInt32(int32(v.UnixNano()/60e9))...)
}

/////////////////////////////////////////
// Double
/////////////////////////////////////////

// # 64-bit IEEE double
// ::= 'D' b7 b6 b5 b4 b3 b2 b1 b0
// ::= x5b                   # 0.0
// ::= x5c                   # 1.0
// ::= x5d b0                # byte cast to double (-128.0 to 127.0)
// ::= x5e b1 b0             # short cast to double
// ::= x5f b3 b2 b1 b0       # 32-bit float cast to double
func encFloat(v float64, b []byte) []byte {
	fv := float64(int64(v))
	if fv == v {
		iv := int64(v)
		switch iv {
		case 0:
			return encBT(b, BC_DOUBLE_ZERO)
		case 1:
			return encBT(b, BC_DOUBLE_ONE)
		}
		if iv >= -0x80 && iv < 0x80 {
			return encBT(b, BC_DOUBLE_BYTE, byte(iv))
		} else if iv >= -0x8000 && iv < 0x8000 {
			return encBT(b, BC_DOUBLE_BYTE, byte(iv>>8), byte(iv))
		}

		goto END
	}

END:
	bits := uint64(math.Float64bits(v))
	return encBT(b, BC_DOUBLE, byte(bits>>56), byte(bits>>48), byte(bits>>40),
		byte(bits>>32), byte(bits>>24), byte(bits>>16), byte(bits>>8), byte(bits))
}

/////////////////////////////////////////
// String
/////////////////////////////////////////

// # UTF-8 encoded character string split into 64k chunks
// ::= x52 b1 b0 <utf8-data> string  # non-final chunk
// ::= 'S' b1 b0 <utf8-data>         # string of length 0-65535
// ::= [x00-x1f] <utf8-data>         # string of length 0-31
// ::= [x30-x34] <utf8-data>         # string of length 0-1023
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
		return encBT(b, BC_STRING_DIRECT)
	}

	for {
		vLen = utf8.RuneCount(vBuf.Bytes())
		if vLen == 0 {
			break
		}
		if vLen > CHUNK_SIZE {
			b = encBT(b, BC_STRING_CHUNK)
			b = encBT(b, PackUint16(uint16(CHUNK_SIZE))...)
			vChunk(CHUNK_SIZE)
		} else {
			if vLen <= int(STRING_DIRECT_MAX) {
				b = encBT(b, byte(vLen+int(BC_STRING_DIRECT)))
			} else if vLen <= int(STRING_SHORT_MAX) {
				b = encBT(b, byte((vLen>>8)+int(BC_STRING_SHORT)), byte(vLen))
			} else {
				b = encBT(b, BC_STRING)
				b = encBT(b, PackUint16(uint16(vLen))...)
			}
			vChunk(vLen)
		}
	}

	return b
}

/////////////////////////////////////////
// Binary, []byte
/////////////////////////////////////////

// # 8-bit binary data split into 64k chunks
// ::= x41 b1 b0 <binary-data> binary # non-final chunk
// ::= 'B' b1 b0 <binary-data>        # final chunk
// ::= [x20-x2f] <binary-data>        # binary data of length 0-15
// ::= [x34-x37] <binary-data>        # binary data of length 0-1023
func encBinary(v []byte, b []byte) []byte {
	var (
		length  uint16
		vLength int
	)

	if len(v) == 0 {
		return encBT(b, BC_BINARY_DIRECT)
	}

	vLength = len(v)
	for vLength > 0 {
		// if vBuf.Len() > CHUNK_SIZE {
		if vLength > CHUNK_SIZE {
			length = CHUNK_SIZE
			b = encBT(b, byte(BC_BINARY_CHUNK), byte(length>>8), byte(length))
		} else {
			length = uint16(vLength)
			if vLength <= int(BINARY_DIRECT_MAX) {
				b = encBT(b, byte(int(BC_BINARY_DIRECT)+vLength))
			} else if vLength <= int(BINARY_SHORT_MAX) {
				b = encBT(b, byte(int(BC_BINARY_SHORT)+vLength>>8), byte(vLength))
			} else {
				b = encBT(b, byte(BC_BINARY), byte(vLength>>8), byte(vLength))
			}
		}

		b = append(b, v[:length]...)
		v = v[length:]
		vLength = len(v)
	}

	return b
}

/////////////////////////////////////////
// List
/////////////////////////////////////////

// # list/vector
// ::= x55 type value* 'Z'   # variable-length list
// ::= 'V' type int value*   # fixed-length list
// ::= x57 value* 'Z'        # variable-length untyped list
// ::= x58 int value*        # fixed-length untyped list
// ::= [x70-77] type value*  # fixed-length typed list
// ::= [x78-7f] value*       # fixed-length untyped list
func encUntypedList(v interface{}, b []byte) []byte {
	reflectValue := reflect.ValueOf(v)
	b = encBT(b, BC_LIST_FIXED_UNTYPED) // x58
	b = encInt32(int32(reflectValue.Len()), b)
	for i := 0; i < reflectValue.Len(); i++ {
		b = Encode(reflectValue.Index(i).Interface(), b)
	}

	return b
}

/////////////////////////////////////////
// map/object
/////////////////////////////////////////

// ::= 'M' type (value value)* 'Z'  # key, value map pairs
// ::= 'H' (value value)* 'Z'       # untyped key, value
func encUntypedMap(m map[interface{}]interface{}, b []byte) []byte {
	if len(m) == 0 {
		return b
	}

	b = encBT(b, BC_MAP_UNTYPED)

	for k, v := range m {
		b = Encode(k, b)
		b = Encode(v, b)
	}

	b = encBT(b, BC_END) // 'Z'

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

	// return nil
	return newCodecError("unsuport key kind " + typ.Kind().String())
}

func encMap(m interface{}, b []byte) []byte {
	var (
		buf   []byte // 如果map encode失败，也不会影响b中已有的内容
		typ   reflect.Type
		value reflect.Value
		keys  []reflect.Value
	)

	// buf = append(buf, 'M')
	buf = encBT(buf, BC_MAP_UNTYPED)
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
	buf = append(buf, BC_END)

	return append(b, buf...)
}

/////////////////////////////////////////
// map/object
/////////////////////////////////////////

//class-def  ::= 'C' string int string* //  mandatory type string, the number of fields, and the field names.
//object     ::= 'O' int value* // class-def id, value list
//           ::= [x60-x6f] value* // class-def id, value list
//
//Object serialization
//
//class Car {
//  String color;
//  String model;
//}
//
//out.writeObject(new Car("red", "corvette"));
//out.writeObject(new Car("green", "civic"));
//
//---
//
//C                        # object definition (#0)
//  x0b example.Car        # type is example.Car
//  x92                    # two fields
//  x05 color              # color field name
//  x05 model              # model field name
//
//O                        # object def (long form)
//  x90                    # object definition #0
//  x03 red                # color field value
//  x08 corvette           # model field value
//
//x60                      # object def #0 (short form)
//  x05 green              # color field value
//  x05 civic              # model field value
//
//enum Color {
//  RED,
//  GREEN,
//  BLUE,
//}
//
//out.writeObject(Color.RED);
//out.writeObject(Color.GREEN);
//out.writeObject(Color.BLUE);
//out.writeObject(Color.GREEN);
//
//---
//
//C                         # class definition #0
//  x0b example.Color       # type is example.Color
//  x91                     # one field
//  x04 name                # enumeration field is "name"
//
//x60                       # object #0 (class def #0)
//  x03 RED                 # RED value
//
//x60                       # object #1 (class def #0)
//  x90                     # object definition ref #0
//  x05 GREEN               # GREEN value
//
//x60                       # object #2 (class def #0)
//  x04 BLUE                # BLUE value
//
//x51 x91                   # object ref #1, i.e. Color.GREEN
func encStruct(v POJO, b []byte) []byte {
	vv := reflect.ValueOf(v)

	// write object definition
	l, ok := checkPOJORegistry(typeof(v))
	if !ok { // 不存在
		l = RegisterPOJO(v)
	}

	// write object instance
	if byte(l) <= OBJECT_DIRECT_MAX {
		b = encBT(b, byte(l)+BC_OBJECT_DIRECT)
	} else {
		b = encBT(b, BC_OBJECT)
		b = encInt32(int32(l), b)
	}
	num := vv.NumField()
	for i := 0; i < num; i++ {
		b = Encode(vv.Field(i).Interface(), b)
	}

	return b
}
