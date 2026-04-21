IF NOT EXISTS (SELECT 1 FROM sys.tables WHERE name = 'teams')
BEGIN
    CREATE TABLE teams (
        id              NVARCHAR(255) NOT NULL,
        organization_id NVARCHAR(255) NOT NULL,
        name            NVARCHAR(255) NOT NULL,
        slug            NVARCHAR(255) NOT NULL,
        created_at      DATETIME2     NOT NULL,
        CONSTRAINT PK_teams PRIMARY KEY (id),
        CONSTRAINT UQ_teams_org_slug UNIQUE (organization_id, slug)
    );
    CREATE INDEX IX_teams_organization_id ON teams (organization_id);
END;
