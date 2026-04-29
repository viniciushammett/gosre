IF EXISTS (SELECT 1 FROM sys.columns WHERE object_id = OBJECT_ID('targets') AND name = 'service_id')
BEGIN
    ALTER TABLE targets DROP COLUMN service_id;
END;
