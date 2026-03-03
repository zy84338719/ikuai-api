package ikuaisdk

import (
	"context"

	"github.com/zy84338719/ikuai-api/internal"
	"github.com/zy84338719/ikuai-api/types"
)

func (c *Client) Login(ctx context.Context) error {
	req := &types.LoginRequest{
		Username: c.username,
		Password: internal.MD5Hash(c.password),
		Pass:     internal.Base64Password(c.password),
	}

	var loginResp types.BaseResponse
	if err := c.doRequest(ctx, "/Action/login", req, &loginResp); err != nil {
		return err
	}

	if !loginResp.IsSuccess() {
		return NewSDKError(ErrCodeLoginFailed, loginResp.GetErrorMessage(), nil)
	}

	if loginResp.IsV4() {
		c.version = VersionV4
	} else {
		c.version = VersionV3
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
	if err := c.doRequest(ctx, "/Action/logout", req, &baseResp); err != nil {
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
	if err := c.doRequest(ctx, "/Action/call", req, &baseResp); err != nil {
		return false, err
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
