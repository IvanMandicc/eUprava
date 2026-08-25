package repository

import (
	"context"
	"errors"
	"time"

	"euprava/vehicles/internal/model"
)

// ErrNotFound se vraća kada traženi zapis ne postoji.
var ErrNotFound = errors.New("zapis nije pronađen")

// Mali, fokusirani interfejsi (Interface Segregation): svaki service sloj
// zavisi samo od repozitorijuma koji mu zaista trebaju.

type VehicleRepository interface {
	Create(ctx context.Context, v *model.Vehicle) error
	GetByID(ctx context.Context, id int64) (*model.Vehicle, error)
	GetByVIN(ctx context.Context, vin string) (*model.Vehicle, error)
	GetByPlate(ctx context.Context, plate string) (*model.Vehicle, error)
	PlateExists(ctx context.Context, plate string) (bool, error)
	ListByOwner(ctx context.Context, ownerCitizenID int64) ([]model.Vehicle, error)
	List(ctx context.Context) ([]model.Vehicle, error)
	Update(ctx context.Context, v *model.Vehicle) error
}

type TransferRepository interface {
	Create(ctx context.Context, t *model.OwnershipTransfer) error
	ListByVehicle(ctx context.Context, vehicleID int64) ([]model.OwnershipTransfer, error)
}

type TheftRepository interface {
	Create(ctx context.Context, r *model.TheftReport) error
	GetOpenByVehicle(ctx context.Context, vehicleID int64) (*model.TheftReport, error)
	MarkFound(ctx context.Context, id int64) error
}

type PlateReservationRepository interface {
	Create(ctx context.Context, r *model.PlateReservation) error
	GetByID(ctx context.Context, id int64) (*model.PlateReservation, error)
	List(ctx context.Context) ([]model.PlateReservation, error)
	ListByRequester(ctx context.Context, citizenID int64) ([]model.PlateReservation, error)
	PlateReservedActive(ctx context.Context, plate string) (bool, error)
	UpdateStatus(ctx context.Context, id int64, status model.PlateReservationStatus, decidedAt time.Time) error
	ExpirePending(ctx context.Context, before time.Time) error
}

type ReportRepository interface {
	Create(ctx context.Context, r *model.VehicleReport) error
	GetByCode(ctx context.Context, code string) (*model.VehicleReport, error)
}
