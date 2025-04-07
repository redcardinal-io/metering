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
	// Call the repository to create the meter
	err := s.olap.CreateMeter(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to create meter: %w", err)
	}

	// Store the meter in the database
	m, err := s.store.CreateMeter(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to store meter: %w", err)
	}

	return m, nil
}
