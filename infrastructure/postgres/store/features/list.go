package features

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/constants"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
	"github.com/redcardinal-io/metering/infrastructure/postgres"
	"github.com/redcardinal-io/metering/infrastructure/postgres/gen"
	"go.uber.org/zap"
)

func (p *PgFeatureRepository) ListFeatures(ctx context.Context, page pagination.Pagination) (*pagination.PaginationView[models.Feature], error) {
	tenantSlug := ctx.Value(constants.TenantSlugKey).(string)
	// extract type from pagination query params
	featureType := page.Queries["type"]

	m, err := p.q.ListFeaturesPaginated(ctx, gen.ListFeaturesPaginatedParams{
		Limit:      int32(page.Limit),
		Offset:     int32(page.GetOffset()),
		TenantSlug: tenantSlug,
		Type:       createFeatureTypeEnum(featureType),
	})
	if err != nil {
		p.logger.Error("Error listing features: ", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.ListFeatures")
	}

	features := make([]models.Feature, 0, len(m))
	for _, feature := range m {
		config := make(map[string]any)
		_ = json.Unmarshal(feature.Config, &config)
		id, _ := uuid.FromBytes(feature.ID.Bytes[:])

		features = append(features, models.Feature{
			Name:        feature.Name,
			Description: feature.Description.String,
			Slug:        feature.Slug,
			TenantSlug:  feature.TenantSlug,
			Type:        models.FeatureTypeEnum(feature.Type),
			Config:      config,
			Base: models.Base{
				ID:        id,
				CreatedAt: feature.CreatedAt,
				CreatedBy: feature.CreatedBy,
				UpdatedBy: feature.UpdatedBy,
				UpdatedAt: feature.UpdatedAt,
			},
		})
	}

	count, err := p.q.CountFeatures(ctx, gen.CountFeaturesParams{
		TenantSlug: tenantSlug,
		Type:       createFeatureTypeEnum(featureType),
	})
	if err != nil {
		p.logger.Error("Error counting features: ", zap.Error(err))
		return nil, postgres.MapError(err, "Postgres.CountFeatures")
	}

	result := pagination.FormatWith(page, int(count), features)
	return &result, nil
}

// createFeatureTypeEnum converts a feature type string into a nullable FeatureEnum for database queries.
// The returned enum is marked valid only if the input string is non-empty.
func createFeatureTypeEnum(featureType string) gen.NullFeatureEnum {
	return gen.NullFeatureEnum{
		FeatureEnum: gen.FeatureEnum(featureType),
		Valid:       featureType != "",
	}
}
