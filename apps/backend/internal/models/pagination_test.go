package models

import (
	"net/url"
	"testing"
)

func TestParsePaginationDefaults(t *testing.T) {
	req, err := ParsePagination(url.Values{}, "name", "created_at")
	if err != nil {
		t.Fatalf("ParsePagination() error = %v", err)
	}
	if req.Page != DefaultPage || req.Limit != DefaultLimit || req.Sort != "created_at" || req.Order != "desc" {
		t.Fatalf("unexpected defaults: %#v", req)
	}
}

func TestParsePaginationRejectsUnsafeValues(t *testing.T) {
	tests := []struct {
		name   string
		values url.Values
	}{
		{name: "page below one", values: url.Values{"page": {"0"}}},
		{name: "limit above maximum", values: url.Values{"limit": {"101"}}},
		{name: "invalid sort field", values: url.Values{"sort": {"name; DROP TABLE projects"}}},
		{name: "invalid sort order", values: url.Values{"order": {"sideways"}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := ParsePagination(test.values, "name", "created_at"); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestParsePaginationNormalizesWhitelistedSort(t *testing.T) {
	req, err := ParsePagination(url.Values{
		"page":  {"2"},
		"limit": {"50"},
		"sort":  {" NAME "},
		"order": {"ASC"},
	}, "name", "created_at")
	if err != nil {
		t.Fatalf("ParsePagination() error = %v", err)
	}
	if req.Page != 2 || req.Limit != 50 || req.Sort != "name" || req.Order != "asc" {
		t.Fatalf("unexpected parsed request: %#v", req)
	}
}
