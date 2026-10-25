# Changelog

## v1.3.1 — 2026-10-26

Dependency refresh plus one behavioural fix found while aligning with the
current oekit releases.

### Changed

- **Dependencies**: `oekit/logx` v0.7.0 → v0.7.1, `oekit/pkg` v0.13.0 →
  v0.28.0. No API change on any path this SDK uses.

- **Retry-After hints are now clamped** by `oekit/pkg/httpjson.SelectRetryDelay`
  instead of being trusted unconditionally. A hint is adopted only when it is
  ≤ MaxServerRetryAfter (5 minutes) or shorter than the local exponential
  curve; `APIError.RetryAfter` still carries the raw hint for callers that
  want to make their own call. Previously a pathological header such as
  `Retry-After: 86400` would park a retrying client for a day. A regression
  test locks both directions of the clamp.

## v1.3.0 — 2026-10-18

Adopts the oekit v0.13 / v0.7 releases. One change is visible to callers;
the rest brings this SDK into line with a rule it was previously working
around.

### Fixed

- **Retry back-off could fire before its own baseline.** The delay was
  fully jittered into [0, d], so a retry scheduled for a 4-second back-off
  could fire three milliseconds later. Under a failing router — the one
  moment back-off exists for — every client doing that at once turns into a
  retry storm exactly when the device can least absorb it. The jitter is now
  one-sided via `mathx.Unstable.UpwardDuration`, giving [d, d*(1+0.25)]:
  never earlier than the exponential baseline, still spread enough that
  clients do not resynchronise. `TestBackoffNeverShorterThanBaseline` locks
  the lower bound; `TestBackoffSpreadsClients` locks the upper purpose.

### Changed

- **Dependencies**: `oekit/pkg` v0.11.0 → v0.13.0, `oekit/logx` v0.6.0 →
  v0.7.0. No breaking API change affects this SDK; logx's one breaking
  change (`ParseLogLevel` now returns an error) is on no path used here.

- **One errors package in the process.** `oekit/pkg/errors` v0.12.0 added
  `New`/`As`/`Is`/`Unwrap` with the standard library's shapes, which removes
  the reason this SDK still imported the standard library alongside it. The
  sentinels and every match now go through the oekit package, so the "same
  process, one errors package" rule holds without a caveat. The `errors.Is`
  and `errors.As` contracts callers already rely on are unchanged.

### Not adopted

- `pkg/jsonrepair` was evaluated and rejected. It repairs LLM output —
  stripping markdown fences and extracting an object out of prose — which is
  the wrong behaviour for a router response. The firmware emits well-formed
  JSON apart from the bare `nil` tokens `SanitizeNil` already handles;
  letting a repairer normalise the payload would obscure a malformed
  response rather than surface it.

### Documentation

- `doc.go` documents redacting SDK logs by key through `logx.WithFilter`,
  for callers whose URLs are themselves sensitive.

## v1.2.0 — 2026-10-18

Two independent bodies of work landed here.

**Query and body-contract correctness**, ported from a field-verified
survey of the iKuai v4 surface. The headline fix is that paging silently
did not work.

**Adoption of the oekit engineering contract**, so that errors and logs
from this SDK are understood by an oekit service with no adapter in
between. That part is breaking; see Compatibility.

### Breaking changes

- **`Client.Close()` → `Client.Close(ctx context.Context) error`.**
  Follows the oekit lifecycle rule: a context bounds the shutdown, the
  error is returned instead of swallowed, and repeated or concurrent calls
  return the first result without redoing the work.

  ```go
  // before
  defer client.Close()
  // after
  defer func() { _ = client.Close(context.Background()) }()
  ```

- **`WithLogger(fn func(format string, args ...any))` is removed.** The
  printf-style callback has no place in the kv logging contract. Logging
  now goes through the oekit `logx` facade; call `logx.UseSlog("svc")` (or
  `logx.SetDefault`) once in `main` and the SDK follows.

- **`NewDefaultLogger(LogLevel) Logger` and the `Logger` interface are
  removed** from this package, along with `Client.Logger`. Use
  `logx.NewStdLogger` / `logx.UseSlog` and, for per-client narrowing,
  `WithStructuredLogger(logx.Logger)`.

- **`output.Options` → `output.RenderOptions`.** A bare `Options` type
  collides when two packages are co-imported; every options type in this
  repo is now package-qualified.

- **No exported identifier was renamed except the three above.**
  `APIError`, `NetworkError`, `IsAuthFailure`, `IsPermissionFailure`,
  `Remedy`, `CoerceBody`, `CreateBody`, `UpdatePayload` and the service
  layer are unchanged, so most callers only need to touch `Close`.

### Fixed

- **Pagination sent the wrong key.** Every generated `List<Name>` helper
  sent `page_size=<n>`. The v4 request contract names that parameter
  `limit`; `page_size` is ignored by the router, so paging was a no-op and
  callers silently received an unpaged result. `codegen` now emits
  `limit`, and `service/codegen_test.go` asserts that `page_size` never
  appears. The Go-side `PageSize` field is unchanged, so this is a
  behaviour fix with no source-level migration for callers.

- **`PageInfo` could not read the page size back.** Firmware echoes either
  `page_size` or `limit` depending on the revision; `PageInfo` only had the
  former. It now decodes both and `PageInfo.Size()` prefers `page_size`,
  falling back to `limit`.

- **Sorting was reversed on the security and advanced-service
  collections.** Almost everywhere the contract is `order=<direction>` and
  `order_by=<field>`. Those two groups want the opposite pairing, so the
  SDK was sorting by a column literally named `asc`/`desc` — on 13
  endpoints, with no error raised. `V4Endpoint.SortFlipped` records the
  exception and `codegen` translates it. `Order` still means the direction
  and `OrderBy` still means the field on every endpoint, so no caller
  change is needed.

### Added

- **oekit error contract.** Every `APIError` now embeds an oekit
  `BizError`, so `oekerrors.AsBizError` — and therefore `mw.Fail` — sees a
  business code and an HTTP status instead of falling through to a
  generic 500. The iKuai code becomes the business code; `StatusFor`
  derives the status (3001 → 400, 3007/1008 → 401, 10014 → 403), and a
  status the router actually sent always wins. SDK-local validation
  failures use the 4xxx business-code band, so "the SDK refused to send
  this" stays distinguishable from "the router said no".
  `APIError.Unwrap()` exposes the embedded `BizError` to keep the chain
  intact; `AsBizError` is re-exported for callers that do not want to know
  which layer produced the error.

- **Package-level sentinel errors** (`ErrInvalidToken`, `ErrNotConfigured`,
  `ErrRequiredField`, `ErrUnknownObjectGroup`, `ErrEmptyResponse`,
  `ErrNoObject`, `ErrNoSchema`) so local failures are matchable with
  `errors.Is` instead of only by message text.

- **Structured logging through `logx`.** The package logger is
  `logx.Delegating("ikuai-api")`, resolved on every call, so a consumer's
  `logx.SetDefault` in `main` is picked up without this package choosing a
  backend or hard-coding a service name. Events are kv-style and carry the
  request context, so a trace id attaches automatically. `SetLogLevel`
  controls the floor for the SDK's own events.

- **Dependency gate.** `make zero-dep` and a CI job enforce that only
  `github.com/oekit/*` may appear in the module graph. `go mod tidy` keeps
  it at exactly two direct requires (`oekit/logx`, `oekit/pkg`) — logx
  pulls hertz only through its `hlog` backend, which this SDK does not
  import.

- **Per-endpoint body contracts (`V4Schemas`).** 73 endpoints now carry
  their create defaults, update allowlist and required fields (36 with an
  allowlist, 39 with required fields). `CreateBody` fills a create payload
  in (returning a deep copy, so no caller shares structure with the table)
  and reports a missing required field locally instead of surfacing it
  later as a 3001. `Schema`, `SchemaPaths` and `V4Schema.AllowsField`
  expose the contract for inspection.

  The firmware's own inconsistencies are reproduced rather than
  normalized, because normalizing them breaks the write — see
  `skills/writing.md` for the list.

- **`UpdatePayload` consults the schema.** Passing `nil` for the allowlist
  now applies the endpoint's recorded allowlist instead of keeping
  everything. An explicit `keep` list still wins. This is a behaviour
  change for callers that previously passed `nil`.

- **Per-endpoint query contracts.** `V4Endpoint.Query` carries a compact
  spec (`ParseQuerySpec`) describing the dialect an endpoint actually
  speaks plus its extra required/optional parameters. `codegen` turns each
  spec into a typed `Query<Name>(ctx, *<Name>QueryOptions)` helper that
  builds the exact query string and rejects incomplete requests before they
  reach the network. Six dialects are modelled: `paged`, `legacy`
  (`TYPE`/`LIMIT="<offset>,<count>"`/`OFFSET`/`ORDER`/`ORDER_BY`), `mixed`
  (lowercase paging with uppercase `KEYWORDS`/`FINDS`), `window`
  (`starttime`/`stoptime`), `none`, and `load`
  (`datetype`/`start_time`/`end_time`/`math`). 20 endpoints carry a spec.

- **Missing endpoint.** `/log/url-visits` (GET, DELETE) was absent from the
  catalog; it is now present, with its `pattern`/`starttime`/`stoptime`
  contract recorded.

- **`fields.go` — the write-body conventions, encoded.** `Enabled` /
  `ParseEnabled` (`"yes"`/`"no"`, never a JSON boolean), `CoerceBody` and
  the `IntegerAPIFields` allowlist (only documented fields become numbers,
  so an SNMP community of `"161"` or a VLAN name of `"0010"` survive),
  `AddressValues` (the `{"custom":[...],"object":[...]}` selector pair),
  `ObjectGroup` / `TimeObjectGroup` / `WeekSchedule` (`group_value`
  members), `FirstObject`, `MergeBody`, and `UpdatePayload`.

- **`session` package.** Credential persistence with atomic `0600` writes,
  environment fallback (`IKUAI_API_BASE_URL` / `IKUAI_API_TOKEN`, also
  honouring the `IKUAI_CLI_*` spellings), explicit source reporting, and
  `NormalizeBaseURL`.

- **`output` package.** Table and JSON rendering with column selection,
  timestamp humanization, whole-column dropping for narrow terminals, and
  flattening for the firmware's selector and object-group structures.

- **Actionable error codes.** `APIError` gained `Remedy()`,
  `IsAuthFailure()` (3007, 1008), `IsPermissionFailure()` (10014) and
  `IsBadRequest()` (3001). Code 10014 was previously unmapped.

- **Agent-facing docs.** `AGENTS.md` (the repo's hard rules), `llms.txt`,
  and `skills/` — `SKILL.md`, `monitoring.md`, `writing.md`, `rendering.md`.
  Package documentation moved to `doc.go`.

### Sub-resource routing

The catalog stores paths as strings, which cannot express a resource that
lives under a parent — WireGuard peers belong to a tunnel. The generator
now understands `{placeholder}` segments:

- `{tunnel_id}` names a parent and is bound to a leading `int64` argument.
- A trailing placeholder (e.g. `{peer_id}`) is bound to the `id` argument
  that Get/Update/Patch/Delete already take, and those verbs no longer send
  `?id=` as well.

```go
api.Vpn().ListVpnWireguardPeers(ctx, opts)
api.Vpn().GetVpnWireguardPeer(ctx, tunnelID, peerID)
api.Vpn().CreateVpnWireguardPeer(ctx, tunnelID, body)
api.Vpn().DeleteVpnWireguardPeer(ctx, tunnelID, peerID)
```

An endpoint whose path ends in a placeholder gets no `List` helper — there
is no collection at that path. Endpoints without placeholders generate
exactly what they did before, which `TestFlatEndpointsAreUnaffected` locks.

- **Object groups and DHCP contracts.** `V4Schema` gained three fields
  the earlier shape could not express:

  - `UpdateDefaults` — some endpoints genuinely differ per verb. A DHCP
    service create writes no `opt_type*` fields at all, while an update
    must send twelve of them (eleven defaulting to 0, `opt_type121` to 2).
    Sending one verb's body to the other reconfigures options nobody
    touched.
  - `Mirror` — the VLAN endpoint treats its rule name and display name as
    one field and copies one onto the other unconditionally, so
    `UpdatePayload` does too.
  - `DropBlanks` — an option value that came back from the router as an
    empty string is dropped rather than re-sent, because the firmware reads
    `""` as "no option configured", not "unchanged". A blank the caller
    supplies is left alone.

  `UpdatePayload` applies them in the firmware's own order: read-back →
  drop blanks → caller updates → create defaults → update defaults →
  allowlist → mirrors.

- **Object-group reverse lookups.** Every object group answers "which rules
  reference this group" on a `/{group}/ref` sub-path with a `group_name`
  query parameter. That is 7 endpoints and a new `DialectRef`, which takes
  one required name and no paging.

- **Read-back normalization on update.** Three more firmware behaviours are
  now modelled, all of which silently corrupt an update rather than
  failing loudly:

  - `DropSuffix` — the routing endpoints return a derived `_int` companion
    for several fields and reject a body carrying one back.
  - `NormalizeSelectors` — an empty selector pair read back as
    `{"custom":{},"object":{}}` is rewritten to the array form the firmware
    wrote, because it rejects the object form on the next write.
  - `id` is now dropped unconditionally, not only when the endpoint
    declares an allowlist. The firmware rejects it on any update.

  `UpdatePayload` decodes with `json.Number` rather than `float64`. A row
  id or counter that round-trips through `float64` comes back altered, and
  an update sends whatever came back — so the corruption used to land on
  the router. `TestReadModifyWriteKeepsLargeIntegers` locks it.

- **More endpoint contracts.** Static routes, DNS configuration, the DHCP
  access-control rule sets and their modes, the security singleton configs
  (MAC mode, advanced, secondary route) and CPU frequency now carry
  contracts. That brings the table to 73 endpoints.

### Verified against hardware

Every entry below was probed against a live iKuai OS 4.x router, and three
of the results contradicted what the SDK previously shipped. A 404 from
this firmware comes back as an HTML body, not the usual JSON envelope, so
"unknown path" and "feature absent" are distinguishable.

**Fixed — the VPN server path was wrong.** `/vpn/pptp`, `/vpn/l2tp` and
`/vpn/openvpn` all 404. The configuration lives at `/vpn/<proto>/services`,
which returns it. Server-side VPN configuration was therefore unreachable
through this SDK. `/vpn/ikev2/services` also 404s on this firmware, but
that is the feature being absent, not the path being wrong — it now follows
the same shape as the others.

**Fixed — two object groups were misnamed.** `/ipv6-objects` and
`/protocol-objects` 404; `/ip6-objects` and `/proto-objects` answer. The
other five object groups were already correct. The `ref` sub-paths moved
with them.

**Removed — four AC entries that never existed.** They sat at the API root
(`/services`, `/clients`, `/peers`, `/ref`) and all 404. The controller's
endpoints are `/network/ac/services` and `/network/ac/ap-config`, which the
network group already carried correctly, so the four root-level entries are
gone rather than relocated on a guess.

**Confirmed — pagination.** The `page_size` → `limit` correction is real.
Against a populated online-client collection:

| query | rows returned |
|-------|---------------|
| `limit=1` | 1 |
| `limit=3` | 3 |
| `limit=5` | 5 |
| `page_size=1` | every row |
| `page_size=3` | every row |
| `page_size=5` | every row |

Page 2 returns a different slice, while the same page requested with
`page_size` repeats page 1 exactly. The old spelling was ignored outright.

**Corrected — the sort pair was never reversed.** This release had carried
a `SortFlipped` marker asserting that the security and advanced-service
collections pair their sort keys the other way round, on the strength of an
independent CLI's flag mapping. Hardware says otherwise. On a security collection with several rows:

| query | result |
|-------|--------|
| `order=desc&order_by=id` | highest id first — descending |
| `order=asc&order_by=id` | lowest id first — ascending |
| `order=id&order_by=desc` | no rows |
| `order=id&order_by=asc` | no rows |

The standard pairing works on security collections exactly as it does
everywhere else, and the reversed pairing returns nothing at all. Translating the
keys would have silently emptied every security and advanced-service list.
The marker is removed; every collection now uses `order=<direction>` and
`order_by=<field>`.

**Fixed — token validation rejected every real token.** `ValidateToken`
required 32 hex characters. A current 4.x router issues a base64 token
(32 characters, decoding to a UUID prefix), so the check rejected tokens
that authenticate fine. Verified: the base64 token passed every endpoint
while failing validation. Both shapes are accepted now, and the check is
explicitly a shape check rather than a pre-authentication.

**Fixed — eleven writable fields were being dropped by the update
allowlists.** Every allowlist was cross-checked against live responses.
The router returned fields the allowlist omitted, so a read-modify-write
would clear them silently:

| endpoint | field that was dropped |
|----------|------------------------|
| `/system/remote-access` | `open_ftp`, `ftp_port`, `ftp_access` |
| `/security/advanced/config` | `limit_tcp2p`, `limit_tcp2p_num`, `limit_udp2p`, `limit_udp2p_num` |
| `/security/secondary-route/config` | `time` |
| `/vpn/openvpn/services` | `auth` |
| `/network/dhcp/static` | `termname` |
| `/system/backup-auto` | `tagname` |
| `/advanced-service/samba-config` | `interface` |

All seven allowlists now carry them. Thirteen fields were confirmed
read-only and deliberately stay out — hardware readouts under
`/system/cpufreq`, and `available`/`status` on the DHCP service list —
because the router computes them and rejects them on write.

**Fixed — a third response envelope shape.** Alongside `data` and
`results`, several endpoints answer `{"<group>_data": [...],
"<group>_total": n}`: the object groups (`/ip-objects` → `ip_data`) and the
DHCP static list (`static_data`). `FirstObject` did not recognise it, so a
read-modify-write on those endpoints would have handed back the wrapper —
counters included — as if it were the row.

**Confirmed — action endpoints.** All eleven `:start` / `:stop` /
`:restart` / `:sync` / `:check` / `:start` / `:restore` endpoints answer
`{"message":"Success"}` under POST. They 404 under GET, so a GET-based
sweep reports them as missing; they are correct.

**Full sweep.** 147 of 155 GET-addressable endpoints answer 200. The
remaining eight are the five action endpoints above (POST-only), the two
IKEv2 endpoints, and `/vpn/wireguard`, which answers 200 with
"Access denied" — the feature is disabled on this firmware, and the path
is right.

**Fixed — SDK-local validation errors answered 500.** An error the SDK
raises before anything reaches the router carries neither an iKuai code nor
a router status, so the status fallback assumed the server had failed and
produced HTTP 500. Missing options, a page number of zero, an endpoint
absent from the catalog — all client mistakes, all reported as server
errors. They now answer 400. This matters most in an oekit service, where
`mw.Fail` turns that number straight into the response: a bad call would
have blamed the server and tripped alerting for a request never sent.

**Verified — the write path, end to end.** The read path had been checked
against hardware; the write path had not. A reversible cycle on an empty
object group — create, read back, update, read back, delete — exercises
every layer the SDK adds between a caller's map and the wire:

```
CreateBody → {"group_name":…,"group_value":[{"ip":"10.255.255.10"},…]}
POST       → {"message":"Success","rowid":1}
read back  → byte-for-byte what was sent; the group_value encoding is accepted
PUT        → {"message":"Success"}   (id dropped, as documented)
DELETE     → {"message":"Success"}   → group empty again
```

The same cycle on `/time-objects` confirmed the schedule encoding, Chinese
comments included. A row comes back as exactly
`{group_name, group_value, id, ref_count}`, so `ref_count` is a computed
counter: the seven object-group allowlists, previously empty, now name
exactly `group_name` and `group_value`.

**Not verifiable on this device.** The legacy uppercase dialect and the
`/auth/online-users` mixed dialect could not be settled: `downstream`,
`cameras` and `switch` are empty here, so both spellings return an empty
result and nothing can be distinguished. They stay encoded as documented.

### Changed

- The endpoint count is now 152 (was 151) with `/log/url-visits` added.
- `codegen` reuses the parent package's `ParseQuerySpec` instead of
  re-parsing the spec grammar, so the catalog syntax has one
  implementation.
- Generated `Query<Name>` options no longer repeat a parameter their
  dialect already covers (e.g. `limit` on a paged endpoint).

### Compatibility

- Wire-format changes: list queries now send `limit` rather than
  `page_size`, and security / advanced-service list queries send the
  reversed sort pairing. Both are the point of this release.
- `V4Endpoint.Load` is deprecated in favour of `Query: "load"` but is still
  honoured and decodes as `DialectLoad`.
- The SDK is no longer standard-library-only. It depends on two oekit
  modules, and no third-party require is permitted; `make zero-dep`
  enforces that.

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
