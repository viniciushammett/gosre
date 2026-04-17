CREATE TABLE IF NOT EXISTS targets (
    id        TEXT PRIMARY KEY,
    name      TEXT NOT NULL,
    type      TEXT NOT NULL,
    address   TEXT NOT NULL,
    tags      TEXT NOT NULL DEFAULT '[]',
    metadata  TEXT NOT NULL DEFAULT '{}'
);
