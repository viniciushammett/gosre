IF EXISTS (SELECT 1 FROM sys.columns WHERE object_id = OBJECT_ID('results') AND name = 'target_name')
    ALTER TABLE results DROP COLUMN target_name;
