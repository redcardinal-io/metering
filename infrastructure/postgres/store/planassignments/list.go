package planassignments

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
	"go.uber.org/zap"
)

func (p *PgPlanAssignmentsStoreRepository) ListAssignments(ctx context.Context, arg models.QueryPlanAssignmentInput, page pagination.Pagination) (*pagination.PaginationView[models.PlanAssignment], error) {
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)
	planId := pgtype.UUID{Valid: false}
	validFrom := pgtype.Timestamptz{Valid: false}
	validUntil := pgtype.Timestamptz{Valid: false}

	if !arg.ValidFrom.IsZero() {
		validFrom = pgtype.Timestamptz{Time: arg.ValidFrom, Valid: true}
	}

	if !arg.ValidUntil.IsZero() {
		validUntil = pgtype.Timestamptz{Time: arg.ValidUntil, Valid: true}
	}

	if arg.PlanID != nil {
		planId = pgtype.UUID{Bytes: *arg.PlanID, Valid: true}
	}

	m, err := p.q.ListAssignmentsPaginated(ctx, gen.ListAssignmentsPaginatedParams{
		Limit:          int32(page.Limit),
		Offset:         int32(page.GetOffset()),
		OrganizationID: pgtype.Text{String: arg.OrganizationID, Valid: arg.OrganizationID != ""},
		PlanID:         planId,
		UserID:         pgtype.Text{String: arg.UserID, Valid: arg.UserID != ""},
		ValidFrom:      validFrom,
		ValidUntil:     validUntil,
		TenantSlug:     tenantSlug,
	})
	if err != nil {
		p.logger.Error("Error listing assignments: ", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.ListAssignments")
	}

	planassignments := make([]models.PlanAssignment, 0, len(m))
	for _, planassignment := range m {
		id, _ := uuid.FromBytes(planassignment.ID.Bytes[:])
		if planassignment.ValidUntil.Time.IsZero() {
			planassignment.ValidUntil.Time = planassignment.ValidUntil.Time.UTC()
		}
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

	count, err := p.q.CountAssignments(ctx, gen.CountAssignmentsParams{
		OrganizationID: pgtype.Text{String: arg.OrganizationID, Valid: arg.OrganizationID != ""},
		PlanID:         planId,
		UserID:         pgtype.Text{String: arg.UserID, Valid: arg.UserID != ""},
		ValidFrom:      validFrom,
		ValidUntil:     validUntil,
		TenantSlug:     tenantSlug,
	})
	if err != nil {
		p.logger.Error("Error counting assignments: ", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.CountAssignments")
	}

	result := pagination.FormatWith(page, int(count), planassignments)

	return &result, nil
}

func (p *PgPlanAssignmentsStoreRepository) ListAllAssignments(ctx context.Context, page pagination.Pagination) (*pagination.PaginationView[models.PlanAssignment], error) {
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)

	m, err := p.q.ListAllAssignmentsPaginated(ctx, gen.ListAllAssignmentsPaginatedParams{
		Limit:      int32(page.Limit),
		Offset:     int32(page.GetOffset()),
		TenantSlug: tenantSlug,
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

	count, err := p.q.CountAllAssignments(ctx, tenantSlug)
	if err != nil {
		p.logger.Error("Error counting assignments: ", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.CountAssignments")
	}

	result := pagination.FormatWith(page, int(count), planassignments)

	return &result, nil
}

func (p *PgPlanAssignmentsStoreRepository) ListAssignmentsHistory(ctx context.Context, arg models.QueryPlanAssignmentHistoryInput, page pagination.Pagination) (*pagination.PaginationView[models.PlanAssignmentHistory], error) {
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)
	planId := pgtype.UUID{Valid: false}
	validFromBefore := pgtype.Timestamptz{Valid: false}
	validFromAfter := pgtype.Timestamptz{Valid: false}
	validUntilBefore := pgtype.Timestamptz{Valid: false}
	validUntilAfter := pgtype.Timestamptz{Valid: false}

	if !arg.ValidFromBefore.IsZero() {
		validFromBefore = pgtype.Timestamptz{Time: arg.ValidFromBefore, Valid: true}
	}

	if !arg.ValidFromAfter.IsZero() {
		validFromAfter = pgtype.Timestamptz{Time: arg.ValidFromAfter, Valid: true}
	}

	if !arg.ValidUntilBefore.IsZero() {
		validUntilBefore = pgtype.Timestamptz{Time: arg.ValidUntilBefore, Valid: true}
	}

	if !arg.ValidUntilAfter.IsZero() {
		validUntilAfter = pgtype.Timestamptz{Time: arg.ValidUntilAfter, Valid: true}
	}
	if arg.PlanID != nil {
		planId = pgtype.UUID{Bytes: *arg.PlanID, Valid: true}
	}

	m, err := p.q.ListAssignmentsHistoryPaginated(ctx, gen.ListAssignmentsHistoryPaginatedParams{
		Limit:          int32(page.Limit),
		Offset:         int32(page.GetOffset()),
		OrganizationID: pgtype.Text{String: arg.OrganizationID, Valid: arg.OrganizationID != ""},
		Action:         pgtype.Text{String: arg.Action, Valid: arg.Action != ""},
		PlanID:         planId,
		UserID:         pgtype.Text{String: arg.UserID, Valid: arg.UserID != ""},
		ValidFrom:      validFromBefore,
		ValidFrom_2:    validFromAfter,
		ValidUntil:     validUntilBefore,
		ValidUntil_2:   validUntilAfter,
		TenantSlug:     tenantSlug,
	})
	if err != nil {
		p.logger.Error("Error listing assignments history: ", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.ListAssignmentsHistory")
	}

	planassignments := make([]models.PlanAssignmentHistory, 0, len(m))
	for _, planassignment := range m {
		id, _ := uuid.FromBytes(planassignment.ID.Bytes[:])
		planassignments = append(planassignments, models.PlanAssignmentHistory{
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
			Action:         planassignment.Action.String,
		})
	}

	count, err := p.q.CountAssignmentsHistory(ctx, gen.CountAssignmentsHistoryParams{
		OrganizationID: pgtype.Text{String: arg.OrganizationID, Valid: arg.OrganizationID != ""},
		Action:         pgtype.Text{String: arg.Action, Valid: arg.Action != ""},
		PlanID:         planId,
		UserID:         pgtype.Text{String: arg.UserID, Valid: arg.UserID != ""},
		ValidFrom:      validFromBefore,
		ValidFrom_2:    validFromAfter,
		ValidUntil:     validUntilBefore,
		ValidUntil_2:   validUntilAfter,
		TenantSlug:     tenantSlug,
	})
	if err != nil {
		p.logger.Error("Error counting assignments history ", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.CountAssignmentsHistory")
	}

	result := pagination.FormatWith(page, int(count), planassignments)

	return &result, nil
}
