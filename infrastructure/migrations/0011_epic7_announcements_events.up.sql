-- Epic 7: Pengumuman, target audiens, statistik baca, dan agenda.

CREATE TABLE announcements (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    author_user_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL CHECK (title <> ''),
    content TEXT NOT NULL CHECK (content <> ''),
    category VARCHAR(50) NOT NULL
        CHECK (category IN ('general', 'security', 'health', 'billing', 'event', 'emergency')),
    priority VARCHAR(20) NOT NULL DEFAULT 'normal'
        CHECK (priority IN ('normal', 'important')),
    publish_at TIMESTAMPTZ,
    expire_at TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'scheduled', 'published', 'archived')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_announcements_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT fk_announcements_author_tenant
        FOREIGN KEY (organization_id, author_user_id)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT chk_announcements_publish_expire
        CHECK (expire_at IS NULL OR publish_at IS NULL OR expire_at > publish_at),
    CONSTRAINT chk_announcements_scheduled_publish_at
        CHECK (status <> 'scheduled' OR publish_at IS NOT NULL)
);

CREATE INDEX idx_announcements_organization_status_publish
    ON announcements (organization_id, status, publish_at DESC);

CREATE TRIGGER trg_announcements_updated_at
    BEFORE UPDATE ON announcements
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE announcement_targets (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    announcement_id UUID NOT NULL,
    target_type VARCHAR(20) NOT NULL
        CHECK (target_type IN ('all', 'role', 'household', 'house_unit')),
    target_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_announcement_targets_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT fk_announcement_targets_announcement_tenant
        FOREIGN KEY (organization_id, announcement_id)
        REFERENCES announcements (organization_id, id)
        ON DELETE CASCADE,
    CONSTRAINT chk_announcement_targets_target
        CHECK (
            (target_type = 'all' AND target_id IS NULL)
            OR (target_type <> 'all' AND target_id IS NOT NULL)
        ),
    CONSTRAINT uq_announcement_targets_definition
        UNIQUE NULLS NOT DISTINCT (organization_id, announcement_id, target_type, target_id)
);

CREATE INDEX idx_announcement_targets_organization_target
    ON announcement_targets (organization_id, target_type, target_id);

CREATE TABLE announcement_read_receipts (
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    announcement_id UUID NOT NULL,
    user_id UUID NOT NULL,
    read_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (organization_id, announcement_id, user_id),
    CONSTRAINT fk_announcement_read_receipts_announcement_tenant
        FOREIGN KEY (organization_id, announcement_id)
        REFERENCES announcements (organization_id, id)
        ON DELETE CASCADE,
    CONSTRAINT fk_announcement_read_receipts_user_tenant
        FOREIGN KEY (organization_id, user_id)
        REFERENCES users (organization_id, id)
        ON DELETE CASCADE
);

CREATE INDEX idx_announcement_read_receipts_announcement
    ON announcement_read_receipts (organization_id, announcement_id, read_at DESC);

CREATE TABLE events (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    author_user_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL CHECK (title <> ''),
    description TEXT,
    location TEXT,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL DEFAULT 'planned'
        CHECK (status IN ('planned', 'ongoing', 'completed', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_events_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT fk_events_author_tenant
        FOREIGN KEY (organization_id, author_user_id)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT chk_events_starts_ends
        CHECK (ends_at IS NULL OR ends_at > starts_at)
);

CREATE INDEX idx_events_organization_status_starts
    ON events (organization_id, status, starts_at);

CREATE TRIGGER trg_events_updated_at
    BEFORE UPDATE ON events
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE file_attachments
    DROP CONSTRAINT file_attachments_entity_type_check,
    ADD CONSTRAINT file_attachments_entity_type_check
        CHECK (entity_type IN ('announcement', 'event', 'complaint', 'letter_request', 'payment', 'cash_transaction'));