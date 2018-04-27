/******************************************************
# DESC    : dubbo request
# AUTHOR  : Alex Stocks
# EMAIL   : alexstocks@foxmail.com
# MOD     : 2017-10-15 22:42
# FILE    : request.go
******************************************************/

package hessian

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

import (
	jerrors "github.com/juju/errors"
)

/////////////////////////////////////////
// dubbo
/////////////////////////////////////////

/**
 * 协议头是16字节的定长数据
 * 2字节magic字符串0xdabb,0-7高位，8-15低位
 * 1字节的消息标志位。16-20序列id,21 event,22 two way,23请求或响应标识
 * 1字节状态。当消息类型为响应时，设置响应状态。24-31位。
 * 8字节，消息ID,long类型，32-95位。
 * 4字节，消息长度，96-127位
 **/
const (
	// header length.
	HEADER_LENGTH = 16

	// magic header.
	MAGIC      = uint16(0xdabb)
	MAGIC_HIGH = byte(0xda)
	MAGIC_LOW  = byte(0xbb)

	// message flag.
	FLAG_REQUEST = byte(0x80)
	FLAG_TWOWAY  = byte(0x40)
	FLAG_EVENT   = byte(0x20) // for heartbeat

	SERIALIZATION_MASK = byte(0x1f)
)

const (
	DUBBO_VERSION = "2.5.4"
)

const (
	DEFAULT_LEN        = 8388608 // 8 * 1024 * 1024 default body max length
	PACKET_DEFAULT_LEN = 131400  // 128 * 1024 default one buf max length
)

var (
	REQUEST_HEADER = [HEADER_LENGTH]byte{MAGIC_HIGH, MAGIC_LOW, FLAG_TWOWAY | FLAG_REQUEST}
)

// com.alibaba.dubbo.common.utils.ReflectUtils.ReflectUtils.java line245 getDesc
func getArgType(v interface{}) string {
	if v == nil {
		return "V"
	}

	switch v.(type) {
	// 基本类型的序列化tag
	case nil:
		return "V"
	case bool:
		return "Z"
	case byte:
		return "B"
	case int8:
		return "B"
	case int16:
		return "S"
	case uint16: // 相当于Java的Char
		return "C"
	// case rune:
	//	return "C"
	case int:
		return "I"
	case int32:
		return "I"
	case int64:
		return "J"
	case time.Time:
		return "java.util.Date"
	case float32:
		return "F"
	case float64:
		return "D"
	case string:
		return "java.lang.String"
	case []byte:
		return "[B"
	case map[interface{}]interface{}:
		// return  "java.util.HashMap"
		return "java.util.Map"

	//  复杂类型的序列化tag
	default:
		t := reflect.TypeOf(v)
		if reflect.Ptr == t.Kind() {
			t = reflect.TypeOf(reflect.ValueOf(v).Elem())
		}
		switch t.Kind() {
		case reflect.Struct:
			return "java.lang.Object"
		case reflect.Slice, reflect.Array:
			// return "java.util.ArrayList"
			return "java.util.List"
		case reflect.Map: // 进入这个case，就说明map可能是map[string]int这种类型
			return "java.util.Map"
		default:
			return ""
		}
	}

	return "java.lang.RuntimeException"
}

func getArgsTypeList(args []interface{}) (string, error) {
	var (
		typ   string
		types string
	)

	for i := range args {
		typ = getArgType(args[i])
		if typ == "" {
			return types, fmt.Errorf("cat not get arg %#v type", args[i])
		}
		if !strings.Contains(typ, ".") {
			types += typ
		} else {
			// java.util.List -> Ljava/util/List;
			types += "L" + strings.Replace(typ, ".", "/", -1) + ";"
		}
	}

	return types, nil
}

// create request buffer body
func PackRequest(reqID int64, path, dubboInterface, version, method string, args []interface{}, timeout int) ([]byte, error) {
	var sirializationID = byte(78) // java 中标识一个class的ID
	var header = make([]byte, HEADER_LENGTH<<10)
	header = header[:0]

	//////////////////////////////////////////
	// header
	//////////////////////////////////////////

	// magic
	header = append(header, REQUEST_HEADER[:]...)
	// serialization id, two way flag, event, request/response flag
	header[2] |= byte(sirializationID & SERIALIZATION_MASK)
	// request id
	binary.BigEndian.PutUint64(header[4:], uint64(reqID))

	// com.alibaba.dubbo.rpc.protocol.dubbo.DubboCodec.DubboCodec.java line144 encodeRequestData
	//////////////////////////////////////////
	// body
	//////////////////////////////////////////

	//模拟dubbo的请求
	var encoder Encoder
	encoder.Append(header[:HEADER_LENGTH])

	// dubbo version + path + version + method
	encoder.Encode(DUBBO_VERSION)
	encoder.Encode(dubboInterface)
	encoder.Encode(version)
	encoder.Encode(method)

	// args = args type list + args value list
	var types string
	var err error
	types, err = getArgsTypeList(args)
	if err != nil {
		return nil, jerrors.Annotatef(err, " PackRequest(args:%+v)", args)
	}
	encoder.Encode(types) //"Ljava/lang/Integer;"
	for _, v := range args {
		encoder.Encode(v)
	}

	serviceParams := make(map[string]string)
	serviceParams["path"] = path
	serviceParams["interface"] = dubboInterface
	if len(version) != 0 {
		serviceParams["version"] = version
	}
	if timeout != 0 {
		serviceParams["timeout"] = strconv.Itoa(timeout)
	}

	encoder.Encode(serviceParams)

	encBuf := encoder.Buffer()
	encBufLen := len(encBuf)
	// header{body length}
	if encBufLen > int(DEFAULT_LEN) { // 8M
		return nil, fmt.Errorf("Data length %d too large, max payload %d", encBufLen, DEFAULT_LEN)
	}
	binary.BigEndian.PutUint32(encBuf[12:], uint32(encBufLen-HEADER_LENGTH))

	return encBuf, nil
}
