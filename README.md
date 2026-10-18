# ikuai-api (v4)

A focused Go SDK for the **iKuai v4.0 REST API** used by iKuai routers
(running iKuai OS 4.x). The SDK targets the v4 surface only — the v3
`/Action/call` protocol has been removed in this major version.

[![Go Report Card](https://goreportcard.com/badge/github.com/zy84338719/ikuai-api)](https://goreportcard.com/badge/github.com/zy84338719/ikuai-api)
[![GoDoc](https://godoc.org/github.com/zy84338719/ikuai-api?status.svg)](https://godoc.org/github.com/zy84338719/ikuai-api)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Features

- 🧩 **Built on the [oekit](https://oekit.github.io) contract** — errors
  carry an oekit `BizError` so an oekit service's `mw.Fail` understands
  them without an adapter; logging goes through the `logx` facade with a
  lazily resolved package logger. Two oekit modules are the only
  dependencies, and `make zero-dep` fails the build if a third-party
  require ever appears.
- 🧰 **157 typed methods** across 12 functional groups, generated
  directly from the v4 endpoint catalog.
- 🔍 **Per-endpoint query dialects** — iKuai v4 does not speak one query
  shape. Most collections use `page`/`limit`/`filter`/`order`/`order_by`,
  but monitoring and auth mix in an uppercase `TYPE`/`LIMIT`/`OFFSET`
  dialect, a `KEYWORDS`/`FINDS` hybrid, and a `starttime`/`stoptime`
  window. Every endpoint's real contract is recorded in the catalog, and
  a typed `Query<Name>` helper is generated for the ones that need it.
  Sending the wrong shape makes the router silently ignore the query —
  see [`skills/SKILL.md`](skills/SKILL.md).
- ✍️ **Write-body conventions encoded** — `enabled` is `"yes"`/`"no"`,
  selectors are `{"custom":[...],"object":[...]}`, object groups use a
  `group_value` array, and only a documented allowlist of fields may be
  JSON numbers. `CoerceBody`, `Enabled`, `AddressValues`, `ObjectGroup`
  and `UpdatePayload` (read-modify-write) handle all of it.
- 📋 **Per-endpoint body contracts** — 73 endpoints record their create
  defaults, update allowlist and required fields. `CreateBody` fills a
  create payload in and catches a missing required field locally rather
  than letting it surface as an opaque 3001; `UpdatePayload` applies the
  right allowlist automatically. Some contracts differ per verb — a DHCP
  service create writes no `opt_type*` fields while its update must send
  twelve — and the table records that too. Firmware inconsistencies are
  reproduced rather than smoothed over — see `skills/writing.md`.
- 🔁 **Retry with exponential back-off** for transient failures. The jitter
  is one-sided: a retry never fires *earlier* than its exponential
  baseline, so back-off cannot collapse into a retry storm against a
  struggling router.
- 🧪 **Dry-run mode** — every call returns a JSON preview of the
  request it would have made, no router traffic.
- 🔍 **Catalog discovery** — iterate all 157 endpoints at runtime with
  `service.Endpoints()` or `service.Path<Group>()`; `ep.QuerySpec()`
  decodes an endpoint's paging dialect.
- 🔐 **Credential management** (`session`) — session file with atomic
  0600 writes, environment fallback, and a documented precedence.
- 🖥️ **Response rendering** (`output`) — table for humans, compact JSON
  for scripts, with column selection, timestamp humanization, and
  whole-column dropping for narrow terminals.
- 🛡️ **iKuai-specific quirks handled** — the firmware emits bare `nil`
  instead of `null`; the SDK normalizes that before parsing. Envelope
  fields `data`, `results`, and `rowid` are normalized across
  endpoints. Error codes carry a hint, an actionable remedy, and the oekit
  business/HTTP code pair (3007 → 401, 10014 → 403, 3001 → 400).

## Installation

```bash
go get github.com/zy84338719/ikuai-api
```

Requires Go 1.24+. The only dependencies are `github.com/oekit/logx` and
`github.com/oekit/pkg`.

## Quick start

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/oekit/logx"

    ikuaiapi "github.com/zy84338719/ikuai-api"
    "github.com/zy84338719/ikuai-api/service"
)

func main() {
    client, err := ikuaiapi.NewClient("https://192.168.1.1",
        ikuaiapi.WithToken("deadbeefcafebabe1234567890abcdef"),
        ikuaiapi.WithTimeout(15*time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer func() { _ = client.Close(context.Background()) }()

    api := service.NewAPIClient(client)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    raw, err := api.Monitoring().GetMonitoringSystem(ctx)
    if err != nil {
        log.Fatal(err)
    }
    var overview map[string]any
    if err := json.Unmarshal(raw, &overview); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("hostname: %v\n", overview["hostname"])
}
```

## How to get a token

iKuai v4 uses Bearer tokens. Generate one in the router web UI:

1. Sign in to the router as `admin`.
2. Go to **System → Auth → API Token**.
3. Click **Generate** and copy the 32-character hex string.

The SDK exposes `ikuaiapi.ValidateToken(token)` for early validation.

## Service groups

| Group | Service | Endpoints |
| --- | --- | ---: |
| `advanced` | FTP / Samba / SNMP / HTTP | 6 |
| `auth` | users, packages, web services, online users | 4 |
| `interfaces` | LAN / WAN / physical / VLAN | 4 |
| `log` | arp, auth, dhcp, ddns, notice, pppoe, system, web, wireless | 9 |
| `monitoring` | system, cpu, memory, disk, network, clients, traffic, … | 37 |
| `network` | dhcp, dmz, dnat, dns, nat, pppoe, qos, vlan, ac | 25 |
| `objects` | domain / ip / ipv6 / mac / port / protocol / time | 7 |
| `routing` | static / policy / 5-tuple / load-balance / app-protocols | 6 |
| `security` | acl / mac / url / domain / peerconn / terminals | 13 |
| `system` | basic / alg / ntp / cpufreq / kernel / reboot / backup / upgrade / … | 28 |
| `vpn` | pptp / l2tp / openvpn / ikev2 / ipsec / wireguard | 10 |
| `wireless` | access-control, vlan | 2 |

Access any group via `api.<Group>()` (CamelCased). For example:

```go
api.Network().ListNetworkDhcpServices(ctx, &service.NetworkDhcpServicesListOptions{
    Page: 1, PageSize: 50, Order: "desc", OrderBy: "id",
})

api.System().GetSystemBasicConfig(ctx)

api.Monitoring().GetMonitoringClientsOnline(ctx)
```

## Method shapes

For each catalog entry, the generator emits one of two shapes:

1. **Single GET** — `<Group>Service.Get<Name>(ctx) (json.RawMessage, error)`
2. **CRUD** — a `<Name>ListOptions` struct plus
   `List<Name>` / `Get<Name>` / `Create<Name>` / `Update<Name>` /
   `Patch<Name>` (if PATCH is supported) / `Delete<Name>`. `Create`
   returns the `rowid` parsed out of the synthetic envelope the router
   emits on create responses.

The full path and supported verbs are documented above each generated
method — for example:

```go
// NetworkDhcpServices wraps network dhcp-services.
//
// Methods: GET, POST, PUT, PATCH
//
// Path: /network/dhcp/services
//
// Use ListNetworkDhcpServices(ctx, opts...) for paginated reads.
```

## Escape hatch: catalog-driven calls

If a generated helper does not exist yet, or you need to call a method
that is not in the catalog, use `APIClient.Call`:

```go
raw, err := api.Call(ctx, "interfaces", "wan-config", "GET", nil, nil)
```

`Call` resolves the (group, name) pair from the catalog and dispatches
the request with the supplied method, body and query params.

## Options

```go
client, _ := ikuaiapi.NewClient("https://192.168.1.1",
    ikuaiapi.WithToken("..."),                     // Bearer token
    ikuaiapi.WithTimeout(15*time.Second),          // per-request timeout
    ikuaiapi.WithInsecureSkipVerify(true),         // trust self-signed cert
    ikuaiapi.WithHTTPClient(customHTTP),           // bring your own http.Client
    ikuaiapi.WithAPIBase("/api/v4.0"),             // override the path prefix
    ikuaiapi.WithRawMode(true),                    // return the full envelope
    ikuaiapi.WithDryRun(true),                     // print requests, do not send
    ikuaiapi.WithRetry(3),                         // total attempt count
    ikuaiapi.WithRetryDelay(200*time.Millisecond, 5*time.Second),
    ikuaiapi.WithMetrics(ikuaiapi.NewMetrics()), // count requests / latency
    ikuaiapi.WithStructuredLogger(logx.NewStdLogger("poller", logx.LevelDebug)), // optional, narrows this client
)
```

## Observability

- **`WithMetrics(*Metrics)`** wires a request counter/latency collector into
  every call. Read snapshot stats with `client.Metrics().GetStats()` →
  `(requestCount, errorCount, avgDuration)`; handy for a `/metrics` endpoint
  or a health check. Use `Metrics.Reset()` between scrapes.
- **Logging goes through the oekit `logx` facade.** The SDK logs through
  `logx.Delegating("ikuai-api")`, resolved on every call, so installing a
  backend once in `main` is enough:

  ```go
  logx.UseSlog("my-service")              // structured JSON
  ikuaiapi.SetLogLevel(ikuaiapi.LogLevelInfo)
  ```

  `WithStructuredLogger(logx.Logger)` narrows the logger for one client
  without touching the global facade. Events are kv-style and carry the
  request context, so a trace id attaches automatically. There is no
  printf-style hook: `WithLogger` and `NewDefaultLogger` were removed in
  v1.2.0.
- **`SDKVersion`** is the semantic version string for runtime
  introspection.

## Error model

Two typed errors come back from every call. Router-side failures always
come back as `*APIError`; SDK-local validation failures carry a 4xxx
business code and wrap a sentinel from this package (`ErrRequiredField`,
`ErrInvalidToken`, …), so `errors.Is` can tell them apart from "the
router said no".

- `*ikuaiapi.APIError` — protocol-level failure: non-zero `code`,
  HTTP 4xx/5xx, or unparseable JSON. Fields: `HTTPStatus`, `Code`,
  `Message`, `Details` (per-field validation errors), and `RetryAfter`
  (parsed from the HTTP `Retry-After` header on 429/503). It embeds an
  oekit `BizError`, so `errors.As(err, &bizErr)` — and oekit's
  `mw.Fail` — reads `BizCode` and `HTTPCode` without an adapter.
  `IsAuthFailure()`, `IsPermissionFailure()`, `IsBadRequest()` and
  `Remedy()` classify and explain the failure.
- `*ikuaiapi.NetworkError` — transport failure: DNS, refused, TLS,
  timeout. Wraps the original `net` / `tls` / `http` error.

Both error types implement **`IsRetryable() bool`**, exposing the
SDK's own retryability judgement so applications can reuse it for custom
retry loops or circuit breakers: `APIError` is retryable on HTTP 429 / 5xx;
`NetworkError` is retryable. Note the SDK only auto-retries network errors
for **idempotent** verbs (GET/HEAD/OPTIONS/DELETE) — writes are never
auto-retried, since the request may have reached the router.

```go
_, err := api.Network().GetNetworkDnsConfig(ctx)
var apiErr *ikuaiapi.APIError
switch {
case errors.As(err, &apiErr):
    fmt.Printf("router said no: code=%d status=%d msg=%q\n",
        apiErr.Code, apiErr.HTTPStatus, apiErr.Message)
case errors.As(err, &netErr):
    fmt.Printf("transport: %v\n", netErr)
}
```

## Regenerating the service layer

The `service/` package is generated from `v4_catalog.go`. After editing
the catalog, regenerate:

```bash
go run ./codegen
```

The generator parses the catalog literal with a regex (no parent-package
import needed, so no import cycle), then writes one file per group plus
`service/root.go`. The output is deterministic and checked in.

## Project layout

```
.
├── README.md             — this file
├── go.mod                — two oekit modules, nothing else (enforced by make zero-dep)
├── auth.go               — token validation helpers
├── client.go             — net/http + retry + envelope + sanitization
├── doc.go                — package documentation
├── errors.go             — APIError (oekit BizError), NetworkError, sentinels
├── fields.go             — iKuai write-body conventions (see skills/writing.md)
├── logger.go             — logx.Delegating package logger + level floor
├── metrics.go            — request counters and latency
├── version.go            — Version enum (V4 only)
├── v4_catalog.go         — source of truth: 157 endpoints + query contracts
├── v4_schemas.go         — per-endpoint body contracts (73 endpoints)
├── codegen/              — service-layer generator (run with `go run ./codegen`)
├── service/              — generated, one file per group
│   ├── root.go           — APIClient entry point + Call()
│   ├── advanced.go
│   ├── auth.go
│   ├── interfaces.go
│   ├── log.go
│   ├── monitoring.go
│   ├── network.go
│   ├── objects.go
│   ├── routing.go
│   ├── security.go
│   ├── system.go
│   ├── vpn.go
│   └── wireless.go
├── session/              — credential persistence + resolution
├── output/               — table / JSON response rendering
├── skills/               — agent-facing guides (SKILL.md, monitoring, writing, rendering)
├── AGENTS.md             — hard rules for humans and agents
├── llms.txt              — machine-readable entry point
├── internal/             — small helpers shared by core + generated
├── example/              — runnable demo (env-driven)
└── *_test.go             — unit tests for every layer
```

## Testing

```bash
go test ./...
```

The tests use `httptest.Server` to stand in for a real router. They
cover:

- `SanitizeNil` edge cases (CRLF, escaped quotes, identifier
  boundaries).
- Token validation.
- Envelope handling for `data`, `results`, `rowid` and bare `nil`.
- Retry on 5xx.
- Dry-run mode.
- Every endpoint in the catalog (round-trip via `api.Call`).

## License

MIT. See [LICENSE](LICENSE).
