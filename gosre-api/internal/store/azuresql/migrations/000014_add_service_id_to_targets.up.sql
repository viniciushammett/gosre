IF NOT EXISTS (SELECT 1 FROM sys.columns WHERE object_id = OBJECT_ID('targets') AND name = 'service_id')
BEGIN
    ALTER TABLE targets ADD service_id NVARCHAR(255) NULL;
END;
