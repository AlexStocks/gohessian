/******************************************************
# DESC    : hessian decode
# AUTHOR  : Alex Stocks
# EMAIL   : alexstocks@foxmail.com
# MOD     : 2016-10-22 20:25
# FILE    : decode.go
******************************************************/

package hessian

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"
	"time"
)

type Decoder struct {
	reader *bufio.Reader
	refs   []interface{}
}

var (
	ErrNotEnoughBuf    = fmt.Errorf("not enough buf")
	ErrIllegalRefIndex = fmt.Errorf("illegal ref index")
)

func NewDecoder(b []byte) *Decoder {
	return &Decoder{reader: bufio.NewReader(bytes.NewReader(b))}
}

/////////////////////////////////////////
// utilities
/////////////////////////////////////////

//读取当前字节,指针不前移
func (d *Decoder) peekByte() byte {
	return d.peek(1)[0]
}

//添加引用
func (d *Decoder) appendRefs(v interface{}) {
	d.refs = append(d.refs, v)
}

//获取缓冲长度
func (d *Decoder) len() int {
	d.peek(1) //需要先读一下资源才能得到已缓冲的长度
	return d.reader.Buffered()
}

//读取 Decoder 结构中的一个字节,并后移一个字节
func (d *Decoder) readByte() (byte, error) {
	return d.reader.ReadByte()
}

//读取指定长度的字节,并后移len(b)个字节
func (d *Decoder) next(b []byte) (int, error) {
	return d.reader.Read(b)
}

//读取指定长度字节,指针不后移
// func (d *Decoder) peek(n int) ([]byte, error) {
func (d *Decoder) peek(n int) []byte {
	// return d.reader.Peek(n)
	b, _ := d.reader.Peek(n)
	return b
}

//读取len(s)的 utf8 字符
func (d *Decoder) nextRune(s []rune) []rune {
	var (
		n  int
		i  int
		r  rune
		ri int
		e  error
	)

	n = len(s)
	s = s[:0]
	for i = 0; i < n; i++ {
		if r, ri, e = d.reader.ReadRune(); e == nil && ri > 0 {
			s = append(s, r)
		}
	}

	return s
}

//读取数据类型描述,用于 list 和 map
func (d *Decoder) decType() (string, error) {
	var (
		err error
		arr [1]byte
		buf []byte
		tag byte
		idx int32
	)

	buf = arr[:1]
	if _, err = io.ReadFull(d.reader, buf); err != nil {
		return "", newCodecError("decType reading tag", err)
	}
	tag = buf[0]
	if (tag >= BC_STRING_DIRECT && tag <= STRING_DIRECT_MAX) ||
		(tag >= 0x30 && tag <= 0x33) || (tag == BC_STRING) || (tag == BC_STRING_CHUNK) {
		return d.decString(int32(tag))
	}

	if idx, err = d.decInt32(TAG_READ); err != nil {
		return "", newCodecError("decType reading tag", err)
	}

	return getGoNameByIndex(int(idx))
}

//解析 hessian 数据包
func (d *Decoder) Decode() (interface{}, error) {
	var (
		err error
		t   byte
	)

	t, err = d.readByte()
	if err == io.EOF {
		return nil, err
	}
	switch {
	case t == BC_END:
		return nil, nil

	case t == BC_NULL: // 'N': //null
		return nil, nil

	case t == BC_TRUE: // 'T': //true
		return true, nil

	case t == BC_FALSE: //'F': //false
		return false, nil

	case (0x80 <= t && t <= 0xbf) || (0xc0 <= t && t <= 0xcf) ||
		(0xd0 <= t && t <= 0xd7) || t == BC_INT: //'I': //int
		return d.decInt32(int32(t))

	case (t >= 0xd8 && t <= 0xef) || (t >= 0xf4 && t <= 0xff) ||
		(t >= 0x38 && t <= 0x3f) || (t == BC_LONG_INT) || (t == BC_LONG): //'L': //long
		return d.decInt64(int32(t))

	case (t == BC_DATE_MINUTE) || (t == BC_DATE): //'d': //date
		return d.decDate(int32(t))

	case (t == BC_DOUBLE_ZERO) || (t == BC_DOUBLE_ONE) || (t == BC_DOUBLE_BYTE) ||
		(t == BC_DOUBLE_SHORT) || (t == BC_DOUBLE_MILL) || (t == BC_DOUBLE): //'D': //double
		return d.decDouble(int32(t))

	// case 'S', 's', 'X', 'x': //string,xml
	case (t == BC_STRING_CHUNK || t == BC_STRING) ||
		(t >= BC_STRING_DIRECT && t <= STRING_DIRECT_MAX) ||
		(t >= 0x30 && t <= 0x33):
		return d.decString(int32(t))

		// case 'B', 'b': //binary
	case (t == BC_BINARY) || (t == BC_BINARY_CHUNK) || (t >= 0x20 && t <= 0x2f):
		return d.decBinary(int32(t))

	// case 'V': //list
	case (t >= BC_LIST_DIRECT && t <= 0x77) || (t == BC_LIST_FIXED || t == BC_LIST_VARIABLE) ||
		(t >= BC_LIST_DIRECT_UNTYPED && t <= 0x7f) ||
		(t == BC_LIST_FIXED_UNTYPED || t == BC_LIST_VARIABLE_UNTYPED):
		return d.decList(int32(t))

	case (t == BC_MAP) || (t == BC_MAP_UNTYPED):
		return d.decMap(int32(t))

	case (t == BC_OBJECT_DEF) || (t == BC_OBJECT):
		return d.decObject(int32(t))

	case (t == BC_REF): // 'R': //ref, 一个整数，用以指代前面的list 或者 map
		return d.decRef(int32(t))

	default:
		return nil, fmt.Errorf("Invalid type: %v,>>%v<<<", string(t), d.peek(d.len()))
	}
}

/////////////////////////////////////////
// Int32
/////////////////////////////////////////

// # 32-bit signed integer
// ::= 'I' b3 b2 b1 b0
// ::= [x80-xbf]             # -x10 to x3f
// ::= [xc0-xcf] b0          # -x800 to x7ff
// ::= [xd0-xd7] b1 b0       # -x40000 to x3ffff
func (d *Decoder) decInt32(flag int32) (int32, error) {
	var (
		err error
		tag byte
		buf [4]byte
	)

	if flag != TAG_READ {
		tag = byte(flag)
	} else {
		tag, _ = d.readByte()
	}

	switch {
	//direct integer
	case tag >= 0x80 && tag <= 0xbf:
		return int32(tag - BC_INT_ZERO), nil

	case tag >= 0xc0 && tag <= 0xcf:
		if _, err = io.ReadFull(d.reader, buf[:1]); err != nil {
			return 0, newCodecError("decInt32 short integer", err)
		}
		return int32(tag-BC_INT_BYTE_ZERO)<<8 + int32(buf[0]), nil

	case tag >= 0xd0 && tag <= 0xd7:
		if _, err = io.ReadFull(d.reader, buf[:2]); err != nil {
			return 0, newCodecError("decInt32 short integer", err)
		}
		return int32(tag-BC_INT_SHORT_ZERO)<<16 + int32(buf[0])<<8 + int32(buf[1]), nil

	case tag == BC_INT:
		if _, err := io.ReadFull(d.reader, buf[:4]); err != nil {
			return 0, newCodecError("decInt32 parse int", err)
		}
		return int32(buf[0])<<24 + int32(buf[1])<<16 + int32(buf[2])<<8 + int32(buf[3]), nil

	default:
		return 0, newCodecError("decInt32 integer wrong tag:" + fmt.Sprintf("%d-%#x", int(tag), tag))
	}
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
func (d *Decoder) decInt64(flag int32) (int64, error) {
	var (
		err error
		tag byte
		buf [8]byte
	)

	if flag != TAG_READ {
		tag = byte(flag)
	} else {
		tag, _ = d.readByte()
	}

	switch {
	case tag >= 0xd8 && tag <= 0xef:
		return int64(tag - BC_LONG_ZERO), nil

	case tag >= 0xf4 && tag <= 0xff:
		if _, err = io.ReadFull(d.reader, buf[:1]); err != nil {
			return 0, newCodecError("decInt64 short integer", err)
		}
		return int64(tag-BC_LONG_BYTE_ZERO)<<8 + int64(buf[0]), nil

	case tag >= 0x38 && tag <= 0x3f:
		if _, err := io.ReadFull(d.reader, buf[:2]); err != nil {
			return 0, newCodecError("decInt64 short integer", err)
		}
		return int64(tag-BC_LONG_SHORT_ZERO)<<16 + int64(buf[0])<<8 + int64(buf[1]), nil
		// return int64(tag-BC_LONG_SHORT_ZERO)<<16 + int64(buf[0])*256 + int64(buf[1]), nil

	case tag == BC_LONG:
		if _, err := io.ReadFull(d.reader, buf[:8]); err != nil {
			return 0, newCodecError("decInt64 parse long", err)
		}
		return int64(buf[0])<<56 + int64(buf[1])<<48 + int64(buf[2])<<40 + int64(buf[3])<<32 +
			int64(buf[4])<<24 + int64(buf[5])<<16 + int64(buf[6])<<8 + int64(buf[7]), nil

	default:
		return 0, newCodecError("decInt64 long wrong tag " + fmt.Sprintf("%d-%#x", int(tag), tag))
	}
}

/////////////////////////////////////////
// Date
/////////////////////////////////////////

// # time in UTC encoded as 64-bit long milliseconds since epoch
// ::= x4a b7 b6 b5 b4 b3 b2 b1 b0
// ::= x4b b3 b2 b1 b0       # minutes since epoch
func (d *Decoder) decDate(flag int32) (time.Time, error) {
	var (
		err error
		l   int
		tag byte
		buf [8]byte
		s   []byte
		i64 int64
		t   time.Time
	)

	if flag != TAG_READ {
		tag = byte(flag)
	} else {
		tag, _ = d.readByte()
	}

	switch {
	case tag == BC_DATE: //'d': //date
		s = buf[:8]
		l, err = d.next(s)
		if err != nil {
			return t, err
		}
		if l != 8 {
			return t, ErrNotEnoughBuf
		}
		i64 = UnpackInt64(s)
		// return time.Unix(i64/1000, i64%1000*10e5), nil
		return time.Unix(i64/1000, i64*100), nil

	case tag == BC_DATE_MINUTE:
		s = buf[:4]
		l, err = d.next(s)
		if err != nil {
			return t, err
		}
		if l != 4 {
			return t, ErrNotEnoughBuf
		}
		i64 = int64(UnpackInt32(s))
		return time.Unix(i64*60, 0), nil

	default:
		return t, fmt.Errorf("decDate Invalid type: %v", tag)
	}
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
func (d *Decoder) decDouble(flag int32) (interface{}, error) {
	var (
		err error
		tag byte
		buf [8]byte
	)

	if flag != TAG_READ {
		tag = byte(flag)
	} else {
		tag, _ = d.readByte()
	}
	switch tag {
	case BC_LONG_INT:
		return d.decInt32(TAG_READ)

	case BC_DOUBLE_ZERO:
		return float64(0), nil

	case BC_DOUBLE_ONE:
		return float64(1), nil

	case BC_DOUBLE_BYTE:
		tag, _ = d.readByte()
		return float64(tag), nil

	case BC_DOUBLE_SHORT:
		if _, err = io.ReadFull(d.reader, buf[:2]); err != nil {
			return nil, newCodecError("decDouble short integer", err)
		}

		return float64(int(buf[0])<<8 + int(buf[1])), nil

	case BC_DOUBLE_MILL:
		i, _ := d.decInt32(TAG_READ)
		return float64(i), nil

	case BC_DOUBLE:
		if _, err = io.ReadFull(d.reader, buf[:8]); err != nil {
			return nil, newCodecError("decDouble short integer", err)
		}

		bits := binary.BigEndian.Uint64(buf[:8])
		datum := math.Float64frombits(bits)
		return datum, nil
	}

	return nil, newCodecError("decDouble parse double wrong tag:" + fmt.Sprintf("%d-%#x", int(tag), tag))
}

/////////////////////////////////////////
// String
/////////////////////////////////////////

// # UTF-8 encoded character string split into 64k chunks
// ::= x52 b1 b0 <utf8-data> string  # non-final chunk
// ::= 'S' b1 b0 <utf8-data>         # string of length 0-65535
// ::= [x00-x1f] <utf8-data>         # string of length 0-31
// ::= [x30-x34] <utf8-data>         # string of length 0-1023
func (d *Decoder) getStrLen(tag byte) (int32, error) {
	var (
		err    error
		buf    [2]byte
		length int32
	)

	switch {
	case tag >= BC_STRING_DIRECT && tag <= STRING_DIRECT_MAX:
		return int32(tag - 0x00), nil

	case tag >= 0x30 && tag <= 0x33:
		_, err = io.ReadFull(d.reader, buf[:1])
		if err != nil {
			return -1, newCodecError("getStrLen byte4 integer", err)
		}

		length = int32(tag-0x30)<<8 + int32(buf[0])
		return length, nil

	case tag == BC_STRING_CHUNK || tag == BC_STRING:
		_, err = io.ReadFull(d.reader, buf[:2])
		if err != nil {
			return -1, newCodecError("getStrLen byte5 integer", err)
		}
		length = int32(buf[0])<<8 + int32(buf[1])
		return length, nil

	default:
		return -1, newCodecError("getStrLen getStrLen")
	}
}

func getRune(reader io.Reader) (rune, int, error) {
	var (
		runeNil rune
		typ     reflect.Type
	)

	typ = reflect.TypeOf(reader.(interface{}))

	if (typ == reflect.TypeOf(&bufio.Reader{})) {
		byteReader := reader.(interface{}).(*bufio.Reader)
		return byteReader.ReadRune()
	}

	if (typ == reflect.TypeOf(&bytes.Buffer{})) {
		byteReader := reader.(interface{}).(*bytes.Buffer)
		return byteReader.ReadRune()
	}

	if (typ == reflect.TypeOf(&bytes.Reader{})) {
		byteReader := reader.(interface{}).(*bytes.Reader)
		return byteReader.ReadRune()
	}

	return runeNil, 0, nil
}

func (d *Decoder) decString(flag int32) (string, error) {
	var (
		tag    byte
		length int32
		last   bool
		s      string
		r      rune
	)

	if flag != TAG_READ {
		tag = byte(flag)
	} else {
		tag, _ = d.readByte()
	}

	last = true
	if (tag >= BC_STRING_DIRECT && tag <= STRING_DIRECT_MAX) ||
		(tag >= 0x30 && tag <= 0x33) ||
		(tag == BC_STRING_CHUNK || tag == BC_STRING) {

		if tag == BC_STRING_CHUNK {
			last = false
		} else {
			last = true
		}

		l, err := d.getStrLen(tag)
		if err != nil {
			return s, newCodecError("decString->getStrLen", err)
		}
		length = l
		runeDate := make([]rune, length)
		for i := 0; ; {
			if int32(i) == length {
				if last {
					return string(runeDate), nil
				}

				b, _ := d.readByte()
				switch {
				case (tag >= BC_STRING_DIRECT && tag <= STRING_DIRECT_MAX) ||
					(tag >= 0x30 && tag <= 0x33) ||
					(tag == BC_STRING_CHUNK || tag == BC_STRING):

					if b == BC_STRING_CHUNK {
						last = false
					} else {
						last = true
					}

					l, err := d.getStrLen(b)
					if err != nil {
						return s, newCodecError("decString->getStrLen", err)
					}
					length += l
					bs := make([]rune, length)
					copy(bs, runeDate)
					runeDate = bs

				default:
					return s, newCodecError("decString tag error ", err)
				}

			} else {
				r, _, err = d.reader.ReadRune()
				if err != nil {
					return s, newCodecError("decString->ReadRune", err)
				}
				runeDate[i] = r
				i++
			}
		}

		return string(runeDate), nil
	} else {
		return s, newCodecError("decString byte3 integer")
	}
}

/////////////////////////////////////////
// Binary, []byte
/////////////////////////////////////////

// # 8-bit binary data split into 64k chunks
// ::= x41 b1 b0 <binary-data> binary # non-final chunk
// ::= 'B' b1 b0 <binary-data>        # final chunk
// ::= [x20-x2f] <binary-data>        # binary data of length 0-15
// ::= [x34-x37] <binary-data>        # binary data of length 0-1023
func (d *Decoder) getBinaryLength(tag byte) (int, error) {
	var (
		err error
		buf [2]byte
	)

	if tag >= BC_BINARY_DIRECT && tag <= INT_DIRECT_MAX {
		return int(tag - BC_BINARY_DIRECT), nil
	}

	if _, err = io.ReadFull(d.reader, buf[:2]); err != nil {
		return 0, newCodecError("getBinaryLength parse binary", err)
	}

	return int(buf[0]<<8 + buf[1]), nil
}

func (d *Decoder) decBinary(flag int32) ([]byte, error) {
	var (
		tag    byte
		last   bool
		length int32
	)

	if flag != TAG_READ {
		tag = byte(flag)
	} else {
		tag, _ = d.readBufByte()
	}

	last = true
	if (tag >= BC_BINARY_DIRECT && tag <= INT_DIRECT_MAX) ||
		(tag == BC_BINARY) || (tag == BC_BINARY_CHUNK) {
		if tag == BC_BINARY_CHUNK {
			last = false
		} else {
			last = true
		}
		l, err := d.getBinaryLength(tag)
		if err != nil {
			return nil, newCodecError("decBinary->getBinaryLength", err)
		}
		length = int32(l)
		data := make([]byte, length)
		for i := 0; ; {
			if int32(i) == length {
				if last {
					return data, nil
				}

				var buf [1]byte
				_, err := io.ReadFull(d.reader, buf[:1])
				if err != nil {
					return nil, newCodecError("decBinary byte1 integer", err)
				}
				b := buf[0]
				switch {
				case b == BC_BINARY_CHUNK || b == BC_BINARY:
					if b == BC_BINARY_CHUNK {
						last = false
					} else {
						last = true
					}
					l, err := d.getStrLen(b)
					if err != nil {
						return nil, newCodecError("decBinary getStrLen", err)
					}
					length += l
					bs := make([]byte, 0, length)
					copy(bs, data)
					data = bs
				default:
					return nil, newCodecError("decBinary tag error ", err)
				}
			} else {
				var buf [1]byte
				_, err := io.ReadFull(d.reader, buf[:1])
				if err != nil {
					return nil, newCodecError("decBinary byte2 integer", err)
				}
				data[i] = buf[0]
				i++
			}
		}

		return data, nil
	} else {
		return nil, newCodecError("decBinary byte3 integer")
	}
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

func (d *Decoder) readBufByte() (byte, error) {
	var (
		err error
		buf [1]byte
	)

	_, err = io.ReadFull(d.reader, buf[:1])
	if err != nil {
		return 0, newCodecError("readBufByte", err)
	}

	return buf[0], nil
}

func (d *Decoder) decSlice(value reflect.Value) (interface{}, error) {
	var (
		i   int
		tag byte
	)

	tag, _ = d.readBufByte()
	if tag >= BC_LIST_DIRECT_UNTYPED && tag <= 0x7f {
		i = int(tag - BC_LIST_DIRECT_UNTYPED)
	} else {
		ii, err := d.decInt32(TAG_READ)
		if err != nil {
			return nil, newCodecError("decSlice->decInt32", err)
		}
		i = int(ii)
	}

	ary := reflect.MakeSlice(value.Type(), i, i)
	for j := 0; j < i; j++ {
		it, err := d.Decode()
		if err != nil {
			return nil, newCodecError("decSlice->ReadList", err)
		}
		ary.Index(j).Set(reflect.ValueOf(it))
	}
	d.readBufByte()
	value.Set(ary)

	return ary, nil
}

//func isBuildInType(typeName string) bool {
//	switch typeName {
//	case ARRAY_STRING:
//		return true
//	case ARRAY_INT:
//		return true
//	case ARRAY_FLOAT:
//		return true
//	case ARRAY_DOUBLE:
//		return true
//	case ARRAY_BOOL:
//		return true
//	case ARRAY_LONG:
//		return true
//	default:
//		return false
//	}
//}

func (d *Decoder) decList(flag int32) (interface{}, error) {
	var (
		tag  byte
		size int
	)

	if flag != TAG_READ {
		tag = byte(flag)
	} else {
		tag, _ = d.readByte()
	}

	switch {
	case (tag >= BC_LIST_DIRECT && tag <= 0x77) || (tag == BC_LIST_FIXED || tag == BC_LIST_VARIABLE):
		// str, err := d.decType()
		// if err != nil {
		// 	return nil, newCodecError("ReadType", err)
		// }
		d.decType() // 忽略
		if tag >= BC_LIST_DIRECT && tag <= 0x77 {
			size = int(tag - BC_LIST_DIRECT)
		} else {
			i32, err := d.decInt32(TAG_READ)
			if err != nil {
				return nil, newCodecError("decList->decInt32", err)
			}
			size = int(i32)
		}
		// bl := isBuildInType(str)
		ary := make([]interface{}, size)
		for j := 0; j < size; j++ {
			it, err := d.Decode()
			if err != nil {
				return nil, newCodecError("decList->Decode", err)
			}
			ary[j] = it
		}

		d.appendRefs(&ary)

		return ary, nil

	case (tag >= BC_LIST_DIRECT_UNTYPED && tag <= 0x7f) || (tag == BC_LIST_FIXED_UNTYPED || tag == BC_LIST_VARIABLE_UNTYPED):
		if tag >= BC_LIST_DIRECT_UNTYPED && tag <= 0x7f {
			size = int(tag - BC_LIST_DIRECT_UNTYPED)
		} else {
			i32, err := d.decInt32(TAG_READ)
			if err != nil {
				return nil, newCodecError("decList->ReadType", err)
			}
			size = int(i32)
		}
		ary := make([]interface{}, size)
		for j := 0; j < size; j++ {
			it, err := d.Decode()
			if err != nil {
				return nil, newCodecError("decList->Decode", err)
			}
			ary[j] = it
		}
		//read the endbyte of list
		d.readBufByte()

		d.appendRefs(&ary)

		return ary, nil

	default:
		return nil, newCodecError("illegal list type tag:", tag)
	}
}

/////////////////////////////////////////
// Map
/////////////////////////////////////////

// ::= 'M' type (value value)* 'Z'  # key, value map pairs
// ::= 'H' (value value)* 'Z'       # untyped key, value
func (d *Decoder) decMapByValue(value reflect.Value) (interface{}, error) {
	var (
		tag byte
	)

	tag, _ = d.readBufByte()
	if tag == BC_MAP {
		d.decString(TAG_READ)
	} else if tag == BC_MAP_UNTYPED {
		//do nothing
	} else {
		return nil, newCodecError("wrong header BC_MAP_UNTYPED")
	}

	m := reflect.MakeMap(value.Type())
	//read key and value
	for {
		key, err := d.Decode()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, newCodecError("decMapByValue->ReadType", err)
			}
		}
		vl, err := d.Decode()
		m.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(vl))
	}
	value.Set(m)

	return m, nil
}

func (d *Decoder) decMap(flag int32) (interface{}, error) {
	var (
		err        error
		tag        byte
		ok         bool
		k          interface{}
		v          interface{}
		t          string
		keyName    string
		methodName string
		key        interface{}
		value      interface{}
		inst       interface{}
		m          map[interface{}]interface{}
		fieldValue reflect.Value
		args       []reflect.Value
	)

	if flag != TAG_READ {
		tag = byte(flag)
	} else {
		tag, _ = d.readByte()
	}

	switch {
	case tag == BC_MAP:
		if t, err = d.decType(); err != nil {
			return nil, err
		}
		if _, ok = checkPOJORegistry(t); ok {
			m = make(map[interface{}]interface{}) // 此处假设了map的定义形式，这是不对的
			// d.decType() // 忽略
			for d.peekByte() != byte('z') {
				k, err = d.Decode()
				if err != nil {
					if err == io.EOF {
						break
					}

					return nil, err
				}
				v, err = d.Decode()
				if err != nil {
					return nil, err
				}
				m[k] = v
			}
			d.readByte()
			d.appendRefs(&m)
			return m, nil
		} else {
			inst = createInstance(t)
			for d.peekByte() != 'z' {
				if key, err = d.Decode(); err != nil {
					return nil, err
				}
				if value, err = d.Decode(); err != nil {
					return nil, err
				}
				//set value of the struct to Zero
				if fieldValue = reflect.ValueOf(value); fieldValue.IsValid() {
					keyName = key.(string)
					if keyName[0] >= 'a' { //convert to Upper
						methodName = "Set" + string(keyName[0]-32) + keyName[1:]
					} else {
						methodName = "Set" + keyName
					}

					args = args[:0]
					args = append(args, fieldValue)
					reflect.ValueOf(inst).MethodByName(methodName).Call(args)
				}
			}
			// v = inst
			d.appendRefs(&inst)
			return inst, nil
		}

	case tag == BC_MAP_UNTYPED:
		m = make(map[interface{}]interface{})
		for d.peekByte() != byte(BC_END) {
			k, err = d.Decode()
			if err != nil {
				if err == io.EOF {
					break
				}

				return nil, err
			}
			v, err = d.Decode()
			if err != nil {
				return nil, err
			}
			m[k] = v
		}
		d.readByte()
		d.appendRefs(&m)
		return m, nil

	default:
		return nil, newCodecError("illegal map type tag:", tag)
	}
}

/////////////////////////////////////////
// Object
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
//
//
//
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

func (d *Decoder) decClassDef() (interface{}, error) {
	var (
		err       error
		clsName   string
		fieldNum  int32
		fieldName string
		fieldList []string
	)

	clsName, err = d.decString(TAG_READ)
	if err != nil {
		return nil, newCodecError("decClassDef->decString", err)
	}
	fieldNum, err = d.decInt32(TAG_READ)
	if err != nil {
		return nil, newCodecError("decClassDef->decInt32", err)
	}
	fieldList = make([]string, fieldNum)
	for i := 0; i < int(fieldNum); i++ {
		fieldName, err = d.decString(TAG_READ)
		if err != nil {
			return nil, newCodecError("decClassDef->decString", err)
		}
		fieldList[i] = fieldName
	}

	return classDef{javaName: clsName, fieldNameList: fieldList}, nil
}

func findField(name string, typ reflect.Type) (int, error) {
	for i := 0; i < typ.NumField(); i++ {
		str := typ.Field(i).Name
		if strings.Compare(str, name) == 0 {
			return i, nil
		}
		str1 := strings.Title(name)
		if strings.Compare(str, str1) == 0 {
			return i, nil
		}
	}
	return 0, newCodecError("findField")
}

func (d *Decoder) decInstance(typ reflect.Type, cls classDef) (interface{}, error) {
	if typ.Kind() != reflect.Struct {
		return nil, newCodecError("wrong type expect Struct but get " + typ.String())
	}
	vv := reflect.New(typ)
	st := reflect.ValueOf(vv.Interface()).Elem()
	for i := 0; i < len(cls.fieldNameList); i++ {
		fldName := cls.fieldNameList[i]
		index, err := findField(fldName, typ)
		if err != nil {
			// Log.Printf("%s is not found, will ski type ->p %v", fldName, typ)
			continue
		}
		fldValue := st.Field(index)
		//fmt.Println("fld", fldName, fldValue, fldValue.Kind())
		if !fldValue.CanSet() {
			return nil, newCodecError("decInstance CanSet false for " + fldName)
		}
		kind := fldValue.Kind()
		// fmt.Println("fld name:", fldName, ", index:", index, ", fld kind:", kind,
		// 	", flag:", fldValue.Type() == reflect.TypeOf(hessianNow))
		switch {
		case kind == reflect.String:
			str, err := d.decString(TAG_READ)
			if err != nil {
				return nil, newCodecError("decInstance->ReadString:"+fldName, err)
			}
			fldValue.SetString(str)

		case kind == reflect.Int32 || kind == reflect.Int || kind == reflect.Int16:
			i, err := d.decInt32(TAG_READ)
			if err != nil {
				return nil, newCodecError("decInstance->ParseInt:"+fldName, err)
			}
			v := int64(i)
			fldValue.SetInt(v)

		case kind == reflect.Int64 || kind == reflect.Uint64:
			i, err := d.decInt64(TAG_READ)
			if err != nil {
				return nil, newCodecError("decInstance->decInt64:"+fldName, err)
			}
			fldValue.SetInt(i)

		case kind == reflect.Bool:
			b, err := d.Decode()
			if err != nil {
				return nil, newCodecError("decInstance->Decode:"+fldName, err)
			}
			fldValue.SetBool(b.(bool))

		case kind == reflect.Float32 || kind == reflect.Float64:
			d, err := d.decDouble(TAG_READ)
			if err != nil {
				return nil, newCodecError("decInstance->decDouble"+fldName, err)
			}
			fldValue.SetFloat(d.(float64))

		case kind == reflect.Map:
			d.decMapByValue(fldValue)

		case kind == reflect.Slice || kind == reflect.Array:
			m, _ := d.Decode()
			v := reflect.ValueOf(m)
			if v.Len() > 0 {
				sl := reflect.MakeSlice(fldValue.Type(), v.Len(), v.Len())
				for i := 0; i < v.Len(); i++ {
					sl.Index(i).Set(reflect.ValueOf(v.Index(i).Interface()))
				}
				fldValue.Set(sl)
			}

		case kind == reflect.Struct:
			s, err := d.Decode()
			if err != nil {
				return nil, newCodecError("decInstance->Decode", err)
			}

			fldValue.Set(reflect.Indirect(s.(reflect.Value)))
		//fmt.Println("s with struct", s)
		default:
		}
	}

	return vv, nil
}

func (d *Decoder) decObject(flag int32) (interface{}, error) {
	var (
		tag byte
	)

	if flag != TAG_READ {
		tag = byte(flag)
	} else {
		tag, _ = d.readByte()
	}

	switch {
	case tag == BC_OBJECT_DEF:
		clsDef, err := d.decClassDef()
		if err != nil {
			return nil, newCodecError("decObject->decClassDef byte double", err)
		}
		clsD, _ := clsDef.(classDef)
		//add to slice
		appendClsDef(clsD)
		return d.Decode()

	case tag == BC_OBJECT:
		idx, _ := d.decInt32(TAG_READ)
		clsName, _ := getGoNameByIndex(int(idx))
		return createInstance(clsName), nil

	default:
		return nil, newCodecError("decObject illegal object type tag:", tag)
	}
}

/////////////////////////////////////////
// Ref
/////////////////////////////////////////

// # value reference (e.g. circular trees and graphs)
// ref        ::= x51 int            # reference to nth map/list/object
func (d *Decoder) decRef(flag int32) (interface{}, error) {
	var (
		err error
		tag byte
		buf [4]byte
		i   int
	)

	if flag != TAG_READ {
		tag = byte(flag)
	} else {
		tag, _ = d.readByte()
	}

	switch {
	case tag == BC_REF:
		i, err = d.next(buf[:4])
		if err != nil {
			return nil, err
		}
		if i != 4 {
			return nil, ErrNotEnoughBuf
		}
		i = int(UnpackInt32(buf[:4])) // ref index

		if len(d.refs) <= i {
			return nil, ErrIllegalRefIndex
		}
		return &d.refs[i], nil

	default:
		return nil, newCodecError("decRef illegal ref type tag:", tag)
	}
}
