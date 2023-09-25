package pics1f

import (
	"reflect"
	"sync"
)

type pagearray []Page

func (parr *pagearray) Reset() {
	for _, p := range *parr {
		p.Reset()
	}
}

var (
	pagepool        = new(GenericPool[pagearray]).Init()
	stringarraypool = new(GenericPool[[]string]).Init()
)

type GenericPool[T any] struct {
	pool sync.Pool
}

func (pool *GenericPool[T]) Init() *GenericPool[T] {
	pool.pool.New = func() any {
		return new(T)
	}
	return pool
}

// SelectFromPool ...
func (pool *GenericPool[T]) SelectFromPool() *T {
	return pool.pool.Get().(*T)
}

// PutIntoPool ...
func (pool *GenericPool[T]) PutIntoPool(x *T) {
	rst := reflect.ValueOf(x).MethodByName("Reset")
	if rst.IsValid() {
		rst.Call(nil)
	}
	pool.pool.Put(x)
}
