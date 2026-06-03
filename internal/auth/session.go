package auth

import (
	"time"

	"classified-vault/internal/domain"
)

type Session struct {
	UserID    string
	Username  string
	Role      domain.Role
	Clearance domain.ClearanceLevel
	ExpiresAt time.Time
}
