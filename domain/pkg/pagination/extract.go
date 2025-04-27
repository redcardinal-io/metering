package pagination

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// DefaultPage is the default page number if not specified
const DefaultPage = 1

// DefaultPerPage is the default number of items per page if not specified
const DefaultPerPage = 10

// Reserved query parameter names that are used specifically for pagination
var ReservedQueryParams = map[string]bool{
	"page":         true,
	"limit":        true,
	"search_query": true,
	"sort":         true,
}

// ExtractPaginationFromContext extracts pagination parameters from a Fiber context
func ExtractPaginationFromContext(ctx *fiber.Ctx) Pagination {
	// Parse pagination parameters from query string
	page, err := strconv.Atoi(ctx.Query("page", strconv.Itoa(DefaultPage)))
	if err != nil || page < 1 {
		page = DefaultPage
	}

	limit, err := strconv.Atoi(ctx.Query("limit", strconv.Itoa(DefaultPerPage)))
	if err != nil || limit < 1 {
		limit = DefaultPerPage
	}

	searchQuery := ctx.Query("search_query", "")
	sort := ctx.Query("sort", "desc")

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
