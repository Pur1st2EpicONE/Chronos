CREATE TABLE IF NOT EXISTS Notifications (
    id          INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    uuid        VARCHAR(36) NOT NULL UNIQUE,
    channel     VARCHAR(20) NOT NULL,
    message     TEXT NOT NULL,
    status      VARCHAR(30) NOT NULL DEFAULT 'pending',
    send_at     TIMESTAMP WITH TIME ZONE NOT NULL,
    send_to     VARCHAR(254) NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_notifications_uuid ON Notifications(uuid);