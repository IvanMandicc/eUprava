package service

import (
	"context"
	"fmt"
	"log"

	"euprava/traffic-police/internal/client"
	"euprava/traffic-police/internal/model"
	"euprava/traffic-police/internal/repository"
)

// FineService upravlja novčanim kaznama.
type FineService struct {
	fines      repository.FineRepository
	violations repository.ViolationRepository
	notifier   client.NotificationClient
}

func NewFineService(fines repository.FineRepository, violations repository.ViolationRepository, notifier client.NotificationClient) *FineService {
	return &FineService{fines: fines, violations: violations, notifier: notifier}
}

func (s *FineService) Get(ctx context.Context, id int64) (*model.FineDetails, error) {
	return s.fines.GetByID(ctx, id)
}

func (s *FineService) List(ctx context.Context) ([]model.FineDetails, error) {
	return s.fines.List(ctx)
}

func (s *FineService) ListByCitizen(ctx context.Context, citizenID int64) ([]model.FineDetails, error) {
	return s.fines.ListByCitizen(ctx, citizenID)
}

// Pay evidentira plaćanje kazne (poziva ga Payment servis): kazna postaje
// plaćena, prekršaj rešen, a građanin dobija potvrdu.
func (s *FineService) Pay(ctx context.Context, id int64) (*model.FineDetails, error) {
	if err := s.fines.MarkPaid(ctx, id); err != nil {
		return nil, err
	}
	fd, err := s.fines.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if v, err := s.violations.GetByID(ctx, fd.ViolationID); err == nil {
		v.Status = model.ViolationResolved
		if err := s.violations.Update(ctx, v); err != nil {
			log.Printf("ažuriranje statusa prekršaja %d nije uspelo: %v", v.ID, err)
		}
	}

	if err := s.notifier.Notify(ctx, fd.CitizenID,
		fmt.Sprintf("Kazna br. %d u iznosu od %.2f RSD je uspešno plaćena.", fd.ID, fd.Amount),
		"FINE_PAID"); err != nil {
		log.Printf("slanje obaveštenja nije uspelo (citizenId=%d): %v", fd.CitizenID, err)
	}
	return fd, nil
}
