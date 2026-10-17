package ikuaiapi

import (
	"encoding/base64"
	"strings"
)

// TokenHelp is a short, copy-pasteable instruction for obtaining a token
// from an iKuai router web UI. The SDK does not log in on behalf of the
// caller: tokens are generated manually in System → Auth → API Token.
const TokenHelp = "obtain a token from the router web UI (System → Auth → API Token)"

// TokenLength is the character length of an iKuai API token, in both of the
// shapes firmware has shipped.
const TokenLength = 32

// ValidateToken returns an error if the token is obviously malformed.
//
// iKuai has shipped two token shapes and both authenticate:
//
//   - 32 lowercase hex characters, on older firmware.
//   - 32 base64 characters, which decode to a UUID prefix. This is what a
//     current 4.x router issues — verified against hardware, where the token
//     worked on every endpoint while failing this check.
//
// The check is therefore deliberately a shape check rather than a format
// one. Rejecting a working token is far worse than accepting a malformed
// one: the router is the authority, and a bad token fails there with a clear
// error anyway. Call ValidateToken to catch a pasted-then-truncated string,
// not to pre-authenticate.
func ValidateToken(token string) error {
	t := strings.TrimSpace(token)
	if t == "" {
		return NewError("empty token; " + TokenHelp)
	}
	if len(t) != TokenLength {
		return NewError(
			"token must be 32 characters (hex on older firmware, base64 on current); " + TokenHelp)
	}
	if strings.ContainsAny(t, " \t\r\n") {
		return NewError("token must not contain whitespace; " + TokenHelp)
	}
	if isHex(t) {
		return nil
	}
	if _, err := base64.RawStdEncoding.DecodeString(t); err == nil {
		return nil
	}
	// Neither shape decoded. The token still reaches the router, so this is
	// a hint rather than a wall — but a token with no recognisable alphabet
	// is almost always a copy-paste accident.
	return NewError("token is not valid hex or base64; " + TokenHelp)
}

func isHex(s string) bool {
	for _, c := range s {
		switch {
		case c >= '0' && c <= '9':
		case c >= 'a' && c <= 'f':
		case c >= 'A' && c <= 'F':
		default:
			return false
		}
	}
	return true
}
