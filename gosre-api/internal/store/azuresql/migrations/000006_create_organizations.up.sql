IF NOT EXISTS (SELECT 1 FROM sys.tables WHERE name = 'organizations')
BEGIN
    CREATE TABLE organizations (
        id         NVARCHAR(255) NOT NULL,
        name       NVARCHAR(255) NOT NULL,
        slug       NVARCHAR(255) NOT NULL,
        created_at DATETIME2     NOT NULL,
        CONSTRAINT PK_organizations PRIMARY KEY (id),
        CONSTRAINT UQ_organizations_slug UNIQUE (slug)
    );
END;
