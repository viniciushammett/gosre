IF OBJECT_ID('slos', 'U') IS NULL
CREATE TABLE slos (
    id             NVARCHAR(255) NOT NULL,
    target_id      NVARCHAR(255) NOT NULL,
    name           NVARCHAR(255) NOT NULL,
    metric         NVARCHAR(255) NOT NULL,
    threshold      FLOAT         NOT NULL,
    window_seconds BIGINT        NOT NULL,
    CONSTRAINT pk_slos PRIMARY KEY (id)
);

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_slos_target_id' AND object_id = OBJECT_ID('slos'))
CREATE INDEX idx_slos_target_id ON slos (target_id);
