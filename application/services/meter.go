package services

import (
	"context"

	"github.com/redcardinal-io/metering/application/repositories"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
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

func (s *MeterService) GetMeterIDorSlug(ctx context.Context, IDorSlug string) (*models.Meter, error) {
	// Call the repository to get the meter
	m, err := s.store.GetMeterByIDorSlug(ctx, IDorSlug)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *MeterService) QueryMeter(ctx context.Context, arg models.QueryMeterParams) (*models.QueryMeterResult, error) {
	m, err := s.store.GetMeterByIDorSlug(ctx, arg.MeterSlug)
	if err != nil {
		return nil, err
	}
	result, err := s.olap.QueryMeter(ctx, arg, &m.Aggregation)
	return result, err
}

// TODO: implement recovery if store deletion fails
func (s *MeterService) DeleteMeter(ctx context.Context, iDorSlug string) error {
	meter, err := s.store.GetMeterByIDorSlug(ctx, iDorSlug)
	if err != nil {
		return err
	}

	// Call the OLAP repository to delete the meter
	err = s.olap.DeleteMeter(ctx, meter.Slug)
	if err != nil {
		return err
	}

	err = s.store.DeleteMeterByIDorSlug(ctx, iDorSlug)
	if err != nil {
		return err
	}

	return nil
}

func (s *MeterService) UpdateMeter(ctx context.Context, IDorSlug string, arg models.UpdateMeterInput) (*models.Meter, error) {
	// Call the store repository to update the meter
	m, err := s.store.UpdateMeterByIDorSlug(ctx, IDorSlug, arg)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (s *MeterService) ListMeters(ctx context.Context, pagination pagination.Pagination) (*pagination.PaginationView[models.Meter], error) {
	// Call the store repository to list the meters
	m, err := s.store.ListMeters(ctx, pagination)
	if err != nil {
		return nil, err
	}

	return m, nil
}
