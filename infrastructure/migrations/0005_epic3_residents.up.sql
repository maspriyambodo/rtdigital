-- Epic 3: Data Keluarga dan Warga.

CREATE TABLE house_units (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    code VARCHAR(50) NOT NULL,
    address_detail TEXT,
    occupancy_status VARCHAR(20) NOT NULL CHECK (occupancy_status IN ('owned', 'rented', 'contract', 'empty')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_house_units_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT uq_house_units_organization_code UNIQUE (organization_id, code)
);

CREATE TRIGGER trg_house_units_updated_at
    BEFORE UPDATE ON house_units
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE residents (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    national_id_encrypted TEXT,
    national_id_blind_index CHAR(64),
    full_name VARCHAR(255) NOT NULL,
    birth_place VARCHAR(100),
    birth_date DATE,
    gender VARCHAR(20) CHECK (gender IN ('male', 'female')),
    marital_status VARCHAR(30),
    occupation VARCHAR(100),
    education VARCHAR(100),
    phone VARCHAR(30),
    email VARCHAR(255),
    resident_status VARCHAR(20) NOT NULL CHECK (resident_status IN ('active', 'moved', 'deceased', 'inactive')),
    verification_status VARCHAR(20) NOT NULL CHECK (verification_status IN ('unverified', 'verified', 'rejected')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_residents_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT chk_residents_national_id_pair CHECK (
        (national_id_encrypted IS NULL) = (national_id_blind_index IS NULL)
    )
);

CREATE UNIQUE INDEX uq_residents_organization_national_id
    ON residents (organization_id, national_id_blind_index)
    WHERE national_id_blind_index IS NOT NULL;

CREATE INDEX idx_residents_organization_name
    ON residents (organization_id, full_name);

CREATE TRIGGER trg_residents_updated_at
    BEFORE UPDATE ON residents
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE households (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    house_unit_id UUID NOT NULL,
    internal_number VARCHAR(50) NOT NULL,
    family_card_number_encrypted TEXT,
    family_card_blind_index CHAR(64),
    head_resident_id UUID,
    domicile_status VARCHAR(20) NOT NULL CHECK (domicile_status IN ('permanent', 'temporary')),
    move_in_date DATE,
    move_out_date DATE,
    verification_status VARCHAR(20) NOT NULL CHECK (verification_status IN ('unverified', 'verified', 'rejected')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_households_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT uq_households_organization_internal_number UNIQUE (organization_id, internal_number),
    CONSTRAINT chk_households_family_card_pair CHECK (
        (family_card_number_encrypted IS NULL) = (family_card_blind_index IS NULL)
    ),
    CONSTRAINT chk_households_move_dates CHECK (
        move_out_date IS NULL OR move_in_date IS NULL OR move_out_date >= move_in_date
    ),
    CONSTRAINT fk_households_house_unit_tenant
        FOREIGN KEY (organization_id, house_unit_id)
        REFERENCES house_units (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_households_head_resident_tenant
        FOREIGN KEY (organization_id, head_resident_id)
        REFERENCES residents (organization_id, id)
        ON DELETE RESTRICT
);

CREATE UNIQUE INDEX uq_households_organization_family_card
    ON households (organization_id, family_card_blind_index)
    WHERE family_card_blind_index IS NOT NULL;

CREATE INDEX idx_households_organization_internal_number
    ON households (organization_id, internal_number);

CREATE TRIGGER trg_households_updated_at
    BEFORE UPDATE ON households
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE household_members (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    household_id UUID NOT NULL,
    resident_id UUID NOT NULL,
    relationship VARCHAR(50) NOT NULL CHECK (relationship IN ('head', 'spouse', 'child', 'parent', 'other')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    started_at DATE NOT NULL DEFAULT current_date,
    ended_at DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_household_members_organization_id_id UNIQUE (organization_id, id),
    CONSTRAINT chk_household_members_dates CHECK (
        (is_active AND ended_at IS NULL) OR (NOT is_active AND ended_at IS NOT NULL AND ended_at >= started_at)
    ),
    CONSTRAINT fk_household_members_household_tenant
        FOREIGN KEY (organization_id, household_id)
        REFERENCES households (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_household_members_resident_tenant
        FOREIGN KEY (organization_id, resident_id)
        REFERENCES residents (organization_id, id)
        ON DELETE RESTRICT
);

CREATE UNIQUE INDEX uq_active_household_member_per_resident
    ON household_members (organization_id, resident_id)
    WHERE is_active;

CREATE UNIQUE INDEX uq_active_household_head
    ON household_members (organization_id, household_id)
    WHERE is_active AND relationship = 'head';

CREATE INDEX idx_household_members_household_active
    ON household_members (organization_id, household_id)
    WHERE is_active;

CREATE TRIGGER trg_household_members_updated_at
    BEFORE UPDATE ON household_members
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE users
    ADD CONSTRAINT fk_users_resident_tenant
    FOREIGN KEY (organization_id, resident_id)
    REFERENCES residents (organization_id, id)
    ON DELETE SET NULL;

CREATE OR REPLACE FUNCTION validate_household_head()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.head_resident_id IS NOT NULL AND NOT EXISTS (
        SELECT 1
        FROM household_members hm
        WHERE hm.organization_id = NEW.organization_id
          AND hm.household_id = NEW.id
          AND hm.resident_id = NEW.head_resident_id
          AND hm.relationship = 'head'
          AND hm.is_active
    ) THEN
        RAISE EXCEPTION 'head_resident_id must be an active household head member';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER trg_households_validate_head
    AFTER INSERT OR UPDATE OF head_resident_id, organization_id ON households
    DEFERRABLE INITIALLY DEFERRED
    FOR EACH ROW EXECUTE FUNCTION validate_household_head();