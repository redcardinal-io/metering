package models

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Base struct {
	ID        uuid.UUID          `json:"id"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
	CreatedBy string             `json:"created_by"`
	UpdatedBy string             `json:"updated_by"`
}
