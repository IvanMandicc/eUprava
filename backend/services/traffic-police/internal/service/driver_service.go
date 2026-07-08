package service

import (
	"context"
	"log"

	"euprava/traffic-police/internal/client"
	"euprava/traffic-police/internal/model"
	"euprava/traffic-police/internal/repository"
)

// RegisterDriverInput su podaci za evidentiranje vozača.
type RegisterDriverInput struct {
	CitizenID     int64  `json:"citizenId" binding:"required"`
	LicenseNumber string `json:"licenseNumber" binding:"required"`
}

// DriverService upravlja evidencijom vozača.
type DriverService struct {
	drivers  repository.DriverRepository
	citizens client.CitizenClient
}

func NewDriverService(drivers repository.DriverRepository, citizens client.CitizenClient) *DriverService {
	return &DriverService{drivers: drivers, citizens: citizens}
}

// Register evidentira vozača — prvo proverava kod Citizen servisa da građanin postoji.
func (s *DriverService) Register(ctx context.Context, in RegisterDriverInput) (*model.Driver, error) {
	if _, err := s.citizens.GetCitizen(ctx, in.CitizenID); err != nil {
		return nil, err
	}
	d := &model.Driver{CitizenID: in.CitizenID, LicenseNumber: in.LicenseNumber}
	if err := s.drivers.Create(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

// Get vraća vozača obogaćenog podacima o građaninu iz Citizen servisa.
func (s *DriverService) Get(ctx context.Context, id int64) (*model.DriverDetails, error) {
	d, err := s.drivers.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	details := &model.DriverDetails{Driver: *d}
	// Ako Citizen servis trenutno nije dostupan, vraćamo vozača bez detalja
	// o građaninu umesto greške (labava povezanost servisa).
	if info, err := s.citizens.GetCitizen(ctx, d.CitizenID); err == nil {
		details.Citizen = info
	} else {
		log.Printf("citizen servis nedostupan za citizenId=%d: %v", d.CitizenID, err)
	}
	return details, nil
}

// GetByCitizen vraća vozača po ID-ju građanina (za /me rute).
func (s *DriverService) GetByCitizen(ctx context.Context, citizenID int64) (*model.Driver, error) {
	return s.drivers.GetByCitizenID(ctx, citizenID)
}

// List vraća sve vozače.
func (s *DriverService) List(ctx context.Context) ([]model.Driver, error) {
	return s.drivers.List(ctx)
}
