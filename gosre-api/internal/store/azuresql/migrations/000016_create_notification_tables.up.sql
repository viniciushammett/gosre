IF OBJECT_ID('notification_channels', 'U') IS NULL
BEGIN
    CREATE TABLE notification_channels (
        id         NVARCHAR(36)  NOT NULL PRIMARY KEY,
        project_id NVARCHAR(36)  NOT NULL,
        name       NVARCHAR(255) NOT NULL,
        kind       NVARCHAR(50)  NOT NULL,
        config     NVARCHAR(MAX) NOT NULL DEFAULT '{}'
    );
    CREATE INDEX idx_notif_channels_project ON notification_channels (project_id);
END;

IF OBJECT_ID('notification_rules', 'U') IS NULL
BEGIN
    CREATE TABLE notification_rules (
        id         NVARCHAR(36)  NOT NULL PRIMARY KEY,
        project_id NVARCHAR(36)  NOT NULL,
        channel_id NVARCHAR(36)  NOT NULL,
        event_kind NVARCHAR(255) NOT NULL,
        tag_filter NVARCHAR(MAX) NOT NULL DEFAULT '[]'
    );
    CREATE INDEX idx_notif_rules_project ON notification_rules (project_id);
END;
