package collector

import (
	"fmt"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/logger"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/repository"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"math/rand"
	"runtime"
	"time"
)

type Collector struct {
	memStore     *repository.MemStore
	pollInterval int
}

func New(store *repository.MemStore, pollInt int) *Collector {
	return &Collector{memStore: store, pollInterval: pollInt}
}

// CollectMetrics Функция сбора метрик
func (c *Collector) CollectMetrics() {
	for {
		ticker := time.NewTicker(time.Duration(c.pollInterval) * time.Second)
		for range ticker.C {

			m := runtime.MemStats{}
			runtime.ReadMemStats(&m)
			c.memStore.SetGauge("Alloc", float64(m.Alloc))
			c.memStore.SetGauge("BuckHashSys", float64(m.BuckHashSys))
			c.memStore.SetGauge("Frees", float64(m.Frees))
			c.memStore.SetGauge("GCCPUFraction", float64(m.GCCPUFraction))
			c.memStore.SetGauge("GCSys", float64(m.GCSys))
			c.memStore.SetGauge("HeapAlloc", float64(m.HeapAlloc))
			c.memStore.SetGauge("HeapIdle", float64(m.HeapIdle))
			c.memStore.SetGauge("HeapInuse", float64(m.HeapInuse))
			c.memStore.SetGauge("HeapObjects", float64(m.HeapObjects))
			c.memStore.SetGauge("HeapReleased", float64(m.HeapReleased))
			c.memStore.SetGauge("HeapSys", float64(m.HeapSys))
			c.memStore.SetGauge("LastGC", float64(m.LastGC))
			c.memStore.SetGauge("Lookups", float64(m.Lookups))
			c.memStore.SetGauge("MCacheInuse", float64(m.MCacheInuse))
			c.memStore.SetGauge("MCacheSys", float64(m.MCacheSys))
			c.memStore.SetGauge("MSpanInuse", float64(m.MSpanInuse))
			c.memStore.SetGauge("MSpanSys", float64(m.MSpanSys))
			c.memStore.SetGauge("Mallocs", float64(m.Mallocs))
			c.memStore.SetGauge("NextGC", float64(m.NextGC))
			c.memStore.SetGauge("NumForcedGC", float64(m.NumForcedGC))
			c.memStore.SetGauge("NumGC", float64(m.NumGC))
			c.memStore.SetGauge("OtherSys", float64(m.OtherSys))
			c.memStore.SetGauge("PauseTotalNs", float64(m.PauseTotalNs))
			c.memStore.SetGauge("StackInuse", float64(m.StackInuse))
			c.memStore.SetGauge("StackSys", float64(m.StackSys))
			c.memStore.SetGauge("Sys", float64(m.Sys))
			c.memStore.SetGauge("TotalAlloc", float64(m.TotalAlloc))
			c.memStore.SetGauge("RandomValue", rand.Float64())
			c.memStore.AddCounter("PollCount", 1)
		}
	}
}

// CollectSystemMetrics Функция сбора метрик
func (c *Collector) CollectSystemMetrics() {
	for {
		ticker := time.NewTicker(time.Duration(c.pollInterval) * time.Second)
		for range ticker.C {
			memStat, err := mem.VirtualMemory()
			if err != nil {
				logger.Log.Infow("Ошибка опроса системных переменных", err)
			}

			c.memStore.SetGauge("TotalMemory", float64(memStat.Total))
			c.memStore.SetGauge("FreeMemory", float64(memStat.Free))

			cpuStat, err := cpu.Percent(0, true)
			if err != nil {
				logger.Log.Infow("Ошибка опроса количества CPU", err)
				continue
			}
			for i, percent := range cpuStat {
				c.memStore.SetGauge(fmt.Sprintf("CPUutilization%d", i), percent)
			}
		}
	}
}
