IF NOT EXISTS (SELECT 1 FROM sys.tables WHERE name = 'projects')
BEGIN
    CREATE TABLE projects (
        id              NVARCHAR(255) NOT NULL,
        organization_id NVARCHAR(255) NOT NULL,
        team_id         NVARCHAR(255) NOT NULL DEFAULT '',
        name            NVARCHAR(255) NOT NULL,
        slug            NVARCHAR(255) NOT NULL,
        created_at      DATETIME2     NOT NULL,
        CONSTRAINT PK_projects PRIMARY KEY (id),
        CONSTRAINT UQ_projects_org_slug UNIQUE (organization_id, slug)
    );
    CREATE INDEX IX_projects_organization_id ON projects (organization_id);
    CREATE INDEX IX_projects_team_id ON projects (team_id);
END;
