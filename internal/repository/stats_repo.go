package repository

import (
	"database/sql"
)

type StatsResponse struct {
	TierCounts       map[string]int `json:"tier_counts"`
	DepartmentCounts map[string]int `json:"department_counts"`
	MostActive       string         `json:"most_active"`
	MostActiveCount  int            `json:"most_active_count"`
	CreatedThisMonth int            `json:"created_this_month"`
	TotalScrolls     int            `json:"total_scrolls"`
	TotalVillagers   int            `json:"total_villagers"`
}

type StatsRepository struct {
	db *sql.DB
}

func NewStatsRepository(db *sql.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

func (r *StatsRepository) GetStats() (*StatsResponse, error) {
	resp := &StatsResponse{
		TierCounts:       make(map[string]int),
		DepartmentCounts: make(map[string]int),
	}

	rows, err := r.db.Query(`SELECT classification, COUNT(*) FROM documents GROUP BY classification`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var tier int
		var count int
		if err := rows.Scan(&tier, &count); err != nil {
			return nil, err
		}
		key := tierToString(tier)
		resp.TierCounts[key] = count
	}

	rows2, err := r.db.Query(`SELECT department, COUNT(*) FROM documents GROUP BY department`)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var dept string
		var count int
		if err := rows2.Scan(&dept, &count); err != nil {
			return nil, err
		}
		resp.DepartmentCounts[dept] = count
	}

	row := r.db.QueryRow(`SELECT created_by, COUNT(*) as cnt FROM documents GROUP BY created_by ORDER BY cnt DESC LIMIT 1`)
	if err := row.Scan(&resp.MostActive, &resp.MostActiveCount); err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	r.db.QueryRow(`SELECT COUNT(*) FROM documents WHERE created_at >= date('now', 'start of month')`).Scan(&resp.CreatedThisMonth)
	r.db.QueryRow(`SELECT COUNT(*) FROM documents`).Scan(&resp.TotalScrolls)
	r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&resp.TotalVillagers)

	return resp, nil
}

func tierToString(tier int) string {
	switch tier {
	case 0:
		return "TOWN NOTICE"
	case 1:
		return "GUILD SEALED"
	case 2:
		return "COUNCIL SEALED"
	case 3:
		return "VAULT SEALED"
	case 4:
		return "ARCANE SEALED"
	case 5:
		return "JUNIMO SCRIPT"
	default:
		return "UNKNOWN"
	}
}
