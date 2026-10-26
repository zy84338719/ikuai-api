package ikuaiapi

// Wire-format handling for iKuai v4 responses: the nil-sanitizer and the
// response-envelope contract (code/data/results/rowid unwrapping).

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
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
func (c *Client) check(resp *http.Response) (json.RawMessage, error) {
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewError("read response body: "+err.Error(), WithHTTPStatus(resp.StatusCode), WithCause(err))
	}
	body := SanitizeNil(raw)

	if resp.StatusCode >= 400 {
		return nil, httpStatusError(resp, body)
	}

	var env envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, NewError("non-JSON response: "+string(body), WithHTTPStatus(resp.StatusCode))
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
func httpStatusError(resp *http.Response, body []byte) error {
	msg := string(body)
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
		WithRetryAfter(parseRetryAfter(resp.Header.Get("Retry-After"))),
	)
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
