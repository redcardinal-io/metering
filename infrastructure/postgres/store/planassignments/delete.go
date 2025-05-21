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

func (p *PgPlanAssignmentsStoreRepository) TerminatePlan(ctx context.Context, arg *models.TerminateAssignedPlanInput) error {
	// using must parse because the http handler should have already validated the UUID
	planID := uuid.MustParse(arg.PlanId)

	err := p.q.TerminateAssignedPlan(ctx, gen.TerminateAssignedPlanParams{
		PlanID:         pgtype.UUID{Bytes: planID, Valid: true},
		OrganizationID: pgtype.Text{String: arg.OrganizationId, Valid: arg.OrganizationId != ""},
		UserID:         pgtype.Text{String: arg.UserId, Valid: arg.UserId != ""},
	})
	if err != nil {
		p.logger.Error("failed to un-assign plan to the organization", zap.Error(err))
		return postgres.MapError(err, "Postgres.UnAssignPlan")
	}

	return nil
}
