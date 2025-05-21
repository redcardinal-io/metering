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

func (p *PgPlanAssignmentsStoreRepository) UpdateAssignment(ctx context.Context, arg models.UpdateAssignmentInput) (*models.PlanAssignment, error) {
	// using must parse because the http handler should have already validated the UUID
	planId := uuid.MustParse(arg.PlanID)

	m, err := p.q.UpdateAssignedPlan(ctx, gen.UpdateAssignedPlanParams{
		PlanID:         pgtype.UUID{Bytes: planId, Valid: true},
		OrganizationID: pgtype.Text{String: arg.OrganizationID, Valid: arg.OrganizationID != ""},
		UserID:         pgtype.Text{String: arg.UserID, Valid: arg.UserID != ""},
		UpdatedBy:      arg.UpdatedBy,
		ValidFrom:      pgtype.Timestamptz{Time: arg.ValidFrom, Valid: true},
		ValidUntil:     pgtype.Timestamptz{Time: arg.ValidUntil, Valid: true},
	})
	if err != nil {
		p.logger.Error("failed to update assigned plan to the organization", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.UpdateAssignPlan")
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
		UserId:         m.UserID.String,
		ValidFrom:      m.ValidFrom.Time,
		ValidUntil:     m.ValidUntil.Time,
	}

	return planAssignment, nil
}
