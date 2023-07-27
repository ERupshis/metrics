package metricsgetter

import "runtime"

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
