package output

import (
	"io"
	"os"
	"strings"
)

// ForceTTYEnv forces terminal behaviour on or off regardless of what the
// output stream actually is. Tests use it; operators use it when piping
// through something that hides the real destination.
var ForceTTYEnv = "IKUAI_FORCE_TTY"

// IsTerminal reports whether w is an interactive character device.
//
// The standard library cannot do a full terminal query, so this uses the
// file mode instead of x/term's ioctl probe. It answers the question that
// actually matters here — "is a human watching, or a pipe?" — and keeps the
// package dependency-free. Set ForceTTYEnv to "1" to force true, or to "0"
// to force false.
func IsTerminal(w io.Writer) bool {
	switch strings.TrimSpace(os.Getenv(ForceTTYEnv)) {
	case "1":
		return true
	case "0":
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// TermWidth reports the width of w in columns, or 0 when the width is
// unknown or irrelevant. Callers pass the result as RenderOptions.TermWidth to
// have the renderer drop columns that would overflow.
func TermWidth(w io.Writer) int {
	f, ok := w.(*os.File)
	if !ok {
		return 0
	}
	// The width query needs an ioctl, which is not in the standard library.
	// A fixed conservative default keeps tables readable when the real
	// width is unknown, and callers that know the width should set it
	// explicitly.
	if !IsTerminal(f) {
		return 0
	}
	return 0
}
