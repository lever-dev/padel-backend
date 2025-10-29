package reservation

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lever-dev/padel-backend/internal/entities"
)

type Repository struct {
	connectionURL string
	pool          *pgxpool.Pool
}

func NewRepository(connectionURL string) *Repository {
	return &Repository{connectionURL: connectionURL}
}

func (r *Repository) Connect(ctx context.Context) error {
	p, err := pgxpool.New(ctx, r.connectionURL)
	if err != nil {
		return fmt.Errorf("pgxpool new: %w", err)
	}

	r.pool = p

	return nil
}

func (r *Repository) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}

func (r *Repository) Create(ctx context.Context, reservation *entities.Reservation) error {
	if r.pool == nil {
		return fmt.Errorf("not connected to pool")
	}

	if reservation.CreatedAt.IsZero() {
		reservation.CreatedAt = time.Now().UTC()
	}

	d := newDTO(reservation)

	_, err := r.pool.Exec(
		ctx,
		createReservationQuery,
		d.ID,
		d.CourtID,
		d.Status,
		d.ReservedFrom,
		d.ReservedTo,
		d.ReservedBy,
		nullableString(d.CancelledBy),
		d.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	return nil
}

const createReservationQuery = `
INSERT INTO reservations (
    id,
    court_id,
    status,
    reserved_from,
    reserved_to,
    reserved_by,
    cancelled_by,
    created_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
`

func (r *Repository) ListByCourtAndTimeRange(
	ctx context.Context,
	courtID string,
	from, to time.Time,
) ([]entities.Reservation, error) {
	if r.pool == nil {
		return nil, fmt.Errorf("not connected to pool")
	}

	rows, err := r.pool.Query(ctx, listReservationsQuery, courtID, from, to)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var results []entities.Reservation

	for rows.Next() {
		rsv, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan reservation: %w", err)
		}

		results = append(results, rsv)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return results, nil
}

const listReservationsQuery = `
SELECT
    id,
    court_id,
    status,
    reserved_from,
    reserved_to,
    reserved_by,
    cancelled_by,
    created_at
FROM reservations
WHERE court_id = $1
    AND reserved_from < $3
    AND reserved_to > $2
ORDER BY reserved_from ASC
`

func (r *Repository) GetByID(ctx context.Context, reservationID string) (*entities.Reservation, error) {
	if r.pool == nil {
		return nil, fmt.Errorf("not connected to pool")
	}

	rsv, err := scan(r.pool.QueryRow(ctx, getReservationByIDQuery, reservationID))
	if err != nil {
		return nil, fmt.Errorf("scan reservation: %w", err)
	}

	return &rsv, nil
}

const getReservationByIDQuery = `
	SELECT
	    id,
	    court_id,
	    status,
	    reserved_from,
	    reserved_to,
	    reserved_by,
	    cancelled_by,
	    created_at
	FROM reservations
	WHERE id = $1
	LIMIT 1
`

func (r *Repository) CancelReservation(
	ctx context.Context,
	reservationID string,
	cancelledByUser string,
) error {
	if r.pool == nil {
		return fmt.Errorf("not connected to pool")
	}

	tag, err := r.pool.Exec(
		ctx,
		cancelReservationQuery,
		entities.CancelledReservationStatus,
		cancelledByUser,
		reservationID,
	)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return entities.ErrNotFound
	}

	return nil
}

const cancelReservationQuery = `
UPDATE reservations
SET status = $1,
    cancelled_by = $2
WHERE id = $3
`

func nullableString(s string) any {
	if s == "" {
		return nil
	}

	return s
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scan(scanner rowScanner) (entities.Reservation, error) {
	var (
		d           dto
		cancelledBy sql.NullString
	)

	err := scanner.Scan(
		&d.ID,
		&d.CourtID,
		&d.Status,
		&d.ReservedFrom,
		&d.ReservedTo,
		&d.ReservedBy,
		&cancelledBy,
		&d.CreatedAt,
	)
	if err != nil {
		return entities.Reservation{}, err
	}

	if cancelledBy.Valid {
		d.CancelledBy = cancelledBy.String
	}

	d.ReservedFrom = d.ReservedFrom.UTC()
	d.ReservedTo = d.ReservedTo.UTC()
	d.CreatedAt = d.CreatedAt.UTC()

	return d.toEntity(), nil
}
