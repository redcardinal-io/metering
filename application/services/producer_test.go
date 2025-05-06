package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domainerrors "github.com/redcardinal-io/metering/domain/errors"
	"github.com/redcardinal-io/metering/domain/models"
	"github.com/redcardinal-io/metering/domain/pkg/pagination"
)

// MockProducerRepository is a mock implementation of ProducerRepository
type MockProducerRepository struct {
	mock.Mock
}

func (m *MockProducerRepository) PublishEvents(topic string, eventBatch *models.EventBatch) error {
	args := m.Called(topic, eventBatch)
	return args.Error(0)
}

func (m *MockProducerRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockMeterStoreRepository is a mock implementation of MeterStoreRepository
type MockMeterStoreRepository struct {
	mock.Mock
}

func (m *MockMeterStoreRepository) CreateMeter(ctx context.Context, arg models.CreateMeterInput) (*models.Meter, error) {
	return nil, nil
}

func (m *MockMeterStoreRepository) GetMeterByIDorSlug(ctx context.Context, idOrSlug string) (*models.Meter, error) {
	return nil, nil
}

func (m *MockMeterStoreRepository) ListMeters(ctx context.Context, pagination pagination.Pagination) (*pagination.PaginationView[models.Meter], error) {
	return nil, nil
}

func (m *MockMeterStoreRepository) ListMetersByEventTypes(ctx context.Context, eventTypes []string) ([]*models.Meter, error) {
	args := m.Called(ctx, eventTypes)
	if pierwszym := args.Get(0); pierwszym == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Meter), args.Error(1)
}

func (m *MockMeterStoreRepository) DeleteMeterByIDorSlug(ctx context.Context, idOrSlug string) error {
	return nil
}

func (m *MockMeterStoreRepository) UpdateMeterByIDorSlug(ctx context.Context, idOrSlug string, arg models.UpdateMeterInput) (*models.Meter, error) {
	return nil, nil
}

func TestProducerService_PublishEvents(t *testing.T) {
	const testTopic = "test-topic"
	ctx := context.Background()

	newTestEvent := func(id, eventType string, properties map[string]any) *models.Event {
		propsJSON := "{}"
		if properties != nil {
			data, _ := json.Marshal(properties)
			propsJSON = string(data)
		}
		return &models.Event{
			ID:           id,
			Type:         eventType,
			Source:       fmt.Sprintf("source-for-%s", id),
			Organization: "org-default",
			User:         "user-default",
			Timestamp:    "2024-05-06T12:00:00Z",
			Properties:   propsJSON,
		}
	}

	baseMeter := &models.Meter{
		EventType:     "type1",
		Properties:    []string{"propA", "propB"},
		ValueProperty: "valueProp",
		Slug:          "meter-type1",
	}

	baseMeterOnlyValueProp := &models.Meter{
		EventType:     "type2",
		ValueProperty: "cost",
		Slug:          "meter-type2",
	}

	baseMeterNoProps := &models.Meter{
		EventType: "type3",
		Slug:      "meter-type3",
	}

	t.Run("successfully publish all valid events", func(t *testing.T) {
		mockProducer := new(MockProducerRepository)
		mockStore := new(MockMeterStoreRepository)
		service := NewProducerService(mockProducer, mockStore)

		events := &models.EventBatch{
			Events: []*models.Event{
				newTestEvent("ev1", "type1", map[string]any{"propA": "valA", "propB": 123, "valueProp": 10.5}),
				newTestEvent("ev2", "type2", map[string]any{"cost": 5}),
			},
		}

		mockStore.On("ListMetersByEventTypes", ctx, mock.AnythingOfType("[]string")).Run(func(args mock.Arguments) {
			eventTypes := args.Get(1).([]string)
			assert.ElementsMatch(t, []string{"type1", "type2"}, eventTypes)
		}).Return([]*models.Meter{baseMeter, baseMeterOnlyValueProp}, nil).Once()

		mockProducer.On("PublishEvents", testTopic, mock.AnythingOfType("*models.EventBatch")).Run(func(args mock.Arguments) {
			batch := args.Get(1).(*models.EventBatch)
			assert.Len(t, batch.Events, 2)
		}).Return(nil).Once()

		result, err := service.PublishEvents(ctx, testTopic, events, true)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.SuccessCount)
		assert.Empty(t, result.FailedEvents)
		mockStore.AssertExpectations(t)
		mockProducer.AssertExpectations(t)
	})

	t.Run("batch size exceeds maximum", func(t *testing.T) {
		mockProducer := new(MockProducerRepository)
		mockStore := new(MockMeterStoreRepository)
		service := NewProducerService(mockProducer, mockStore)

		var oversizedEvents []*models.Event
		for i := range MaxBatchSize + 1 {
			oversizedEvents = append(oversizedEvents, newTestEvent(fmt.Sprintf("ev%d", i), "type1", nil))
		}
		events := &models.EventBatch{Events: oversizedEvents}

		result, err := service.PublishEvents(ctx, testTopic, events, true)

		assert.Error(t, err)
		var appErr *domainerrors.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, domainerrors.EINVALID, domainerrors.ErrorCode(appErr.Code))
		assert.Contains(t, appErr.Message, "batch size too large")
		assert.Nil(t, result)
	})

	t.Run("fetchAndPrepareMeterConfig returns error - no valid event types in non-empty batch", func(t *testing.T) {
		mockProducer := new(MockProducerRepository)
		mockStore := new(MockMeterStoreRepository)
		service := NewProducerService(mockProducer, mockStore)

		events := &models.EventBatch{Events: []*models.Event{newTestEvent("ev1", "", nil)}} // Empty type

		result, err := service.PublishEvents(ctx, testTopic, events, true)

		assert.Error(t, err)
		var appErr *domainerrors.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, domainerrors.EINVALID, domainerrors.ErrorCode(appErr.Code))
		assert.Contains(t, appErr.Internal, "no valid event types found in the batch")
		assert.Equal(t, "no valid event types", appErr.Message)
		assert.Nil(t, result)
	})

	t.Run("fetchAndPrepareMeterConfig returns error - completely empty batch", func(t *testing.T) {
		mockProducer := new(MockProducerRepository)
		mockStore := new(MockMeterStoreRepository)
		service := NewProducerService(mockProducer, mockStore)

		events := &models.EventBatch{Events: []*models.Event{}} // Empty batch

		result, err := service.PublishEvents(ctx, testTopic, events, true)

		assert.Error(t, err)
		var appErr *domainerrors.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, domainerrors.EINVALID, domainerrors.ErrorCode(appErr.Code))
		assert.Contains(t, appErr.Internal, "no valid event types found in the batch")
		assert.Equal(t, "no valid event types", appErr.Message)
		assert.Nil(t, result)
	})

	t.Run("fetchAndPrepareMeterConfig returns error - store fails", func(t *testing.T) {
		mockProducer := new(MockProducerRepository)
		mockStore := new(MockMeterStoreRepository)
		service := NewProducerService(mockProducer, mockStore)

		events := &models.EventBatch{Events: []*models.Event{newTestEvent("ev1", "type1", nil)}}
		storeErr := errors.New("database connection failed")
		mockStore.On("ListMetersByEventTypes", ctx, []string{"type1"}).Return(nil, storeErr).Once()

		result, err := service.PublishEvents(ctx, testTopic, events, true)

		assert.Error(t, err)
		var appErr *domainerrors.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, domainerrors.EINTERNAL, domainerrors.ErrorCode(appErr.Code))
		assert.Contains(t, appErr.Internal, "failed to fetch meters")
		assert.Equal(t, "meter fetching error", appErr.Message)
		assert.Nil(t, result)
		mockStore.AssertExpectations(t)
	})

	t.Run("validation fails - nil event, allowPartialSuccess=true", func(t *testing.T) {
		mockProducer := new(MockProducerRepository)
		mockStore := new(MockMeterStoreRepository)
		service := NewProducerService(mockProducer, mockStore)

		events := &models.EventBatch{
			Events: []*models.Event{
				nil,
				newTestEvent("ev2", "type1", map[string]any{"propA": "valA", "propB": 1, "valueProp": 1.0}),
			},
		}
		mockStore.On("ListMetersByEventTypes", ctx, []string{"type1"}).Return([]*models.Meter{baseMeter}, nil).Once()
		mockProducer.On("PublishEvents", testTopic, mock.MatchedBy(func(batch *models.EventBatch) bool {
			return len(batch.Events) == 1 && batch.Events[0].ID == "ev2"
		})).Return(nil).Once()

		result, err := service.PublishEvents(ctx, testTopic, events, true)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.SuccessCount)
		assert.Len(t, result.FailedEvents, 1)
		assert.Nil(t, result.FailedEvents[0].Event)
		var failedAppErr *domainerrors.AppError
		assert.True(t, errors.As(result.FailedEvents[0].Error, &failedAppErr))
		assert.Equal(t, domainerrors.EINVALID, domainerrors.ErrorCode(failedAppErr.Code))
		assert.Equal(t, "nil event in batch", failedAppErr.Message)

		mockStore.AssertExpectations(t)
		mockProducer.AssertExpectations(t)
	})

	t.Run("validation fails - nil event, allowPartialSuccess=false", func(t *testing.T) {
		mockProducer := new(MockProducerRepository)
		mockStore := new(MockMeterStoreRepository)
		service := NewProducerService(mockProducer, mockStore)

		events := &models.EventBatch{
			Events: []*models.Event{
				nil,
				newTestEvent("ev2", "type1", map[string]any{"propA": "valA", "propB": 1, "valueProp": 1.0}),
			},
		}
		mockStore.On("ListMetersByEventTypes", ctx, []string{"type1"}).Return([]*models.Meter{baseMeter}, nil).Once()

		result, err := service.PublishEvents(ctx, testTopic, events, false)

		assert.Error(t, err)
		var appErr *domainerrors.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, domainerrors.EINVALID, domainerrors.ErrorCode(appErr.Code))
		assert.Equal(t, "nil event in batch", appErr.Message)
		assert.Nil(t, result)
		mockStore.AssertExpectations(t)
		mockProducer.AssertNotCalled(t, "PublishEvents", mock.Anything, mock.Anything)
	})

	t.Run("validation fails - empty event type, allowPartialSuccess=true", func(t *testing.T) {
		mockProducer := new(MockProducerRepository)
		mockStore := new(MockMeterStoreRepository)
		service := NewProducerService(mockProducer, mockStore)

		events := &models.EventBatch{
			Events: []*models.Event{
				newTestEvent("ev1", "", map[string]any{"propA": "valA"}), // Empty type
				newTestEvent("ev2", "type1", map[string]any{"propA": "valA", "propB": 1, "valueProp": 1.0}),
			},
		}
		mockStore.On("ListMetersByEventTypes", ctx, []string{"type1"}).Return([]*models.Meter{baseMeter}, nil).Once()
		mockProducer.On("PublishEvents", testTopic, mock.MatchedBy(func(batch *models.EventBatch) bool {
			return len(batch.Events) == 1 && batch.Events[0].ID == "ev2"
		})).Return(nil).Once()

		result, err := service.PublishEvents(ctx, testTopic, events, true)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.SuccessCount)
		assert.Len(t, result.FailedEvents, 1)
		assert.Equal(t, "ev1", result.FailedEvents[0].Event.ID)
		var failedAppErr *domainerrors.AppError
		assert.True(t, errors.As(result.FailedEvents[0].Error, &failedAppErr))
		assert.Equal(t, domainerrors.EINVALID, domainerrors.ErrorCode(failedAppErr.Code))
		assert.Equal(t, "missing event type", failedAppErr.Message)

		mockStore.AssertExpectations(t)
		mockProducer.AssertExpectations(t)
	})

	t.Run("validation fails - no meter configured for event type, allowPartialSuccess=true", func(t *testing.T) {
		mockProducer := new(MockProducerRepository)
		mockStore := new(MockMeterStoreRepository)
		service := NewProducerService(mockProducer, mockStore)

		events := &models.EventBatch{
			Events: []*models.Event{
				newTestEvent("ev1", "unknown-type", map[string]any{"propA": "valA"}),
				newTestEvent("ev2", "type1", map[string]any{"propA": "valA", "propB": 1, "valueProp": 1.0}),
			},
		}
		mockStore.On("ListMetersByEventTypes", ctx, mock.AnythingOfType("[]string")).Run(func(args mock.Arguments) {
			eventTypes := args.Get(1).([]string)
			assert.ElementsMatch(t, []string{"unknown-type", "type1"}, eventTypes)
		}).Return([]*models.Meter{baseMeter /* only for type1 */}, nil).Once()

		mockProducer.On("PublishEvents", testTopic, mock.MatchedBy(func(batch *models.EventBatch) bool {
			return len(batch.Events) == 1 && batch.Events[0].ID == "ev2"
		})).Return(nil).Once()

		result, err := service.PublishEvents(ctx, testTopic, events, true)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.SuccessCount)
		assert.Len(t, result.FailedEvents, 1)
		assert.Equal(t, "ev1", result.FailedEvents[0].Event.ID)
		var failedAppErr *domainerrors.AppError
		assert.True(t, errors.As(result.FailedEvents[0].Error, &failedAppErr))
		assert.Equal(t, domainerrors.EINVALID, domainerrors.ErrorCode(failedAppErr.Code))
		assert.Equal(t, "missing meter configuration", failedAppErr.Message)
		assert.Contains(t, failedAppErr.Internal, "no meter configured for event type: unknown-type")

		mockStore.AssertExpectations(t)
		mockProducer.AssertExpectations(t)
	})

	t.Run("validation fails - missing required property, allowPartialSuccess=false", func(t *testing.T) {
		mockProducer := new(MockProducerRepository)
		mockStore := new(MockMeterStoreRepository)
		service := NewProducerService(mockProducer, mockStore)

		events := &models.EventBatch{
			Events: []*models.Event{
				newTestEvent("ev1", "type1", map[string]any{"propA": "valA"}),
			},
		}
		mockStore.On("ListMetersByEventTypes", ctx, []string{"type1"}).Return([]*models.Meter{baseMeter}, nil).Once()

		result, err := service.PublishEvents(ctx, testTopic, events, false)

		assert.Error(t, err)
		var appErr *domainerrors.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, domainerrors.EINVALID, domainerrors.ErrorCode(appErr.Code))
		assert.Equal(t, "missing required event properties", appErr.Message)
		assert.Contains(t, appErr.Internal, "missing or empty required properties")
		assert.Nil(t, result)
		mockStore.AssertExpectations(t)
		mockProducer.AssertNotCalled(t, "PublishEvents", mock.Anything, mock.Anything)
	})

	t.Run("validation fails - empty string for required property, allowPartialSuccess=true", func(t *testing.T) {
		mockProducer := new(MockProducerRepository)
		mockStore := new(MockMeterStoreRepository)
		service := NewProducerService(mockProducer, mockStore)

		events := &models.EventBatch{
			Events: []*models.Event{
				newTestEvent("ev1", "type1", map[string]any{"propA": "", "propB": 123, "valueProp": 10.5}),
				newTestEvent("ev2", "type1", map[string]any{"propA": "valA", "propB": 1, "valueProp": 1.0}),
			},
		}
		mockStore.On("ListMetersByEventTypes", ctx, []string{"type1"}).Return([]*models.Meter{baseMeter}, nil).Once()
		mockProducer.On("PublishEvents", testTopic, mock.MatchedBy(func(batch *models.EventBatch) bool {
			return len(batch.Events) == 1 && batch.Events[0].ID == "ev2"
		})).Return(nil).Once()

		result, err := service.PublishEvents(ctx, testTopic, events, true)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.SuccessCount)
		assert.Len(t, result.FailedEvents, 1)
		assert.Equal(t, "ev1", result.FailedEvents[0].Event.ID)
		var failedAppErr *domainerrors.AppError
		assert.True(t, errors.As(result.FailedEvents[0].Error, &failedAppErr))
		assert.Equal(t, domainerrors.EINVALID, domainerrors.ErrorCode(failedAppErr.Code))
		assert.Equal(t, "missing required event properties", failedAppErr.Message)
		assert.Contains(t, failedAppErr.Internal, "missing or empty required properties: [propA]")

		mockStore.AssertExpectations(t)
		mockProducer.AssertExpectations(t)
	})

	t.Run("validation fails - invalid event properties JSON, allowPartialSuccess=true", func(t *testing.T) {
		mockProducer := new(MockProducerRepository)
		mockStore := new(MockMeterStoreRepository)
		service := NewProducerService(mockProducer, mockStore)

		invalidPropsEvent := newTestEvent("ev1", "type1", nil)
		invalidPropsEvent.Properties = "{not-a-valid-json"

		events := &models.EventBatch{
			Events: []*models.Event{
				invalidPropsEvent,
				newTestEvent("ev2", "type1", map[string]any{"propA": "valA", "propB": 1, "valueProp": 1.0}),
			},
		}
		mockStore.On("ListMetersByEventTypes", ctx, []string{"type1"}).Return([]*models.Meter{baseMeter}, nil).Once()
		mockProducer.On("PublishEvents", testTopic, mock.MatchedBy(func(batch *models.EventBatch) bool {
			return len(batch.Events) == 1 && batch.Events[0].ID == "ev2"
		})).Return(nil).Once()

		result, err := service.PublishEvents(ctx, testTopic, events, true)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.SuccessCount)
		assert.Len(t, result.FailedEvents, 1)
		assert.Equal(t, "ev1", result.FailedEvents[0].Event.ID)
		var failedAppErr *domainerrors.AppError
		assert.True(t, errors.As(result.FailedEvents[0].Error, &failedAppErr))
		assert.Equal(t, domainerrors.EINVALID, domainerrors.ErrorCode(failedAppErr.Code))
		assert.Equal(t, "invalid event properties format", failedAppErr.Message)
		assert.Contains(t, failedAppErr.Internal, "failed to unmarshal event properties")

		mockStore.AssertExpectations(t)
		mockProducer.AssertExpectations(t)
	})

	t.Run("all events fail validation, allowPartialSuccess=true", func(t *testing.T) {
		mockProducer := new(MockProducerRepository)
		mockStore := new(MockMeterStoreRepository)
		service := NewProducerService(mockProducer, mockStore)

		events := &models.EventBatch{
			Events: []*models.Event{
				newTestEvent("ev1", "type1", map[string]any{"propA": "valA"}),
				newTestEvent("ev2", "type1", map[string]any{"propB": 123}),
			},
		}
		mockStore.On("ListMetersByEventTypes", ctx, []string{"type1"}).Return([]*models.Meter{baseMeter}, nil).Once()

		result, err := service.PublishEvents(ctx, testTopic, events, true)

		assert.Error(t, err)
		var appErr *domainerrors.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, domainerrors.EINVALID, domainerrors.ErrorCode(appErr.Code))
		assert.Equal(t, "all events failed validation", appErr.Message)
		assert.Contains(t, appErr.Internal, "all 2 events failed validation")
		assert.Nil(t, result)
		mockStore.AssertExpectations(t)
		mockProducer.AssertNotCalled(t, "PublishEvents", mock.Anything, mock.Anything)
	})

	t.Run("all events fail validation (one nil, one missing prop), allowPartialSuccess=true", func(t *testing.T) {
		mockProducer := new(MockProducerRepository)
		mockStore := new(MockMeterStoreRepository)
		service := NewProducerService(mockProducer, mockStore)

		events := &models.EventBatch{
			Events: []*models.Event{
				nil,
				newTestEvent("ev2", "type1", map[string]any{}),
			},
		}
		mockStore.On("ListMetersByEventTypes", ctx, []string{"type1"}).Return([]*models.Meter{baseMeter}, nil).Once()

		result, err := service.PublishEvents(ctx, testTopic, events, true)

		assert.Error(t, err)
		var appErr *domainerrors.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, domainerrors.EINVALID, domainerrors.ErrorCode(appErr.Code))
		assert.Equal(t, "all events failed validation", appErr.Message)
		assert.Contains(t, appErr.Internal, "all 2 events failed validation")
		assert.Nil(t, result)
		mockStore.AssertExpectations(t)
		mockProducer.AssertNotCalled(t, "PublishEvents", mock.Anything, mock.Anything)
	})

	t.Run("producer PublishEvents fails", func(t *testing.T) {
		mockProducer := new(MockProducerRepository)
		mockStore := new(MockMeterStoreRepository)
		service := NewProducerService(mockProducer, mockStore)

		events := &models.EventBatch{
			Events: []*models.Event{
				newTestEvent("ev1", "type1", map[string]any{"propA": "valA", "propB": 123, "valueProp": 10.5}),
			},
		}
		producerErr := errors.New("kafka unavailable")

		mockStore.On("ListMetersByEventTypes", ctx, []string{"type1"}).Return([]*models.Meter{baseMeter}, nil).Once()
		mockProducer.On("PublishEvents", testTopic, mock.AnythingOfType("*models.EventBatch")).Return(producerErr).Once()

		result, err := service.PublishEvents(ctx, testTopic, events, true)

		assert.Error(t, err)
		var appErr *domainerrors.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, domainerrors.EINTERNAL, domainerrors.ErrorCode(appErr.Code))
		assert.Equal(t, "event publishing failed", appErr.Message)
		assert.Contains(t, appErr.Internal, "failed to publish valid events: kafka unavailable")
		assert.Nil(t, result)
		mockStore.AssertExpectations(t)
		mockProducer.AssertExpectations(t)
	})

	t.Run("meter has no specific properties defined, event has some", func(t *testing.T) {
		mockProducer := new(MockProducerRepository)
		mockStore := new(MockMeterStoreRepository)
		service := NewProducerService(mockProducer, mockStore)

		events := &models.EventBatch{
			Events: []*models.Event{
				newTestEvent("ev3", "type3", map[string]any{"randomProp": "xyz"}),
			},
		}

		mockStore.On("ListMetersByEventTypes", ctx, []string{"type3"}).Return([]*models.Meter{baseMeterNoProps}, nil).Once()
		mockProducer.On("PublishEvents", testTopic, mock.AnythingOfType("*models.EventBatch")).Run(func(args mock.Arguments) {
			batch := args.Get(1).(*models.EventBatch)
			assert.Len(t, batch.Events, 1)
			assert.Equal(t, "ev3", batch.Events[0].ID)
		}).Return(nil).Once()

		result, err := service.PublishEvents(ctx, testTopic, events, true)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.SuccessCount)
		assert.Empty(t, result.FailedEvents)
		mockStore.AssertExpectations(t)
		mockProducer.AssertExpectations(t)
	})

	t.Run("event properties is empty string, but required props exist for meter", func(t *testing.T) {
		mockProducer := new(MockProducerRepository)
		mockStore := new(MockMeterStoreRepository)
		service := NewProducerService(mockProducer, mockStore)

		eventWithEmptyPropsString := newTestEvent("ev-empty-props", "type1", nil)
		eventWithEmptyPropsString.Properties = ""

		events := &models.EventBatch{
			Events: []*models.Event{eventWithEmptyPropsString},
		}
		mockStore.On("ListMetersByEventTypes", ctx, []string{"type1"}).Return([]*models.Meter{baseMeter}, nil).Once()

		result, err := service.PublishEvents(ctx, testTopic, events, false)

		assert.Error(t, err)
		var appErr *domainerrors.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, domainerrors.EINVALID, domainerrors.ErrorCode(appErr.Code))
		assert.Equal(t, "missing required event properties", appErr.Message)
		assert.Contains(t, appErr.Internal, "propA")
		assert.Contains(t, appErr.Internal, "propB")
		assert.Contains(t, appErr.Internal, "valueProp")
		assert.Nil(t, result)

		mockStore.AssertExpectations(t)
		mockProducer.AssertNotCalled(t, "PublishEvents", mock.Anything, mock.Anything)
	})

	t.Run("event properties is literal empty JSON object string, but required props exist", func(t *testing.T) {
		mockProducer := new(MockProducerRepository)
		mockStore := new(MockMeterStoreRepository)
		service := NewProducerService(mockProducer, mockStore)

		eventWithEmptyJSONObject := newTestEvent("ev-empty-json", "type1", nil)
		eventWithEmptyJSONObject.Properties = "{}"

		events := &models.EventBatch{
			Events: []*models.Event{eventWithEmptyJSONObject},
		}
		mockStore.On("ListMetersByEventTypes", ctx, []string{"type1"}).Return([]*models.Meter{baseMeter}, nil).Once()

		result, err := service.PublishEvents(ctx, testTopic, events, false)

		assert.Error(t, err)
		var appErr *domainerrors.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, domainerrors.EINVALID, domainerrors.ErrorCode(appErr.Code))
		assert.Equal(t, "missing required event properties", appErr.Message)
		assert.Contains(t, appErr.Internal, "propA")
		assert.Contains(t, appErr.Internal, "propB")
		assert.Contains(t, appErr.Internal, "valueProp")
		assert.Nil(t, result)

		mockStore.AssertExpectations(t)
		mockProducer.AssertNotCalled(t, "PublishEvents", mock.Anything, mock.Anything)
	})
}
