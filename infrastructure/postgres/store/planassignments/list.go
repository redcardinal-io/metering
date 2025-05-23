package planassignments

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
	"go.uber.org/zap"
)

func (p *PgPlanAssignmentsStoreRepository) ListOrgOrUserPlanAssignments(ctx context.Context, orgId string, userId string, page pagination.Pagination) (*pagination.PaginationView[models.PlanAssignment], error) {

	m, err := p.q.ListOrgOrUserAssignmentsPaginated(ctx, gen.ListOrgOrUserAssignmentsPaginatedParams{
		Limit:          int32(page.Limit),
		Offset:         int32(page.GetOffset()),
		OrganizationID: pgtype.Text{String: orgId, Valid: orgId != ""},
		UserID:         pgtype.Text{String: userId, Valid: userId != ""},
	})

	if err != nil {
		p.logger.Error("Error listing assignments: ", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.ListAssignments")
	}

	planassignments := make([]models.PlanAssignment, 0, len(m))
	for _, planassignment := range m {
		id, _ := uuid.FromBytes(planassignment.ID.Bytes[:])
		planassignments = append(planassignments, models.PlanAssignment{
			Base: models.Base{
				ID:        id,
				CreatedAt: planassignment.CreatedAt,
				CreatedBy: planassignment.CreatedBy,
				UpdatedBy: planassignment.UpdatedBy,
				UpdatedAt: planassignment.UpdatedAt,
			},
			PlanID:         planassignment.PlanID.String(),
			OrganizationID: planassignment.OrganizationID.String,
			UserID:         planassignment.UserID.String,
			ValidFrom:      planassignment.ValidFrom.Time,
			ValidUntil:     planassignment.ValidUntil.Time,
		})
	}

	count, err := p.q.CountOrgOrUserAssignments(ctx, gen.CountOrgOrUserAssignmentsParams{
		OrganizationID: pgtype.Text{String: orgId, Valid: orgId != ""},
		UserID:         pgtype.Text{String: userId, Valid: userId != ""},
	})

	if err != nil {
		p.logger.Error("Error counting meters: ", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.CountAssignments")
	}

	result := pagination.FormatWith(page, int(count), planassignments)

	return &result, nil
}
