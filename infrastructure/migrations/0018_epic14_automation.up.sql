-- Epic 14: Otomatisasi dan Layanan Operasional Proaktif.
-- Semua konfigurasi dan riwayat dibatasi organization_id. Tidak ada mutasi
-- historis yang dihapus; alokasi pembayaran dan event memakai append-only.

-- 14.1: penerbitan tagihan rutin.
ALTER TABLE due_types
    ADD COLUMN automatic_generation_enabled BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN generation_lead_days SMALLINT NOT NULL DEFAULT 0
        CHECK (generation_lead_days BETWEEN 0 AND 366),
    ADD COLUMN reminder_lead_days SMALLINT NOT NULL DEFAULT 3
        CHECK (reminder_lead_days BETWEEN 0 AND 30);

CREATE TABLE invoice_generation_runs (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    due_type_id UUID NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    due_date DATE NOT NULL,
    run_key VARCHAR(255) NOT NULL CHECK (run_key <> ''),
    status VARCHAR(20) NOT NULL
        CHECK (status IN ('running', 'completed', 'failed')),
    total_targeted INTEGER NOT NULL DEFAULT 0 CHECK (total_targeted >= 0),
    total_created INTEGER NOT NULL DEFAULT 0 CHECK (total_created >= 0),
    total_skipped INTEGER NOT NULL DEFAULT 0 CHECK (total_skipped >= 0),
    error_message TEXT,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ,
    created_by UUID,
    CONSTRAINT uq_invoice_generation_runs_organization_due_type_period
        UNIQUE (organization_id, due_type_id, period_start, period_end),
    CONSTRAINT uq_invoice_generation_runs_organization_run_key
        UNIQUE (organization_id, run_key),
    CONSTRAINT chk_invoice_generation_runs_period
        CHECK (period_end >= period_start),
    CONSTRAINT chk_invoice_generation_runs_completion
        CHECK (
            (status = 'running' AND completed_at IS NULL)
            OR (status IN ('completed', 'failed') AND completed_at IS NOT NULL)
        ),
    CONSTRAINT fk_invoice_generation_runs_due_type_tenant
        FOREIGN KEY (organization_id, due_type_id)
        REFERENCES due_types (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_invoice_generation_runs_created_by_tenant
        FOREIGN KEY (organization_id, created_by)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_invoice_generation_runs_organization_status_started_at
    ON invoice_generation_runs (organization_id, status, started_at DESC);

-- 14.2: preferensi kanal dan log pengingat ber-idempotensi.
CREATE TABLE notification_preferences (
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    user_id UUID NOT NULL,
    in_app_enabled BOOLEAN NOT NULL DEFAULT true,
    email_enabled BOOLEAN NOT NULL DEFAULT true,
    whatsapp_enabled BOOLEAN NOT NULL DEFAULT false,
    due_reminder_enabled BOOLEAN NOT NULL DEFAULT true,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (organization_id, user_id),
    CONSTRAINT fk_notification_preferences_user_tenant
        FOREIGN KEY (organization_id, user_id)
        REFERENCES users (organization_id, id)
        ON DELETE CASCADE
);

CREATE TABLE invoice_reminder_deliveries (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    invoice_id UUID NOT NULL,
    user_id UUID NOT NULL,
    channel VARCHAR(20) NOT NULL CHECK (channel IN ('in_app', 'email', 'whatsapp')),
    reminder_kind VARCHAR(30) NOT NULL
        CHECK (reminder_kind IN ('before_due', 'due_today', 'overdue')),
    scheduled_for DATE NOT NULL,
    status VARCHAR(20) NOT NULL
        CHECK (status IN ('sent', 'skipped', 'failed')),
    failure_message TEXT,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_invoice_reminder_deliveries_idempotency
        UNIQUE (organization_id, invoice_id, user_id, channel, reminder_kind, scheduled_for),
    CONSTRAINT fk_invoice_reminder_deliveries_invoice_tenant
        FOREIGN KEY (organization_id, invoice_id)
        REFERENCES invoices (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_invoice_reminder_deliveries_user_tenant
        FOREIGN KEY (organization_id, user_id)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_invoice_reminder_deliveries_organization_created_at
    ON invoice_reminder_deliveries (organization_id, created_at DESC);

-- 14.3: satu pembayaran dapat dialokasikan ke beberapa invoice.
-- invoice_id legacy dipertahankan sebagai referensi utama kompatibilitas.
ALTER TABLE payments
    ALTER COLUMN invoice_id DROP NOT NULL;

CREATE TABLE payment_allocations (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    payment_id UUID NOT NULL,
    invoice_id UUID NOT NULL,
    amount NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_payment_allocations_payment_invoice
        UNIQUE (organization_id, payment_id, invoice_id),
    CONSTRAINT fk_payment_allocations_payment_tenant
        FOREIGN KEY (organization_id, payment_id)
        REFERENCES payments (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_payment_allocations_invoice_tenant
        FOREIGN KEY (organization_id, invoice_id)
        REFERENCES invoices (organization_id, id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_payment_allocations_organization_invoice
    ON payment_allocations (organization_id, invoice_id);

-- 14.5, 14.6, 14.12: validasi pra-pengajuan, SLA, verifikasi publik surat.
ALTER TABLE letter_types
    ADD COLUMN sla_hours INTEGER
        CHECK (sla_hours IS NULL OR sla_hours > 0);

ALTER TABLE letter_requests
    ADD COLUMN sla_due_at TIMESTAMPTZ,
    ADD COLUMN sla_escalated_at TIMESTAMPTZ,
    ADD COLUMN public_verification_code VARCHAR(64),
    ADD COLUMN cancelled_by UUID,
    ADD COLUMN cancelled_at TIMESTAMPTZ,
    ADD COLUMN cancellation_reason TEXT,
    ADD CONSTRAINT uq_letter_requests_public_verification_code
        UNIQUE (public_verification_code),
    ADD CONSTRAINT fk_letter_requests_cancelled_by_tenant
        FOREIGN KEY (organization_id, cancelled_by)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT,
    ADD CONSTRAINT chk_letter_requests_cancellation_data
        CHECK (
            status <> 'cancelled'
            OR (
                cancelled_at IS NOT NULL
                AND NULLIF(trim(cancellation_reason), '') IS NOT NULL
            )
        );

CREATE INDEX idx_letter_requests_organization_sla_due_at
    ON letter_requests (organization_id, sla_due_at)
    WHERE sla_due_at IS NOT NULL AND status IN ('submitted', 'under_review', 'needs_revision', 'awaiting_approval', 'approved');

-- 14.7, 14.8: SLA dan timeline aduan.
ALTER TABLE complaint_categories
    ADD COLUMN target_response_hours INTEGER
        CHECK (target_response_hours IS NULL OR target_response_hours > 0),
    ADD COLUMN target_resolution_hours INTEGER
        CHECK (target_resolution_hours IS NULL OR target_resolution_hours > 0),
    ADD COLUMN target_reporter_confirmation_hours INTEGER
        CHECK (
            target_reporter_confirmation_hours IS NULL
            OR target_reporter_confirmation_hours > 0
        );

ALTER TABLE complaints
    ADD COLUMN response_due_at TIMESTAMPTZ,
    ADD COLUMN responded_at TIMESTAMPTZ,
    ADD COLUMN resolution_due_at TIMESTAMPTZ,
    ADD COLUMN reporter_confirmation_due_at TIMESTAMPTZ,
    ADD COLUMN reporter_confirmed_at TIMESTAMPTZ,
    ADD COLUMN closure_reason TEXT,
    ADD COLUMN closed_by UUID,
    ADD CONSTRAINT fk_complaints_closed_by_tenant
        FOREIGN KEY (organization_id, closed_by)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT;

CREATE TABLE complaint_events (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    complaint_id UUID NOT NULL,
    actor_user_id UUID,
    event_type VARCHAR(50) NOT NULL CHECK (event_type <> ''),
    data JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_complaint_events_complaint_tenant
        FOREIGN KEY (organization_id, complaint_id)
        REFERENCES complaints (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_complaint_events_actor_tenant
        FOREIGN KEY (organization_id, actor_user_id)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_complaint_events_organization_complaint_created_at
    ON complaint_events (organization_id, complaint_id, created_at ASC);

CREATE INDEX idx_complaints_organization_response_due_at
    ON complaints (organization_id, response_due_at)
    WHERE response_due_at IS NOT NULL AND responded_at IS NULL;

CREATE INDEX idx_complaints_organization_resolution_due_at
    ON complaints (organization_id, resolution_due_at)
    WHERE resolution_due_at IS NOT NULL AND status NOT IN ('closed', 'rejected');

-- 14.9, 14.10: daftar kerja kualitas data dan evaluasi domisili.
ALTER TABLE households
    ADD COLUMN domicile_review_due_at DATE,
    ADD COLUMN domicile_last_confirmed_at TIMESTAMPTZ;

CREATE INDEX idx_households_organization_domicile_review_due_at
    ON households (organization_id, domicile_review_due_at)
    WHERE domicile_review_due_at IS NOT NULL;

CREATE TABLE domicile_reminder_deliveries (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    household_id UUID NOT NULL,
    user_id UUID NOT NULL,
    reminder_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('sent', 'failed')),
    failure_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_domicile_reminder_deliveries_daily
        UNIQUE (organization_id, household_id, user_id, reminder_date),
    CONSTRAINT fk_domicile_reminder_deliveries_household_tenant
        FOREIGN KEY (organization_id, household_id)
        REFERENCES households (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_domicile_reminder_deliveries_user_tenant
        FOREIGN KEY (organization_id, user_id)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT
);

-- 14.13: serah-terima jabatan append-only.
CREATE TABLE office_handovers (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    outgoing_user_id UUID NOT NULL,
    incoming_user_id UUID,
    status VARCHAR(20) NOT NULL CHECK (status IN ('draft', 'completed', 'cancelled')),
    checklist JSONB NOT NULL DEFAULT '{}'::jsonb,
    notes TEXT,
    completed_by UUID,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_office_handovers_outgoing_tenant
        FOREIGN KEY (organization_id, outgoing_user_id)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_office_handovers_incoming_tenant
        FOREIGN KEY (organization_id, incoming_user_id)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_office_handovers_completed_by_tenant
        FOREIGN KEY (organization_id, completed_by)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT chk_office_handovers_completion
        CHECK (
            (status = 'completed' AND completed_by IS NOT NULL AND completed_at IS NOT NULL)
            OR status <> 'completed'
        )
);

CREATE INDEX idx_office_handovers_organization_status_created_at
    ON office_handovers (organization_id, status, created_at DESC);

-- 14.11: hanya snapshot agregat yang dapat dilihat warga. Detail transaksi,
-- identitas, dan bukti privat tidak pernah dipublikasikan lewat tabel ini.
CREATE TABLE cash_publication_policies (
    organization_id UUID PRIMARY KEY REFERENCES organizations(id) ON DELETE RESTRICT,
    is_public BOOLEAN NOT NULL DEFAULT false,
    public_until DATE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE cash_publication_categories (
    organization_id UUID NOT NULL,
    cash_category_id UUID NOT NULL,
    PRIMARY KEY (organization_id, cash_category_id),
    CONSTRAINT fk_cash_publication_categories_category_tenant
        FOREIGN KEY (organization_id, cash_category_id)
        REFERENCES cash_categories (organization_id, id)
        ON DELETE RESTRICT
);

CREATE TABLE public_cash_summaries (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    total_income NUMERIC(15,2) NOT NULL DEFAULT 0 CHECK (total_income >= 0),
    total_expense NUMERIC(15,2) NOT NULL DEFAULT 0 CHECK (total_expense >= 0),
    ending_balance NUMERIC(15,2) NOT NULL,
    published_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_public_cash_summaries_organization_period
        UNIQUE (organization_id, period_start, period_end),
    CONSTRAINT chk_public_cash_summaries_period
        CHECK (period_end >= period_start)
);

CREATE TABLE public_cash_summary_categories (
    public_cash_summary_id UUID NOT NULL
        REFERENCES public_cash_summaries(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL,
    cash_category_id UUID NOT NULL,
    category_name VARCHAR(255) NOT NULL,
    transaction_type VARCHAR(20) NOT NULL
        CHECK (transaction_type IN ('income', 'expense')),
    total_amount NUMERIC(15,2) NOT NULL DEFAULT 0 CHECK (total_amount >= 0),
    PRIMARY KEY (public_cash_summary_id, cash_category_id, transaction_type),
    CONSTRAINT fk_public_cash_summary_categories_category_tenant
        FOREIGN KEY (organization_id, cash_category_id)
        REFERENCES cash_categories (organization_id, id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_public_cash_summaries_organization_period
    ON public_cash_summaries (organization_id, period_end DESC);

CREATE TRIGGER trg_office_handovers_updated_at
    BEFORE UPDATE ON office_handovers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();