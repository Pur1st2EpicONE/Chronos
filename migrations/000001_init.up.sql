CREATE TABLE IF NOT EXISTS Notifications (
    id          INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    uuid        VARCHAR(36) NOT NULL UNIQUE,
    channel     VARCHAR(20) NOT NULL,
    message     TEXT NOT NULL,
    status      VARCHAR(30) NOT NULL DEFAULT 'pending',
    send_at     TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS Recipients (
    id                INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    notification_uuid VARCHAR(36) NOT NULL REFERENCES Notifications(uuid) ON DELETE CASCADE,
    recipient         VARCHAR(254) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_notifications_uuid ON Notifications(uuid);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON Notifications(status) WHERE status = 'canceled';