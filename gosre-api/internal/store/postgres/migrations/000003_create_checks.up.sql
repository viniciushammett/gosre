CREATE TABLE IF NOT EXISTS checks (
    id        TEXT PRIMARY KEY,
    type      TEXT NOT NULL,
    target_id TEXT NOT NULL,
    interval  BIGINT NOT NULL DEFAULT 0,
    timeout   BIGINT NOT NULL DEFAULT 0,
    params    TEXT NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_checks_target_id ON checks (target_id);
