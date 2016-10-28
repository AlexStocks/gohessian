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
	"sync"
)

var (
	pojoReg = POJORegistry{registry: make(map[string]reflect.Type)}
)

type POJO interface {
	GetType() string
}

type POJORegistry struct {
	sync.Mutex
	registry map[string]reflect.Type
}

// 解析struct
func showPOJORegistry() {
	pojoReg.Lock()
	for k, v := range pojoReg.registry {
		fmt.Println("-->> show Registered types <<----")
		fmt.Println(k, v)
	}
	pojoReg.Unlock()
}

// the return value is false if @o has been registered.
func RegisterPOJO(o POJO) bool {
	var ok bool
	pojoReg.Lock()
	if _, ok = pojoReg.registry[o.GetType()]; !ok {
		pojoReg.registry[o.GetType()] = reflect.TypeOf(o)
	}
	pojoReg.Unlock()

	return !ok
}

// check if @typeName has been registered or not.
func checkPOJORegistry(typeName string) bool {
	var ok bool
	pojoReg.Lock()
	_, ok = pojoReg.registry[typeName]
	pojoReg.Unlock()

	return ok
}

// Create a new instance whose type name is @t.
// the return value is nil if @o has been registered.
func createInstance(typeName string) interface{} {
	var (
		ok  bool
		typ reflect.Type
	)

	pojoReg.Lock()
	typ, ok = pojoReg.registry[typeName]
	pojoReg.Unlock()
	if !ok {
		return nil
	}

	return reflect.New(typ).Interface()
}
