package model

import "time"

// Role je uloga korisnika u sistemu.
type Role string

const (
	RoleCitizen Role = "citizen"
	RoleOfficer Role = "officer"
	RoleAdmin   Role = "admin"
)

// User predstavlja građanina, policajca ili administratora.
type User struct {
	ID           int64     `json:"id"`
	JMBG         string    `json:"jmbg"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	Address      string    `json:"address"`
	CreatedAt    time.Time `json:"createdAt"`
}
