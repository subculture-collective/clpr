package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stripe/stripe-go/v81"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// MockSubscriptionRepository is a mock implementation of SubscriptionRepositoryInterface
type MockSubscriptionRepository struct {
	mock.Mock
}

func (m *MockSubscriptionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Subscription, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) GetByStripeCustomerID(ctx context.Context, customerID string) (*models.Subscription, error) {
	args := m.Called(ctx, customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) GetByStripeSubscriptionID(ctx context.Context, subscriptionID string) (*models.Subscription, error) {
	args := m.Called(ctx, subscriptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) Create(ctx context.Context, sub *models.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) Update(ctx context.Context, sub *models.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) GetEventByStripeEventID(ctx context.Context, eventID string) (*models.SubscriptionEvent, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SubscriptionEvent), args.Error(1)
}

func (m *MockSubscriptionRepository) LogSubscriptionEvent(ctx context.Context, subscriptionID *uuid.UUID, eventType string, stripeEventID *string, eventData interface{}) error {
	args := m.Called(ctx, subscriptionID, eventType, stripeEventID, eventData)
	return args.Error(0)
}

// MockUserRepository is a mock implementation of UserRepositoryInterface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// MockWebhookRepository is a mock implementation of WebhookRepositoryInterface
type MockWebhookRepository struct {
	mock.Mock
}

func (m *MockWebhookRepository) AddToRetryQueue(ctx context.Context, stripeEventID string, eventType string, payload interface{}, maxRetries int) error {
	args := m.Called(ctx, stripeEventID, eventType, payload, maxRetries)
	return args.Error(0)
}

func (m *MockWebhookRepository) GetPendingRetries(ctx context.Context, limit int) ([]*models.WebhookRetryQueue, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.WebhookRetryQueue), args.Error(1)
}

func (m *MockWebhookRepository) UpdateRetryQueueItem(ctx context.Context, id uuid.UUID, retryCount int, nextRetryAt *time.Time, lastError string) error {
	args := m.Called(ctx, id, retryCount, nextRetryAt, lastError)
	return args.Error(0)
}

func (m *MockWebhookRepository) RemoveFromRetryQueue(ctx context.Context, stripeEventID string) error {
	args := m.Called(ctx, stripeEventID)
	return args.Error(0)
}

// MockAuditLogService is a mock implementation of AuditLogService for testing
type MockAuditLogService struct {
	mock.Mock
}

func (m *MockAuditLogService) LogSubscriptionEvent(ctx context.Context, userID uuid.UUID, eventType string, metadata map[string]interface{}) error {
	args := m.Called(ctx, userID, eventType, metadata)
	return args.Error(0)
}

// MockDunningService is a mock implementation of DunningService for testing
type MockDunningService struct {
	mock.Mock
}

func (m *MockDunningService) HandlePaymentSuccess(ctx context.Context, invoice *stripe.Invoice) error {
	args := m.Called(ctx, invoice)
	return args.Error(0)
}

func (m *MockDunningService) HandlePaymentFailure(ctx context.Context, invoice *stripe.Invoice) error {
	args := m.Called(ctx, invoice)
	return args.Error(0)
}

// MockEmailService is a mock implementation of EmailService for testing
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendNotificationEmail(ctx context.Context, user *models.User, notificationType string, notificationID uuid.UUID, data map[string]interface{}) error {
	args := m.Called(ctx, user, notificationType, notificationID, data)
	return args.Error(0)
}

func (m *MockEmailService) SendDisputeNotification(ctx context.Context, user *models.User, dispute *stripe.Dispute) error {
	args := m.Called(ctx, user, dispute)
	return args.Error(0)
}
