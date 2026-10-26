// Package-level client wiring: the Client type, its options, and the verb
// methods. Response handling lives in envelope.go, the retry engine in
// retry.go, dry-run previews in dryrun.go.
package ikuaiapi

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/oekit/logx"
	oekerrors "github.com/oekit/pkg/errors"
)

// V4APIBase is the canonical root for all iKuai v4 REST endpoints.
const V4APIBase = "/api/v4.0"

// defaultTimeout is the per-request timeout when none is configured.
const defaultTimeout = 30 * time.Second

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

	timeout    time.Duration
	retryMax   int
	retryBase  time.Duration
	retryMaxD  time.Duration
	apiBaseURL *url.URL
	metrics    *Metrics    // optional request counters/latency
	slogger    logx.Logger // optional structured logger
	closeOnce  closeOnce   // guards Close against repeated calls
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
// total attempt count (initial + retries). The default is 3. See retry.go
// for the retryability rules and the back-off curve.
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

// WithStructuredLogger attaches a structured logger for this client only.
//
// The default is the package-level Delegating logger (see logger.go),
// which follows whatever logx.SetDefault installed. Attaching a logger
// here narrows it to this client — useful in a library that serves several
// routers with different verbosity — without touching the global facade.
func WithStructuredLogger(l logx.Logger) ClientOption {
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
		return nil, ErrNotConfigured
	}
	apiBase, err := url.Parse(baseURL + V4APIBase)
	if err != nil {
		return nil, oekerrors.Wrap(err, "ikuaiapi: parse baseURL")
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

// Close releases the underlying transport and makes the client unusable.
//
// It follows the oekit lifecycle contract: the ctx is accepted so a caller
// can bound the shutdown, the error is returned rather than swallowed, and
// repeated or concurrent calls are safe — a second Close returns the first
// one's result without doing the work again.
//
// Close does not cancel in-flight requests; it closes idle connections so
// the process can exit promptly. To bound a request, pass a context with a
// deadline to the call itself.
func (c *Client) Close(ctx context.Context) error {
	c.closeOnce.once.Do(func() {
		if c.HTTPClient == nil {
			return
		}
		transport, ok := c.HTTPClient.Transport.(interface{ CloseIdleConnections() })
		if !ok {
			return
		}
		transport.CloseIdleConnections()
		if err := ctx.Err(); err != nil {
			c.closeOnce.err = oekerrors.Wrap(err, "ikuai-api: close")
		}
	})
	return c.closeOnce.err
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

// logCtx emits a structured event at the given level, carrying the context
// so logx can attach a trace id, and falling back to the package logger when
// no per-client logger is attached.
func (c *Client) logCtx(ctx context.Context, level LogLevel, msg string, kv ...any) {
	l := c.slogger
	if l == nil {
		l = log()
	}
	switch level {
	case LogLevelDebug:
		l.CtxDebug(ctx, msg, kv...)
	case LogLevelInfo:
		l.CtxInfo(ctx, msg, kv...)
	case LogLevelWarn:
		l.CtxWarn(ctx, msg, kv...)
	default:
		l.CtxError(ctx, msg, kv...)
	}
}

// Do executes a typed REST call and decodes the result into out (which
// may be nil for requests that only return a rowid/message).
func (c *Client) Do(ctx context.Context, method, p string, body, out any) error {
	if c.DryRun {
		if out == nil {
			return nil
		}
		return json.Unmarshal(dryRunPreview(method, p, body, nil), out)
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
		return NewError("decode response: "+err.Error(), WithHTTPStatus(200), WithCause(err))
	}
	return nil
}

// Get issues a GET request. params is optional and added to the query
// string.
func (c *Client) Get(ctx context.Context, p string, params map[string]string) (json.RawMessage, error) {
	if c.DryRun {
		return dryRunPreview(http.MethodGet, p, nil, params), nil
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

// FormatQuery assembles a query string from a map. It backs the Call escape
// hatch, which appends ?key=value to the path for DELETE requests that iKuai
// drives with query params.
func (c *Client) FormatQuery(q map[string]string) string {
	return formatQuery(q)
}

func formatQuery(q map[string]string) string {
	if len(q) == 0 {
		return ""
	}
	uv := make(url.Values, len(q))
	for k, v := range q {
		uv.Set(k, v)
	}
	return uv.Encode()
}

func (c *Client) doPayload(ctx context.Context, method, p string, body any) (json.RawMessage, error) {
	if c.DryRun {
		return dryRunPreview(method, p, body, nil), nil
	}
	fullURL, err := c.fullURL(p)
	if err != nil {
		return nil, err
	}
	return c.doWithRetry(ctx, method, fullURL, body)
}

func (c *Client) doOnce(ctx context.Context, method, fullURL string, body any) (json.RawMessage, error) {
	var bodyBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, NewError("encode request body: "+err.Error(), WithCause(err))
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
		return nil, NewError("build request: "+err.Error(), WithCause(err))
	}
	req.Header = c.headers()

	c.logCtx(reqCtx, LogLevelDebug, "sending request", "method", method, "url", fullURL)
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
