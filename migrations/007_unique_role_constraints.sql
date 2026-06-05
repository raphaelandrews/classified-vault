CREATE UNIQUE INDEX IF NOT EXISTS idx_users_unique_mayor ON users(role) WHERE role = 'Mayor';
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_unique_director ON users(department, role) WHERE role = 'Director';
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_unique_lead ON users(department, role) WHERE role NOT IN ('Mayor', 'Director', 'Member', 'Visitor');
