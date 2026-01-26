package services

import (
	"context"
	"time"

	"github.com/femisowemimo/booking-appointment/backend/pkg/core/domain"
	"github.com/femisowemimo/booking-appointment/backend/pkg/core/ports"
	"github.com/google/uuid"
)

type ReservationService struct {
	repo      ports.ReservationRepository
	publisher ports.EventPublisher
}

func NewReservationService(repo ports.ReservationRepository, publisher ports.EventPublisher) *ReservationService {
	return &ReservationService{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *ReservationService) Create(ctx context.Context, userID, eventID string, start, end time.Time, ticketCount int) (*domain.Reservation, error) {
	// 1. Create Domain Entity (Validation happens here)
	res, err := domain.NewReservation(userID, eventID, start, end, ticketCount)
	if err != nil {
		return nil, err
	}
	res.ID = uuid.New().String()

	// 2. Check Availability (Simplified)
	// Assuming DB constraints handle concurrency/overlap.

	// 3. Persist to DB
	if err := s.repo.Save(ctx, res); err != nil {
		return nil, err
	}

	// 4. Publish Event
	if s.publisher != nil {
		event := struct {
			EventID       string    `json:"event_id"`
			EventType     string    `json:"event_type"`
			ReservationID string    `json:"reservation_id"`
			UserID        string    `json:"user_id"`
			TicketCount   int       `json:"ticket_count"`
			Status        string    `json:"status"`
			Timestamp     time.Time `json:"timestamp"`
		}{
			EventID:       res.EventID, // The actual event (e.g., concert id)
			EventType:     "ReservationCreated",
			ReservationID: res.ID,
			UserID:        res.UserID,
			TicketCount:   res.TicketCount,
			Status:        string(res.Status),
			Timestamp:     time.Now(),
		}

		if err := s.publisher.Publish(ctx, event); err != nil {
			// In production: return success but log error
			return nil, err
		}
	}

	return res, nil
}

func (s *ReservationService) Get(ctx context.Context, id string) (*domain.Reservation, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ReservationService) ListByEvent(ctx context.Context, eventID string, start, end time.Time) ([]*domain.Reservation, error) {
	return s.repo.GetByEventAndRange(ctx, eventID, start, end)
}
