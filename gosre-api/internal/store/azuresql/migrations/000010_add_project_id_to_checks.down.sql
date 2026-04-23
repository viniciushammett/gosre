IF EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'IX_checks_project_id')
    DROP INDEX IX_checks_project_id ON checks;
