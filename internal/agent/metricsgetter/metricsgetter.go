// Package metricsgetter provides functionality to get  application's runtime and extra stats.
//
// Consists of maps with callback functions.
package metricsgetter

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// GaugeMetricsGetter is map with callback functions as values.
// Callbacks return float64 value.
//
// Example of use:
//
//	 type Metric struct {
//			name string
//			value float64
//		}
//
//		var stats runtime.MemStats
//		runtime.ReadMemStats(&stats)
//		var metrics []Metric
//		for name, valueGetter := range GaugeMetricsGetter {
//			metrics = append(metrics, Metric{name, valueGetter(&a.stats)})
//		}
var GaugeMetricsGetter = map[string]func(stats *runtime.MemStats) float64{
	"Alloc":         func(stats *runtime.MemStats) float64 { return float64(stats.Alloc) },
	"BuckHashSys":   func(stats *runtime.MemStats) float64 { return float64(stats.BuckHashSys) },
	"Frees":         func(stats *runtime.MemStats) float64 { return float64(stats.Frees) },
	"GCCPUFraction": func(stats *runtime.MemStats) float64 { return stats.GCCPUFraction },
	"GCSys":         func(stats *runtime.MemStats) float64 { return float64(stats.GCSys) },
	"HeapAlloc":     func(stats *runtime.MemStats) float64 { return float64(stats.HeapAlloc) },
	"HeapIdle":      func(stats *runtime.MemStats) float64 { return float64(stats.HeapIdle) },
	"HeapInuse":     func(stats *runtime.MemStats) float64 { return float64(stats.HeapInuse) },
	"HeapObjects":   func(stats *runtime.MemStats) float64 { return float64(stats.HeapObjects) },
	"HeapReleased":  func(stats *runtime.MemStats) float64 { return float64(stats.HeapReleased) },
	"HeapSys":       func(stats *runtime.MemStats) float64 { return float64(stats.HeapSys) },
	"LastGC":        func(stats *runtime.MemStats) float64 { return float64(stats.LastGC) },
	"Lookups":       func(stats *runtime.MemStats) float64 { return float64(stats.Lookups) },
	"MCacheInuse":   func(stats *runtime.MemStats) float64 { return float64(stats.MCacheInuse) },
	"MCacheSys":     func(stats *runtime.MemStats) float64 { return float64(stats.MCacheSys) },
	"MSpanInuse":    func(stats *runtime.MemStats) float64 { return float64(stats.MSpanInuse) },
	"MSpanSys":      func(stats *runtime.MemStats) float64 { return float64(stats.MSpanSys) },
	"Mallocs":       func(stats *runtime.MemStats) float64 { return float64(stats.Mallocs) },
	"NextGC":        func(stats *runtime.MemStats) float64 { return float64(stats.NextGC) },
	"NumForcedGC":   func(stats *runtime.MemStats) float64 { return float64(stats.NumForcedGC) },
	"NumGC":         func(stats *runtime.MemStats) float64 { return float64(stats.NumGC) },
	"OtherSys":      func(stats *runtime.MemStats) float64 { return float64(stats.OtherSys) },
	"PauseTotalNs":  func(stats *runtime.MemStats) float64 { return float64(stats.PauseTotalNs) },
	"StackInuse":    func(stats *runtime.MemStats) float64 { return float64(stats.StackInuse) },
	"StackSys":      func(stats *runtime.MemStats) float64 { return float64(stats.StackSys) },
	"Sys":           func(stats *runtime.MemStats) float64 { return float64(stats.Sys) },
	"TotalAlloc":    func(stats *runtime.MemStats) float64 { return float64(stats.TotalAlloc) },
}

// ExtraStats stores extra stats.
type ExtraStats struct {
	Data map[string]float64
}

// AdditionalGaugeMetricsGetter is map with callback functions as values.
// Callbacks return float64 value and error.
//
// Example of use:
//
//	 type Metric struct {
//			name string
//			value float64
//		}
//
//		var metrics []Metric
//		for name, valueGetter := range AdditionalGaugeMetricsGetter {
//			metricVal,err := valueGetter()
//			if err != nil {
//				//Handle error
//			}
//			metrics = append(metrics, Metric{name, metricVal})
//		}
var AdditionalGaugeMetricsGetter = map[string]func() (float64, error){
	"TotalMemory": func() (float64, error) {
		vm, err := mem.VirtualMemory()
		if err != nil {
			return 0., err
		}

		totalMemory := vm.Total
		return float64(totalMemory), nil
	},
	"FreeMemory": func() (float64, error) {
		vm, err := mem.VirtualMemory()
		if err != nil {
			return 0., err
		}

		freeMemory := vm.Free
		return float64(freeMemory), nil
	},
	"CPUutilization1": func() (float64, error) {
		cpuPercentages, err := cpu.Percent(time.Second, false)
		if err != nil {
			return 0., err
		}
		return cpuPercentages[0], nil
	},
}
