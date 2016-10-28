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

func TestEncInt32(t *testing.T) {
	var b = make([]byte, 4)
	b = Encode(20161024, b[:0])
	if len(b) == 0 {
		t.Fail()
	}
	b = Encode(int32(20161024), b[:0])
	if len(b) == 0 {
		t.Fail()
	}
}

func TestEncInt64(t *testing.T) {
	var b = make([]byte, 8)
	b = Encode(int64(20161024), b[:0])
	if len(b) == 0 {
		t.Fail()
	}
}

func TestEncDate(t *testing.T) {
	var b = make([]byte, 8)
	tz, _ := time.Parse("2006-01-02 15:04:05", "2014-02-09 06:15:23")
	b = Encode(tz, b[:0])
	if len(b) == 0 {
		t.Fail()
	}
	want := []byte{0x64, 0x00, 0x00, 0x01, 0x44, 0x15, 0x49, 0x34, 0x78}
	assert(want, b, t)
}

func TestEncDouble(t *testing.T) {
	var b = make([]byte, 8)
	b = Encode(2016.1024, b[:0])
	if len(b) == 0 {
		t.Fail()
	}
}

func TestEncString(t *testing.T) {
	var b = make([]byte, 64)
	b = Encode("hello", b[:0])
	if len(b) == 0 {
		t.Fail()
	}
}

func TestEncBinary(t *testing.T) {
	var b = make([]byte, 64)
	b = Encode([]byte{}, b[:0])
	want := []byte{0x42, 0x00, 0x00}
	assert(b, want, t)

	raw := []byte{10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 'a', 'b', 'c', 'd'}
	b = Encode(raw, b[:0])
	want = []byte{0x42, 0x00, 0x0e, 0x0a, 0x09, 0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, 0x61, 0x62, 0x63, 0x64}
	assert(b, want, t)
}

func TestEncList(t *testing.T) {
	var b = make([]byte, 128)
	list := []Any{100, 10.001, "hello", []byte{0, 2, 4, 6, 8, 10}, true, nil, false}
	b = Encode(list, b[:0])
	if len(b) == 0 {
		t.Fail()
	}
}

func TestEncMap(t *testing.T) {
	var b = make([]byte, 128)
	var m = make(map[Any]Any)
	m["hello"] = "world"
	m[100] = "100"
	m[100.1010] = 101910
	m[true] = true
	m[false] = true
	b = Encode(m, b[:0])
	if len(b) == 0 {
		t.Fail()
	}
}
