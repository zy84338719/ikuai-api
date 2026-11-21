package ikuaiapi

import (
	"sync/atomic"

	"github.com/oekit/logx"
)

// LogLevel selects the severity of an SDK event.
type LogLevel int

const (
	// LogLevelDebug is per-request detail.
	LogLevelDebug LogLevel = iota
	// LogLevelInfo is lifecycle events such as retry scheduling.
	LogLevelInfo
	// LogLevelWarn is a recoverable condition, e.g. a retried request.
	LogLevelWarn
	// LogLevelError is an operation that failed.
	LogLevelError
	// LogLevelNone silences the SDK's own logging.
	LogLevelNone
)

// String renders the level name.
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "debug"
	case LogLevelInfo:
		return "info"
	case LogLevelWarn:
		return "warn"
	case LogLevelError:
		return "error"
	case LogLevelNone:
		return "none"
	default:
		return "unknown"
	}
}

// minLevel is the floor for the SDK's own events. Lower levels are
// dropped before they reach the backend, so a chatty SDK never costs a
// caller log volume they did not ask for. Atomic: SetLogLevel writes from
// the caller's goroutine while logCtx reads it on every request goroutine.
var minLevel atomic.Int64

func init() { minLevel.Store(int64(LogLevelWarn)) }

// SetLogLevel sets the floor for the SDK's own events. The global log
// backend still applies on top of this.
func SetLogLevel(l LogLevel) { minLevel.Store(int64(l)) }

// log returns the package-level logger.
//
// It uses logx.Delegating rather than capturing logx.Default(): the SDK
// must not hard-code a service name, because the service name belongs to
// whoever consumes this package, and a logger captured at init time would
// never see a logx.SetDefault call made later in main. Delegating resolves
// lazily on every call, so a consumer that installs a backend — slog,
// hlog, anything — picks it up without this package knowing.
func log() logx.Logger { return logx.Delegating("ikuai-api") }

// enabled reports whether an event at the given level should be emitted at
// all.
func enabled(level LogLevel) bool {
	m := LogLevel(minLevel.Load())
	return level >= m && m != LogLevelNone
}
