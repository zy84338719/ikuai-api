# ikuai-api

[![Go Report Card](https://goreportcard.com/badge/github.com/zy84338719/ikuai-api)](https://goreportcard.com/badge/github.com/zy84338719/ikuai-api)
[![GoDoc](https://godoc.org/github.com/zy84338719/ikuai-api?status.svg)](https://godoc.org/github.com/zy84338719/ikuai-api)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A Go SDK for the **iKuai v4.0 REST API** — the local management API served by
routers running iKuai OS 4.x. v3's `/Action/call` protocol is not supported.

The distinguishing thing about this SDK is that it does not treat the v4
API as uniform, because it isn't. The page-size parameter, the sort keys,
the query dialect and the response envelope all differ between groups of
endpoints, and getting any of them wrong makes the router **silently ignore
you** rather than fail. Those contracts are recorded in
[`v4_catalog.go`](v4_catalog.go) and [`v4_schemas.go`](v4_schemas.go) and
turned into typed helpers, so callers write `QueryMonitoringDownstream(ctx,
opts)` instead of hand-assembling a query string that quietly does nothing.

## Contents

- [Install](#install)
- [Quick start](#quick-start)
- [Three contracts to know before you write anything](#three-contracts-to-know-before-you-write-anything)
- [Reading](#reading)
- [Writing](#writing)
- [Response envelopes](#response-envelopes)
- [Errors](#errors)
- [Logging and metrics](#logging-and-metrics)
- [Credentials](#credentials)
- [Rendering](#rendering)
- [Service groups](#service-groups)
- [Options](#options)
- [Generated API shapes](#generated-api-shapes)
- [Development](#development)
- [Project layout](#project-layout)
- [Verification status](#verification-status)
- [Known limitations](#known-limitations)

## Install

```bash
go get github.com/zy84338719/ikuai-api
```

Requires Go 1.24+. The only dependencies are `github.com/oekit/logx` and
`github.com/oekit/pkg` — `make zero-dep` fails the build if a third-party
require ever appears.

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
        ikuaiapi.WithToken(token),
        ikuaiapi.WithTimeout(15*time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }
    // Close takes a context, returns an error, and is safe to call twice.
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
    fmt.Printf("hostname: %v  uptime: %v\n", overview["hostname"], overview["uptime"])
}
```

### Getting a token

iKuai v4 authenticates with a Bearer token, generated in the router web UI:
**System → Auth → API Token**.

Two token shapes exist in the field, and both work:

| Shape | Looks like | Seen on |
|---|---|---|
| hex | 32 characters, `0-9a-f` | older firmware |
| base64 | 32 characters, decodes to a UUID prefix | current 4.x firmware |

A base64 token authenticates fine on every endpoint — it just fails a
hex-shaped validator, which is why `ValidateToken` accepts both rather than
rejecting tokens that work. It is a shape check to catch a truncated
paste, not a pre-authentication: the router is the authority.

```go
if err := ikuaiapi.ValidateToken(token); err != nil {
    return fmt.Errorf("token looks wrong: %w", err)
}
```

## Three contracts to know before you write anything

Each of these was found the hard way, and each one fails silently.

### 1. The page-size parameter is `limit`

`page_size` is accepted by the router and **ignored** — you get every row
back. Against a collection of nine online clients:

| query | rows returned |
|---|---|
| `limit=1` / `3` / `5` | 1 / 3 / 5 |
| `page_size=1` / `3` / `5` | 9 / 9 / 9 |

Page 2 with `limit=3` returns a different slice; the same page requested
with `page_size` repeats page 1 exactly.

The generated `List<Name>` helpers send `limit`.

### 2. Sort keys are `order` = direction, `order_by` = field

Everywhere, including the security and advanced-service collections.
Swapping the two returns **no rows at all**, on a table that plainly has
data. `Order` always means the direction and `OrderBy` always means the
field, on every endpoint.

### 3. Collections speak seven different query dialects

Sending the wrong shape is silently ignored, exactly like `page_size`.

| Dialect | Endpoints | Query |
|---|---|---|
| paged | most collections | `page`, `limit`, `filter`, `order`, `order_by` |
| legacy | `/monitoring/downstream`, `/monitoring/cameras`, `/monitoring/switch` | `TYPE=data`, `LIMIT="<offset>,<count>"`, `OFFSET`, `ORDER`, `ORDER_BY` |
| mixed | `/auth/online-users` | `page`, `limit` + `KEYWORDS`, `FINDS`, `ORDER`, `ORDER_BY` |
| window | `/monitoring/protocols*`, `/monitoring/clients/protocols*`, `/log/url-visits` | `starttime`, `stoptime` (no underscore) |
| none | `/monitoring/ssid-clients`, `channel-clients`, `wireless-traffic`, `clients-traffic-load` | only the endpoint's own selectors |
| load | `/monitoring/cpu`, `memory`, `disk`, `cputemp`, `terminals`, `connections`, `network` | `datetype`, `start_time`, `end_time`, `math` (underscored) |
| ref | `/…-objects/ref` | `group_name`, no paging |

29 endpoints carry a recorded dialect; the rest use the default. For the
non-default ones the generator emits a `Query<Name>` helper that builds the
right shape and rejects an incomplete request before it leaves the process:

```go
// legacy dialect: LIMIT is "<offset>,<count>", there is no "page"
raw, err := api.Monitoring().QueryMonitoringDownstream(ctx,
    &service.MonitoringDownstreamQueryOptions{
        Page: 1, PageSize: 20,
        Order: "desc", OrderBy: "id",
        Status: "1", Device: "camera",
    })

// window dialect: ip and mac are required
raw, err = api.Monitoring().QueryMonitoringClientsProtocols(ctx,
    &service.MonitoringClientsProtocolsQueryOptions{
        IP: "192.168.1.100", MAC: "08:9b:4b:01:7e:7c",
        StartTime: 1773300000, StopTime: 1773303600,
    })
```

Inspect any endpoint's contract at runtime:

```go
for _, ep := range service.Endpoints() {
    dialect, required, optional, err := ep.QuerySpec()
    // ...
}
```

## Reading

```go
// plain collection: page/limit/filter/order/order_by
raw, err := api.Network().ListNetworkDhcpServices(ctx,
    &service.NetworkDhcpServicesListOptions{
        Page: 1, PageSize: 50,
        Filter: "enabled==yes",
        Order: "desc", OrderBy: "id",
    })

// no parameters at all
raw, err = api.Monitoring().GetMonitoringSystem(ctx)

// load window over a time bucket
raw, err = api.Monitoring().LoadCpu(ctx, &service.MonitoringCpuLoadOptions{
    DataType: "hour", // hour | day | week | month
    StartTime: time.Now().Add(-time.Hour).Unix(),
    EndTime:   time.Now().Unix(),
    Math:      "avg", // avg | max
})
```

### Two-level resources

WireGuard peers live under a tunnel, so both identifiers are path
parameters:

```go
api.Vpn().ListVpnWireguardPeers(ctx, opts)
api.Vpn().GetVpnWireguardPeer(ctx, tunnelID, peerID)
api.Vpn().CreateVpnWireguardPeers(ctx, tunnelID, body)
api.Vpn().DeleteVpnWireguardPeer(ctx, tunnelID, peerID)
```

An endpoint whose path ends in a placeholder gets no `List` — there is no
collection at that path.

## Writing

The firmware validates writes strictly, and for most rule-like resources a
partial update is **rejected** rather than completed with the router's own
defaults — which would silently reconfigure whatever you left out. 73
endpoints record their create defaults, update allowlist and required
fields; the helpers apply them.

```go
// create: defaults filled in, required fields checked locally
body, err := ikuaiapi.CreateBody("/security/acl-rules", map[string]any{
    "tagname":  "block_ssh",
    "action":   "drop",
    "protocol": "tcp",
    "dst_port": ikuaiapi.AddressValues("22"),
})
// enabled=yes prio=50 dir=forward, untouched selectors are the empty pair,
// time is the all-day window — and a missing required field is reported
// here rather than surfacing later as an opaque 3001.
rowid, err := api.Security().CreateSecurityAclRules(ctx, body)

// update: read-modify-write with the endpoint's own allowlist
updated, err := ikuaiapi.UpdatePayload(ctx, client, "/network/nat/rules", id,
    map[string]any{"nat_addr": "203.0.113.5"}, nil) // nil = use the schema's list
```

### Body conventions the firmware insists on

| Convention | Helper |
|---|---|
| `enabled` is the string `"yes"` / `"no"`, never a JSON boolean | `Enabled(true)` |
| address/port/protocol selectors are `{"custom":[…],"object":[]}`, not an array | `AddressValues("80,443")` |
| object-group members live in `group_value`, one single-key object each | `ObjectGroup("ip", "10.0.0.1")` |
| the time group holds schedules instead of values | `TimeObjectGroup(...)`, `WeekSchedule()` |
| only an allowlisted set of fields may be JSON numbers | `CoerceBody(body)` |

That last one is the subtle one. Most numeric-looking fields must stay
strings, or an SNMP community of `"161"` or a VLAN name of `"0010"` gets
mangled. `IsIntegerAPIField` and `IntegerAPIFields` expose the
classification.

Three firmware behaviours are reproduced rather than smoothed over,
because normalizing them breaks the write — see
[`skills/writing.md`](skills/writing.md) for the full list.

## Response envelopes

Three wrapping shapes come back, and confusing them is easy:

| Shape | Example | Where |
|---|---|---|
| `data` | `{"data":[…],"total":n}` | most collections |
| `results` | `{"results":[…]}` | monitoring load endpoints |
| group paging | `{"ip_data":[],"ip_total":0}` | object groups, DHCP static |

The third is the trap: `/ip-objects` answers with `ip_data`, `/ip6-objects`
with `ipv6_data`, and the members are **not** at the top level. Read the
array and the counter separately:

```go
var page struct {
    IPData  []map[string]any `json:"ip_data"`
    IPTotal int              `json:"ip_total"`
}
json.Unmarshal(raw, &page)
```

`FirstObject` unwraps all three, and a read-modify-write update goes
through it. An empty group correctly reports "no object" rather than
handing back the wrapper with its counters as if they were a row.

The client also normalizes the bare `nil` tokens some firmware emits in
place of `null` before parsing, and `Create` methods return the `rowid` the
router reports on success.

## Errors

Router-side failures come back as `*ikuaiapi.APIError`, which embeds an
[oekit](https://oekit.github.io) `BizError`. The iKuai code becomes the business code, and an oekit
handler's `mw.Fail` reads a status without any adapter:

| iKuai code | Meaning | BizCode | HTTPCode | Predicate |
|---|---|---|---|---|
| 3001 | parameter error | 3001 | 400 | `IsBadRequest()` |
| 3007 | token expired/invalid | 3007 | 401 | `IsAuthFailure()` |
| 1008 | session expired | 1008 | 401 | `IsAuthFailure()` |
| 10014 | permission denied | 10014 | 403 | `IsPermissionFailure()` |
| 4xxx | SDK-local validation | 4900 | 400 | — |

The last row is a deliberate split: a 4xxx code means **the request was
never sent**. An SDK-local failure — missing options, a page number of
zero, an endpoint absent from the catalog — answers 400 rather than
falling through to 500, so a service does not blame its own server for a
caller's mistake.

```go
_, err := api.Network().GetNetworkDnsConfig(ctx)

var apiErr *ikuaiapi.APIError
switch {
case errors.As(err, &apiErr):
    fmt.Printf("code=%d status=%d msg=%q\n",
        apiErr.Code, apiErr.Biz().HTTPCode, apiErr.Message)
    fmt.Println("try:", apiErr.Remedy())
}
```

`*ikuaiapi.NetworkError` wraps transport failures (DNS, refused, TLS,
timeout) and also carries the original error for `errors.Is`.

Both types implement `IsRetryable() bool`. The SDK retries transient
failures with exponential back-off and honours `Retry-After`. Its jitter is
one-sided — a retry never fires *earlier* than its exponential baseline,
so back-off cannot collapse into a retry storm against a struggling
router. A write is never auto-retried on a transport error, because that
request may already have reached the device.

## Logging and metrics

Logging goes through the oekit `logx` facade with a lazily resolved
package logger, so the SDK never picks a backend or a service name on your
behalf. Install one in `main` and everything follows:

```go
logx.UseSlog("my-service") // structured JSON
ikuaiapi.SetLogLevel(ikuaiapi.LogLevelInfo)
```

What it emits is request metadata — method, URL, retry count, delay — and
never the bearer token. If a URL is itself sensitive on your network,
redact by key before installing:

```go
logx.SetDefault(logx.WithFilter(logx.UseSlog("svc"),
    logx.FilterKey("url"), logx.FilterKey("path")))
```

`WithStructuredLogger(logx.Logger)` narrows the logger for one client
without touching the global facade.

For counters, `Metrics` collects request count, failure count and mean
latency without forcing a metrics backend on you:

```go
metrics := ikuaiapi.NewMetrics()
client, _ := ikuaiapi.NewClient(base, ikuaiapi.WithToken(t), ikuaiapi.WithMetrics(metrics))
// later
count, errs, avg := metrics.GetStats()
```

## Credentials

Rather than hard-coding an address and a token, resolve them:

```go
client, resolved, err := session.ResolveClient()
fmt.Println("credentials came from:", resolved.Source()) // session | env | none
```

Resolution is session file → environment → unset, and the source is
reported rather than guessed. The session file is written with 0600
permissions and replaced atomically; environment values stay in memory and
are never persisted.

```go
export IKUAI_API_BASE_URL=https://192.168.1.1   # or IKUAI_CLI_BASE_URL
export IKUAI_API_TOKEN=<token>                  # or IKUAI_CLI_TOKEN
```

`NormalizeBaseURL` canonicalizes an address — upgrading a bare host to
https, stripping any path — because the v4 API lives at a fixed prefix and
a stray path silently breaks every call.

## Rendering

`output` turns a response into something readable or pipeable:

```go
r := output.New(os.Stdout, output.FormatAuto) // table on a TTY, JSON when piped
r = r.WithOptions(output.RenderOptions{
    HumanTime: true,
    Columns:   []string{"ip_addr", "mac", "upload", "download"},
})
err := r.Value(decoded)
```

It flattens the firmware's own structures — the `{"custom":[…],"object":[…]}`
selector pair and `group_value` members render as lines rather than
opaque JSON — converts `timestamp` fields to local time, converts headers
to `UPPER_SNAKE_CASE`, and drops whole trailing columns rather than
truncating a value in a narrow terminal.

## Service groups

| Group | Covers | Endpoints |
|---|---|---:|
| `advanced` | FTP, HTTP, Samba, SNMPD | 6 |
| `auth` | accounts, packages, web portal, online sessions | 4 |
| `interfaces` | LAN, WAN, WAN-VLAN, physical NICs | 4 |
| `log` | system, arp, auth, dhcp, pppoe, web, ddns, notice, wireless, url-visits | 10 |
| `monitoring` | system, cpu/memory/disk/…, clients, traffic, protocols, wireless | 37 |
| `network` | dhcp, dhcpv6, vlan, nat/dnat/dmz, dns, qos, pppoe, ac | 25 |
| `objects` | 7 object groups + 7 reverse-lookup endpoints | 14 |
| `routing` | static routes, stream shunting (domain / 5-tuple / l7 / load-balance / updown) | 6 |
| `security` | acl, mac, l7, url rules, domain blacklist, peerconn, terminals, advanced | 13 |
| `system` | basic, ntp, alg, kernel, cpufreq, disks, files, backup, upgrade, web-admin, … | 24 |
| `vpn` | pptp / l2tp / openvpn / ikev2 servers and clients, ipsec, wireguard + peers | 12 |
| `wireless` | blacklist, VLAN rules | 2 |

Reach any group with `api.<Group>()`:

```go
api.Network().ListNetworkDhcpServices(ctx, opts)
api.System().GetSystemBasicConfig(ctx)
api.Monitoring().QueryMonitoringClientsOnline(ctx, opts)
```

## Options

```go
client, err := ikuaiapi.NewClient("https://192.168.1.1",
    ikuaiapi.WithToken(token),                  // Bearer token
    ikuaiapi.WithTimeout(15*time.Second),       // per-request timeout
    ikuaiapi.WithInsecureSkipVerify(true),      // trust the router's self-signed cert
    ikuaiapi.WithHTTPClient(custom),            // bring your own *http.Client
    ikuaiapi.WithAPIBase("/api/v4.0"),          // override the path prefix
    ikuaiapi.WithRawMode(true),                 // return the full envelope
    ikuaiapi.WithDryRun(true),                  // preview requests, send nothing
    ikuaiapi.WithRetry(3),                      // total attempt count
    ikuaiapi.WithRetryDelay(200*time.Millisecond, 5*time.Second),
    ikuaiapi.WithMetrics(ikuaiapi.NewMetrics()),
    ikuaiapi.WithStructuredLogger(logx.NewStdLogger("poller", logx.LevelInfo)),
)
```

`WithDryRun` is the safe way to see what a call would do: it returns the
request as JSON and contacts nothing. Use it before any destructive
write.

## Generated API shapes

The generator emits one of these per catalog entry:

1. **Parameterless GET** — `Get<Name>(ctx)`
2. **Query helper** — `Query<Name>(ctx, *<Name>QueryOptions)` for endpoints
   whose dialect is not the default
3. **CRUD** — `List<Name>` / `Get<Name>` / `Create<Name>` / `Update<Name>` /
   `Patch<Name>` / `Delete<Name>`, where `Create` returns the `rowid`

Each carries its path and verbs in a doc comment above it:

```go
// ListNetworkDhcpServices lists items at /network/dhcp/services.
//
// Methods: GET, POST, PUT, PATCH
```

If a generated helper does not exist yet, `Call` is the escape hatch: it
resolves the (group, name) pair from the catalog and dispatches the request
with the method, body and query params you supply.

```go
raw, err := api.Call(ctx, "interfaces", "wan-config", "GET", nil, nil)
```

`Call` resolves static catalog paths, so it cannot address a sub-resource
— use the generated helpers for those.

## Development

```bash
make test        # go test ./...
make check       # fmt + vet + test + dependency gate
make race        # go test -race ./...
make generate    # regenerate service/ from the contract tables
make zero-dep    # only github.com/oekit/* may appear in the module graph
make example     # runnable demo, honours IKUAI_DRY_RUN=1
```

The `service/` package is generated. To change what the SDK exposes, edit
`v4_catalog.go` (paths, verbs, query dialect, sort behaviour) or
`v4_schemas.go` (body defaults, allowlists, required fields), then run
`make generate`. CI fails if regeneration produces a diff.

## Project layout

```
.
├── doc.go                 — package documentation
├── client.go              — HTTP, retry/back-off, envelope, sanitization
├── errors.go              — APIError (oekit BizError), NetworkError, sentinels
├── fields.go              — write-body conventions and read-modify-write
├── auth.go                — token shape validation
├── logger.go              — logx.Delegating package logger, level floor
├── metrics.go             — request counters and latency
├── version.go             — SDKVersion, Version enum
├── v4_catalog.go          — 157 endpoints + query dialect contracts
├── v4_schemas.go          — 73 endpoint body contracts
├── codegen/               — service-layer generator
├── service/               — generated, one file per group + root.go
├── session/               — credential persistence and resolution
├── output/                — table / JSON rendering
├── skills/                — agent-facing guides
├── example/               — runnable demo
├── AGENTS.md              — hard rules for humans and agents
└── llms.txt               — machine-readable entry point
```

## Verification status

Endpoints, query dialects and body contracts were checked against a live
iKuai OS 4.x router, and several entries exist because a probe proved the
obvious spelling wrong:

- **157 endpoints**; 29 carry a query-dialect contract, 73 a body contract.
- **147 of 155** GET-addressable endpoints answer 200. The other eight are
  five POST-only action endpoints, the two IKEv2 endpoints (the feature is
  absent on that firmware), and `/vpn/wireguard` (disabled there).
- **All 11 action endpoints** answer `{"message":"Success"}` under POST.
- **Write path verified end to end** on an object group: create, read back,
  update, read back, delete — the `group_value` encoding round-trips
  byte-for-byte, including non-ASCII comments.
- **Every allowlist was cross-checked** field by field against live
  responses. Fields the router returns and accepts on write were added;
  hardware readouts and computed counters stay out deliberately, and
  `v4_schemas.go` records which is which.

Two things could **not** be settled on that device: the legacy uppercase
dialect and the `/auth/online-users` mixed dialect, because
`downstream`, `cameras` and `switch` are empty there and both spellings
return an empty result. They are encoded from the upstream client and
flagged as unverified.

## Known limitations

- The legacy and mixed dialects are encoded from an independent
  implementation, not confirmed on hardware — see above.
- `/vpn/ikev2/services` 404s on the firmware tested; that is the feature
  being absent, not the path being wrong.
- There is no concurrency helper. `Client` is safe for concurrent use, and
  the generated methods take a context, but the SDK deliberately does not
  start goroutines on your behalf.
- `Metrics` is a small in-process counter, not a Prometheus integration.
  Feed the numbers into whatever your service already runs.

## License

MIT. See [LICENSE](LICENSE).
