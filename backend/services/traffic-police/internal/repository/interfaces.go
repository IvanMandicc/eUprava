package repository

import (
	"context"
	"errors"
	"time"

	"euprava/traffic-police/internal/model"
)

// ErrNotFound se vraća kada traženi zapis ne postoji.
var ErrNotFound = errors.New("zapis nije pronađen")

// Mali, fokusirani interfejsi (Interface Segregation): svaki service sloj
// zavisi samo od repozitorijuma koji mu zaista trebaju.

type DriverRepository interface {
	Create(ctx context.Context, d *model.Driver) error
	GetByID(ctx context.Context, id int64) (*model.Driver, error)
	GetByCitizenID(ctx context.Context, citizenID int64) (*model.Driver, error)
	List(ctx context.Context) ([]model.Driver, error)
	UpdatePoints(ctx context.Context, id int64, points int, status model.LicenseStatus) error
}

type ViolationRepository interface {
	Create(ctx context.Context, v *model.Violation) error
	GetByID(ctx context.Context, id int64) (*model.Violation, error)
	ListByDriver(ctx context.Context, driverID int64) ([]model.Violation, error)
	List(ctx context.Context) ([]model.Violation, error)
	Update(ctx context.Context, v *model.Violation) error
	Delete(ctx context.Context, id int64) error
	// Stats vraća statistiku po tipu prekršaja; from/to (opciono) filtriraju po datumu prekršaja.
	Stats(ctx context.Context, from, to *time.Time) ([]model.ViolationStat, error)
}

type FineRepository interface {
	Create(ctx context.Context, f *model.Fine) error
	GetByID(ctx context.Context, id int64) (*model.FineDetails, error)
	GetByViolationID(ctx context.Context, violationID int64) (*model.Fine, error)
	List(ctx context.Context) ([]model.FineDetails, error)
	ListByCitizen(ctx context.Context, citizenID int64) ([]model.FineDetails, error)
	MarkPaid(ctx context.Context, id int64) error
	DeleteByViolationID(ctx context.Context, violationID int64) error
	UnpaidSummaryByCitizen(ctx context.Context, citizenID int64) (*model.UnpaidFinesSummary, error)
}
