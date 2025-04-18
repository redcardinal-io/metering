package clickhouse

import (
	"errors"
	"fmt"
	"strings"

	"database/sql"

	domainerrors "github.com/redcardinal-io/metering/domain/errors"
)

// ClickHouse error message patterns
const (
	// Connection errors
	ChErrConnectionRefused = "connection refused"
	ChErrConnectionReset   = "connection reset"
	ChErrBrokenPipe        = "broken pipe"

	// Authentication errors
	ChErrAuthFailed   = "authentication failed"
	ChErrAccessDenied = "access denied"

	// Schema errors
	ChErrNoSuchTable      = "no such table"
	ChErrTableNotFound    = "table does not exist"
	ChErrNoSuchColumn     = "no such column"
	ChErrUnknownColumn    = "unknown column"
	ChErrDatabaseNotFound = "database does not exist"

	// Query errors
	ChErrSyntaxError = "syntax error"
	ChErrTypeError   = "type mismatch"
)

// MapError translates ClickHouse errors (via sqlx) to domain errors
func MapError(err error, op string) error {
	if err == nil {
		return nil
	}

	// Handle sql.ErrNoRows specifically
	if errors.Is(err, sql.ErrNoRows) {
		return domainerrors.New(
			err,
			domainerrors.ENOTFOUND,
			"Resource not found",
			domainerrors.WithOperation(op),
		)
	}

	// Get error message in lowercase for consistent string matching
	errMsg := strings.ToLower(err.Error())

	// Parse ClickHouse error codes - ClickHouse often includes error codes like "Code: 60"
	errCode := extractErrorCode(errMsg)

	// First check for specific error codes if they exist
	if errCode > 0 {
		switch errCode {
		case 516, 497: // Authentication failed, Access denied
			return domainerrors.New(
				err,
				domainerrors.EUNAUTHORIZED,
				"Authentication failed with ClickHouse",
				domainerrors.WithOperation(op),
			)

		case 81: // Database not found
			return domainerrors.New(
				err,
				domainerrors.ENOTFOUND,
				"Database does not exist",
				domainerrors.WithOperation(op),
			)

		case 60: // Table not found
			return domainerrors.New(
				err,
				domainerrors.ENOTFOUND,
				"Table does not exist",
				domainerrors.WithOperation(op),
			)

		case 47: // Column not found
			return domainerrors.New(
				err,
				domainerrors.EINVALID,
				"Column does not exist",
				domainerrors.WithOperation(op),
			)

		case 57: // Table already exists
			return domainerrors.New(
				err,
				domainerrors.ECONFLICT,
				"Table already exists",
				domainerrors.WithOperation(op),
			)

		case 62: // Syntax error
			return domainerrors.New(
				err,
				domainerrors.EINVALID,
				"SQL syntax error",
				domainerrors.WithOperation(op),
			)

		case 160: // Quota exceeded
			return domainerrors.New(
				err,
				domainerrors.EUNPROCESSABLE,
				"Resource quota exceeded",
				domainerrors.WithOperation(op),
			)

		case 159: // Timeout
			return domainerrors.New(
				err,
				domainerrors.ETIMEOUT,
				"Query execution timed out",
				domainerrors.WithOperation(op),
			)

		case 241: // Memory limit exceeded
			return domainerrors.New(
				err,
				domainerrors.EUNPROCESSABLE,
				"Memory limit exceeded",
				domainerrors.WithOperation(op),
			)
		}
	}

	// Otherwise, check for common error patterns in the message

	// Connection errors
	if strings.Contains(errMsg, ChErrConnectionRefused) ||
		strings.Contains(errMsg, ChErrConnectionReset) ||
		strings.Contains(errMsg, ChErrBrokenPipe) {
		return domainerrors.New(
			err,
			domainerrors.EUNAVAILABLE,
			"ClickHouse service unavailable",
			domainerrors.WithOperation(op),
		)
	}

	// Authentication errors
	if strings.Contains(errMsg, ChErrAuthFailed) ||
		strings.Contains(errMsg, ChErrAccessDenied) {
		return domainerrors.New(
			err,
			domainerrors.EUNAUTHORIZED,
			"Failed to authenticate with ClickHouse",
			domainerrors.WithOperation(op),
		)
	}

	// Schema errors
	if strings.Contains(errMsg, ChErrNoSuchTable) ||
		strings.Contains(errMsg, ChErrTableNotFound) {
		return domainerrors.New(
			err,
			domainerrors.ENOTFOUND,
			"Table does not exist",
			domainerrors.WithOperation(op),
		)
	}

	if strings.Contains(errMsg, ChErrDatabaseNotFound) {
		return domainerrors.New(
			err,
			domainerrors.ENOTFOUND,
			"Database does not exist",
			domainerrors.WithOperation(op),
		)
	}

	if strings.Contains(errMsg, ChErrNoSuchColumn) ||
		strings.Contains(errMsg, ChErrUnknownColumn) {
		return domainerrors.New(
			err,
			domainerrors.EINVALID,
			"Column does not exist",
			domainerrors.WithOperation(op),
		)
	}

	// Query errors
	if strings.Contains(errMsg, ChErrSyntaxError) {
		return domainerrors.New(
			err,
			domainerrors.EINVALID,
			"SQL syntax error",
			domainerrors.WithOperation(op),
		)
	}

	if strings.Contains(errMsg, ChErrTypeError) {
		return domainerrors.New(
			err,
			domainerrors.EINVALID,
			"Type mismatch in query",
			domainerrors.WithOperation(op),
		)
	}

	// Check for duplicate/conflict
	if strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "already exists") {
		return domainerrors.New(
			err,
			domainerrors.ECONFLICT,
			"Resource already exists",
			domainerrors.WithOperation(op),
		)
	}

	// Timeout errors
	if strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "deadline exceeded") {
		return domainerrors.New(
			err,
			domainerrors.ETIMEOUT,
			"ClickHouse operation timed out",
			domainerrors.WithOperation(op),
		)
	}

	// Fallback for unhandled ClickHouse errors
	return domainerrors.New(
		err,
		domainerrors.EOLAP,
		"ClickHouse error",
		domainerrors.WithOperation(op),
	)
}

// Helper function to extract error code from ClickHouse error messages
func extractErrorCode(errMsg string) int {
	// Look for patterns like "Code: 60" or "code: 60"
	codeIndex := strings.Index(strings.ToLower(errMsg), "code: ")
	if codeIndex == -1 {
		return 0
	}

	codeIndex += 6 // Move past "code: "
	endIndex := codeIndex

	// Find the end of the code number
	for endIndex < len(errMsg) && errMsg[endIndex] >= '0' && errMsg[endIndex] <= '9' {
		endIndex++
	}

	if endIndex == codeIndex {
		return 0
	}

	var code int
	_, err := fmt.Sscanf(errMsg[codeIndex:endIndex], "%d", &code)
	if err != nil {
		return 0
	}

	return code
}
