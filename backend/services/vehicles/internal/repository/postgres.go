package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"euprava/vehicles/internal/model"
)

const schema = `
CREATE TABLE IF NOT EXISTS vehicles (
	id                          BIGSERIAL PRIMARY KEY,
	owner_citizen_id            BIGINT      NOT NULL,
	vin                         TEXT        NOT NULL UNIQUE,
	plate_number                TEXT        NOT NULL UNIQUE,
	make                        TEXT        NOT NULL,
	model                       TEXT        NOT NULL,
	year                        INT         NOT NULL,
	category                    TEXT        NOT NULL,
	color                       TEXT        NOT NULL DEFAULT '',
	engine_power_kw             INT         NOT NULL DEFAULT 0,
	fuel_type                   TEXT        NOT NULL DEFAULT '',
	first_registration_date     TIMESTAMPTZ NOT NULL DEFAULT now(),
	insurance_valid_until       TIMESTAMPTZ NOT NULL,
	tech_inspection_valid_until TIMESTAMPTZ NOT NULL,
	registration_valid_until    TIMESTAMPTZ NOT NULL,
	status                      TEXT        NOT NULL DEFAULT 'registered',
	created_at                  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS ownership_transfers (
	id              BIGSERIAL PRIMARY KEY,
	vehicle_id      BIGINT      NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
	from_citizen_id BIGINT      NOT NULL,
	to_citizen_id   BIGINT      NOT NULL,
	transfer_date   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS theft_reports (
	id                     BIGSERIAL PRIMARY KEY,
	vehicle_id             BIGINT      NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
	reported_by_citizen_id BIGINT      NOT NULL,
	status                 TEXT        NOT NULL DEFAULT 'reported',
	reported_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
	found_at               TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS plate_reservations (
	id                       BIGSERIAL PRIMARY KEY,
	requested_by_citizen_id  BIGINT           NOT NULL,
	requested_plate          TEXT             NOT NULL,
	fee_amount               DOUBLE PRECISION NOT NULL,
	status                   TEXT             NOT NULL DEFAULT 'pending',
	requested_at             TIMESTAMPTZ      NOT NULL DEFAULT now(),
	decided_at               TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS vehicle_reports (
	id                BIGSERIAL PRIMARY KEY,
	vehicle_id        BIGINT      NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
	verification_code TEXT        NOT NULL UNIQUE,
	generated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);`

// InitSchema kreira tabele ako ne postoje.
func InitSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, schema)
	return err
}

// ---------------- VehicleRepository ----------------

type PostgresVehicleRepository struct{ pool *pgxpool.Pool }

func NewPostgresVehicleRepository(pool *pgxpool.Pool) *PostgresVehicleRepository {
	return &PostgresVehicleRepository{pool: pool}
}

const selectVehicle = `
SELECT id, owner_citizen_id, vin, plate_number, make, model, year, category, color,
       engine_power_kw, fuel_type, first_registration_date, insurance_valid_until,
       tech_inspection_valid_until, registration_valid_until, status, created_at
FROM vehicles`

func (r *PostgresVehicleRepository) Create(ctx context.Context, v *model.Vehicle) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO vehicles (owner_citizen_id, vin, plate_number, make, model, year, category, color,
		                       engine_power_kw, fuel_type, first_registration_date, insurance_valid_until,
		                       tech_inspection_valid_until, registration_valid_until, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		 RETURNING id, created_at`,
		v.OwnerCitizenID, v.VIN, v.PlateNumber, v.Make, v.Model, v.Year, v.Category, v.Color,
		v.EnginePowerKw, v.FuelType, v.FirstRegistrationDate, v.InsuranceValidUntil,
		v.TechInspectionValidUntil, v.RegistrationValidUntil, v.Status,
	).Scan(&v.ID, &v.CreatedAt)
}

func (r *PostgresVehicleRepository) GetByID(ctx context.Context, id int64) (*model.Vehicle, error) {
	return scanVehicle(r.pool.QueryRow(ctx, selectVehicle+` WHERE id = $1`, id))
}

func (r *PostgresVehicleRepository) GetByVIN(ctx context.Context, vin string) (*model.Vehicle, error) {
	return scanVehicle(r.pool.QueryRow(ctx, selectVehicle+` WHERE vin = $1`, vin))
}

func (r *PostgresVehicleRepository) GetByPlate(ctx context.Context, plate string) (*model.Vehicle, error) {
	return scanVehicle(r.pool.QueryRow(ctx, selectVehicle+` WHERE plate_number = $1`, plate))
}

func (r *PostgresVehicleRepository) PlateExists(ctx context.Context, plate string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM vehicles WHERE plate_number = $1)`, plate).Scan(&exists)
	return exists, err
}

func (r *PostgresVehicleRepository) ListByOwner(ctx context.Context, ownerCitizenID int64) ([]model.Vehicle, error) {
	return r.list(ctx, selectVehicle+` WHERE owner_citizen_id = $1 ORDER BY id`, ownerCitizenID)
}

func (r *PostgresVehicleRepository) List(ctx context.Context) ([]model.Vehicle, error) {
	return r.list(ctx, selectVehicle+` ORDER BY id`)
}

func (r *PostgresVehicleRepository) list(ctx context.Context, query string, args ...any) ([]model.Vehicle, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	vehicles := []model.Vehicle{}
	for rows.Next() {
		v, err := scanVehicleRow(rows)
		if err != nil {
			return nil, err
		}
		vehicles = append(vehicles, *v)
	}
	return vehicles, rows.Err()
}

func (r *PostgresVehicleRepository) Update(ctx context.Context, v *model.Vehicle) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE vehicles SET owner_citizen_id = $2, insurance_valid_until = $3,
		                     tech_inspection_valid_until = $4, registration_valid_until = $5, status = $6
		 WHERE id = $1`,
		v.ID, v.OwnerCitizenID, v.InsuranceValidUntil, v.TechInspectionValidUntil, v.RegistrationValidUntil, v.Status)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

// row je zajednički interfejs za pgx.Row i pgx.Rows (obe imaju Scan).
type row interface {
	Scan(dest ...any) error
}

func scanVehicle(r row) (*model.Vehicle, error) {
	v, err := scanVehicleRow(r)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return v, err
}

func scanVehicleRow(r row) (*model.Vehicle, error) {
	var v model.Vehicle
	err := r.Scan(&v.ID, &v.OwnerCitizenID, &v.VIN, &v.PlateNumber, &v.Make, &v.Model, &v.Year, &v.Category, &v.Color,
		&v.EnginePowerKw, &v.FuelType, &v.FirstRegistrationDate, &v.InsuranceValidUntil,
		&v.TechInspectionValidUntil, &v.RegistrationValidUntil, &v.Status, &v.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// ---------------- TransferRepository ----------------

type PostgresTransferRepository struct{ pool *pgxpool.Pool }

func NewPostgresTransferRepository(pool *pgxpool.Pool) *PostgresTransferRepository {
	return &PostgresTransferRepository{pool: pool}
}

func (r *PostgresTransferRepository) Create(ctx context.Context, t *model.OwnershipTransfer) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO ownership_transfers (vehicle_id, from_citizen_id, to_citizen_id, transfer_date)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		t.VehicleID, t.FromCitizenID, t.ToCitizenID, t.TransferDate,
	).Scan(&t.ID)
}

func (r *PostgresTransferRepository) ListByVehicle(ctx context.Context, vehicleID int64) ([]model.OwnershipTransfer, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, vehicle_id, from_citizen_id, to_citizen_id, transfer_date
		 FROM ownership_transfers WHERE vehicle_id = $1 ORDER BY transfer_date`, vehicleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	transfers := []model.OwnershipTransfer{}
	for rows.Next() {
		var t model.OwnershipTransfer
		if err := rows.Scan(&t.ID, &t.VehicleID, &t.FromCitizenID, &t.ToCitizenID, &t.TransferDate); err != nil {
			return nil, err
		}
		transfers = append(transfers, t)
	}
	return transfers, rows.Err()
}

// ---------------- TheftRepository ----------------

type PostgresTheftRepository struct{ pool *pgxpool.Pool }

func NewPostgresTheftRepository(pool *pgxpool.Pool) *PostgresTheftRepository {
	return &PostgresTheftRepository{pool: pool}
}

func (r *PostgresTheftRepository) Create(ctx context.Context, t *model.TheftReport) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO theft_reports (vehicle_id, reported_by_citizen_id, status, reported_at)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		t.VehicleID, t.ReportedByID, t.Status, t.ReportedAt,
	).Scan(&t.ID)
}

func (r *PostgresTheftRepository) GetOpenByVehicle(ctx context.Context, vehicleID int64) (*model.TheftReport, error) {
	var t model.TheftReport
	err := r.pool.QueryRow(ctx,
		`SELECT id, vehicle_id, reported_by_citizen_id, status, reported_at, found_at
		 FROM theft_reports WHERE vehicle_id = $1 AND status = 'reported'
		 ORDER BY reported_at DESC LIMIT 1`, vehicleID,
	).Scan(&t.ID, &t.VehicleID, &t.ReportedByID, &t.Status, &t.ReportedAt, &t.FoundAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *PostgresTheftRepository) MarkFound(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE theft_reports SET status = 'found', found_at = now() WHERE id = $1 AND status = 'reported'`, id)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

// ---------------- PlateReservationRepository ----------------

type PostgresPlateReservationRepository struct{ pool *pgxpool.Pool }

func NewPostgresPlateReservationRepository(pool *pgxpool.Pool) *PostgresPlateReservationRepository {
	return &PostgresPlateReservationRepository{pool: pool}
}

const selectReservation = `
SELECT id, requested_by_citizen_id, requested_plate, fee_amount, status, requested_at, decided_at
FROM plate_reservations`

func (r *PostgresPlateReservationRepository) Create(ctx context.Context, res *model.PlateReservation) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO plate_reservations (requested_by_citizen_id, requested_plate, fee_amount, status, requested_at)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		res.RequestedByID, res.RequestedPlate, res.FeeAmount, res.Status, res.RequestedAt,
	).Scan(&res.ID)
}

func (r *PostgresPlateReservationRepository) GetByID(ctx context.Context, id int64) (*model.PlateReservation, error) {
	res, err := scanReservation(r.pool.QueryRow(ctx, selectReservation+` WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return res, err
}

func (r *PostgresPlateReservationRepository) List(ctx context.Context) ([]model.PlateReservation, error) {
	return r.list(ctx, selectReservation+` ORDER BY requested_at DESC`)
}

func (r *PostgresPlateReservationRepository) ListByRequester(ctx context.Context, citizenID int64) ([]model.PlateReservation, error) {
	return r.list(ctx, selectReservation+` WHERE requested_by_citizen_id = $1 ORDER BY requested_at DESC`, citizenID)
}

func (r *PostgresPlateReservationRepository) list(ctx context.Context, query string, args ...any) ([]model.PlateReservation, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	reservations := []model.PlateReservation{}
	for rows.Next() {
		res, err := scanReservation(rows)
		if err != nil {
			return nil, err
		}
		reservations = append(reservations, *res)
	}
	return reservations, rows.Err()
}

func (r *PostgresPlateReservationRepository) PlateReservedActive(ctx context.Context, plate string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM plate_reservations
		               WHERE requested_plate = $1 AND status IN ('pending', 'approved'))`, plate,
	).Scan(&exists)
	return exists, err
}

func (r *PostgresPlateReservationRepository) UpdateStatus(ctx context.Context, id int64, status model.PlateReservationStatus, decidedAt time.Time) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE plate_reservations SET status = $2, decided_at = $3 WHERE id = $1`, id, status, decidedAt)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

func (r *PostgresPlateReservationRepository) ExpirePending(ctx context.Context, before time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE plate_reservations SET status = 'expired', decided_at = now()
		 WHERE status = 'pending' AND requested_at < $1`, before)
	return err
}

func scanReservation(r row) (*model.PlateReservation, error) {
	var res model.PlateReservation
	err := r.Scan(&res.ID, &res.RequestedByID, &res.RequestedPlate, &res.FeeAmount, &res.Status, &res.RequestedAt, &res.DecidedAt)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// ---------------- ReportRepository ----------------

type PostgresReportRepository struct{ pool *pgxpool.Pool }

func NewPostgresReportRepository(pool *pgxpool.Pool) *PostgresReportRepository {
	return &PostgresReportRepository{pool: pool}
}

func (r *PostgresReportRepository) Create(ctx context.Context, rep *model.VehicleReport) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO vehicle_reports (vehicle_id, verification_code, generated_at)
		 VALUES ($1, $2, $3) RETURNING id`,
		rep.VehicleID, rep.VerificationCode, rep.GeneratedAt,
	).Scan(&rep.ID)
}

func (r *PostgresReportRepository) GetByCode(ctx context.Context, code string) (*model.VehicleReport, error) {
	var rep model.VehicleReport
	err := r.pool.QueryRow(ctx,
		`SELECT id, vehicle_id, verification_code, generated_at FROM vehicle_reports WHERE verification_code = $1`, code,
	).Scan(&rep.ID, &rep.VehicleID, &rep.VerificationCode, &rep.GeneratedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rep, nil
}
