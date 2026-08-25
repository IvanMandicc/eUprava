package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"euprava/vehicles/internal/model"
	"euprava/vehicles/internal/repository"
)

// ReportService generiše digitalni istorijat vozila (vlasništvo +
// registracija) sa verifikacionim kodom koji se može javno proveriti bez
// prijave — isti duh kao "Otvoreni podaci" kod Traffic Police servisa,
// ovde primenjen na proveru autentičnosti jednog konkretnog dokumenta.
type ReportService struct {
	reports   repository.ReportRepository
	vehicles  repository.VehicleRepository
	transfers repository.TransferRepository
}

func NewReportService(reports repository.ReportRepository, vehicles repository.VehicleRepository, transfers repository.TransferRepository) *ReportService {
	return &ReportService{reports: reports, vehicles: vehicles, transfers: transfers}
}

// Generate kreira novi izveštaj o vozilu — dostupno vlasniku ili službeniku.
func (s *ReportService) Generate(ctx context.Context, vehicleID, requesterCitizenID int64, requesterRole string) (*model.VehicleReportDetails, error) {
	v, err := s.vehicles.GetByID(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	if requesterRole == "citizen" && requesterCitizenID != v.OwnerCitizenID {
		return nil, ErrNotOwner
	}

	code, err := generateVerificationCode()
	if err != nil {
		return nil, err
	}
	rep := &model.VehicleReport{VehicleID: vehicleID, VerificationCode: code, GeneratedAt: time.Now()}
	if err := s.reports.Create(ctx, rep); err != nil {
		return nil, err
	}

	return s.buildDetails(ctx, *rep, *v)
}

// Verify proverava izveštaj po javnom verifikacionom kodu — bez prijave.
func (s *ReportService) Verify(ctx context.Context, code string) (*model.VehicleReportDetails, error) {
	rep, err := s.reports.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	v, err := s.vehicles.GetByID(ctx, rep.VehicleID)
	if err != nil {
		return nil, err
	}
	return s.buildDetails(ctx, *rep, *v)
}

func (s *ReportService) buildDetails(ctx context.Context, rep model.VehicleReport, v model.Vehicle) (*model.VehicleReportDetails, error) {
	transfers, err := s.transfers.ListByVehicle(ctx, v.ID)
	if err != nil {
		return nil, err
	}
	return &model.VehicleReportDetails{VehicleReport: rep, Vehicle: v, Transfers: transfers}, nil
}

// generateVerificationCode pravi kriptografski nepredvidiv kod — javno
// izložen podatak ne sme biti pogodiv (npr. redni broj).
func generateVerificationCode() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
