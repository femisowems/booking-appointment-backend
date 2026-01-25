package ports

import (
	"context"
	"time"

	"github.com/femisowemimo/booking-appointment/backend/pkg/core/domain"
)

type AppointmentRepository interface {
	Save(ctx context.Context, appointment *domain.Appointment) error
	GetByID(ctx context.Context, id string) (*domain.Appointment, error)
	GetByProviderAndRange(ctx context.Context, providerID string, start, end time.Time) ([]*domain.Appointment, error)
}

type EventPublisher interface {
	Publish(ctx context.Context, event interface{}) error
}

type AppointmentService interface {
	Create(ctx context.Context, userID, providerID string, start, end time.Time) (*domain.Appointment, error)
	Get(ctx context.Context, id string) (*domain.Appointment, error)
	ListByProvider(ctx context.Context, providerID string, start, end time.Time) ([]*domain.Appointment, error)
}
