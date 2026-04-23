ALTER TABLE results ADD COLUMN project_id TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_results_project_id ON results (project_id);
