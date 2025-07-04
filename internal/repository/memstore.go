package repository

type MemStore struct {
	gauge   map[string]float64
	counter map[string]int64
}

// Создание нового экземпляра
func New() *MemStore {
	return &MemStore{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

// Обновление значения Gauge
func (ms *MemStore) SetGauge(key string, value float64) {
	ms.gauge[key] = value
}

// Увеличение счетчика Counter
func (ms *MemStore) SetCounter(key string, value int64) {
	v, ok := ms.counter[key]
	if ok {
		ms.counter[key] = v + value
	} else {
		ms.counter[key] = value
	}
}

// Получить значения Gauge
func (ms *MemStore) GetGauges() map[string]float64 {
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
