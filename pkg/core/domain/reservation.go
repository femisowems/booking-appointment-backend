package domain

import (
	"errors"
	"time"
)

type ReservationStatus string

const (
	StatusBooked    ReservationStatus = "BOOKED"
	StatusCheckedIn ReservationStatus = "CHECKED_IN"
	StatusCancelled ReservationStatus = "CANCELLED"
	StatusCompleted ReservationStatus = "COMPLETED"
)

var (
	ErrInvalidTime        = errors.New("invalid reservation time")
	ErrPastTime           = errors.New("cannot make reservation in the past")
	ErrDuration           = errors.New("reservation duration must be positive")
	ErrInvalidTicketCount = errors.New("ticket count must be between 1 and 6")
	ErrCheckInNotAllowed  = errors.New("reservation cannot be checked in in its current state")
)

type Reservation struct {
	ID          string            `json:"id"`
	UserID      string            `json:"user_id"`
	EventID     string            `json:"event_id"`
	StartTime   time.Time         `json:"start_time"`
	EndTime     time.Time         `json:"end_time"`
	TicketCount int               `json:"ticket_count"`
	Status      ReservationStatus `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Version     int               `json:"version"` // Optimistic locking
}

func NewReservation(userID, eventID string, start, end time.Time, ticketCount int) (*Reservation, error) {
	if start.After(end) {
		return nil, ErrInvalidTime
	}
	if start.Before(time.Now()) {
		return nil, ErrPastTime
	}

	duration := end.Sub(start)
	if duration <= 0 {
		return nil, ErrDuration
	}

	if ticketCount < 1 || ticketCount > 6 {
		return nil, ErrInvalidTicketCount
	}

	return &Reservation{
		UserID:      userID,
		EventID:     eventID,
		StartTime:   start,
		EndTime:     end,
		TicketCount: ticketCount,
		Status:      StatusBooked,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Version:     1,
	}, nil
}

func (r *Reservation) Cancel() {
	r.Status = StatusCancelled
	r.UpdatedAt = time.Now()
}

func (r *Reservation) CheckIn() error {
	switch r.Status {
	case StatusCancelled, StatusCompleted:
		return ErrCheckInNotAllowed
	case StatusCheckedIn:
		return nil
	default:
		r.Status = StatusCheckedIn
		r.Version++
		r.UpdatedAt = time.Now()
		return nil
	}
}
