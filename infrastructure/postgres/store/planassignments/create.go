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

func (p *PgPlanAssignmentsStoreRepository) AssignPlan(ctx context.Context, arg models.AssignPlanInput) (*models.PlanAssignment, error) {
	// using must parse because the http handler should have already validated the UUID
	planId := uuid.MustParse(arg.PlanID)

	m, err := p.q.AssignPlan(ctx, gen.AssignPlanParams{
		PlanID:         pgtype.UUID{Bytes: planId, Valid: true},
		ValidFrom:      pgtype.Timestamptz{Time: arg.ValidFrom, Valid: true},
		ValidUntil:     pgtype.Timestamptz{Time: arg.ValidUntil, Valid: true},
		OrganizationID: pgtype.Text{String: arg.OrganizationID, Valid: arg.OrganizationID != ""},
		CreatedBy:      arg.CreatedBy,
		UpdatedBy:      arg.CreatedBy,
	})
	if err != nil {
		p.logger.Error("failed to assign plan to the organization", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.AssignPlan")
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
		PlanId:         m.PlanID.String(),
		OrganizationId: m.OrganizationID.String,
		UserId:         m.UserID.String,
		ValidFrom:      m.ValidFrom.Time,
		ValidUntil:     m.ValidUntil.Time,
	}

	return planAssignment, nil
}
