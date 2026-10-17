# Monitoring: the endpoints that need a query helper

Monitoring is where the v4 API is least uniform. Most of these endpoints
ignore a generic `page`/`limit` query, so they have a generated
`Query<Name>` helper that builds the right shape and rejects a call that is
missing something the router requires.

## Point-in-time reads — no query at all

```go
api.Monitoring().GetMonitoringSystem(ctx)            // /monitoring/system
api.Monitoring().GetMonitoringInterfacesStatus(ctx)  // WAN interface state
api.Monitoring().GetMonitoringInterfacesTraffic(ctx)
api.Monitoring().GetMonitoringInterfacesPhysical(ctx)
api.Monitoring().GetMonitoringWirelessStatistics(ctx)
api.Monitoring().GetMonitoringWirelessScore(ctx)
```

These take no options and send no query string.

## Load windows — `datetype` / `start_time` / `end_time` / `math`

CPU, memory, disk, temperature, terminal count, connection count and
network load all answer over a time bucket. The names here use
**underscores**, unlike the `starttime` window below.

```go
raw, err := api.Monitoring().LoadCpu(ctx, &service.MonitoringCpuLoadOptions{
    DataType:  "hour", // hour | day | week | month
    StartTime: 1773300000,
    EndTime:   1773303600,
    Math:      "avg", // avg | max
})
```

The generated helper validates the enum values and rejects
`StartTime >= EndTime` before the request leaves the process.

## Client lists — paged, plus fuzzy search

```go
raw, err := api.Monitoring().QueryMonitoringClientsOnline(ctx,
    &service.MonitoringClientsOnlineQueryOptions{
        Page: 1, PageSize: 100,
        Filter:  "interface==lan1",
        Key:     "ip_addr,mac", // which fields to search
        Pattern: "192.168.1",   // what to search for
    })
```

`key` and `pattern` are the firmware's fuzzy-match pair and only exist on
the client-list endpoints. `clients-offline`, `clients-ip6-online` and
`clients-ip6-offline` take the same shape.

## Per-client drill-downs — `ip` and `mac` required

```go
// live 5-minute traffic for one client
api.Monitoring().QueryMonitoringClientsTrafficLoad(ctx,
    &service.MonitoringClientsTrafficLoadQueryOptions{
        IP: "192.168.1.100", MAC: "08:9b:4b:01:7e:7c",
    })

// protocol breakdown, optionally bounded by a time window
api.Monitoring().QueryMonitoringClientsProtocols(ctx,
    &service.MonitoringClientsProtocolsQueryOptions{
        IP: "192.168.1.100", MAC: "08:9b:4b:01:7e:7c",
        StartTime: 1773300000, StopTime: 1773303600,
    })

// app-protocol rate for one client
api.Monitoring().QueryMonitoringClientsAppProtocolsLoad(ctx,
    &service.MonitoringClientsAppProtocolsLoadQueryOptions{
        Page: 1, PageSize: 20,
        IP: "192.168.1.100", MAC: "08:9b:4b:01:7e:7c",
    })
```

Omitting `ip` or `mac` fails fast with a clear error rather than returning
a confusing empty result.

## Protocol and application views

```go
// distribution over a time window
api.Monitoring().QueryMonitoringProtocols(ctx,
    &service.MonitoringProtocolsQueryOptions{
        StartTime: 1773300000, StopTime: 1773303600,
    })

// history for one application set; appids is required
api.Monitoring().QueryMonitoringAppProtocolsHistoryLoad(ctx,
    &service.MonitoringAppProtocolsHistoryLoadQueryOptions{
        AppIDs:    "2580003,2580004",
        StartTime: 1773300000, StopTime: 1773303600,
    })

// which terminals are using one application
api.Monitoring().QueryMonitoringAppProtocolsTerminalLoad(ctx,
    &service.MonitoringAppProtocolsTerminalLoadQueryOptions{
        AppID: 2580003,
    })

// paged application load
api.Monitoring().QueryMonitoringAppProtocolsLoad(ctx,
    &service.MonitoringAppProtocolsLoadQueryOptions{
        Page: 1, PageSize: 20, Order: "desc", OrderBy: "total_down",
    })
```

Note the two spellings: `AppIDs` (plural, comma-separated) for history,
`AppID` (single number) for terminals.

## Traffic summaries

```go
// generic List is correct here — these two are plain paged endpoints
api.Monitoring().ListMonitoringClientsTrafficSummary(ctx,
    &service.MonitoringClientsTrafficSummaryListOptions{Page: 1, PageSize: 20})
api.Monitoring().ListMonitoringAppTrafficSummary(ctx,
    &service.MonitoringAppTrafficSummaryListOptions{Page: 1, PageSize: 20})
```

## Legacy uppercase endpoints

`/monitoring/downstream`, `/monitoring/cameras` and `/monitoring/switch`
use the old dialect: no `page`, a `LIMIT="<offset>,<count>"` pair, and
uppercase sort keys. The helper computes the offset for you.

```go
api.Monitoring().QueryMonitoringCameras(ctx,
    &service.MonitoringCamerasQueryOptions{
        Page: 2, PageSize: 20,
        Order: "desc", OrderBy: "id",
        Status: "1", Keywords: "hikvision",
    })

api.Monitoring().QueryMonitoringDownstream(ctx,
    &service.MonitoringDownstreamQueryOptions{
        Page: 1, PageSize: 20,
        Status: "1", Device: "camera",
    })
```

`device` accepts the router's own vocabulary: `router`, `switches`,
`firewall`, `server`, `camera`, `printer`, `SmartDevices`, `ap`, `other`.
`status` is `"0"` or `"1"`.

## Wireless and flow views

```go
api.Monitoring().QueryMonitoringWirelessTraffic(ctx,
    &service.MonitoringWirelessTrafficQueryOptions{APMAC: "aa:bb:cc:dd:ee:ff"})
api.Monitoring().QueryMonitoringSsidClients(ctx,
    &service.MonitoringSsidClientsQueryOptions{SSID: "Office"})
api.Monitoring().QueryMonitoringChannelClients(ctx,
    &service.MonitoringChannelClientsQueryOptions{Channel: 36})
api.Monitoring().QueryMonitoringFlowShunting(ctx,
    &service.MonitoringFlowShuntingQueryOptions{Type: "data"})
```

## Logs

`/log/url-visits` is a normal paged endpoint with a time window:

```go
api.Log().QueryLogUrlVisits(ctx,
    &service.LogUrlVisitsQueryOptions{
        Page: 1, PageSize: 20,
        Pattern: "example.com",
        StartTime: 1761842000, StopTime: 1761843000,
    })
```

The other log types (`system`, `arp`, `auth`, `dhcp`, `pppoe`, `web`,
`ddns`, `notice`, `wireless`) use the generic List helper, and every log
endpoint also supports Delete to clear itself.

## Response envelopes

Three shapes come back from this firmware, and they are easy to confuse:

| shape | example | seen on |
|-------|---------|---------|
| `data` | `{"data":[...],"total":n}` | most collections |
| `results` | `{"results":[...]}` | monitoring load endpoints |
| group paging | `{"ip_data":[],"ip_total":0}` | object groups, DHCP static |

The third one is the trap. The object groups do not return an array — they
return the members under `<group>_data` with a separate counter, so
`/ip-objects` answers with `ip_data` and `/ip6-objects` with `ipv6_data`.

`FirstObject` unwraps all three, which is what a read-modify-write update
goes through. To read a whole group yourself, decode the member array and
the counter separately:

```go
var page struct {
    IPData  []map[string]any `json:"ip_data"`
    IPTotal int              `json:"ip_total"`
}
json.Unmarshal(raw, &page)
```

An empty group answers `{"ip_data":[],"ip_total":0}`, so
`FirstObject` correctly reports "no object" rather than handing back the
wrapper with its counters as if they were a row.

## Interpreting the response

The client unwraps the envelope: `data` first, then `results` (used by
monitoring load endpoints), and it normalizes the bare `nil` some firmware
emits in place of `null`. Turn `timestamp` fields into something readable
with the output package rather than by hand:

```go
output.New(os.Stdout, output.FormatTable).WithOptions(output.RenderOptions{
    HumanTime: true,
    Columns:   []string{"ip_addr", "mac", "upload", "download"},
}).Value(v)
```
