package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
)

// MockPlanAssignmentsStoreRepository is a mock implementation of the PlanAssignmentsStoreRepository interface
type MockPlanAssignmentsStoreRepository struct {
	mock.Mock
}

// Implement the interface methods for the mock
func (m *MockPlanAssignmentsStoreRepository) CreateAssignment(ctx context.Context, arg models.CreateAssignmentInput) (*models.PlanAssignment, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlanAssignment), args.Error(1)
}

func (m *MockPlanAssignmentsStoreRepository) TerminateAssignment(ctx context.Context, arg models.TerminateAssignmentInput) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockPlanAssignmentsStoreRepository) UpdateAssignment(ctx context.Context, arg models.UpdateAssignmentInput) (*models.PlanAssignment, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlanAssignment), args.Error(1)
}

func (m *MockPlanAssignmentsStoreRepository) ListAssignments(ctx context.Context, arg models.QueryPlanAssignmentInput, p pagination.Pagination) (*pagination.PaginationView[models.PlanAssignment], error) {
	args := m.Called(ctx, arg, p)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pagination.PaginationView[models.PlanAssignment]), args.Error(1)
}

func (m *MockPlanAssignmentsStoreRepository) ListAssignmentsHistory(ctx context.Context, arg models.QueryPlanAssignmentHistoryInput, p pagination.Pagination) (*pagination.PaginationView[models.PlanAssignmentHistory], error) {
	args := m.Called(ctx, arg, p)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pagination.PaginationView[models.PlanAssignmentHistory]), args.Error(1)
}

func (m *MockPlanAssignmentsStoreRepository) ListAllAssignments(ctx context.Context, p pagination.Pagination) (*pagination.PaginationView[models.PlanAssignment], error) {
	args := m.Called(ctx, p)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pagination.PaginationView[models.PlanAssignment]), args.Error(1)
}

func TestValidateAssignmentTimeRange(t *testing.T) {
	// Define common variables for tests
	ctx := context.Background()
	now := time.Now().UTC()
	planID := uuid.New()
	orgID := uuid.New()
	userID := uuid.New()

	// Define test cases
	tests := []struct {
		name         string
		setupMock    func(*MockPlanAssignmentsStoreRepository)
		updateInput  models.UpdateAssignmentInput
		expectedErr  bool
		exactMessage string
	}{
		{
			name: "error with no existing assignments",
			setupMock: func(m *MockPlanAssignmentsStoreRepository) {
				m.On("ListAssignments", mock.Anything, mock.Anything, mock.Anything).
					Return(&pagination.PaginationView[models.PlanAssignment]{
						Results: []models.PlanAssignment{},
						Total:   0,
					}, nil)
			},
			updateInput: models.UpdateAssignmentInput{
				PlanID:         &planID,
				OrganizationID: orgID.String(),
				UserID:         userID.String(),
				ValidFrom:      now.Add(24 * time.Hour),
				ValidUntil:     now.Add(48 * time.Hour),
			},
			expectedErr:  true,
			exactMessage: "no existing assignment found",
		},
		{
			name: "error when listing assignments",
			setupMock: func(m *MockPlanAssignmentsStoreRepository) {
				m.On("ListAssignments", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("database error"))
			},
			updateInput: models.UpdateAssignmentInput{
				PlanID:         &planID,
				OrganizationID: orgID.String(),
				UserID:         userID.String(),
			},
			expectedErr:  true,
			exactMessage: "failed to list assignments",
		},
		{
			name: "error when valid from is before existing assignment's valid from",
			setupMock: func(m *MockPlanAssignmentsStoreRepository) {
				m.On("ListAssignments", mock.Anything, mock.Anything, mock.Anything).
					Return(&pagination.PaginationView[models.PlanAssignment]{
						Results: []models.PlanAssignment{{
							ValidFrom:  now,
							ValidUntil: now.Add(24 * time.Hour),
						}},
						Total: 1,
					}, nil)
			},
			updateInput: models.UpdateAssignmentInput{
				PlanID:         &planID,
				OrganizationID: orgID.String(),
				UserID:         userID.String(),
				ValidFrom:      now.Add(-1 * time.Hour),
				ValidUntil:     now.Add(1 * time.Hour),
			},
			expectedErr:  true,
			exactMessage: "valid_from cannot be before the current valid_from",
		},
		{
			name: "error when valid until is after existing assignment's valid until",
			setupMock: func(m *MockPlanAssignmentsStoreRepository) {
				m.On("ListAssignments", mock.Anything, mock.Anything, mock.Anything).
					Return(&pagination.PaginationView[models.PlanAssignment]{
						Results: []models.PlanAssignment{{
							ValidFrom:  now,
							ValidUntil: now.Add(24 * time.Hour),
						}},
						Total: 1,
					}, nil)
			},
			updateInput: models.UpdateAssignmentInput{
				PlanID:         &planID,
				OrganizationID: orgID.String(),
				UserID:         userID.String(),
				ValidFrom:      now.Add(1 * time.Hour),
				ValidUntil:     now.Add(48 * time.Hour),
			},
			expectedErr:  true,
			exactMessage: "valid_until cannot be after the current valid_until",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize mock
			mockRepo := new(MockPlanAssignmentsStoreRepository)
			tt.setupMock(mockRepo)

			// Initialize service with mock
			service := &PlanManagementService{
				planAssignmentsStore: mockRepo,
			}

			// Execute the method under test
			err := service.validateAssignmentTimeRange(ctx, tt.updateInput)

			// Assertions
			if tt.expectedErr {
				assert.Error(t, err)
				if tt.exactMessage != "" {
					assert.Contains(t, err.Error(), tt.exactMessage)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

// Test with successful validation when assignment exists
func TestValidateAssignmentTimeRangeSuccessful(t *testing.T) {
	// Setup test data
	ctx := context.Background()
	now := time.Now().UTC()
	planID := uuid.New()
	orgID := uuid.New()
	userID := uuid.New()

	// Create mock repository
	mockRepo := new(MockPlanAssignmentsStoreRepository)

	// Create service with mocked dependencies
	service := PlanManagementService{
		planAssignmentsStore: mockRepo,
	}

	// Set up the test case where ListAssignments returns a valid assignment
	updateInput := models.UpdateAssignmentInput{
		PlanID:         &planID,
		OrganizationID: orgID.String(),
		UserID:         userID.String(),
		ValidFrom:      now.Add(1 * time.Hour),
		ValidUntil:     now.Add(23 * time.Hour),
	}

	// Mock the ListAssignments to return a valid assignment
	mockRepo.On("ListAssignments", mock.Anything, mock.Anything, mock.Anything).Return(
		&pagination.PaginationView[models.PlanAssignment]{
			Results: []models.PlanAssignment{
				{
					ValidFrom:  now,
					ValidUntil: now.Add(24 * time.Hour),
				},
			},
			Total: 1,
		},
		nil,
	)

	// Test the function
	err := service.validateAssignmentTimeRange(ctx, updateInput)

	// Verify no error is returned for a valid time range
	assert.NoError(t, err, "Expected no error for valid time range")

	// Verify the mock was called as expected
	mockRepo.AssertExpectations(t)
}
