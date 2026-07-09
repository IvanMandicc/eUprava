package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"euprava/citizen/internal/model"
	"euprava/citizen/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("pogrešan email ili lozinka")
	ErrEmailTaken         = errors.New("korisnik sa ovim email-om već postoji")
)

// RegisterInput su podaci za registraciju građanina.
type RegisterInput struct {
	JMBG      string `json:"jmbg" binding:"required,len=13"`
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	Address   string `json:"address"`
}

// AuthService sadrži poslovnu logiku registracije i prijave.
type AuthService struct {
	users     repository.UserRepository
	jwtSecret []byte
	tokenTTL  time.Duration
}

func NewAuthService(users repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{users: users, jwtSecret: []byte(jwtSecret), tokenTTL: 24 * time.Hour}
}

// Register kreira novog građanina sa heširanom lozinkom.
// Javna registracija UVEK daje ulogu citizen — niko ne može sam sebi
// da dodeli veća ovlašćenja.
func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*model.User, error) {
	return s.createUser(ctx, in, model.RoleCitizen)
}

// CreateOfficer kreira nalog policajca — poziva ga isključivo administrator.
func (s *AuthService) CreateOfficer(ctx context.Context, in RegisterInput) (*model.User, error) {
	return s.createUser(ctx, in, model.RoleOfficer)
}

func (s *AuthService) createUser(ctx context.Context, in RegisterInput, role model.Role) (*model.User, error) {
	if _, err := s.users.GetByEmail(ctx, in.Email); err == nil {
		return nil, ErrEmailTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &model.User{
		JMBG:         in.JMBG,
		FirstName:    in.FirstName,
		LastName:     in.LastName,
		Email:        in.Email,
		PasswordHash: string(hash),
		Role:         role,
		Address:      in.Address,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// Login proverava kredencijale i izdaje JWT token.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, *model.User, error) {
	u, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return "", nil, ErrInvalidCredentials
	}
	if err != nil {
		return "", nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return "", nil, ErrInvalidCredentials
	}

	claims := jwt.MapClaims{
		"sub":  u.ID,
		"role": string(u.Role),
		"name": u.FirstName + " " + u.LastName,
		"exp":  time.Now().Add(s.tokenTTL).Unix(),
		"iat":  time.Now().Unix(),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
	if err != nil {
		return "", nil, err
	}
	return token, u, nil
}

// SeedOfficer upisuje podrazumevani nalog policajca ako ne postoji (za demo).
func (s *AuthService) SeedOfficer(ctx context.Context, email, password string) error {
	return s.seed(ctx, email, password, model.RoleOfficer, "Saobraćajna", "Policija", "0000000000000")
}

// SeedAdmin upisuje podrazumevani nalog administratora ako ne postoji.
func (s *AuthService) SeedAdmin(ctx context.Context, email, password string) error {
	return s.seed(ctx, email, password, model.RoleAdmin, "Sistem", "Administrator", "0000000000001")
}

func (s *AuthService) seed(ctx context.Context, email, password string, role model.Role, firstName, lastName, jmbg string) error {
	if _, err := s.users.GetByEmail(ctx, email); err == nil {
		return nil
	} else if !errors.Is(err, repository.ErrNotFound) {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.users.Create(ctx, &model.User{
		JMBG:         jmbg,
		FirstName:    firstName,
		LastName:     lastName,
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
		Address:      "MUP Republike Srbije",
	})
}
