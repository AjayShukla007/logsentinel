package team

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Team struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	OwnerID   string    `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Repository interface {
	Create(ctx context.Context, name, ownerID string) (*Team, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Team, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, name, ownerID string) (*Team, error) {
	team := &Team{
		ID:      uuid.New(),
		Name:    name,
		OwnerID: ownerID,
	}

	query := `
		INSERT INTO teams (id, name, owner_id) 
		VALUES ($1, $2, $3) 
		RETURNING id, name, owner_id, created_at
	`

	err := r.db.QueryRow(ctx, query, team.ID, team.Name, team.OwnerID).Scan(
		&team.ID, &team.Name, &team.OwnerID, &team.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*Team, error) {
	team := &Team{}

	query := `
		SELECT id, name, owner_id, created_at 
		FROM teams 
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&team.ID, &team.Name, &team.OwnerID, &team.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM teams WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
