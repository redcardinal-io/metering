package postgres

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	domainerrors "github.com/redcardinal-io/metering/domain/errors"
)

// PostgreSQL error codes
const (
	// Connection errors
	PgErrConnectionFailure = "08006" // connection_failure
	PgErrConnectionLoss    = "08003" // connection_does_not_exist

	// Constraint violations
	PgErrUniqueViolation     = "23505" // unique_violation
	PgErrForeignKeyViolation = "23503" // foreign_key_violation
	PgErrCheckViolation      = "23514" // check_violation
	PgErrNotNullViolation    = "23502" // not_null_violation

)

// MapError converts a PostgreSQL or database-related error into a domain-specific error with contextual information.
// It interprets common PostgreSQL error codes and database error patterns, returning a corresponding domain error with a human-readable message and relevant metadata. If the error is not recognized, it returns a generic database domain error.
func MapError(err error, op string) error {
	if err == nil {
		return nil
	}

	// Check if this is a PostgreSQL error
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// Handle specific PostgreSQL errors
		switch pgErr.Code {
		case PgErrUniqueViolation:
			// Extract constraint name for more specific error message
			constraint := extractConstraintName(pgErr.Detail)
			msg := "A record with this information already exists"
			if constraint != "" {
				msg = "Duplicate value for " + humanizeConstraint(constraint)
			}
			return domainerrors.New(
				err,
				domainerrors.ECONFLICT,
				msg,
				domainerrors.WithOperation(op),
				domainerrors.WithData("constraint", constraint),
			)

		case PgErrForeignKeyViolation:
			constraint := extractConstraintName(pgErr.Detail)
			msg := "Referenced record does not exist"
			if constraint != "" {
				msg = "Referenced " + humanizeConstraint(constraint) + " does not exist"
			}
			return domainerrors.New(
				err,
				domainerrors.EINVALID,
				msg,
				domainerrors.WithOperation(op),
				domainerrors.WithData("constraint", constraint),
			)

		case PgErrNotNullViolation:
			column := extractColumnName(pgErr.Detail)
			msg := "Required field cannot be empty"
			if column != "" {
				msg = "Required field " + humanizeColumn(column) + " cannot be empty"
			}
			return domainerrors.New(
				err,
				domainerrors.EINVALID,
				msg,
				domainerrors.WithOperation(op),
				domainerrors.WithData("column", column),
			)

		case PgErrCheckViolation:
			constraint := extractConstraintName(pgErr.Detail)
			msg := "Constraint check failed"
			if constraint != "" {
				msg = humanizeConstraint(constraint) + " check failed"
			}
			return domainerrors.New(
				err,
				domainerrors.EINVALID,
				msg,
				domainerrors.WithOperation(op),
				domainerrors.WithData("constraint", constraint),
			)

		case PgErrConnectionFailure, PgErrConnectionLoss:
			return domainerrors.New(
				err,
				domainerrors.EUNAVAILABLE,
				"Database connection error",
				domainerrors.WithOperation(op),
			)

		default:
			// For all other Postgres errors, create a storage error
			return domainerrors.New(
				err,
				domainerrors.EDATABASE,
				"Database error occurred",
				domainerrors.WithOperation(op),
				domainerrors.WithInternal("Postgres error: "+pgErr.Message),
			)
		}
	}

	// Handle non-PostgreSQL specific database errors
	if err == ErrNoRows || strings.Contains(err.Error(), "no rows") {
		return domainerrors.New(
			err,
			domainerrors.ENOTFOUND,
			"Resource not found",
			domainerrors.WithOperation(op),
		)
	}

	if strings.Contains(err.Error(), "connection") && strings.Contains(err.Error(), "refused") {
		return domainerrors.New(
			err,
			domainerrors.EUNAVAILABLE,
			"Database unavailable",
			domainerrors.WithOperation(op),
		)
	}

	if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline exceeded") {
		return domainerrors.New(
			err,
			domainerrors.ETIMEOUT,
			"Database operation timed out",
			domainerrors.WithOperation(op),
		)
	}

	// Fallback for unhandled database errors
	return domainerrors.New(
		err,
		domainerrors.EDATABASE,
		"Database error",
		domainerrors.WithOperation(op),
	)
}

// ErrNoRows is a sentinel error for when no rows are found
var ErrNoRows = errors.New("no rows found")

// extractConstraintName returns the constraint name from a PostgreSQL error detail string, or an empty string if not found.
func extractConstraintName(detail string) string {
	if idx := strings.Index(detail, "constraint \""); idx >= 0 {
		end := strings.Index(detail[idx+12:], "\"")
		if end > 0 {
			return detail[idx+12 : idx+12+end]
		}
	}
	return ""
}

// extractColumnName parses and returns the column name from a PostgreSQL error detail string.
// Returns an empty string if no column name is found.
func extractColumnName(detail string) string {
	if idx := strings.Index(detail, "column \""); idx >= 0 {
		end := strings.Index(detail[idx+8:], "\"")
		if end > 0 {
			return detail[idx+8 : idx+8+end]
		}
	}
	return ""
}

// humanizeConstraint converts a technical database constraint name into a human-readable string by removing common prefixes, replacing underscores with spaces, and capitalizing each word.
func humanizeConstraint(constraint string) string {
	// Remove common prefixes
	constraint = strings.TrimPrefix(constraint, "pk_")
	constraint = strings.TrimPrefix(constraint, "fk_")
	constraint = strings.TrimPrefix(constraint, "uq_")
	constraint = strings.TrimPrefix(constraint, "ck_")

	// Replace underscores with spaces and capitalize
	words := strings.Split(constraint, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[0:1]) + word[1:]
		}
	}

	return strings.Join(words, " ")
}

// humanizeColumn converts a technical column name to a human-readable string by replacing underscores with spaces and capitalizing each word.
func humanizeColumn(column string) string {
	// Replace underscores with spaces and capitalize
	words := strings.Split(column, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[0:1]) + word[1:]
		}
	}

	return strings.Join(words, " ")
}
