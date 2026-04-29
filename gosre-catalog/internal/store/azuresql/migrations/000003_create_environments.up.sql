IF NOT EXISTS (SELECT 1 FROM sys.tables WHERE name = 'environments')
BEGIN
    CREATE TABLE environments (
        id         NVARCHAR(255) NOT NULL,
        name       NVARCHAR(255) NOT NULL,
        project_id NVARCHAR(255) NOT NULL,
        kind       NVARCHAR(50)  NOT NULL DEFAULT 'dev',
        created_at DATETIME2     NOT NULL,
        CONSTRAINT PK_environments PRIMARY KEY (id)
    );
END;
