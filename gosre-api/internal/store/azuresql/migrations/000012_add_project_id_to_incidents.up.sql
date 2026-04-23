IF NOT EXISTS (SELECT 1 FROM sys.columns WHERE object_id = OBJECT_ID('incidents') AND name = 'project_id')
BEGIN
    ALTER TABLE incidents ADD project_id NVARCHAR(255) NOT NULL DEFAULT '';
    CREATE INDEX IX_incidents_project_id ON incidents (project_id);
END;
