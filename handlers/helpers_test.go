package handlers

import (
	"net/http/httptest"
	"testing"
)

func TestParseFilters(t *testing.T) {
	r := httptest.NewRequest("GET", "/partials/posts?category_ids=abc&q=go&limit=12&offset=24", nil)
	filters := parseFilters(r)
	if filters.CategoryIDs != "abc" || filters.Search != "go" || filters.Limit != 12 || filters.Offset != 24 {
		t.Fatalf("unexpected filters: %#v", filters)
	}
}

func TestParseFiltersUsesSafeDefaults(t *testing.T) {
	r := httptest.NewRequest("GET", "/partials/posts?limit=1000&offset=-1", nil)
	filters := parseFilters(r)
	if filters.Limit != 20 || filters.Offset != 0 {
		t.Fatalf("unexpected defaults: %#v", filters)
	}
}

func TestParseUUIDsSupportsRepeatedAndCommaSeparatedValues(t *testing.T) {
	values := []string{"5919fccf-0a1f-4f98-bfc5-f1e2bfd9cdb9,4b749076-12a3-40df-8a1c-9cc1e28a3ed3", "5919fccf-0a1f-4f98-bfc5-f1e2bfd9cdb9"}
	ids, err := parseUUIDs(values)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 unique IDs, got %d", len(ids))
	}
}
