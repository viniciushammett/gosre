IF NOT EXISTS (SELECT 1 FROM sys.columns WHERE object_id = OBJECT_ID('results') AND name = 'target_name')
BEGIN
    ALTER TABLE results ADD target_name NVARCHAR(255) NOT NULL CONSTRAINT df_results_target_name DEFAULT '';
END;
