ALTER TABLE checks ADD COLUMN project_id TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_checks_project_id ON checks (project_id);
