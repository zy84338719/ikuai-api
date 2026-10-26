package ikuaiapi

// The retry engine: attempt loop, retryability judgement, one-sided-jitter
// exponential back-off, and Retry-After handling.

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	oekerrors "github.com/oekit/pkg/errors"
	"github.com/oekit/pkg/httpjson"
	"github.com/oekit/pkg/mathx"
)

// defaultRetryMax is the default maximum number of attempts (initial + retries).
const defaultRetryMax = 3

// defaultRetryBaseDelay is the base delay for the first retry.
const defaultRetryBaseDelay = 200 * time.Millisecond

// defaultRetryMaxDelay caps the back-off window.
const defaultRetryMaxDelay = 5 * time.Second

// doWithRetry runs doOnce up to retryMax times, applying exponential
// back-off with one-sided jitter on transient failures.
func (c *Client) doWithRetry(ctx context.Context, method, fullURL string, body any) (json.RawMessage, error) {
	idem := method == http.MethodGet || method == http.MethodHead ||
		method == http.MethodOptions || method == http.MethodDelete
	var lastErr error
	for attempt := 1; attempt <= c.retryMax; attempt++ {
		raw, err := c.doOnce(ctx, method, fullURL, body)
		if err == nil {
			return raw, nil
		}
		lastErr = err
		if !shouldRetry(err, idem) || attempt == c.retryMax {
			break
		}
		delay := backoffDelay(attempt, c.retryBase, c.retryMaxD)
		// Honour a server-advised Retry-After (429/503), clamped by the oekit
		// rule: a hint is trusted only when it is ≤ MaxServerRetryAfter (5min)
		// or shorter than our own curve — a pathological header can't park a
		// client for hours and a retry storm can't resynchronise.
		if apiErr, ok := isAPIError(err); ok && apiErr.RetryAfter > 0 {
			delay = httpjson.SelectRetryDelay(delay, apiErr.RetryAfter)
		}
		c.logCtx(ctx, LogLevelWarn, "retrying request",
			"attempt", attempt+1, "max_attempts", c.retryMax,
			"delay", delay.String(), "err", err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}
	return nil, lastErr
}

// shouldRetry decides whether err warrants another attempt. Network errors
// are retried only for idempotent verbs — a write that failed at the transport
// layer may still have reached the router, so retrying it risks a duplicate.
func shouldRetry(err error, idempotent bool) bool {
	if err == nil {
		return false
	}
	if oekerrors.Is(err, context.Canceled) {
		return false
	}
	var netErr *NetworkError
	if oekerrors.As(err, &netErr) {
		// Only retry transport failures on idempotent verbs.
		return idempotent
	}
	var apiErr *APIError
	if oekerrors.As(err, &apiErr) {
		return apiErr.IsRetryable()
	}
	return false
}

// backoffDelay returns how long to wait before attempt+1.
//
// The jitter is deliberately one-sided. A symmetric or full jitter picks a
// value anywhere in [0, d], which means a retry scheduled for a 4-second
// back-off can fire after 3 milliseconds — and when a router is failing
// under load, every client doing that at once turns into a retry storm
// precisely when the device is least able to absorb it. mathx.Unstable's
// upward jitter keeps the delay in [d, d*(1+dev)]: never earlier than the
// exponential baseline, still spread enough that clients do not resynchronise.
func backoffDelay(attempt int, base, max time.Duration) time.Duration {
	if base <= 0 {
		base = defaultRetryBaseDelay
	}
	if max <= 0 {
		max = defaultRetryMaxDelay
	}
	d := time.Duration(float64(base) * math.Pow(2, float64(attempt-1)))
	if d > max {
		d = max
	}
	if d <= 0 {
		return 0
	}
	return retryJitter.UpwardDuration(d)
}

// retryJitter spreads concurrent clients' back-off without ever pulling a
// retry earlier than its exponential baseline. The deviation is modest: it
// buys desynchronisation without materially inflating the wait.
var retryJitter = mathx.NewUnstable(retryJitterDeviation)

// retryJitterDeviation is the fraction added to the baseline at most.
const retryJitterDeviation = 0.25

// parseRetryAfter parses an HTTP Retry-After header value, which may be either
// a delta-seconds integer or an HTTP-date. Returns 0 when the header is absent
// or unparseable.
func parseRetryAfter(v string) time.Duration {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0
	}
	// Integer seconds form.
	if n, err := strconv.Atoi(v); err == nil && n > 0 {
		return time.Duration(n) * time.Second
	}
	// HTTP-date form.
	if t, err := http.ParseTime(v); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}
