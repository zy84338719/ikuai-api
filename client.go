// Package ikuaiapi provides a Go SDK for interacting with iKuai routers
// using the local v4.0 REST API.
//
// The SDK uses the Go standard library net/http with a small custom layer
// (retry + timeout + sanitization) and exposes a typed service layer per
// functional area (system, network, firewall, monitor, ...).
//
// Authentication uses a Bearer token obtained from the router web UI
// (System → Auth → API Token). iKuai OS v4.x exposes all router
// configuration under /api/v4.0/*.
//
// Basic usage:
//
//	client, err := ikuaiapi.NewClient("https://192.168.1.1",
//	    ikuaiapi.WithToken("<router-api-token>"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	api := service.NewAPIClient(client)
//	iface, err := api.Network().GetInterfaces(ctx)
package ikuaiapi

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

// V4APIBase is the canonical root for all iKuai v4 REST endpoints.
const V4APIBase = "/api/v4.0"

// defaultTimeout is the per-request timeout when none is configured.
const defaultTimeout = 30 * time.Second

// defaultRetryMax is the default maximum number of attempts (initial + retries).
const defaultRetryMax = 3

// defaultRetryBaseDelay is the base delay for the first retry.
const defaultRetryBaseDelay = 200 * time.Millisecond

// defaultRetryMaxDelay caps the back-off window.
const defaultRetryMaxDelay = 5 * time.Second

// Client is the iKuai HTTP API client.
type Client struct {
	BaseURL    string
	Token      string
	APIBase    string
	HTTPClient *http.Client
	UserAgent  string

	// RawMode returns the full JSON envelope (data/results/rowid/code/message)
	// instead of just the data field. Useful for debugging.
	RawMode bool
	// DryRun reports the request it would have made without contacting the
	// router. Read methods return the preview as a JSON object, write
	// methods return without executing.
	DryRun bool
	// Logger, if set, receives short human-readable status lines.
	Logger func(format string, args ...any)

	timeout    time.Duration
	retryMax   int
	retryBase  time.Duration
	retryMaxD  time.Duration
	apiBaseURL *url.URL
	metrics    *Metrics         // optional request counters/latency
	slogger    Logger           // optional structured logger
}

// ClientOption configures a Client at construction time.
type ClientOption func(*Client)

// WithToken sets the Bearer token used on every request.
func WithToken(token string) ClientOption {
	return func(c *Client) { c.Token = token }
}

// WithTimeout sets the per-request timeout. The same value is also used
// as the overall upper bound for retried requests.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) { c.timeout = d }
}

// WithInsecureSkipVerify disables TLS certificate verification. iKuai
// routers use self-signed certificates by default, so this is normally
// the desired behaviour. Use only on trusted networks.
func WithInsecureSkipVerify(skip bool) ClientOption {
	return func(c *Client) {
		if !skip {
			return
		}
		if c.HTTPClient == nil {
			c.HTTPClient = &http.Client{Timeout: defaultTimeout}
		}
		if tr, ok := c.HTTPClient.Transport.(*http.Transport); ok {
			if tr.TLSClientConfig == nil {
				tr.TLSClientConfig = &tls.Config{}
			}
			tr.TLSClientConfig.InsecureSkipVerify = true //nolint:gosec
		} else if c.HTTPClient.Transport == nil {
			c.HTTPClient.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			}
		}
	}
}

// WithHTTPClient replaces the underlying *http.Client. Callers that need
// proxy, custom CA, or tracing support can pass their own.
func WithHTTPClient(h *http.Client) ClientOption {
	return func(c *Client) { c.HTTPClient = h }
}

// WithAPIBase overrides the default /api/v4.0 prefix.
func WithAPIBase(base string) ClientOption {
	return func(c *Client) {
		c.APIBase = "/" + strings.TrimLeft(strings.TrimSpace(base), "/")
	}
}

// WithRawMode enables envelope-level responses (see Client.RawMode).
func WithRawMode(raw bool) ClientOption { return func(c *Client) { c.RawMode = raw } }

// WithDryRun reports the request it would have made without contacting
// the router.
func WithDryRun(dry bool) ClientOption { return func(c *Client) { c.DryRun = dry } }

// WithRetry configures exponential-back-off retries. retryMax is the
// total attempt count (initial + retries). The default is 3.
func WithRetry(retryMax int) ClientOption {
	return func(c *Client) {
		if retryMax < 1 {
			retryMax = 1
		}
		c.retryMax = retryMax
	}
}

// WithRetryDelay sets the base delay and maximum delay for retries.
func WithRetryDelay(base, max time.Duration) ClientOption {
	return func(c *Client) {
		if base > 0 {
			c.retryBase = base
		}
		if max > 0 {
			c.retryMaxD = max
		}
	}
}

// WithLogger sets a logging callback. The callback is invoked once per
// request with a short status line. Prefer WithStructuredLogger for new code.
func WithLogger(fn func(format string, args ...any)) ClientOption {
	return func(c *Client) { c.Logger = fn }
}

// WithStructuredLogger attaches a leveled, structured Logger (see logger.go).
// When set, retry / timeout / token-failure events are emitted through it
// instead of the printf-style Logger callback.
func WithStructuredLogger(l Logger) ClientOption {
	return func(c *Client) { c.slogger = l }
}

// WithMetrics attaches a Metrics collector. When set, every request records
// its duration and outcome (see Metrics.RecordRequest). Use GetStats to read
// counters, e.g. for a /metrics endpoint or health check.
func WithMetrics(m *Metrics) ClientOption {
	return func(c *Client) { c.metrics = m }
}

// Metrics returns the attached Metrics collector, or nil if none was set.
func (c *Client) Metrics() *Metrics { return c.metrics }

// NewClient creates a Client targeting the given router. baseURL should
// be of the form "http://192.168.1.1" or "https://router.lan:443".
func NewClient(baseURL string, opts ...ClientOption) (*Client, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, errors.New("ikuaiapi: baseURL is required")
	}
	apiBase, err := url.Parse(baseURL + V4APIBase)
	if err != nil {
		return nil, fmt.Errorf("ikuaiapi: parse baseURL: %w", err)
	}
	c := &Client{
		BaseURL: baseURL,
		APIBase: V4APIBase,
		HTTPClient: &http.Client{
			Timeout: defaultTimeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			},
		},
		UserAgent:  "ikuai-api-go/4",
		timeout:    defaultTimeout,
		retryMax:   defaultRetryMax,
		retryBase:  defaultRetryBaseDelay,
		retryMaxD:  defaultRetryMaxDelay,
		apiBaseURL: apiBase,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// Close releases the underlying transport. Safe to call multiple times.
func (c *Client) Close() {
	if c.HTTPClient != nil && c.HTTPClient.Transport != nil {
		if t, ok := c.HTTPClient.Transport.(*http.Transport); ok {
			t.CloseIdleConnections()
		}
	}
}

// SanitizeNil replaces bare `nil` tokens in JSON value positions with
// `null`. Some iKuai firmware emits `nil` instead of `null`; the function
// tracks string state to avoid corrupting legitimate string content.
func SanitizeNil(body []byte) []byte {
	body = bytes.ReplaceAll(body, []byte("\r\n"), []byte("\n"))
	n := len(body)
	if n < 3 {
		return body
	}
	out := make([]byte, 0, n+8)
	inString := false
	for i := 0; i < n; i++ {
		ch := body[i]
		if ch == '\\' && inString && i+1 < n {
			out = append(out, ch, body[i+1])
			i++
			continue
		}
		if ch == '"' {
			inString = !inString
			out = append(out, ch)
			continue
		}
		if inString {
			out = append(out, ch)
			continue
		}
		if ch == 'n' && i+2 < n && body[i+1] == 'i' && body[i+2] == 'l' {
			nextOK := i+3 >= n || !isIdentByte(body[i+3])
			if nextOK {
				out = append(out, 'n', 'u', 'l', 'l')
				i += 2
				continue
			}
		}
		out = append(out, ch)
	}
	return out
}

func isIdentByte(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '_'
}

type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Results json.RawMessage `json:"results"`
	RowID   json.RawMessage `json:"rowid"`
}

// check inspects the HTTP response, normalizes the body, and returns the
// payload callers should use (Data preferred, then Results, then a
// synthetic envelope carrying RowID/message for create-style endpoints).
func (c *Client) check(resp *http.Response) (json.RawMessage, error) {
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &APIError{HTTPStatus: resp.StatusCode, Message: "read response body: " + err.Error()}
	}
	body := SanitizeNil(raw)

	if resp.StatusCode >= 400 {
		msg := string(body)
		var env envelope
		_ = json.Unmarshal(body, &env)
		if env.Message != "" {
			msg = env.Message
		}
		if hint, ok := errorHints[env.Code]; ok {
			msg = msg + " (" + hint + ")"
		}
		var detailsEnv struct {
			Details []APIErrorDetail `json:"details"`
		}
		_ = json.Unmarshal(body, &detailsEnv)
		return nil, &APIError{
			HTTPStatus: resp.StatusCode,
			Code:       env.Code,
			Message:    msg,
			Details:    detailsEnv.Details,
			RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After")),
		}
	}

	var env envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, &APIError{
			HTTPStatus: resp.StatusCode,
			Message:    "non-JSON response: " + string(body),
		}
	}

	if env.Code != 0 && env.Code != 20000 {
		msg := env.Message
		if msg == "" {
			msg = "request failed"
		}
		if hint, ok := errorHints[env.Code]; ok {
			msg = msg + " (" + hint + ")"
		}
		var detailsEnv struct {
			Details []APIErrorDetail `json:"details"`
		}
		_ = json.Unmarshal(body, &detailsEnv)
		return nil, &APIError{
			HTTPStatus: resp.StatusCode,
			Code:       env.Code,
			Message:    msg,
			Details:    detailsEnv.Details,
		}
	}

	if c.RawMode {
		return body, nil
	}

	payload := env.Data
	if (len(payload) == 0 || string(payload) == "null") &&
		len(env.Results) > 0 && string(env.Results) != "null" {
		payload = env.Results
	}
	if len(payload) == 0 || string(payload) == "null" {
		msg := env.Message
		if msg == "" {
			msg = "ok"
		}
		if len(env.RowID) > 0 && string(env.RowID) != "null" {
			var rowid any
			if err := json.Unmarshal(env.RowID, &rowid); err == nil {
				synthetic, _ := json.Marshal(map[string]any{
					"message": msg,
					"rowid":   rowid,
				})
				return synthetic, nil
			}
		}
		synthetic, _ := json.Marshal(map[string]any{"message": msg})
		return synthetic, nil
	}
	return payload, nil
}

func (c *Client) headers() http.Header {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	h.Set("Accept", "application/json")
	if c.UserAgent != "" {
		h.Set("User-Agent", c.UserAgent)
	}
	if c.Token != "" {
		h.Set("Authorization", "Bearer "+c.Token)
	}
	return h
}

func (c *Client) fullURL(p string) (string, error) {
	if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
		return p, nil
	}
	rel := strings.TrimLeft(p, "/")
	if !strings.HasPrefix(rel, "api/v4.0/") && !strings.HasPrefix(rel, "api/") {
		rel = strings.TrimLeft(c.APIBase, "/") + "/" + rel
	}
	u := *c.apiBaseURL
	u.Path = path.Clean("/" + strings.SplitN(rel, "?", 2)[0])
	if i := strings.Index(rel, "?"); i >= 0 {
		u.RawQuery = rel[i+1:]
	}
	return u.String(), nil
}

func (c *Client) log(format string, args ...any) {
	if c.Logger != nil {
		c.Logger(format, args...)
	}
}

// logf logs at the given structured level. Falls back to the printf Logger
// when no structured logger is attached.
func (c *Client) logf(level LogLevel, msg string, args ...any) {
	if c.slogger != nil {
		switch level {
		case LogLevelDebug:
			c.slogger.Debug(msg, args...)
		case LogLevelInfo:
			c.slogger.Info(msg, args...)
		case LogLevelWarn:
			c.slogger.Warn(msg, args...)
		default:
			c.slogger.Error(msg, args...)
		}
		return
	}
	c.log(msg, args...)
}

// Do executes a typed REST call and decodes the result into out (which
// may be nil for requests that only return a rowid/message).
func (c *Client) Do(ctx context.Context, method, p string, body, out any) error {
	if c.DryRun {
		preview := map[string]any{
			"dry_run": true,
			"method":  strings.ToUpper(method),
			"path":    p,
		}
		if body != nil {
			bb, _ := json.Marshal(body)
			preview["body"] = json.RawMessage(bb)
		}
		raw, _ := json.Marshal(preview)
		if out != nil {
			return json.Unmarshal(raw, out)
		}
		return nil
	}

	fullURL, err := c.fullURL(p)
	if err != nil {
		return err
	}
	payload, err := c.doWithRetry(ctx, method, fullURL, body)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	if len(payload) == 0 || string(payload) == "null" {
		return nil
	}
	if err := json.Unmarshal(payload, out); err != nil {
		return &APIError{HTTPStatus: 200, Message: "decode response: " + err.Error()}
	}
	return nil
}

// Get issues a GET request. params is optional and added to the query
// string.
func (c *Client) Get(ctx context.Context, p string, params map[string]string) (json.RawMessage, error) {
	if c.DryRun {
		preview := map[string]any{
			"dry_run": true,
			"method":  "GET",
			"path":    p,
		}
		if len(params) > 0 {
			q := url.Values{}
			for k, v := range params {
				q.Set(k, v)
			}
			preview["query"] = q.Encode()
		}
		return json.Marshal(preview)
	}
	if len(params) > 0 {
		q := url.Values{}
		for k, v := range params {
			q.Set(k, v)
		}
		sep := "?"
		if strings.Contains(p, "?") {
			sep = "&"
		}
		p = p + sep + q.Encode()
	}
	fullURL, err := c.fullURL(p)
	if err != nil {
		return nil, err
	}
	return c.doWithRetry(ctx, http.MethodGet, fullURL, nil)
}

// Post issues a POST request with a JSON body.
func (c *Client) Post(ctx context.Context, p string, body any) (json.RawMessage, error) {
	return c.doPayload(ctx, http.MethodPost, p, body)
}

// Put issues a PUT request with a JSON body.
func (c *Client) Put(ctx context.Context, p string, body any) (json.RawMessage, error) {
	return c.doPayload(ctx, http.MethodPut, p, body)
}

// Patch issues a PATCH request with a JSON body.
func (c *Client) Patch(ctx context.Context, p string, body any) (json.RawMessage, error) {
	return c.doPayload(ctx, http.MethodPatch, p, body)
}

// Delete issues a DELETE request. The optional body is sent as JSON.
func (c *Client) Delete(ctx context.Context, p string, body any) (json.RawMessage, error) {
	return c.doPayload(ctx, http.MethodDelete, p, body)
}

// FormatQuery is exported for callers that need to assemble a query
// string from a map (e.g. the Call escape hatch appends ?key=value to
// the path for DELETE requests that iKuai drives with query params).
func (c *Client) FormatQuery(q map[string]string) string {
	if len(q) == 0 {
		return ""
	}
	uv := make(url.Values)
	for k, v := range q {
		uv.Set(k, v)
	}
	return uv.Encode()
}

func (c *Client) doPayload(ctx context.Context, method, p string, body any) (json.RawMessage, error) {
	if c.DryRun {
		preview := map[string]any{
			"dry_run": true,
			"method":  method,
			"path":    p,
		}
		if body != nil {
			bb, _ := json.Marshal(body)
			preview["body"] = json.RawMessage(bb)
		}
		return json.Marshal(preview)
	}
	fullURL, err := c.fullURL(p)
	if err != nil {
		return nil, err
	}
	return c.doWithRetry(ctx, method, fullURL, body)
}

// doWithRetry runs doRequest up to retryMax times, applying exponential
// back-off with jitter on transient failures.
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
		// Honour a server-advised Retry-After when present (429/503).
		if apiErr, ok := isAPIError(err); ok && apiErr.RetryAfter > 0 {
			delay = apiErr.RetryAfter
		}
		c.logf(LogLevelWarn, "retry %d/%d after %s: %v", attempt+1, c.retryMax, delay, err)
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
	if errors.Is(err, context.Canceled) {
		return false
	}
	var netErr *NetworkError
	if errors.As(err, &netErr) {
		// Only retry transport failures on idempotent verbs.
		return idempotent
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.IsRetryable()
	}
	return false
}

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
	// Full jitter (0..d).
	if d > 0 {
		d = time.Duration(rand.Int63n(int64(d) + 1)) //nolint:gosec
	}
	return d
}

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

func (c *Client) doOnce(ctx context.Context, method, fullURL string, body any) (json.RawMessage, error) {
	var bodyBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, &APIError{Message: "encode request body: " + err.Error()}
		}
		bodyBytes = b
	} else {
		bodyBytes = []byte("{}")
	}

	reqCtx := ctx
	if c.timeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}
	req, err := http.NewRequestWithContext(reqCtx, method, fullURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, &APIError{Message: "build request: " + err.Error()}
	}
	req.Header = c.headers()

	c.logf(LogLevelDebug, "%s %s", method, fullURL)
	start := time.Now()
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		dur := time.Since(start)
		c.recordMetrics(dur, true)
		return nil, &NetworkError{Message: "request failed", Cause: err}
	}
	defer func() { _ = resp.Body.Close() }()
	payload, err := c.check(resp)
	c.recordMetrics(time.Since(start), err != nil)
	return payload, err
}

// recordMetrics records the request if a Metrics collector is attached.
func (c *Client) recordMetrics(dur time.Duration, hasError bool) {
	if c.metrics != nil {
		c.metrics.RecordRequest(dur, hasError)
	}
}
