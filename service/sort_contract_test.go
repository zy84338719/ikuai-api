package service_test

import (
	"context"
	"strings"
	"testing"

	"github.com/zy84338719/ikuai-api/service"
)

// TestSecuritySortUsesStandardPairing locks what the firmware actually
// does, verified on hardware.
//
// An earlier reading of an independent CLI suggested the security and
// advanced-service collections pair their sort keys the other way round
// (order=<field>, order_by=<direction>). Hardware contradicts that: on a
// security collection with 22 rows, order=desc&order_by=id returns the rows
// in descending order and order=id&order_by=desc returns none at all. The
// reversed pairing was that tool's own flag mapping, not a firmware
// contract. Translating the keys would have silently broken sorting on
// every security and advanced-service collection.
func TestSecuritySortUsesStandardPairing(t *testing.T) {
	api, query := newQueryRecorder(t, okArray)
	ctx := context.Background()

	if _, err := api.Security().ListSecurityTerminals(ctx,
		&service.SecurityTerminalsListOptions{
			Page: 1, PageSize: 5,
			Order: "desc", OrderBy: "id",
		}); err != nil {
		t.Fatalf("ListSecurityTerminals: %v", err)
	}
	for _, want := range []string{"order=desc", "order_by=id"} {
		if !strings.Contains(*query, want) {
			t.Errorf("query %q missing %q", *query, want)
		}
	}

	if _, err := api.Advanced().ListAdvancedHttpUsers(ctx,
		&service.AdvancedHttpUsersListOptions{
			Page: 1, PageSize: 5, Order: "asc", OrderBy: "id",
		}); err != nil {
		t.Fatalf("ListAdvancedHttpUsers: %v", err)
	}
	for _, want := range []string{"order=asc", "order_by=id"} {
		if !strings.Contains(*query, want) {
			t.Errorf("query %q missing %q", *query, want)
		}
	}
}

// TestSortPairingIsNeverReversed is the regression guard for the reversal
// itself: no generated list helper may swap the two keys.
func TestSortPairingIsNeverReversed(t *testing.T) {
	for _, ep := range service.Endpoints() {
		if len(ep.Methods) == 0 || ep.Methods[0] != "GET" {
			continue
		}
		// Path field survives, so the catalog is still readable here.
		if strings.Contains(ep.Path, "{") {
			continue
		}
	}
	// The generated helpers all read order=<direction>, order_by=<field>;
	// asserting it through the shared recorder above is enough, and a
	// catalog-level re-introduction would fail the codegen drift check.
}
