package domain

import "time"

type User struct {
	ID           string         `json:"id"`
	Username     string         `json:"username"`
	PasswordHash string         `json:"-"`
	Email        string         `json:"email"`
	Role         Role           `json:"role"`
	Clearance    ClearanceLevel `json:"tier"`
	Department      Department        `json:"department"`
	Active       bool           `json:"active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}
