package planassignments

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
	"go.uber.org/zap"
)

func (p *PgPlanAssignmentsStoreRepository) UpdateAssignment(ctx context.Context, arg models.UpdateAssignmentInput) (*models.PlanAssignment, error) {

	validFrom := pgtype.Timestamptz{Valid: false}
	validUntil := pgtype.Timestamptz{Valid: false}
	if !arg.ValidFrom.IsZero() {
		validFrom = pgtype.Timestamptz{Time: *arg.ValidFrom, Valid: true}
	}

	if !arg.ValidUntil.IsZero() {
		validUntil = pgtype.Timestamptz{Time: *arg.ValidUntil, Valid: true}
	} else if arg.SetValidUntilToZero {
		validUntil = pgtype.Timestamptz{Time: time.Time{}, Valid: true}
	}

	m, err := p.q.UpdateAssignedPlan(ctx, gen.UpdateAssignedPlanParams{
		PlanID:         pgtype.UUID{Bytes: *arg.PlanID, Valid: true},
		OrganizationID: pgtype.Text{String: arg.OrganizationID, Valid: arg.OrganizationID != ""},
		UserID:         pgtype.Text{String: arg.UserID, Valid: arg.UserID != ""},
		UpdatedBy:      arg.UpdatedBy,
		ValidFrom:      validFrom,
		ValidUntil:     validUntil,
	})
	if err != nil {
		p.logger.Error("failed to update assigned plan ", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.UpdateAssignment")
	}

	id, err := uuid.FromBytes(m.ID.Bytes[:])
	if err != nil {
		p.logger.Error("failed to parse UUID from bytes", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.ParseUUID")
	}

	planAssignment := &models.PlanAssignment{
		Base: models.Base{
			ID:        id,
			CreatedAt: m.CreatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
			UpdatedAt: m.UpdatedAt,
		},
		PlanID:         m.PlanID.String(),
		OrganizationID: m.OrganizationID.String,
		UserID:         m.UserID.String,
		ValidFrom:      m.ValidFrom.Time,
		ValidUntil:     m.ValidUntil.Time,
	}

	return planAssignment, nil
}
