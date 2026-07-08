package service

import (
	"context"

	"euprava/citizen/internal/model"
	"euprava/citizen/internal/repository"
)

// CitizenService vraća podatke o građanima (koristi ga i Traffic Police servis).
type CitizenService struct {
	users repository.UserRepository
}

func NewCitizenService(users repository.UserRepository) *CitizenService {
	return &CitizenService{users: users}
}

func (s *CitizenService) GetByID(ctx context.Context, id int64) (*model.User, error) {
	return s.users.GetByID(ctx, id)
}
