IF EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'IX_targets_project_id')
    DROP INDEX IX_targets_project_id ON targets;
