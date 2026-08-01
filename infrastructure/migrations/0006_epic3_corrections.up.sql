-- Epic 3: Pengajuan koreksi data warga.

CREATE TABLE resident_corrections (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    resident_id UUID NOT NULL,
    requester_user_id UUID NOT NULL,
    requested_changes JSONB NOT NULL CHECK (jsonb_typeof(requested_changes) = 'object' AND requested_changes <> '{}'::jsonb),
    reason TEXT NOT NULL CHECK (btrim(reason) <> ''),
    status VARCHAR(20) NOT NULL CHECK (status IN ('submitted', 'approved', 'rejected', 'needs_revision')),
    reviewer_user_id UUID,
    reviewer_note TEXT,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_resident_corrections_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT fk_resident_corrections_resident_tenant
        FOREIGN KEY (organization_id, resident_id)
        REFERENCES residents (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_resident_corrections_requester_tenant
        FOREIGN KEY (organization_id, requester_user_id)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_resident_corrections_reviewer_tenant
        FOREIGN KEY (organization_id, reviewer_user_id)
        REFERENCES users (organization_id, id)
        ON DELETE SET NULL
);

CREATE INDEX idx_resident_corrections_org_status
    ON resident_corrections (organization_id, status, created_at DESC);

CREATE TRIGGER trg_resident_corrections_updated_at
    BEFORE UPDATE ON resident_corrections
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();