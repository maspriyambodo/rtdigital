-- Epic 10: Notifikasi

CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    user_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type <> ''),
    title VARCHAR(255) NOT NULL CHECK (title <> ''),
    body TEXT,
    reference_type VARCHAR(50),
    reference_id UUID,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_notifications_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT fk_notifications_user_tenant
        FOREIGN KEY (organization_id, user_id)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_notifications_user_unread
    ON notifications (user_id, created_at DESC)
    WHERE read_at IS NULL;

CREATE INDEX idx_notifications_user_created_at
    ON notifications (user_id, created_at DESC);