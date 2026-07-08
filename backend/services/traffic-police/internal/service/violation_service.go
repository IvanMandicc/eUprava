package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"euprava/traffic-police/internal/client"
	"euprava/traffic-police/internal/model"
	"euprava/traffic-police/internal/repository"
)

var (
	ErrUnknownViolationType = errors.New("nepoznat tip prekršaja")
	ErrFinePaid             = errors.New("prekršaj sa plaćenom kaznom ne može da se obriše")
)

// finePaymentDeadline je rok za plaćanje kazne od dana prekršaja.
const finePaymentDeadline = 15 * 24 * time.Hour

// CreateViolationInput su podaci za unos novog prekršaja.
type CreateViolationInput struct {
	DriverID    int64     `json:"driverId" binding:"required"`
	Type        string    `json:"type" binding:"required"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	Date        time.Time `json:"date"`
}

// UpdateViolationInput su podaci za izmenu prekršaja.
type UpdateViolationInput struct {
	Description string                `json:"description"`
	Location    string                `json:"location"`
	Status      model.ViolationStatus `json:"status" binding:"omitempty,oneof=active resolved"`
}

// ViolationService sadrži centralnu poslovnu logiku servisa:
// unos prekršaja automatski kreira kaznu, dodaje kaznene poene
// i obaveštava građanina; na 18+ poena suspenduje dozvolu.
type ViolationService struct {
	drivers    repository.DriverRepository
	violations repository.ViolationRepository
	fines      repository.FineRepository
	notifier   client.NotificationClient
}

func NewViolationService(
	drivers repository.DriverRepository,
	violations repository.ViolationRepository,
	fines repository.FineRepository,
	notifier client.NotificationClient,
) *ViolationService {
	return &ViolationService{drivers: drivers, violations: violations, fines: fines, notifier: notifier}
}

// Create evidentira prekršaj i sprovodi sva poslovna pravila.
func (s *ViolationService) Create(ctx context.Context, in CreateViolationInput) (*model.Violation, error) {
	entry, ok := model.CatalogEntry(in.Type)
	if !ok {
		return nil, ErrUnknownViolationType
	}
	driver, err := s.drivers.GetByID(ctx, in.DriverID)
	if err != nil {
		return nil, err
	}

	date := in.Date
	if date.IsZero() {
		date = time.Now()
	}
	v := &model.Violation{
		DriverID:    driver.ID,
		Type:        entry.Code,
		Description: in.Description,
		Date:        date,
		Location:    in.Location,
		Points:      entry.Points,
		FineAmount:  entry.Fine,
		Status:      model.ViolationActive,
	}
	if err := s.violations.Create(ctx, v); err != nil {
		return nil, err
	}

	// 1) Automatski se kreira novčana kazna sa rokom plaćanja.
	fine := &model.Fine{
		ViolationID:     v.ID,
		Amount:          entry.Fine,
		PaymentDeadline: date.Add(finePaymentDeadline),
	}
	if err := s.fines.Create(ctx, fine); err != nil {
		return nil, err
	}

	// 2) Dodaju se kazneni poeni; na dostignut limit dozvola se suspenduje.
	newPoints := driver.PenaltyPoints + entry.Points
	status := driver.LicenseStatus
	if newPoints >= model.PenaltyPointsLimit {
		status = model.LicenseSuspended
	}
	if err := s.drivers.UpdatePoints(ctx, driver.ID, newPoints, status); err != nil {
		return nil, err
	}

	// 3) Građanin dobija obaveštenja (neuspeh slanja ne obara unos prekršaja).
	s.notify(ctx, driver.CitizenID,
		fmt.Sprintf("Izrečena vam je nova kazna: %s (%s). Iznos: %.2f RSD, rok za plaćanje: %s.",
			entry.Label, in.Location, entry.Fine, fine.PaymentDeadline.Format("02.01.2006.")),
		"FINE")
	if status == model.LicenseSuspended && driver.LicenseStatus != model.LicenseSuspended {
		s.notify(ctx, driver.CitizenID,
			fmt.Sprintf("Vozačka dozvola vam je suspendovana — dostigli ste %d kaznenih poena (limit je %d).",
				newPoints, model.PenaltyPointsLimit),
			"LICENSE_SUSPENDED")
	}
	return v, nil
}

// Update menja opis, lokaciju ili status prekršaja.
func (s *ViolationService) Update(ctx context.Context, id int64, in UpdateViolationInput) (*model.Violation, error) {
	v, err := s.violations.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.Description != "" {
		v.Description = in.Description
	}
	if in.Location != "" {
		v.Location = in.Location
	}
	if in.Status != "" {
		v.Status = in.Status
	}
	if err := s.violations.Update(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

// Delete briše prekršaj: oduzima njegove poene vozaču (uz ponovnu proveru
// statusa dozvole) i ukida nevezanu kaznu. Plaćena kazna blokira brisanje.
func (s *ViolationService) Delete(ctx context.Context, id int64) error {
	v, err := s.violations.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if fine, err := s.fines.GetByViolationID(ctx, v.ID); err == nil && fine.Paid {
		return ErrFinePaid
	} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return err
	}

	driver, err := s.drivers.GetByID(ctx, v.DriverID)
	if err != nil {
		return err
	}
	if err := s.fines.DeleteByViolationID(ctx, v.ID); err != nil {
		return err
	}
	if err := s.violations.Delete(ctx, v.ID); err != nil {
		return err
	}

	newPoints := driver.PenaltyPoints - v.Points
	if newPoints < 0 {
		newPoints = 0
	}
	status := driver.LicenseStatus
	if newPoints < model.PenaltyPointsLimit {
		status = model.LicenseValid
	}
	return s.drivers.UpdatePoints(ctx, driver.ID, newPoints, status)
}

// ListByDriver vraća istoriju prekršaja vozača.
func (s *ViolationService) ListByDriver(ctx context.Context, driverID int64) ([]model.Violation, error) {
	return s.violations.ListByDriver(ctx, driverID)
}

// List vraća sve prekršaje.
func (s *ViolationService) List(ctx context.Context) ([]model.Violation, error) {
	return s.violations.List(ctx)
}

// PenaltyPoints vraća ukupne kaznene poene i status dozvole vozača.
func (s *ViolationService) PenaltyPoints(ctx context.Context, driverID int64) (*model.Driver, error) {
	return s.drivers.GetByID(ctx, driverID)
}

func (s *ViolationService) notify(ctx context.Context, citizenID int64, message, notifType string) {
	if err := s.notifier.Notify(ctx, citizenID, message, notifType); err != nil {
		log.Printf("slanje obaveštenja nije uspelo (citizenId=%d): %v", citizenID, err)
	}
}
