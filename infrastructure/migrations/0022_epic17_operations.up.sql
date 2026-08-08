CREATE TABLE patrol_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    location TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'deleted')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (organization_id, code)
);

CREATE TABLE patrol_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    post_id UUID NOT NULL REFERENCES patrol_posts(id),
    shift_date DATE NOT NULL,
    shift_start_time TIME NOT NULL,
    shift_end_time TIME NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'in_progress', 'completed', 'cancelled')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, post_id, shift_date)
);

CREATE TABLE patrol_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    schedule_id UUID NOT NULL REFERENCES patrol_schedules(id),
    resident_id UUID NOT NULL REFERENCES residents(id),
    substitute_resident_id UUID REFERENCES residents(id),
    status VARCHAR(20) NOT NULL DEFAULT 'assigned' CHECK (status IN ('assigned', 'substituted', 'excused', 'absent', 'attended')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (schedule_id, resident_id)
);

CREATE TABLE patrol_attendances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    assignment_id UUID NOT NULL REFERENCES patrol_assignments(id) UNIQUE,
    resident_id UUID NOT NULL REFERENCES residents(id),
    check_in_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    check_out_time TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE patrol_incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    schedule_id UUID REFERENCES patrol_schedules(id),
    reporter_id UUID NOT NULL REFERENCES users(id),
    incident_time TIMESTAMP WITH TIME ZONE NOT NULL,
CREATE TABLE community_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    activity_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME,
    location TEXT,
    target_type VARCHAR(20) NOT NULL DEFAULT 'all' CHECK (target_type IN ('all', 'household', 'resident', 'role')),
    is_mandatory BOOLEAN NOT NULL DEFAULT false,
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled' CHECK (status IN ('draft', 'scheduled', 'ongoing', 'completed', 'cancelled')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE activity_attendances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    activity_id UUID NOT NULL REFERENCES community_activities(id),
    household_id UUID REFERENCES households(id),
    resident_id UUID REFERENCES residents(id),
    status VARCHAR(20) NOT NULL DEFAULT 'attended' CHECK (status IN ('attended', 'absent', 'excused')),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (
        (household_id IS NOT NULL) OR (resident_id IS NOT NULL)
    ),
    UNIQUE(activity_id, household_id),
    UNIQUE(activity_id, resident_id)
);

CREATE TABLE visitor_invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    host_resident_id UUID NOT NULL REFERENCES residents(id),
    visitor_name VARCHAR(255) NOT NULL,
    purpose TEXT,
    valid_from TIMESTAMP WITH TIME ZONE NOT NULL,
    valid_until TIMESTAMP WITH TIME ZONE NOT NULL,
    qr_code_hash VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'used', 'expired', 'cancelled')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE visitor_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    invite_id UUID REFERENCES visitor_invites(id),
    host_resident_id UUID REFERENCES residents(id),
    visitor_name VARCHAR(255) NOT NULL,
    identity_type VARCHAR(50),
    identity_number VARCHAR(100),
    vehicle_plate VARCHAR(50),
    purpose TEXT,
    check_in_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    check_out_time TIMESTAMP WITH TIME ZONE,
    guard_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'checked_in' CHECK (status IN ('checked_in', 'checked_out')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE emergency_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    reporter_id UUID NOT NULL REFERENCES residents(id),
    category VARCHAR(50) NOT NULL CHECK (category IN ('fire', 'medical', 'crime', 'accident', 'other')),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    location_details TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'acknowledged', 'resolved', 'false_alarm')),
    acknowledged_by UUID REFERENCES users(id),
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolution_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_patrol_posts_updated_at BEFORE UPDATE ON patrol_posts FOR EACH ROW EXECUTE FUNCTION update_modified_column();
CREATE TRIGGER update_patrol_schedules_updated_at BEFORE UPDATE ON patrol_schedules FOR EACH ROW EXECUTE FUNCTION update_modified_column();
CREATE TRIGGER update_patrol_assignments_updated_at BEFORE UPDATE ON patrol_assignments FOR EACH ROW EXECUTE FUNCTION update_modified_column();
CREATE TRIGGER update_patrol_attendances_updated_at BEFORE UPDATE ON patrol_attendances FOR EACH ROW EXECUTE FUNCTION update_modified_column();
CREATE TRIGGER update_patrol_incidents_updated_at BEFORE UPDATE ON patrol_incidents FOR EACH ROW EXECUTE FUNCTION update_modified_column();
CREATE TRIGGER update_community_activities_updated_at BEFORE UPDATE ON community_activities FOR EACH ROW EXECUTE FUNCTION update_modified_column();
CREATE TRIGGER update_activity_attendances_updated_at BEFORE UPDATE ON activity_attendances FOR EACH ROW EXECUTE FUNCTION update_modified_column();
CREATE TRIGGER update_visitor_invites_updated_at BEFORE UPDATE ON visitor_invites FOR EACH ROW EXECUTE FUNCTION update_modified_column();
CREATE TRIGGER update_visitor_logs_updated_at BEFORE UPDATE ON visitor_logs FOR EACH ROW EXECUTE FUNCTION update_modified_column();
CREATE TRIGGER update_emergency_alerts_updated_at BEFORE UPDATE ON emergency_alerts FOR EACH ROW EXECUTE FUNCTION update_modified_column();

    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'low' CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    status VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'resolved', 'closed')),
    resolution_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
