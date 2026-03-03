package types

type BaseRequest struct {
	FuncName string      `json:"func_name"`
	Action   string      `json:"action"`
	Param    interface{} `json:"param,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"passwd"`
	Pass     string `json:"pass"`
}

type BaseResponse struct {
	Result  int    `json:"Result"`
	ErrMsg  string `json:"ErrMsg"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (r *BaseResponse) IsV4() bool {
	return r.Message != ""
}

func (r *BaseResponse) IsSuccess() bool {
	if r.IsV4() {
		return r.Code == 0
	}
	return r.Result == 10000 || r.Result == 30000
}

func (r *BaseResponse) GetErrorMessage() string {
	if r.ErrMsg != "" {
		return r.ErrMsg
	}
	if r.Message != "" {
		return r.Message
	}
	return ""
}
