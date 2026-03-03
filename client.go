package ikuaisdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/zy84338719/ikuai-api/internal"
	"github.com/zy84338719/ikuai-api/types"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	username   string
	password   string
	version    Version
	loggedIn   bool
}

type ClientOption func(*Client)

func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

func WithInsecureSkipVerify(skip bool) ClientOption {
	return func(c *Client) {
		if skip {
			if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
				transport.TLSClientConfig.InsecureSkipVerify = true
			}
		}
	}
}

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

func NewClient(baseURL, username, password string, opts ...ClientOption) *Client {
	baseURL = internal.NormalizeAddr(baseURL)

	cookieJar, _ := cookiejar.New(nil)
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	c := &Client{
		baseURL:  baseURL,
		username: username,
		password: password,
		version:  VersionUnknown,
		loggedIn: false,
		httpClient: &http.Client{
			Transport: transport,
			Jar:       cookieJar,
			Timeout:   30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) GetVersion() Version {
	return c.version
}

func (c *Client) IsLoggedIn() bool {
	return c.loggedIn
}

func (c *Client) doRequest(ctx context.Context, path string, req interface{}) (*http.Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, NewSDKError(ErrCodeRequestFailed, "failed to marshal request", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, NewSDKError(ErrCodeRequestFailed, "failed to create request", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(httpReq)
}

func (c *Client) parseResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewSDKError(ErrCodeInvalidResponse, "failed to read response body", err)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return NewSDKError(ErrCodeInvalidResponse, fmt.Sprintf("failed to unmarshal response: %s", string(body)), err)
	}

	return nil
}

func (c *Client) Call(ctx context.Context, funcName, action string, param interface{}, result interface{}) error {
	if !c.loggedIn {
		return NewSDKError(ErrCodeNotLoggedIn, "client not logged in", nil)
	}

	req := &types.BaseRequest{
		FuncName: funcName,
		Action:   action,
		Param:    param,
	}

	resp, err := c.doRequest(ctx, "/Action/call", req)
	if err != nil {
		return err
	}

	if err := c.parseResponse(resp, result); err != nil {
		return err
	}

	return nil
}

func (c *Client) Close() {
	c.httpClient.CloseIdleConnections()
}
