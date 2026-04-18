CREATE TABLE IF NOT EXISTS incidents (
    id         TEXT PRIMARY KEY,
    target_id  TEXT NOT NULL,
    state      TEXT NOT NULL DEFAULT 'open',
    first_seen TIMESTAMPTZ NOT NULL,
    last_seen  TIMESTAMPTZ NOT NULL,
    result_ids TEXT NOT NULL DEFAULT '[]'
);

CREATE INDEX IF NOT EXISTS idx_incidents_target_id ON incidents (target_id);
CREATE INDEX IF NOT EXISTS idx_incidents_state     ON incidents (state);
