package pagination

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

// GetPaginationFromReq extracts pagination info from an HTTP request
func GetPaginationFromReq(r *http.Request) (*Pagination, error) {
	query := r.URL.Query()

	page := 1
	if p := query.Get("page"); p != "" {
		var err error
		page, err = strconv.Atoi(p)
		if err != nil || page < 1 {
			return nil, errors.New("invalid page parameter")
		}
	}

	limit := DefaultLimit
	if l := query.Get("limit"); l != "" {
		var err error
		limit, err = strconv.Atoi(l)
		if err != nil || limit < 1 || limit > 100 {
			return nil, errors.New("invalid limit parameter")
		}
	}

	searchQuery := query.Get("search_query")
	sort := query.Get("sort")
	if sort == "" {
		sort = "desc"
	}

	// Optional strict sort validation
	if sort != "asc" && sort != "desc" {
		return nil, errors.New("invalid sort parameter")
	}

	return &Pagination{
		Page:        page,
		Limit:       limit,
		SearchQuery: searchQuery,
		Sort:        sort,
		Queries: map[string]string{
			"sort": sort,
		},
	}, nil
}

// Pagination represents pagination parameters and search options
type Pagination struct {
	Page        int               `json:"page"`
	Limit       int               `json:"limit"`
	SearchQuery string            `json:"search_query,omitempty"`
	Queries     map[string]string `json:"queries,omitempty"`
	Sort        string            `json:"sort,omitempty"`
}

// PaginationView represents a paginated view of results
type PaginationView[T any] struct {
	Results []T `json:"results"`
	Page    int `json:"page"`
	Limit   int `json:"limit"`
	Total   int `json:"total"`
}

// UnmarshalJSON implements custom JSON unmarshaling for Pagination
func (p *Pagination) UnmarshalJSON(data []byte) error {
	type Alias Pagination
	aux := &struct {
		*Alias
		Limit *int `json:"limit,omitempty"`
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Limit == nil {
		p.Limit = DefaultLimit
	} else {
		p.Limit = *aux.Limit
	}

	return nil
}

// GetOffset calculates the offset based on page and limit
func (p Pagination) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

// AutoPaginate automatically paginates the given content
func AutoPaginate[T any](p Pagination, content []T) PaginationView[T] {
	start := p.GetOffset()
	end := start + p.Limit

	if start > len(content) {
		start = len(content)
	}
	if end > len(content) {
		end = len(content)
	}

	results := content[start:end]
	return FormatWith(p, len(content), results)
}

// FormatWith formats the given results into a PaginationView
func FormatWith[T any](p Pagination, total int, results []T) PaginationView[T] {
	return PaginationView[T]{
		Results: results,
		Page:    p.Page,
		Limit:   p.Limit,
		Total:   total,
	}
}

// NewPaginationView returns a PaginationView containing the specified page, limit, total count, and results.
func NewPaginationView[T any](page, limit, total int, results []T) PaginationView[T] {
	return PaginationView[T]{
		Results: results,
		Page:    page,
		Limit:   limit,
		Total:   total,
	}
}
