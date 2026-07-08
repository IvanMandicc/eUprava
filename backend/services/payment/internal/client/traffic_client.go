package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"euprava/payment/internal/model"
)

// TrafficClient komunicira sa Traffic Police servisom: čita podatke o kazni
// i evidentira plaćanje. NotificationClient šalje obaveštenja.
type TrafficClient interface {
	GetFine(ctx context.Context, fineID int64) (*model.FineInfo, error)
	PayFine(ctx context.Context, fineID int64) error
}

type NotificationClient interface {
	Notify(ctx context.Context, citizenID int64, message, notifType string) error
}

// ---------------- HTTP implementacije ----------------

type HTTPTrafficClient struct {
	baseURL string
	http    *http.Client
}

func NewHTTPTrafficClient(baseURL string) *HTTPTrafficClient {
	return &HTTPTrafficClient{baseURL: baseURL, http: &http.Client{Timeout: 5 * time.Second}}
}

func (c *HTTPTrafficClient) GetFine(ctx context.Context, fineID int64) (*model.FineInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/fines/%d", c.baseURL, fineID), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("kazna %d ne postoji", fineID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("traffic servis vratio status %d", resp.StatusCode)
	}
	var info model.FineInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (c *HTTPTrafficClient) PayFine(ctx context.Context, fineID int64) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/fines/%d/pay", c.baseURL, fineID), nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("traffic servis vratio status %d", resp.StatusCode)
	}
	return nil
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
