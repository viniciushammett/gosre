IF NOT EXISTS (SELECT 1 FROM sys.columns WHERE object_id = OBJECT_ID('results') AND name = 'project_id')
BEGIN
    ALTER TABLE results ADD project_id NVARCHAR(255) NOT NULL DEFAULT '';
    CREATE INDEX IX_results_project_id ON results (project_id);
END;
