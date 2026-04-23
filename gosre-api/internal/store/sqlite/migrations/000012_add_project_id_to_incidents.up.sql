ALTER TABLE incidents ADD COLUMN project_id TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_incidents_project_id ON incidents (project_id);
