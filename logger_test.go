package ikuaisdk

import (
	"testing"
	"time"
)

func TestMetrics(t *testing.T) {
	m := NewMetrics()

	// Record some requests
	m.RecordRequest(100*time.Millisecond, false)
	m.RecordRequest(200*time.Millisecond, false)
	m.RecordRequest(150*time.Millisecond, true)

	count, errors, avgDuration := m.GetStats()
	if count != 3 {
		t.Errorf("GetStats() count = %d, want 3", count)
	}
	if errors != 1 {
		t.Errorf("GetStats() errors = %d, want 1", errors)
	}
	expectedAvg := time.Duration((100 + 200 + 150) / 3 * int64(time.Millisecond))
	if avgDuration != expectedAvg {
		t.Errorf("GetStats() avgDuration = %v, want %v", avgDuration, expectedAvg)
	}

	// Reset
	m.Reset()
	count, errors, _ = m.GetStats()
	if count != 0 || errors != 0 {
		t.Error("Reset() should reset all metrics to 0")
	}
}

func TestDefaultLogger(t *testing.T) {
	// Just verify that creating loggers doesn't panic
	logger := NewDefaultLogger(LogLevelDebug)
	logger.Debug("test debug %s", "message")
	logger.Info("test info %s", "message")
	logger.Warn("test warn %s", "message")
	logger.Error("test error %s", "message")

	// Test with different levels
	logger = NewDefaultLogger(LogLevelError)
	logger.Debug("should not print")
	logger.Error("should print")

	logger = NewDefaultLogger(LogLevelNone)
	logger.Error("should not print")
}

func TestLogLevel(t *testing.T) {
	levels := []LogLevel{
		LogLevelDebug,
		LogLevelInfo,
		LogLevelWarn,
		LogLevelError,
		LogLevelNone,
	}

	// Just verify levels are defined correctly
	for i, level := range levels {
		if int(level) != i {
			t.Errorf("LogLevel %d has unexpected value %d", i, level)
		}
	}
}
