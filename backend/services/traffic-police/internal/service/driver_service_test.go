package service

import (
	"context"
	"testing"

	"euprava/traffic-police/internal/model"
	"euprava/traffic-police/internal/repository"
)

// mockCitizenClient simulira Citizen servis — vraća korisnike sa različitim ulogama.
type mockCitizenClient struct {
	users map[int64]*model.CitizenInfo
}

func (m *mockCitizenClient) GetCitizen(_ context.Context, id int64) (*model.CitizenInfo, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, repository.ErrNotFound
}

func newTestDriverService() *DriverService {
	drivers := &mockDriverRepo{drivers: map[int64]*model.Driver{}}
	citizens := &mockCitizenClient{users: map[int64]*model.CitizenInfo{
		1: {ID: 1, FirstName: "Sistem", LastName: "Administrator", Role: "admin"},
		2: {ID: 2, FirstName: "Saobraćajna", LastName: "Policija", Role: "officer"},
		3: {ID: 3, FirstName: "Petar", LastName: "Petrović", Role: "citizen"},
	}}
	return NewDriverService(drivers, citizens)
}

func TestRegisterDriverForCitizenSucceeds(t *testing.T) {
	svc := newTestDriverService()
	d, err := svc.Register(context.Background(), RegisterDriverInput{CitizenID: 3, LicenseNumber: "NS-001"})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if d.CitizenID != 3 || d.LicenseStatus != model.LicenseValid {
		t.Errorf("neočekivan vozač: %+v", d)
	}
}

func TestRegisterDriverRejectsNonCitizenRoles(t *testing.T) {
	svc := newTestDriverService()
	for _, id := range []int64{1, 2} { // admin i policajac
		if _, err := svc.Register(context.Background(), RegisterDriverInput{CitizenID: id, LicenseNumber: "XX-001"}); err != ErrNotACitizen {
			t.Errorf("citizenId=%d: očekivana greška ErrNotACitizen, dobijeno: %v", id, err)
		}
	}
}

func TestRegisterDriverRejectsUnknownCitizen(t *testing.T) {
	svc := newTestDriverService()
	if _, err := svc.Register(context.Background(), RegisterDriverInput{CitizenID: 999, LicenseNumber: "XX-002"}); err == nil {
		t.Error("očekivana greška za nepostojećeg građanina")
	}
}
