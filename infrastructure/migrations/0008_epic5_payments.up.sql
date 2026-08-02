-- Epic 5: Pembayaran dan lampiran privat.

CREATE TABLE file_objects (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    storage_key TEXT NOT NULL UNIQUE CHECK (storage_key <> ''),
    original_name VARCHAR(255) NOT NULL CHECK (original_name <> ''),
    mime_type VARCHAR(100) NOT NULL CHECK (mime_type <> ''),
    size_bytes BIGINT NOT NULL CHECK (size_bytes > 0),
    checksum CHAR(64),
    visibility VARCHAR(20) NOT NULL DEFAULT 'private'
        CHECK (visibility = 'private'),
    uploaded_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_file_objects_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT fk_file_objects_uploaded_by_tenant
        FOREIGN KEY (organization_id, uploaded_by)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_file_objects_organization_uploaded_by
    ON file_objects (organization_id, uploaded_by, created_at DESC);

CREATE TABLE file_attachments (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    file_id UUID NOT NULL,
    entity_type VARCHAR(50) NOT NULL
        CHECK (entity_type IN ('announcement', 'complaint', 'letter_request', 'payment')),
    entity_id UUID NOT NULL,
    purpose VARCHAR(50) NOT NULL CHECK (purpose <> ''),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_file_attachments_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT fk_file_attachments_file_tenant
        FOREIGN KEY (organization_id, file_id)
        REFERENCES file_objects (organization_id, id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_file_attachments_organization_entity
    ON file_attachments (organization_id, entity_type, entity_id);

CREATE TABLE payments (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    invoice_id UUID NOT NULL,
    payment_number VARCHAR(50) NOT NULL CHECK (payment_number <> ''),
    method VARCHAR(20) NOT NULL CHECK (method IN ('cash', 'transfer', 'other')),
    amount NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    paid_at TIMESTAMPTZ NOT NULL,
    proof_file_id UUID,
    verification_status VARCHAR(20) NOT NULL
        CHECK (verification_status IN ('pending', 'verified', 'rejected', 'cancelled')),
    verified_by UUID,
    verified_at TIMESTAMPTZ,
    rejection_reason TEXT,
    cancelled_by UUID,
    cancelled_at TIMESTAMPTZ,
    cancellation_reason TEXT,
    idempotency_key VARCHAR(255),
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_payments_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT uq_payments_organization_number UNIQUE (organization_id, payment_number),
    CONSTRAINT fk_payments_invoice_tenant
        FOREIGN KEY (organization_id, invoice_id)
        REFERENCES invoices (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_payments_proof_file_tenant
        FOREIGN KEY (organization_id, proof_file_id)
        REFERENCES file_objects (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_payments_verified_by_tenant
        FOREIGN KEY (organization_id, verified_by)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_payments_cancelled_by_tenant
        FOREIGN KEY (organization_id, cancelled_by)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_payments_created_by_tenant
        FOREIGN KEY (organization_id, created_by)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT chk_payments_transfer_proof
        CHECK (method <> 'transfer' OR proof_file_id IS NOT NULL),
    CONSTRAINT chk_payments_verified
        CHECK (
            verification_status <> 'verified'
            OR (
                verified_by IS NOT NULL
                AND verified_at IS NOT NULL
                AND verified_by <> created_by
            )
        ),
    CONSTRAINT chk_payments_rejected
        CHECK (
            verification_status <> 'rejected'
            OR (
                verified_by IS NOT NULL
                AND verified_at IS NOT NULL
                AND verified_by <> created_by
                AND NULLIF(trim(rejection_reason), '') IS NOT NULL
            )
        ),
    CONSTRAINT chk_payments_cancelled
        CHECK (
            verification_status <> 'cancelled'
            OR (
                cancelled_by IS NOT NULL
                AND cancelled_at IS NOT NULL
                AND NULLIF(trim(cancellation_reason), '') IS NOT NULL
            )
        )
);

CREATE UNIQUE INDEX uq_payments_organization_idempotency_key
    ON payments (organization_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE INDEX idx_payments_organization_status_created_at
    ON payments (organization_id, verification_status, created_at DESC);

CREATE INDEX idx_payments_organization_invoice
    ON payments (organization_id, invoice_id);

CREATE TRIGGER trg_payments_updated_at
    BEFORE UPDATE ON payments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();