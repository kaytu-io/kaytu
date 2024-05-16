package sdk

import "sync"

type LazyLoadCounter struct {
	counter int
	mutex   sync.RWMutex
}

func (m *LazyLoadCounter) Get() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.counter
}

func (m *LazyLoadCounter) Increment() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.counter++
}
