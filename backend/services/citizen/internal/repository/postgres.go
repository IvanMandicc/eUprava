package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"euprava/citizen/internal/model"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
	id            BIGSERIAL PRIMARY KEY,
	jmbg          VARCHAR(13) NOT NULL UNIQUE,
	first_name    TEXT        NOT NULL,
	last_name     TEXT        NOT NULL,
	email         TEXT        NOT NULL UNIQUE,
	password_hash TEXT        NOT NULL,
	role          TEXT        NOT NULL DEFAULT 'citizen',
	address       TEXT        NOT NULL DEFAULT '',
	created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);`

// PostgresUserRepository je Postgres implementacija UserRepository interfejsa.
type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(ctx context.Context, pool *pgxpool.Pool) (*PostgresUserRepository, error) {
	if _, err := pool.Exec(ctx, schema); err != nil {
		return nil, err
	}
	return &PostgresUserRepository{pool: pool}, nil
}

func (r *PostgresUserRepository) Create(ctx context.Context, u *model.User) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO users (jmbg, first_name, last_name, email, password_hash, role, address)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, created_at`,
		u.JMBG, u.FirstName, u.LastName, u.Email, u.PasswordHash, u.Role, u.Address,
	).Scan(&u.ID, &u.CreatedAt)
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	return r.scanOne(r.pool.QueryRow(ctx, selectUser+` WHERE id = $1`, id))
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.scanOne(r.pool.QueryRow(ctx, selectUser+` WHERE email = $1`, email))
}

func (r *PostgresUserRepository) GetByJMBG(ctx context.Context, jmbg string) (*model.User, error) {
	return r.scanOne(r.pool.QueryRow(ctx, selectUser+` WHERE jmbg = $1`, jmbg))
}

func (r *PostgresUserRepository) List(ctx context.Context) ([]model.User, error) {
	rows, err := r.pool.Query(ctx, selectUser+` ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := []model.User{}
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.JMBG, &u.FirstName, &u.LastName, &u.Email, &u.PasswordHash, &u.Role, &u.Address, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

const selectUser = `SELECT id, jmbg, first_name, last_name, email, password_hash, role, address, created_at FROM users`

func (r *PostgresUserRepository) scanOne(row pgx.Row) (*model.User, error) {
	var u model.User
	err := row.Scan(&u.ID, &u.JMBG, &u.FirstName, &u.LastName, &u.Email, &u.PasswordHash, &u.Role, &u.Address, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
