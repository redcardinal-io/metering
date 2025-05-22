package planassignments

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
	"go.uber.org/zap"
)

func (p *PgPlanAssignmentsStoreRepository) TerminateAssignment(ctx context.Context, arg models.TerminateAssignmentInput) error {

	err := p.q.TerminateAssignedPlan(ctx, gen.TerminateAssignedPlanParams{
		PlanID:         pgtype.UUID{Bytes: *arg.PlanID, Valid: true},
		OrganizationID: pgtype.Text{String: arg.OrganizationID, Valid: arg.OrganizationID != ""},
		UserID:         pgtype.Text{String: arg.UserID, Valid: arg.UserID != ""},
	})
	if err != nil {
		p.logger.Error("failed to un-assign plan to the organization", zap.Error(err))
		return postgres.MapError(err, "Postgres.TerminatePlan")
	}

	return nil
}
