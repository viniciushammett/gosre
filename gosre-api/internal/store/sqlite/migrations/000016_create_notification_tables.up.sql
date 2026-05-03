CREATE TABLE IF NOT EXISTS notification_channels (
    id         TEXT NOT NULL PRIMARY KEY,
    project_id TEXT NOT NULL,
    name       TEXT NOT NULL,
    kind       TEXT NOT NULL,
    config     TEXT NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_notif_channels_project ON notification_channels (project_id);

CREATE TABLE IF NOT EXISTS notification_rules (
    id         TEXT NOT NULL PRIMARY KEY,
    project_id TEXT NOT NULL,
    channel_id TEXT NOT NULL,
    event_kind TEXT NOT NULL,
    tag_filter TEXT NOT NULL DEFAULT '[]'
);

CREATE INDEX IF NOT EXISTS idx_notif_rules_project ON notification_rules (project_id);
