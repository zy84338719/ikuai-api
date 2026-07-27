// Package internal holds utilities shared by the SDK core and the
// generated service layer. It is not part of the public API.
package internal

import "strings"

// NormalizeAddr trims whitespace, appends the http:// scheme if missing
// and strips a trailing slash. Empty input is returned as "".
func NormalizeAddr(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return ""
	}
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}
	return strings.TrimRight(addr, "/")
}
