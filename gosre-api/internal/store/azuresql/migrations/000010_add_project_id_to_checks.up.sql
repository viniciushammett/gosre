IF NOT EXISTS (SELECT 1 FROM sys.columns WHERE object_id = OBJECT_ID('checks') AND name = 'project_id')
BEGIN
    ALTER TABLE checks ADD project_id NVARCHAR(255) NOT NULL DEFAULT '';
    CREATE INDEX IX_checks_project_id ON checks (project_id);
END;
