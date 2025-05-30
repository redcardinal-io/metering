package pagination

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// DefaultPage is the default page number if not specified
const DefaultPage = 1

// DefaultLimit is the default number of items per page if not specified
const DefaultLimit = 20

// Reserved query parameter names that are used specifically for pagination
var ReservedQueryParams = map[string]bool{
	"page":         true,
	"limit":        true,
	"search_query": true,
	"sort":         true,
}

// ExtractPaginationFromContext retrieves pagination and filtering parameters from the Fiber context's query string.
// 
// It parses "page" and "limit" as integers, applying default values and validation. The "search_query" and "sort" parameters are extracted as strings, with "sort" restricted to "asc" or "desc" (defaulting to "desc"). All other non-reserved query parameters are collected as filter queries. Returns a Pagination struct containing the extracted values.
func ExtractPaginationFromContext(ctx *fiber.Ctx) Pagination {
	// Parse pagination parameters from query string
	page, err := strconv.Atoi(ctx.Query("page", strconv.Itoa(DefaultPage)))
	if err != nil || page < 1 {
		page = DefaultPage
	}

	limit, err := strconv.Atoi(ctx.Query("limit", strconv.Itoa(DefaultLimit)))
	if err != nil || limit < 1 || limit > 100 {
		limit = DefaultLimit
	}

	searchQuery := ctx.Query("search_query", "")
	sort := ctx.Query("sort", "desc")
	if sort != "asc" && sort != "desc" {
		sort = "desc"
	}

	// Extract filter queries (non-pagination parameters)
	allQueries := ctx.Queries()
	filterQueries := make(map[string]string)

	for key, value := range allQueries {
		// Only include non-reserved parameters in the filter queries
		if !ReservedQueryParams[key] {
			filterQueries[key] = value
		}
	}

	// Create pagination input
	return Pagination{
		Page:        page,
		Limit:       limit,
		SearchQuery: searchQuery,
		Queries:     filterQueries,
		Sort:        sort,
	}
}
