package domain

import (
	"errors"
	"time"
)

type AppointmentStatus string

const (
	StatusBooked    AppointmentStatus = "BOOKED"
	StatusCancelled AppointmentStatus = "CANCELLED"
	StatusCompleted AppointmentStatus = "COMPLETED"
)

var (
	ErrInvalidTime  = errors.New("invalid appointment time")
	ErrPastTime     = errors.New("cannot book appointment in the past")
	ErrDuration     = errors.New("appointment duration must be positive")
)

type Appointment struct {
	ID         string            `json:"id"`
	UserID     string            `json:"user_id"`
	ProviderID string            `json:"provider_id"`
	StartTime  time.Time         `json:"start_time"`
	EndTime    time.Time         `json:"end_time"`
	Status     AppointmentStatus `json:"status"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Version    int               `json:"version"` // Optimistic locking
}

func NewAppointment(userID, providerID string, start, end time.Time) (*Appointment, error) {
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

	return &Appointment{
		UserID:     userID,
		ProviderID: providerID,
		StartTime:  start,
		EndTime:    end,
		Status:     StatusBooked,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Version:    1,
	}, nil
}

func (a *Appointment) Cancel() {
	a.Status = StatusCancelled
	a.UpdatedAt = time.Now()
}
