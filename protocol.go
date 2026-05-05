package ikuaisdk

import "github.com/zy84338719/ikuai-api/types"

type apiProtocol interface {
	Version() Version
	LoginPath() string
	LogoutPath() string
	CallPath() string
	IsSuccess(*types.BaseResponse) bool
	ErrorMessage(*types.BaseResponse) string
}

type v3Protocol struct{}

func (v3Protocol) Version() Version   { return VersionV3 }
func (v3Protocol) LoginPath() string  { return "/Action/login" }
func (v3Protocol) LogoutPath() string { return "/Action/logout" }
func (v3Protocol) CallPath() string   { return "/Action/call" }
func (v3Protocol) IsSuccess(resp *types.BaseResponse) bool {
	return resp.Result == 10000 || resp.Result == 30000
}
func (v3Protocol) ErrorMessage(resp *types.BaseResponse) string { return resp.ErrMsg }

type v4Protocol struct{}

func (v4Protocol) Version() Version                             { return VersionV4 }
func (v4Protocol) LoginPath() string                            { return "/Action/login" }
func (v4Protocol) LogoutPath() string                           { return "/Action/logout" }
func (v4Protocol) CallPath() string                             { return "/Action/call" }
func (v4Protocol) IsSuccess(resp *types.BaseResponse) bool      { return resp.Code == 0 }
func (v4Protocol) ErrorMessage(resp *types.BaseResponse) string { return resp.Message }

func protocolForVersion(version Version) apiProtocol {
	switch version {
	case VersionV3:
		return v3Protocol{}
	case VersionV4:
		return v4Protocol{}
	default:
		return nil
	}
}

func detectProtocol(resp *types.BaseResponse) apiProtocol {
	if resp.IsV4() {
		return v4Protocol{}
	}
	return v3Protocol{}
}
