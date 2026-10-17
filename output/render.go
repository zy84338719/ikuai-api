// Package output renders iKuai API responses for humans and for machines.
//
// The same response is useful in two shapes: a table for a person reading
// a terminal, and JSON for a script or an agent piping it somewhere. The
// renderer picks the shape, and both handle the two structures the
// firmware leans on heavily — the {"custom": [...], "object": [...]}
// selector pair and the group_value member array — instead of dumping raw
// JSON into a cell.
//
// Only the standard library is used. YAML is intentionally not implemented
// here; callers that want it can marshal to JSON and convert.
package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

// Format selects the rendering shape.
type Format string

const (
	// FormatTable renders aligned columns for a terminal.
	FormatTable Format = "table"
	// FormatJSON renders compact JSON for scripts and agents.
	FormatJSON Format = "json"
	// FormatAuto renders a table for a terminal and JSON otherwise.
	FormatAuto Format = "auto"
)

// ParseFormat validates a format name. The empty string and "auto" both
// resolve to FormatAuto, which is the sensible default for a program that
// may be run interactively or piped.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "auto":
		return FormatAuto, nil
	case "table":
		return FormatTable, nil
	case "json":
		return FormatJSON, nil
	}
	return "", fmt.Errorf("output: unknown format %q (want table, json or auto)", s)
}

// RenderOptions controls a single render.
type RenderOptions struct {
	// Format selects the shape. Zero value means FormatAuto.
	Format Format
	// Columns restricts and orders the table columns. Empty means every
	// column, in sorted order. It has no effect on JSON.
	Columns []string
	// HumanTime converts timestamp fields from Unix seconds to local time.
	HumanTime bool
	// TermWidth, when greater than zero, drops trailing columns that would
	// overflow the terminal instead of truncating them.
	TermWidth int
	// Indent pretty-prints JSON. Compact JSON is the better default for
	// scripts and agents, so this is off unless asked for.
	Indent bool
}

// Renderer writes responses to an output stream.
type Renderer struct {
	w      io.Writer
	format Format
	opts   RenderOptions
}

// New builds a renderer. A zero Format is treated as FormatAuto.
func New(w io.Writer, format Format) *Renderer {
	return &Renderer{w: w, format: format}
}

// WithOptions returns a copy of the renderer carrying opts. Fields the
// caller left unset keep their current value.
func (r *Renderer) WithOptions(opts RenderOptions) *Renderer {
	next := *r
	if opts.Format != "" {
		next.format = opts.Format
	}
	if len(opts.Columns) > 0 {
		next.opts.Columns = opts.Columns
	}
	if opts.HumanTime {
		next.opts.HumanTime = true
	}
	if opts.TermWidth > 0 {
		next.opts.TermWidth = opts.TermWidth
	}
	if opts.Indent {
		next.opts.Indent = true
	}
	return &next
}

// JSON writes the raw payload exactly as the client returned it, which is
// the closest thing to the wire format the SDK offers.
func (r *Renderer) JSON(raw json.RawMessage) error {
	if r.format != FormatTable {
		return r.writeJSON(raw)
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		_, writeErr := fmt.Fprintln(r.w, string(raw))
		return writeErr
	}
	return r.render(r.format, v)
}

// Value renders a decoded value.
func (r *Renderer) Value(v any) error { return r.render(r.format, v) }

func (r *Renderer) render(format Format, v any) error {
	if r.opts.HumanTime {
		v = humanizeTimestamps(v)
	}
	// With no explicit format, a table is only useful on a terminal; a
	// piped program gets JSON it can actually consume.
	if format == FormatAuto || format == "" {
		if IsTerminal(r.w) {
			format = FormatTable
		} else {
			format = FormatJSON
		}
	}
	if format == FormatTable {
		return r.table(v)
	}
	return r.writeJSONValue(v)
}

func (r *Renderer) writeJSON(raw json.RawMessage) error {
	if r.opts.Indent {
		return writeIndented(r.w, raw)
	}
	_, err := fmt.Fprintln(r.w, strings.TrimRight(string(raw), "\n"))
	return err
}

func (r *Renderer) writeJSONValue(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if r.opts.Indent {
		return writeIndented(r.w, data)
	}
	_, err = fmt.Fprintln(r.w, string(data))
	return err
}

func writeIndented(w io.Writer, data []byte) error {
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err != nil {
		_, werr := fmt.Fprintln(w, string(data))
		return werr
	}
	_, err := fmt.Fprintln(w, buf.String())
	return err
}

func (r *Renderer) table(v any) error {
	switch t := v.(type) {
	case nil:
		_, err := fmt.Fprintln(r.w, "(no data)")
		return err
	case []any:
		if len(t) == 0 {
			_, err := fmt.Fprintln(r.w, "(no data)")
			return err
		}
		if allObjects(t) {
			return r.arrayTable(t)
		}
		return r.list(t)
	case map[string]any:
		if len(t) == 0 {
			_, err := fmt.Fprintln(r.w, "(no data)")
			return err
		}
		// A lone collection under one key renders as that collection, so
		// {"data":[{...}]} does not print a useless DATA column.
		for _, key := range []string{"data", "results", "items"} {
			if inner, ok := t[key]; ok && len(t) == 1 {
				return r.table(inner)
			}
		}
		if objects, ok := t["objects"]; ok && len(t) == 1 {
			return r.table(objects)
		}
		return r.keyValue("", t)
	default:
		_, err := fmt.Fprintln(r.w, Cell(v))
		return err
	}
}

func (r *Renderer) arrayTable(items []any) error {
	keys := r.opts.Columns
	if len(keys) == 0 {
		keys = collectKeys(items)
	}
	if len(keys) == 0 {
		_, err := fmt.Fprintln(r.w, "(no data)")
		return err
	}
	if r.opts.TermWidth > 0 && len(keys) > 1 {
		fit, hidden := fitColumns(keys, items, r.opts.TermWidth)
		keys = fit
		defer func() {
			if hidden > 0 {
				fmt.Fprintf(r.w, "(%d columns hidden, raise TermWidth or pass Columns to see all)\n", hidden)
			}
		}()
	}

	headers := make([]string, len(keys))
	for i, k := range keys {
		headers[i] = HeaderName(k)
	}

	tw := tabwriter.NewWriter(r.w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, strings.Join(headers, "\t"))
	for _, item := range items {
		obj, _ := item.(map[string]any)
		cells := make([]string, len(keys))
		for i, k := range keys {
			cells[i] = Cell(obj[k])
		}
		fmt.Fprintln(tw, strings.Join(cells, "\t"))
	}
	return tw.Flush()
}

func (r *Renderer) list(items []any) error {
	tw := tabwriter.NewWriter(r.w, 0, 0, 2, ' ', 0)
	for _, v := range items {
		fmt.Fprintf(tw, "%s\n", Cell(v))
	}
	return tw.Flush()
}

func (r *Renderer) keyValue(title string, obj map[string]any) error {
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if title != "" {
		fmt.Fprintf(r.w, "== %s ==\n", strings.ToUpper(HeaderName(title)))
	}
	tw := tabwriter.NewWriter(r.w, 0, 0, 2, ' ', 0)
	for _, k := range keys {
		fmt.Fprintf(tw, "%s\t%s\n", HeaderName(k), Cell(obj[k]))
	}
	return tw.Flush()
}

// Cell renders one value for a table. The nested structures the firmware
// uses are flattened into readable lines; anything else falls back to
// compact JSON rather than a Go-flavoured value.
func Cell(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return sanitizeCell(t)
	case bool:
		if t {
			return "true"
		}
		return "false"
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%g", t)
	case map[string]any:
		if s := flattenSelector(t); s != "" {
			return s
		}
	case []any:
		if s := flattenMembers(t); s != "" {
			return s
		}
	}
	if data, err := json.Marshal(v); err == nil {
		return sanitizeCell(string(data))
	}
	return sanitizeCell(fmt.Sprintf("%v", v))
}

// flattenSelector renders the {"custom": [...], "object": [...]} pair the
// firmware uses for every address, port and protocol selector.
func flattenSelector(m map[string]any) string {
	custom, hasCustom := m["custom"]
	object, hasObject := m["object"]
	if !hasCustom && !hasObject {
		return ""
	}
	var lines []string
	if arr, ok := custom.([]any); ok {
		for _, v := range arr {
			if s := memberLine(v); s != "" {
				lines = append(lines, s)
			}
		}
	}
	if arr, ok := object.([]any); ok {
		for _, v := range arr {
			if s := memberLine(v); s != "" {
				lines = append(lines, "["+s+"]")
			}
		}
	}
	if len(lines) == 0 {
		// The key pair exists but selects nothing; "*" is the web UI's
		// marker for "any", so mirror it.
		return "*"
	}
	return strings.Join(lines, "\n")
}

// flattenMembers renders a group_value array, whose members are
// single-key objects such as {"ip": "192.168.1.10"}.
func flattenMembers(arr []any) string {
	rows := make([]map[string]any, 0, len(arr))
	for _, v := range arr {
		if m, ok := v.(map[string]any); ok && len(m) > 0 {
			rows = append(rows, m)
		}
	}
	if len(rows) == 0 {
		return ""
	}
	singleKey := true
	for _, m := range rows {
		if len(m) != 1 {
			singleKey = false
			break
		}
	}
	var lines []string
	for _, m := range rows {
		if singleKey {
			for k, v := range m {
				lines = append(lines, fmt.Sprintf("%s: %v", k, v))
			}
			continue
		}
		sub := make([]string, 0, len(m))
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if s := Cell(m[k]); s != "" {
				sub = append(sub, fmt.Sprintf("%s: %s", k, s))
			}
		}
		lines = append(lines, strings.Join(sub, ", "))
	}
	return strings.Join(lines, "\n")
}

// memberLine renders one element of a selector array.
func memberLine(v any) string {
	if m, ok := v.(map[string]any); ok {
		sub := make([]string, 0, len(m))
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sub = append(sub, fmt.Sprintf("%s: %v", k, m[k]))
		}
		return strings.Join(sub, ", ")
	}
	return fmt.Sprintf("%v", v)
}

func allObjects(items []any) bool {
	if len(items) == 0 {
		return false
	}
	for _, v := range items {
		if _, ok := v.(map[string]any); !ok {
			return false
		}
	}
	return true
}

func collectKeys(items []any) []string {
	seen := map[string]bool{}
	for _, v := range items {
		obj, ok := v.(map[string]any)
		if !ok {
			continue
		}
		for k := range obj {
			seen[k] = true
		}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func columnWidths(headers, keys []string, items []any, humanTime bool) []int {
	widths := make([]int, len(keys))
	for i, h := range headers {
		widths[i] = displayWidth(h)
	}
	for _, v := range items {
		obj, _ := v.(map[string]any)
		for i, k := range keys {
			cell := Cell(obj[k])
			if humanTime && k == "timestamp" {
				cell = FormatTimestamp(obj[k])
			}
			// A multi-line cell contributes its widest line, not its
			// total length.
			widest := 0
			for _, line := range strings.Split(cell, "\n") {
				if w := displayWidth(line); w > widest {
					widest = w
				}
			}
			if widest > widths[i] {
				widths[i] = widest
			}
		}
	}
	return widths
}

// fitColumns keeps as many leading columns as fit within termWidth,
// dropping whole trailing columns rather than truncating their contents.
func fitColumns(keys []string, items []any, termWidth int) (fit []string, hidden int) {
	total := 0
	for i, k := range keys {
		headers := []string{HeaderName(k)}
		w := columnWidths(headers, []string{k}, items, false)[0]
		if i > 0 {
			total += 2 // column separator
		}
		if total+w > termWidth && i > 0 {
			return keys[:i], len(keys) - i
		}
		total += w
	}
	return keys, 0
}

// HeaderName converts a field name to the UPPER_SNAKE_CASE form the web UI
// uses, so CLI output and router output read the same. Acronym boundaries
// are handled: srcIP becomes SRC_IP.
func HeaderName(field string) string {
	var b strings.Builder
	runes := []rune(field)
	for i, r := range runes {
		if r == '_' || r == '-' || r == ' ' {
			if b.Len() > 0 && !strings.HasSuffix(b.String(), "_") {
				b.WriteRune('_')
			}
			continue
		}
		isUpper := r >= 'A' && r <= 'Z'
		isLower := r >= 'a' && r <= 'z'
		isDigit := r >= '0' && r <= '9'
		if !isUpper && !isLower && !isDigit {
			b.WriteRune(r)
			continue
		}

		// A word starts when we leave a lower-case run (lanIp) or when an
		// acronym ends (srcIPAddr).
		newWord := false
		if i > 0 {
			prev := runes[i-1]
			prevLower := prev >= 'a' && prev <= 'z'
			prevDigit := prev >= '0' && prev <= '9'
			prevUpper := prev >= 'A' && prev <= 'Z'
			if prevLower && isUpper {
				newWord = true
			} else if prevUpper && isUpper && i+1 < len(runes) {
				nextLower := runes[i+1] >= 'a' && runes[i+1] <= 'z'
				newWord = nextLower
			} else if prevDigit && isUpper {
				newWord = true
			}
		}
		if newWord && b.Len() > 0 && !strings.HasSuffix(b.String(), "_") {
			b.WriteRune('_')
		}
		if isLower {
			r -= 'a' - 'A'
		}
		b.WriteRune(r)
	}
	return b.String()
}

// FormatTimestamp renders a Unix-seconds field as local time. A non-timestamp
// value is rendered as a normal cell.
func FormatTimestamp(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case float64:
		return time.Unix(int64(t), 0).Local().Format("2006-01-02 15:04:05")
	case int64:
		return time.Unix(t, 0).Local().Format("2006-01-02 15:04:05")
	}
	return Cell(v)
}

// humanizeTimestamps rewrites every timestamp field in place. It applies to
// JSON output too, because a Unix timestamp is unreadable in a log dump.
func humanizeTimestamps(v any) any {
	switch t := v.(type) {
	case []any:
		for i := range t {
			t[i] = humanizeTimestamps(t[i])
		}
		return t
	case map[string]any:
		for k, val := range t {
			if k == "timestamp" {
				if s := FormatTimestamp(val); s != "" {
					t[k] = s
				}
				continue
			}
			t[k] = humanizeTimestamps(val)
		}
		return t
	}
	return v
}

// sanitizeCell keeps control characters from breaking table alignment.
func sanitizeCell(s string) string {
	if !strings.ContainsAny(s, "\t\n\r") {
		return s
	}
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.ReplaceAll(s, "\r", "\\r")
	return strings.ReplaceAll(s, "\n", "\\n")
}

// displayWidth counts runes rather than bytes. It is a close approximation
// for CJK text, which the standard library cannot measure exactly, and it
// never splits a multi-byte character.
func displayWidth(s string) int {
	return len([]rune(s))
}
