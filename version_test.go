package ikuaisdk

import "testing"

func TestVersionString(t *testing.T) {
	tests := []struct {
		version  Version
		expected string
	}{
		{VersionUnknown, "unknown"},
		{VersionV3, "v3"},
		{VersionV4, "v4"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.version.String(); got != tt.expected {
				t.Errorf("Version.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}
