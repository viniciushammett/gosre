IF NOT EXISTS (SELECT 1 FROM sys.tables WHERE name = 'dependencies')
BEGIN
    CREATE TABLE dependencies (
        id                NVARCHAR(255) NOT NULL,
        source_service_id NVARCHAR(255) NOT NULL,
        target_service_id NVARCHAR(255) NOT NULL,
        kind              NVARCHAR(50)  NOT NULL DEFAULT 'generic',
        created_at        DATETIME2     NOT NULL,
        CONSTRAINT PK_dependencies PRIMARY KEY (id)
    );
END;
