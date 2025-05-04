package models

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Base struct {
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
	DeletedAt pgtype.Timestamptz `json:"deleted_at"`
	CreatedBy string             `json:"created_by"`
	UpdatedBy string             `json:"updated_by"`
}
