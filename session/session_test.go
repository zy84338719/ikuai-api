package session

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// isolate points the package at a temp config file and clears every
// environment variable it reads, so tests never touch the developer's real
// credentials.
func isolate(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	t.Setenv(PathEnv, path)
	for _, name := range append(append([]string{}, BaseURLEnv...), append(TokenEnv, ForceTTY...)...) {
		t.Setenv(name, "")
	}
	return path
}

// ForceTTY lists the remaining environment variables the package reads; it
// is only used to clear them during tests.
var ForceTTY = []string{}

func TestLoadMissingFileIsNotAnError(t *testing.T) {
	isolate(t)
	s, err := Load()
	if err != nil {
		t.Fatalf("Load on a missing file: %v", err)
	}
	if s.Configured() {
		t.Errorf("expected an empty session, got %+v", s)
	}
	if s.Source() != SourceNone {
		t.Errorf("Source = %q, want %q", s.Source(), SourceNone)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := isolate(t)
	s := &Session{BaseURL: "https://192.168.1.1", Token: "deadbeef"}
	if err := s.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.BaseURL != s.BaseURL || got.Token != s.Token {
		t.Errorf("round trip = %+v, want %+v", got, s)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("config permissions = %v, want 0600", perm)
	}
}

// TestSaveCreatesPrivateDirectory checks that a directory the package
// creates for itself is private. An existing directory is left alone:
// quietly changing the mode of a directory the operator already made would
// be surprising.
func TestSaveCreatesPrivateDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "config.json")
	t.Setenv(PathEnv, path)
	for _, name := range append(append([]string{}, BaseURLEnv...), TokenEnv...) {
		t.Setenv(name, "")
	}
	if err := (&Session{BaseURL: "https://192.168.1.1", Token: "t"}).Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	info, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o700 {
		t.Errorf("created directory permissions = %v, want 0700", perm)
	}
}

func TestSaveIsAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	t.Setenv(PathEnv, path)
	for _, name := range append(append([]string{}, BaseURLEnv...), TokenEnv...) {
		t.Setenv(name, "")
	}

	if err := (&Session{BaseURL: "https://192.168.1.1", Token: "a"}).Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := (&Session{BaseURL: "https://192.168.1.1", Token: "b"}).Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	// No temporary files may survive a successful save.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "config.json" {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("directory should hold only config.json, got %v", names)
	}
}

func TestResolvePrefersSessionFile(t *testing.T) {
	isolate(t)
	t.Setenv(BaseURLEnv[0], "https://10.0.0.1")
	t.Setenv(TokenEnv[0], "env-token")

	if err := (&Session{BaseURL: "https://192.168.1.1", Token: "file-token"}).Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	s, err := Resolve()
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if s.Token != "file-token" || s.BaseURL != "https://192.168.1.1" {
		t.Errorf("the session file must win, got %+v", s)
	}
	if s.Source() != SourceFile {
		t.Errorf("Source = %q, want %q", s.Source(), SourceFile)
	}
}

func TestResolveFallsBackToEnv(t *testing.T) {
	path := isolate(t)
	t.Setenv(BaseURLEnv[0], "https://10.0.0.1/")
	t.Setenv(TokenEnv[0], "env-token")

	s, err := Resolve()
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if s.BaseURL != "https://10.0.0.1" {
		t.Errorf("BaseURL = %q, want the trailing slash stripped", s.BaseURL)
	}
	if s.Token != "env-token" {
		t.Errorf("Token = %q, want the env value", s.Token)
	}
	if s.Source() != SourceEnv {
		t.Errorf("Source = %q, want %q", s.Source(), SourceEnv)
	}
	// Environment credentials must never be written to disk.
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("environment credentials must not be persisted")
	}
}

func TestResolveDoesNotPatchPartialSession(t *testing.T) {
	isolate(t)
	t.Setenv(TokenEnv[0], "env-token")

	if err := (&Session{BaseURL: "https://192.168.1.1"}).Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	s, err := Resolve()
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if s.Token != "" {
		t.Errorf("Token = %q, want the half-configured session left alone", s.Token)
	}
	if _, err := s.Client(); err == nil {
		t.Error("Client should refuse to build a client without a token")
	}
}

func TestClearRemovesFile(t *testing.T) {
	isolate(t)
	if err := (&Session{BaseURL: "https://192.168.1.1", Token: "t"}).Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := Clear(); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	if err := Clear(); err != nil {
		t.Errorf("Clear on a missing file should be a no-op, got %v", err)
	}
}

func TestNormalizeBaseURL(t *testing.T) {
	tests := []struct {
		name        string
		in          string
		want        string
		wantChanged bool
	}{
		{"bare host gains https", "192.168.1.1", "https://192.168.1.1", true},
		{"http is upgraded", "http://192.168.1.1", "https://192.168.1.1", true},
		{"path is dropped", "https://192.168.1.1/admin", "https://192.168.1.1", true},
		{"trailing slash is dropped", "https://192.168.1.1/", "https://192.168.1.1", true},
		{"port is kept", "https://192.168.1.1:8443", "https://192.168.1.1:8443", false},
		{"already canonical", "https://router.lan", "https://router.lan", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, changed, err := NormalizeBaseURL(tc.in)
			if err != nil {
				t.Fatalf("NormalizeBaseURL(%q): %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("NormalizeBaseURL(%q) = %q, want %q", tc.in, got, tc.want)
			}
			if changed != tc.wantChanged {
				t.Errorf("changed = %v, want %v", changed, tc.wantChanged)
			}
		})
	}
}

func TestNormalizeBaseURLRejectsBadInput(t *testing.T) {
	for _, in := range []string{
		"",
		"   ",
		"ftp://192.168.1.1",
		"https://user:pass@192.168.1.1",
		"https://",
		"192.168.1.1 with spaces",
	} {
		if got, _, err := NormalizeBaseURL(in); err == nil {
			t.Errorf("NormalizeBaseURL(%q) = %q, want an error", in, got)
		}
	}
}

func TestClientRequiresBothCredentials(t *testing.T) {
	if _, err := (Session{Token: "t"}).Client(); err == nil {
		t.Error("expected an error without a base URL")
	}
	if _, err := (Session{BaseURL: "https://192.168.1.1"}).Client(); err == nil {
		t.Error("expected an error without a token")
	}
	c, err := (Session{BaseURL: "https://192.168.1.1", Token: "t"}).Client()
	if err != nil {
		t.Fatalf("Client: %v", err)
	}
	defer func() { _ = c.Close(context.Background()) }()
	if c.Token != "t" {
		t.Errorf("client token = %q, want the session token", c.Token)
	}
}

func TestLoadReportsMalformedFile(t *testing.T) {
	path := isolate(t)
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	_, err := Load()
	if err == nil {
		t.Fatal("expected an error for a malformed config file")
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("error should name the file, got %v", err)
	}
}
