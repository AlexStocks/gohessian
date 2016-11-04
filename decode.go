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
	"time"
)

type Decoder struct {
	reader *bufio.Reader
	// structRegistry map[string]reflect.Type // 注册struct
	// typList        []string
	// refs           []interface{}
	// clsDefList     []classDef // 解析过程中使用
	refs []Any
}

var (
	ErrNotEnoughBuf    = fmt.Errorf("not enough buf")
	ErrIllegalRefIndex = fmt.Errorf("illegal ref index")
)

// func NewDecoder(r io.Reader) *Decoder {
// 	return &Decoder{reader: bufio.NewReader(r)}
// }

// func NewDecoderWithBuf(b []byte) *Decoder {
// 	return NewDecoder(bytes.NewReader(b))
// }

func NewDecoder(b []byte) *Decoder {
	return &Decoder{reader: bufio.NewReader(bytes.NewReader(b))}
}

//读取当前字节,指针不前移
func (this *Decoder) peekByte() byte {
	return this.peek(1)[0]
}

//添加引用
func (this *Decoder) appendRefs(v interface{}) {
	this.refs = append(this.refs, v)
}

//获取缓冲长度
func (this *Decoder) len() int {
	this.peek(1) //需要先读一下资源才能得到已缓冲的长度
	return this.reader.Buffered()
}

//读取 Decoder 结构中的一个字节,并后移一个字节
func (this *Decoder) readByte() (byte, error) {
	return this.reader.ReadByte()
}

//读取指定长度的字节,并后移len(b)个字节
func (this *Decoder) next(b []byte) (int, error) {
	return this.reader.Read(b)
}

//读取指定长度字节,指针不后移
// func (this *Decoder) peek(n int) ([]byte, error) {
func (this *Decoder) peek(n int) []byte {
	// return this.reader.Peek(n)
	b, _ := this.reader.Peek(n)
	return b
}

//读取len(s)的 utf8 字符
func (this *Decoder) nextRune(s []rune) []rune {
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
		if r, ri, e = this.reader.ReadRune(); e == nil && ri > 0 {
			s = append(s, r)
		}
	}

	return s
}

//读取数据类型描述,用于 list 和 map
func (this *Decoder) readType() string {
	if this.peekByte() != byte('t') {
		return ""
	}

	var tLen = UnpackInt16(this.peek(3)[1:3]) // 取类型字符串长度
	var b = make([]rune, 3+tLen)
	return string(this.nextRune(b)[3:]) //取类型名称
}

//解析 hessian 数据包
func (this *Decoder) Decode() (interface{}, error) {
	var (
		err error
		t   byte
		l   int
		a   []byte
		s   []byte
	)

	a = make([]byte, 16)
	t, err = this.readByte()
	if err == io.EOF {
		return nil, err
	}
	switch t {
	case 'N': //null
		return nil, nil

	case 'T': //true
		return true, nil

	case 'F': //false
		return false, nil

	case 'I': //int
		s = a[:4]
		l, err = this.next(s)
		if err != nil {
			return nil, err
		}
		if l != 4 {
			return nil, ErrNotEnoughBuf
		}
		return UnpackInt32(s), nil

	case 'L': //long
		s = a[:8]
		l, err = this.next(s)
		if err != nil {
			return nil, err
		}
		if l != 8 {
			return nil, ErrNotEnoughBuf
		}
		return UnpackInt64(s), nil

	case 'd': //date
		s = a[:8]
		l, err = this.next(s)
		if err != nil {
			return nil, err
		}
		if l != 8 {
			return nil, ErrNotEnoughBuf
		}
		var ms = UnpackInt64(s)
		return time.Unix(ms/1000, ms%1000*10e5), nil

	case 'D': //double
		s = a[:8]
		l, err = this.next(s)
		if err != nil {
			return nil, err
		}
		if l != 8 {
			return nil, ErrNotEnoughBuf
		}
		return UnpackFloat64(s), nil

	case 'S', 's', 'X', 'x': //string,xml
		var (
			rBuf   []rune
			chunks []rune
		)
		rBuf = make([]rune, CHUNK_SIZE)
		for { //避免递归读取 Chunks
			s = a[:2]
			l, err = this.next(s)
			if err != nil {
				return nil, err
			}
			if l != 2 {
				return nil, ErrNotEnoughBuf
			}
			l = int(UnpackInt16(s))
			chunks = append(chunks, this.nextRune(rBuf[:l])...)
			if t == 'S' || t == 'X' {
				break
			}
			if t, err = this.readByte(); err != nil {
				return nil, err
			}
		}
		return string(chunks), nil

	case 'B', 'b': //binary
		var (
			buf    []byte
			chunks []byte //等同于 []uint8,在 反射判断类型的时候，会得到 []uint8
		)
		buf = make([]byte, CHUNK_SIZE)
		for { //避免递归读取 Chunks
			s = a[:2]
			l, err = this.next(s)
			if err != nil {
				return nil, err
			}
			if l != 2 {
				return nil, ErrNotEnoughBuf
			}
			l = int(UnpackInt16(s))
			if l, err = this.next(buf[:l]); err != nil {
				return nil, err
			}
			chunks = append(chunks, buf[:l]...)
			if t == 'B' {
				break
			}
			if t, err = this.readByte(); err != nil {
				return nil, err
			}
		}

		return chunks, nil

	case 'V': //list
		var (
			v      Any
			chunks []Any
		)
		this.readType() // 忽略
		if this.peekByte() == byte('l') {
			this.next(a[:5])
		}
		for this.peekByte() != byte('z') {
			if v, err = this.Decode(); err != nil {
				return nil, err
			} else {
				chunks = append(chunks, v)
			}
		}
		this.readByte()
		this.appendRefs(&chunks)
		return chunks, nil

	case 'M': //map
		var (
			k          Any
			v          Any
			t          string
			keyName    string
			methodName string
			key        interface{}
			value      interface{}
			inst       interface{}
			m          map[Any]Any
			fieldValue reflect.Value
			args       []reflect.Value
		)

		t = this.readType()
		if !checkPOJORegistry(t) {
			m = make(map[Any]Any) // 此处假设了map的定义形式，这是不对的
			// this.readType() // 忽略
			for this.peekByte() != byte('z') {
				k, err = this.Decode()
				if err != nil {
					if err == io.EOF {
						break
					}

					return nil, err
				}
				v, err = this.Decode()
				if err != nil {
					return nil, err
				}
				m[k] = v
			}
			this.readByte()
			this.appendRefs(&m)
			return m, nil

		} else {
			fmt.Println("hello1")
			inst = createInstance(t)
			for this.peekByte() != 'z' {
				if key, err = this.Decode(); err != nil {
					fmt.Printf("key err:%#v", err)
					return nil, err
				}
				fmt.Printf("key:%#v\n", key)
				if value, err = this.Decode(); err != nil {
					fmt.Printf("value err:%#v", err)
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
					fmt.Println("hello2")
					reflect.ValueOf(inst).MethodByName(methodName).Call(args)
				}
			}
			// v = inst
			this.appendRefs(&inst)
			return inst, nil
		}

	case 'f': //fault
		this.Decode() //drop "code"
		code, _ := this.Decode()
		this.Decode() //drop "message"
		message, _ := this.Decode()
		return nil, fmt.Errorf("%s : %s", code, message)

	case 'r': //reply
		// valid-reply ::= r x01 x00 header* object z
		// fault-reply ::= r x01 x00 header* fault z
		this.next(a[:2])
		return this.Decode()

	case 'R': //ref, 一个整数，用以指代前面的list 或者 map
		s = a[:4]
		l, err = this.next(s)
		if err != nil {
			return nil, err
		}
		if l != 4 {
			return nil, ErrNotEnoughBuf
		}
		l = int(UnpackInt32(s)) // ref index

		if len(this.refs) <= l {
			return nil, ErrIllegalRefIndex
		}
		return &this.refs[l], nil

	default:
		return nil, fmt.Errorf("Invalid type: %v,>>%v<<<", string(t), this.peek(this.len()))
	}
}

func (d *Decoder) readInt32(flag int32) (interface{}, error) {
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
	//direct integer
	case tag >= 0x80 && tag <= 0xbf:
		return int32(tag - BC_INT_ZERO), nil

	case tag >= 0xc0 && tag <= 0xcf:
		if _, err = io.ReadFull(d.reader, buf[:1]); err != nil {
			return nil, newCodecError("short integer", err)
		}
		return int32(tag-BC_INT_BYTE_ZERO)<<8 + int32(buf[0]), nil

	case tag >= 0xd0 && tag <= 0xd7:
		if _, err = io.ReadFull(d.reader, buf[:2]); err != nil {
			return nil, newCodecError("short integer", err)
		}
		return int32(tag-BC_INT_SHORT_ZERO)<<16 + int32(buf[1])<<8 + int32(buf[0]), nil

	case tag == BC_INT:
		buf := make([]byte, 4)
		if _, err := io.ReadFull(d.reader, buf); err != nil {
			return nil, newCodecError("parse int", err)
		}
		return int32(buf[0])<<24 + int32(buf[1])<<16 + int32(buf[2])<<8 + int32(buf[3]), nil

	default:
		return nil, newCodecError("integer wrong tag:" + fmt.Sprintf("%d-%#x", int(tag), tag))
	}
}

func (d *Decoder) readInt64(flag int32) (interface{}, error) {
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
			return nil, newCodecError("short integer", err)
		}
		return int64(tag-BC_LONG_BYTE_ZERO)<<8 + int64(buf[0]), nil

	case tag >= 0x38 && tag <= 0x3f:
		if _, err := io.ReadFull(d.reader, buf[:2]); err != nil {
			return nil, newCodecError("short integer", err)
		}

		return int64(tag-BC_LONG_SHORT_ZERO)<<16 + int64(buf[1])<<8 + int64(buf[0]), nil

	case tag == BC_LONG:
		if _, err := io.ReadFull(d.reader, buf[:8]); err != nil {
			return nil, newCodecError("parse long", err)
		}
		return int64(buf[0])<<56 + int64(buf[1])<<48 + int64(buf[2])<<40 + int64(buf[3])<<32 +
			int64(buf[4])<<24 + int64(buf[5])<<16 + int64(buf[6])<<8 + int64(buf[7]), nil

	default:
		return nil, newCodecError("long wrong tag " + fmt.Sprintf("%d-%#x", int(tag), tag))
	}
}

func (d *Decoder) readDouble(flag int32) (interface{}, error) {
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
		return d.readInt32(TAG_READ)

	case BC_DOUBLE_ZERO:
		return float64(0), nil

	case BC_DOUBLE_ONE:
		return float64(1), nil

	case BC_DOUBLE_BYTE:
		tag, _ = d.readByte()
		return float64(tag), nil

	case BC_DOUBLE_SHORT:
		if _, err = io.ReadFull(d.reader, buf[:2]); err != nil {
			return nil, newCodecError("short integer", err)
		}

		return float64(int(buf[0])*256 + int(buf[1])), nil

	case BC_DOUBLE_MILL:
		i, _ := d.readInt32(TAG_READ)
		return float64(i.(int32)), nil

	case BC_DOUBLE:
		if _, err = io.ReadFull(d.reader, buf[:8]); err != nil {
			return nil, newCodecError("short integer", err)
		}

		bits := binary.BigEndian.Uint64(buf)
		datum := math.Float64frombits(bits)
		return datum, nil
	}

	return nil, newCodecError("parse double wrong tag:" + fmt.Sprintf("%d-%#x", int(tag), tag))
}

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
			return -1, newCodecError("byte4 integer", err)
		}

		length = int32(tag-0x30)<<8 + int32(buf[0])
		return length, nil

	case tag == BC_STRING_CHUNK || tag == BC_STRING:
		_, err = io.ReadFull(d.reader, buf[:2])
		if err != nil {
			return -1, newCodecError("byte5 integer", err)
		}
		length = int32(buf[0])<<8 + int32(buf[1])
		return length, nil

	default:
		return -1, newCodecError("getStrLen")
	}
}

func (d *Decoder) readString(flag int32) (interface{}, error) {
	var (
		err    error
		tag    byte
		length int32
		last   bool
	)

	if flag != TAG_READ {
		tag = byte(flag)
	} else {
		tag, _ = d.readByte()
	}

	last = true
	if (tag >= BC_STRING_DIRECT && tag <= STRING_DIRECT_MAX) ||
		(tag >= 0x30 && tag <= 0x33) ||
		(tag == BC_STRING_CHUNK ||
			tag == BC_STRING) {

		//fmt.Println("inside ", tag)
		if tag == BC_STRING_CHUNK {
			last = false
		} else {
			last = true
		}

		l, err := d.getStrLen(tag)
		if err != nil {
			return nil, newCodecError("getStrLen", err)
		}
		length = l
		runeDate := make([]rune, length)
		for i := 0; ; {
			if int32(i) == length {
				if last {
					//fmt.Println("last ", last, "i", i, "length", length)
					return string(runeDate), nil
				}

				b, _ := d.readByte()
				switch {
				case (tag >= BC_STRING_DIRECT && tag <= STRING_DIRECT_MAX) ||
					(tag >= 0x30 && tag <= 0x33) ||
					(tag == BC_STRING_CHUNK ||
						tag == BC_STRING):

					if b == BC_STRING_CHUNK {
						last = false
					} else {
						last = true
					}

					l, err := d.getStrLen(b)
					if err != nil {
						return nil, newCodecError("getStrLen", err)
					}
					length += l
					bs := make([]rune, length)
					copy(bs, runeDate)
					runeDate = bs

				default:
					return nil, newCodecError("tag error ", err)
				}

			} else {
				runeOne, _, err := getRune(d.reader)
				runeDate[i] = runeOne
				if err != nil {
					return nil, newCodecError("byte2 integer", err)
				}
				i++
			}
		}

		return string(runeDate), nil
	} else {
		fmt.Println(tag, length, last)
		return nil, newCodecError("byte3 integer")
	}

}
