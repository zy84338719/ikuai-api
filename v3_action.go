package ikuaiapi

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type V3ActionClient struct {
	client  *Client
	rawMode bool
	dryRun  bool
}

type V3ActionOption func(*V3ActionClient)

func WithV3RawMode(raw bool) V3ActionOption {
	return func(c *V3ActionClient) {
		c.rawMode = raw
	}
}

func WithV3DryRun(dryRun bool) V3ActionOption {
	return func(c *V3ActionClient) {
		c.dryRun = dryRun
	}
}

func NewV3ActionClient(client *Client, opts ...V3ActionOption) *V3ActionClient {
	c := &V3ActionClient{client: client}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func NewV3ActionClientWithLogin(ctx context.Context, baseURL, username, password string, opts ...ClientOption) (*V3ActionClient, error) {
	client, err := NewV3ClientWithLoginContext(ctx, baseURL, username, password, opts...)
	if err != nil {
		return nil, err
	}
	return NewV3ActionClient(client), nil
}

func (c *Client) V3Action(opts ...V3ActionOption) *V3ActionClient {
	return NewV3ActionClient(c, opts...)
}

type V3ActionRequest struct {
	FuncName string      `json:"func_name"`
	Action   string      `json:"action"`
	Param    interface{} `json:"param,omitempty"`
}

type V3ActionEnvelope struct {
	Result  int             `json:"Result"`
	ErrMsg  string          `json:"ErrMsg"`
	Data    json.RawMessage `json:"Data"`
	Results json.RawMessage `json:"results"`
	ID      json.RawMessage `json:"id"`
	RowID   json.RawMessage `json:"rowid"`
}

func (c *V3ActionClient) CallRaw(ctx context.Context, funcName, action string, param interface{}) (json.RawMessage, error) {
	if c == nil || c.client == nil {
		return nil, NewSDKError(ErrCodeInvalidResponse, "v3 action client is nil", nil)
	}
	if !c.client.IsLoggedIn() {
		return nil, NewSDKError(ErrCodeNotLoggedIn, "client not logged in", nil)
	}
	if c.client.protocol == nil {
		c.client.protocol = protocolForVersion(VersionV3)
	}

	if c.dryRun {
		return json.Marshal(V3ActionRequest{
			FuncName: funcName,
			Action:   action,
			Param:    param,
		})
	}

	var raw json.RawMessage
	req := V3ActionRequest{
		FuncName: funcName,
		Action:   action,
		Param:    param,
	}
	if err := c.client.doRequest(ctx, c.client.protocol.CallPath(), req, &raw); err != nil {
		return nil, err
	}
	if c.rawMode {
		return raw, nil
	}
	return unwrapV3Action(raw)
}

func (c *V3ActionClient) Call(ctx context.Context, funcName, action string, param interface{}, result interface{}) error {
	raw, err := c.CallRaw(ctx, funcName, action, param)
	if err != nil {
		return err
	}
	if result == nil {
		return nil
	}
	return json.Unmarshal(raw, result)
}

func (c *V3ActionClient) DoRaw(ctx context.Context, endpointName, action string, param interface{}) (json.RawMessage, error) {
	endpoint, ok := V3EndpointByName(endpointName)
	if !ok {
		return nil, NewSDKError(ErrCodeVersionNotSupported, "unknown v3 compatibility endpoint: "+endpointName, nil)
	}
	if !endpoint.Supports(action) {
		return nil, NewSDKError(ErrCodeVersionNotSupported, endpointName+" does not support action "+action+" on v3", nil)
	}
	if param == nil {
		param = endpoint.DefaultParam
	}
	return c.CallRaw(ctx, endpoint.FuncName, action, param)
}

func (c *V3ActionClient) Do(ctx context.Context, endpointName, action string, param interface{}, result interface{}) error {
	raw, err := c.DoRaw(ctx, endpointName, action, param)
	if err != nil {
		return err
	}
	if result == nil {
		return nil
	}
	return json.Unmarshal(raw, result)
}

func (c *V3ActionClient) ShowRaw(ctx context.Context, endpointName string, param interface{}) (json.RawMessage, error) {
	return c.DoRaw(ctx, endpointName, "show", param)
}

func (c *V3ActionClient) Show(ctx context.Context, endpointName string, param interface{}, result interface{}) error {
	return c.Do(ctx, endpointName, "show", param, result)
}

func (c *V3ActionClient) AddRaw(ctx context.Context, endpointName string, param interface{}) (json.RawMessage, error) {
	return c.DoRaw(ctx, endpointName, "add", param)
}

func (c *V3ActionClient) Add(ctx context.Context, endpointName string, param interface{}, result interface{}) error {
	return c.Do(ctx, endpointName, "add", param, result)
}

func (c *V3ActionClient) EditRaw(ctx context.Context, endpointName string, param interface{}) (json.RawMessage, error) {
	return c.DoRaw(ctx, endpointName, "edit", param)
}

func (c *V3ActionClient) Edit(ctx context.Context, endpointName string, param interface{}, result interface{}) error {
	return c.Do(ctx, endpointName, "edit", param, result)
}

func (c *V3ActionClient) DeleteRaw(ctx context.Context, endpointName string, id int) (json.RawMessage, error) {
	return c.DoRaw(ctx, endpointName, "del", map[string]int{"id": id})
}

func (c *V3ActionClient) Delete(ctx context.Context, endpointName string, id int, result interface{}) error {
	return c.Do(ctx, endpointName, "del", map[string]int{"id": id}, result)
}

func unwrapV3Action(raw json.RawMessage) (json.RawMessage, error) {
	var env V3ActionEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("v3 action non-JSON response: %w", err)
	}
	if env.Result != 10000 && env.Result != 30000 {
		msg := env.ErrMsg
		if msg == "" {
			msg = string(raw)
		}
		return nil, NewSDKError(ErrCodeRequestFailed, msg, nil)
	}

	payload := firstNonEmptyJSON(env.Results, env.Data)
	if !isEmptyJSON(payload) {
		return unwrapV3Payload(payload), nil
	}

	result := map[string]interface{}{"message": "success"}
	if !isEmptyJSON(env.ID) {
		var id interface{}
		if err := json.Unmarshal(env.ID, &id); err == nil {
			result["id"] = id
		}
	}
	if !isEmptyJSON(env.RowID) {
		var rowid interface{}
		if err := json.Unmarshal(env.RowID, &rowid); err == nil {
			result["rowid"] = rowid
		}
	}
	return json.Marshal(result)
}

func unwrapV3Payload(payload json.RawMessage) json.RawMessage {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(payload, &obj); err != nil {
		return payload
	}
	for _, key := range []string{"data", "list", "items"} {
		if value, ok := obj[key]; ok && !isEmptyJSON(value) {
			return value
		}
	}
	return payload
}

func firstNonEmptyJSON(values ...json.RawMessage) json.RawMessage {
	for _, value := range values {
		if !isEmptyJSON(value) {
			return value
		}
	}
	return nil
}

func normalizeV3Name(name string) string {
	return strings.ReplaceAll(strings.TrimSpace(strings.ToLower(name)), "_", "-")
}
