package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

func Connect(path string) (*sql.DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

func RunMigrations(db *sql.DB) error {
	if err := ensureMigrationsTable(db); err != nil {
		return err
	}

	entries, err := os.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}

	type migration struct {
		version int
		name    string
		path    string
	}

	var pending []migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) != 2 {
			continue
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		if version == 0 {
			continue
		}

		name := strings.TrimSuffix(parts[1], ".sql")
		applied, err := isMigrationApplied(db, version)
		if err != nil {
			return fmt.Errorf("check migration %d: %w", version, err)
		}

		if applied {
			continue
		}

		pending = append(pending, migration{
			version: version,
			name:    name,
			path:    filepath.Join("migrations", entry.Name()),
		})
	}

	sort.Slice(pending, func(i, j int) bool {
		return pending[i].version < pending[j].version
	})

	for _, m := range pending {
		content, err := os.ReadFile(m.path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", m.path, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx for migration %d: %w", m.version, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("execute migration %d (%s): %w", m.version, m.name, err)
		}

		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (version, name) VALUES (?, ?)",
			m.version, m.name,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %d: %w", m.version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", m.version, err)
		}
	}

	return nil
}

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    INTEGER PRIMARY KEY,
		name       TEXT NOT NULL,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	return err
}

func isMigrationApplied(db *sql.DB, version int) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
