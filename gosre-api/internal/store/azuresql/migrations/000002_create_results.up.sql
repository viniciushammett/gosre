IF OBJECT_ID('results', 'U') IS NULL
CREATE TABLE results (
    id          NVARCHAR(255) NOT NULL,
    check_id    NVARCHAR(255) NOT NULL,
    target_id   NVARCHAR(255) NOT NULL,
    agent_id    NVARCHAR(255) NOT NULL CONSTRAINT df_results_agent_id    DEFAULT '',
    status      NVARCHAR(50)  NOT NULL,
    duration_ns BIGINT        NOT NULL CONSTRAINT df_results_duration_ns DEFAULT 0,
    error       NVARCHAR(MAX) NOT NULL CONSTRAINT df_results_error       DEFAULT '',
    timestamp   DATETIME2     NOT NULL,
    metadata    NVARCHAR(MAX) NOT NULL CONSTRAINT df_results_metadata    DEFAULT '{}',
    CONSTRAINT pk_results PRIMARY KEY (id)
);

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_results_target_id' AND object_id = OBJECT_ID('results'))
CREATE INDEX idx_results_target_id ON results (target_id);

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_results_status' AND object_id = OBJECT_ID('results'))
CREATE INDEX idx_results_status ON results (status);

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_results_timestamp' AND object_id = OBJECT_ID('results'))
CREATE INDEX idx_results_timestamp ON results (timestamp);
