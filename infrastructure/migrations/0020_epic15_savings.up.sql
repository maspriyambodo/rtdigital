-- Epic 15: Tabungan Warga (Dana Titipan Non-Kas)

CREATE TABLE savings_products (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    code VARCHAR(50) NOT NULL CHECK (code <> ''),
    name VARCHAR(100) NOT NULL CHECK (name <> ''),
    description TEXT,
    minimum_deposit NUMERIC(15,2) NOT NULL DEFAULT 0 CHECK (minimum_deposit >= 0),
    withdrawal_rule VARCHAR(50) NOT NULL, -- e.g. 'anytime', 'end_of_period', 'on_demand'
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_savings_products_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT uq_savings_products_organization_code UNIQUE (organization_id, code),
    CONSTRAINT fk_savings_products_created_by_tenant
        FOREIGN KEY (organization_id, created_by)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_savings_products_organization_status ON savings_products (organization_id, status);

CREATE TRIGGER trg_savings_products_updated_at
    BEFORE UPDATE ON savings_products
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE savings_accounts (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    product_id UUID NOT NULL,
    household_id UUID NOT NULL,
    account_number VARCHAR(50) NOT NULL CHECK (account_number <> ''),
    balance NUMERIC(15,2) NOT NULL DEFAULT 0 CHECK (balance >= 0),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'closed')),
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_savings_accounts_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT uq_savings_accounts_organization_account_number UNIQUE (organization_id, account_number),
    CONSTRAINT uq_savings_accounts_organization_product_household UNIQUE (organization_id, product_id, household_id) WHERE status = 'active',
    CONSTRAINT fk_savings_accounts_product_tenant
        FOREIGN KEY (organization_id, product_id)
        REFERENCES savings_products (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_savings_accounts_household_tenant
        FOREIGN KEY (organization_id, household_id)
        REFERENCES households (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_savings_accounts_created_by_tenant
        FOREIGN KEY (organization_id, created_by)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_savings_accounts_organization_product ON savings_accounts (organization_id, product_id);
CREATE INDEX idx_savings_accounts_organization_household ON savings_accounts (organization_id, household_id);


CREATE TABLE savings_transactions (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    account_id UUID NOT NULL,
    transaction_number VARCHAR(50) NOT NULL CHECK (transaction_number <> ''),
    type VARCHAR(20) NOT NULL CHECK (type IN ('deposit', 'withdrawal', 'reversal')),
    amount NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    balance_after NUMERIC(15,2) NOT NULL CHECK (balance_after >= 0),
    transaction_date DATE NOT NULL,
    description TEXT NOT NULL CHECK (description <> ''),
    proof_file_id UUID,
    reversal_of_id UUID,
    verification_status VARCHAR(20) NOT NULL DEFAULT 'pending' 
        CHECK (verification_status IN ('pending', 'verified', 'rejected')),
    verified_by UUID,
    verified_at TIMESTAMPTZ,
    rejection_reason TEXT,
    idempotency_key VARCHAR(255),
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_savings_transactions_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT uq_savings_transactions_organization_number UNIQUE (organization_id, transaction_number),
    CONSTRAINT fk_savings_transactions_account_tenant
        FOREIGN KEY (organization_id, account_id)
        REFERENCES savings_accounts (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_savings_transactions_proof_file_tenant
        FOREIGN KEY (organization_id, proof_file_id)
        REFERENCES file_objects (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_savings_transactions_reversal_tenant
        FOREIGN KEY (organization_id, reversal_of_id)
        REFERENCES savings_transactions (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_savings_transactions_verified_by_tenant
        FOREIGN KEY (organization_id, verified_by)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_savings_transactions_created_by_tenant
        FOREIGN KEY (organization_id, created_by)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT chk_savings_transactions_reversal_not_self
        CHECK (reversal_of_id IS NULL OR reversal_of_id <> id),
    CONSTRAINT chk_savings_transactions_verified
        CHECK (
            verification_status <> 'verified'
            OR (
                verified_by IS NOT NULL
                AND verified_at IS NOT NULL
                AND verified_by <> created_by
            )
        ),
    CONSTRAINT chk_savings_transactions_rejected
        CHECK (
            verification_status <> 'rejected'
            OR (
                verified_by IS NOT NULL
                AND verified_at IS NOT NULL
                AND verified_by <> created_by
                AND NULLIF(trim(rejection_reason), '') IS NOT NULL
            )
        )
);

CREATE UNIQUE INDEX uq_savings_transactions_organization_idempotency_key
    ON savings_transactions (organization_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE UNIQUE INDEX uq_savings_transactions_reversal_of
    ON savings_transactions (organization_id, reversal_of_id)
    WHERE reversal_of_id IS NOT NULL;

CREATE INDEX idx_savings_transactions_organization_account_date
    ON savings_transactions (organization_id, account_id, transaction_date DESC, created_at DESC);

CREATE INDEX idx_savings_transactions_organization_status
    ON savings_transactions (organization_id, verification_status, created_at DESC);

CREATE TRIGGER trg_savings_transactions_updated_at
    BEFORE UPDATE ON savings_transactions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE file_attachments
    DROP CONSTRAINT file_attachments_entity_type_check,
    ADD CONSTRAINT file_attachments_entity_type_check
        CHECK (entity_type IN ('announcement', 'complaint', 'letter_request', 'payment', 'cash_transaction', 'savings_transaction'));

CREATE OR REPLACE FUNCTION prevent_savings_transactions_delete()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'savings transactions are append-only and cannot be deleted';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_prevent_savings_transactions_delete
    BEFORE DELETE ON savings_transactions
    FOR EACH ROW EXECUTE FUNCTION prevent_savings_transactions_delete();

CREATE TRIGGER trg_savings_accounts_updated_at
    BEFORE UPDATE ON savings_accounts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();