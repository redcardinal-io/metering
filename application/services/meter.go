package services

import (
	"context"

	"github.com/redcardinal-io/metering/application/repositories"
	"github.com/redcardinal-io/metering/domain/models"
)

type MeterService struct {
	olap  repositories.OlapRepository
	store repositories.MeterStoreRepository
}

// NewMeterService creates a new MeterService with the provided OLAP and meter store repositories.
func NewMeterService(olap repositories.OlapRepository, store repositories.MeterStoreRepository) *MeterService {
	return &MeterService{
		olap:  olap,
		store: store,
	}
}

func (s *MeterService) CreateMeter(ctx context.Context, arg models.CreateMeterInput) (*models.Meter, error) {
	// Store the meter in the database
	m, err := s.store.CreateMeter(ctx, arg)
	if err != nil {
		return nil, err
	}

	// Call the repository to create the meter
	err = s.olap.CreateMeter(ctx, arg)
	if err != nil {
		// delete the meter from the database if OLAP creation fails
		if deleteErr := s.store.DeleteMeterByIDorSlug(ctx, m.Slug); deleteErr != nil {
			return nil, deleteErr
		}
		return nil, err
	}

	return m, nil
}

func (s *MeterService) GetMeter(ctx context.Context, IDorSlug string) (*models.Meter, error) {
	// Call the repository to get the meter
	m, err := s.store.GetMeterByIDorSlug(ctx, IDorSlug)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *MeterService) QueryMeter(ctx context.Context, arg models.QueryMeterInput) (*models.QueryMeterOutput, error) {
	m, err := s.store.GetMeterByIDorSlug(ctx, arg.MeterSlug)
	if err != nil {
		return nil, err
	}
	result, err := s.olap.QueryMeter(ctx, arg, &m.Aggregation)
	return result, err
}

func (s *MeterService) ListMeters(ctx context.Context)
