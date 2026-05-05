package ikuaisdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"github.com/zy84338719/ikuai-api/internal"
)

const V4APIBase = "/api/v4.0"

type V4Client struct {
	client  *req.Client
	baseURL string
	token   string
	apiBase string
	rawMode bool
	dryRun  bool
}

type V4Option func(*V4Client)

func WithV4APIBase(apiBase string) V4Option {
	return func(c *V4Client) {
		c.apiBase = "/" + strings.Trim(apiBase, "/")
	}
}

func WithV4RawMode(raw bool) V4Option {
	return func(c *V4Client) {
		c.rawMode = raw
	}
}

func WithV4DryRun(dryRun bool) V4Option {
	return func(c *V4Client) {
		c.dryRun = dryRun
	}
}

func WithV4HTTPClient(httpClient *req.Client) V4Option {
	return func(c *V4Client) {
		c.client = httpClient
	}
}

func NewV4RESTClient(baseURL, token string, opts ...V4Option) *V4Client {
	baseURL = internal.NormalizeAddr(baseURL)
	c := &V4Client{
		baseURL: baseURL,
		token:   token,
		apiBase: V4APIBase,
		client: req.C().
			SetBaseURL(baseURL).
			SetTimeout(15 * time.Second).
			EnableInsecureSkipVerify(),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) V4REST(token ...string) *V4Client {
	tok := c.token
	if len(token) > 0 {
		tok = token[0]
	}
	return NewV4RESTClient(c.baseURL, tok, WithV4HTTPClient(c.client))
}

type V4APIError struct {
	Code    int
	Message string
	Details []V4APIErrorDetail
}

type V4APIErrorDetail struct {
	Field string `json:"field"`
	Type  string `json:"type"`
	Msg   string `json:"msg"`
}

func (e *V4APIError) Error() string {
	msg := fmt.Sprintf("[v4 API %d] %s", e.Code, e.Message)
	for _, d := range e.Details {
		msg += fmt.Sprintf("\n  - %s: %s", d.Field, d.Msg)
	}
	return msg
}

type V4Envelope struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    json.RawMessage    `json:"data"`
	Results json.RawMessage    `json:"results"`
	RowID   json.RawMessage    `json:"rowid"`
	Details []V4APIErrorDetail `json:"details,omitempty"`
}

type V4ListParams struct {
	Page        int
	PageSize    int
	PageSizeKey string
	Filter      string
	Order       string
	OrderBy     string
	Key         string
	Pattern     string
	Extra       map[string]string
}

func (p V4ListParams) Query() map[string]string {
	q := map[string]string{}
	if p.Page > 0 {
		q["page"] = fmt.Sprint(p.Page)
	}
	if p.PageSize > 0 {
		key := p.PageSizeKey
		if key == "" {
			key = "page_size"
		}
		q[key] = fmt.Sprint(p.PageSize)
	}
	if p.Filter != "" {
		q["filter"] = p.Filter
	}
	if p.Order != "" {
		q["order"] = p.Order
	}
	if p.OrderBy != "" {
		q["order_by"] = p.OrderBy
	}
	if p.Key != "" {
		q["key"] = p.Key
	}
	if p.Pattern != "" {
		q["pattern"] = p.Pattern
	}
	for k, v := range p.Extra {
		q[k] = v
	}
	return q
}

func (c *V4Client) GetRaw(ctx context.Context, path string, params map[string]string) (json.RawMessage, error) {
	return c.do(ctx, http.MethodGet, path, params, nil)
}

func (c *V4Client) PostRaw(ctx context.Context, path string, body interface{}) (json.RawMessage, error) {
	return c.do(ctx, http.MethodPost, path, nil, body)
}

func (c *V4Client) PutRaw(ctx context.Context, path string, body interface{}) (json.RawMessage, error) {
	return c.do(ctx, http.MethodPut, path, nil, body)
}

func (c *V4Client) PatchRaw(ctx context.Context, path string, body interface{}) (json.RawMessage, error) {
	return c.do(ctx, http.MethodPatch, path, nil, body)
}

func (c *V4Client) DeleteRaw(ctx context.Context, path string) (json.RawMessage, error) {
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

func (c *V4Client) Get(ctx context.Context, path string, params map[string]string, result interface{}) error {
	raw, err := c.GetRaw(ctx, path, params)
	return decodeV4Raw(raw, result, err)
}

func (c *V4Client) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	raw, err := c.PostRaw(ctx, path, body)
	return decodeV4Raw(raw, result, err)
}

func (c *V4Client) Put(ctx context.Context, path string, body interface{}, result interface{}) error {
	raw, err := c.PutRaw(ctx, path, body)
	return decodeV4Raw(raw, result, err)
}

func (c *V4Client) Patch(ctx context.Context, path string, body interface{}, result interface{}) error {
	raw, err := c.PatchRaw(ctx, path, body)
	return decodeV4Raw(raw, result, err)
}

func (c *V4Client) Delete(ctx context.Context, path string, result interface{}) error {
	raw, err := c.DeleteRaw(ctx, path)
	return decodeV4Raw(raw, result, err)
}

func (c *V4Client) do(ctx context.Context, method, path string, params map[string]string, body interface{}) (json.RawMessage, error) {
	fullPath := c.fullPath(path)
	if c.dryRun {
		preview := map[string]interface{}{
			"dry_run": true,
			"method":  method,
			"url":     c.baseURL + fullPath,
		}
		if len(params) > 0 {
			preview["query"] = params
		}
		if body != nil {
			preview["body"] = body
		}
		return json.Marshal(preview)
	}

	req := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json")
	if c.token != "" {
		req.SetHeader("Authorization", "Bearer "+c.token)
	}
	if len(params) > 0 {
		req.SetQueryParams(params)
	}
	if body != nil {
		req.SetBody(body)
	} else if method != http.MethodGet {
		req.SetBody(map[string]string{})
	}

	resp, err := req.Send(method, fullPath)
	if err != nil {
		return nil, NewSDKError(ErrCodeRequestFailed, "v4 REST request failed", err)
	}
	if resp.Err != nil {
		return nil, NewSDKError(ErrCodeRequestFailed, "v4 REST response error", resp.Err)
	}

	return c.check(resp.StatusCode, resp.Bytes())
}

func (c *V4Client) fullPath(path string) string {
	path = "/" + strings.TrimLeft(path, "/")
	if strings.HasPrefix(path, c.apiBase+"/") || path == c.apiBase {
		return path
	}
	return c.apiBase + path
}

func (c *V4Client) check(statusCode int, raw []byte) (json.RawMessage, error) {
	body := sanitizeV4Nil(raw)
	var env V4Envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("HTTP %d: non-JSON response", statusCode)
	}

	if statusCode >= 400 || (env.Code != 0 && env.Code != 20000) {
		msg := env.Message
		if msg == "" {
			msg = string(body)
		}
		return nil, &V4APIError{Code: env.Code, Message: msg, Details: env.Details}
	}
	if c.rawMode {
		return body, nil
	}

	payload := env.Data
	if isEmptyJSON(payload) && !isEmptyJSON(env.Results) {
		payload = env.Results
	}
	if isEmptyJSON(payload) {
		msg := env.Message
		if msg == "" {
			msg = "ok"
		}
		result := map[string]interface{}{"message": msg}
		if !isEmptyJSON(env.RowID) {
			var rowid interface{}
			if err := json.Unmarshal(env.RowID, &rowid); err == nil {
				result["rowid"] = rowid
			}
		}
		return json.Marshal(result)
	}
	return payload, nil
}

func decodeV4Raw(raw json.RawMessage, result interface{}, err error) error {
	if err != nil {
		return err
	}
	if result == nil {
		return nil
	}
	return json.Unmarshal(raw, result)
}

func isEmptyJSON(raw json.RawMessage) bool {
	raw = bytes.TrimSpace(raw)
	return len(raw) == 0 || bytes.Equal(raw, []byte("null"))
}

func sanitizeV4Nil(body []byte) []byte {
	body = bytes.ReplaceAll(body, []byte("\r\n"), []byte("\n"))
	out := make([]byte, 0, len(body))
	inString := false
	for i := 0; i < len(body); i++ {
		c := body[i]
		if c == '\\' && inString && i+1 < len(body) {
			out = append(out, c, body[i+1])
			i++
			continue
		}
		if c == '"' {
			inString = !inString
			out = append(out, c)
			continue
		}
		if !inString && c == 'n' && i+2 < len(body) && body[i+1] == 'i' && body[i+2] == 'l' {
			if i+3 >= len(body) || !isV4Alpha(body[i+3]) {
				out = append(out, 'n', 'u', 'l', 'l')
				i += 2
				continue
			}
		}
		out = append(out, c)
	}
	return out
}

func isV4Alpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}
