IF OBJECT_ID('targets', 'U') IS NULL
CREATE TABLE targets (
    id       NVARCHAR(255) NOT NULL,
    name     NVARCHAR(255) NOT NULL,
    type     NVARCHAR(50)  NOT NULL,
    address  NVARCHAR(500) NOT NULL,
    tags     NVARCHAR(MAX) NOT NULL CONSTRAINT df_targets_tags     DEFAULT '[]',
    metadata NVARCHAR(MAX) NOT NULL CONSTRAINT df_targets_metadata DEFAULT '{}',
    CONSTRAINT pk_targets PRIMARY KEY (id)
);
