package pagination

import (
	"encoding/json"
)

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

// NewPaginationView creates a new PaginationView instance
func NewPaginationView[T any](page, limit, total int, results []T) PaginationView[T] {
	return PaginationView[T]{
		Results: results,
		Page:    page,
		Limit:   limit,
		Total:   total,
	}
}
