IF NOT EXISTS (SELECT 1 FROM sys.columns WHERE object_id = OBJECT_ID('targets') AND name = 'project_id')
BEGIN
    ALTER TABLE targets ADD project_id NVARCHAR(255) NOT NULL DEFAULT '';
    CREATE INDEX IX_targets_project_id ON targets (project_id);
END;
