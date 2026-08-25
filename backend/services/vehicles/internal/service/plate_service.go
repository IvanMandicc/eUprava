package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"euprava/vehicles/internal/client"
	"euprava/vehicles/internal/model"
	"euprava/vehicles/internal/repository"
)

var (
	ErrInvalidPlateFormat    = errors.New("tablica mora imati 4-7 znakova (slova A-Z bez Q,W,X,Y i cifre)")
	ErrPlateForbidden        = errors.New("tražena kombinacija nije dozvoljena")
	ErrPlateTaken            = errors.New("tablica je već zauzeta ili rezervisana")
	ErrReservationNotPending = errors.New("zahtev je već obrađen")
)

var plateFormatPattern = regexp.MustCompile(`^[A-Z0-9]{4,7}$`)

// forbiddenPlateWords je kratka lista zabranjenih reči/kombinacija —
// personalizovana tablica ne sme sadržati uvredljiv sadržaj.
var forbiddenPlateWords = []string{"KURAC", "PICKA", "JEBEM", "SRANJE", "NAZI"}

// RequestPlateInput su podaci za zahtev za personalizovanu tablicu.
type RequestPlateInput struct {
	RequestedPlate string `json:"requestedPlate" binding:"required"`
}

// PlateService upravlja rezervacijama personalizovanih registarskih tablica:
// validacija formata, provera kolizije, obračun takse, odobravanje/odbijanje.
type PlateService struct {
	reservations repository.PlateReservationRepository
	vehicles     repository.VehicleRepository
	notifier     client.NotificationClient
}

func NewPlateService(reservations repository.PlateReservationRepository, vehicles repository.VehicleRepository, notifier client.NotificationClient) *PlateService {
	return &PlateService{reservations: reservations, vehicles: vehicles, notifier: notifier}
}

// Request podnosi zahtev za personalizovanu tablicu građanina.
func (s *PlateService) Request(ctx context.Context, citizenID int64, in RequestPlateInput) (*model.PlateReservation, error) {
	s.expirePending(ctx)

	plate := strings.ToUpper(strings.TrimSpace(in.RequestedPlate))
	if !plateFormatPattern.MatchString(plate) {
		return nil, ErrInvalidPlateFormat
	}
	for _, word := range forbiddenPlateWords {
		if strings.Contains(plate, word) {
			return nil, ErrPlateForbidden
		}
	}

	taken, err := s.vehicles.PlateExists(ctx, plate)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, ErrPlateTaken
	}
	reserved, err := s.reservations.PlateReservedActive(ctx, plate)
	if err != nil {
		return nil, err
	}
	if reserved {
		return nil, ErrPlateTaken
	}

	fee := model.VanityPlateBaseFee
	if len(plate) <= model.VanityPlateShortMaxLen {
		fee = model.VanityPlateShortFee
	}

	r := &model.PlateReservation{
		RequestedByID: citizenID, RequestedPlate: plate, FeeAmount: fee,
		Status: model.PlateReservationPending, RequestedAt: time.Now(),
	}
	if err := s.reservations.Create(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

// Decide odobrava ili odbija zahtev za personalizovanu tablicu (službenik).
func (s *PlateService) Decide(ctx context.Context, id int64, approve bool) (*model.PlateReservation, error) {
	s.expirePending(ctx)

	r, err := s.reservations.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if r.Status != model.PlateReservationPending {
		return nil, ErrReservationNotPending
	}

	status := model.PlateReservationRejected
	if approve {
		status = model.PlateReservationApproved
	}
	now := time.Now()
	if err := s.reservations.UpdateStatus(ctx, id, status, now); err != nil {
		return nil, err
	}
	r.Status = status
	r.DecidedAt = &now

	msg := fmt.Sprintf("Zahtev za personalizovanu tablicu %s je odbijen.", r.RequestedPlate)
	if approve {
		msg = fmt.Sprintf("Zahtev za personalizovanu tablicu %s je odobren. Taksa za uplatu: %.2f RSD.", r.RequestedPlate, r.FeeAmount)
	}
	if err := s.notifier.Notify(ctx, r.RequestedByID, msg, "PLATE_RESERVATION_DECIDED"); err != nil {
		log.Printf("slanje obaveštenja nije uspelo (citizenId=%d): %v", r.RequestedByID, err)
	}
	return r, nil
}

// List vraća sve zahteve (za službenika).
func (s *PlateService) List(ctx context.Context) ([]model.PlateReservation, error) {
	s.expirePending(ctx)
	return s.reservations.List(ctx)
}

// ListByRequester vraća sopstvene zahteve građanina (za /me rute).
func (s *PlateService) ListByRequester(ctx context.Context, citizenID int64) ([]model.PlateReservation, error) {
	s.expirePending(ctx)
	return s.reservations.ListByRequester(ctx, citizenID)
}

// expirePending ističe nepotvrđene zahteve starije od VanityPlateReservationTTL
// (best-effort, poziva se pri svakom čitanju/odlučivanju).
func (s *PlateService) expirePending(ctx context.Context) {
	if err := s.reservations.ExpirePending(ctx, time.Now().Add(-model.VanityPlateReservationTTL)); err != nil {
		log.Printf("isticanje nepotvrđenih rezervacija nije uspelo: %v", err)
	}
}
