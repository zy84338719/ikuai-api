package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseCatalogSanity ensures the regex picks up every entry in the
// real catalog. If the catalog grew past 100 endpoints and the regex
// dropped one, this would catch the regression.
func TestParseCatalogSanity(t *testing.T) {
	eps, err := parseCatalog(locateCatalog())
	if err != nil {
		t.Fatal(err)
	}
	if len(eps) < 100 {
		t.Errorf("expected at least 100 endpoints, got %d", len(eps))
	}
	seen := make(map[string]int)
	for _, ep := range eps {
		seen[ep.Group]++
		if !strings.HasPrefix(ep.Path, "/") {
			t.Errorf("path %q does not start with /", ep.Path)
		}
		if len(ep.Methods) == 0 {
			t.Errorf("%s/%s has no methods", ep.Group, ep.Name)
		}
	}
	for g, n := range seen {
		if n < 2 {
			t.Errorf("group %q has only %d endpoint(s)", g, n)
		}
	}
}

// TestToGoName locks the CamelCase conversion. A change here would
// silently rename every generated method.
func TestToGoName(t *testing.T) {
	cases := []struct{ in, out string }{
		{"dhcp-services", "DhcpServices"},
		{"app-protocols-history-load", "AppProtocolsHistoryLoad"},
		{"vlan", "Vlan"},
		{"web-services", "WebServices"},
		{"ac-services-start", "AcServicesStart"},
	}
	for _, c := range cases {
		if got := toGoName(c.in); got != c.out {
			t.Errorf("toGoName(%q) = %q, want %q", c.in, got, c.out)
		}
	}
}

func locateCatalog() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		p := filepath.Join(dir, "v4_catalog.go")
		if _, err := os.Stat(p); err == nil {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
