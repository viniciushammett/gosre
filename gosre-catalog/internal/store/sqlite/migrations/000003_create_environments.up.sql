CREATE TABLE IF NOT EXISTS environments (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    project_id TEXT NOT NULL,
    kind       TEXT NOT NULL DEFAULT 'dev',
    created_at DATETIME NOT NULL
);
