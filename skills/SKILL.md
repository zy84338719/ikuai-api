---
name: ikuai-api
description: |
  Go SDK for the iKuai v4.0 REST API (routers running iKuai OS 4.x).
  Use when reading router status or traffic, managing DHCP/DNS/VLAN/NAT/PPPoE,
  configuring VPN/firewall/ACL/QoS/wireless, managing users and web-admin
  accounts, reading logs, or diagnosing why a query returned nothing.
---

# ikuai-api — iKuai v4.0 SDK

A zero-dependency Go SDK: 152 catalog endpoints across 12 groups, typed
query helpers per endpoint dialect, and helpers for the body conventions
the firmware insists on.

```go
client, err := ikuaiapi.NewClient("https://192.168.1.1",
    ikuaiapi.WithToken(token),
    ikuaiapi.WithDryRun(true), // preview every call, touch nothing
)
if err != nil {
    return err
}
defer func() { _ = client.Close(context.Background()) }()

api := service.NewAPIClient(client)
```

Credentials can come from a session file or the environment instead of
being hard-coded — see `session.ResolveClient`.

## Read this first: not every endpoint speaks the same query dialect

iKuai v4 is not uniform. Sending the wrong query shape makes the router
**silently ignore the entire query string**, so a call that looks like it
returned data may actually have returned everything, or nothing useful.

| Dialect | Endpoints | Query sent |
|---------|-----------|------------|
| paged | most collections | `page`, `limit`, `filter`, `order`, `order_by` |
| legacy | `/monitoring/downstream`, `/monitoring/cameras`, `/monitoring/switch` | `TYPE=data`, `LIMIT="<offset>,<count>"`, `OFFSET`, `ORDER`, `ORDER_BY` |
| mixed | `/auth/online-users` | `page`, `limit` + uppercase `KEYWORDS`, `FINDS`, `ORDER`, `ORDER_BY` |
| window | `/monitoring/protocols*`, `/monitoring/clients/protocols*`, `/monitoring/app-protocols/history-load`, `/log/url-visits` | `starttime`, `stoptime` (no underscore) |
| none | `/monitoring/ssid-clients`, `/monitoring/channel-clients`, `/monitoring/wireless-traffic`, `/monitoring/flow-shunting`, `/monitoring/clients-traffic-load` | only that endpoint's own selectors |
| load | `/monitoring/cpu`, `memory`, `disk`, `cputemp`, `terminals`, `connections`, `network` | `datetype`, `start_time`, `end_time`, `math` (with underscores) |

> **The page-size key is `limit`, not `page_size`.** `page_size` is silently
> ignored. This trips up most hand-written v4 clients.

For a non-paged endpoint use the generated `Query<Name>` helper — it
validates required parameters and builds the exact query string:

```go
// legacy dialect
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

For a plain paged collection use the generic `List<Name>` helper:

```go
raw, err := api.Network().ListNetworkDhcpServices(ctx,
    &service.NetworkDhcpServicesListOptions{
        Page: 1, PageSize: 50,
        Filter: "enabled==yes",
        Order: "desc", OrderBy: "id",
    })
```

`service.Endpoints()` walks the whole catalog at runtime; each group also
exposes `service.PathNetwork`, `service.PathMonitoring()` and so on, which
is useful for building a CLI or a discovery screen without hard-coding.

## Error handling

```go
raw, err := api.Monitoring().GetMonitoringSystem(ctx)
if err != nil {
    var apiErr *ikuaiapi.APIError
    if errors.As(err, &apiErr) {
        switch {
        case apiErr.IsAuthFailure():       // 3007, 1008
            // token expired — refresh it, retrying will not help
        case apiErr.IsPermissionFailure(): // 10014
            // authenticated but not allowed; needs a different account
        default:
            // 3001 and friends: the request is wrong
        }
        fmt.Println(apiErr.Message, "->", apiErr.Remedy())
    }
}
```

| Code | Meaning | Fix |
|------|---------|-----|
| 3001 | parameter error | re-check the body against the endpoint's field hints |
| 3007 | token expired/invalid | get a new token in the router web UI (System → Auth → API Token) |
| 1008 | session expired | same |
| 10014 | permission denied | use an account that has the feature granted |

The SDK retries transient failures (429, 5xx) with exponential back-off and
honours `Retry-After`. It never auto-retries a write on a network error,
because the request may have reached the router.

## Writing is where the firmware gets opinionated

See `skills/writing.md`. The short version:

- `enabled` is the **string** `"yes"` / `"no"`, never a JSON boolean —
  use `ikuaiapi.Enabled(true)`.
- A few fields must be JSON numbers; everything else stays a string so an
  SNMP community of `"161"` or a VLAN name of `"0010"` survives intact.
  Use `ikuaiapi.CoerceBody(body)`.
- Address/port/protocol selectors are `{"custom":[...],"object":[]}`, not
  a bare array — use `ikuaiapi.AddressValues("80,443")`.
- Object-group members live in a `group_value` array of single-key objects
  — use `ikuaiapi.ObjectGroup("ip", "192.168.1.10")`.
- Most updates must send the **whole** object. Use
  `ikuaiapi.UpdatePayload(ctx, client, path, id, updates, allowedFields)`
  for the read-modify-write round trip.

## Rendering

`output` turns a response into something readable or pipeable — a table
for a human, compact JSON for a script or agent, plus column selection and
timestamp humanization. See `skills/rendering.md`.

## Capability map

| Group | Covers |
|-------|--------|
| `system` | basic config, NTP, reboot schedules, remote access, VRRP, ALG, kernel params, CPU freq, disks, files, backup/restore, firmware upgrade, web-admin groups & accounts |
| `network` | DNS + DNS proxy, DHCP + DHCPv6, static leases, access control, VLAN, NAT/DNAT/DMZ, QoS, PPPoE, AC |
| `interfaces` | WAN, WAN-VLAN, LAN, physical NICs |
| `monitoring` | system, CPU/memory/disk/temp load, online & offline clients (v4/v6), traffic, protocol & app-protocol breakdowns, wireless stats, cameras, switches, flow shunting |
| `security` | ACL, MAC filter + mode, L7 app rules, URL black/keywords/redirect/replace, domain blacklist, peerconn, terminals, advanced, secondary route |
| `vpn` | PPTP, L2TP, OpenVPN, IKEv2 servers and clients; IPsec; WireGuard tunnels and peers |
| `routing` | static routes, stream shunting by domain / five-tuple / L7 / load-balance / up-down |
| `wireless` | blacklist, VLAN rules, AC services and AP config |
| `objects` | IP, IPv6, MAC, port, protocol, domain and time object groups, plus reverse references |
| `log` | system, ARP, auth, DHCP, PPPoE, web activity, DDNS, notice, wireless, URL visits |
| `auth` | online sessions (with kick), accounts, packages, web portal config |
| `advanced` | FTP, HTTP, Samba, SNMPD |

## Discovery

```go
for _, ep := range service.Endpoints() {
    if ep.Group != "network" {
        continue
    }
    dialect, required, optional, err := ep.QuerySpec()
    // ...
}
```

`ep.QuerySpec()` decodes the per-endpoint query contract, so a generic
tool can tell an endpoint's paging style without hard-coding a table.

## Safety

- `ikuaiapi.WithDryRun(true)` turns every call into a request preview — use
  it to show the operator what a write would do.
- `WithRawMode(true)` returns the full `code`/`data`/`results`/`rowid`
  envelope instead of the unwrapped payload.
- Destructive calls are ordinary methods; there is no confirmation
  callback in the SDK. Gate them in your own code.
