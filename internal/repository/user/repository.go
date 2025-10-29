package userrepo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	UserID      string    `json:"user_id"`
	ClientID    string    `json:"client_id"`
	AccountType string    `json:"account_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserQuota struct {
	ProjectCount       int
	LogCountLastMinute int
	LastLogTime        *time.Time
}

type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByUserID(ctx context.Context, userID string) (*User, error)
	GetAllUsers(ctx context.Context) ([]*User, error)
	UpdateAccountType(ctx context.Context, userID string, accountType string) error
	UserExists(ctx context.Context, userID, clientID string) (bool, error)
	Delete(ctx context.Context, userID string) error
	GetQuota(ctx context.Context, userID string) (*UserQuota, error)
	GetUserProjectCount(ctx context.Context, userID string) (int, error)
	ValidateUserCredentials(ctx context.Context, clientID, projectID, apiKey string) (
		bool, string, error)
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, user *User) error {
	query := `
        INSERT INTO users (id, user_id, client_id, account_type)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, updated_at
    `
	err := r.db.QueryRow(ctx, query,
		uuid.New(), user.UserID, user.ClientID, user.AccountType,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}
func (r *PostgresRepository) UserExists(ctx context.Context, userID, clientID string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM users 
			WHERE user_id = $1 AND client_id = $2
		)
	`

	err := r.db.QueryRow(ctx, query, userID, clientID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
func (r *PostgresRepository) ValidateUserCredentials(ctx context.Context, clientID, projectID, apiKey string) (bool, string, error) {
	var accountType string

	err := r.db.QueryRow(ctx, `
				SELECT u.account_type
				FROM users u
				JOIN projects p ON p.user_id = u.id
				WHERE u.client_id = $1 AND p.id = $2 AND p.api_key = $3`,
		clientID, projectID, apiKey,
	).Scan(&accountType)

	if err != nil {
		return false, "", err
	}

	return true, accountType, nil
}

func (r *PostgresRepository) GetByUserID(ctx context.Context, userID string) (*User, error) {
	user := &User{}
	query := `
        SELECT id, user_id, client_id, account_type, created_at, updated_at
        FROM users 
        WHERE user_id = $1
    `

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.UserID,
		&user.ClientID,
		&user.AccountType,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *PostgresRepository) GetAllUsers(ctx context.Context) ([]*User, error) {
	query := `
		SELECT id, user_id, client_id, account_type, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID,
			&user.UserID,
			&user.ClientID,
			&user.AccountType,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *PostgresRepository) UpdateAccountType(ctx context.Context, userID string, accountType string) error {
	query := `
        UPDATE users 
        SET account_type = $1, updated_at = CURRENT_TIMESTAMP
        WHERE user_id = $2
    `

	result, err := r.db.Exec(ctx, query, accountType, userID)
	if err != nil {
		return err
	}
	var ErrUserNotFound = errors.New("user not found")
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, userID string) error {
	query := `DELETE FROM users WHERE user_id = $1`

	result, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return err
	}
	var ErrUserNotFound = errors.New("user not found")
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *PostgresRepository) GetQuota(ctx context.Context, userID string) (*UserQuota, error) {
	quota := &UserQuota{}
	query := `
        WITH stats AS (
            SELECT 
                COUNT(DISTINCT p.id) as project_count,
                COUNT(l.*) FILTER (WHERE l.created_at > NOW() - INTERVAL '1 minute') as log_count,
                MAX(l.created_at) as last_log_time
            FROM users u
            LEFT JOIN projects p ON p.user_id = u.id
            LEFT JOIN logs l ON l.project_id = p.id
            WHERE u.user_id = $1
        )
        SELECT project_count, log_count, last_log_time FROM stats
    `

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&quota.ProjectCount,
		&quota.LogCountLastMinute,
		&quota.LastLogTime,
	)
	if err != nil {
		return nil, err
	}

	return quota, nil
}

func (r *PostgresRepository) GetUserProjectCount(ctx context.Context, userID string) (int, error) {
	var count int

	quota, err := r.GetQuota(ctx, userID)

	if err != nil {
		return 0, err
	}

	count = quota.ProjectCount

	return count, nil
}
