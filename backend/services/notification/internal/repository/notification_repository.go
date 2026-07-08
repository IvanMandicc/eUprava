package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"euprava/notification/internal/model"
)

var ErrNotFound = errors.New("zapis nije pronađen")

// NotificationRepository apstrahuje skladištenje obaveštenja.
type NotificationRepository interface {
	Create(ctx context.Context, n *model.Notification) error
	ListByCitizen(ctx context.Context, citizenID int64) ([]model.Notification, error)
	MarkRead(ctx context.Context, id, citizenID int64) error
}

const schema = `
CREATE TABLE IF NOT EXISTS notifications (
	id         BIGSERIAL PRIMARY KEY,
	citizen_id BIGINT      NOT NULL,
	message    TEXT        NOT NULL,
	type       TEXT        NOT NULL DEFAULT 'INFO',
	read       BOOLEAN     NOT NULL DEFAULT FALSE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);`

type PostgresNotificationRepository struct{ pool *pgxpool.Pool }

func NewPostgresNotificationRepository(ctx context.Context, pool *pgxpool.Pool) (*PostgresNotificationRepository, error) {
	if _, err := pool.Exec(ctx, schema); err != nil {
		return nil, err
	}
	return &PostgresNotificationRepository{pool: pool}, nil
}

func (r *PostgresNotificationRepository) Create(ctx context.Context, n *model.Notification) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO notifications (citizen_id, message, type) VALUES ($1, $2, $3)
		 RETURNING id, read, created_at`,
		n.CitizenID, n.Message, n.Type,
	).Scan(&n.ID, &n.Read, &n.CreatedAt)
}

func (r *PostgresNotificationRepository) ListByCitizen(ctx context.Context, citizenID int64) ([]model.Notification, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, citizen_id, message, type, read, created_at
		 FROM notifications WHERE citizen_id = $1 ORDER BY created_at DESC`, citizenID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	notifications := []model.Notification{}
	for rows.Next() {
		var n model.Notification
		if err := rows.Scan(&n.ID, &n.CitizenID, &n.Message, &n.Type, &n.Read, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

func (r *PostgresNotificationRepository) MarkRead(ctx context.Context, id, citizenID int64) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE notifications SET read = TRUE WHERE id = $1 AND citizen_id = $2`, id, citizenID)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}
