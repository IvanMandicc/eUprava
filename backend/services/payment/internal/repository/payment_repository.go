package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"euprava/payment/internal/model"
)

var ErrNotFound = errors.New("zapis nije pronađen")

// PaymentRepository apstrahuje skladištenje plaćanja.
type PaymentRepository interface {
	Create(ctx context.Context, p *model.Payment) error
	GetByID(ctx context.Context, id int64) (*model.Payment, error)
	GetPendingByFineID(ctx context.Context, fineID int64) (*model.Payment, error)
	ListByCitizen(ctx context.Context, citizenID int64) ([]model.Payment, error)
	MarkCompleted(ctx context.Context, id int64) error
}

const schema = `
CREATE TABLE IF NOT EXISTS payments (
	id           BIGSERIAL PRIMARY KEY,
	fine_id      BIGINT           NOT NULL,
	citizen_id   BIGINT           NOT NULL,
	amount       DOUBLE PRECISION NOT NULL,
	status       TEXT             NOT NULL DEFAULT 'pending',
	created_at   TIMESTAMPTZ      NOT NULL DEFAULT now(),
	completed_at TIMESTAMPTZ
);`

type PostgresPaymentRepository struct{ pool *pgxpool.Pool }

func NewPostgresPaymentRepository(ctx context.Context, pool *pgxpool.Pool) (*PostgresPaymentRepository, error) {
	if _, err := pool.Exec(ctx, schema); err != nil {
		return nil, err
	}
	return &PostgresPaymentRepository{pool: pool}, nil
}

func (r *PostgresPaymentRepository) Create(ctx context.Context, p *model.Payment) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO payments (fine_id, citizen_id, amount) VALUES ($1, $2, $3)
		 RETURNING id, status, created_at`,
		p.FineID, p.CitizenID, p.Amount,
	).Scan(&p.ID, &p.Status, &p.CreatedAt)
}

const selectPayment = `SELECT id, fine_id, citizen_id, amount, status, created_at, completed_at FROM payments`

func (r *PostgresPaymentRepository) GetByID(ctx context.Context, id int64) (*model.Payment, error) {
	return scanPayment(r.pool.QueryRow(ctx, selectPayment+` WHERE id = $1`, id))
}

func (r *PostgresPaymentRepository) GetPendingByFineID(ctx context.Context, fineID int64) (*model.Payment, error) {
	return scanPayment(r.pool.QueryRow(ctx, selectPayment+` WHERE fine_id = $1 AND status = 'pending'`, fineID))
}

func (r *PostgresPaymentRepository) ListByCitizen(ctx context.Context, citizenID int64) ([]model.Payment, error) {
	rows, err := r.pool.Query(ctx, selectPayment+` WHERE citizen_id = $1 ORDER BY created_at DESC`, citizenID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	payments := []model.Payment{}
	for rows.Next() {
		p, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		payments = append(payments, *p)
	}
	return payments, rows.Err()
}

func (r *PostgresPaymentRepository) MarkCompleted(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE payments SET status = 'completed', completed_at = now() WHERE id = $1 AND status = 'pending'`, id)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

func scanPayment(row pgx.Row) (*model.Payment, error) {
	var p model.Payment
	err := row.Scan(&p.ID, &p.FineID, &p.CitizenID, &p.Amount, &p.Status, &p.CreatedAt, &p.CompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}
