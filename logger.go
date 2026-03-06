package ikuaisdk

import (
	"log"
	"os"
	"sync"
	"time"
)

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelNone
)

type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type defaultLogger struct {
	level  LogLevel
	logger *log.Logger
}

func NewDefaultLogger(level LogLevel) Logger {
	return &defaultLogger{
		level:  level,
		logger: log.New(os.Stderr, "[iKuai SDK] ", log.LstdFlags),
	}
}

func (l *defaultLogger) Debug(msg string, args ...interface{}) {
	if l.level <= LogLevelDebug {
		l.logger.Printf("[DEBUG] "+msg, args...)
	}
}

func (l *defaultLogger) Info(msg string, args ...interface{}) {
	if l.level <= LogLevelInfo {
		l.logger.Printf("[INFO] "+msg, args...)
	}
}

func (l *defaultLogger) Warn(msg string, args ...interface{}) {
	if l.level <= LogLevelWarn {
		l.logger.Printf("[WARN] "+msg, args...)
	}
}

func (l *defaultLogger) Error(msg string, args ...interface{}) {
	if l.level <= LogLevelError {
		l.logger.Printf("[ERROR] "+msg, args...)
	}
}

type Metrics struct {
	mu            sync.RWMutex
	requestCount  int64
	errorCount    int64
	totalDuration time.Duration
	avgDuration   time.Duration
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

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

func (m *Metrics) GetStats() (count int64, errors int64, avgDuration time.Duration) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.requestCount, m.errorCount, m.avgDuration
}

func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestCount = 0
	m.errorCount = 0
	m.totalDuration = 0
	m.avgDuration = 0
}
