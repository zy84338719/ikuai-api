package ikuaiapi

// Wire-format handling for iKuai v4 responses: the nil-sanitizer and the
// response-envelope contract (code/data/results/rowid unwrapping).

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

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

// envelopeCodeOK reports whether the router's business code means success.
// Firmware uses 0 in most responses and 20000 on some monitor endpoints.
func envelopeCodeOK(code int) bool {
	return code == 0 || code == 20000
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
// maxResponseBody caps how much of a response is read into memory. iKuai
// responses are kilobytes; anything near this ceiling is a hostile or
// broken endpoint, and an unbounded io.ReadAll hands a remote party a
// lever on this process's heap. Same ceiling and rationale as
// oekit/pkg/httpjson's maxBody ("异常服务端刷屏防内存放大").
const maxResponseBody = 64 << 20

// retryAfter carries the server-advised back-off hint captured by
// RetryAfterTransport for this response (0 when absent/unparseable).
func (c *Client) check(resp *http.Response, retryAfter time.Duration) (json.RawMessage, error) {
	// Read one byte past the cap so a body landing exactly on the cap is
	// not mistaken for one over it (the probe byte is dropped by the
	// length check below — same trick as httpjson.DoRaw).
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody+1))
	if err != nil {
		return nil, NewError("read response body: "+err.Error(), WithHTTPStatus(resp.StatusCode), WithCause(err))
	}
	if len(raw) > maxResponseBody {
		// Fail fast instead of parsing a truncated body: half a JSON
		// document surfaces as a baffling "non-JSON response". The
		// connection is dropped unread (deferred Close above) — this is
		// the exceptional path, not worth draining for reuse.
		return nil, NewError(
			fmt.Sprintf("response body exceeds %d bytes", maxResponseBody),
			WithHTTPStatus(resp.StatusCode),
		)
	}
	body := SanitizeNil(raw)

	if resp.StatusCode >= 400 {
		return nil, httpStatusError(resp, body, retryAfter)
	}

	var env envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, NewError("non-JSON response: "+summarizeBody(body), WithHTTPStatus(resp.StatusCode))
	}

	if !envelopeCodeOK(env.Code) {
		return nil, envelopeError(resp.StatusCode, env, body)
	}

	if c.RawMode {
		return body, nil
	}
	return unwrapEnvelope(env), nil
}

// httpStatusError builds the typed error for a non-2xx response, preferring
// the envelope's message/code/details over the raw body.
func httpStatusError(resp *http.Response, body []byte, retryAfter time.Duration) error {
	msg := summarizeBody(body)
	var env envelope
	_ = json.Unmarshal(body, &env)
	if env.Message != "" {
		msg = env.Message
	}
	var detailsEnv struct {
		Details []APIErrorDetail `json:"details"`
	}
	_ = json.Unmarshal(body, &detailsEnv)
	return NewError(msg,
		WithHTTPStatus(resp.StatusCode),
		WithCode(env.Code),
		WithDetails(detailsEnv.Details),
		WithRetryAfter(retryAfter),
	)
}

// summarizeBody bounds how much of a raw response body can enter an error
// message: at most 200 runes (cut on a rune boundary, so a multi-byte
// character is never split) followed by the original size, which keeps the
// truncation observable. A whole body here would ride along into every log
// line that prints the error — httpjson's DoJSON keeps a 200-rune summary
// for the same reason. An empty body stays empty: " (0 bytes)" would be
// noise where the caller expects no message at all.
func summarizeBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	const summaryRunes = 200
	runes := []rune(string(body))
	if len(runes) <= summaryRunes {
		return fmt.Sprintf("%s (%d bytes)", string(runes), len(body))
	}
	return fmt.Sprintf("%s… (%d bytes)", string(runes[:summaryRunes]), len(body))
}

// envelopeError builds the typed error for a 2xx response whose business
// code reports failure.
func envelopeError(status int, env envelope, body []byte) error {
	msg := env.Message
	if msg == "" {
		msg = "request failed"
	}
	var detailsEnv struct {
		Details []APIErrorDetail `json:"details"`
	}
	_ = json.Unmarshal(body, &detailsEnv)
	return NewError(msg,
		WithHTTPStatus(status),
		WithCode(env.Code),
		WithDetails(detailsEnv.Details),
	)
}

// unwrapEnvelope applies the payload precedence: Data first, then Results,
// then a synthetic {"message"[,"rowid"]} object for create-style endpoints
// whose only meaningful response is the new row id.
func unwrapEnvelope(env envelope) json.RawMessage {
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
				return synthetic
			}
		}
		synthetic, _ := json.Marshal(map[string]any{"message": msg})
		return synthetic
	}
	return payload
}
