-- Audit log tersedia sejak Epic 1 untuk tindakan autentikasi.
-- Append-only: aplikasi tidak diberi operasi UPDATE/DELETE.

CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL CHECK (action <> ''),
    entity_type VARCHAR(100) NOT NULL CHECK (entity_type <> ''),
    entity_id UUID,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_logs_organization_time
    ON audit_logs (organization_id, created_at DESC)
    WHERE organization_id IS NOT NULL;

CREATE INDEX idx_audit_logs_actor_time
    ON audit_logs (actor_user_id, created_at DESC)
    WHERE actor_user_id IS NOT NULL;

CREATE INDEX idx_audit_logs_entity
    ON audit_logs (entity_type, entity_id)
    WHERE entity_id IS NOT NULL;

CREATE INDEX idx_audit_logs_action_time
    ON audit_logs (action, created_at DESC);