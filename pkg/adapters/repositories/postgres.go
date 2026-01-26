package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/femisowemimo/booking-appointment/backend/pkg/core/domain"
	_ "github.com/lib/pq" // Postgres driver
)

type PostgresReservationRepository struct {
	db *sql.DB
}

func NewPostgresReservationRepository(db *sql.DB) *PostgresReservationRepository {
	return &PostgresReservationRepository{db: db}
}

func (r *PostgresReservationRepository) Save(ctx context.Context, res *domain.Reservation) error {
	query := `
		INSERT INTO reservations (id, user_id, event_id, start_time, end_time, ticket_count, status, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		res.ID, res.UserID, res.EventID, res.StartTime, res.EndTime, res.TicketCount, res.Status, res.Version, res.CreatedAt, res.UpdatedAt,
	)
	return err
}

func (r *PostgresReservationRepository) GetByID(ctx context.Context, id string) (*domain.Reservation, error) {
	query := `
		SELECT id, user_id, event_id, start_time, end_time, ticket_count, status, version, created_at, updated_at
		FROM reservations WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var res domain.Reservation
	err := row.Scan(
		&res.ID, &res.UserID, &res.EventID, &res.StartTime, &res.EndTime, &res.TicketCount, &res.Status, &res.Version, &res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &res, nil
}

func (r *PostgresReservationRepository) GetByEventAndRange(ctx context.Context, eventID string, start, end time.Time) ([]*domain.Reservation, error) {
	query := `
		SELECT id, user_id, event_id, start_time, end_time, ticket_count, status, version, created_at, updated_at
		FROM reservations 
		WHERE event_id = $1 AND start_time >= $2 AND start_time < $3 AND status != 'CANCELLED'
		ORDER BY start_time ASC
	`
	rows, err := r.db.QueryContext(ctx, query, eventID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reservations []*domain.Reservation
	for rows.Next() {
		var res domain.Reservation
		if err := rows.Scan(
			&res.ID, &res.UserID, &res.EventID, &res.StartTime, &res.EndTime, &res.TicketCount, &res.Status, &res.Version, &res.CreatedAt, &res.UpdatedAt,
		); err != nil {
			return nil, err
		}
		reservations = append(reservations, &res)
	}
	return reservations, nil
}
