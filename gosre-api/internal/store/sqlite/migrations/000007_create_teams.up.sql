CREATE TABLE IF NOT EXISTS teams (
    id              TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL,
    name            TEXT NOT NULL,
    slug            TEXT NOT NULL,
    created_at      DATETIME NOT NULL,
    UNIQUE (organization_id, slug)
);
CREATE INDEX IF NOT EXISTS idx_teams_organization_id ON teams (organization_id);
