package model

import "time"

// Notification je elektronsko obaveštenje građaninu.
type Notification struct {
	ID        int64     `json:"id"`
	CitizenID int64     `json:"citizenId"`
	Message   string    `json:"message"`
	Type      string    `json:"type"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"createdAt"`
}
