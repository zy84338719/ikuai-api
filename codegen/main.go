// codegen generates the v4 service layer from v4_catalog.go.
//
// Usage:
//
//	go run ./codegen
//
// The generator reads the V4EndpointCatalog literal exported by the
// parent package, partitions the endpoints by group, and writes one
// service/<group>.go file per group plus a service/root.go that wires
// all groups into a single APIClient entry point.
//
// The generated code has no business logic: each method is a thin
// wrapper that picks the right path from the catalog and delegates to
// the underlying *ikuaiapi.Client. Typed request/response structs are
// deliberately omitted so the SDK stays decoupled from the iKuai
// firmware's wire format; callers decode the returned json.RawMessage
// into a struct of their own.
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	// The generator reuses the parent package's query-spec parser so the
	// catalog grammar has exactly one implementation. The parent package
	// does not import this directory, so there is no cycle.
	ikuaiapi "github.com/zy84338719/ikuai-api"
)

// V4Endpoint is a copy of the runtime type used purely for codegen.
// It is never imported by the generated code.
type V4Endpoint struct {
	Group, Name, Path string
	Methods           []string
	Load              bool
	Action            string
	Query             string
}

const (
	catalogFile = "v4_catalog.go"
	outputDir   = "service"
)

func main() {
	catalog, err := parseCatalog(catalogFile)
	if err != nil {
		log.Fatalf("parse catalog: %v", err)
	}
	log.Printf("loaded %d endpoints", len(catalog))

	byGroup := make(map[string][]V4Endpoint)
	for _, ep := range catalog {
		byGroup[ep.Group] = append(byGroup[ep.Group], ep)
	}
	groups := make([]string, 0, len(byGroup))
	for g := range byGroup {
		groups = append(groups, g)
	}
	sort.Strings(groups)

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		log.Fatalf("mkdir %s: %v", outputDir, err)
	}

	seen := make(map[string]struct{})
	for _, g := range groups {
		eps := byGroup[g]
		sort.Slice(eps, func(i, j int) bool { return eps[i].Name < eps[j].Name })
		fp := filepath.Join(outputDir, g+".go")
		buf := &bytes.Buffer{}
		if err := renderGroup(buf, g, eps); err != nil {
			log.Fatalf("render %s: %v", g, err)
		}
		if err := os.WriteFile(fp, buf.Bytes(), 0o644); err != nil {
			log.Fatalf("write %s: %v", fp, err)
		}
		seen[g] = struct{}{}
		log.Printf("  wrote %s (%d methods)", fp, len(eps))
	}

	if err := renderRoot(filepath.Join(outputDir, "root.go"), groups); err != nil {
		log.Fatalf("render root: %v", err)
	}
	log.Printf("  wrote %s/root.go (entry point)", outputDir)
}

// parseCatalog reads v4_catalog.go and extracts every V4Endpoint
// composite literal. The file is small (~200 lines) and the literals
// follow a single stable shape, so a regex pass is sufficient and
// avoids the import cycle that would come from importing the parent
// package.
//
// Each entry has the shape:
//
//	{Group: "...", Name: "...", Path: "...", Methods: []string{"..."}, Load: true, Action: "start"},
//
// The trailing Load / Action are optional. Load marks monitoring load
// endpoints (datetype/start_time/end_time/math). Action marks
// :start/:stop/:restart/:sync/:restore/:check endpoints.
var (
	// entryRe captures the full composite literal (up to the closing
	// "}," at end of line) so we can scan for the optional Load and
	// Action fields inside the whole entry rather than just the part
	// matched by epRe.
	entryRe  = regexp.MustCompile(`\{[^{}]*Group:\s*"([^"]+)"[^{}]*Name:\s*"([^"]+)"[^{}]*Path:\s*"([^"]+)"[^{}]*Methods:\s*\[\]string\{([^}]*)\}[^{}]*\}`)
	loadRe   = regexp.MustCompile(`Load:\s*true`)
	actionRe = regexp.MustCompile(`Action:\s*"([^"]+)"`)
	queryRe  = regexp.MustCompile(`Query:\s*"([^"]*)"`)
)

func parseCatalog(file string) ([]V4Endpoint, error) {
	src, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	matches := entryRe.FindAllStringSubmatchIndex(string(src), -1)
	catalog := make([]V4Endpoint, 0, len(matches))
	for _, m := range matches {
		ep := V4Endpoint{
			Group: string(matches_get(string(src), m, 1)),
			Name:  string(matches_get(string(src), m, 2)),
			Path:  string(matches_get(string(src), m, 3)),
		}
		methods := string(matches_get(string(src), m, 4))
		for _, ms := range strings.Split(methods, ",") {
			s := strings.TrimSpace(ms)
			s = strings.TrimPrefix(s, `"`)
			s = strings.TrimSuffix(s, `"`)
			if s != "" {
				ep.Methods = append(ep.Methods, s)
			}
		}
		// Scan the entire entry for the optional flags.
		entry := string(src[m[0]:m[1]])
		if loadRe.MatchString(entry) {
			ep.Load = true
		}
		if am := actionRe.FindStringSubmatch(entry); am != nil {
			ep.Action = am[1]
		}
		if qm := queryRe.FindStringSubmatch(entry); qm != nil {
			ep.Query = qm[1]
		}
		catalog = append(catalog, ep)
	}
	return catalog, nil
}

// matches_get returns the capture group n as a string from a regex
// match with byte indices.
func matches_get(src string, m []int, n int) []byte {
	return []byte(src[m[2*n]:m[2*n+1]])
}

// toGoName converts a kebab-case catalog name (e.g. "app-protocols-load")
// to a Go-style CamelCase identifier (e.g. "AppProtocolsLoad").
func toGoName(s string) string {
	var b strings.Builder
	upper := true
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			if upper {
				r -= 'a' - 'A'
			}
			b.WriteRune(r)
			upper = false
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
			upper = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			upper = true
		default:
			upper = true
		}
	}
	return b.String()
}

// methodsOn returns the HTTP methods a generated method should support
// for a given endpoint. The returned set is a subset of {GET, POST, PUT,
// PATCH, DELETE}; some endpoints only allow one verb (e.g. action
// endpoints expose POST only).
func verbs(ep V4Endpoint) map[string]bool {
	out := make(map[string]bool, len(ep.Methods))
	for _, m := range ep.Methods {
		out[strings.ToUpper(m)] = true
	}
	return out
}

func renderGroup(buf *bytes.Buffer, group string, eps []V4Endpoint) error {
	groupTitle := toGoName(group)
	needsStrconv := false
	needsJSON := true // every generated method returns json.RawMessage
	for _, ep := range eps {
		v := verbs(ep)
		if len(v) > 1 || (len(v) == 1 && !v["GET"] && !v["POST"]) {
			needsStrconv = true
		}
		if ep.Load {
			needsStrconv = true
		}
		if ep.Query != "" {
			needsStrconv = true
		}
	}
	imports := []string{"context"}
	if needsJSON {
		imports = append(imports, "encoding/json")
	}
	if needsStrconv {
		imports = append(imports, "strconv")
	}
	imports = append(imports, "ikuaiapi")

	fmt.Fprintln(buf, "// Code generated by ikuai-api/codegen. DO NOT EDIT.")
	fmt.Fprintln(buf, "")
	fmt.Fprintln(buf, "package service")
	fmt.Fprintln(buf, "")
	fmt.Fprintln(buf, "import (")
	for _, imp := range imports {
		if imp == "ikuaiapi" {
			fmt.Fprintln(buf, "\tikuaiapi \"github.com/zy84338719/ikuai-api\"")
		} else {
			fmt.Fprintf(buf, "\t%q\n", imp)
		}
	}
	fmt.Fprintln(buf, ")")

	fmt.Fprintf(buf, `
// %[1]sService is the typed v4 entry point for the "%[2]s" group.
// Method names follow the catalog Name field (CamelCased); the
// underlying HTTP verb and path are documented above each method.
type %[1]sService struct {
	client *ikuaiapi.Client
}

// Path%[1]s is the v4 catalog table for this group; it is exported so
// callers can iterate endpoints or build dynamic requests.
var Path%[1]s = pathGroup%[1]s()

func pathGroup%[1]s() []V4Endpoint {
	src := ikuaiapi.V4EndpointCatalog
	out := make([]V4Endpoint, 0, 16)
	for _, ep := range src {
		if ep.Group == %[3]q {
			out = append(out, ep)
		}
	}
	return out
}

`, groupTitle, group, group)

	// Render one method block per endpoint.
	for _, ep := range eps {
		if err := renderEndpoint(buf, groupTitle, ep); err != nil {
			return err
		}
	}
	// Per-group field-hint comments. These document the field names
	// the iKuai firmware expects in the JSON body for write
	// operations, so callers don't have to grep the web UI source.
	renderFieldHints(buf, group)
	return nil
}

// fieldHints maps a catalog group to a list of (name, description)
// pairs covering the most common write fields. The list is appended
// to the generated service file as a doc-only constant.
var fieldHints = map[string][]struct{ Name, Description string }{
	"network": {
		{"name", "Rule name (required for create, used in list views)"},
		{"action", "NAT action: filter | dnat | snat"},
		{"protocol", "Protocol: tcp | udp | any"},
		{"in_interface", "Inbound interface (wan1, lan1, …)"},
		{"out_interface", "Outbound interface"},
		{"comment", "Free-form comment, shown in the web UI"},
		{"wan_port", "WAN port (NAT rule). Single port, range, or comma list"},
		{"lan_addr", "LAN target address (DNAT). IP or CIDR"},
		{"lan_port", "LAN target port (DNAT)"},
		{"src_addr", "Source address(es). Comma-separated IP/CIDR groups"},
		{"dst_addr", "Destination address(es). Comma-separated IP/CIDR groups"},
		{"src_port", "Source port(s). Comma-separated ports/ranges"},
		{"dst_port", "Destination port(s). Comma-separated ports/ranges"},
		{"enabled", "yes | no. Whether the rule is active"},
	},
	"objects": {
		{"name", "Object name (required)"},
		{"comment", "Free-form comment"},
		{"ip_group", "Array of IP/CIDR strings (for ip-objects)"},
		{"mac", "MAC address (for mac-objects)"},
		{"port_group", "Array of port strings (for port-objects)"},
		{"domain_group", "Array of domain strings (for domain-objects)"},
		{"enabled", "yes | no"},
	},
	"system": {
		{"hostname", "Router hostname"},
		{"timezone", "POSIX timezone (Asia/Shanghai, UTC, ...)"},
		{"dns1", "Primary DNS"},
		{"dns2", "Secondary DNS"},
		{"srcfile", "Backup file name (used by /system/backup:restore)"},
		{"dstfile", "Destination file (used by /system/backup)"},
	},
	"vpn": {
		{"name", "Tunnel name"},
		{"server_ip", "VPN server IP (PPTP / L2TP / OpenVPN server)"},
		{"username", "VPN account username"},
		{"password", "VPN account password"},
		{"enabled", "yes | no"},
	},
	"auth": {
		{"username", "Auth user name (required)"},
		{"password", "Auth user password (raw, not md5)"},
		{"group_id", "Group id (1-based; required)"},
		{"enabled", "yes | no"},
		{"comment", "Free-form comment"},
		{"ip_addr", "Bind IP (optional)"},
		{"sesstimeout", "Session timeout in seconds (0 = no limit)"},
	},
}

func renderFieldHints(buf *bytes.Buffer, group string) {
	hints, ok := fieldHints[group]
	if !ok {
		return
	}
	buf.WriteString("\n// Field hints for this group (iKuai firmware field names):\n")
	for _, h := range hints {
		fmt.Fprintf(buf, "//   %-15s  %s\n", h.Name, h.Description)
	}
}

// prefixed returns the group-prefixed Go name to avoid collisions
// across groups (e.g. /network/ac/services and /system/ac-services both
// yield AcServices if we just CamelCase the catalog name).
func prefixed(groupTitle, name string) string {
	return groupTitle + name
}

func renderEndpoint(buf *bytes.Buffer, groupTitle string, ep V4Endpoint) error {
	name := prefixed(groupTitle, toGoName(ep.Name))
	short := toGoName(ep.Name)
	verbs := verbs(ep)
	hasID := strings.Contains(ep.Path, "?id=") || strings.HasSuffix(ep.Path, "/{id}")

	fmt.Fprintf(buf, "\n// %s wraps %s %s.\n//\n// Methods: %s\n//\n// Path: %s\n",
		name, ep.Group, ep.Name, strings.Join(ep.Methods, ", "), ep.Path)
	if !hasID && verbs["GET"] && !ep.Load && ep.Query == "" {
		fmt.Fprintf(buf, "//\n// Use List%s(ctx, opts...) for paginated reads.\n", name)
	}

	// Load-style monitoring endpoints get a typed LoadOptions struct
	// plus enum validation (datetype/math).
	if ep.Load {
		renderLoad(buf, groupTitle, ep, name)
		return nil
	}
	// Action-style endpoints (path ends in :start/:stop/etc.) become
	// semantically named helpers instead of a generic Do<Name>.
	if ep.Action != "" {
		renderAction(buf, groupTitle, ep, name)
		return nil
	}

	// Endpoints with a non-default query contract get a typed
	// Query<Name> helper that speaks their dialect. The generic CRUD /
	// Get helpers are still emitted so existing callers keep working.
	if ep.Query != "" {
		if err := renderQuery(buf, groupTitle, ep, name); err != nil {
			return err
		}
	}
	if len(verbs) == 1 && verbs["GET"] {
		renderGetOnly(buf, groupTitle, ep, name)
		return nil
	}
	if len(verbs) == 1 && verbs["POST"] {
		renderPostOnly(buf, groupTitle, ep, name)
		return nil
	}
	renderCRUD(buf, groupTitle, ep, name, short, verbs)
	return nil
}

func renderGetOnly(buf *bytes.Buffer, groupTitle string, ep V4Endpoint, name string) {
	fmt.Fprintf(buf, `func (s *%sService) Get%s(ctx context.Context) (json.RawMessage, error) {
	return s.client.Get(ctx, %q, nil)
}

`, groupTitle, name, ep.Path)
}

// dialectReserved lists the query keys each dialect already emits on its
// own. Extra fields whose name collides with one of these are dropped from
// the generated options struct, because the dialect already covers them.
var dialectReserved = map[ikuaiapi.QueryDialect][]string{
	ikuaiapi.DialectPaged:  {"page", "limit", "filter", "order", "order_by"},
	ikuaiapi.DialectLegacy: {"type", "limit", "offset", "order", "order_by"},
	ikuaiapi.DialectMixed:  {"page", "limit", "order", "order_by"},
	ikuaiapi.DialectNone:   nil,
	ikuaiapi.DialectWindow: nil,
	ikuaiapi.DialectRef:    nil,
}

// toQueryFieldName maps a query key to an exported Go field name. The
// router's own casing is preserved on the wire; this only affects the Go
// side. Acronyms are expanded so the generated API reads naturally.
func toQueryFieldName(key string) string {
	switch strings.ToLower(key) {
	case "ip":
		return "IP"
	case "mac":
		return "MAC"
	case "ssid":
		return "SSID"
	case "apmac":
		return "APMAC"
	case "appid":
		return "AppID"
	case "appids":
		return "AppIDs"
	case "starttime":
		return "StartTime"
	case "stoptime":
		return "StopTime"
	case "keywords":
		return "Keywords"
	case "finds":
		return "Finds"
	}
	lower := strings.ToLower(key)
	var b strings.Builder
	upper := true
	for _, r := range lower {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			if upper {
				r -= 'a' - 'A'
			}
			b.WriteRune(r)
			upper = false
			continue
		}
		upper = true
	}
	return b.String()
}

// renderQuery emits <Name>QueryOptions plus a Query<Name> method that
// builds the exact query string the endpoint's dialect expects.
//
// The generic List helper always speaks the default page/limit dialect.
// Endpoints listed in the catalog with a different contract would silently
// ignore that query string, so they get this typed helper instead.
func renderQuery(buf *bytes.Buffer, groupTitle string, ep V4Endpoint, name string) error {
	dialect, required, optional, err := ikuaiapi.ParseQuerySpec(ep.Query)
	if err != nil {
		return fmt.Errorf("%s: %w", ep.Path, err)
	}
	if dialect == ikuaiapi.DialectLoad {
		return nil // already covered by the typed Load<Name> helper
	}

	reserved := map[string]bool{}
	for _, k := range dialectReserved[dialect] {
		reserved[strings.ToLower(k)] = true
	}
	extras := make([]ikuaiapi.QueryField, 0, len(required)+len(optional))
	for _, f := range append(append([]ikuaiapi.QueryField{}, required...), optional...) {
		if reserved[strings.ToLower(f.Name)] {
			continue
		}
		extras = append(extras, f)
	}

	optsName := name + "QueryOptions"
	labels := map[ikuaiapi.QueryDialect]string{
		ikuaiapi.DialectLegacy: "legacy uppercase TYPE/LIMIT/OFFSET dialect",
		ikuaiapi.DialectMixed:  "lowercase paging with uppercase sort/search dialect",
		ikuaiapi.DialectNone:   "unpaginated dialect",
		ikuaiapi.DialectWindow: "starttime/stoptime window dialect",
		ikuaiapi.DialectRef:    "reverse-lookup dialect: one group name, no paging",
	}

	fmt.Fprintf(buf, "\n// %s configures a query against %s.\n", optsName, ep.Path)
	fmt.Fprintf(buf, "//\n// The endpoint speaks the %s.\n", labels[dialect])
	fmt.Fprintln(buf, "// The generic List helper is NOT usable here: it always sends the")
	fmt.Fprintln(buf, "// default page/limit paging, which this endpoint ignores.")
	fmt.Fprintln(buf, "type", optsName, "struct {")

	paged := dialect == ikuaiapi.DialectPaged || dialect == ikuaiapi.DialectMixed || dialect == ikuaiapi.DialectLegacy
	if paged {
		fmt.Fprintln(buf, "\t// Page is the 1-based page number.")
		fmt.Fprintln(buf, "\tPage int")
		fmt.Fprintln(buf, "\t// PageSize is the number of items per page.")
		fmt.Fprintln(buf, "\tPageSize int")
		if dialect == ikuaiapi.DialectPaged {
			fmt.Fprintln(buf, "\t// Filter uses the iKuai expression syntax, e.g. \"enabled==yes\".")
			fmt.Fprintln(buf, "\t// \"&\" is AND and \",\" is OR.")
			fmt.Fprintln(buf, "\tFilter string")
			fmt.Fprintln(buf, "\t// Order is the sort direction: asc | desc.")
			fmt.Fprintln(buf, "\tOrder string")
			fmt.Fprintln(buf, "\t// OrderBy is the sort field.")
			fmt.Fprintln(buf, "\tOrderBy string")
		} else {
			fmt.Fprintln(buf, "\t// Order is the sort direction: asc | desc. Sent as ORDER.")
			fmt.Fprintln(buf, "\tOrder string")
			fmt.Fprintln(buf, "\t// OrderBy is the sort field. Sent as ORDER_BY.")
			fmt.Fprintln(buf, "\tOrderBy string")
		}
	}
	for _, f := range extras {
		field := toQueryFieldName(f.Name)
		typ := "string"
		if f.Numeric {
			typ = "int64"
		}
		req := "optional"
		if f.Required {
			req = "required"
		}
		fmt.Fprintf(buf, "\t// %s is %s (sent as %q).\n", field, req, f.Name)
		fmt.Fprintf(buf, "\t%s %s\n", field, typ)
	}
	fmt.Fprintln(buf, "}")

	fmt.Fprintf(buf, "\nfunc (o *%s) query() (map[string]string, error) {\n", optsName)
	fmt.Fprintf(buf, "\tif o == nil {\n\t\treturn nil, ikuaiapi.NewError(%q)\n\t}\n",
		fmt.Sprintf("%s: options are required", optsName))
	if paged {
		fmt.Fprintf(buf, "\tif o.Page < 1 {\n\t\treturn nil, ikuaiapi.NewError(%q)\n\t}\n",
			optsName+": Page must be greater than 0")
		fmt.Fprintf(buf, "\tif o.PageSize < 1 {\n\t\treturn nil, ikuaiapi.NewError(%q)\n\t}\n",
			optsName+": PageSize must be greater than 0")
	}

	switch dialect {
	case ikuaiapi.DialectLegacy:
		fmt.Fprintln(buf, "\toffset := (o.Page - 1) * o.PageSize")
		fmt.Fprintln(buf, "\tq := map[string]string{")
		fmt.Fprintln(buf, "\t\t\"TYPE\":   \"data\",")
		fmt.Fprintln(buf, "\t\t\"LIMIT\":  strconv.Itoa(offset) + \",\" + strconv.Itoa(o.PageSize),")
		fmt.Fprintln(buf, "\t\t\"OFFSET\": strconv.Itoa(offset),")
		fmt.Fprintln(buf, "\t}")
		fmt.Fprintln(buf, "\tif o.Order != \"\" {")
		fmt.Fprintln(buf, "\t\tq[\"ORDER\"] = o.Order")
		fmt.Fprintln(buf, "\t}")
		fmt.Fprintln(buf, "\tif o.OrderBy != \"\" {")
		fmt.Fprintln(buf, "\t\tq[\"ORDER_BY\"] = o.OrderBy")
		fmt.Fprintln(buf, "\t}")
	case ikuaiapi.DialectMixed:
		fmt.Fprintln(buf, "\tq := map[string]string{")
		fmt.Fprintln(buf, "\t\t\"page\":  strconv.Itoa(o.Page),")
		fmt.Fprintln(buf, "\t\t\"limit\": strconv.Itoa(o.PageSize),")
		fmt.Fprintln(buf, "\t}")
		fmt.Fprintln(buf, "\tif o.Order != \"\" {")
		fmt.Fprintln(buf, "\t\tq[\"ORDER\"] = o.Order")
		fmt.Fprintln(buf, "\t}")
		fmt.Fprintln(buf, "\tif o.OrderBy != \"\" {")
		fmt.Fprintln(buf, "\t\tq[\"ORDER_BY\"] = o.OrderBy")
		fmt.Fprintln(buf, "\t}")
	case ikuaiapi.DialectPaged:
		fmt.Fprintln(buf, "\tq := map[string]string{")
		fmt.Fprintln(buf, "\t\t\"page\":  strconv.Itoa(o.Page),")
		fmt.Fprintln(buf, "\t\t\"limit\": strconv.Itoa(o.PageSize),")
		fmt.Fprintln(buf, "\t}")
		fmt.Fprintln(buf, "\tif o.Filter != \"\" {")
		fmt.Fprintln(buf, "\t\tq[\"filter\"] = o.Filter")
		fmt.Fprintln(buf, "\t}")
		fmt.Fprintln(buf, "\tif o.Order != \"\" {")
		fmt.Fprintln(buf, "\t\tq[\"order\"] = o.Order")
		fmt.Fprintln(buf, "\t}")
		fmt.Fprintln(buf, "\tif o.OrderBy != \"\" {")
		fmt.Fprintln(buf, "\t\tq[\"order_by\"] = o.OrderBy")
		fmt.Fprintln(buf, "\t}")
	default: // DialectNone, DialectWindow
		fmt.Fprintln(buf, "\tq := map[string]string{}")
	}

	for _, f := range extras {
		field := toQueryFieldName(f.Name)
		switch {
		case f.Required && f.Numeric:
			fmt.Fprintf(buf, "\tif o.%s <= 0 {\n\t\treturn nil, ikuaiapi.NewError(%q)\n\t}\n",
				field, fmt.Sprintf("%s: %s must be greater than 0", optsName, field))
			fmt.Fprintf(buf, "\tq[%q] = strconv.FormatInt(o.%s, 10)\n", f.Name, field)
		case f.Required:
			fmt.Fprintf(buf, "\tif o.%s == \"\" {\n\t\treturn nil, ikuaiapi.NewError(%q)\n\t}\n",
				field, fmt.Sprintf("%s: %s is required", optsName, field))
			fmt.Fprintf(buf, "\tq[%q] = o.%s\n", f.Name, field)
		case f.Numeric:
			fmt.Fprintln(buf, "\tif o."+field+" != 0 {")
			fmt.Fprintf(buf, "\t\tq[%q] = strconv.FormatInt(o.%s, 10)\n", f.Name, field)
			fmt.Fprintln(buf, "\t}")
		default:
			fmt.Fprintln(buf, "\tif o."+field+" != \"\" {")
			fmt.Fprintf(buf, "\t\tq[%q] = o.%s\n", f.Name, field)
			fmt.Fprintln(buf, "\t}")
		}
	}
	fmt.Fprintln(buf, "\tif len(q) == 0 {")
	fmt.Fprintln(buf, "\t\treturn nil, nil")
	fmt.Fprintln(buf, "\t}")
	fmt.Fprintln(buf, "\treturn q, nil")
	fmt.Fprintln(buf, "}")

	// A sub-resource query carries the parent identifier before its options.
	parentDeclStr := ""
	for _, pp := range parentParams(ep.Path) {
		parentDeclStr += toGoParam(pp) + " int64, "
	}
	queryPath := pathExpr(ep.Path, parentParams(ep.Path), false)
	fmt.Fprintf(buf, `
// Query%s runs a query against %s using its documented query dialect.
func (s *%sService) Query%s(ctx context.Context, %sopts *%s) (json.RawMessage, error) {
	q, err := opts.query()
	if err != nil {
		return nil, err
	}
	return s.client.Get(ctx, %s, q)
}

`, name, ep.Path, groupTitle, name, parentDeclStr, optsName, queryPath)

	return nil
}

func renderPostOnly(buf *bytes.Buffer, groupTitle string, ep V4Endpoint, name string) {
	fmt.Fprintf(buf, `func (s *%sService) Do%s(ctx context.Context, body any) (json.RawMessage, error) {
	return s.client.Post(ctx, %q, body)
}

`, groupTitle, name, ep.Path)
}

// renderLoad emits a typed Load<Name> method that validates the
// datetype/math enums and the start<end relationship, then sends the
// request as datetype/start_time/end_time/math query params.
func renderLoad(buf *bytes.Buffer, groupTitle string, ep V4Endpoint, name string) {
	optsName := name + "LoadOptions"
	methodName := "Load" + name[len(groupTitle):]
	if methodName == "Load"+name {
		methodName = "Get" + name
	}
	fmt.Fprintf(buf, `// %[1]s configures a load query against %[2]s.
// All fields are required; the router rejects partial queries.
//
// DataType selects the time bucket. iKuai accepts one of
// "hour", "day", "week", "month".
//
// Math is the aggregation. iKuai accepts one of "avg", "max".
//
// StartTime / EndTime are Unix epoch seconds. StartTime must be
// strictly less than EndTime.
type %[1]s struct {
	DataType  string
	StartTime int64
	EndTime   int64
	Math      string
}

func (o *%[1]s) query() (map[string]string, error) {
	if o == nil {
		return nil, ikuaiapi.NewError("%[1]s: options are required")
	}
	switch o.DataType {
	case "hour", "day", "week", "month":
	default:
		return nil, ikuaiapi.NewError("%[1]s: DataType must be one of: hour, day, week, month")
	}
	switch o.Math {
	case "avg", "max":
	default:
		return nil, ikuaiapi.NewError("%[1]s: Math must be one of: avg, max")
	}
	if o.StartTime <= 0 || o.EndTime <= 0 {
		return nil, ikuaiapi.NewError("%[1]s: StartTime and EndTime are required (Unix seconds)")
	}
	if o.StartTime >= o.EndTime {
		return nil, ikuaiapi.NewError("%[1]s: StartTime must be less than EndTime")
	}
	return map[string]string{
		"datetype":   o.DataType,
		"start_time": strconv.FormatInt(o.StartTime, 10),
		"end_time":   strconv.FormatInt(o.EndTime, 10),
		"math":       o.Math,
	}, nil
}

// %[3]s runs a typed load query against %[2]s.
func (s *%[4]sService) %[3]s(ctx context.Context, opts *%[1]s) (json.RawMessage, error) {
	q, err := opts.query()
	if err != nil {
		return nil, err
	}
	return s.client.Get(ctx, %[2]q, q)
}

`, optsName, ep.Path, methodName, groupTitle)
}

// renderAction emits a semantically named helper (Start<Name>,
// Stop<Name>, Restart<Name>, Sync<Name>, Restore<Name>, Check<Name>)
// for action-style endpoints. The catalog name usually ends with the
// action verb (e.g. "ac-services-start" → "AcServicesStart"); we strip
// the trailing verb before re-prefixing so the method reads naturally
// (Start<AcServices> instead of Start<AcServicesStart>).
func renderAction(buf *bytes.Buffer, groupTitle string, ep V4Endpoint, name string) {
	stripped := stripTrailingVerb(name, ep.Action)
	verbCap := strings.ToUpper(ep.Action[:1]) + ep.Action[1:]
	methodName := verbCap + stripped
	usesBody := ep.Action == "restore" || ep.Action == "check" || ep.Action == "start" || ep.Action == "sync"
	if usesBody {
		fmt.Fprintf(buf, `// %[1]s %[2]s the resource at %[3]s. body is sent as JSON;
// pass nil if the action takes no parameters.
func (s *%[4]sService) %[1]s(ctx context.Context, body any) (json.RawMessage, error) {
	return s.client.Post(ctx, %[3]q, body)
}

`, methodName, verbCap, ep.Path, groupTitle)
		return
	}
	fmt.Fprintf(buf, `// %[1]s %[2]s the resource at %[3]s. No body is required.
func (s *%[4]sService) %[1]s(ctx context.Context) (json.RawMessage, error) {
	return s.client.Post(ctx, %[3]q, map[string]any{})
}

`, methodName, verbCap, ep.Path, groupTitle)
}

// stripTrailingVerb removes a case-insensitive trailing verb from
// name. "AcServicesStart" with verb "start" → "AcServices". "VrrpStart"
// → "Vrrp". If the verb is not at the end, name is returned unchanged.
func stripTrailingVerb(name, verb string) string {
	verbCap := strings.ToUpper(verb[:1]) + verb[1:]
	if strings.HasSuffix(name, verbCap) {
		return name[:len(name)-len(verbCap)]
	}
	if strings.HasSuffix(name, verb) {
		return name[:len(name)-len(verb)]
	}
	return name
}

// pathParamRe matches a path parameter placeholder, e.g. "{tunnel_id}".
var pathParamRe = regexp.MustCompile(`\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// pathParams returns the names of a path's placeholders, in order.
func pathParams(p string) []string {
	var out []string
	for _, m := range pathParamRe.FindAllStringSubmatch(p, -1) {
		out = append(out, m[1])
	}
	return out
}

// goInitialisms are the words Go style capitalizes whole, so a path
// placeholder reads as a Go parameter rather than a mechanical transliteration.
var goInitialisms = map[string]string{
	"id": "ID", "url": "URL", "api": "API", "ip": "IP", "mac": "MAC",
	"uuid": "UUID", "tcp": "TCP", "udp": "UDP", "vlan": "VLAN",
}

// toGoParam converts a path placeholder to a Go parameter name:
// "tunnel_id" -> "tunnelID", "peer_id" -> "peerID".
func toGoParam(s string) string {
	parts := strings.Split(s, "_")
	var b strings.Builder
	for i, p := range parts {
		if p == "" {
			continue
		}
		if i == 0 {
			b.WriteString(strings.ToLower(p))
			continue
		}
		lower := strings.ToLower(p)
		if up, ok := goInitialisms[lower]; ok {
			b.WriteString(up)
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]))
		b.WriteString(p[1:])
	}
	return b.String()
}

// idSegmentParam returns the placeholder that occupies the last path
// segment, if any. When a path ends in a parameter, the item id travels in
// the path rather than in an ?id= query, so the generated helpers bind it
// there.
func idSegmentParam(p string) string {
	params := pathParams(p)
	if len(params) == 0 {
		return ""
	}
	last := p[strings.LastIndex(p, "/")+1:]
	m := pathParamRe.FindStringSubmatch(last)
	if m == nil {
		return ""
	}
	return m[1]
}

// pathExpr renders a Go expression that builds the request path from the
// method's parameters. For a path without placeholders it yields a plain
// string literal, which keeps every existing endpoint generating exactly
// what it did before.
//
// The trailing placeholder binds to the method's "id" parameter — that is
// the value Get/Update/Patch/Delete already take — while the leading ones
// keep their own names.
func pathExpr(path string, params []string, bindLastToID bool) string {
	if len(params) == 0 {
		return fmt.Sprintf("%q", path)
	}
	var b strings.Builder
	rest := path
	for i, name := range params {
		loc := pathParamRe.FindStringIndex(rest)
		if loc == nil {
			break
		}
		b.WriteString(fmt.Sprintf("%q", rest[:loc[0]]))
		variable := toGoParam(name)
		if bindLastToID && i == len(params)-1 {
			variable = "id"
		}
		b.WriteString(fmt.Sprintf(" + strconv.FormatInt(%s, 10) + ", variable))
		rest = rest[loc[1]:]
	}
	b.WriteString(fmt.Sprintf("%q", rest))
	return b.String()
}

// parentParams returns the leading placeholders — the parent resource ids
// that appear before the item segment.
func parentParams(p string) []string {
	params := pathParams(p)
	if id := idSegmentParam(p); id != "" {
		for i, name := range params {
			if name == id {
				return params[:i]
			}
		}
	}
	return params
}

// renderCRUD generates the standard method block (List / Get / Create /
// Update / Patch / Delete) for endpoints that expose a normal CRUD shape.
func renderCRUD(buf *bytes.Buffer, groupTitle string, ep V4Endpoint, name, _ string, verbs map[string]bool) {
	params := pathParams(ep.Path)
	idInPath := idSegmentParam(ep.Path)
	parents := parentParams(ep.Path)
	// parentArgs are the leading path parameters carried by every method.
	var parentArgs []string
	var parentDecl []string
	for _, p := range parents {
		g := toGoParam(p)
		parentArgs = append(parentArgs, g)
		parentDecl = append(parentDecl, g+" int64")
	}
	parentDeclStr := strings.Join(parentDecl, ", ")
	if parentDeclStr != "" {
		parentDeclStr += ", "
	}
	// idDeclStr adds the item identifier for the verbs that target one
	// resource when that identifier lives in the path.
	idDeclStr := parentDeclStr
	if idInPath != "" {
		idDeclStr += "id int64, "
	}

	// The item path binds its trailing placeholder to the method's "id".
	path := pathExpr(ep.Path, params, idInPath != "")
	// A collection List is only generated for a path without a trailing
	// placeholder, and then it is the literal path — nothing to bind.

	orderKey, orderByKey := "order", "order_by"
	sortNote := ""
	if len(params) > 0 {
		sortNote += fmt.Sprintf("\n//\n// This endpoint is a sub-resource of %s; the parent identifier is a\n"+
			"// path parameter rather than a query one.", strings.Join(parents, "_"))
	}
	fmt.Fprintf(buf, `// %[1]sListOptions configures paginated reads of %[1]s.%[5]s
type %[1]sListOptions struct {
	Page     int
	PageSize int
	Filter   string
	// Order is the sort direction: asc | desc.
	Order string
	// OrderBy is the sort field.
	OrderBy string
}

func (o *%[1]sListOptions) query() map[string]string {
	if o == nil {
		return nil
	}
	q := map[string]string{
		"page": strconv.Itoa(o.Page),
		"limit": strconv.Itoa(o.PageSize),
	}
	if o.Filter != "" {
		q["filter"] = o.Filter
	}
	if o.Order != "" {
		q[%[2]q] = o.Order
	}
	if o.OrderBy != "" {
		q[%[3]q] = o.OrderBy
	}
	return q
}

`, name, orderKey, orderByKey, "", sortNote)

	if verbs["GET"] && idInPath == "" {
		fmt.Fprintf(buf, `// List%[1]s lists items at %[2]s.
func (s *%[3]sService) List%[1]s(ctx context.Context, opts *%[1]sListOptions) (json.RawMessage, error) {
	return s.client.Get(ctx, %[2]q, opts.query())
}

`, name, ep.Path, groupTitle)
	}

	// Get one: the id travels in the path when the endpoint is a
	// sub-resource, and in ?id= otherwise.
	if verbs["GET"] {
		if idInPath != "" {
			fmt.Fprintf(buf, `// Get%[1]s fetches one item at %[2]s.
func (s *%[3]sService) Get%[1]s(ctx context.Context, %[6]sid int64) (json.RawMessage, error) {
	return s.client.Get(ctx, %[5]s, nil)
}

`, name, ep.Path, groupTitle, ep.Path, path, parentDeclStr)
		} else {
			fmt.Fprintf(buf, `// Get%[1]s fetches one item at %[2]s.
func (s *%[3]sService) Get%[1]s(ctx context.Context, id int64) (json.RawMessage, error) {
	q := map[string]string{"id": strconv.FormatInt(id, 10)}
	return s.client.Get(ctx, %[2]q, q)
}

`, name, ep.Path, groupTitle)
		}
	}

	if verbs["POST"] {
		fmt.Fprintf(buf, `// Create%[1]s posts a new item to %[2]s. The router returns a rowid on
// success which is propagated to the caller.
func (s *%[3]sService) Create%[1]s(ctx context.Context, %[6]sbody any) (int64, error) {
	raw, err := s.client.Post(ctx, %[5]s, body)
	if err != nil {
		return 0, err
	}
	return extractRowID(raw)
}

`, name, ep.Path, groupTitle, ep.Path, path, parentDeclStr)
	}

	if verbs["PUT"] {
		fmt.Fprintf(buf, `// Update%[1]s replaces the resource at %[2]s.
func (s *%[3]sService) Update%[1]s(ctx context.Context, %[7]sbody any) error {
	_, err := s.client.Put(ctx, %[5]s, body)
	return err
}

`, name, ep.Path, groupTitle, ep.Path, path, parentDeclStr, idDeclStr)
	}

	if verbs["PATCH"] {
		fmt.Fprintf(buf, `// Patch%[1]s partial-updates the resource at %[2]s.
func (s *%[3]sService) Patch%[1]s(ctx context.Context, %[7]sbody any) error {
	_, err := s.client.Patch(ctx, %[5]s, body)
	return err
}

`, name, ep.Path, groupTitle, ep.Path, path, parentDeclStr, idDeclStr)
	}

	if verbs["DELETE"] {
		if idInPath != "" {
			fmt.Fprintf(buf, `// Delete%[1]s removes the resource at %[2]s. The identifier travels
// in the path, not as an ?id= query parameter.
func (s *%[3]sService) Delete%[1]s(ctx context.Context, %[6]sid int64) error {
	_, err := s.client.Delete(ctx, %[5]s, nil)
	return err
}

`, name, ep.Path, groupTitle, ep.Path, path, parentDeclStr)
		} else {
			fmt.Fprintf(buf, `// Delete%[1]s removes the resource at %[2]s. iKuai expects the
// resource identifier as the ?id= query parameter, not in the JSON
// body. Pass Delete%[1]sWithBody to send a custom JSON body instead.
func (s *%[3]sService) Delete%[1]s(ctx context.Context, id int64) error {
	_, err := s.client.Delete(ctx, %[2]q+"?id="+strconv.FormatInt(id, 10), nil)
	return err
}

// Delete%[1]sWithBody removes the resource at %[2]s with a custom
// JSON body. iKuai also accepts a few DELETE requests that carry
// parameters in the body (e.g. /system/backup?srcfile=… which the
// wrapper Delete%[1]s cannot model directly).
func (s *%[3]sService) Delete%[1]sWithBody(ctx context.Context, body any) error {
	_, err := s.client.Delete(ctx, %[2]q, body)
	return err
}

`, name, ep.Path, groupTitle)
		}
	}
}

// trimIDSegment drops a trailing path parameter segment, yielding the
// collection path that a list call targets.
func trimIDSegment(p string) string {
	i := strings.LastIndex(p, "/")
	if i < 0 {
		return p
	}
	return p[:i]
}

var _ = trimIDSegment

// renderRoot writes service/root.go: the APIClient entry point plus
// small helper functions shared by all generated service files.
func renderRoot(file string, groups []string) error {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, `// Code generated by ikuai-api/codegen. DO NOT EDIT.

package service

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	ikuaiapi "github.com/zy84338719/ikuai-api"
)

// V4Endpoint is a re-export of ikuaiapi.V4Endpoint so generated
// service files don't need to know the parent package layout.
type V4Endpoint = ikuaiapi.V4Endpoint

// Endpoints returns the full v4 catalog. It is provided as a helper for
// callers that want to discover the API at runtime.
func Endpoints() []V4Endpoint {
	return ikuaiapi.V4EndpointCatalog
}

// extractRowID parses the synthetic envelope emitted by Create responses
// ({"message": "...", "rowid": N}) and returns the rowid.
func extractRowID(raw json.RawMessage) (int64, error) {
	var env struct {
		RowID json.RawMessage `+"`json:\"rowid\"`"+`
	}
	if err := json.Unmarshal(raw, &env); err != nil || len(env.RowID) == 0 {
		return 0, nil
	}
	if v, err := strconv.ParseInt(string(env.RowID), 10, 64); err == nil {
		return v, nil
	}
	// Fall back to unmarshalling as a JSON number.
	var n json.Number
	if err := json.Unmarshal(env.RowID, &n); err != nil {
		return 0, nil
	}
	return n.Int64()
}

// APIClient is the typed v4 entry point. Construct it with NewAPIClient
// and call any of the per-group service accessors.
type APIClient struct {
	client *ikuaiapi.Client
`)

	for _, g := range groups {
		fmt.Fprintf(buf, "\tsvc%[1]s *%[1]sService\n", toGoName(g))
	}
	fmt.Fprintln(buf, "}")

	fmt.Fprintln(buf, "\n// NewAPIClient wires every group service to the supplied client.")
	fmt.Fprintln(buf, "func NewAPIClient(client *ikuaiapi.Client) *APIClient {")
	fmt.Fprintln(buf, "\treturn &APIClient{client: client,")
	for _, g := range groups {
		fmt.Fprintf(buf, "\t\tsvc%[1]s: &%[1]sService{client: client},\n", toGoName(g))
	}
	fmt.Fprintln(buf, "\t}")
	fmt.Fprintln(buf, "}")

	fmt.Fprintln(buf, "\n// Client returns the underlying ikuaiapi.Client. Use it for raw,")
	fmt.Fprintln(buf, "// catalog-driven calls when a typed helper does not exist yet.")
	fmt.Fprintln(buf, "func (a *APIClient) Client() *ikuaiapi.Client { return a.client }")

	for _, g := range groups {
		name := toGoName(g)
		fmt.Fprintf(buf, `
func (a *APIClient) %[1]s() *%[1]sService { return a.svc%[1]s }
`, name)
	}

	fmt.Fprintln(buf, "\n// Call is the lowest-level escape hatch. It resolves a (group, name)")
	fmt.Fprintln(buf, "// pair from the catalog and dispatches the request using method,")
	fmt.Fprintln(buf, "// path, optional query params and optional JSON body. group is an")
	fmt.Fprintln(buf, "// optional sanity check; pass \"\" to look up by name only.")
	fmt.Fprintln(buf, "func (a *APIClient) Call(ctx context.Context, group, name, method string, body any, params map[string]string) (json.RawMessage, error) {")
	fmt.Fprintln(buf, "\tep, ok := ikuaiapi.V4EndpointByName(name)")
	fmt.Fprintln(buf, "\tif !ok || (group != \"\" && ep.Group != group) {")
	fmt.Fprintln(buf, "\t\t// Some names live under more than one group (e.g. \"system\")")
	fmt.Fprintln(buf, "\t\t// so fall back to the explicit (group, name) lookup.")
	fmt.Fprintln(buf, "\t\tif group != \"\" {")
	fmt.Fprintln(buf, "\t\t\tif ep2, ok2 := ikuaiapi.V4EndpointByGroupName(group, name); ok2 {")
	fmt.Fprintln(buf, "\t\t\t\tep = ep2")
	fmt.Fprintln(buf, "\t\t\t} else {")
	fmt.Fprintln(buf, "\t\t\t\treturn nil, ikuaiapi.NewError(\"endpoint not found: \" + group + \"/\" + name)")
	fmt.Fprintln(buf, "\t\t\t}")
	fmt.Fprintln(buf, "\t\t} else {")
	fmt.Fprintln(buf, "\t\t\treturn nil, ikuaiapi.NewError(\"endpoint not found: \" + name)")
	fmt.Fprintln(buf, "\t\t}")
	fmt.Fprintln(buf, "\t}")
	fmt.Fprintln(buf, "\tswitch method {")
	fmt.Fprintln(buf, "\tcase \"GET\":")
	fmt.Fprintln(buf, "\t\treturn a.client.Get(ctx, ep.Path, params)")
	fmt.Fprintln(buf, "\tcase \"POST\":")
	fmt.Fprintln(buf, "\t\treturn a.client.Post(ctx, ep.Path, body)")
	fmt.Fprintln(buf, "\tcase \"PUT\":")
	fmt.Fprintln(buf, "\t\treturn a.client.Put(ctx, ep.Path, body)")
	fmt.Fprintln(buf, "\tcase \"PATCH\":")
	fmt.Fprintln(buf, "\t\treturn a.client.Patch(ctx, ep.Path, body)")
	fmt.Fprintln(buf, "\tcase \"DELETE\":")
	fmt.Fprintln(buf, "\t\tdelURL := ep.Path")
	fmt.Fprintln(buf, "\t\tif len(params) > 0 {")
	fmt.Fprintln(buf, "\t\t\tsep := \"?\"")
	fmt.Fprintln(buf, "\t\t\tif strings.Contains(ep.Path, \"?\") {")
	fmt.Fprintln(buf, "\t\t\t\tsep = \"&\"")
	fmt.Fprintln(buf, "\t\t\t}")
	fmt.Fprintln(buf, "\t\t\tdelURL = ep.Path + sep + a.client.FormatQuery(params)")
	fmt.Fprintln(buf, "\t\t}")
	fmt.Fprintln(buf, "\t\treturn a.client.Delete(ctx, delURL, body)")
	fmt.Fprintln(buf, "\tdefault:")
	fmt.Fprintln(buf, "\t\treturn nil, ikuaiapi.NewError(\"unsupported method: \" + method)")
	fmt.Fprintln(buf, "\t}")
	fmt.Fprintln(buf, "}")

	// PageInfo is a small helper for parsing pagination metadata that
	// iKuai echoes back on list endpoints.
	fmt.Fprintln(buf, "\n// PageInfo mirrors the pagination metadata iKuai returns with list")
	fmt.Fprintln(buf, "// endpoints. It is exposed so callers can drive follow-up pages")
	fmt.Fprintln(buf, "// without having to declare a local struct.")
	fmt.Fprintln(buf, "//")
	fmt.Fprintln(buf, "// Firmware revisions disagree on the page-size key they echo back:")
	fmt.Fprintln(buf, "// some return \"page_size\", others \"limit\" (the key the v4")
	fmt.Fprintln(buf, "// request contract actually accepts). Both are captured; use")
	fmt.Fprintln(buf, "// Size rather than reading the fields directly.")
	fmt.Fprintln(buf, "type PageInfo struct {")
	fmt.Fprintln(buf, "\tPage     int `json:\"page\"`")
	fmt.Fprintln(buf, "\tPageSize int `json:\"page_size\"`")
	fmt.Fprintln(buf, "\tLimit    int `json:\"limit\"`")
	fmt.Fprintln(buf, "\tTotal    int `json:\"total\"`")
	fmt.Fprintln(buf, "}")
	fmt.Fprintln(buf, "\n// Size returns the echoed page size, preferring page_size and falling")
	fmt.Fprintln(buf, "// back to limit. It returns 0 when the router sent neither.")
	fmt.Fprintln(buf, "func (p PageInfo) Size() int {")
	fmt.Fprintln(buf, "\tif p.PageSize > 0 {")
	fmt.Fprintln(buf, "\t\treturn p.PageSize")
	fmt.Fprintln(buf, "\t}")
	fmt.Fprintln(buf, "\treturn p.Limit")
	fmt.Fprintln(buf, "}")

	// FormatURLValues is exported for callers building query strings by
	// hand. It returns nil for an empty map.
	fmt.Fprintln(buf, "\n// FormatURLValues is exposed for callers that build query strings")
	fmt.Fprintln(buf, "// by hand. It returns an empty string for a nil/empty map.")
	fmt.Fprintln(buf, "func FormatURLValues(q map[string]string) string {")
	fmt.Fprintln(buf, "\tif len(q) == 0 {")
	fmt.Fprintln(buf, "\t\treturn \"\"")
	fmt.Fprintln(buf, "\t}")
	fmt.Fprintln(buf, "\treturn urlEncode(q)")
	fmt.Fprintln(buf, "}")

	fmt.Fprintln(buf, "\nfunc urlEncode(q map[string]string) string {")
	fmt.Fprintln(buf, "\tuv := make(url.Values)")
	fmt.Fprintln(buf, "\tfor k, v := range q {")
	fmt.Fprintln(buf, "\t\tuv.Set(k, v)")
	fmt.Fprintln(buf, "\t}")
	fmt.Fprintln(buf, "\treturn uv.Encode()")
	fmt.Fprintln(buf, "}")

	return os.WriteFile(file, buf.Bytes(), 0o644)
}
