package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"euprava/payment/internal/client"
	"euprava/payment/internal/model"
	"euprava/payment/internal/repository"
)

var (
	ErrFineAlreadyPaid = errors.New("kazna je već plaćena")
	ErrNotYourFine     = errors.New("kazna ne pripada ovom građaninu")
	ErrPaymentExists   = errors.New("plaćanje za ovu kaznu je već u toku")
	ErrNotYourPayment  = errors.New("plaćanje ne pripada ovom građaninu")
)

// PaymentService sadrži poslovnu logiku plaćanja kazni.
type PaymentService struct {
	payments repository.PaymentRepository
	traffic  client.TrafficClient
	notifier client.NotificationClient
}

func NewPaymentService(payments repository.PaymentRepository, traffic client.TrafficClient, notifier client.NotificationClient) *PaymentService {
	return &PaymentService{payments: payments, traffic: traffic, notifier: notifier}
}

// Create inicira plaćanje: proverava kod Traffic Police servisa da kazna
// postoji, da nije plaćena i da pripada građaninu koji plaća.
func (s *PaymentService) Create(ctx context.Context, citizenID, fineID int64) (*model.Payment, error) {
	fine, err := s.traffic.GetFine(ctx, fineID)
	if err != nil {
		return nil, err
	}
	if fine.Paid {
		return nil, ErrFineAlreadyPaid
	}
	if fine.CitizenID != citizenID {
		return nil, ErrNotYourFine
	}
	if _, err := s.payments.GetPendingByFineID(ctx, fineID); err == nil {
		return nil, ErrPaymentExists
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	p := &model.Payment{FineID: fineID, CitizenID: citizenID, Amount: fine.Amount}
	if err := s.payments.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Confirm simulira uspešno plaćanje: kompletira plaćanje, evidentira ga
// kod Traffic Police servisa i obaveštava građanina.
func (s *PaymentService) Confirm(ctx context.Context, id, citizenID int64) (*model.Payment, error) {
	p, err := s.payments.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.CitizenID != citizenID {
		return nil, ErrNotYourPayment
	}
	if p.Status != model.PaymentPending {
		return nil, ErrFineAlreadyPaid
	}
	// Prvo evidencija kod Traffic Police servisa, pa tek onda kompletiranje
	// plaćanja — da neuspeh spoljnog poziva ne ostavi nekonzistentno stanje.
	if err := s.traffic.PayFine(ctx, p.FineID); err != nil {
		return nil, fmt.Errorf("evidentiranje plaćanja kod saobraćajne policije nije uspelo: %w", err)
	}
	if err := s.payments.MarkCompleted(ctx, id); err != nil {
		return nil, err
	}
	if err := s.notifier.Notify(ctx, p.CitizenID,
		fmt.Sprintf("Uplata od %.2f RSD za kaznu br. %d je uspešno izvršena.", p.Amount, p.FineID),
		"PAYMENT"); err != nil {
		log.Printf("slanje obaveštenja nije uspelo (citizenId=%d): %v", p.CitizenID, err)
	}
	return s.payments.GetByID(ctx, id)
}

// ListByCitizen vraća istoriju plaćanja građanina.
func (s *PaymentService) ListByCitizen(ctx context.Context, citizenID int64) ([]model.Payment, error) {
	return s.payments.ListByCitizen(ctx, citizenID)
}
