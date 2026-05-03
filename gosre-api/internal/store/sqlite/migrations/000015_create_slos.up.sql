CREATE TABLE IF NOT EXISTS slos (
    id             TEXT    PRIMARY KEY,
    target_id      TEXT    NOT NULL,
    name           TEXT    NOT NULL,
    metric         TEXT    NOT NULL,
    threshold      REAL    NOT NULL,
    window_seconds INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_slos_target_id ON slos (target_id);
