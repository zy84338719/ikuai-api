# Rendering responses

`output` turns a raw response into a table for a person or JSON for a
machine, without pulling in a dependency.

```go
r := output.New(os.Stdout, output.FormatAuto)

r.JSON(raw) // full response, closest to the wire format
r.Value(v)  // an already-decoded value
```

`FormatAuto` renders a table when the stream is a terminal and JSON
otherwise, so the same program is readable interactively and pipeable in
CI. Override the detection with `IKUAI_FORCE_TTY=1` or `=0`.

## Options

```go
r = r.WithOptions(output.RenderOptions{
    Format:    output.FormatTable,
    Columns:   []string{"id", "ip_addr", "mac"}, // restrict and order
    HumanTime: true,                              // timestamp → local time
    TermWidth: 100,                               // drop columns past this width
    Indent:    true,                              // pretty-print JSON
})
```

`WithOptions` only overwrites the fields you set, so it composes cleanly.

## What the renderer handles

The firmware's responses are not plain records, and the renderer flattens
the two structures that would otherwise dump raw JSON into a cell.

**Selector pairs** — every address, port and protocol field:

```json
{"custom": ["192.168.1.0/24", "10.0.0.5"], "object": []}
```

renders as one value per line, and an empty pair renders as `*`, matching
what the web UI shows for "any".

**Object-group members** — the `group_value` array:

```json
[{"ip": "192.168.1.10"}, {"ip": "192.168.1.11"}]
```

renders as `ip: 192.168.1.10` / `ip: 192.168.1.11`. Members with several
fields render as comma-separated `key: value` pairs.

**Single-key envelopes** — `{"data": [...]}`, `{"results": [...]}`, and
`{"objects": {...}}` are unwrapped, so you get the table rather than one
useless `DATA` column.

**Empty results** render as `(no data)` rather than an empty screen.

## Column naming

Headers are converted to `UPPER_SNAKE_CASE` so the output reads like the
router's own UI, including acronym boundaries:

| Field | Header |
|-------|--------|
| `tagname` | `TAGNAME` |
| `src_addr` | `SRC_ADDR` |
| `srcIPAddr` | `SRC_IP_ADDR` |
| `enableHTTPS` | `ENABLE_HTTPS` |

## Narrow terminals

With `TermWidth` set, the renderer keeps the leading columns that fit and
drops the rest **whole** — truncating a value mid-cell loses information
that the next column might clarify. Dropped columns are reported:

```
(2 columns hidden, raise TermWidth or pass Columns to see all)
```

At least one column always survives, so a narrow terminal still shows
something useful.

Note that the width measurement counts runes, not display cells. It never
splits a multi-byte character, but very wide CJK text may render wider than
measured; set `TermWidth` a little conservatively if that matters.

## For agents and scripts

JSON is the right shape to emit for machine consumption, and it is the
default off a terminal. Timestamps are humanized before rendering, so a
Unix stamp becomes `2026-03-11 08:00:00` rather than `1773300000`:

```bash
# compact, one line per response
ikuai_tool monitor system
```

For a stable, machine-only stream, build the renderer explicitly:

```go
r := output.New(w, output.FormatJSON) // never a table, never pretty
```

Empty results stay valid JSON (`[]` / `{}`), so a consumer never has to
special-case the "(no data)" marker.
