IF NOT EXISTS (SELECT 1 FROM sys.tables WHERE name = 'services')
BEGIN
    CREATE TABLE services (
        id          NVARCHAR(255)  NOT NULL,
        name        NVARCHAR(255)  NOT NULL,
        owner       NVARCHAR(255)  NOT NULL,
        criticality NVARCHAR(50)   NOT NULL DEFAULT 'medium',
        runbook_url NVARCHAR(2048) NOT NULL DEFAULT '',
        repo_url    NVARCHAR(2048) NOT NULL DEFAULT '',
        project_id  NVARCHAR(255)  NOT NULL DEFAULT '',
        created_at  DATETIME2      NOT NULL,
        CONSTRAINT PK_services PRIMARY KEY (id)
    );
END;
