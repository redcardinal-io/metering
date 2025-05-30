package planassignments

import (
	"context"

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

	planAssignment := toPlanAssignmentModel(m)
	if planAssignment.ValidUntil.IsZero() {
		planAssignment.ValidUntil = planAssignment.ValidUntil.UTC()
	}

	return planAssignment, nil
}
