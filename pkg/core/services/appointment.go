package services

import (
	"context"
	"time"

	"github.com/femisowemimo/booking-appointment/backend/pkg/core/domain"
	"github.com/femisowemimo/booking-appointment/backend/pkg/core/ports"
	"github.com/google/uuid"
)

type AppointmentService struct {
	repo      ports.AppointmentRepository
	publisher ports.EventPublisher
}

func NewAppointmentService(repo ports.AppointmentRepository, publisher ports.EventPublisher) *AppointmentService {
	return &AppointmentService{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *AppointmentService) Create(ctx context.Context, userID, providerID string, start, end time.Time) (*domain.Appointment, error) {
	// 1. Create Domain Entity (Validation happens here)
	appt, err := domain.NewAppointment(userID, providerID, start, end)
	if err != nil {
		return nil, err
	}
	appt.ID = uuid.New().String()

	// 2. Check Availability (Simplified: Rely on DB constraints or Repository check)
	// In a real system, we might query the repo here to check for overlaps manually if DB constraints aren't enough.
	// For this demo, assuming DB constraints handle concurrency/overlap.

	// 3. Persist to DB
	if err := s.repo.Save(ctx, appt); err != nil {
		return nil, err
	}

	// 4. Publish Event
	if s.publisher != nil {
		event := struct {
			EventID       string    `json:"event_id"`
			EventType     string    `json:"event_type"`
			AppointmentID string    `json:"appointment_id"`
			UserID        string    `json:"user_id"`
			ProviderID    string    `json:"provider_id"`
			Timestamp     time.Time `json:"timestamp"`
		}{
			EventID:       uuid.New().String(),
			EventType:     "AppointmentCreated",
			AppointmentID: appt.ID,
			UserID:        appt.UserID,
			ProviderID:    appt.ProviderID,
			Timestamp:     time.Now(),
		}

		if err := s.publisher.Publish(ctx, event); err != nil {
			// In production: Return success but log CRITICAL error that event wasn't published.
			// Or use a transactional outbox so this never happens.
			return nil, err
		}
	}

	return appt, nil
}

func (s *AppointmentService) Get(ctx context.Context, id string) (*domain.Appointment, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *AppointmentService) ListByProvider(ctx context.Context, providerID string, start, end time.Time) ([]*domain.Appointment, error) {
	return s.repo.GetByProviderAndRange(ctx, providerID, start, end)
}
