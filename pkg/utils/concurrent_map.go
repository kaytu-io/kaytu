package utils

import (
	"sync"
)

type ConcurrentMap[K comparable, V any] struct {
	data sync.Map
}

func NewConcurrentMap[K comparable, V any]() ConcurrentMap[K, V] {
	return ConcurrentMap[K, V]{
		data: sync.Map{},
	}
}

func (cm *ConcurrentMap[K, V]) Set(key K, value V) {
	cm.data.Store(key, value)
}

func (cm *ConcurrentMap[K, V]) Delete(key K) {
	cm.data.Delete(key)
}

func (cm *ConcurrentMap[K, V]) Get(key K) (V, bool) {
	v, ok := cm.data.Load(key)
	if !ok {
		return *new(V), false
	}
	return v.(V), true
}

func (cm *ConcurrentMap[K, V]) Range(f func(key K, value V) bool) {
	cm.data.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}

func (cm *ConcurrentMap[K, V]) CompareAndSwap(key K, old, new V) bool {
	return cm.data.CompareAndSwap(key, old, new)
}

func (cm *ConcurrentMap[K, V]) CompareAndDelete(key K, value V) bool {
	return cm.data.CompareAndDelete(key, value)
}

func (cm *ConcurrentMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	return cm.data.LoadOrStore(key, value)
}

func (cm *ConcurrentMap[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	return cm.data.LoadAndDelete(key)
}
