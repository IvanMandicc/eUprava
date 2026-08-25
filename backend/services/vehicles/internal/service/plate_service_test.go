package service

import (
	"context"
	"testing"
	"time"

	"euprava/vehicles/internal/model"
	"euprava/vehicles/internal/repository"
)

type mockReservationRepo struct {
	reservations map[int64]*model.PlateReservation
	nextID       int64
}

func (m *mockReservationRepo) Create(_ context.Context, r *model.PlateReservation) error {
	m.nextID++
	r.ID = m.nextID
	cp := *r
	m.reservations[r.ID] = &cp
	return nil
}

func (m *mockReservationRepo) GetByID(_ context.Context, id int64) (*model.PlateReservation, error) {
	r, ok := m.reservations[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	cp := *r
	return &cp, nil
}

func (m *mockReservationRepo) List(_ context.Context) ([]model.PlateReservation, error) {
	out := []model.PlateReservation{}
	for _, r := range m.reservations {
		out = append(out, *r)
	}
	return out, nil
}

func (m *mockReservationRepo) ListByRequester(_ context.Context, citizenID int64) ([]model.PlateReservation, error) {
	out := []model.PlateReservation{}
	for _, r := range m.reservations {
		if r.RequestedByID == citizenID {
			out = append(out, *r)
		}
	}
	return out, nil
}

func (m *mockReservationRepo) PlateReservedActive(_ context.Context, plate string) (bool, error) {
	for _, r := range m.reservations {
		if r.RequestedPlate == plate && (r.Status == model.PlateReservationPending || r.Status == model.PlateReservationApproved) {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockReservationRepo) UpdateStatus(_ context.Context, id int64, status model.PlateReservationStatus, decidedAt time.Time) error {
	r, ok := m.reservations[id]
	if !ok {
		return repository.ErrNotFound
	}
	r.Status = status
	r.DecidedAt = &decidedAt
	return nil
}

func (m *mockReservationRepo) ExpirePending(_ context.Context, _ time.Time) error { return nil }

func newTestPlateService() (*PlateService, *mockVehicleRepo, *mockNotifier) {
	reservations := &mockReservationRepo{reservations: map[int64]*model.PlateReservation{}}
	vehicles := &mockVehicleRepo{vehicles: map[int64]*model.Vehicle{}}
	notifier := &mockNotifier{}
	return NewPlateService(reservations, vehicles, notifier), vehicles, notifier
}

func TestRequestPlateRejectsInvalidFormat(t *testing.T) {
	svc, _, _ := newTestPlateService()
	if _, err := svc.Request(context.Background(), 10, RequestPlateInput{RequestedPlate: "ab"}); err != ErrInvalidPlateFormat {
		t.Errorf("očekivana greška ErrInvalidPlateFormat, dobijeno: %v", err)
	}
}

func TestRequestPlateRejectsForbiddenWord(t *testing.T) {
	svc, _, _ := newTestPlateService()
	if _, err := svc.Request(context.Background(), 10, RequestPlateInput{RequestedPlate: "KURAC1"}); err != ErrPlateForbidden {
		t.Errorf("očekivana greška ErrPlateForbidden, dobijeno: %v", err)
	}
}

func TestRequestPlateChargesHigherFeeForShortPlate(t *testing.T) {
	svc, _, _ := newTestPlateService()
	r, err := svc.Request(context.Background(), 10, RequestPlateInput{RequestedPlate: "MARKO"})
	if err != nil {
		t.Fatalf("Request: %v", err)
	}
	if r.FeeAmount != model.VanityPlateShortFee {
		t.Errorf("očekivana viša taksa %.2f za kratku tablicu, dobijeno %.2f", model.VanityPlateShortFee, r.FeeAmount)
	}
}

func TestRequestPlateRejectsCollisionWithExistingVehicle(t *testing.T) {
	svc, vehicles, _ := newTestPlateService()
	vehicles.vehicles[1] = &model.Vehicle{ID: 1, PlateNumber: "MARKO01"}

	if _, err := svc.Request(context.Background(), 10, RequestPlateInput{RequestedPlate: "MARKO01"}); err != ErrPlateTaken {
		t.Errorf("očekivana greška ErrPlateTaken, dobijeno: %v", err)
	}
}

func TestRequestPlateRejectsDuplicatePendingReservation(t *testing.T) {
	svc, _, _ := newTestPlateService()
	if _, err := svc.Request(context.Background(), 10, RequestPlateInput{RequestedPlate: "MARKO01"}); err != nil {
		t.Fatalf("prvi Request: %v", err)
	}
	if _, err := svc.Request(context.Background(), 20, RequestPlateInput{RequestedPlate: "MARKO01"}); err != ErrPlateTaken {
		t.Errorf("očekivana greška ErrPlateTaken, dobijeno: %v", err)
	}
}

func TestDecideApprovePlateNotifiesRequester(t *testing.T) {
	svc, _, notifier := newTestPlateService()
	r, err := svc.Request(context.Background(), 10, RequestPlateInput{RequestedPlate: "MARKO01"})
	if err != nil {
		t.Fatalf("Request: %v", err)
	}

	decided, err := svc.Decide(context.Background(), r.ID, true)
	if err != nil {
		t.Fatalf("Decide: %v", err)
	}
	if decided.Status != model.PlateReservationApproved {
		t.Errorf("očekivan status approved, dobijeno %s", decided.Status)
	}
	if len(notifier.messages) != 1 {
		t.Errorf("očekivano 1 obaveštenje, dobijeno %d", len(notifier.messages))
	}
}

func TestDecideRejectsAlreadyDecidedReservation(t *testing.T) {
	svc, _, _ := newTestPlateService()
	r, err := svc.Request(context.Background(), 10, RequestPlateInput{RequestedPlate: "MARKO01"})
	if err != nil {
		t.Fatalf("Request: %v", err)
	}
	if _, err := svc.Decide(context.Background(), r.ID, true); err != nil {
		t.Fatalf("prva Decide: %v", err)
	}
	if _, err := svc.Decide(context.Background(), r.ID, false); err != ErrReservationNotPending {
		t.Errorf("očekivana greška ErrReservationNotPending, dobijeno: %v", err)
	}
}
