package repository

import (
	"context"
	"errors"

	"euprava/citizen/internal/model"
)

// ErrNotFound se vraća kada traženi zapis ne postoji.
var ErrNotFound = errors.New("zapis nije pronađen")

// UserRepository apstrahuje skladištenje korisnika (Interface Segregation +
// Dependency Inversion: service sloj zavisi od ovog interfejsa, ne od Postgresa).
type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByJMBG(ctx context.Context, jmbg string) (*model.User, error)
	List(ctx context.Context) ([]model.User, error)
}
