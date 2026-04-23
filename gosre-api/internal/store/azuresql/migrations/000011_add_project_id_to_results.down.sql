IF EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'IX_results_project_id')
    DROP INDEX IX_results_project_id ON results;
