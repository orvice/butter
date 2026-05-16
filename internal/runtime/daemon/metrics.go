package daemon

import (
	"runtime"
	"sync"
	"time"

	rmetrics "runtime/metrics"
)

// LatencySample is a single bridge invocation latency observation.
type LatencySample struct {
	Timestamp time.Time
	Latency   time.Duration
}

// BridgeMetricsSnapshot is the value returned by Metrics.Snapshot.
type BridgeMetricsSnapshot struct {
	MemoryUsedBytes uint64
	Goroutines      int
	CPUPercent      float64
	Latency         []LatencySample
}

// Metrics collects router-side diagnostics for the daemon bridge.
//
// It exposes a small ring buffer of recent bridge latencies and a one-shot
// process-level resource snapshot. Memory and goroutine counts use the Go
// runtime; CPU% is the average since process start (cpuSeconds / wallSeconds).
type Metrics struct {
	mu      sync.Mutex
	samples []LatencySample // ring buffer (newest at end)
	limit   int

	startCPU       float64
	startWallClock time.Time
}

// NewMetrics constructs a Metrics collector keeping the last `limit` samples.
// limit <= 0 defaults to 60.
func NewMetrics(limit int) *Metrics {
	if limit <= 0 {
		limit = 60
	}
	return &Metrics{
		limit:          limit,
		startCPU:       readCPUSeconds(),
		startWallClock: time.Now(),
	}
}

// RecordLatency appends a single observation. Safe for concurrent use.
func (m *Metrics) RecordLatency(d time.Duration) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.samples = append(m.samples, LatencySample{Timestamp: time.Now().UTC(), Latency: d})
	if len(m.samples) > m.limit {
		m.samples = m.samples[len(m.samples)-m.limit:]
	}
}

// Snapshot returns the current rolled-up metrics.
func (m *Metrics) Snapshot() BridgeMetricsSnapshot {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	var samples []LatencySample
	if m != nil {
		m.mu.Lock()
		samples = make([]LatencySample, len(m.samples))
		copy(samples, m.samples)
		m.mu.Unlock()
	}

	cpu := 0.0
	if m != nil {
		now := time.Now()
		curCPU := readCPUSeconds()
		wall := now.Sub(m.startWallClock).Seconds()
		if wall > 0 {
			cpu = (curCPU - m.startCPU) / wall * 100.0
			if cpu < 0 {
				cpu = 0
			}
		}
	}

	return BridgeMetricsSnapshot{
		MemoryUsedBytes: memStats.Sys,
		Goroutines:      runtime.NumGoroutine(),
		CPUPercent:      cpu,
		Latency:         samples,
	}
}

// readCPUSeconds returns cumulative CPU seconds across all classes. Returns 0
// if the runtime metric is not available.
func readCPUSeconds() float64 {
	sample := []rmetrics.Sample{{Name: "/cpu/classes/total:cpu-seconds"}}
	rmetrics.Read(sample)
	if sample[0].Value.Kind() != rmetrics.KindFloat64 {
		return 0
	}
	return sample[0].Value.Float64()
}
