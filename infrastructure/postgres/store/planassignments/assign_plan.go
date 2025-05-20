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

func (p *PgPlanAssignmentsStoreRepository) AssignPlanToOrg(ctx context.Context, planId uuid.UUID, arg models.AssignOrUpdateAssignedPlanInput) (*models.PlanAssignment, error) {
	m, err := p.q.AssignPlanToOrg(ctx, gen.AssignPlanToOrgParams{
		PlanID:         pgtype.UUID{Bytes: planId, Valid: true},
		ValidFrom:      arg.ValidFrom,
		ValidUntil:     arg.ValidUntil,
		OrganizationID: arg.OrganizationOrUserId,
		CreatedBy:      arg.By,
		UpdatedBy:      arg.By,
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
		PlanId:         m.PlanID,
		OrganizationId: m.OrganizationID,
		ValidFrom:      m.ValidFrom,
		ValidUntil:     m.ValidUntil,
	}

	return planAssignment, nil
}

func (p *PgPlanAssignmentsStoreRepository) AssignPlanToUser(ctx context.Context, planId uuid.UUID, arg models.AssignOrUpdateAssignedPlanInput) (*models.PlanAssignment, error) {
	m, err := p.q.AssignPlanToUser(ctx, gen.AssignPlanToUserParams{
		PlanID:     pgtype.UUID{Bytes: planId, Valid: true},
		UserID:     arg.OrganizationOrUserId,
		ValidFrom:  arg.ValidFrom,
		ValidUntil: arg.ValidUntil,
		CreatedBy:  arg.By,
		UpdatedBy:  arg.By,
	})
	if err != nil {
		p.logger.Error("failed to assign plan to the user", zap.Error(err))
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
		PlanId:     m.PlanID,
		UserId:     m.UserID,
		ValidFrom:  m.ValidFrom,
		ValidUntil: m.ValidUntil,
	}

	return planAssignment, nil
}
