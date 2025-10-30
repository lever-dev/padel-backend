package court

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5" 
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
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

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

func (r *Repository) Create(ctx context.Context, court *entities.Court) error {
	if r.pool == nil {
		return fmt.Errorf("not connected to pool")
	}

	if court.CreatedAt.IsZero() {
		court.CreatedAt = time.Now().UTC()
	}

	court.UpdatedAt = time.Now().UTC()

	d := newDTO(court)

	_, err := r.pool.Exec(
		ctx,
		createCourtQuery,
		d.ID,
		d.OrganizationID,
		d.Name,
		d.CreatedAt,
		d.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("exec create court: %w", err)
	}

	return nil
}

const createCourtQuery = `
	INSERT INTO courts(
		id,
		organization_id,
		name,
		created_at,
		updated_at
	) VALUES ($1, $2, $3, $4, $5)
`

func (r *Repository) GetByID(ctx context.Context, court_id string) (*entities.Court, error) {
	if r.pool == nil {
		return nil, fmt.Errorf("not connected to pool")
	}

	crt, err := scan(r.pool.QueryRow(ctx, getCourtByIDQuery, court_id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entities.ErrNotFound
		}
		return nil, fmt.Errorf("scan organization: %w", err)
	}

	return &crt, nil
}

const getCourtByIDQuery = `
	SELECT
		id,
		organization_id,
		name,
		created_at,
		updated_at
	FROM courts
	WHERE id = $1
	LIMIT 1
`

func (r* Repository) ListByOrganizationID(ctx context.Context, organizationID string) ([]entities.Court, error) {
	if r.pool == nil {
		return nil, fmt.Errorf("not connected to pool")
	}

	rows, err := r.pool.Query(ctx, listCourtsByOrganizationIDQuery, organizationID)
	if err != nil {
		return nil, fmt.Errorf("query courts: %w", err)
	}
	defer rows.Close()

	var courts []entities.Court
	for rows.Next() {
		crt, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan court: %w", err)
		}
		courts = append(courts, crt)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows err: %w", rows.Err())
	}

	return courts, nil
}

const listCourtsByOrganizationIDQuery = `
	SELECT
		id,
		organization_id,
		name,
		created_at,
		updated_at
	FROM courts
	WHERE organization_id = $1
	ORDER BY name ASC
`



func (r *Repository) Update(ctx context.Context, crt *entities.Court) error {
	if r.pool == nil {
		return fmt.Errorf("not connected to pool")
	}

	crt.UpdatedAt = time.Now().UTC()

	tag, err := r.pool.Exec(
		ctx,
		updateCourtQuery,
		crt.Name,
		crt.OrganizationID,
		crt.UpdatedAt,
		crt.ID,
	)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return entities.ErrNotFound
	}

	return nil
}

const updateCourtQuery = `
	UPDATE courts
	SET 
    	name = $1,
    	organization_id = $2,
		updated_at = $3
	WHERE id = $4
`

func (r *Repository) Delete(ctx context.Context, id string) error {
	if r.pool == nil {
		return fmt.Errorf("not connected to pool")
	}

	tag, err := r.pool.Exec(ctx, deleteCourtQuery, id)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return entities.ErrNotFound
	}

	return nil
}

const deleteCourtQuery = `
	DELETE FROM courts
	WHERE id = $1
`

type rowScanner interface {
	Scan(dest ...any) error
}

func scan(scanner rowScanner) (entities.Court, error) {
	var d dto

	err := scanner.Scan(
		&d.ID,
		&d.OrganizationID,
		&d.Name,
		&d.CreatedAt,
		&d.UpdatedAt,
)
	if err != nil {
		return entities.Court{}, err
	}

	d.CreatedAt = d.CreatedAt.UTC()
	d.UpdatedAt = d.UpdatedAt.UTC()

	return d.toEntity(), nil
}