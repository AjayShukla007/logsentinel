package project

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Project struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	UserID    uuid.UUID `json:"user_id"`
	ApiKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
}

type Repository interface {
	Create(ctx context.Context, name string, userID uuid.UUID, apiKey string) (*Project, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Project, error)
	GetAllProjects(ctx context.Context) ([]*Project, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, name string, userID uuid.UUID, apiKey string) (*Project, error) {
	project := &Project{
		ID:     uuid.New(),
		Name:   name,
		UserID: userID,
		ApiKey: apiKey,
	}

	query := `
		INSERT INTO projects (id, name, user_id, api_key) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, name, user_id, api_key, created_at
	`

	err := r.db.QueryRow(ctx, query, project.ID, project.Name, project.UserID, project.ApiKey).Scan(
		&project.ID, &project.Name, &project.UserID, &project.ApiKey, &project.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (r *PostgresRepository) GetAllProjects(ctx context.Context) ([]*Project, error) {
	query := `
			SELECT id, name, user_id, api_key, created_at 
			FROM projects 
			ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		project := &Project{}
		err := rows.Scan(
			&project.ID, &project.Name, &project.UserID, &project.ApiKey, &project.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, nil
}
func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*Project, error) {
	project := &Project{}

	query := `
		SELECT id, name, user_id, api_key, created_at 
		FROM projects 
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&project.ID, &project.Name, &project.UserID, &project.ApiKey, &project.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (r *PostgresRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Project, error) {
	query := `
		SELECT id, name, user_id, api_key, created_at 
		FROM projects 
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		project := &Project{}
		err := rows.Scan(
			&project.ID, &project.Name, &project.UserID, &project.ApiKey, &project.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM projects WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	var ErrProjectNotFound = errors.New("project not found")
	if result.RowsAffected() == 0 {
		return ErrProjectNotFound
	}

	return nil
}

func (r *PostgresRepository) VerifyApiKey(ctx context.Context, clientID, projectID, apiKey string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 
			FROM projects p
			JOIN users u ON u.id = p.user_id
			WHERE u.client_id = $1 
			AND p.id = $2 
			AND p.api_key = $3
		)
	`

	err := r.db.QueryRow(ctx, query, clientID, projectID, apiKey).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
