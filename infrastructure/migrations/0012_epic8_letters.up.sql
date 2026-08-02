-- Epic 8: Surat Pengantar

CREATE TABLE letter_types (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    name VARCHAR(100) NOT NULL CHECK (name <> ''),
    requirements JSONB NOT NULL DEFAULT '[]'::jsonb,
    form_schema JSONB NOT NULL DEFAULT '{}'::jsonb,
    template TEXT NOT NULL CHECK (template <> ''),
    number_pattern VARCHAR(100) NOT NULL CHECK (number_pattern <> ''),
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_letter_types_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT uq_letter_types_organization_name UNIQUE (organization_id, name)
);

CREATE INDEX idx_letter_types_organization_status
    ON letter_types (organization_id, status);

CREATE TRIGGER trg_letter_types_updated_at
    BEFORE UPDATE ON letter_types
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE letter_requests (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    requester_user_id UUID NOT NULL,
    resident_id UUID NOT NULL,
    letter_type_id UUID NOT NULL,
    request_number VARCHAR(50) NOT NULL CHECK (request_number <> ''),
    letter_number VARCHAR(100),
    form_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(30) NOT NULL DEFAULT 'draft'
        CHECK (status IN (
            'draft',
            'submitted',
            'under_review',
            'needs_revision',
            'awaiting_approval',
            'approved',
            'issued',
            'rejected',
            'cancelled'
        )),
    resident_note TEXT,
    internal_note TEXT,
    submitted_at TIMESTAMPTZ,
    processed_by UUID,
    approved_by UUID,
    approved_at TIMESTAMPTZ,
    issued_file_id UUID,
    issued_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_letter_requests_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT uq_letter_requests_organization_request_number
        UNIQUE (organization_id, request_number),
    CONSTRAINT fk_letter_requests_requester_tenant
        FOREIGN KEY (organization_id, requester_user_id)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_letter_requests_resident_tenant
        FOREIGN KEY (organization_id, resident_id)
        REFERENCES residents (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_letter_requests_type_tenant
        FOREIGN KEY (organization_id, letter_type_id)
        REFERENCES letter_types (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_letter_requests_processed_tenant
        FOREIGN KEY (organization_id, processed_by)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_letter_requests_approved_tenant
        FOREIGN KEY (organization_id, approved_by)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_letter_requests_file_tenant
        FOREIGN KEY (organization_id, issued_file_id)
        REFERENCES file_objects (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT chk_letter_requests_issued_data
        CHECK (
            status <> 'issued'
            OR (
                letter_number IS NOT NULL
                AND issued_file_id IS NOT NULL
                AND issued_at IS NOT NULL
            )
        )
);

CREATE UNIQUE INDEX uq_letter_requests_organization_letter_number
    ON letter_requests (organization_id, letter_number)
    WHERE letter_number IS NOT NULL;

CREATE INDEX idx_letter_requests_org_status_created_at
    ON letter_requests (organization_id, status, created_at DESC);

CREATE INDEX idx_letter_requests_org_requester_created_at
    ON letter_requests (organization_id, requester_user_id, created_at DESC);

CREATE INDEX idx_letter_requests_org_resident_created_at
    ON letter_requests (organization_id, resident_id, created_at DESC);

CREATE TRIGGER trg_letter_requests_updated_at
    BEFORE UPDATE ON letter_requests
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();