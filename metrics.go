package ikuaiapi

import (
	"sync"
	"time"
)

// Metrics is a small request counter for the client. It is deliberately not
// a Prometheus dependency: a library that forces a metrics backend on its
// consumers is a library nobody can embed in a script.
//
// Record the numbers into whatever the host already runs:
//
//	stats, _ := client.Metrics().GetStats()
//	promGauge.Set(float64(stats.RequestCount))
//
// or attach the SDK to an oekit metrics facade from the host side instead.
type Metrics struct {
	mu            sync.RWMutex
	requestCount  int64
	errorCount    int64
	totalDuration time.Duration
	avgDuration   time.Duration
}

// NewMetrics returns an empty collector.
func NewMetrics() *Metrics { return &Metrics{} }

// RecordRequest notes one completed request.
func (m *Metrics) RecordRequest(duration time.Duration, hasError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requestCount++
	m.totalDuration += duration
	m.avgDuration = time.Duration(int64(m.totalDuration) / m.requestCount)

	if hasError {
		m.errorCount++
	}
}

// GetStats returns the request count, the failure count and the mean
// latency observed so far.
func (m *Metrics) GetStats() (count int64, errors int64, avgDuration time.Duration) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.requestCount, m.errorCount, m.avgDuration
}

// Reset clears the collected statistics.
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestCount = 0
	m.errorCount = 0
	m.totalDuration = 0
	m.avgDuration = 0
}
