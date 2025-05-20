package planassignments

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
	"go.uber.org/zap"
)

func (p *PgPlanAssignmentsStoreRepository) UnAssignPlanToOrg(ctx context.Context, planId uuid.UUID, orgOrUserId uuid.UUID) error {
	err := p.q.UnAssignPlanToOrg(ctx, gen.UnAssignPlanToOrgParams{
		PlanID:         pgtype.UUID{Bytes: planId, Valid: true},
		OrganizationID: pgtype.UUID{Bytes: orgOrUserId, Valid: true},
	})
	if err != nil {
		p.logger.Error("failed to un-assign plan to the organization", zap.Error(err))
		return postgres.MapError(err, "Postgres.UnAssignPlan")
	}

	return nil
}

func (p *PgPlanAssignmentsStoreRepository) UnAssignPlanToUser(ctx context.Context, planId uuid.UUID, orgOrUserId uuid.UUID) error {
	err := p.q.UnAssignPlanToUser(ctx, gen.UnAssignPlanToUserParams{
		PlanID: pgtype.UUID{Bytes: planId, Valid: true},
		UserID: pgtype.UUID{Bytes: orgOrUserId, Valid: true},
	})
	if err != nil {
		p.logger.Error("failed to un-assign plan to the user", zap.Error(err))
		return postgres.MapError(err, "Postgres.UnAssignPlan")
	}

	return nil
}
