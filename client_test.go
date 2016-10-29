/******************************************************
# DESC    : client.go unit test
# AUTHOR  : Alex Stocks
# EMAIL   : alexstocks@foxmail.com
# MOD     : 2016-10-22 20:26
# FILE    : client_test.go
******************************************************/

package hessian

import (
	"bytes"
	"fmt"
	"log"
	"testing"
)

const (
	MATH_URL     = "http://localhost:8000/math"     //整数四则运算hessian测试接口
	DATATYPE_URL = "http://localhost:8000/datatype" //数据类型测试结果,传入参数，返回该参数的(响应)编码结果
	// DATATYPE_URL = "http://192.168.35.3:8000/datatype" //数据类型测试结果,传入参数，返回该参数的(响应)编码结果
)

func TestHttpPost(t *testing.T) {
	t.SkipNow() // 跳过这个错误测试
	fmt.Println("\ntest http post")
	data := bytes.NewBuffer([]byte{0, 1, 3, 4})
	rb, _ := httpPost(DATATYPE_URL, bytes.NewReader(data.Bytes())) // 没有method，非法请求
	log.Println(rb)
	log.Println(string(rb))
}

//整数 数学运算测试
func TestMath(t *testing.T) {
	var (
		err error
		res interface{}
	)

	fmt.Println("\ntest math")
	res, err = Request(MATH_URL, "Add", 100, 200)
	fmt.Println("Add(100, 200) = res:", res, ", err:", err)
	res, err = Request(MATH_URL, "Sub", 100, 200)
	fmt.Println("Sub(100, 200) = res:", res, ", err:", err)
	res, err = Request(MATH_URL, "Mul", 100, 200)
	fmt.Println("Mul(100, 200) = res:", res, ", err:", err)
	res, err = Request(MATH_URL, "Div", 200, 50)
	fmt.Println("Div(200, 50) = res:", res, ", err:", err)
}

//数据类型测试
func TestDataType(t *testing.T) {
	var (
		err error
		res interface{}
	)

	fmt.Println("\ntest data type")

	// void
	res, err = Request(DATATYPE_URL, "DTVoid")
	fmt.Println("DTVoid(Map) = res:", res, ", err:", err)

	// null
	res, err = Request(DATATYPE_URL, "DTNull")
	fmt.Println("DTNull() = res:", res, ", err:", err)

	// bool true
	res, err = Request(DATATYPE_URL, "DTBoolean", true)
	fmt.Println("DTBoolean(true) = res:", res, ", err:", err)

	// bool false
	res, err = Request(DATATYPE_URL, "DTBoolean", false)
	fmt.Println("DTBoolean(false) = res:", res, ", err:", err)

	// int
	res, err = Request(DATATYPE_URL, "DTInt", 1000)
	fmt.Println("DTInt(int) = res:", res, ", err:", err)

	// double
	res, err = Request(DATATYPE_URL, "DTDouble", 2016.1026)
	fmt.Println("DTDouble(1016.1026) = res:", res, ", err:", err)

	// bytes
	res, err = Request(DATATYPE_URL, "DTBytes", []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	fmt.Println("DTBytes([]byte) = res:", res, ", err:", err)

	// string
	res, err = Request(DATATYPE_URL, "DTString", "_BEGIN_兔兔你小姨子_END_")
	fmt.Println("DTString(string) = res:", res, ", err:", err)

	// list
	list := []Any{100, 10.001, "hello", []byte{0, 2, 4, 6, 8, 10}, true, nil, false}
	res, err = Request(DATATYPE_URL, "DTList", list)
	fmt.Println("DTList([]Any) = res:", res, ", err:", err)

	// map
	var m = make(map[Any]Any)
	m["你好"] = "哈哈哈"
	m[100] = "嘿嘿"
	m[100.1010] = 101910
	m[true] = true
	m[false] = true
	m["list"] = list
	res, err = Request(DATATYPE_URL, "DTMap", m)
	fmt.Println("DTMap(Map) = res:", res, ", err:", err)

	var untypedMap = make(map[string]int)
	untypedMap["你好"] = 0
	untypedMap["hello"] = 1
	res, err = Request(DATATYPE_URL, "DTMap", untypedMap)
	fmt.Println("DTMap(Map) = res:", res, ", err:", err)
}

//异常测试,服务器已经抛出异常，但客户端看到的是200和空响应
//curl -vvv --data-binary "c\x00\x01m\x00\adataIntz" -H "Content-Type: application/binary" http://localhost:8080/DecoderTest/dt
//curl -vvv --data-binary "c\x00\x01m\x00\x0EthorwExceptionz" -H "Content-Type: application/binary" http://localhost:8080/DecoderTest/dt
func TestException(t *testing.T) {
	var (
		err error
		res interface{}
	)

	fmt.Println("\ntest excetion")
	res, err = Request(DATATYPE_URL, "thorwException")
	// DT(exception) = res: <nil> , err: NoSuchMethodException : The service has no method named: thorwException
	fmt.Println("DT(exception) = res:", res, ", err:", err)
}
