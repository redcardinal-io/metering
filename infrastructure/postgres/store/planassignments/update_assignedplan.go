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

func (p *PgPlanAssignmentsStoreRepository) UpdateAssignedPlanToOrg(ctx context.Context, planId uuid.UUID, arg models.AssignOrUpdateAssignedPlanInput) (*models.PlanAssignment, error) {
	m, err := p.q.UpdateOrgsValidFromAndUntil(ctx, gen.UpdateOrgsValidFromAndUntilParams{
		PlanID:         pgtype.UUID{Bytes: planId, Valid: true},
		OrganizationID: arg.OrganizationOrUserId,
		UpdatedBy:      arg.By,
		ValidFrom:      arg.ValidFrom,
		ValidUntil:     arg.ValidUntil,
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
		PlanId:         m.PlanID,
		OrganizationId: m.OrganizationID,
		ValidFrom:      m.ValidFrom,
		ValidUntil:     m.ValidUntil,
	}

	return planAssignment, nil
}

func (p *PgPlanAssignmentsStoreRepository) UpdateAssignedPlanToUser(ctx context.Context, planId uuid.UUID, arg models.AssignOrUpdateAssignedPlanInput) (*models.PlanAssignment, error) {

	m, err := p.q.UpdateUsersValidFromAndUntil(ctx, gen.UpdateUsersValidFromAndUntilParams{
		PlanID:     pgtype.UUID{Bytes: planId, Valid: true},
		UserID:     arg.OrganizationOrUserId,
		UpdatedBy:  arg.By,
		ValidFrom:  arg.ValidFrom,
		ValidUntil: arg.ValidUntil,
	})

	if err != nil {
		p.logger.Error("failed to update assigned plan to the user", zap.Error(err))
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
		PlanId:         m.PlanID,
		OrganizationId: m.OrganizationID,
		ValidFrom:      m.ValidFrom,
		ValidUntil:     m.ValidUntil,
	}

	return planAssignment, nil
}
