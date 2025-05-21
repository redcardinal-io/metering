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

func (p *PgPlanAssignmentsStoreRepository) TerminateAssignment(ctx context.Context, arg models.TerminateAssignmentInput) error {
	// using must parse because the http handler should have already validated the UUID
	planID := uuid.MustParse(arg.PlanID)

	err := p.q.TerminateAssignedPlan(ctx, gen.TerminateAssignedPlanParams{
		PlanID:         pgtype.UUID{Bytes: planID, Valid: true},
		OrganizationID: pgtype.Text{String: arg.OrganizationId, Valid: arg.OrganizationId != ""},
		UserID:         pgtype.Text{String: arg.UserID, Valid: arg.UserID != ""},
	})
	if err != nil {
		p.logger.Error("failed to un-assign plan to the organization", zap.Error(err))
		return postgres.MapError(err, "Postgres.UnAssignPlan")
	}

	return nil
}
