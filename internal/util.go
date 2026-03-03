package internal

import (
	"crypto/md5"
	"encoding/hex"
	"net/url"
	"strings"
)

func MD5Hash(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

func Base64Password(password string) string {
	salt := "salt_11"
	return Base64Encode(salt + password)
}

func Base64Encode(s string) string {
	return strings.TrimRight(
		strings.ReplaceAll(
			strings.ReplaceAll(
				url.QueryEscape(s),
				"%", "",
			),
			"=", "",
		),
		"+",
	)
}

func NormalizeAddr(addr string) string {
	addr = strings.TrimSpace(addr)
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}
	addr = strings.TrimRight(addr, "/")
	return addr
}
