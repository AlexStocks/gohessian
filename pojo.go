/******************************************************
# DESC    : pojo registry
# AUTHOR  : Alex Stocks
# EMAIL   : alexstocks@foxmail.com
# MOD     : 2016-10-22 20:25
# FILE    : pojo.go
******************************************************/

package hessian

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// Pls attention that Every field name should be upper case. Otherwise the app may be panic.
type POJO interface {
	JavaClassName() string // 获取对应的java classs的package name
}

type JavaEnum interface {
	JavaClassName() string
	EnumStringArray() []string
}

type classDef struct {
	javaName      string
	fieldNameList []string
	b             []byte // encoded buffer
}

type structInfo struct {
	typ      reflect.Type
	goName   string
	javaName string
	index    int // clsDefList index
}

type POJORegistry struct {
	sync.RWMutex
	clsDefList []classDef            // {class name, field name list...} list
	j2g        map[string]string     // java class name --> go struct name
	registry   map[string]structInfo // go class name --> go struct info
}

var (
	pojoRegistry = POJORegistry{
		j2g:      make(map[string]string),
		registry: make(map[string]structInfo),
	}
)

// 解析struct
func showPOJORegistry() {
	pojoRegistry.Lock()
	for k, v := range pojoRegistry.registry {
		fmt.Println("-->> show Registered types <<----")
		fmt.Println(k, v)
	}
	pojoRegistry.Unlock()
}

// get @v go struct name
func typeof(v interface{}) string {
	return reflect.TypeOf(v).String()
}

// the return value is -1 if @o has been registered.
func RegisterPOJO(o POJO) int {
	var (
		ok bool
		b  []byte
		i  int32
		n  int32
		f  string
		l  []string
		t  structInfo
		c  classDef
	)

	pojoRegistry.Lock()
	if _, ok = pojoRegistry.registry[o.JavaClassName()]; !ok {
		t.goName = typeof(o)
		t.javaName = o.JavaClassName()
		t.typ = reflect.TypeOf(o)
		pojoRegistry.j2g[t.javaName] = t.goName

		b = b[:0]
		b = encBT(b, BC_OBJECT_DEF)
		b = encString(t.javaName, b)
		l = l[:0]
		n = int32(t.typ.NumField())
		b = encInt32(n, b)
		for i = 0; i < n; i++ {
			f = strings.ToLower(t.typ.Field(i).Name)
			l = append(l, f)
			b = encString(f, b)
		}

		c = classDef{javaName: t.javaName, fieldNameList: l}
		c.b = append(c.b, b[:]...)
		t.index = len(pojoRegistry.clsDefList)
		pojoRegistry.clsDefList = append(pojoRegistry.clsDefList, c)
		pojoRegistry.registry[t.goName] = t
		i = t.index
	} else {
		i = -1
	}
	pojoRegistry.Unlock()

	return i
}

// check if go struct name @typeName has been registered or not.
func checkPOJORegistry(typeName string) (int, bool) {
	var (
		ok bool
		s  structInfo
	)
	pojoRegistry.RLock()
	s, ok = pojoRegistry.registry[typeName]
	pojoRegistry.RUnlock()

	return s.index, ok
}

// Create a new instance whose go struct name is @t.
// the return value is nil if @o has been registered.
func createInstance(typeName string) interface{} {
	var (
		ok  bool
		typ reflect.Type
	)

	pojoRegistry.RLock()
	typ, ok = pojoRegistry.registry[typeName]
	pojoRegistry.RUnlock()
	if !ok {
		return nil
	}

	return reflect.New(typ).Interface()
}
