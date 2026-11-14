package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func renderJSON(t *testing.T, v any, opts RenderOptions) string {
	t.Helper()
	var buf bytes.Buffer
	r := New(&buf, FormatJSON)
	if opts.HumanTime {
		r = r.WithOptions(opts)
	}
	if err := r.Value(v); err != nil {
		t.Fatalf("render: %v", err)
	}
	return buf.String()
}

func TestCellFlattensSelectorPair(t *testing.T) {
	// The firmware's address/port selector pair must read as lines, not as
	// an opaque JSON blob.
	got := Cell(map[string]any{
		"custom": []any{"192.168.1.0/24", "10.0.0.5"},
		"object": []any{},
	})
	want := "192.168.1.0/24\n10.0.0.5"
	if got != want {
		t.Errorf("Cell = %q, want %q", got, want)
	}

	// An empty selector pair means "any" in the web UI.
	if got := Cell(map[string]any{"custom": []any{}, "object": []any{}}); got != "*" {
		t.Errorf("empty selector = %q, want \"*\"", got)
	}
}

func TestCellFlattensObjectGroupMembers(t *testing.T) {
	got := Cell([]any{
		map[string]any{"ip": "192.168.1.10"},
		map[string]any{"ip": "192.168.1.11"},
	})
	want := "ip: 192.168.1.10\nip: 192.168.1.11"
	if got != want {
		t.Errorf("Cell = %q, want %q", got, want)
	}
}

func TestCellScalars(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"nil", nil, ""},
		{"bool", true, "true"},
		{"integral float", float64(1500), "1500"},
		{"fractional float", 1.5, "1.5"},
		{"newline is escaped", "a\nb", `a\nb`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Cell(tc.in); got != tc.want {
				t.Errorf("Cell(%#v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestHeaderName(t *testing.T) {
	tests := map[string]string{
		"tagname":     "TAGNAME",
		"src_addr":    "SRC_ADDR",
		"lan_ip":      "LAN_IP",
		"srcIP":       "SRC_IP",
		"srcIPAddr":   "SRC_IP_ADDR",
		"ip_addr":     "IP_ADDR",
		"total_down":  "TOTAL_DOWN",
		"already_ok":  "ALREADY_OK",
		"enableHTTPS": "ENABLE_HTTPS",
	}
	for in, want := range tests {
		if got := HeaderName(in); got != want {
			t.Errorf("HeaderName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestTableRendersColumns(t *testing.T) {
	var buf bytes.Buffer
	r := New(&buf, FormatTable)
	data := []any{
		map[string]any{"tagname": "office", "ip_addr": "192.168.1.100", "enabled": "yes"},
		map[string]any{"tagname": "guest", "ip_addr": "192.168.1.101", "enabled": "no"},
	}
	if err := r.Value(data); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"TAGNAME", "IP_ADDR", "ENABLED", "office", "192.168.1.100"} {
		if !strings.Contains(out, want) {
			t.Errorf("table output missing %q:\n%s", want, out)
		}
	}
}

func TestTableColumnSelection(t *testing.T) {
	var buf bytes.Buffer
	r := New(&buf, FormatTable).WithOptions(RenderOptions{Columns: []string{"tagname"}})
	if err := r.Value([]any{map[string]any{"tagname": "a", "ip_addr": "10.0.0.1"}}); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "TAGNAME") {
		t.Errorf("expected the selected column:\n%s", out)
	}
	if strings.Contains(out, "IP_ADDR") {
		t.Errorf("unselected column should be omitted:\n%s", out)
	}
}

func TestTableUnwrapsSingleKeyEnvelope(t *testing.T) {
	var buf bytes.Buffer
	r := New(&buf, FormatTable)
	if err := r.Value(map[string]any{
		"data": []any{map[string]any{"tagname": "office"}},
	}); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "DATA") {
		t.Errorf("a single-key envelope should render its collection, not a DATA column:\n%s", out)
	}
	if !strings.Contains(out, "office") {
		t.Errorf("expected the row to be rendered:\n%s", out)
	}
}

func TestTableEmptyData(t *testing.T) {
	var buf bytes.Buffer
	if err := New(&buf, FormatTable).Value([]any{}); err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(buf.String(), "(no data)") {
		t.Errorf("want an empty-data marker, got %q", buf.String())
	}
}

func TestHumanTimeAppliesToJSONToo(t *testing.T) {
	out := renderJSON(t,
		map[string]any{"timestamp": float64(1773300000)},
		RenderOptions{HumanTime: true},
	)
	if strings.Contains(out, "1773300000") {
		t.Errorf("timestamp should be humanized in JSON output too, got %s", out)
	}
	if !strings.Contains(out, "-") {
		t.Errorf("expected a formatted local time, got %s", out)
	}
}

func TestJSONRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	r := New(&buf, FormatJSON)
	if err := r.JSON(json.RawMessage(`{"code":20000,"data":[{"id":1}]}`)); err != nil {
		t.Fatalf("render: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v (%s)", err, buf.String())
	}
}

func TestParseFormat(t *testing.T) {
	for _, in := range []string{"", "auto", "table", "JSON", " json "} {
		if _, err := ParseFormat(in); err != nil {
			t.Errorf("ParseFormat(%q) failed: %v", in, err)
		}
	}
	if _, err := ParseFormat("yaml"); err == nil {
		t.Error("expected an error for an unsupported format")
	}
}

func TestTermWidthDropsTrailingColumns(t *testing.T) {
	var buf bytes.Buffer
	// Columns are ordered by name, so "comment" leads and "tagname" is last.
	items := []any{map[string]any{
		"tagname": "office", "ip_addr": "192.168.1.100", "comment": "a long comment here",
	}}
	r := New(&buf, FormatTable).WithOptions(RenderOptions{TermWidth: 30})
	if err := r.Value(items); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "COMMENT") {
		t.Errorf("the leading column must always survive:\n%s", out)
	}
	if strings.Contains(out, "IP_ADDR") || strings.Contains(out, "TAGNAME") {
		t.Errorf("columns that overflow the width should be dropped whole:\n%s", out)
	}
	if !strings.Contains(out, "columns hidden") {
		t.Errorf("dropping columns should be reported:\n%s", out)
	}

	// A width large enough for everything keeps every column.
	buf.Reset()
	r = New(&buf, FormatTable).WithOptions(RenderOptions{TermWidth: 200})
	if err := r.Value(items); err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(buf.String(), "TAGNAME") {
		t.Errorf("a wide terminal should keep every column:\n%s", buf.String())
	}
}

// termCols measures what a terminal actually renders: fullwidth CJK
// characters take two columns. It is deliberately independent of
// displayWidth — the alignment tests need a referee that is not the
// implementation under test (using displayWidth made both sides count
// runes, and the two errors cancelled out).
func termCols(s string) int {
	w := 0
	for _, r := range s {
		if (r >= 0x1100 && r <= 0x115F) ||
			(r >= 0x2E80 && r <= 0xA4CF && r != 0x303F) ||
			(r >= 0xAC00 && r <= 0xD7A3) ||
			(r >= 0xF900 && r <= 0xFAFF) ||
			(r >= 0xFF01 && r <= 0xFF60) ||
			(r >= 0xFFE0 && r <= 0xFFE6) ||
			(r >= 0x20000 && r <= 0x3FFFD) {
			w += 2
			continue
		}
		w++
	}
	return w
}

// TestDisplayWidthCJK locks the width contract: East-Asian fullwidth
// characters occupy two terminal columns, not one. Counting runes made
// 主网关 measure 3 while the terminal renders it across 6 columns.
func TestDisplayWidthCJK(t *testing.T) {
	cases := map[string]int{
		"":            0,
		"gw0":         3,
		"主网关":         6,
		"备用":          4,
		"192.168.1.1": 11,
		"全角Ａ":         6, // 全角 forms are wide too
	}
	for in, want := range cases {
		if got := displayWidth(in); got != want {
			t.Errorf("displayWidth(%q) = %d, want %d", in, got, want)
		}
	}
}

// TestTableCJKAlignment runs the real render path (arrayTable) and asserts
// every row's IP cell starts at the same terminal column. tabwriter and the
// old rune-counting displayWidth both measured CJK cells one column per
// character, so rows with Chinese descriptions shifted the columns after
// them (26 vs 25 columns in the probe that found this).
func TestTableCJKAlignment(t *testing.T) {
	var buf bytes.Buffer
	items := []any{
		map[string]any{"name": "gw0", "desc": "主网关", "ip": "192.168.1.1"},
		map[string]any{"name": "gw1", "desc": "备用", "ip": "10.0.0.1"},
	}
	r := New(&buf, FormatTable)
	if err := r.Value(items); err != nil {
		t.Fatalf("render: %v", err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("want header + 2 rows, got %d:\n%s", len(lines), buf.String())
	}
	termCol := func(line, cell string) int {
		t.Helper()
		i := strings.Index(line, cell)
		if i < 0 {
			t.Fatalf("cell %q missing from line %q", cell, line)
		}
		return termCols(line[:i])
	}
	colA := termCol(lines[1], "192.168.1.1")
	colB := termCol(lines[2], "10.0.0.1")
	if colA != colB {
		t.Errorf("IP column starts at terminal column %d vs %d (misaligned):\n%s",
			colA, colB, buf.String())
	}
}

// TestKeyValueCJKAlignment covers the second alignment path: a Chinese key
// must count as two columns per character when padding the value column.
func TestKeyValueCJKAlignment(t *testing.T) {
	var buf bytes.Buffer
	r := New(&buf, FormatTable)
	if err := r.Value(map[string]any{"网关名": "主网关", "a": "x"}); err != nil {
		t.Fatalf("render: %v", err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 rows, got %d:\n%s", len(lines), buf.String())
	}
	termCol := func(line, cell string) int {
		t.Helper()
		i := strings.Index(line, cell)
		if i < 0 {
			t.Fatalf("cell %q missing from line %q", cell, line)
		}
		return termCols(line[:i])
	}
	colA := termCol(lines[0], "x")
	colB := termCol(lines[1], "主网关")
	if colA != colB {
		t.Errorf("value column starts at terminal column %d vs %d (misaligned):\n%s",
			colA, colB, buf.String())
	}
}
