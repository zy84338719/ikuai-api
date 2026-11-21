package ikuaiapi

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"

	"github.com/oekit/logx"
)

// installCapture swaps the global logx backend for one writing to w with a
// debug-level floor, so the only filter that can drop an event is this
// package's own minLevel. The SDK floor is restored after the test.
func installCapture(t *testing.T, w io.Writer) {
	t.Helper()
	logx.SetDefault(logx.NewSlogLogger("ikuai-api-test",
		logx.WithHandler(slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug})),
		logx.WithLevel(logx.LevelDebug),
	))
	t.Cleanup(func() { SetLogLevel(LogLevelWarn) })
}

// TestSetLogLevelFloorsSDKEvents: SetLogLevel 只写 minLevel，而唯一读它的
// enabled() 从未接到日志出口 logCtx —— 公开 API 是 no-op，doc.go:80 承诺的
// 级别过滤根本没生效：默认 warn floor 下 debug 事件照样发出去。
func TestSetLogLevelFloorsSDKEvents(t *testing.T) {
	var buf bytes.Buffer
	installCapture(t, &buf)
	c := &Client{}

	// 默认 floor = warn：debug 事件必须被丢弃。
	c.logCtx(context.Background(), LogLevelDebug, "floor-debug")
	if buf.Len() != 0 {
		t.Fatalf("debug event must be dropped at the default warn floor, got %q", buf.String())
	}

	// 调到 debug 后放行。
	SetLogLevel(LogLevelDebug)
	c.logCtx(context.Background(), LogLevelDebug, "floor-debug")
	if !strings.Contains(buf.String(), "floor-debug") {
		t.Fatalf("debug event must pass after SetLogLevel(LogLevelDebug), got %q", buf.String())
	}

	// LogLevelNone 连 error 都静音。
	buf.Reset()
	SetLogLevel(LogLevelNone)
	c.logCtx(context.Background(), LogLevelError, "floor-none")
	if buf.Len() != 0 {
		t.Fatalf("LogLevelNone must silence every event, got %q", buf.String())
	}
}

// TestSetLogLevelConcurrentWithLogging: enabled() 接进日志热路径后，
// minLevel 变成 SetLogLevel 与 logCtx 并发读写的共享状态——必须 race-free
// （AGENTS.md 并发安全：共享状态 mutex / atomic 保护）。
func TestSetLogLevelConcurrentWithLogging(t *testing.T) {
	// 内容无意义，只求后端不拖慢也不参与断言；用 Discard 避免
	// bytes.Buffer 自身成为竞争源。
	installCapture(t, io.Discard)

	c := &Client{}
	levels := []LogLevel{LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError, LogLevelNone}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				SetLogLevel(levels[(i+j)%len(levels)])
			}
		}(i)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				c.logCtx(context.Background(), LogLevelDebug, "race-probe")
				c.logCtx(context.Background(), LogLevelWarn, "race-probe")
			}
		}()
	}
	wg.Wait()
}
