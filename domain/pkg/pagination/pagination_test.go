package pagination

import (
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestPagination_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		expected    Pagination
		expectError bool
	}{
		{
			name:     "With Limit",
			jsonData: `{"page": 2, "limit": 10, "search_query": "test"}`,
			expected: Pagination{
				Page:        2,
				Limit:       10,
				SearchQuery: "test",
			},
			expectError: false,
		},
		{
			name:     "Without Limit",
			jsonData: `{"page": 3, "search_query": "query"}`,
			expected: Pagination{
				Page:        3,
				Limit:       DefaultLimit,
				SearchQuery: "query",
			},
			expectError: false,
		},
		{
			name:        "Invalid JSON",
			jsonData:    `{"page": "invalid"}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Pagination
			err := json.Unmarshal([]byte(tt.jsonData), &p)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if p.Page != tt.expected.Page {
				t.Errorf("Expected Page %d, got %d", tt.expected.Page, p.Page)
			}
			if p.Limit != tt.expected.Limit {
				t.Errorf("Expected Limit %d, got %d", tt.expected.Limit, p.Limit)
			}
			if p.SearchQuery != tt.expected.SearchQuery {
				t.Errorf("Expected SearchQuery %s, got %s", tt.expected.SearchQuery, p.SearchQuery)
			}
		})
	}
}

func TestPagination_GetOffset(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		limit    int
		expected int
	}{
		{
			name:     "Page 1 Limit 10",
			page:     1,
			limit:    10,
			expected: 0,
		},
		{
			name:     "Page 2 Limit 10",
			page:     2,
			limit:    10,
			expected: 10,
		},
		{
			name:     "Page 3 Limit 5",
			page:     3,
			limit:    5,
			expected: 10,
		},
		{
			name:     "Page 0 Limit 10",
			page:     0,
			limit:    10,
			expected: -10, // edge case - negative page
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Pagination{
				Page:  tt.page,
				Limit: tt.limit,
			}
			if got := p.GetOffset(); got != tt.expected {
				t.Errorf("GetOffset() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAutoPaginate(t *testing.T) {
	type testItem struct {
		ID   int
		Name string
	}

	items := []testItem{
		{ID: 1, Name: "Item 1"},
		{ID: 2, Name: "Item 2"},
		{ID: 3, Name: "Item 3"},
		{ID: 4, Name: "Item 4"},
		{ID: 5, Name: "Item 5"},
		{ID: 6, Name: "Item 6"},
		{ID: 7, Name: "Item 7"},
		{ID: 8, Name: "Item 8"},
		{ID: 9, Name: "Item 9"},
		{ID: 10, Name: "Item 10"},
	}

	tests := []struct {
		name     string
		page     int
		limit    int
		content  []testItem
		expected PaginationView[testItem]
	}{
		{
			name:    "Empty Content",
			page:    1,
			limit:   5,
			content: []testItem{},
			expected: PaginationView[testItem]{
				Results: []testItem{},
				Page:    1,
				Limit:   5,
				Total:   0,
			},
		},
		{
			name:    "First Page Partial Content",
			page:    1,
			limit:   5,
			content: items[:3], // Only 3 items
			expected: PaginationView[testItem]{
				Results: items[:3],
				Page:    1,
				Limit:   5,
				Total:   3,
			},
		},
		{
			name:    "First Page Full Content",
			page:    1,
			limit:   5,
			content: items,
			expected: PaginationView[testItem]{
				Results: items[:5],
				Page:    1,
				Limit:   5,
				Total:   10,
			},
		},
		{
			name:    "Second Page",
			page:    2,
			limit:   5,
			content: items,
			expected: PaginationView[testItem]{
				Results: items[5:10],
				Page:    2,
				Limit:   5,
				Total:   10,
			},
		},
		{
			name:    "Page Beyond Content",
			page:    3,
			limit:   5,
			content: items,
			expected: PaginationView[testItem]{
				Results: []testItem{},
				Page:    3,
				Limit:   5,
				Total:   10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Pagination{
				Page:  tt.page,
				Limit: tt.limit,
			}
			result := AutoPaginate(p, tt.content)

			if !reflect.DeepEqual(result.Results, tt.expected.Results) {
				t.Errorf("Results = %v, want %v", result.Results, tt.expected.Results)
			}
			if result.Page != tt.expected.Page {
				t.Errorf("Page = %v, want %v", result.Page, tt.expected.Page)
			}
			if result.Limit != tt.expected.Limit {
				t.Errorf("Limit = %v, want %v", result.Limit, tt.expected.Limit)
			}
			if result.Total != tt.expected.Total {
				t.Errorf("Total = %v, want %v", result.Total, tt.expected.Total)
			}
		})
	}
}

func TestFormatWith(t *testing.T) {
	type testItem struct {
		ID   int
		Name string
	}

	items := []testItem{
		{ID: 1, Name: "Item 1"},
		{ID: 2, Name: "Item 2"},
		{ID: 3, Name: "Item 3"},
	}

	p := Pagination{
		Page:  2,
		Limit: 10,
	}

	result := FormatWith(p, 25, items)

	if len(result.Results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(result.Results))
	}
	if result.Page != 2 {
		t.Errorf("Expected Page 2, got %d", result.Page)
	}
	if result.Limit != 10 {
		t.Errorf("Expected Limit 10, got %d", result.Limit)
	}
	if result.Total != 25 {
		t.Errorf("Expected Total 25, got %d", result.Total)
	}
}

func TestNewPaginationView(t *testing.T) {
	type testItem struct {
		ID   int
		Name string
	}

	items := []testItem{
		{ID: 1, Name: "Item 1"},
		{ID: 2, Name: "Item 2"},
	}

	result := NewPaginationView(3, 15, 30, items)

	if len(result.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result.Results))
	}
	if result.Page != 3 {
		t.Errorf("Expected Page 3, got %d", result.Page)
	}
	if result.Limit != 15 {
		t.Errorf("Expected Limit 15, got %d", result.Limit)
	}
	if result.Total != 30 {
		t.Errorf("Expected Total 30, got %d", result.Total)
	}
}

func TestGetPaginationFromReq(t *testing.T) {
	tests := []struct {
		name        string
		queryParams map[string]string
		expected    *Pagination
		expectError bool
	}{
		{
			name:        "Default Values",
			queryParams: map[string]string{},
			expected: &Pagination{
				Page:        1,
				Limit:       20,
				SearchQuery: "",
				Queries: map[string]string{
					"sort": "desc",
				},
			},
			expectError: false,
		},
		{
			name: "Valid Parameters",
			queryParams: map[string]string{
				"page":         "2",
				"limit":        "15",
				"search_query": "test query",
				"sort":         "asc",
			},
			expected: &Pagination{
				Page:        2,
				Limit:       15,
				SearchQuery: "test query",
				Queries: map[string]string{
					"sort": "asc",
				},
			},
			expectError: false,
		},
		{
			name: "Invalid Page",
			queryParams: map[string]string{
				"page": "invalid",
			},
			expectError: true,
		},
		{
			name: "Invalid Limit",
			queryParams: map[string]string{
				"limit": "invalid",
			},
			expectError: true,
		},
		{
			name: "Invalid Sort",
			queryParams: map[string]string{
				"sort": "invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with query parameters
			req := &http.Request{
				URL: &url.URL{
					RawQuery: buildQueryString(tt.queryParams),
				},
			}

			p, err := GetPaginationFromReq(req)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if p.Page != tt.expected.Page {
				t.Errorf("Expected Page %d, got %d", tt.expected.Page, p.Page)
			}
			if p.Limit != tt.expected.Limit {
				t.Errorf("Expected Limit %d, got %d", tt.expected.Limit, p.Limit)
			}
			if p.SearchQuery != tt.expected.SearchQuery {
				t.Errorf("Expected SearchQuery %s, got %s", tt.expected.SearchQuery, p.SearchQuery)
			}
			if !reflect.DeepEqual(p.Queries, tt.expected.Queries) {
				t.Errorf("Expected Queries %v, got %v", tt.expected.Queries, p.Queries)
			}
		})
	}
}

// Helper function to build query string
func buildQueryString(params map[string]string) string {
	values := url.Values{}
	for key, value := range params {
		values.Add(key, value)
	}
	return values.Encode()
}
