CREATE TABLE IF NOT EXISTS users (
    id           TEXT PRIMARY KEY,
    username     TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    email        TEXT UNIQUE NOT NULL,
    role         TEXT NOT NULL DEFAULT 'intern',
    clearance    INTEGER NOT NULL DEFAULT 0,
    active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO users (id, username, password_hash, email, role, clearance)
VALUES (
    'usr_admin_0000000001',
    'admin',
    '$2a$12$LJ3m4ys3Kk0mMfFq1B8qHeEqHqFm3jNrSPRQukPMtRzPVQqK9aGhu',
    'admin@vault.local',
    'admin',
    4
);
