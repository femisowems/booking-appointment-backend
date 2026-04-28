package repositories

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/femisowemimo/booking-appointment/backend/pkg/core/domain"
)

type InMemoryReservationRepository struct {
	mu           sync.RWMutex
	reservations map[string]*domain.Reservation
}

func NewInMemoryReservationRepository() *InMemoryReservationRepository {
	return &InMemoryReservationRepository{
		reservations: map[string]*domain.Reservation{},
	}
}

func (r *InMemoryReservationRepository) Save(_ context.Context, res *domain.Reservation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copyRes := *res
	r.reservations[res.ID] = &copyRes
	return nil
}

func (r *InMemoryReservationRepository) GetByID(_ context.Context, id string) (*domain.Reservation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res, ok := r.reservations[id]
	if !ok {
		return nil, nil
	}

	copyRes := *res
	return &copyRes, nil
}

func (r *InMemoryReservationRepository) GetByEventAndRange(_ context.Context, eventID string, start, end time.Time) ([]*domain.Reservation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make([]*domain.Reservation, 0)
	for _, res := range r.reservations {
		if res.EventID != eventID {
			continue
		}
		if res.Status == domain.StatusCancelled {
			continue
		}
		if res.StartTime.Before(start) || !res.StartTime.Before(end) {
			continue
		}

		copyRes := *res
		results = append(results, &copyRes)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].StartTime.Before(results[j].StartTime)
	})

	return results, nil
}