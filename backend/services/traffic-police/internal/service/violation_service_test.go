package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"euprava/traffic-police/internal/client"
	"euprava/traffic-police/internal/model"
	"euprava/traffic-police/internal/repository"
)

// In-memory mock implementacije interfejsa — dokaz da service sloj
// zavisi samo od apstrakcija (Dependency Inversion princip).

type mockDriverRepo struct {
	drivers map[int64]*model.Driver
}

func (m *mockDriverRepo) Create(_ context.Context, d *model.Driver) error {
	d.ID = int64(len(m.drivers) + 1)
	d.LicenseStatus = model.LicenseValid
	m.drivers[d.ID] = d
	return nil
}

func (m *mockDriverRepo) GetByID(_ context.Context, id int64) (*model.Driver, error) {
	d, ok := m.drivers[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	copy := *d
	return &copy, nil
}

func (m *mockDriverRepo) GetByCitizenID(_ context.Context, citizenID int64) (*model.Driver, error) {
	for _, d := range m.drivers {
		if d.CitizenID == citizenID {
			copy := *d
			return &copy, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *mockDriverRepo) List(_ context.Context) ([]model.Driver, error) { return nil, nil }

func (m *mockDriverRepo) UpdatePoints(_ context.Context, id int64, points int, status model.LicenseStatus) error {
	d, ok := m.drivers[id]
	if !ok {
		return repository.ErrNotFound
	}
	d.PenaltyPoints = points
	d.LicenseStatus = status
	return nil
}

type mockViolationRepo struct {
	violations map[int64]*model.Violation
	nextID     int64
}

func (m *mockViolationRepo) Create(_ context.Context, v *model.Violation) error {
	m.nextID++
	v.ID = m.nextID
	m.violations[v.ID] = v
	return nil
}

func (m *mockViolationRepo) GetByID(_ context.Context, id int64) (*model.Violation, error) {
	v, ok := m.violations[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	copy := *v
	return &copy, nil
}

func (m *mockViolationRepo) ListByDriver(_ context.Context, driverID int64) ([]model.Violation, error) {
	out := []model.Violation{}
	for _, v := range m.violations {
		if v.DriverID == driverID {
			out = append(out, *v)
		}
	}
	return out, nil
}

func (m *mockViolationRepo) List(_ context.Context) ([]model.Violation, error) { return nil, nil }

func (m *mockViolationRepo) Stats(_ context.Context, _, _ *time.Time) ([]model.ViolationStat, error) {
	return nil, nil
}

func (m *mockViolationRepo) Update(_ context.Context, v *model.Violation) error {
	if _, ok := m.violations[v.ID]; !ok {
		return repository.ErrNotFound
	}
	m.violations[v.ID] = v
	return nil
}

func (m *mockViolationRepo) Delete(_ context.Context, id int64) error {
	if _, ok := m.violations[id]; !ok {
		return repository.ErrNotFound
	}
	delete(m.violations, id)
	return nil
}

type mockFineRepo struct {
	fines  map[int64]*model.Fine
	nextID int64
}

func (m *mockFineRepo) Create(_ context.Context, f *model.Fine) error {
	m.nextID++
	f.ID = m.nextID
	m.fines[f.ID] = f
	return nil
}

func (m *mockFineRepo) GetByID(_ context.Context, id int64) (*model.FineDetails, error) {
	f, ok := m.fines[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return &model.FineDetails{Fine: *f}, nil
}

func (m *mockFineRepo) GetByViolationID(_ context.Context, violationID int64) (*model.Fine, error) {
	for _, f := range m.fines {
		if f.ViolationID == violationID {
			copy := *f
			return &copy, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *mockFineRepo) List(_ context.Context) ([]model.FineDetails, error) { return nil, nil }

func (m *mockFineRepo) ListByCitizen(_ context.Context, _ int64) ([]model.FineDetails, error) {
	return nil, nil
}

func (m *mockFineRepo) MarkPaid(_ context.Context, id int64) error {
	f, ok := m.fines[id]
	if !ok {
		return repository.ErrNotFound
	}
	f.Paid = true
	return nil
}

func (m *mockFineRepo) DeleteByViolationID(_ context.Context, violationID int64) error {
	for id, f := range m.fines {
		if f.ViolationID == violationID {
			delete(m.fines, id)
		}
	}
	return nil
}

func (m *mockFineRepo) UnpaidSummaryByCitizen(_ context.Context, _ int64) (*model.UnpaidFinesSummary, error) {
	return &model.UnpaidFinesSummary{}, nil
}

type mockNotifier struct {
	messages []string
}

func (m *mockNotifier) Notify(_ context.Context, _ int64, message, _ string) error {
	m.messages = append(m.messages, message)
	return nil
}

// mockVehiclesClient simulira Vehicles servis — mapira tablicu na vlasnika.
type mockVehiclesClient struct {
	ownersByPlate map[string]int64
}

func (m *mockVehiclesClient) OwnerByPlate(_ context.Context, plate string) (*model.VehicleOwnerInfo, error) {
	citizenID, ok := m.ownersByPlate[plate]
	if !ok {
		return nil, client.ErrVehicleNotFound
	}
	return &model.VehicleOwnerInfo{OwnerCitizenID: citizenID, PlateNumber: plate}, nil
}

// ---------------- pomoćna funkcija ----------------

func newTestService() (*ViolationService, *mockDriverRepo, *mockFineRepo, *mockNotifier) {
	svc, drivers, fines, notifier, _ := newTestServiceWithVehicles()
	return svc, drivers, fines, notifier
}

func newTestServiceWithVehicles() (*ViolationService, *mockDriverRepo, *mockFineRepo, *mockNotifier, *mockVehiclesClient) {
	drivers := &mockDriverRepo{drivers: map[int64]*model.Driver{
		1: {ID: 1, CitizenID: 10, LicenseNumber: "B-12345", PenaltyPoints: 0, LicenseStatus: model.LicenseValid},
	}}
	violations := &mockViolationRepo{violations: map[int64]*model.Violation{}}
	fines := &mockFineRepo{fines: map[int64]*model.Fine{}}
	notifier := &mockNotifier{}
	vehicles := &mockVehiclesClient{ownersByPlate: map[string]int64{"NS-001-AA": 10}}
	return NewViolationService(drivers, violations, fines, notifier, vehicles), drivers, fines, notifier, vehicles
}

// ---------------- testovi poslovnih pravila ----------------

func TestCreateViolationAddsPointsFineAndNotification(t *testing.T) {
	svc, drivers, fines, notifier := newTestService()

	v, err := svc.Create(context.Background(), CreateViolationInput{
		DriverID: 1, Type: "SPEEDING", Location: "Bulevar oslobođenja", Description: "92 km/h u zoni 50",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if v.Points != 6 || v.FineAmount != 20000 {
		t.Errorf("očekivano 6 poena i 20000 RSD iz šifarnika, dobijeno %d/%f", v.Points, v.FineAmount)
	}
	if d := drivers.drivers[1]; d.PenaltyPoints != 6 {
		t.Errorf("vozač treba da ima 6 poena, ima %d", d.PenaltyPoints)
	}
	if len(fines.fines) != 1 {
		t.Errorf("očekivana automatski kreirana kazna, ima ih %d", len(fines.fines))
	}
	if len(notifier.messages) != 1 || !strings.Contains(notifier.messages[0], "nova kazna") {
		t.Errorf("očekivano obaveštenje o novoj kazni, dobijeno: %v", notifier.messages)
	}
}

func TestLicenseSuspendedAtPointLimit(t *testing.T) {
	svc, drivers, _, notifier := newTestService()
	drivers.drivers[1].PenaltyPoints = 12 // 12 + 6 (SPEEDING) = 18 → limit

	if _, err := svc.Create(context.Background(), CreateViolationInput{DriverID: 1, Type: "SPEEDING"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if d := drivers.drivers[1]; d.LicenseStatus != model.LicenseSuspended {
		t.Errorf("dozvola treba da bude suspendovana, status je %s", d.LicenseStatus)
	}
	found := false
	for _, msg := range notifier.messages {
		if strings.Contains(msg, "suspendovana") {
			found = true
		}
	}
	if !found {
		t.Errorf("očekivano obaveštenje o suspenziji, dobijeno: %v", notifier.messages)
	}
}

func TestCreateByPlateFindsOwnerAndCreatesViolation(t *testing.T) {
	svc, drivers, _, _, _ := newTestServiceWithVehicles()

	v, err := svc.CreateByPlate(context.Background(), CreateViolationByPlateInput{
		PlateNumber: "NS-001-AA", Type: "SPEEDING", Location: "Autoput",
	})
	if err != nil {
		t.Fatalf("CreateByPlate: %v", err)
	}
	if v.DriverID != 1 {
		t.Errorf("očekivan vozač 1 (vlasnik tablice), dobijeno driverId=%d", v.DriverID)
	}
	if drivers.drivers[1].PenaltyPoints != 6 {
		t.Errorf("vozač treba da ima 6 poena posle prekršaja preko tablice, ima %d", drivers.drivers[1].PenaltyPoints)
	}
}

func TestCreateByPlateRejectsUnregisteredOwner(t *testing.T) {
	svc, _, _, _, vehicles := newTestServiceWithVehicles()
	vehicles.ownersByPlate["ZR-999-ZZ"] = 999 // vlasnik postoji, ali nije evidentiran kao vozač

	_, err := svc.CreateByPlate(context.Background(), CreateViolationByPlateInput{
		PlateNumber: "ZR-999-ZZ", Type: "SPEEDING",
	})
	if err != ErrOwnerNotRegistered {
		t.Errorf("očekivana greška ErrOwnerNotRegistered, dobijeno: %v", err)
	}
}

func TestCreateByPlateRejectsUnknownPlate(t *testing.T) {
	svc, _, _, _, _ := newTestServiceWithVehicles()

	_, err := svc.CreateByPlate(context.Background(), CreateViolationByPlateInput{
		PlateNumber: "NEPOSTOJECA", Type: "SPEEDING",
	})
	if err == nil {
		t.Error("očekivana greška za nepostojeću tablicu")
	}
}

func TestDeleteViolationRestoresPoints(t *testing.T) {
	svc, drivers, fines, _ := newTestService()

	v, err := svc.Create(context.Background(), CreateViolationInput{DriverID: 1, Type: "RED_LIGHT"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if drivers.drivers[1].PenaltyPoints != 8 {
		t.Fatalf("očekivano 8 poena posle prekršaja, ima %d", drivers.drivers[1].PenaltyPoints)
	}

	if err := svc.Delete(context.Background(), v.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if d := drivers.drivers[1]; d.PenaltyPoints != 0 || d.LicenseStatus != model.LicenseValid {
		t.Errorf("poeni treba da se vrate na 0 i dozvola bude važeća, dobijeno %d/%s", d.PenaltyPoints, d.LicenseStatus)
	}
	if len(fines.fines) != 0 {
		t.Errorf("kazna treba da bude obrisana sa prekršajem, ima ih %d", len(fines.fines))
	}
}

func TestDeleteViolationWithPaidFineFails(t *testing.T) {
	svc, _, fines, _ := newTestService()

	v, err := svc.Create(context.Background(), CreateViolationInput{DriverID: 1, Type: "NO_SEATBELT"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	for _, f := range fines.fines {
		f.Paid = true
	}
	if err := svc.Delete(context.Background(), v.ID); err != ErrFinePaid {
		t.Errorf("očekivana greška ErrFinePaid, dobijeno: %v", err)
	}
}

func TestUnknownViolationTypeRejected(t *testing.T) {
	svc, _, _, _ := newTestService()
	if _, err := svc.Create(context.Background(), CreateViolationInput{DriverID: 1, Type: "NEPOSTOJEĆI"}); err != ErrUnknownViolationType {
		t.Errorf("očekivana greška ErrUnknownViolationType, dobijeno: %v", err)
	}
}
