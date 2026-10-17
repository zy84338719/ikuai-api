# Writing to iKuai: body conventions

The v4 write surface is where a generic JSON client breaks most often. The
firmware validates field names, types and completeness strictly, and it
rejects a partial update on most rule-like endpoints rather than defaulting
the missing fields.

## enabled is a string

Every rule-like resource uses `yes` / `no`, not `true` / `false`.

```go
body := map[string]any{
    "tagname": "block_ssh",
    "action":  "drop",
    "enabled": ikuaiapi.Enabled(false), // "no"
}
```

`ikuaiapi.ParseEnabled` reads the value back (it also accepts the boolean
spellings that turn up in payloads read from the router), so a value can be
round-tripped.

## Only some fields are numbers

A handful of fields must arrive as JSON numbers. Everything else must stay
a string, because several iKuai fields accept numeric-looking strings and
coercing them corrupts the value — an SNMP community of `"161"`, a VLAN
name of `"0010"`, a host comment of `"42"`.

```go
body := map[string]any{
    "mtu":       "1500", // allowlisted → becomes 1500
    "prio":      "31",   // allowlisted → becomes 31
    "community": "161",  // not allowlisted → stays "161"
}
ikuaiapi.CoerceBody(body)
ikuaiapi.IsIntegerAPIField("mtu")          // true
fields := ikuaiapi.IntegerAPIFields()      // a copy of the allowlist
```

Call `CoerceBody` on the map you hand to Create/Update/Patch. It is
idempotent and leaves fields that are already numbers alone.

## Selectors are an object, not an array

Addresses, ports, protocols and domains are always
`{"custom": [...], "object": [...]}` — the `object` half is how the web UI
stores references to object groups, and omitting the key makes the router
reject the rule.

```go
body["dst_port"] = ikuaiapi.AddressValues("80,443") // one argument
body["dst_port"] = ikuaiapi.AddressValues("80", "443") // or several
```

`AddressValues` trims each entry, drops empties, and always emits both keys.
A selector that selects nothing becomes `{"custom":[],"object":[]}`, which
the web UI shows as "any".

## Object groups

A create/read/update/delete cycle on hardware confirms the writable surface
is exactly `group_name` plus `group_value`. A row also comes back with `id`
and a computed `ref_count`, and neither belongs in an update body.

Members live in `group_value`, one single-key object per value, named after
the group type: `ip`→`ip`, `ip6`/`ipv6`→`ipv6`, `mac`→`mac`, `port`→`port`,
`proto`→`proto`, `domain`→`domain`.

```go
vals, err := ikuaiapi.ObjectGroup("ip", "192.168.1.10,192.168.1.11")
// [{"ip":"192.168.1.10"},{"ip":"192.168.1.11"}]
body["group_value"] = vals
```

The time group holds schedules instead of values:

```go
body["group_value"] = []any{
    ikuaiapi.TimeObjectGroup("weekly", "1234567", "00:00", "23:59", ""),
}
ikuaiapi.WeekSchedule() // the same all-day schedule, as a convenience
```

Combined with `CreateBody`, a whole create reads:

```go
vals, _ := ikuaiapi.ObjectGroup("ip", "10.0.0.1,10.0.0.2")
body, _ := ikuaiapi.CreateBody("/ip-objects", map[string]any{
    "group_name":  "servers",
    "group_value": vals,
})
```

## Updates need the whole object

Most endpoints reject a partial update: the firmware revalidates every
field and applies its own defaults to anything absent, which quietly
reconfigures unrelated settings. Fetch the current object, merge, and send
the result back.

```go
body, err := ikuaiapi.UpdatePayload(ctx, client, "/network/nat/rules", id,
    map[string]any{"nat_addr": "203.0.113.5"}, nil)
if err != nil {
    return err
}
_, err = api.Network().UpdateNetworkNatRules(ctx, id, body)
```

`UpdatePayload` reads the current item (path-segment form, falling back to
`?id=`), merges your updates over it, drops `id` (the firmware rejects it on
update), applies the endpoint's update allowlist, and runs the result
through `CoerceBody`.

The allowlist comes from `V4Schemas` when you pass `nil`. Pass an explicit
list to override it, or when you need to send a field the table does not
know about.

## Create bodies come pre-filled

`CreateBody` applies the defaults the firmware expects for an endpoint and
validates the required fields, so you never have to reconstruct a rule body
from memory:

```go
body, err := ikuaiapi.CreateBody("/security/acl-rules", map[string]any{
    "tagname":  "block_ssh",
    "action":   "drop",
    "protocol": "tcp",
    "dst_port": ikuaiapi.AddressValues("22"),
})
// enabled=yes, prio=50, dir=forward, every untouched selector is the empty
// pair, time is the all-day weekly window — and a missing required field is
// reported here rather than coming back as a 3001.
```

Inspect the contract directly when you need to:

```go
s, ok := ikuaiapi.Schema("/security/acl-rules")
s.CreateDefaults   // what a create gets filled in with
s.InputFields       // what an update may send
s.RequiredFields    // what a create must set explicitly
s.AllowsField("action")

ikuaiapi.SchemaPaths() // every endpoint carrying a contract
```

`CreateBody` returns a deep copy, so mutating the result — or building a
second body — cannot corrupt the shared table.

### Where the firmware is inconsistent

The table reproduces the firmware's own quirks rather than smoothing them
over, because normalizing them breaks the write:

| Endpoint | Quirk |
|----------|-------|
| `/auth/users` | `src_addr` holds a bare map with no `object` key |
| `/routing/load-balance-rules` | `weight` is the string `"1"`, not a number |
| `/vpn/openvpn/clients` | mixes types — `remote_port`, `tun_mtu`, `comp_lzo` are strings while `accept_push_route` is a number |
| `/wireless/access-control/rules` | `time` is a flat `"00:00-23:59"` plus a `week` string, not the nested weekly structure |
| `/security/url-black/rules` | the only rule whose `time` is an empty selector pair instead of the all-day window |
| `/routing/updown` | has no `time` field at all |

### Paths that were wrong, and how we know

Several paths in this SDK were originally wrong and were corrected against
a live router. If an endpoint 404s, suspect the path rather than the
feature — this firmware answers an unknown path with an HTML body:

| endpoint | shape |
|----------|-------|
| VPN server configuration | `/vpn/<proto>/services` — the bare `/vpn/<proto>` 404s |
| IPv6 object group | `/ip6-objects`, not `/ipv6-objects` |
| protocol object group | `/proto-objects`, not `/protocol-objects` |
| wireless controller | `/network/ac/services`, `/network/ac/ap-config` |


Lower-level building blocks, if you need to do the merge yourself:

```go
current, err := ikuaiapi.FirstObject(raw)          // unwrap data/results
merged := ikuaiapi.MergeBody(current, updates)     // updates win
```

`FirstObject` copes with a bare array, a bare object or a full envelope,
so it works whether or not the client is in raw mode.

## Resource-specific notes worth knowing

- **DHCP services** carry `opt_type*` fields whose defaults differ between
  create and update. Send them explicitly rather than relying on defaults.
- **VLAN and DHCP static** mirror `vlan_name` into `tagname`. Set both.
- **Routing rules** carry read-only helper keys suffixed `_int`; strip them
  before an update or the firmware rejects the body.
- **WireGuard** is two-level: a tunnel, then peers under
  `/vpn/wireguard/{tunnel_id}/peers/{peer_id}`.
- **Action endpoints** are path-suffixed, not bodies:
  `/network/dhcp/services:start`, `/system/vrrp:stop`,
  `/system/backup:restore`, `/system/upgrade:check`.
- **Backup restore and firmware upgrade** are irreversible. Gate them
  behind explicit confirmation in your own code, and preview with
  `WithDryRun(true)` first.

## Method shapes

| Shape | Example |
|-------|---------|
| List | `api.Network().ListNetworkDhcpServices(ctx, opts)` |
| Get one | `api.Network().GetNetworkDhcpServices(ctx, id)` |
| Create | `api.Network().CreateNetworkDhcpServices(ctx, body)` |
| Update | `api.Network().UpdateNetworkDhcpServices(ctx, id, body)` |
| Toggle | `api.Network().PatchNetworkDhcpServices(ctx, id, body)` |
| Delete | `api.Network().DeleteNetworkDhcpServices(ctx, id)` |
| Action | `api.System().StartSystemVrrp(ctx)` |
| Load window | `api.Monitoring().LoadCpu(ctx, &service.MonitoringCpuLoadOptions{...})` |
| Dialect query | `api.Monitoring().QueryMonitoringDownstream(ctx, opts)` |

Method names follow the catalog: group prefix, then the endpoint name in
CamelCase. `service.Endpoints()` and the per-group `Path<Group>` tables let
you confirm the exact name at runtime instead of guessing.
