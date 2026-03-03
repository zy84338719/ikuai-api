package ikuaisdk

type Version int

const (
	VersionUnknown Version = iota
	VersionV3
	VersionV4
)

func (v Version) String() string {
	switch v {
	case VersionV3:
		return "v3"
	case VersionV4:
		return "v4"
	default:
		return "unknown"
	}
}
