# v1.0.0 Release Notes

## Headline

First **v4-only stable release**. The v3 protocol is gone, the SDK is
regenerated from a single source-of-truth catalog (151 endpoints
across 13 groups), and the heavy `imroc/req/v3` dependency has been
replaced with the Go standard library.

## Install

```bash
go get github.com/zy84338719/ikuai-api@v1.0.0
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
    client, _ := ikuaiapi.NewClient("https://192.168.1.1",
        ikuaiapi.WithToken("deadbeefcafebabe1234567890abcdef"),
        ikuaiapi.WithTimeout(15*time.Second),
    )
    defer client.Close()

    api := service.NewAPIClient(client)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    raw, err := api.Monitoring().GetMonitoringSystem(ctx)
    if err != nil {
        log.Fatal(err)
    }
    var overview map[string]any
    _ = json.Unmarshal(raw, &overview)
    fmt.Printf("hostname: %v\n", overview["hostname"])
}
```

## How to get a token

1. Sign in to the router as `admin`.
2. **System → Auth → API Token**.
3. Click **Generate**, copy the 32-character hex string.

## What's in the box

- **151 typed methods** across 13 functional groups, code-generated
  from `v4_catalog.go`. Re-generate with `go run ./codegen` after
  editing the catalog.
- **Zero external dependencies** — only the Go standard library.
- **Retry** with exponential back-off + jitter for transient failures.
- **Dry-run mode** returns a JSON preview of the request without
  contacting the router.
- **iKuai-specific quirks** handled: the firmware emits bare `nil`
  instead of `null`, envelope fields are normalised across
  `data` / `results` / `rowid`.

## Breaking changes from v0.2.x

- v3 (`/Action/call` + `func_name/action/param`) is gone.
- No more username/password login. Use a Bearer token.
- `*ikuaiapi.V3ActionClient`, `*ikuaiapi.V4Client`, `*SDKError`,
  `types.BaseRequest`, `types.ShowResponse*` and friends are
  removed. See `CHANGELOG.md` for the full migration cheatsheet.

## Service groups

| Group | Endpoints |
| --- | ---: |
| `advanced` (FTP / Samba / SNMP / HTTP) | 6 |
| `auth` (users, packages, web, online) | 4 |
| `interfaces` (LAN / WAN / physical / VLAN) | 4 |
| `log` (arp, auth, dhcp, ddns, notice, pppoe, …) | 9 |
| `monitoring` (system, cpu, memory, clients, traffic, …) | 37 |
| `network` (dhcp, dmz, dnat, dns, nat, pppoe, qos, vlan, ac) | 25 |
| `objects` (domain / ip / ipv6 / mac / port / protocol / time) | 7 |
| `routing` (static / policy / 5-tuple / load-balance) | 6 |
| `security` (acl / mac / url / domain / peerconn / terminals) | 13 |
| `system` (basic / alg / ntp / cpufreq / kernel / reboot / backup / upgrade / web-admin / ac) | 28 |
| `vpn` (pptp / l2tp / openvpn / ikev2 / ipsec / wireguard) | 10 |
| `wireless` (access-control, vlan) | 2 |

Access any group via `api.<Group>()` (CamelCased).

## Tests

- 22 unit tests covering `SanitizeNil` edge cases, envelope
  handling, retry, dry-run, network-error wrapping, pagination, and
  a round-trip across every catalog entry.

## License

MIT. See `LICENSE`.
