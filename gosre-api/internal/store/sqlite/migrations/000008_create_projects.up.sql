CREATE TABLE IF NOT EXISTS projects (
    id              TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL,
    team_id         TEXT NOT NULL DEFAULT '',
    name            TEXT NOT NULL,
    slug            TEXT NOT NULL,
    created_at      DATETIME NOT NULL,
    UNIQUE (organization_id, slug)
);
CREATE INDEX IF NOT EXISTS idx_projects_organization_id ON projects (organization_id);
CREATE INDEX IF NOT EXISTS idx_projects_team_id ON projects (team_id);
