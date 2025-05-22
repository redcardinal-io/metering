package planassignments

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
	"go.uber.org/zap"
)

func (p *PgPlanAssignmentsStoreRepository) CreateAssignment(ctx context.Context, arg models.CreateAssignmentInput) (*models.PlanAssignment, error) {
	m, err := p.q.AssignPlan(ctx, gen.AssignPlanParams{
		PlanID:         pgtype.UUID{Bytes: *arg.PlanID, Valid: true},
		ValidFrom:      pgtype.Timestamptz{Time: arg.ValidFrom, Valid: true},
		ValidUntil:     pgtype.Timestamptz{Time: arg.ValidUntil, Valid: true},
		OrganizationID: pgtype.Text{String: arg.OrganizationID, Valid: arg.OrganizationID != ""},
		UserID:         pgtype.Text{String: arg.UserID, Valid: arg.UserID != ""},
		CreatedBy:      arg.CreatedBy,
		UpdatedBy:      arg.CreatedBy,
	})
	if err != nil {
		p.logger.Error("failed to assign plan", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.CreateAssignment")
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
