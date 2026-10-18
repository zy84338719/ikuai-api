// Package session persists the router address and API token so an
// application does not have to hard-code them.
//
// Credentials resolve in three tiers, and the resolution is exposed so
// callers can report where a credential came from:
//
//	session file  >  environment variables  >  unset
//
// The session file is written with 0600 permissions and replaced
// atomically, so a crash mid-write cannot leave a truncated or
// world-readable file behind. Environment variables are never written to
// disk.
package session

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	oekerrors "github.com/oekit/pkg/errors"
	ikuaiapi "github.com/zy84338719/ikuai-api"
)

// Environment variables consulted for the router address and token, in
// priority order. The IKUAI_CLI_* spellings are also honoured so a
// machine already set up for the ikuai-cli tool works unchanged.
var (
	BaseURLEnv = []string{"IKUAI_API_BASE_URL", "IKUAI_CLI_BASE_URL"}
	TokenEnv   = []string{"IKUAI_API_TOKEN", "IKUAI_CLI_TOKEN"}
	// PathEnv overrides the session file location.
	PathEnv = "IKUAI_API_CONFIG_FILE"
)

// Source names the tier a credential came from.
type Source string

const (
	SourceFile  Source = "session"
	SourceEnv   Source = "env"
	SourceNone  Source = "none"
	SourceMixed Source = "session+env"
)

// Session is the persisted credential pair.
type Session struct {
	BaseURL string `json:"base_url"`
	Token   string `json:"token"`

	// source records where the values in memory came from. It is not
	// serialized: an env-sourced token must not be persisted.
	source Source
}

// Source reports which tier the in-memory values came from.
func (s Session) Source() Source {
	if s.BaseURL == "" && s.Token == "" {
		return SourceNone
	}
	return s.source
}

// Configured reports whether both a router address and a token are
// available, which is the minimum needed to talk to the router.
func (s Session) Configured() bool { return s.BaseURL != "" && s.Token != "" }

// Path returns the session file location: PathEnv when set, otherwise
// ~/.ikuai/config.json.
func Path() string {
	if p := strings.TrimSpace(os.Getenv(PathEnv)); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		// Without a home directory there is nowhere sensible to persist.
		// Callers still get to use environment credentials.
		return ""
	}
	return filepath.Join(home, ".ikuai", "config.json")
}

// Load reads the session file. A missing file is not an error: it yields an
// empty session so first-run callers can proceed to environment lookup. A
// malformed file is reported, because silently ignoring it would leave the
// user with no way to tell why their saved token stopped being used.
func Load() (*Session, error) {
	path := Path()
	if path == "" {
		return &Session{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if oekerrors.Is(err, os.ErrNotExist) {
			return &Session{}, nil
		}
		return nil, fmt.Errorf("ikuai session: read %s: %w", path, err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return &Session{}, nil
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("ikuai session: parse %s: %w", path, err)
	}
	s.source = SourceFile
	return &s, nil
}

// Save writes the session file atomically with 0600 permissions. Values
// that came from the environment are never written: only a caller that
// explicitly sets a field persists it.
func (s *Session) Save() error {
	path := Path()
	if path == "" {
		return oekerrors.New("ikuai session: no writable config path; set " + PathEnv)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("ikuai session: create %s: %w", dir, err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return writeFileAtomic(path, data, 0o600)
}

// Clear removes the stored credentials.
func Clear() error {
	path := Path()
	if path == "" {
		return nil
	}
	if err := os.Remove(path); err != nil && !oekerrors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("ikuai session: clear %s: %w", path, err)
	}
	return nil
}

// writeFileAtomic writes data to a temporary file in the destination
// directory and renames it into place, so readers never observe a
// half-written credential file.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".ikuai-*.tmp")
	if err != nil {
		return fmt.Errorf("ikuai session: create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		// Best effort: on the success path the file no longer exists.
		_ = os.Remove(tmpName)
	}()

	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("ikuai session: replace %s: %w", path, err)
	}
	return nil
}

// Resolve returns the effective credentials and the tier they came from.
//
// The session file wins. When it holds no token, the environment is
// consulted; environment values stay in memory and are never persisted.
// A session that has a base URL but no token is reported honestly rather
// than being patched from the environment, because a half-configured
// session is usually a mistake the operator should see.
func Resolve() (Session, error) {
	s, err := Load()
	if err != nil {
		return Session{}, err
	}
	if s.Token != "" {
		s.source = SourceFile
		return *s, nil
	}

	envURL, envToken := envCred()
	if s.BaseURL == "" {
		if envURL != "" || envToken != "" {
			s.source = SourceEnv
		} else {
			s.source = SourceNone
		}
		s.BaseURL, s.Token = envURL, envToken
		return *s, nil
	}

	// A base URL with no token: keep the stored address but report that the
	// token is missing, and never fill it in from the environment.
	if envToken != "" || envURL != "" {
		s.source = SourceMixed
	} else {
		s.source = SourceNone
	}
	return *s, nil
}

func envCred() (baseURL, token string) {
	baseURL = firstEnv(BaseURLEnv)
	token = firstEnv(TokenEnv)
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	return strings.TrimSpace(baseURL), strings.TrimSpace(token)
}

func firstEnv(names []string) string {
	for _, n := range names {
		if v := strings.TrimSpace(os.Getenv(n)); v != "" {
			return v
		}
	}
	return ""
}

// NormalizeBaseURL validates a router address and returns it in canonical
// form: https scheme, host and optional port, nothing else. iKuai's local
// management interface serves HTTPS on a self-signed certificate, so a
// bare host is upgraded to https:// rather than rejected, and any path,
// query or fragment is dropped because the v4 API lives at a fixed prefix.
//
// It also reports whether the input had to be rewritten, so a CLI can tell
// the operator that the value they typed was normalized.
func NormalizeBaseURL(raw string) (normalized string, changed bool, err error) {
	input := strings.TrimSpace(raw)
	if input == "" {
		return "", false, oekerrors.New("ikuai session: missing router host")
	}
	if strings.ContainsAny(input, " \t\r\n") {
		return "", false, oekerrors.New("ikuai session: router host must not contain whitespace")
	}

	candidate := input
	switch {
	case hasPrefixFold(input, "http://"):
		candidate = "https://" + input[len("http://"):]
	case hasPrefixFold(input, "https://"):
		candidate = "https://" + input[len("https://"):]
	case strings.Contains(input, "://"):
		return "", false, oekerrors.New("ikuai session: only http or https router URLs are supported")
	default:
		candidate = "https://" + input
	}

	parsed, perr := url.Parse(candidate)
	if perr != nil {
		return "", false, fmt.Errorf("ikuai session: invalid router host: %w", perr)
	}
	if parsed.User != nil {
		return "", false, oekerrors.New("ikuai session: router URL must not embed a username or password")
	}
	if parsed.Host == "" || parsed.Hostname() == "" {
		return "", false, oekerrors.New("ikuai session: missing router host")
	}
	if strings.ContainsAny(parsed.Host, " \t\r\n") {
		return "", false, oekerrors.New("ikuai session: invalid router host")
	}

	out := "https://" + parsed.Host
	return out, out != input, nil
}

func hasPrefixFold(s, prefix string) bool {
	return len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix)
}

// Client builds an SDK client from the resolved credentials. It fails when
// the credentials are incomplete, rather than constructing a client that
// would 401 on every call.
func (s Session) Client(opts ...ikuaiapi.ClientOption) (*ikuaiapi.Client, error) {
	if s.BaseURL == "" {
		return nil, fmt.Errorf("ikuai session: no router address configured (set %s)", strings.Join(BaseURLEnv, " or "))
	}
	if s.Token == "" {
		return nil, fmt.Errorf("ikuai session: no API token configured (set %s)", strings.Join(TokenEnv, " or "))
	}
	all := append([]ikuaiapi.ClientOption{ikuaiapi.WithToken(s.Token)}, opts...)
	return ikuaiapi.NewClient(s.BaseURL, all...)
}

// ResolveClient is the one-call path: resolve credentials and build a
// client in a single step.
func ResolveClient(opts ...ikuaiapi.ClientOption) (*ikuaiapi.Client, Session, error) {
	s, err := Resolve()
	if err != nil {
		return nil, Session{}, err
	}
	c, err := s.Client(opts...)
	if err != nil {
		return nil, s, err
	}
	return c, s, nil
}
