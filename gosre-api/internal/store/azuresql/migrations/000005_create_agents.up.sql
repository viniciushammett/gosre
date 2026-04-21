IF OBJECT_ID('agents', 'U') IS NULL
CREATE TABLE agents (
    id        NVARCHAR(255) NOT NULL,
    hostname  NVARCHAR(255) NOT NULL,
    version   NVARCHAR(100) NOT NULL,
    last_seen DATETIME2     NOT NULL CONSTRAINT df_agents_last_seen DEFAULT SYSUTCDATETIME(),
    CONSTRAINT pk_agents PRIMARY KEY (id)
);
