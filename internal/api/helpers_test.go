package api

import (
	"net/http/httptest"
	"testing"
)

func TestParsePagination(t *testing.T) {
	request := httptest.NewRequest("GET", "/v1/posts?limit=25&offset=10", nil)
	limit, offset, err := parsePagination(request)
	if err != nil {
		t.Fatal(err)
	}
	if limit != 25 || offset != 10 {
		t.Fatalf("got limit=%d offset=%d", limit, offset)
	}
}

func TestParsePaginationRejectsInvalidValues(t *testing.T) {
	request := httptest.NewRequest("GET", "/v1/posts?limit=101&offset=-1", nil)
	if _, _, err := parsePagination(request); err == nil {
		t.Fatal("expected invalid pagination error")
	}
}

func TestNormalizeCategoryFilter(t *testing.T) {
	input := "550e8400-e29b-41d4-a716-446655440000, 123e4567-e89b-12d3-a456-426614174000"
	got, err := normalizeCategoryFilter(input)
	if err != nil {
		t.Fatal(err)
	}
	want := "550e8400-e29b-41d4-a716-446655440000,123e4567-e89b-12d3-a456-426614174000"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestNormalizeCategoryFilterRejectsInvalidUUID(t *testing.T) {
	if _, err := normalizeCategoryFilter("not-a-uuid"); err == nil {
		t.Fatal("expected invalid UUID error")
	}
}
