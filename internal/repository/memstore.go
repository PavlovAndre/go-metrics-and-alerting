package repository

import "sync"

type MemStore struct {
	mu      sync.RWMutex
	gauge   map[string]float64
	counter map[string]int64
}

// Создание нового экземпляра
func New() *MemStore {
	return &MemStore{
		mu:      sync.RWMutex{},
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

// Обновление значения Gauge
func (ms *MemStore) SetGauge(key string, value float64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.gauge[key] = value
}

// Увеличение счетчика Counter
func (ms *MemStore) AddCounter(key string, value int64) {
	v, ok := ms.counter[key]
	if ok {
		ms.counter[key] = v + value
	} else {
		ms.counter[key] = value
	}
}

// Увеличение счетчика Counter
func (ms *MemStore) SetCounter(key string, value int64) {
	ms.counter[key] = value
}

// Получить значения Gauge
func (ms *MemStore) GetGauges() map[string]float64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.gauge
}

func (ms *MemStore) GetGauge(name string) (float64, bool) {
	value, ok := ms.gauge[name]
	return value, ok
}

// Получить значения Counter
func (ms *MemStore) GetCounters() map[string]int64 {
	return ms.counter
}

// Получить значение одного Counter
func (ms *MemStore) GetCounter(name string) (int64, bool) {
	counter, ok := ms.counter[name]
	return counter, ok
}

var Store *MemStore
