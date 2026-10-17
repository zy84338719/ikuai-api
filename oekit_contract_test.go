package ikuaiapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	oekerrors "github.com/oekit/pkg/errors"
)

// TestErrorIsVisibleToTheOekitChain is the contract that justifies this
// SDK taking an oekit dependency: an error raised here must be understood
// by an oekit handler without any adapter in between. If a caller does
//
//	mw.Fail(ctx, c, err)
//
// it has to see a business code and an HTTP status, not fall through to a
// generic 500.
func TestErrorIsVisibleToTheOekitChain(t *testing.T) {
	routerRejects := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":3007,"message":"Invalid token"}`))
	}))
	defer routerRejects.Close()

	c, err := NewClient(routerRejects.URL, WithToken("t"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer func() { _ = c.Close(context.Background()) }()

	_, err = c.Get(t.Context(), "/monitoring/system", nil)
	if err == nil {
		t.Fatal("expected an error")
	}

	// 1. The caller's own type is still reachable.
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("errors.As should reach *APIError, got %T", err)
	}
	if apiErr.Code != 3007 {
		t.Errorf("Code = %d, want the iKuai code 3007", apiErr.Code)
	}

	// 2. The oekit business error is reachable through the same chain.
	biz, ok := oekerrors.AsBizError(err)
	if !ok {
		t.Fatal("oekit errors.AsBizError must recognize an error from this SDK")
	}
	if biz.BizCode != 3007 {
		t.Errorf("BizCode = %d, want 3007", biz.BizCode)
	}
	if biz.HTTPCode != 401 {
		t.Errorf("HTTPCode = %d, want 401 for an expired token", biz.HTTPCode)
	}
	if biz.Message == "" {
		t.Error("BizError.Message should not be empty")
	}
}

// TestErrorCodeMapping pins the iKuai-code to HTTP-status translation, which
// is what an oekit handler answers with.
func TestErrorCodeMapping(t *testing.T) {
	tests := []struct {
		ikuaiCode   int
		routerHTTP  int
		wantBizCode int
		wantStatus  int
		authFailure bool
		permFailure bool
		badRequest  bool
	}{
		{ikuaiCode: 3001, wantBizCode: 3001, wantStatus: 400, badRequest: true},
		{ikuaiCode: 3007, wantBizCode: 3007, wantStatus: 401, authFailure: true},
		{ikuaiCode: 1008, wantBizCode: 1008, wantStatus: 401, authFailure: true},
		{ikuaiCode: 10014, wantBizCode: 10014, wantStatus: 403, permFailure: true},
		// A status the router actually sent wins over the derived one.
		{ikuaiCode: 3007, routerHTTP: 503, wantBizCode: 3007, wantStatus: 503, authFailure: true},
	}
	for _, tc := range tests {
		e := NewError("boom", WithCode(tc.ikuaiCode), WithHTTPStatus(tc.routerHTTP))
		if e.BizError.BizCode != tc.wantBizCode {
			t.Errorf("code %d: BizCode = %d, want %d", tc.ikuaiCode, e.BizError.BizCode, tc.wantBizCode)
		}
		if e.BizError.HTTPCode != tc.wantStatus {
			t.Errorf("code %d: HTTPCode = %d, want %d", tc.ikuaiCode, e.BizError.HTTPCode, tc.wantStatus)
		}
		if e.IsAuthFailure() != tc.authFailure {
			t.Errorf("code %d: IsAuthFailure = %v, want %v", tc.ikuaiCode, e.IsAuthFailure(), tc.authFailure)
		}
		if e.IsPermissionFailure() != tc.permFailure {
			t.Errorf("code %d: IsPermissionFailure = %v, want %v", tc.ikuaiCode, e.IsPermissionFailure(), tc.permFailure)
		}
		if e.IsBadRequest() != tc.badRequest {
			t.Errorf("code %d: IsBadRequest = %v, want %v", tc.ikuaiCode, e.IsBadRequest(), tc.badRequest)
		}
		if e.Biz() == nil {
			t.Errorf("code %d: Biz() returned nil", tc.ikuaiCode)
		}
	}
}

// TestSDKLocalValidationIsDistinct checks that a request the SDK refuses to
// send is distinguishable from one the router rejected.
func TestSDKLocalValidationIsDistinct(t *testing.T) {
	_, err := CreateBody("/security/acl-rules", map[string]any{"tagname": "x"})
	if err == nil {
		t.Fatal("expected a validation error")
	}
	if !errors.Is(err, ErrRequiredField) {
		t.Error("the error should wrap ErrRequiredField so callers can match it")
	}
	biz, ok := oekerrors.AsBizError(err)
	if !ok {
		t.Fatal("a local validation error should still be a BizError")
	}
	if biz.HTTPCode != 400 {
		t.Errorf("HTTPCode = %d, want 400", biz.HTTPCode)
	}
	if biz.BizCode < 4000 || biz.BizCode >= 5000 {
		t.Errorf("BizCode = %d, want the 4xxx SDK-local band", biz.BizCode)
	}
}

// TestErrorCauseSurvivesTheChain guards the unwrap path: a caller must be
// able to match the sentinel that caused the failure.
func TestErrorCauseSurvivesTheChain(t *testing.T) {
	e := NewError("outer", WithCode(3001), WithCause(ErrRequiredField))
	if !errors.Is(e, ErrRequiredField) {
		t.Error("errors.Is must reach the cause through the BizError")
	}
	var bz *oekerrors.BizError
	if !errors.As(error(e), &bz) {
		t.Error("errors.As must reach the BizError")
	}
	if !errors.Is(bz, ErrRequiredField) {
		t.Error("the BizError itself must carry the cause")
	}
}

// TestErrorLiteralIsSafe checks the defensive path: an APIError built as a
// struct literal has no BizError, and must not panic when used.
func TestErrorLiteralIsSafe(t *testing.T) {
	var e APIError
	e.Code = 3007
	e.Message = "literal"
	if got := e.Error(); got == "" {
		t.Error("Error() should render a literal APIError")
	}
	if biz := e.Biz(); biz == nil || biz.HTTPCode != 401 {
		t.Errorf("Biz() on a literal should derive the status, got %+v", biz)
	}
}

// TestCloseIsIdempotent locks the oekit lifecycle rule: Close takes a
// context, returns an error, and may be called more than once.
func TestCloseIsIdempotent(t *testing.T) {
	c, err := NewClient("https://192.0.2.1", WithToken("t"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	ctx := t.Context()
	for i := 0; i < 3; i++ {
		if err := c.Close(ctx); err != nil {
			t.Fatalf("Close call %d: %v", i+1, err)
		}
	}
	// A cancelled context is reported rather than swallowed.
	c2, err := NewClient("https://192.0.2.1", WithToken("t"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if err := c2.Close(cancelled); err == nil {
		t.Error("Close should report a cancelled context")
	}
}

// TestLocalValidationIsNotServerError locks the mapping for errors the SDK
// raises before anything reaches the router.
//
// These reached oekit handlers as HTTP 500 because the code was neither an
// iKuai code nor a router status: the fallback assumed something had gone
// wrong on the server, when in fact the SDK had rejected the call. A
// validation failure answered with 500 makes the caller blame the wrong
// tier and, worse, trips alerting on a request that was never sent.
func TestLocalValidationIsNotServerError(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{"options are required", NewError("options are required")},
		{"page below one", NewError("Page must be greater than 0")},
		{"endpoint not in catalog", NewError("endpoint not found: a/b")},
	}
	for _, tc := range cases {
		ae, ok := tc.err.(*APIError)
		if !ok {
			t.Fatalf("%s: want *APIError, got %T", tc.name, tc.err)
		}
		if got := ae.Biz().HTTPCode; got != 400 {
			t.Errorf("%s: HTTPCode = %d, want 400 for a request that was never sent",
				tc.name, got)
		}
	}

	// A router failure must still read as a server failure.
	srvErr := NewError("boom", WithHTTPStatus(503))
	if got := srvErr.Biz().HTTPCode; got != 503 {
		t.Errorf("router 503: HTTPCode = %d, want 503", got)
	}
	// And an unmapped iKuai code with no status stays a server error.
	unknown := NewError("weird", WithCode(50001))
	if got := unknown.Biz().HTTPCode; got != 500 {
		t.Errorf("unknown iKuai code: HTTPCode = %d, want 500", got)
	}
}
