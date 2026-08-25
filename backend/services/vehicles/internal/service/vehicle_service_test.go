package service

import (
	"context"
	"testing"
	"time"

	"euprava/vehicles/internal/model"
	"euprava/vehicles/internal/repository"
)

// In-memory mock implementacije interfejsa — dokaz da service sloj zavisi
// samo od apstrakcija (Dependency Inversion princip), isti obrazac kao
// traffic-police/internal/service/violation_service_test.go.

type mockVehicleRepo struct {
	vehicles map[int64]*model.Vehicle
	nextID   int64
}

func (m *mockVehicleRepo) Create(_ context.Context, v *model.Vehicle) error {
	m.nextID++
	v.ID = m.nextID
	v.CreatedAt = time.Now()
	cp := *v
	m.vehicles[v.ID] = &cp
	return nil
}

func (m *mockVehicleRepo) GetByID(_ context.Context, id int64) (*model.Vehicle, error) {
	v, ok := m.vehicles[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	cp := *v
	return &cp, nil
}

func (m *mockVehicleRepo) GetByVIN(_ context.Context, vin string) (*model.Vehicle, error) {
	for _, v := range m.vehicles {
		if v.VIN == vin {
			cp := *v
			return &cp, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *mockVehicleRepo) GetByPlate(_ context.Context, plate string) (*model.Vehicle, error) {
	for _, v := range m.vehicles {
		if v.PlateNumber == plate {
			cp := *v
			return &cp, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *mockVehicleRepo) PlateExists(_ context.Context, plate string) (bool, error) {
	for _, v := range m.vehicles {
		if v.PlateNumber == plate {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockVehicleRepo) ListByOwner(_ context.Context, ownerCitizenID int64) ([]model.Vehicle, error) {
	out := []model.Vehicle{}
	for _, v := range m.vehicles {
		if v.OwnerCitizenID == ownerCitizenID {
			out = append(out, *v)
		}
	}
	return out, nil
}

func (m *mockVehicleRepo) List(_ context.Context) ([]model.Vehicle, error) { return nil, nil }

func (m *mockVehicleRepo) Update(_ context.Context, v *model.Vehicle) error {
	if _, ok := m.vehicles[v.ID]; !ok {
		return repository.ErrNotFound
	}
	cp := *v
	m.vehicles[v.ID] = &cp
	return nil
}

type mockTransferRepo struct {
	transfers []model.OwnershipTransfer
	nextID    int64
}

func (m *mockTransferRepo) Create(_ context.Context, t *model.OwnershipTransfer) error {
	m.nextID++
	t.ID = m.nextID
	m.transfers = append(m.transfers, *t)
	return nil
}

func (m *mockTransferRepo) ListByVehicle(_ context.Context, vehicleID int64) ([]model.OwnershipTransfer, error) {
	out := []model.OwnershipTransfer{}
	for _, t := range m.transfers {
		if t.VehicleID == vehicleID {
			out = append(out, t)
		}
	}
	return out, nil
}

type mockTheftRepo struct {
	reports map[int64]*model.TheftReport
	nextID  int64
}

func (m *mockTheftRepo) Create(_ context.Context, r *model.TheftReport) error {
	m.nextID++
	r.ID = m.nextID
	cp := *r
	m.reports[r.ID] = &cp
	return nil
}

func (m *mockTheftRepo) GetOpenByVehicle(_ context.Context, vehicleID int64) (*model.TheftReport, error) {
	for _, r := range m.reports {
		if r.VehicleID == vehicleID && r.Status == model.TheftReported {
			cp := *r
			return &cp, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *mockTheftRepo) MarkFound(_ context.Context, id int64) error {
	r, ok := m.reports[id]
	if !ok {
		return repository.ErrNotFound
	}
	r.Status = model.TheftFound
	return nil
}

type mockCitizenClient struct {
	citizens map[int64]*model.CitizenInfo
}

func (m *mockCitizenClient) GetCitizen(_ context.Context, id int64) (*model.CitizenInfo, error) {
	c, ok := m.citizens[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return c, nil
}

type mockTrafficClient struct {
	summary *model.UnpaidFinesSummary
	err     error
}

func (m *mockTrafficClient) UnpaidFinesSummary(_ context.Context, _ int64) (*model.UnpaidFinesSummary, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.summary, nil
}

type mockNotifier struct {
	messages []string
}

func (m *mockNotifier) Notify(_ context.Context, _ int64, message, _ string) error {
	m.messages = append(m.messages, message)
	return nil
}

// ---------------- pomoćna funkcija ----------------

func newTestService() (*VehicleService, *mockVehicleRepo, *mockTheftRepo, *mockTrafficClient, *mockNotifier) {
	vehicles := &mockVehicleRepo{vehicles: map[int64]*model.Vehicle{}}
	transfers := &mockTransferRepo{}
	thefts := &mockTheftRepo{reports: map[int64]*model.TheftReport{}}
	citizens := &mockCitizenClient{citizens: map[int64]*model.CitizenInfo{
		10: {ID: 10, Role: "citizen", FirstName: "Marko", LastName: "Marković"},
		20: {ID: 20, Role: "citizen", FirstName: "Ana", LastName: "Anić"},
		30: {ID: 30, Role: "officer", FirstName: "Petar", LastName: "Petrović"},
	}}
	traffic := &mockTrafficClient{summary: &model.UnpaidFinesSummary{UnpaidCount: 0, UnpaidTotal: 0}}
	notifier := &mockNotifier{}
	return NewVehicleService(vehicles, transfers, thefts, citizens, traffic, notifier), vehicles, thefts, traffic, notifier
}

func validRegisterInput(ownerID int64) RegisterVehicleInput {
	return RegisterVehicleInput{
		OwnerCitizenID:            ownerID,
		VIN:                       "1HGCM82633A004352",
		Make:                      "Škoda",
		Model:                     "Octavia",
		Year:                      2020,
		Category:                  "M1",
		FuelType:                  "dizel",
		InsuranceValidUntil:       time.Now().Add(180 * 24 * time.Hour),
		TechInspectionValidUntil:  time.Now().Add(180 * 24 * time.Hour),
	}
}

// ---------------- testovi poslovnih pravila ----------------

func TestRegisterGeneratesPlateAndValidatesOwner(t *testing.T) {
	svc, _, _, _, _ := newTestService()

	v, err := svc.Register(context.Background(), validRegisterInput(10))
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if v.PlateNumber == "" {
		t.Error("očekivana automatski generisana tablica")
	}
	if v.Status != model.VehicleRegistered {
		t.Errorf("očekivan status registered, dobijeno %s", v.Status)
	}
	if !v.RegistrationValidUntil.After(time.Now()) {
		t.Error("rok registracije treba da bude u budućnosti")
	}
}

func TestRegisterRejectsNonCitizenOwner(t *testing.T) {
	svc, _, _, _, _ := newTestService()

	if _, err := svc.Register(context.Background(), validRegisterInput(30)); err != ErrOwnerNotCitizen {
		t.Errorf("očekivana greška ErrOwnerNotCitizen, dobijeno: %v", err)
	}
}

func TestRegisterRejectsDuplicateVIN(t *testing.T) {
	svc, _, _, _, _ := newTestService()

	if _, err := svc.Register(context.Background(), validRegisterInput(10)); err != nil {
		t.Fatalf("prvi Register: %v", err)
	}
	if _, err := svc.Register(context.Background(), validRegisterInput(20)); err != ErrDuplicateVIN {
		t.Errorf("očekivana greška ErrDuplicateVIN, dobijeno: %v", err)
	}
}

func TestTransferBlockedWhenStolen(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	v, err := svc.Register(context.Background(), validRegisterInput(10))
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if _, err := svc.ReportTheft(context.Background(), v.ID, 10, "citizen"); err != nil {
		t.Fatalf("ReportTheft: %v", err)
	}

	if _, err := svc.Transfer(context.Background(), v.ID, TransferVehicleInput{NewOwnerCitizenID: 20}); err != ErrVehicleStolen {
		t.Errorf("očekivana greška ErrVehicleStolen, dobijeno: %v", err)
	}
}

func TestTransferBlockedByUnpaidFines(t *testing.T) {
	svc, _, _, traffic, _ := newTestService()
	v, err := svc.Register(context.Background(), validRegisterInput(10))
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	traffic.summary = &model.UnpaidFinesSummary{UnpaidCount: 3, UnpaidTotal: 60000}

	if _, err := svc.Transfer(context.Background(), v.ID, TransferVehicleInput{NewOwnerCitizenID: 20}); err != ErrBlockedByFines {
		t.Errorf("očekivana greška ErrBlockedByFines, dobijeno: %v", err)
	}
}

func TestTransferSucceedsAndResetsRegistration(t *testing.T) {
	svc, _, _, _, notifier := newTestService()
	v, err := svc.Register(context.Background(), validRegisterInput(10))
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	originalDeadline := v.RegistrationValidUntil

	updated, err := svc.Transfer(context.Background(), v.ID, TransferVehicleInput{NewOwnerCitizenID: 20})
	if err != nil {
		t.Fatalf("Transfer: %v", err)
	}
	if updated.OwnerCitizenID != 20 {
		t.Errorf("očekivan novi vlasnik 20, dobijeno %d", updated.OwnerCitizenID)
	}
	if !updated.RegistrationValidUntil.Before(originalDeadline) {
		t.Error("rok registracije treba da bude poništen (u prošlosti/sada) nakon prenosa")
	}
	if len(notifier.messages) != 2 {
		t.Errorf("očekivana 2 obaveštenja (stari i novi vlasnik), dobijeno %d", len(notifier.messages))
	}
}

func TestRenewRejectsExpiredTechInspection(t *testing.T) {
	svc, vehicles, _, _, _ := newTestService()
	v, err := svc.Register(context.Background(), validRegisterInput(10))
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	stored := vehicles.vehicles[v.ID]
	stored.TechInspectionValidUntil = time.Now().Add(-24 * time.Hour)

	if _, err := svc.RenewRegistration(context.Background(), v.ID); err != ErrTechInspectionExpired {
		t.Errorf("očekivana greška ErrTechInspectionExpired, dobijeno: %v", err)
	}
}

func TestRenewSucceedsAndExtendsDeadline(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	v, err := svc.Register(context.Background(), validRegisterInput(10))
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	updated, err := svc.RenewRegistration(context.Background(), v.ID)
	if err != nil {
		t.Fatalf("RenewRegistration: %v", err)
	}
	remaining := time.Until(updated.RegistrationValidUntil)
	if remaining < 364*24*time.Hour || remaining > 366*24*time.Hour {
		t.Errorf("novi rok registracije treba da bude ~godinu dana od sada, preostalo: %v", remaining)
	}
}

func TestReportTheftRejectsNonOwnerCitizen(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	v, err := svc.Register(context.Background(), validRegisterInput(10))
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	if _, err := svc.ReportTheft(context.Background(), v.ID, 20, "citizen"); err != ErrNotOwner {
		t.Errorf("očekivana greška ErrNotOwner, dobijeno: %v", err)
	}
}

func TestReportTheftThenFoundRestoresStatus(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	v, err := svc.Register(context.Background(), validRegisterInput(10))
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	if _, err := svc.ReportTheft(context.Background(), v.ID, 10, "citizen"); err != nil {
		t.Fatalf("ReportTheft: %v", err)
	}
	stolen, err := svc.Get(context.Background(), v.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if stolen.Status != model.VehicleStolen {
		t.Fatalf("očekivan status stolen, dobijeno %s", stolen.Status)
	}

	found, err := svc.ReportFound(context.Background(), v.ID, 10, "citizen")
	if err != nil {
		t.Fatalf("ReportFound: %v", err)
	}
	if found.Status != model.VehicleRegistered {
		t.Errorf("očekivan status registered nakon pronalaska, dobijeno %s", found.Status)
	}
}
