/******************************************************
# DESC    :
# AUTHOR  : Alex Stocks
# EMAIL   : alexstocks@foxmail.com
# MOD     : 2016-10-22 20:24
# FILE    : encode_test.go
******************************************************/

package hessian

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

// go test -v encode.go encode_test.go codec.go

var assert = func(want, got []byte, t *testing.T) {
	if !bytes.Equal(want, got) {
		t.Fatalf("want %v , got %v", want, got)
	}
}

func TestEncNull(t *testing.T) {
	var b []byte
	b = Encode(nil, b)
	if b == nil {
		t.Fail()
	}
	t.Logf("nil enc result:%s\n", string(b))
}

func TestEncBool(t *testing.T) {
	var b = make([]byte, 1)
	b = Encode(true, b[:0])
	if b[0] != 'T' {
		t.Fail()
	}
	want := []byte{0x54}
	assert(want, b, t)

	b = Encode(false, b[:0])
	if b[0] != 'F' {
		t.Fail()
	}
	want = []byte{0x46}
	assert(want, b, t)
}

func TestEncInt32Len1B(t *testing.T) {
	var b = make([]byte, 4)
	var value int32 = 0xe6
	// var value int32 = 0xf016
	b = Encode(value, b[:0])
	if len(b) == 0 {
		t.Fail()
	}
	iDecoder := NewDecoder(b)
	iRes, err := iDecoder.Decode()
	t.Logf("decode(%v) = %v, %v\n", value, iRes, err)
}

func TestEncInt32Len2B(t *testing.T) {
	var b = make([]byte, 4)
	// var value int32 = 0x616
	var value int32 = 0xf016
	b = Encode(value, b[:0])
	if len(b) == 0 {
		t.Fail()
	}
	t.Logf("%#v\n", b)
	iDecoder := NewDecoder(b)
	iRes, err := iDecoder.Decode()
	t.Logf("decode(%#x) = %#x, %v\n", value, iRes, err)
}

func TestEncInt32Len4B(t *testing.T) {
	var b = make([]byte, 4)
	var value int32 = 0x20161024
	b = Encode(value, b[:0])
	if len(b) == 0 {
		t.Fail()
	}
	iDecoder := NewDecoder(b)
	iRes, err := iDecoder.Decode()
	t.Logf("decode(%v) = %v, %v\n", value, iRes, err)
}

func TestEncInt64Len1B(t *testing.T) {
	var b = make([]byte, 8)
	var value int64 = 0xf6
	b = Encode(int64(value), b[:0])
	if len(b) == 0 {
		t.Fail()
	}
	iDecoder := NewDecoder(b)
	iRes, err := iDecoder.Decode()
	t.Logf("decode(int64(%#x)) = %#x, %v\n", value, iRes, err)
}

func TestEncInt64Len2B(t *testing.T) {
	var b = make([]byte, 8)
	var value int64 = 0x2016
	b = Encode(int64(value), b[:0])
	if len(b) == 0 {
		t.Fail()
	}
	iDecoder := NewDecoder(b)
	iRes, err := iDecoder.Decode()
	t.Logf("decode(int64(%#x)) = %#x, %v\n", value, iRes, err)
}

func TestEncInt64Len3B(t *testing.T) {
	var b = make([]byte, 8)
	var value int64 = 101910 // 0x18e16
	// b = Encode(int64(20161024), b[:0])
	b = Encode(int64(value), b[:0])
	if len(b) == 0 {
		t.Fail()
	}
	iDecoder := NewDecoder(b)
	iRes, err := iDecoder.Decode()
	t.Logf("decode(int64(%#x)) = %#x, %v\n", value, iRes, err)
}

func TestEncInt64Len8B(t *testing.T) {
	var b = make([]byte, 8)
	var value int64 = 0x20161024114530
	b = Encode(int64(value), b[:0])
	if len(b) == 0 {
		t.Fail()
	}
	iDecoder := NewDecoder(b)
	iRes, err := iDecoder.Decode()
	t.Logf("decode(int64(%#x)) = %#x, %v\n", value, iRes, err)
}

func TestEncDate(t *testing.T) {
	var b = make([]byte, 8)
	ts := "2014-02-09 06:15:23"
	tz, _ := time.Parse("2006-01-02 15:04:05", ts)
	b = Encode(tz, b[:0])
	if len(b) == 0 {
		t.Fail()
	}
	iDecoder := NewDecoder(b)
	iRes, err := iDecoder.Decode()
	t.Logf("decode(%s, %s) = %v, %v\n", ts, tz.Local(), iRes, err)
}

func TestEncDouble(t *testing.T) {
	b := make([]byte, 8)
	v := 2016.1024
	b = Encode(v, b[:0])
	if len(b) == 0 {
		t.Fail()
	}
	iDecoder := NewDecoder(b)
	iRes, err := iDecoder.Decode()
	t.Logf("decode(%v) = %v, %v\n", v, iRes, err)
}

func TestEncString(t *testing.T) {
	var b = make([]byte, 64)
	v := "hello"
	b = Encode(v, b[:0])
	if len(b) == 0 {
		t.Fail()
	}
	iDecoder := NewDecoder(b)
	iRes, err := iDecoder.Decode()
	t.Logf("decode(%v) = %v, %v\n", v, iRes, err)
}

func TestEncShortRune(t *testing.T) {
	var b = make([]byte, 64)
	v := "我化尘埃飞扬，追寻赤裸逆翔"
	b = Encode(v, b[:0])
	if len(b) == 0 {
		t.Fail()
	}
	iDecoder := NewDecoder(b)
	iRes, err := iDecoder.Decode()
	t.Logf("decode(%v) = %v, %v\n", v, iRes, err)
}

func TestEncRune(t *testing.T) {
	var b = make([]byte, 64)
	v := "我化尘埃飞扬，追寻赤裸逆翔, 奔去七月刑场，时间烧灼滚烫, 回忆撕毁臆想，路上行走匆忙, 难能可贵世上，散播留香磁场, 我欲乘风破浪，踏遍黄沙海洋, 与其误会一场，也要不负勇往, 我愿你是个谎，从未出现南墙, 笑是神的伪装，笑是强忍的伤, 我想你就站在，站在大漠边疆, 我化尘埃飞扬，追寻赤裸逆翔," +
		" 奔去七月刑场，时间烧灼滚烫, 回忆撕毁臆想，路上行走匆忙, 难能可贵世上，散播留香磁场, 我欲乘风破浪，踏遍黄沙海洋, 与其误会一场，也要不负勇往, 我愿你是个谎，从未出现南墙, 笑是神的伪装，笑是强忍的伤, 我想你就站在，站在大漠边疆."
	v = v + v + v + v + v
	v = v + v + v + v + v
	v = v + v + v + v + v
	v = v + v + v + v + v
	v = v + v + v + v + v
	fmt.Printf("vlen:%d\n", len(v))
	b = Encode(v, b[:0])
	if len(b) == 0 {
		t.Fail()
	}
	iDecoder := NewDecoder(b)
	iRes, err := iDecoder.Decode()
	if err != nil {
		t.Errorf("Decode() = %v", err)
	}
	// t.Logf("decode(%v) = %v, %v\n", v, iRes, err)
	assert([]byte(iRes.(string)), []byte(v), t)
}

func TestEncBinary(t *testing.T) {
	b := make([]byte, 64)
	raw := []byte{}
	b = Encode(raw, b[:0])
	iDecoder := NewDecoder(b)
	iRes, err := iDecoder.Decode()
	if err != nil {
		t.Errorf("Decode() = %v", err)
	}
	t.Logf("decode(%v) = %v, %v\n", raw, iRes, err)

	b = make([]byte, 64)
	raw = []byte{10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 'a', 'b', 'c', 'd'}
	b = Encode(raw, b[:0])
	t.Logf("encode(%v) = %v\n", raw, b)
	iDecoder = NewDecoder(b)
	iRes, err = iDecoder.Decode()
	if err != nil {
		t.Errorf("Decode() = %v", err)
	}
	t.Logf("decode(%v) = %v, %v, equal:%v\n", raw, iRes, err, bytes.Equal(raw, iRes.([]byte)))
	assert(raw, iRes.([]byte), t)
}

func TestEncList(t *testing.T) {
	var b = make([]byte, 128)
	list := []interface{}{100, 10.001, "hello", []byte{0, 2, 4, 6, 8, 10}, true, nil, false}
	b = Encode(list, b[:0])
	if len(b) == 0 {
		t.Fail()
	}

	iDecoder := NewDecoder(b)
	iRes, err := iDecoder.Decode()
	if err != nil {
		t.Errorf("Decode() = %v", err)
	}
	t.Logf("decode(%v) = %v, %v\n", list, iRes, err)
}

func TestEncUntypedMap(t *testing.T) {
	var b = make([]byte, 128)
	var m = make(map[interface{}]interface{})
	m["hello"] = "world"
	m[100] = "100"
	m[100.1010] = 101910
	m[true] = true
	m[false] = true
	b = Encode(m, b[:0])
	if len(b) == 0 {
		t.Fail()
	}

	iDecoder := NewDecoder(b)
	iRes, err := iDecoder.Decode()
	if err != nil {
		t.Errorf("Decode() = %v", err)
	}
	t.Logf("decode(%v) = %v, %v\n", m, iRes, err)
}
