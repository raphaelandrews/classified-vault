package auth

import (
	"time"

	"classified-vault/internal/domain"
)

type Session struct {
	UserID    string                `json:"user_id"`
	Username  string                `json:"username"`
	Role      domain.Role           `json:"role"`
	Clearance domain.ClearanceLevel `json:"tier"`
	Department   domain.Department        `json:"department"`
	ExpiresAt time.Time             `json:"expires_at"`
}
