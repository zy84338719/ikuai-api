package ikuaiapi

import (
	"errors"
	"strings"
	"time"
)

// APIError is returned when the router replies with a non-success envelope
// or an HTTP 4xx/5xx status. It is the only typed error the SDK raises for
// protocol-level failures; transport failures come back as *NetworkError.
type APIError struct {
	HTTPStatus int
	Code       int
	Message    string
	Details    []APIErrorDetail
	// RetryAfter, when non-zero, carries a server-advised back-off hint
	// (parsed from the HTTP Retry-After header, typically on 429/503).
	RetryAfter time.Duration
}

type APIErrorDetail struct {
	Field string `json:"field"`
	Type  string `json:"type"`
	Msg   string `json:"msg"`
}

func (e *APIError) Error() string {
	var b strings.Builder
	if e.HTTPStatus != 0 {
		b.WriteString("HTTP ")
		writeInt(&b, e.HTTPStatus)
		if e.Code != 0 {
			b.WriteString(", ")
		}
	}
	if e.Code != 0 {
		b.WriteString("code ")
		writeInt(&b, e.Code)
	}
	if b.Len() > 0 {
		b.WriteString(": ")
	}
	b.WriteString(e.Message)
	for _, d := range e.Details {
		b.WriteString("\n  - ")
		b.WriteString(d.Field)
		b.WriteString(": ")
		b.WriteString(d.Msg)
	}
	return b.String()
}

// NetworkError wraps connection-level failures (DNS, refused, TLS, timeout).
type NetworkError struct {
	Message string
	Cause   error
}

func (e *NetworkError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *NetworkError) Unwrap() error { return e.Cause }

// IsRetryable reports whether a caller may safely retry the request that
// produced this error. It encodes the SDK's own retry policy so applications
// can reuse it for custom retry loops or circuit-breaker decisions.
//
//   - *NetworkError: retryable (transport hiccups usually clear up), but only
//     for idempotent verbs — the SDK never auto-retries a write on a network
//     error because the request may have reached the router.
//   - *APIError: retryable on HTTP 429 (rate limited), 5xx, and gateway errors.
//     4xx (other than 429) are not retryable.
//
// Pass the HTTP method to qualify network errors: IsRetryable on a *NetworkError
// returns false for POST/PUT/PATCH to avoid duplicate writes.
func (e *NetworkError) IsRetryable() bool { return true }

// IsRetryable on an *APIError reflects server-side retryability.
func (e *APIError) IsRetryable() bool {
	if e == nil {
		return false
	}
	// 429 Too Many Requests and any 5xx are safe to retry.
	return e.HTTPStatus == 429 || e.HTTPStatus >= 500
}

// errorHints maps known iKuai error codes to human-friendly hints.
var errorHints = map[int]string{
	3001: "parameter error, check the request body or required fields",
	3007: "token expired or invalid, obtain a new one from the router web UI",
	1008: "session expired, obtain a new token from the router web UI",
}

func writeInt(b *strings.Builder, n int) {
	if n == 0 {
		b.WriteByte('0')
		return
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		b.WriteByte('-')
	}
	b.Write(buf[i:])
}

func isAPIError(err error) (*APIError, bool) {
	if err == nil {
		return nil, false
	}
	var e *APIError
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

func isNetworkError(err error) (*NetworkError, bool) {
	if err == nil {
		return nil, false
	}
	var e *NetworkError
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}
