# Changelog

## v1.1.1 — 2026-08-01

Documentation and housekeeping on top of v1.1.0; no behaviour change.

### Changed
- README now documents the v1.1.0 observability features (`WithMetrics`,
  `WithStructuredLogger`, `SDKVersion`) and the `IsRetryable()` / `RetryAfter`
  error-model additions.
- `example/main.go` demonstrates `WithMetrics` + `WithStructuredLogger` and
  prints collected stats; removed a duplicated code block.

### Removed
- `internal.NormalizeAddr` and `V4EndpointsByGroup` — exported helpers with
  no internal or external callers.

## v1.1.0 — 2026-07-31

Production hardening of the HTTP client layer: observability wiring, retry
safety, and retryability introspection. Additive — all v1.0.x callers keep
working.

### Added

- `SDKVersion` constant ("1.1.0") for runtime version introspection.
- `WithMetrics(*Metrics)` wires the existing (previously dead) `Metrics`
  collector into every request: `doOnce` now times each call and records
  duration + outcome. Read via `Client.Metrics()` / `Metrics.GetStats()`.
- `WithStructuredLogger(Logger)` attaches the leveled `Logger` interface
  from logger.go; retry / timeout / debug events flow through it. The
  printf-style `WithLogger` callback is retained for backward compatibility.
- `APIError.IsRetryable()` and `NetworkError.IsRetryable()` expose the
  SDK's own retryability judgement so applications can reuse it for custom
  retry loops or circuit breakers.
- `APIError.RetryAfter` carries the parsed `Retry-After` header (delta-seconds
  or HTTP-date) from 429/503 responses.

### Changed (behaviour)

- **Write operations are no longer retried on network errors.** Previously a
  POST/PUT/PATCH that failed at the transport layer was retried
  unconditionally, which risks duplicate writes since the request may have
  reached the router. Network errors are now retried only for idempotent
  verbs (GET/HEAD/OPTIONS/DELETE).
- `429 Too Many Requests` is now recognised as retryable, and `Retry-After`
  is honoured as the back-off when present.

### Compatibility

All v1.0.x callers continue to work. The write-retry change is a safety fix
that matches what a stateful router API actually requires; callers that
relied on retrying writes on transport errors should wrap the call in their
own idempotency guard.

## v1.0.1 — 2026-07-27

Brings the SDK to parity with `ikuai-cli` for the bits that v1.0.0
shipped as the generic `Do<Name>` / `Get<Name>` shape. The catalog
gains two optional fields that the codegen now understands.

### Added

- `V4Endpoint.Load` — marks monitoring load-style endpoints. The
  codegen now emits a typed `<Name>LoadOptions` struct plus a
  `Load<Name>` method that validates `DataType ∈ {hour, day, week,
  month}`, `Math ∈ {avg, max}`, `StartTime > 0`, `EndTime > 0` and
  `StartTime < EndTime`. Validation errors come back as
  `*ikuaiapi.APIError`. Covers `monitoring/cpu`, `memory`, `disk`,
  `cputemp`, `terminals`.
- `V4Endpoint.Action` — records the verb of action-style endpoints.
  The codegen now emits `Start<Name>`, `Stop<Name>`, `Restart<Name>`,
  `Sync<Name>`, `Restore<Name>`, `Check<Name>` instead of
  `Do<Name>(ctx, body)`. Covers 10 endpoints across `network`,
  `system` groups. No-body actions take no parameters; body-taking
  actions accept any JSON value.
- `Delete<Name>WithBody(ctx, body)` — companion to `Delete<Name>`
  for the rare cases that need a custom JSON body on DELETE (e.g.
  future catalog entries that target iKuai's body-style DELETE).
- `APIClient.Call` now forwards `params` to `DELETE` requests, so
  the `?srcfile=…` style of iKuai DELETE works through the escape
  hatch: `api.Call(ctx, "system", "backup", "DELETE", nil,
  map[string]string{"srcfile": "x"})`.
- `Client.FormatQuery` — small helper for callers that need to
  assemble a query string themselves.
- Per-group Field Hints comments appended to each generated
  service file (`service/<group>.go`). Covers 36 fields across
  `network` (NAT/DNAT), `objects`, `system`, `vpn`, `auth` groups.
  Hints live in `codegen/fieldHints` and can be expanded without
  regenerating from scratch.

### Changed

- `Delete<Name>(ctx, id)` now sends `?id=<n>` as a query parameter
  instead of `{"id": n}` in the JSON body. This matches what the
  iKuai firmware actually expects; the v1.0.0 body form was a
  no-op on the wire. Callers that need the body form can use the
  new `Delete<Name>WithBody`.
- `backup` in the catalog now has `DELETE` listed as a method
  (the missing verb that made `DELETE /system/backup?srcfile=…`
  unreachable through the typed helper in v1.0.0).

### Fixed

- The codegen's catalog parser was anchored to the end of the
  `Methods` array and therefore could not see trailing `Load: true`
  / `Action: "…"` fields. Reshaped the parser to scan the full
  composite literal so the new fields actually take effect.

### Compatibility

- All v1.0.0 callers continue to work. New helpers are additive;
  the existing `Get<Name>` / `Do<Name>` methods for non-Load,
  non-Action endpoints are unchanged.
- The Delete<Name> change is technically a wire change but matches
  what iKuai expects; switching to `Delete<Name>WithBody` is the
  escape hatch for the rare body-style DELETE.

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
