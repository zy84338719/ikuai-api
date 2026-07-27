# Changelog

## v1.0.0 — 2026-07-27

The first v4-only stable release. The entire v3 surface is gone, the
SDK is regenerated from a single source-of-truth catalog, and the
external dependency on `imroc/req/v3` has been removed in favour of
the Go standard library.

### Breaking changes

- v3 (`/Action/call` + `func_name/action/param`) is no longer
  supported. v3-compatible iKuai OS is end-of-life; the dual-version
  abstraction forced every caller to pick a client at compile time
  and was a recurring source of confusion.
- `ikuaiapi.NewClient(baseURL, user, pass)` is gone. The v4 API
  requires a Bearer token obtained from the router web UI, not
  username/password login. Replace with:

  ```go
  client, _ := ikuaiapi.NewClient("https://192.168.1.1",
      ikuaiapi.WithToken("<32-hex-token>"))
  ```

- `*ikuaiapi.V3ActionClient`, `*ikuaiapi.V4Client`, the v3
  `protocolForVersion` switch and the `Call(funcName, action, param)`
  helper have been deleted. Use the typed methods on
  `service.APIClient` or the `APIClient.Call` escape hatch.
- `types.BaseRequest`, `types.LoginRequest`, `types.BaseResponse` and
  every `*ShowResponse` wrapper struct are gone. v4 returns the
  payload directly under `data` (or `results`, or `rowid` for
  create-style endpoints), so the SDK normalises the envelope and
  hands callers `json.RawMessage`. Decode into your own struct.
- `concurrent.go`, `cache.go`, `utils/` are gone. Retry is built into
  the client (`WithRetry(3)`) and the response cache was an
  over-engineered workaround.
- `service.NewAPIClient` now returns a pointer (`*APIClient`) and the
  group accessors are `Network()` / `Monitoring()` / `System()` etc.
  The per-group fields are unexported.
- Errors are now `*ikuaiapi.APIError` and `*ikuaiapi.NetworkError`.
  The old `*ikuaiapi.SDKError` with `ErrorCode` enums is gone. Use
  `errors.As` to type-assert.

### Added

- `v4_catalog.go` is the single source of truth: **151 v4 REST
  endpoints** across 13 groups, merged from the previous SDK and the
  upstream `ikuai-cli` project so the SDK now exposes every path the
  CLI implements (including backup, upgrade, web-admin, AC, disks,
  files, system upgrade status/check/start).
- `codegen/` — a tiny generator that reads `v4_catalog.go` and writes
  one `service/<group>.go` per group plus `service/root.go`. Total:
  13 generated files, 151 typed methods, all generated. Re-run with
  `go run ./codegen` after editing the catalog.
- Every catalog entry becomes a typed method with a consistent shape:
  - Single GET → `Get<Name>(ctx) (json.RawMessage, error)`
  - Read-write → `List<Name>` / `Get<Name>` / `Create<Name>` /
    `Update<Name>` / `Patch<Name>` (if supported) / `Delete<Name>`,
    with a `<Name>ListOptions` struct that mirrors the iKuai
    pagination convention (`page` / `page_size` / `filter` / `order`
    / `order_by`).
  - Action endpoints → `Do<Name>(ctx, body)`.
  - `Create` returns the `rowid` parsed from the synthetic envelope.
- `APIClient.Call(ctx, group, name, method, body, params)` — the
  lowest-level escape hatch for endpoints the typed helpers do not
  cover yet.
- `ikuaiapi.ValidateToken(token)` — early check for the 32-hex
  format the router issues.
- `APIClient.Endpoints()` and `service.Path<Group>()` — runtime
  catalog discovery.

### HTTP layer

- `net/http` from the Go standard library. `go.mod` now lists **no
  external dependencies**.
- `SanitizeNil` normalises the iKuai firmware habit of emitting bare
  `nil` instead of `null` (tracks string state to avoid corrupting
  string content).
- Envelope normalisation: `data` (preferred) → `results` (monitor
  endpoints) → synthetic `{message, rowid}` envelope (create
  responses).
- Exponential-backoff retry with full jitter, default 3 attempts.
  Limited to idempotent methods (GET / HEAD / OPTIONS / DELETE) and
  to 5xx responses.
- Per-request timeout via `WithTimeout`. Same context controls
  cancellation.
- `WithDryRun(true)` returns a JSON preview of the request without
  contacting the router. Useful for scripting / verification.
- `WithRawMode(true)` returns the full envelope (including `code`,
  `message`, `rowid`).
- `WithInsecureSkipVerify(true)` is the default — iKuai routers use
  self-signed certificates. Pass your own `*http.Client` via
  `WithHTTPClient` if you need a custom CA or proxy.
- Error hints for codes 3001 (parameter), 3007 (token invalid) and
  1008 (session expired) are appended to `APIError.Message`.
- Bearer token auth via `Authorization: Bearer <token>`.

### Tests

- 22 unit tests covering:
  - `SanitizeNil` edge cases (CRLF, escaped quotes, identifier
    boundaries, nested arrays).
  - Token validation.
  - Envelope handling for `data`, `results`, `rowid` and bare `nil`.
  - Retry on 5xx with back-off.
  - Dry-run mode (no router traffic).
  - Pagination query string assembly.
  - Network-error wrapping.
  - `Call(group, name, …)` round-trip for **every** catalog entry.
  - List/Create/Delete shaped methods (rowid, `id` in body).

### Files

- 86 files changed.
- Removed: `v3_action.go`, `v3_catalog.go`, `v4_rest.go`, `protocol.go`,
  `concurrent.go`, `cache.go`, `types/` (17 files), `service/`
  (15 v3-style files), `utils/`, `git-commit.sh`, old tests.
- Added: `codegen/`, `service/<group>.go` (13 generated files),
  `service/codegen_test.go`, `example/main.go`, `LICENSE`.
- Rewrote: `client.go` (stdlib + retry), `auth.go` (token
  validation), `errors.go` (`APIError`/`NetworkError`), `version.go`
  (v4 only), `internal/util.go` (kept only `NormalizeAddr`),
  `Makefile`, `README.md`, `.gitignore`.

### Migration cheatsheet

```diff
- client := ikuaiapi.NewClient("http://192.168.1.1", "admin", "pass")
- client.Login(ctx)
+ client, _ := ikuaiapi.NewClient("https://192.168.1.1",
+     ikuaiapi.WithToken("<32-hex-token>"),
+     ikuaiapi.WithInsecureSkipVerify(true),
+ )

- api := service.NewAPIClient(client)
- api.System().GetHomepage(ctx) // returns *types.HomepageSysStat
+ api := service.NewAPIClient(client)
+ raw, _ := api.Monitoring().GetMonitoringSystem(ctx)
+ var overview map[string]any
+ _ = json.Unmarshal(raw, &overview)

- if err := client.Call(ctx, "webuser", "show", nil, &resp); err != nil { ... }
+ if raw, err := api.Auth().ListAuthUsers(ctx, nil); err != nil { ... }

- if e, ok := err.(*ikuaiapi.SDKError); ok && e.Code == ErrCodeNotLoggedIn { ... }
+ var apiErr *ikuaiapi.APIError
+ if errors.As(err, &apiErr) && apiErr.Code == 3007 { ... }
```

## v0.2.0 — earlier

Initial v3 / v4 dual-version SDK, `req/v3` HTTP client. Superseded by
v1.0.0.
