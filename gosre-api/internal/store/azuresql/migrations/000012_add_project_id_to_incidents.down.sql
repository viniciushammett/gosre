IF EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'IX_incidents_project_id')
    DROP INDEX IX_incidents_project_id ON incidents;
