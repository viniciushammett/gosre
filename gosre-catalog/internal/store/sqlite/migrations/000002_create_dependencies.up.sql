CREATE TABLE IF NOT EXISTS dependencies (
    id                TEXT PRIMARY KEY,
    source_service_id TEXT NOT NULL,
    target_service_id TEXT NOT NULL,
    kind              TEXT NOT NULL DEFAULT 'generic',
    created_at        DATETIME NOT NULL
);
