-- Epic 6: Buku kas append-only.

CREATE TABLE cash_categories (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    name VARCHAR(100) NOT NULL CHECK (name <> ''),
    type VARCHAR(10) NOT NULL CHECK (type IN ('income', 'expense')),
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_cash_categories_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT uq_cash_categories_organization_type_name
        UNIQUE (organization_id, type, name)
);

CREATE INDEX idx_cash_categories_organization_type_status
    ON cash_categories (organization_id, type, status);

CREATE TRIGGER trg_cash_categories_updated_at
    BEFORE UPDATE ON cash_categories
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE cash_transactions (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    transaction_number VARCHAR(50) NOT NULL CHECK (transaction_number <> ''),
    type VARCHAR(10) NOT NULL CHECK (type IN ('income', 'expense')),
    category_id UUID,
    amount NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    transaction_date DATE NOT NULL,
    description TEXT NOT NULL CHECK (description <> ''),
    proof_file_id UUID,
    reference_type VARCHAR(50),
    reference_id UUID,
    reversal_of_id UUID,
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'reversed')),
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_cash_transactions_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT uq_cash_transactions_organization_number
        UNIQUE (organization_id, transaction_number),
    CONSTRAINT fk_cash_transactions_category_tenant
        FOREIGN KEY (organization_id, category_id)
        REFERENCES cash_categories (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_cash_transactions_proof_file_tenant
        FOREIGN KEY (organization_id, proof_file_id)
        REFERENCES file_objects (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_cash_transactions_reversal_tenant
        FOREIGN KEY (organization_id, reversal_of_id)
        REFERENCES cash_transactions (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_cash_transactions_created_by_tenant
        FOREIGN KEY (organization_id, created_by)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT chk_cash_transactions_reversal_not_self
        CHECK (reversal_of_id IS NULL OR reversal_of_id <> id)
);

CREATE INDEX idx_cash_transactions_organization_date
    ON cash_transactions (organization_id, transaction_date DESC, created_at DESC);

CREATE INDEX idx_cash_transactions_organization_category
    ON cash_transactions (organization_id, category_id);

CREATE INDEX idx_cash_transactions_organization_reference
    ON cash_transactions (organization_id, reference_type, reference_id);

CREATE UNIQUE INDEX uq_cash_transactions_active_payment_reference
    ON cash_transactions (organization_id, reference_id)
    WHERE reference_type = 'payment' AND status = 'active';

CREATE UNIQUE INDEX uq_cash_transactions_reversal_of
    ON cash_transactions (organization_id, reversal_of_id)
    WHERE reversal_of_id IS NOT NULL;

CREATE TRIGGER trg_cash_transactions_updated_at
    BEFORE UPDATE ON cash_transactions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE file_attachments
    DROP CONSTRAINT file_attachments_entity_type_check,
    ADD CONSTRAINT file_attachments_entity_type_check
        CHECK (entity_type IN ('announcement', 'complaint', 'letter_request', 'payment', 'cash_transaction'));

CREATE OR REPLACE FUNCTION prevent_cash_transactions_delete()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'cash transactions are append-only and cannot be deleted';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_prevent_cash_transactions_delete
    BEFORE DELETE ON cash_transactions
    FOR EACH ROW EXECUTE FUNCTION prevent_cash_transactions_delete();