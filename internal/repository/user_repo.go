package repository

import (
	"database/sql"
	"time"

	"classified-vault/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByUsername(username string) (*domain.User, error) {
	var u domain.User
	var faction string
	err := r.db.QueryRow(
		`SELECT id, username, password_hash, email, role, clearance, faction, active, created_at, updated_at
		 FROM users WHERE username = ?`, username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Email, &u.Role, &u.Clearance, &faction, &u.Active, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u.Faction = domain.Faction(faction)
	return &u, nil
}

func (r *UserRepository) FindByID(id string) (*domain.User, error) {
	var u domain.User
	var faction string
	err := r.db.QueryRow(
		`SELECT id, username, password_hash, email, role, clearance, faction, active, created_at, updated_at
		 FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Email, &u.Role, &u.Clearance, &faction, &u.Active, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u.Faction = domain.Faction(faction)
	return &u, nil
}

func (r *UserRepository) FindAll() ([]*domain.User, error) {
	rows, err := r.db.Query(
		`SELECT id, username, password_hash, email, role, clearance, faction, active, created_at, updated_at
		 FROM users ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var u domain.User
		var faction string
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Email, &u.Role, &u.Clearance, &faction, &u.Active, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		u.Faction = domain.Faction(faction)
		users = append(users, &u)
	}
	return users, rows.Err()
}

func (r *UserRepository) Create(u *domain.User) error {
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now
	if u.Faction == "" {
		u.Faction = domain.FactionMuseum
	}
	_, err := r.db.Exec(
		`INSERT INTO users (id, username, password_hash, email, role, clearance, faction, active, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.Username, u.PasswordHash, u.Email, u.Role, u.Clearance, string(u.Faction), u.Active, u.CreatedAt, u.UpdatedAt,
	)
	return err
}

func (r *UserRepository) Update(u *domain.User) error {
	u.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		`UPDATE users SET username=?, password_hash=?, email=?, role=?, clearance=?, faction=?, active=?, updated_at=?
		 WHERE id=?`,
		u.Username, u.PasswordHash, u.Email, u.Role, u.Clearance, string(u.Faction), u.Active, u.UpdatedAt, u.ID,
	)
	return err
}

func (r *UserRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	return err
}
