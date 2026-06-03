package domain

import "time"

type AuditAction string

const (
	ActionLogin        AuditAction = "VILLAGER_SIGN_IN"
	ActionLogout       AuditAction = "VILLAGER_SIGN_OUT"
	ActionLoginFailed  AuditAction = "SIGN_IN_FAILED"
	ActionScrollRead   AuditAction = "SCROLL_READ"
	ActionScrollCreate AuditAction = "SCROLL_PENNED"
	ActionScrollUpdate AuditAction = "SCROLL_AMENDED"
	ActionScrollDelete AuditAction = "SCROLL_DESTROYED"
	ActionAccessDenied AuditAction = "SEALED"
	ActionUserCreated  AuditAction = "VILLAGER_REGISTERED"
	ActionUserUpdated  AuditAction = "VILLAGER_UPDATED"
	ActionUserDeleted  AuditAction = "VILLAGER_DISMISSED"
)

type AuditLog struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Username  string      `json:"username"`
	Action    AuditAction `json:"action"`
	Resource  string      `json:"resource"`
	IPAddress string      `json:"ip_address"`
	Success   bool        `json:"success"`
	Details   string      `json:"details"`
	Timestamp time.Time   `json:"timestamp"`
}
