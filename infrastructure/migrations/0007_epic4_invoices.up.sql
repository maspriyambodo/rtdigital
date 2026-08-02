-- Epic 4: Iuran dan Tagihan.

CREATE TABLE due_types (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    amount NUMERIC(15,2) CHECK (amount IS NULL OR amount > 0),
    frequency VARCHAR(20) NOT NULL CHECK (frequency IN ('once', 'monthly', 'quarterly', 'yearly')),
    due_day SMALLINT CHECK (due_day IS NULL OR due_day BETWEEN 1 AND 31),
    status VARCHAR(20) NOT NULL CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_due_types_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT uq_due_types_organization_name UNIQUE (organization_id, name)
);

CREATE INDEX idx_due_types_organization_status
    ON due_types (organization_id, status);

CREATE TRIGGER trg_due_types_updated_at
    BEFORE UPDATE ON due_types
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE invoices (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    household_id UUID NOT NULL,
    due_type_id UUID NOT NULL,
    invoice_number VARCHAR(50) NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    due_date DATE NOT NULL,
    amount NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    paid_amount NUMERIC(15,2) NOT NULL DEFAULT 0 CHECK (paid_amount >= 0 AND paid_amount <= amount),
    adjustment_amount NUMERIC(15,2) NOT NULL DEFAULT 0,
    adjustment_reason TEXT,
    status VARCHAR(30) NOT NULL CHECK (status IN ('unpaid', 'pending_verification', 'partial', 'paid', 'cancelled')),
    cancelled_by UUID,
    cancelled_at TIMESTAMPTZ,
    cancellation_reason TEXT,
    bulk_generation_key VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_invoices_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT uq_invoices_organization_number UNIQUE (organization_id, invoice_number),
    CONSTRAINT chk_invoices_period CHECK (period_end >= period_start),
    CONSTRAINT chk_invoices_adjustment_reason CHECK (
        adjustment_amount = 0 OR NULLIF(trim(adjustment_reason), '') IS NOT NULL
    ),
    CONSTRAINT chk_invoices_cancellation CHECK (
        status <> 'cancelled' OR (
            cancelled_by IS NOT NULL
            AND cancelled_at IS NOT NULL
            AND NULLIF(trim(cancellation_reason), '') IS NOT NULL
        )
    ),
    CONSTRAINT fk_invoices_household_tenant
        FOREIGN KEY (organization_id, household_id)
        REFERENCES households (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_invoices_due_type_tenant
        FOREIGN KEY (organization_id, due_type_id)
        REFERENCES due_types (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_invoices_cancelled_by_tenant
        FOREIGN KEY (organization_id, cancelled_by)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_invoices_organization_status_due_date
    ON invoices (organization_id, status, due_date);

CREATE INDEX idx_invoices_organization_household_period
    ON invoices (organization_id, household_id, period_start, period_end);

CREATE INDEX idx_invoices_bulk_generation_key
    ON invoices (organization_id, bulk_generation_key)
    WHERE bulk_generation_key IS NOT NULL;

CREATE TRIGGER trg_invoices_updated_at
    BEFORE UPDATE ON invoices
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();