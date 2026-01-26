package ports

import (
	"context"
	"time"

	"github.com/femisowemimo/booking-appointment/backend/pkg/core/domain"
)

type ReservationRepository interface {
	Save(ctx context.Context, reservation *domain.Reservation) error
	GetByID(ctx context.Context, id string) (*domain.Reservation, error)
	GetByEventAndRange(ctx context.Context, eventID string, start, end time.Time) ([]*domain.Reservation, error)
}

type EventPublisher interface {
	Publish(ctx context.Context, event interface{}) error
}

type ReservationService interface {
	Create(ctx context.Context, userID, eventID string, start, end time.Time, ticketCount int) (*domain.Reservation, error)
	Get(ctx context.Context, id string) (*domain.Reservation, error)
	ListByEvent(ctx context.Context, eventID string, start, end time.Time) ([]*domain.Reservation, error)
}
