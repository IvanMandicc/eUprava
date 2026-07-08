package model

import "time"

// PaymentStatus je status plaćanja.
type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentCompleted PaymentStatus = "completed"
)

// Payment je evidencija plaćanja novčane kazne.
type Payment struct {
	ID          int64         `json:"id"`
	FineID      int64         `json:"fineId"`
	CitizenID   int64         `json:"citizenId"`
	Amount      float64       `json:"amount"`
	Status      PaymentStatus `json:"status"`
	CreatedAt   time.Time     `json:"createdAt"`
	CompletedAt *time.Time    `json:"completedAt"`
}

// FineInfo su podaci o kazni dobijeni od Traffic Police servisa.
type FineInfo struct {
	ID        int64   `json:"id"`
	Amount    float64 `json:"amount"`
	Paid      bool    `json:"paid"`
	CitizenID int64   `json:"citizenId"`
}
