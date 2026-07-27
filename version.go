// Package version: the iKuai API only supports the v4 REST surface.
package ikuaiapi

// Version enumerates the iKuai API generations the SDK recognises.
type Version int

const (
	// VersionUnknown is the zero value; iKuai API v4 is the only supported
	// version, so new clients should pass VersionV4 explicitly.
	VersionUnknown Version = iota
	// VersionV4 is the current iKuai REST API (/api/v4.0).
	VersionV4
)

func (v Version) String() string {
	switch v {
	case VersionV4:
		return "v4"
	default:
		return "unknown"
	}
}
