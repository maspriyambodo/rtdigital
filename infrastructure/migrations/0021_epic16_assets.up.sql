CREATE TABLE asset_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, code)
);

CREATE TABLE asset_locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, code)
);

CREATE TABLE assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    category_id UUID NOT NULL REFERENCES asset_categories(id),
    location_id UUID NOT NULL REFERENCES asset_locations(id),
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    condition VARCHAR(50) NOT NULL CHECK (condition IN ('good', 'fair', 'poor', 'broken')),
    status VARCHAR(50) NOT NULL CHECK (status IN ('available', 'borrowed', 'maintenance', 'inactive', 'disposed')),
    acquisition_date DATE,
    acquisition_value DECIMAL(15, 2),
    pic_id UUID REFERENCES users(id),
    file_object_id UUID REFERENCES file_objects(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, code)
);

CREATE TABLE asset_loans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    asset_id UUID NOT NULL REFERENCES assets(id),
    borrower_id UUID NOT NULL REFERENCES users(id),
    approver_id UUID REFERENCES users(id),
    loan_date DATE NOT NULL,
    due_date DATE NOT NULL,
    return_date DATE,
    condition_out VARCHAR(50) NOT NULL,
    condition_in VARCHAR(50),
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'approved', 'rejected', 'returned')),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE asset_maintenance_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    asset_id UUID NOT NULL REFERENCES assets(id),
    maintenance_date DATE NOT NULL,
    maintenance_type VARCHAR(100) NOT NULL,
    cost DECIMAL(15, 2),
    technician VARCHAR(255),
    notes TEXT,
    file_object_id UUID REFERENCES file_objects(id),
    condition_after VARCHAR(50) NOT NULL,
    status_after VARCHAR(50) NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_asset_categories_org ON asset_categories(organization_id);
CREATE INDEX idx_asset_locations_org ON asset_locations(organization_id);
CREATE INDEX idx_assets_org ON assets(organization_id);
CREATE INDEX idx_asset_loans_org ON asset_loans(organization_id);
CREATE INDEX idx_asset_maintenance_logs_org ON asset_maintenance_logs(organization_id);