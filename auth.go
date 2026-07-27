package ikuaiapi

import "strings"

// TokenHelp is a short, copy-pasteable instruction for obtaining a token
// from an iKuai router web UI. The SDK does not log in on behalf of the
// caller: tokens are generated manually in System → Auth → API Token.
const TokenHelp = "obtain a token from the router web UI (System → Auth → API Token)"

// ValidateToken returns an error if the token looks malformed. iKuai
// router tokens are 32-character lowercase hex strings.
func ValidateToken(token string) error {
	t := strings.TrimSpace(token)
	if t == "" {
		return &APIError{Code: 0, Message: "empty token; " + TokenHelp}
	}
	if len(t) != 32 {
		return &APIError{Code: 0, Message: "token must be 32 hex characters; " + TokenHelp}
	}
	for _, c := range t {
		switch {
		case c >= '0' && c <= '9':
		case c >= 'a' && c <= 'f':
		case c >= 'A' && c <= 'F':
		default:
			return &APIError{Code: 0, Message: "token must be hex; " + TokenHelp}
		}
	}
	return nil
}
