CREATE TABLE IF NOT EXISTS services (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    owner       TEXT NOT NULL,
    criticality TEXT NOT NULL DEFAULT 'medium',
    runbook_url TEXT NOT NULL DEFAULT '',
    repo_url    TEXT NOT NULL DEFAULT '',
    project_id  TEXT NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL
);
