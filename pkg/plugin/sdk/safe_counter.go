package sdk

import "sync"

type SafeCounter struct {
	counter int
	mutex   sync.RWMutex
}

func (m *SafeCounter) Get() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.counter
}

func (m *SafeCounter) Increment() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.counter++
}
