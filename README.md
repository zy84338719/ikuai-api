# ikuai-api (v4)

A focused Go SDK for the **iKuai v4.0 REST API** used by iKuai routers
(running iKuai OS 4.x). The SDK targets the v4 surface only — the v3
`/Action/call` protocol has been removed in this major version.

[![Go Report Card](https://goreportcard.com/badge/github.com/zy84338719/ikuai-api)](https://goreportcard.com/badge/github.com/zy84338719/ikuai-api)
[![GoDoc](https://godoc.org/github.com/zy84338719/ikuai-api?status.svg)](https://godoc.org/github.com/zy84338719/ikuai-api)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Features

- 🚀 **Zero third-party deps** — uses only the Go standard library
  (`net/http`, `encoding/json`, `context`).
- 🧰 **151 typed methods** across 13 functional groups, generated
  directly from the v4 endpoint catalog.
- 🔁 **Retry with exponential back-off** for transient failures; per
  request and total timeout controlled by the same context.
- 🧪 **Dry-run mode** — every call returns a JSON preview of the
  request it would have made, no router traffic.
- 🔍 **Catalog discovery** — iterate all 151 endpoints at runtime with
  `service.Endpoints()` or `service.Path<Group>()`.
- 🛡️ **iKuai-specific quirks handled** — the firmware emits bare `nil`
  instead of `null`; the SDK normalizes that before parsing. Envelope
  fields `data`, `results`, and `rowid` are normalized across
  endpoints.

## Installation

```bash
go get github.com/zy84338719/ikuai-api
```

## Quick start

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

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
    defer client.Close()

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
    ikuaiapi.WithLogger(func(format string, args ...any) { log.Printf(format, args...) }),
)
```

## Error model

Two typed errors come back from every call:

- `*ikuaiapi.APIError` — protocol-level failure: non-zero `code`,
  HTTP 4xx/5xx, or unparseable JSON. Fields: `HTTPStatus`, `Code`,
  `Message`, `Details` (per-field validation errors).
- `*ikuaiapi.NetworkError` — transport failure: DNS, refused, TLS,
  timeout. Wraps the original `net` / `tls` / `http` error.

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
├── go.mod                — zero external dependencies
├── auth.go               — token validation helpers
├── client.go             — net/http + retry + envelope + sanitization
├── errors.go             — APIError, NetworkError, error hints
├── version.go            — Version enum (V4 only)
├── v4_catalog.go         — the source of truth for all 151 endpoints
├── logger.go             — optional structured logging
├── codegen/              — service-layer generator (run with `go run ./codegen`)
├── service/              — generated, 13 files, one per group
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
