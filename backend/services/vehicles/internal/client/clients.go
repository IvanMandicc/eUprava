package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"euprava/vehicles/internal/model"
)

// CitizenClient dobavlja podatke o vlasniku od Citizen servisa.
// TrafficClient proverava neplaćene kazne kod Traffic Police servisa.
// NotificationClient šalje obaveštenja Notification servisu.
// Interfejsi omogućavaju mock-ovanje u testovima (Dependency Inversion) —
// isti obrazac kao kod traffic-police i payment servisa.
type CitizenClient interface {
	GetCitizen(ctx context.Context, id int64) (*model.CitizenInfo, error)
}

type TrafficClient interface {
	UnpaidFinesSummary(ctx context.Context, citizenID int64) (*model.UnpaidFinesSummary, error)
}

type NotificationClient interface {
	Notify(ctx context.Context, citizenID int64, message, notifType string) error
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

// HTTPTrafficClient poziva interni endpoint Traffic Police servisa koji
// nije izložen kroz gateway (isti obrazac kao PUT /fines/{id}/pay koji
// Payment servis koristi).
type HTTPTrafficClient struct {
	baseURL string
	http    *http.Client
}

func NewHTTPTrafficClient(baseURL string) *HTTPTrafficClient {
	return &HTTPTrafficClient{baseURL: baseURL, http: &http.Client{Timeout: 5 * time.Second}}
}

func (c *HTTPTrafficClient) UnpaidFinesSummary(ctx context.Context, citizenID int64) (*model.UnpaidFinesSummary, error) {
	url := fmt.Sprintf("%s/fines/citizen/%d/unpaid-summary", c.baseURL, citizenID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("traffic police servis vratio status %d", resp.StatusCode)
	}
	var summary model.UnpaidFinesSummary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return nil, err
	}
	return &summary, nil
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
