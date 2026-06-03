package domain

import "time"

type AuditAction string

const (
	ActionLogin          AuditAction = "LOGIN"
	ActionLogout         AuditAction = "LOGOUT"
	ActionLoginFailed    AuditAction = "LOGIN_FAILED"
	ActionDocumentRead   AuditAction = "DOCUMENT_READ"
	ActionDocumentCreate AuditAction = "DOCUMENT_CREATE"
	ActionDocumentUpdate AuditAction = "DOCUMENT_UPDATE"
	ActionDocumentDelete AuditAction = "DOCUMENT_DELETE"
	ActionAccessDenied   AuditAction = "ACCESS_DENIED"
	ActionUserCreated    AuditAction = "USER_CREATED"
	ActionUserUpdated    AuditAction = "USER_UPDATED"
	ActionUserDeleted    AuditAction = "USER_DELETED"
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
