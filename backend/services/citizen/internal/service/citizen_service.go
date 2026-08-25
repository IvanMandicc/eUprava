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

// SearchByJMBG pronalazi građanina po JMBG-u — koristi ga policajac pri
// evidentiranju vozača umesto unosa internog ID-ja iz baze.
func (s *CitizenService) SearchByJMBG(ctx context.Context, jmbg string) (*model.User, error) {
	return s.users.GetByJMBG(ctx, jmbg)
}

// jmbgSuggestionLimit ograničava broj predloga u autocomplete listi.
const jmbgSuggestionLimit = 8

// SuggestByJMBGPrefix vraća do nekoliko građana čiji JMBG počinje unetim
// prefiksom — puni padajuću listu dok policajac kuca, bez čekanja na
// ceo broj.
func (s *CitizenService) SuggestByJMBGPrefix(ctx context.Context, prefix string) ([]model.User, error) {
	return s.users.SuggestByJMBGPrefix(ctx, prefix, jmbgSuggestionLimit)
}

// ListUsers vraća sve korisnike sistema (koristi ga administrator).
func (s *CitizenService) ListUsers(ctx context.Context) ([]model.User, error) {
	return s.users.List(ctx)
}
