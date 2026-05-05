package ikuaiapi

import (
	"context"

	"github.com/zy84338719/ikuai-api/internal"
	"github.com/zy84338719/ikuai-api/types"
)

func (c *Client) Login(ctx context.Context) error {
	loginPath := "/Action/login"
	if c.protocol != nil {
		loginPath = c.protocol.LoginPath()
	}

	req := &types.LoginRequest{
		Username: c.username,
		Password: internal.MD5Hash(c.password),
		Pass:     internal.Base64Password(c.password),
	}

	var loginResp types.BaseResponse
	if err := c.doRequest(ctx, loginPath, req, &loginResp); err != nil {
		return err
	}

	if c.protocol == nil {
		c.protocol = detectProtocol(&loginResp)
	}
	c.version = c.protocol.Version()

	if !c.protocol.IsSuccess(&loginResp) {
		return NewSDKError(ErrCodeLoginFailed, c.protocol.ErrorMessage(&loginResp), nil)
	}

	c.loggedIn = true
	return nil
}

func (c *Client) Logout(ctx context.Context) error {
	req := &types.BaseRequest{
		FuncName: "logout",
		Action:   "logout",
	}

	var baseResp types.BaseResponse
	logoutPath := "/Action/logout"
	if c.protocol != nil {
		logoutPath = c.protocol.LogoutPath()
	}
	if err := c.doRequest(ctx, logoutPath, req, &baseResp); err != nil {
		return err
	}

	c.loggedIn = false
	return nil
}

func (c *Client) CheckLogin(ctx context.Context) (bool, error) {
	req := &types.BaseRequest{
		FuncName: "webuser",
		Action:   "show",
	}

	var baseResp types.BaseResponse
	callPath := "/Action/call"
	if c.protocol != nil {
		callPath = c.protocol.CallPath()
	}
	if err := c.doRequest(ctx, callPath, req, &baseResp); err != nil {
		return false, err
	}

	if c.protocol != nil {
		return c.protocol.IsSuccess(&baseResp), nil
	}
	return baseResp.IsSuccess(), nil
}

func NewClientWithLogin(baseURL, username, password string, opts ...ClientOption) (*Client, error) {
	return NewClientWithLoginContext(context.Background(), baseURL, username, password, opts...)
}

func NewClientWithLoginContext(ctx context.Context, baseURL, username, password string, opts ...ClientOption) (*Client, error) {
	client := NewClient(baseURL, username, password, opts...)
	if err := client.Login(ctx); err != nil {
		client.Close()
		return nil, err
	}
	return client, nil
}

func NewV3ClientWithLogin(baseURL, username, password string, opts ...ClientOption) (*Client, error) {
	return NewV3ClientWithLoginContext(context.Background(), baseURL, username, password, opts...)
}

func NewV3ClientWithLoginContext(ctx context.Context, baseURL, username, password string, opts ...ClientOption) (*Client, error) {
	opts = append([]ClientOption{WithVersion(VersionV3)}, opts...)
	return NewClientWithLoginContext(ctx, baseURL, username, password, opts...)
}
