package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"euprava/traffic-police/internal/model"
)

// CitizenClient dobavlja podatke o građaninu od Citizen servisa.
// NotificationClient šalje obaveštenja Notification servisu.
// Interfejsi omogućavaju mock-ovanje u testovima (Dependency Inversion).
type CitizenClient interface {
	GetCitizen(ctx context.Context, id int64) (*model.CitizenInfo, error)
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
