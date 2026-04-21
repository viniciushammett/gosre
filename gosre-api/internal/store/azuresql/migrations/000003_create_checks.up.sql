IF OBJECT_ID('checks', 'U') IS NULL
CREATE TABLE checks (
    id          NVARCHAR(255) NOT NULL,
    type        NVARCHAR(50)  NOT NULL,
    target_id   NVARCHAR(255) NOT NULL,
    interval_ns BIGINT        NOT NULL CONSTRAINT df_checks_interval_ns DEFAULT 0,
    timeout_ns  BIGINT        NOT NULL CONSTRAINT df_checks_timeout_ns  DEFAULT 0,
    params      NVARCHAR(MAX) NOT NULL CONSTRAINT df_checks_params      DEFAULT '{}',
    CONSTRAINT pk_checks PRIMARY KEY (id)
);

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_checks_target_id' AND object_id = OBJECT_ID('checks'))
CREATE INDEX idx_checks_target_id ON checks (target_id);
