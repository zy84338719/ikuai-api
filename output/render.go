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
	hidden := 0
	if r.opts.TermWidth > 0 && len(keys) > 1 {
		fit, h := fitColumns(keys, items, r.opts.TermWidth)
		keys = fit
		hidden = h
	}

	headers := make([]string, len(keys))
	for i, k := range keys {
		headers[i] = HeaderName(k)
	}

	rows := make([][]string, 0, len(items)+1)
	rows = append(rows, headers)
	for _, item := range items {
		obj, _ := item.(map[string]any)
		cells := make([]string, len(keys))
		for i, k := range keys {
			cells[i] = Cell(obj[k])
		}
		rows = append(rows, cells)
	}
	err := writeRows(r.w, rows)
	// The hidden-columns note is best-effort relative to the table: report
	// it only when the table itself went through, so the primary failure
	// is never masked.
	if hidden > 0 {
		if _, ferr := fmt.Fprintf(r.w, "(%d columns hidden, raise TermWidth or pass Columns to see all)\n", hidden); ferr != nil && err == nil {
			err = ferr
		}
	}
	return err
}

func (r *Renderer) list(items []any) error {
	tw := tabwriter.NewWriter(r.w, 0, 0, 2, ' ', 0)
	for _, v := range items {
		// tabwriter buffers and its Write is infallible today; check
		// anyway so the error can never be silently dropped if that
		// ever changes.
		if _, err := fmt.Fprintf(tw, "%s\n", Cell(v)); err != nil {
			return err
		}
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
		if _, err := fmt.Fprintf(r.w, "== %s ==\n", strings.ToUpper(HeaderName(title))); err != nil {
			return err
		}
	}
	rows := make([][]string, 0, len(keys))
	for _, k := range keys {
		rows = append(rows, []string{HeaderName(k), Cell(obj[k])})
	}
	return writeRows(r.w, rows)
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

// writeRows writes rows left-aligned in display columns: two spaces between
// columns, no trailing padding after the last one — the same shape
// text/tabwriter produced with minwidth 0, padding 2, but measured in
// terminal columns. tabwriter counts runes, so a CJK cell pushed every
// column after it out of alignment; arrayTable and keyValue use this
// instead. (list keeps tabwriter: it has no tab-separated columns to align.)
//
// A cell may span lines (selector flattening); its width is that of its
// widest physical line, mirroring columnWidths.
func writeRows(w io.Writer, rows [][]string) error {
	cols := 0
	for _, row := range rows {
		if len(row) > cols {
			cols = len(row)
		}
	}
	widths := make([]int, cols)
	for _, row := range rows {
		for i, cell := range row {
			for _, line := range strings.Split(cell, "\n") {
				if wd := displayWidth(line); wd > widths[i] {
					widths[i] = wd
				}
			}
		}
	}
	var b strings.Builder
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				b.WriteString("  ")
			}
			b.WriteString(cell)
			if i < len(row)-1 {
				// Pad from the cell's last physical line: earlier lines of a
				// multi-line cell already ended with their own newline.
				last := cell
				if nl := strings.LastIndexByte(cell, '\n'); nl >= 0 {
					last = cell[nl+1:]
				}
				if pad := widths[i] - displayWidth(last); pad > 0 {
					b.WriteString(strings.Repeat(" ", pad))
				}
			}
		}
		b.WriteByte('\n')
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// displayWidth estimates how many terminal columns a string occupies.
// East-Asian fullwidth characters (CJK ideographs, kana, hangul, fullwidth
// forms) take two columns each; combining marks and control characters take
// none; everything else takes one. Counting runes instead — the old
// implementation — misaligned every table with Chinese content: 主网关 is
// 3 runes but renders across 6 columns, so every column after it drifted
// left by one per CJK character (and text/tabwriter measured the same way,
// which is why the alignment tests need their own referee).
//
// The tables below cover what router output actually contains (interface
// names, comments and descriptions in CJK); they are not a full East Asian
// Width database. Same zero-dependency approach as oekit/pkg/cli's width
// table — go-runewidth would break this repo's one-require rule.
func displayWidth(s string) int {
	w := 0
	for _, r := range s {
		w += runeWidth(r)
	}
	return w
}

// runeWidth returns the terminal column count of a single rune.
func runeWidth(r rune) int {
	switch {
	case r < 0x20 || (r >= 0x7F && r < 0xA0): // C0/C1 control characters
		return 0
	case inRanges(r, zeroWidthRanges[:]):
		return 0
	case inRanges(r, wideRanges[:]):
		return 2
	}
	return 1
}

// inRanges reports whether r falls inside a lo,hi-pair table sorted by lo.
func inRanges(r rune, table []uint32) bool {
	if r < 0 {
		return false
	}
	v := uint32(r)
	lo, hi := 0, len(table)/2-1
	for lo <= hi {
		mid := (lo + hi) / 2
		start, end := table[mid*2], table[mid*2+1]
		switch {
		case v < start:
			hi = mid - 1
		case v > end:
			lo = mid + 1
		default:
			return true
		}
	}
	return false
}

// wideRanges: two-column code points (East Asian Width W/F), lo,hi pairs
// sorted by lo.
var wideRanges = [...]uint32{
	0x1100, 0x115F, // Hangul Jamo initial consonants
	0x2E80, 0x303E, // CJK radicals, kana, CJK symbols (incl. U+3000 ideographic space)
	0x3041, 0x33FF, // kana, CJK compatibility, enclosed CJK letters
	0x3400, 0x4DBF, // CJK extension A
	0x4E00, 0x9FFF, // CJK unified ideographs
	0xA000, 0xA4CF, // Yi
	0xA960, 0xA97C, // Hangul Jamo extended-A
	0xAC00, 0xD7A3, // Hangul syllables
	0xF900, 0xFAFF, // CJK compatibility ideographs
	0xFE10, 0xFE19, // vertical forms
	0xFE30, 0xFE4F, // CJK compatibility forms
	0xFF01, 0xFF60, // fullwidth ASCII variants
	0xFFE0, 0xFFE6, // fullwidth symbols
	0x1F300, 0x1F64F, // emoji blocks
	0x1F680, 0x1F6FF,
	0x1F900, 0x1F9FF,
	0x20000, 0x3FFFD, // CJK extension B and beyond
}

// zeroWidthRanges: code points that occupy no column (combining marks,
// format characters, variation selectors), lo,hi pairs sorted by lo.
var zeroWidthRanges = [...]uint32{
	0x0300, 0x036F, // combining diacritical marks
	0x0483, 0x0489,
	0x0591, 0x05BD, // Hebrew points
	0x0610, 0x061A, // Arabic marks
	0x064B, 0x065F,
	0x0E31, 0x0E31, // Thai
	0x1AB0, 0x1AFF, // combining diacriticals extended
	0x1DC0, 0x1DFF, // combining diacriticals supplement
	0x200B, 0x200F, // zero-width space / directional marks
	0x2060, 0x206F,
	0x20D0, 0x20F0, // combining marks for symbols
	0xFE00, 0xFE0F, // variation selectors
	0xFE20, 0xFE2F, // combining half marks
	0x1D167, 0x1D169,
	0xE0100, 0xE01EF, // variation selectors supplement
}
