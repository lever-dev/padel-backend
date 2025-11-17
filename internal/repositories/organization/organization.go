package organization

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
	return &Repository{
		connectionURL: connectionURL,
	}
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

func (r *Repository) Create(ctx context.Context, organization *entities.Organization) error {
	if r.pool == nil {
		return fmt.Errorf("not connected to pool")
	}

	if organization.CreatedAt.IsZero() {
		organization.CreatedAt = time.Now().UTC()
	}

	d := newDTO(organization)

	_, err := r.pool.Exec(
		ctx,
		createOrganizationQuery,
		d.ID,
		d.Name,
		d.City,
		d.CreatedAt,
		d.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	return nil
}

const createOrganizationQuery = `
	INSERT INTO organizations(
		id,
		name,
		city,
		created_at,
		updated_at
	) VALUES ($1, $2, $3, $4, $5)
`

func (r *Repository) GetByID(ctx context.Context, organizationID string) (*entities.Organization, error) {
	if r.pool == nil {
		return nil, fmt.Errorf("not connected to pool")
	}

	org, err := scan(r.pool.QueryRow(ctx, getOrganizationByIDQuery, organizationID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entities.ErrNotFound
		}
		return nil, fmt.Errorf("scan organization: %w", err)
	}

	return &org, nil
}

const getOrganizationByIDQuery = `
	SELECT 
		id,
		name,
		city,
		created_at,
		updated_at
	FROM organizations
	WHERE id = $1
	LIMIT 1
`

func (r *Repository) GetOrganizationsByCity(ctx context.Context, city string) ([]entities.Organization, error) {
	if r.pool == nil {
		return nil, fmt.Errorf("not connected to pool")
	}

	rows, err := r.pool.Query(ctx, getOrganizationsByCityQuery, city)
	if err != nil {
		return nil, fmt.Errorf("get organizations by city %q: %w", city, err)
	}
	defer rows.Close()

	var results []entities.Organization

	for rows.Next() {
		org, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan organization: %w", err)
		}

		results = append(results, org)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return results, nil
}

const getOrganizationsByCityQuery = `
	SELECT
		id,
		name,
		city,
		created_at,
		updated_at
	FROM organizations
	WHERE city = $1
`

func (r *Repository) Update(ctx context.Context, org *entities.Organization) error {
	if r.pool == nil {
		return fmt.Errorf("not connected to pool")
	}

	tag, err := r.pool.Exec(
		ctx,
		updateOrganizationQuery,
		org.Name,
		org.City,
		org.UpdatedAt,
		org.ID,
	)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return entities.ErrNotFound
	}

	return nil
}

const updateOrganizationQuery = `
	UPDATE organizations
	SET 
    	name = $1,
    	city = $2,
		updated_at = $3
	WHERE id = $4
`

func (r *Repository) Delete(ctx context.Context, id string) error {
	if r.pool == nil {
		return fmt.Errorf("not connected to pool")
	}

	tag, err := r.pool.Exec(ctx, deleteOrganizationQuery, id)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return entities.ErrNotFound
	}

	return nil
}

const deleteOrganizationQuery = `
	DELETE FROM organizations
	WHERE id = $1
`

type rowScanner interface {
	Scan(dest ...any) error
}

func scan(scanner rowScanner) (entities.Organization, error) {
	var d dto

	err := scanner.Scan(
		&d.ID,
		&d.Name,
		&d.City,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err != nil {
		return entities.Organization{}, err
	}

	d.CreatedAt = d.CreatedAt.UTC()
	d.UpdatedAt = d.UpdatedAt.UTC()

	return d.toEntity(), nil
}
