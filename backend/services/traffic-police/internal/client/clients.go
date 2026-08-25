package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"euprava/traffic-police/internal/model"
)

// ErrVehicleNotFound se vraća kad Vehicles servis ne poznaje traženu tablicu.
var ErrVehicleNotFound = errors.New("vozilo sa unetom tablicom ne postoji")

// CitizenClient dobavlja podatke o građaninu od Citizen servisa.
// NotificationClient šalje obaveštenja Notification servisu.
// Interfejsi omogućavaju mock-ovanje u testovima (Dependency Inversion).
type CitizenClient interface {
	GetCitizen(ctx context.Context, id int64) (*model.CitizenInfo, error)
}

type NotificationClient interface {
	Notify(ctx context.Context, citizenID int64, message, notifType string) error
}

// VehiclesClient pronalazi vlasnika vozila po registarskoj tablici — koristi
// se kad prekršaj snimi kamera i zna se samo tablica, ne i vozač (drugi,
// samostalan smer komunikacije naspram Vehicles → Traffic Police).
type VehiclesClient interface {
	OwnerByPlate(ctx context.Context, plate string) (*model.VehicleOwnerInfo, error)
}

// ---------------- HTTP implementacije ----------------

type HTTPCitizenClient struct {
	baseURL string
	http    *http.Client
}

func NewHTTPCitizenClient(baseURL string) *HTTPCitizenClient {
	return &HTTPCitizenClient{baseURL: baseURL, http: &http.Client{Timeout: 5 * time.Second}}
}

func (c *HTTPCitizenClient) GetCitizen(ctx context.Context, id int64) (*model.CitizenInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/citizens/%d", c.baseURL, id), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("citizen servis vratio status %d", resp.StatusCode)
	}
	var info model.CitizenInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

type HTTPNotificationClient struct {
	baseURL string
	http    *http.Client
}

func NewHTTPNotificationClient(baseURL string) *HTTPNotificationClient {
	return &HTTPNotificationClient{baseURL: baseURL, http: &http.Client{Timeout: 5 * time.Second}}
}

func (c *HTTPNotificationClient) Notify(ctx context.Context, citizenID int64, message, notifType string) error {
	body, err := json.Marshal(map[string]any{
		"citizenId": citizenID,
		"message":   message,
		"type":      notifType,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/notifications", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("notification servis vratio status %d", resp.StatusCode)
	}
	return nil
}

type HTTPVehiclesClient struct {
	baseURL string
	http    *http.Client
}

func NewHTTPVehiclesClient(baseURL string) *HTTPVehiclesClient {
	return &HTTPVehiclesClient{baseURL: baseURL, http: &http.Client{Timeout: 5 * time.Second}}
}

func (c *HTTPVehiclesClient) OwnerByPlate(ctx context.Context, plate string) (*model.VehicleOwnerInfo, error) {
	url := fmt.Sprintf("%s/vehicles/owner-by-plate/%s", c.baseURL, plate)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrVehicleNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vehicles servis vratio status %d", resp.StatusCode)
	}
	var info model.VehicleOwnerInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}
