-- Epic 9: Aduan Warga

CREATE TABLE complaints (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    reporter_user_id UUID NOT NULL,
    ticket_number VARCHAR(50) NOT NULL CHECK (ticket_number <> ''),
    category VARCHAR(50) NOT NULL CHECK (category <> ''),
    title VARCHAR(255) NOT NULL CHECK (title <> ''),
    description TEXT NOT NULL CHECK (description <> ''),
    location_description TEXT,
    priority VARCHAR(20) NOT NULL DEFAULT 'normal'
        CHECK (priority IN ('low', 'normal', 'high')),
    status VARCHAR(30) NOT NULL DEFAULT 'new'
        CHECK (status IN (
            'new',
            'reviewed',
            'in_progress',
            'waiting_information',
            'resolved',
            'rejected',
            'closed'
        )),
    assigned_to UUID,
    resolution_note TEXT,
    resolved_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_complaints_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT uq_complaints_organization_ticket_number UNIQUE (organization_id, ticket_number),
    CONSTRAINT fk_complaints_reporter_tenant
        FOREIGN KEY (organization_id, reporter_user_id)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_complaints_assigned_to_tenant
        FOREIGN KEY (organization_id, assigned_to)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT chk_complaints_resolution_data
        CHECK (
            status <> 'resolved'
            OR (resolution_note IS NOT NULL AND resolved_at IS NOT NULL)
        ),
    CONSTRAINT chk_complaints_closed_data
        CHECK (status <> 'closed' OR closed_at IS NOT NULL)
);

CREATE INDEX idx_complaints_organization_status_created_at
    ON complaints (organization_id, status, created_at DESC);

CREATE INDEX idx_complaints_organization_reporter_created_at
    ON complaints (organization_id, reporter_user_id, created_at DESC);

CREATE INDEX idx_complaints_organization_assigned_to_status
    ON complaints (organization_id, assigned_to, status);

CREATE TRIGGER trg_complaints_updated_at
    BEFORE UPDATE ON complaints
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE complaint_comments (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    complaint_id UUID NOT NULL,
    author_user_id UUID NOT NULL,
    body TEXT NOT NULL CHECK (body <> ''),
    is_internal BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_complaint_comments_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT fk_complaint_comments_complaint_tenant
        FOREIGN KEY (organization_id, complaint_id)
        REFERENCES complaints (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_complaint_comments_author_tenant
        FOREIGN KEY (organization_id, author_user_id)
        REFERENCES users (organization_id, id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_complaint_comments_organization_complaint_created_at
    ON complaint_comments (organization_id, complaint_id, created_at ASC);