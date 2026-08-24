package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"euprava/traffic-police/internal/model"
)

const schema = `
CREATE TABLE IF NOT EXISTS drivers (
	id             BIGSERIAL PRIMARY KEY,
	citizen_id     BIGINT      NOT NULL UNIQUE,
	license_number TEXT        NOT NULL UNIQUE,
	penalty_points INT         NOT NULL DEFAULT 0,
	license_status TEXT        NOT NULL DEFAULT 'valid'
);

CREATE TABLE IF NOT EXISTS violations (
	id          BIGSERIAL PRIMARY KEY,
	driver_id   BIGINT      NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
	type        TEXT        NOT NULL,
	description TEXT        NOT NULL DEFAULT '',
	date        TIMESTAMPTZ NOT NULL DEFAULT now(),
	location    TEXT        NOT NULL DEFAULT '',
	points      INT         NOT NULL,
	fine_amount DOUBLE PRECISION NOT NULL,
	status      TEXT        NOT NULL DEFAULT 'active'
);

CREATE TABLE IF NOT EXISTS fines (
	id               BIGSERIAL PRIMARY KEY,
	violation_id     BIGINT      NOT NULL UNIQUE REFERENCES violations(id) ON DELETE CASCADE,
	amount           DOUBLE PRECISION NOT NULL,
	payment_deadline TIMESTAMPTZ NOT NULL,
	paid             BOOLEAN     NOT NULL DEFAULT FALSE,
	payment_date     TIMESTAMPTZ
);`

// InitSchema kreira tabele ako ne postoje.
func InitSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, schema)
	return err
}

// ---------------- DriverRepository ----------------

type PostgresDriverRepository struct{ pool *pgxpool.Pool }

func NewPostgresDriverRepository(pool *pgxpool.Pool) *PostgresDriverRepository {
	return &PostgresDriverRepository{pool: pool}
}

func (r *PostgresDriverRepository) Create(ctx context.Context, d *model.Driver) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO drivers (citizen_id, license_number) VALUES ($1, $2)
		 RETURNING id, penalty_points, license_status`,
		d.CitizenID, d.LicenseNumber,
	).Scan(&d.ID, &d.PenaltyPoints, &d.LicenseStatus)
}

const selectDriver = `SELECT id, citizen_id, license_number, penalty_points, license_status FROM drivers`

func (r *PostgresDriverRepository) GetByID(ctx context.Context, id int64) (*model.Driver, error) {
	return scanDriver(r.pool.QueryRow(ctx, selectDriver+` WHERE id = $1`, id))
}

func (r *PostgresDriverRepository) GetByCitizenID(ctx context.Context, citizenID int64) (*model.Driver, error) {
	return scanDriver(r.pool.QueryRow(ctx, selectDriver+` WHERE citizen_id = $1`, citizenID))
}

func (r *PostgresDriverRepository) List(ctx context.Context) ([]model.Driver, error) {
	rows, err := r.pool.Query(ctx, selectDriver+` ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	drivers := []model.Driver{}
	for rows.Next() {
		var d model.Driver
		if err := rows.Scan(&d.ID, &d.CitizenID, &d.LicenseNumber, &d.PenaltyPoints, &d.LicenseStatus); err != nil {
			return nil, err
		}
		drivers = append(drivers, d)
	}
	return drivers, rows.Err()
}

func (r *PostgresDriverRepository) UpdatePoints(ctx context.Context, id int64, points int, status model.LicenseStatus) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE drivers SET penalty_points = $2, license_status = $3 WHERE id = $1`, id, points, status)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

func scanDriver(row pgx.Row) (*model.Driver, error) {
	var d model.Driver
	err := row.Scan(&d.ID, &d.CitizenID, &d.LicenseNumber, &d.PenaltyPoints, &d.LicenseStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// ---------------- ViolationRepository ----------------

type PostgresViolationRepository struct{ pool *pgxpool.Pool }

func NewPostgresViolationRepository(pool *pgxpool.Pool) *PostgresViolationRepository {
	return &PostgresViolationRepository{pool: pool}
}

func (r *PostgresViolationRepository) Create(ctx context.Context, v *model.Violation) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO violations (driver_id, type, description, date, location, points, fine_amount, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		v.DriverID, v.Type, v.Description, v.Date, v.Location, v.Points, v.FineAmount, v.Status,
	).Scan(&v.ID)
}

const selectViolation = `SELECT id, driver_id, type, description, date, location, points, fine_amount, status FROM violations`

func (r *PostgresViolationRepository) GetByID(ctx context.Context, id int64) (*model.Violation, error) {
	var v model.Violation
	err := r.pool.QueryRow(ctx, selectViolation+` WHERE id = $1`, id).
		Scan(&v.ID, &v.DriverID, &v.Type, &v.Description, &v.Date, &v.Location, &v.Points, &v.FineAmount, &v.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *PostgresViolationRepository) ListByDriver(ctx context.Context, driverID int64) ([]model.Violation, error) {
	return r.list(ctx, selectViolation+` WHERE driver_id = $1 ORDER BY date DESC`, driverID)
}

func (r *PostgresViolationRepository) List(ctx context.Context) ([]model.Violation, error) {
	return r.list(ctx, selectViolation+` ORDER BY date DESC`)
}

func (r *PostgresViolationRepository) list(ctx context.Context, query string, args ...any) ([]model.Violation, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	violations := []model.Violation{}
	for rows.Next() {
		var v model.Violation
		if err := rows.Scan(&v.ID, &v.DriverID, &v.Type, &v.Description, &v.Date, &v.Location, &v.Points, &v.FineAmount, &v.Status); err != nil {
			return nil, err
		}
		violations = append(violations, v)
	}
	return violations, rows.Err()
}

// Stats vraća anonimnu agregiranu statistiku prekršaja po tipu (open data),
// opciono filtriranu po datumu prekršaja (from/to su inkluzivno/ekskluzivno).
func (r *PostgresViolationRepository) Stats(ctx context.Context, from, to *time.Time) ([]model.ViolationStat, error) {
	query := `SELECT type, COUNT(*), COALESCE(SUM(points), 0),
	        COALESCE(AVG(fine_amount), 0), COALESCE(SUM(fine_amount), 0)
	 FROM violations`
	var args []any
	var conditions []string
	if from != nil {
		args = append(args, *from)
		conditions = append(conditions, fmt.Sprintf("date >= $%d", len(args)))
	}
	if to != nil {
		args = append(args, *to)
		conditions = append(conditions, fmt.Sprintf("date < $%d", len(args)))
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " GROUP BY type ORDER BY COUNT(*) DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	stats := []model.ViolationStat{}
	for rows.Next() {
		var s model.ViolationStat
		if err := rows.Scan(&s.Type, &s.Count, &s.TotalPoints, &s.AvgFine, &s.TotalFines); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

func (r *PostgresViolationRepository) Update(ctx context.Context, v *model.Violation) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE violations SET description = $2, location = $3, status = $4 WHERE id = $1`,
		v.ID, v.Description, v.Location, v.Status)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

func (r *PostgresViolationRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM violations WHERE id = $1`, id)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

// ---------------- FineRepository ----------------

type PostgresFineRepository struct{ pool *pgxpool.Pool }

func NewPostgresFineRepository(pool *pgxpool.Pool) *PostgresFineRepository {
	return &PostgresFineRepository{pool: pool}
}

func (r *PostgresFineRepository) Create(ctx context.Context, f *model.Fine) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO fines (violation_id, amount, payment_deadline) VALUES ($1, $2, $3) RETURNING id`,
		f.ViolationID, f.Amount, f.PaymentDeadline,
	).Scan(&f.ID)
}

const selectFineDetails = `
SELECT f.id, f.violation_id, f.amount, f.payment_deadline, f.paid, f.payment_date,
       v.type, v.date, v.location, v.driver_id, d.citizen_id
FROM fines f
JOIN violations v ON v.id = f.violation_id
JOIN drivers d ON d.id = v.driver_id`

func (r *PostgresFineRepository) GetByID(ctx context.Context, id int64) (*model.FineDetails, error) {
	fd, err := scanFineDetails(r.pool.QueryRow(ctx, selectFineDetails+` WHERE f.id = $1`, id))
	if err != nil {
		return nil, err
	}
	return fd, nil
}

func (r *PostgresFineRepository) GetByViolationID(ctx context.Context, violationID int64) (*model.Fine, error) {
	var f model.Fine
	err := r.pool.QueryRow(ctx,
		`SELECT id, violation_id, amount, payment_deadline, paid, payment_date FROM fines WHERE violation_id = $1`,
		violationID,
	).Scan(&f.ID, &f.ViolationID, &f.Amount, &f.PaymentDeadline, &f.Paid, &f.PaymentDate)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *PostgresFineRepository) List(ctx context.Context) ([]model.FineDetails, error) {
	return r.listDetails(ctx, selectFineDetails+` ORDER BY f.id DESC`)
}

func (r *PostgresFineRepository) ListByCitizen(ctx context.Context, citizenID int64) ([]model.FineDetails, error) {
	return r.listDetails(ctx, selectFineDetails+` WHERE d.citizen_id = $1 ORDER BY f.id DESC`, citizenID)
}

func (r *PostgresFineRepository) listDetails(ctx context.Context, query string, args ...any) ([]model.FineDetails, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	fines := []model.FineDetails{}
	for rows.Next() {
		fd, err := scanFineDetails(rows)
		if err != nil {
			return nil, err
		}
		fines = append(fines, *fd)
	}
	return fines, rows.Err()
}

func (r *PostgresFineRepository) MarkPaid(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE fines SET paid = TRUE, payment_date = now() WHERE id = $1 AND paid = FALSE`, id)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

func (r *PostgresFineRepository) DeleteByViolationID(ctx context.Context, violationID int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM fines WHERE violation_id = $1`, violationID)
	return err
}

func scanFineDetails(row pgx.Row) (*model.FineDetails, error) {
	var fd model.FineDetails
	err := row.Scan(&fd.ID, &fd.ViolationID, &fd.Amount, &fd.PaymentDeadline, &fd.Paid, &fd.PaymentDate,
		&fd.ViolationType, &fd.ViolationDate, &fd.Location, &fd.DriverID, &fd.CitizenID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &fd, nil
}
