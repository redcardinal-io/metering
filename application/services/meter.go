package services

import (
	"context"
	"fmt"

	"github.com/redcardinal-io/metering/application/repositories"
	"github.com/redcardinal-io/metering/domain/models"
)

type MeterService struct {
	olap  repositories.OlapRepository
	store repositories.MeterStoreRepository
}

func NewMeterService(olap repositories.OlapRepository, store repositories.MeterStoreRepository) *MeterService {
	return &MeterService{
		olap:  olap,
		store: store,
	}
}

func (s *MeterService) CreateMeter(ctx context.Context, arg models.CreateMeterInput) (*models.Meter, error) {

	// check if the meter already exists
	m, err := s.store.GetMeterByIDorSlug(ctx, arg.MeterSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to get meter by ID or slug: %w", err)
	}
	if m != nil {
		return nil, fmt.Errorf("meter with slug %s already exists", arg.MeterSlug)
	}

	// Call the repository to create the meter
	err = s.olap.CreateMeter(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to create meter: %w", err)
	}

	// Store the meter in the database
	m, err = s.store.CreateMeter(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to store meter: %w", err)
	}

	return m, nil
}

func (s *MeterService) GetMeter(ctx context.Context, IDorSlug string) (*models.Meter, error) {
	// Call the repository to get the meter
	m, err := s.store.GetMeterByIDorSlug(ctx, IDorSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to get meter: %w", err)
	}
	if m == nil {
		return nil, fmt.Errorf("meter with ID %s not found", IDorSlug)
	}

	return m, nil
}

func (s *MeterService) QueryMeter(ctx context.Context, arg models.QueryMeterInput) (*models.QueryMeterOutput, error) {
	m, err := s.store.GetMeterByIDorSlug(ctx, arg.MeterSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to get meter by ID or slug: %w", err)
	}

	result, err := s.olap.QueryMeter(ctx, arg, &m.Aggregation)
	if err != nil {
		return nil, fmt.Errorf("failed to query meter: %w", err)
	}

	return result, nil
}
