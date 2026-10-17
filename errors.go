package ikuaiapi

import (
	"errors"
	"strings"
	"sync"
	"time"

	oekerrors "github.com/oekit/pkg/errors"
)

// Error handling in this SDK follows the oekit contract so that an error
// raised here is understood by an oekit service unchanged.
//
// The important consequence: every APIError carries an oekit BizError, so
// a handler calling errors.As(err, &bizErr) — including oekit's own
// middleware.Fail — sees a business code and an HTTP status instead of
// falling through to a generic 500. Mixing two error packages in one
// process breaks that recognition; this SDK therefore uses oekit's package
// throughout rather than the standard library's.

// APIError is returned when the router replies with a non-success envelope
// or an HTTP 4xx/5xx status. It is the only typed error the SDK raises for
// protocol-level failures; transport failures come back as *NetworkError.
type APIError struct {
	// BizError is the oekit error contract: a business code and an HTTP
	// status code that an oekit handler can act on. For a router error it
	// is derived from the iKuai code — see CodeFor and StatusFor.
	*oekerrors.BizError

	// Code is the iKuai error code, e.g. 3007 for an expired token.
	Code int
	// HTTPStatus is the status the router actually replied with.
	HTTPStatus int
	// Message is the router's message, with the SDK's hint appended when
	// the code is recognized.
	Message string
	// Details carries per-field validation failures when the router
	// supplies them.
	Details []APIErrorDetail
	// RetryAfter, when non-zero, carries a server-advised back-off hint
	// (parsed from the HTTP Retry-After header, typically on 429/503).
	RetryAfter time.Duration

	// cause holds the underlying error until the BizError is built.
	// It is deliberately unexported: e.Cause would resolve to the embedded
	// BizError's field and dereference a nil pointer before construction.
	cause error
}

// APIErrorDetail is one field-level validation failure reported by the
// router.
type APIErrorDetail struct {
	Field string `json:"field"`
	Type  string `json:"type"`
	Msg   string `json:"msg"`
}

// ErrorOption configures an APIError at construction time.
type ErrorOption func(*APIError)

// WithCode sets the iKuai error code.
func WithCode(code int) ErrorOption {
	return func(e *APIError) { e.Code = code }
}

// WithHTTPStatus sets the status the router replied with.
func WithHTTPStatus(status int) ErrorOption {
	return func(e *APIError) { e.HTTPStatus = status }
}

// WithDetails attaches field-level validation failures.
func WithDetails(details []APIErrorDetail) ErrorOption {
	return func(e *APIError) { e.Details = details }
}

// WithCause attaches the underlying error, preserving it for errors.Is and
// errors.As on the cause.
func WithCause(cause error) ErrorOption {
	return func(e *APIError) { e.cause = cause }
}

// WithRetryAfter records a server-advised back-off.
func WithRetryAfter(d time.Duration) ErrorOption {
	return func(e *APIError) { e.RetryAfter = d }
}

// NewError builds an APIError. Prefer this over a struct literal: it is the
// only path that fills in the oekit BizError the ecosystem expects.
func NewError(msg string, opts ...ErrorOption) *APIError {
	e := &APIError{Message: msg}
	for _, opt := range opts {
		opt(e)
	}
	if e.Message == "" && e.cause != nil {
		e.Message = e.cause.Error()
	}
	// Attach the hint now so the message reads the same whether or not the
	// caller inspects the code.
	if hint, ok := errorHints[e.Code]; ok {
		e.Message = e.Message + " (" + hint + ")"
	}
	e.BizError = oekerrors.NewBizErrorWithStatus(
		bizCodeFor(e.Code, e.HTTPStatus), statusFor(e.Code, e.HTTPStatus), e.Message)
	if e.cause != nil {
		e.BizError.Cause = e.cause
	}
	return e
}

// errorHints maps known iKuai error codes to human-friendly hints.
var errorHints = map[int]string{
	3001:  "parameter error, check the request body or required fields",
	3007:  "token expired or invalid, obtain a new one from the router web UI",
	1008:  "session expired, obtain a new token from the router web UI",
	10014: "permission denied for this operation, use an account that has it granted",
}

// errorRemedies maps known iKuai error codes to the concrete action that
// resolves them. Hints describe the failure; remedies describe the fix, and
// are what a CLI or agent surfaces to the operator.
var errorRemedies = map[int]string{
	3001:  "re-check the request body against the endpoint's field hints; a required field is missing or mistyped",
	3007:  "obtain a fresh API token in the router web UI (System -> Auth -> API Token) and set it on the client",
	1008:  "obtain a fresh API token in the router web UI (System -> Auth -> API Token) and set it on the client",
	10014: "sign in with an administrator account, or grant the current account permission for this feature",
}

// iKuai code classifications.
var (
	authErrorCodes       = map[int]bool{3007: true, 1008: true}
	permissionErrorCodes = map[int]bool{10014: true}
	badRequestCodes      = map[int]bool{3001: true}
)

// CodeFor maps an iKuai error code to an oekit business code. The iKuai
// code is itself a stable business identifier, so it is reused as-is; an
// unrecognized code is passed through so nothing is silently remapped.
func CodeFor(ikuaiCode int) int { return ikuaiCode }

// SDK-local validation codes.
//
// These are deliberately in their own 4xxx band. A router code in the error
// means the request was sent and rejected; an SDK code in the 4xxx band
// means the request never left the process. Keeping the bands apart lets a
// caller tell "my payload is wrong" from "the router said no" without
// inspecting a message string.
const (
	// errCodeValidation marks a request the SDK refused to send.
	errCodeValidation = 4900
)

// StatusFor maps an iKuai error code to the HTTP status an oekit handler
// should answer with.
//
// A status the router actually sent always wins. Failing that:
//
//   - the 4xxx band is the SDK's own validation band, so 400;
//   - a known iKuai code maps to its own meaning (401 auth, 403 permission,
//     400 bad payload);
//   - an SDK error with neither a code nor a router status means the SDK
//     refused to send the request at all — options were missing, a page
//     number was zero, an endpoint is not in the catalog. That is a
//     client-side mistake, so it answers 400 rather than falling through to
//     500 and making an oekit handler blame its own server.
func StatusFor(ikuaiCode, routerStatus int) int {
	if routerStatus >= 400 {
		return routerStatus
	}
	if ikuaiCode >= 4000 && ikuaiCode < 5000 {
		return 400
	}
	switch {
	case authErrorCodes[ikuaiCode]:
		return 401
	case permissionErrorCodes[ikuaiCode]:
		return 403
	case badRequestCodes[ikuaiCode]:
		return 400
	}
	if routerStatus > 0 {
		return routerStatus
	}
	if ikuaiCode == 0 {
		// Raised locally; nothing reached the router.
		return 400
	}
	return 500
}

func bizCodeFor(ikuaiCode, routerStatus int) int {
	if ikuaiCode != 0 {
		return CodeFor(ikuaiCode)
	}
	return routerStatus
}

func statusFor(ikuaiCode, routerStatus int) int { return StatusFor(ikuaiCode, routerStatus) }

// ensureBizError fills in the BizError for an APIError built as a struct
// literal rather than through NewError. It keeps the zero-value path safe
// without making the common path pay for it.
func (e *APIError) ensureBizError() {
	if e.BizError != nil {
		return
	}
	if e.Message == "" && e.cause != nil {
		e.Message = e.cause.Error()
	}
	e.BizError = oekerrors.NewBizErrorWithStatus(
		bizCodeFor(e.Code, e.HTTPStatus), statusFor(e.Code, e.HTTPStatus), e.Message)
	if e.cause != nil {
		e.BizError.Cause = e.cause
	}
}

func (e *APIError) Error() string {
	if e == nil {
		return "<nil>"
	}
	var b strings.Builder
	if e.HTTPStatus != 0 {
		b.WriteString("HTTP ")
		writeInt(&b, e.HTTPStatus)
		if e.Code != 0 {
			b.WriteString(", ")
		}
	}
	if e.Code != 0 {
		b.WriteString("code ")
		writeInt(&b, e.Code)
	}
	if b.Len() > 0 {
		b.WriteString(": ")
	}
	b.WriteString(e.Message)
	for _, d := range e.Details {
		b.WriteString("\n  - ")
		b.WriteString(d.Field)
		b.WriteString(": ")
		b.WriteString(d.Msg)
	}
	return b.String()
}

// Unwrap exposes the embedded BizError so that errors.As can reach it. An
// oekit handler unwrapping the error therefore still sees the business
// code and HTTP status, and a caller can keep reaching the cause.
func (e *APIError) Unwrap() error {
	if e == nil {
		return nil
	}
	if e.BizError == nil {
		return nil
	}
	return e.BizError
}

// Biz returns the oekit business error for this failure, constructing it if
// the APIError was built as a literal.
func (e *APIError) Biz() *oekerrors.BizError {
	if e == nil {
		return nil
	}
	e.ensureBizError()
	return e.BizError
}

// AsBizError extracts the oekit business error from any error chain that
// passes through this SDK. It is the local equivalent of oekit's
// errors.AsBizError and exists so callers do not have to know which layer
// produced the error.
func AsBizError(err error) (*oekerrors.BizError, bool) {
	return oekerrors.AsBizError(err)
}

// Remedy returns the actionable fix for a known iKuai error code, or an
// empty string when the code is not recognized.
func (e *APIError) Remedy() string {
	if e == nil {
		return ""
	}
	return errorRemedies[e.Code]
}

// IsAuthFailure reports whether the error means the token or session is
// missing, expired or invalid. Retrying with the same credentials will not
// help; the caller must refresh the token first.
func (e *APIError) IsAuthFailure() bool {
	if e == nil {
		return false
	}
	return authErrorCodes[e.Code]
}

// IsPermissionFailure reports whether the error means the account is
// authenticated but lacks permission for the operation. Retrying will not
// help; the caller needs a more privileged account.
func (e *APIError) IsPermissionFailure() bool {
	if e == nil {
		return false
	}
	return permissionErrorCodes[e.Code]
}

// IsBadRequest reports whether the router rejected the request itself,
// which a caller can fix by correcting the payload.
func (e *APIError) IsBadRequest() bool {
	if e == nil {
		return false
	}
	return badRequestCodes[e.Code]
}

// IsRetryable reports whether a caller may safely retry the request that
// produced this error. It encodes the SDK's own retry policy so
// applications can reuse it for custom retry loops or circuit-breaker
// decisions.
//
//   - *NetworkError: retryable (transport hiccups usually clear up), but
//     only for idempotent verbs — the SDK never auto-retries a write on a
//     network error because the request may have reached the router.
//   - *APIError: retryable on HTTP 429 (rate limited) and 5xx. Other 4xx
//     are not retryable.
func (e *APIError) IsRetryable() bool {
	if e == nil {
		return false
	}
	// 429 Too Many Requests and any 5xx are safe to retry.
	return e.HTTPStatus == 429 || e.HTTPStatus >= 500
}

// NetworkError wraps connection-level failures (DNS, refused, TLS, timeout).
type NetworkError struct {
	Message string
	Cause   error
}

func (e *NetworkError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *NetworkError) Unwrap() error { return e.Cause }

// IsRetryable reports whether the transport failure may be retried. Pass
// the HTTP method to qualify network errors: a write is not retried,
// because it may already have reached the router.
func (e *NetworkError) IsRetryable() bool { return true }

// ErrInvalidToken is returned when a token fails format validation. It is a
// sentinel so callers can match it with errors.Is.
var ErrInvalidToken = errors.New("ikuaiapi: invalid token")

// ErrNotConfigured is returned when a client is constructed without a base
// URL, which is the one value with no sensible default.
var ErrNotConfigured = errors.New("ikuaiapi: baseURL is required")

// Sentinel errors for the SDK's own pre-flight validation. Router-side
// failures always come back as *APIError.
var (
	// ErrRequiredField reports a missing required body field.
	ErrRequiredField = errors.New("ikuaiapi: required field missing")
	// ErrUnknownObjectGroup reports an object group name the SDK does not
	// know how to encode.
	ErrUnknownObjectGroup = errors.New("ikuaiapi: unknown object group")
	// ErrNoSchema reports that no body contract is recorded for a path.
	ErrNoSchema = errors.New("ikuaiapi: no schema recorded for path")
	// ErrEmptyResponse reports a response with nothing to decode.
	ErrEmptyResponse = errors.New("ikuaiapi: empty response")
	// ErrNoObject reports a response that decoded but held no object.
	ErrNoObject = errors.New("ikuaiapi: response contains no object")
)

func writeInt(b *strings.Builder, n int) {
	if n == 0 {
		b.WriteByte('0')
		return
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		b.WriteByte('-')
	}
	b.Write(buf[i:])
}

func isAPIError(err error) (*APIError, bool) {
	if err == nil {
		return nil, false
	}
	var e *APIError
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

func isNetworkError(err error) (*NetworkError, bool) {
	if err == nil {
		return nil, false
	}
	var e *NetworkError
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

// clientCloseOnce guards Close against concurrent and repeated calls, so a
// double Close in a deferred chain is harmless.
type closeOnce struct {
	once sync.Once
	err  error
}
