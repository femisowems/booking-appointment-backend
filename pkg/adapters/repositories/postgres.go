package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/femisowemimo/booking-appointment/backend/pkg/core/domain"
	_ "github.com/lib/pq" // Postgres driver
)

type PostgresAppointmentRepository struct {
	db *sql.DB
}

func NewPostgresAppointmentRepository(db *sql.DB) *PostgresAppointmentRepository {
	return &PostgresAppointmentRepository{db: db}
}

func (r *PostgresAppointmentRepository) Save(ctx context.Context, a *domain.Appointment) error {
	query := `
		INSERT INTO appointments (id, user_id, provider_id, start_time, end_time, status, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(ctx, query,
		a.ID, a.UserID, a.ProviderID, a.StartTime, a.EndTime, a.Status, a.Version, a.CreatedAt, a.UpdatedAt,
	)
	return err
}

func (r *PostgresAppointmentRepository) GetByID(ctx context.Context, id string) (*domain.Appointment, error) {
	query := `
		SELECT id, user_id, provider_id, start_time, end_time, status, version, created_at, updated_at
		FROM appointments WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)
	
	var a domain.Appointment
	err := row.Scan(
		&a.ID, &a.UserID, &a.ProviderID, &a.StartTime, &a.EndTime, &a.Status, &a.Version, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Or return specific ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *PostgresAppointmentRepository) GetByProviderAndRange(ctx context.Context, providerID string, start, end time.Time) ([]*domain.Appointment, error) {
	query := `
		SELECT id, user_id, provider_id, start_time, end_time, status, version, created_at, updated_at
		FROM appointments 
		WHERE provider_id = $1 AND start_time >= $2 AND start_time < $3 AND status != 'CANCELLED'
		ORDER BY start_time ASC
	`
	rows, err := r.db.QueryContext(ctx, query, providerID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appointments []*domain.Appointment
	for rows.Next() {
		var a domain.Appointment
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.ProviderID, &a.StartTime, &a.EndTime, &a.Status, &a.Version, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		appointments = append(appointments, &a)
	}
	return appointments, nil
}
