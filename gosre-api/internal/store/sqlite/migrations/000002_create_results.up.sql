CREATE TABLE IF NOT EXISTS results (
    id         TEXT PRIMARY KEY,
    check_id   TEXT NOT NULL,
    target_id  TEXT NOT NULL,
    agent_id   TEXT NOT NULL DEFAULT '',
    status     TEXT NOT NULL,
    duration   INTEGER NOT NULL DEFAULT 0,
    error      TEXT NOT NULL DEFAULT '',
    timestamp  DATETIME NOT NULL,
    metadata   TEXT NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_results_target_id ON results (target_id);
CREATE INDEX IF NOT EXISTS idx_results_status    ON results (status);
CREATE INDEX IF NOT EXISTS idx_results_timestamp ON results (timestamp);
