package service

import (
	"context"
	"log"

	"euprava/notification/internal/model"
	"euprava/notification/internal/repository"
)

// CreateNotificationInput su podaci za novo obaveštenje (šalju ga drugi servisi).
type CreateNotificationInput struct {
	CitizenID int64  `json:"citizenId" binding:"required"`
	Message   string `json:"message" binding:"required"`
	Type      string `json:"type"`
}

// NotificationService čuva obaveštenja; slanje email/SMS poruke je
// simulirano logovanjem.
type NotificationService struct {
	notifications repository.NotificationRepository
}

func NewNotificationService(notifications repository.NotificationRepository) *NotificationService {
	return &NotificationService{notifications: notifications}
}

func (s *NotificationService) Create(ctx context.Context, in CreateNotificationInput) (*model.Notification, error) {
	if in.Type == "" {
		in.Type = "INFO"
	}
	n := &model.Notification{CitizenID: in.CitizenID, Message: in.Message, Type: in.Type}
	if err := s.notifications.Create(ctx, n); err != nil {
		return nil, err
	}
	// Simulacija slanja email/SMS poruke.
	log.Printf("[EMAIL/SMS → građanin %d] %s", n.CitizenID, n.Message)
	return n, nil
}

func (s *NotificationService) ListByCitizen(ctx context.Context, citizenID int64) ([]model.Notification, error) {
	return s.notifications.ListByCitizen(ctx, citizenID)
}

func (s *NotificationService) MarkRead(ctx context.Context, id, citizenID int64) error {
	return s.notifications.MarkRead(ctx, id, citizenID)
}
