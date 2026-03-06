package ikuaisdk

import (
	"sync"
	"testing"
	"time"
)

func TestDefaultLogger(t *testing.T) {
	logger := NewDefaultLogger(LogLevelDebug)

	// 测试各级别日志不会 panic
	logger.Debug("test debug message: %s", "value")
	logger.Info("test info message: %d", 123)
	logger.Warn("test warn message")
	logger.Error("test error message")
}

func TestDefaultLogger_Levels(t *testing.T) {
	tests := []struct {
		name  string
		level LogLevel
	}{
		{"Debug level", LogLevelDebug},
		{"Info level", LogLevelInfo},
		{"Warn level", LogLevelWarn},
		{"Error level", LogLevelError},
		{"None level", LogLevelNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewDefaultLogger(tt.level)
			logger.Debug("debug")
			logger.Info("info")
			logger.Warn("warn")
			logger.Error("error")
		})
	}
}

func TestMetrics(t *testing.T) {
	metrics := NewMetrics()

	// 测试记录请求
	metrics.RecordRequest(100*time.Millisecond, false)
	metrics.RecordRequest(200*time.Millisecond, false)
	metrics.RecordRequest(150*time.Millisecond, true)

	count, errors, avgDuration := metrics.GetStats()
	if count != 3 {
		t.Errorf("expected 3 requests, got %d", count)
	}
	if errors != 1 {
		t.Errorf("expected 1 error, got %d", errors)
	}
	expectedAvg := time.Duration((100+200+150)/3) * time.Millisecond
	if avgDuration != expectedAvg {
		t.Errorf("expected avg duration %v, got %v", expectedAvg, avgDuration)
	}

	// 测试重置
	metrics.Reset()
	count, errors, avgDuration = metrics.GetStats()
	if count != 0 {
		t.Errorf("expected 0 requests after reset, got %d", count)
	}
}

func TestMetrics_Concurrent(t *testing.T) {
	metrics := NewMetrics()

	// 并发测试
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			metrics.RecordRequest(100*time.Millisecond, false)
		}()
	}
	wg.Wait()

	count, _, _ := metrics.GetStats()
	if count != 100 {
		t.Errorf("expected 100 requests, got %d", count)
	}
}
