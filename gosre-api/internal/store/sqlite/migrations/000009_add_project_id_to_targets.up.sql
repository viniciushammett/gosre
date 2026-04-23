ALTER TABLE targets ADD COLUMN project_id TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_targets_project_id ON targets (project_id);
