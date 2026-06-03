package repository

import (
	"database/sql"
	"time"

	"classified-vault/internal/domain"
)

type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Save(log *domain.AuditLog) error {
	log.Timestamp = time.Now()
	_, err := r.db.Exec(
		`INSERT INTO audit_logs (id, user_id, username, action, resource, ip_address, success, details, timestamp)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		log.ID, log.UserID, log.Username, log.Action, log.Resource, log.IPAddress, log.Success, log.Details, log.Timestamp,
	)
	return err
}

func (r *AuditRepository) FindAll(limit int) ([]*domain.AuditLog, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, username, action, resource, ip_address, success, details, timestamp
		 FROM audit_logs ORDER BY timestamp DESC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		var l domain.AuditLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.Username, &l.Action, &l.Resource, &l.IPAddress, &l.Success, &l.Details, &l.Timestamp); err != nil {
			return nil, err
		}
		logs = append(logs, &l)
	}
	return logs, rows.Err()
}

func (r *AuditRepository) FindByUser(userID string, limit int) ([]*domain.AuditLog, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, username, action, resource, ip_address, success, details, timestamp
		 FROM audit_logs WHERE user_id = ? ORDER BY timestamp DESC LIMIT ?`, userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		var l domain.AuditLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.Username, &l.Action, &l.Resource, &l.IPAddress, &l.Success, &l.Details, &l.Timestamp); err != nil {
			return nil, err
		}
		logs = append(logs, &l)
	}
	return logs, rows.Err()
}
