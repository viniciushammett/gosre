IF OBJECT_ID('incidents', 'U') IS NULL
CREATE TABLE incidents (
    id         NVARCHAR(255) NOT NULL,
    target_id  NVARCHAR(255) NOT NULL,
    state      NVARCHAR(50)  NOT NULL CONSTRAINT df_incidents_state DEFAULT 'open',
    first_seen DATETIME2     NOT NULL,
    last_seen  DATETIME2     NOT NULL,
    result_ids NVARCHAR(MAX) NOT NULL CONSTRAINT df_incidents_result_ids DEFAULT '[]',
    CONSTRAINT pk_incidents PRIMARY KEY (id)
);

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_incidents_target_id' AND object_id = OBJECT_ID('incidents'))
CREATE INDEX idx_incidents_target_id ON incidents (target_id);

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_incidents_state' AND object_id = OBJECT_ID('incidents'))
CREATE INDEX idx_incidents_state ON incidents (state);
