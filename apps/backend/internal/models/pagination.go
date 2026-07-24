package models

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

type PaginationRequest struct {
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	Search     string `json:"search"`
	Sort       string `json:"sort"`
	Order      string `json:"order"`
	Status     string `json:"status"`
	Severity   string `json:"severity"`
	ProjectID  uint   `json:"project_id"`
	IncidentID uint   `json:"incident_id"`
	Action     string `json:"action"`
	EntityType string `json:"entity_type"`
}

func (p *PaginationRequest) Normalize() {
	if p == nil {
		return
	}
	if p.Page == 0 {
		p.Page = DefaultPage
	}
	if p.Limit == 0 {
		p.Limit = DefaultLimit
	}
	p.Search = strings.TrimSpace(p.Search)
	p.Sort = strings.TrimSpace(strings.ToLower(p.Sort))
	p.Order = strings.TrimSpace(strings.ToLower(p.Order))
	if p.Order == "" {
		p.Order = "desc"
	}
	if p.Sort == "" {
		p.Sort = "created_at"
	}
	p.Status = strings.TrimSpace(p.Status)
	p.Severity = strings.TrimSpace(p.Severity)
	p.Action = strings.TrimSpace(strings.ToUpper(p.Action))
	p.EntityType = strings.TrimSpace(strings.ToLower(p.EntityType))
}

// Validate normalizes defaults and rejects values that could cause inefficient
// queries or unsafe dynamic sorting. Callers must supply the sort fields that
// are valid for their resource.
func (p *PaginationRequest) Validate(allowedSortFields ...string) error {
	if p == nil {
		return fmt.Errorf("pagination parameters are required")
	}

	p.Normalize()

	if p.Page < DefaultPage {
		return fmt.Errorf("page must be at least %d", DefaultPage)
	}
	if p.Limit < 1 || p.Limit > MaxLimit {
		return fmt.Errorf("limit must be between 1 and %d", MaxLimit)
	}
	if p.Order != "asc" && p.Order != "desc" {
		return fmt.Errorf("order must be either asc or desc")
	}
	if !contains(allowedSortFields, p.Sort) {
		return fmt.Errorf("sort must be one of: %s", strings.Join(allowedSortFields, ", "))
	}

	return nil
}

// ParsePagination reads query parameters while preserving the distinction
// between an omitted value (which receives a default) and an invalid supplied
// value (which is rejected).
func ParsePagination(values url.Values, allowedSortFields ...string) (*PaginationRequest, error) {
	req := &PaginationRequest{
		Page:       DefaultPage,
		Limit:      DefaultLimit,
		Search:     values.Get("search"),
		Sort:       values.Get("sort"),
		Order:      values.Get("order"),
		Status:     values.Get("status"),
		Severity:   values.Get("severity"),
		Action:     values.Get("action"),
		EntityType: values.Get("entityType"),
	}

	var err error
	if values.Has("page") {
		req.Page, err = parsePositiveInt(values.Get("page"), "page", 0)
		if err != nil {
			return nil, err
		}
	}
	if values.Has("limit") {
		req.Limit, err = parsePositiveInt(values.Get("limit"), "limit", MaxLimit)
		if err != nil {
			return nil, err
		}
	}

	return req, req.Validate(allowedSortFields...)
}

func parsePositiveInt(value, name string, maximum int) (int, error) {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf("%s must be at least 1", name)
	}
	if maximum > 0 && parsed > maximum {
		return 0, fmt.Errorf("%s must not exceed %d", name, maximum)
	}
	return parsed, nil
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

type PaginationResponse struct {
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	Total      int64       `json:"total"`
	TotalPages int         `json:"totalPages"`
	Items      interface{} `json:"items"`
}
