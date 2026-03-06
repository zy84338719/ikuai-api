// Package ikuaisdk provides a Go SDK for interacting with iKuai routers.
// It supports both v3 and v4 versions of iKuai OS with automatic version detection.
//
// The SDK uses reqv3 HTTP client for better performance, debugging, retry, and HTTP2/3 support.
// All services are defined through interfaces for easy testing and extension.
//
// Basic usage:
//
//	client, err := ikuaisdk.NewClientWithLogin("http://192.168.1.1", "admin", "password")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	api := service.NewAPIClient(client)
//	homepage, err := api.System().GetHomepage(context.Background())
package ikuaisdk

import (
	"context"
	"time"

	"github.com/imroc/req/v3"
	"github.com/zy84338719/ikuai-api/internal"
	"github.com/zy84338719/ikuai-api/types"
)

// Client is the main client for interacting with iKuai router.
// It handles authentication, session management, and API calls.
type Client struct {
	client   *req.Client
	baseURL  string
	username string
	password string
	version  Version
	loggedIn bool
	cache    *ResponseCache
	logger   Logger
	metrics *Metrics
}

type ClientOption func(*Client)

func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.client.SetTimeout(timeout)
	}
}

func WithInsecureSkipVerify(skip bool) ClientOption {
	return func(c *Client) {
		if skip {
			c.client.EnableInsecureSkipVerify()
		}
	}
}

func WithHTTPClient(httpClient *req.Client) ClientOption {
	return func(c *Client) {
		c.client = httpClient
	}
}

func NewClient(baseURL, username, password string, opts ...ClientOption) *Client {
	baseURL = internal.NormalizeAddr(baseURL)

	c := &Client{
		baseURL:  baseURL,
		username: username,
		password: password,
		version:  VersionUnknown,
		loggedIn: false,
		logger:   NewDefaultLogger(LogLevelInfo),
		metrics:  NewMetrics(),
		client: req.C().
			SetBaseURL(baseURL).
			SetTimeout(30 * time.Second),
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

func (c *Client) doRequest(ctx context.Context, path string, reqBody interface{}, result interface{}) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(reqBody).
		SetSuccessResult(result).
		Post(path)

	if err != nil {
		return NewSDKError(ErrCodeRequestFailed, "failed to send request", err)
	}

	if resp.Err != nil {
		return NewSDKError(ErrCodeRequestFailed, "request error", resp.Err)
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

	return c.doRequest(ctx, "/Action/call", req, result)
}

func (c *Client) Close() {
	c.client.CloseIdleConnections()
}
