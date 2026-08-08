package securityops

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) ListPatrolPosts(ctx context.Context, principal *auth.Principal) ([]PatrolPost, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, organization_id, code, name, location, status, created_at, updated_at
		FROM patrol_posts
		WHERE organization_id = $1 AND status != 'deleted'
		ORDER BY name ASC`, principal.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("query patrol posts: %w", err)
	}
	defer rows.Close()

	items := []PatrolPost{}
	for rows.Next() {
		var item PatrolPost
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.Code, &item.Name, &item.Location, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan patrol post: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) CreatePatrolPost(ctx context.Context, principal *auth.Principal, req CreatePatrolPostRequest) (*PatrolPost, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var item PatrolPost
	err := s.db.QueryRow(ctx, `
		INSERT INTO patrol_posts (organization_id, code, name, location)
		VALUES ($1, $2, $3, $4)
		RETURNING id, organization_id, code, name, location, status, created_at, updated_at`,
		principal.OrganizationID, req.Code, req.Name, req.Location,
	).Scan(&item.ID, &item.OrganizationID, &item.Code, &item.Name, &item.Location, &item.Status, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		if isDuplicate(err) {
			return nil, ErrDuplicateCode
		}
		return nil, fmt.Errorf("create patrol post: %w", err)
	}

	s.audit(ctx, principal, "patrol_post.create", "patrol_posts", item.ID)
	return &item, nil
}

func (s *Service) ListPatrolSchedules(ctx context.Context, principal *auth.Principal, postID, date string) ([]PatrolSchedule, error) {
	query := `
		SELECT s.id, s.organization_id, s.post_id, p.name, s.shift_date, s.shift_start_time, s.shift_end_time, s.status, s.created_at, s.updated_at
		FROM patrol_schedules s
		JOIN patrol_posts p ON p.id = s.post_id
		WHERE s.organization_id = $1`
	args := []any{principal.OrganizationID}

	if postID != "" {
		args = append(args, postID)
		query += fmt.Sprintf(" AND s.post_id = $%d", len(args))
	}
	if date != "" {
		args = append(args, date)
		query += fmt.Sprintf(" AND s.shift_date = $%d", len(args))
	}
	query += " ORDER BY s.shift_date DESC, s.shift_start_time ASC"

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query patrol schedules: %w", err)
	}
	defer rows.Close()

	items := []PatrolSchedule{}
	for rows.Next() {
		var item PatrolSchedule
		var sDate time.Time
		var sTime, eTime string
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.PostID, &item.PostName, &sDate, &sTime, &eTime, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan patrol schedule: %w", err)
		}
		item.ShiftDate = sDate.Format("2006-01-02")
		item.ShiftStartTime = sTime
		item.ShiftEndTime = eTime
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) CreatePatrolSchedule(ctx context.Context, principal *auth.Principal, req CreatePatrolScheduleRequest) (*PatrolSchedule, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	shiftDate, err := time.Parse("2006-01-02", req.ShiftDate)
	if err != nil {
		return nil, ErrValidation
	}

	var item PatrolSchedule
	var sTime, eTime string
	err = s.db.QueryRow(ctx, `
		INSERT INTO patrol_schedules (organization_id, post_id, shift_date, shift_start_time, shift_end_time)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, organization_id, post_id, shift_date, shift_start_time, shift_end_time, status, created_at, updated_at`,
		principal.OrganizationID, req.PostID, shiftDate, req.ShiftStartTime, req.ShiftEndTime,
	).Scan(&item.ID, &item.OrganizationID, &item.PostID, &shiftDate, &sTime, &eTime, &item.Status, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		if isDuplicate(err) {
			return nil, ErrDuplicateCode
		}
		return nil, fmt.Errorf("create patrol schedule: %w", err)
	}

	item.ShiftDate = shiftDate.Format("2006-01-02")
	item.ShiftStartTime = sTime
	item.ShiftEndTime = eTime
	s.audit(ctx, principal, "patrol_schedule.create", "patrol_schedules", item.ID)
	return &item, nil
}


func (s *Service) AssignPatrol(ctx context.Context, principal *auth.Principal, req AssignPatrolRequest) (*PatrolAssignment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var item PatrolAssignment
	err := s.db.QueryRow(ctx, `
		INSERT INTO patrol_assignments (organization_id, schedule_id, resident_id)
		VALUES ($1, $2, $3)
		RETURNING id, organization_id, schedule_id, resident_id, substitute_resident_id, status, created_at, updated_at`,
		principal.OrganizationID, req.ScheduleID, req.ResidentID,
	).Scan(&item.ID, &item.OrganizationID, &item.ScheduleID, &item.ResidentID, &item.SubstituteResidentID, &item.Status, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		if isDuplicate(err) {
			return nil, ErrDuplicateCode
		}
		return nil, fmt.Errorf("assign patrol: %w", err)
	}

	s.audit(ctx, principal, "patrol_assignment.create", "patrol_assignments", item.ID)
	return &item, nil
}

func (s *Service) SwapPatrol(ctx context.Context, principal *auth.Principal, assignmentID string, req SwapPatrolRequest) (*PatrolAssignment, error) {
	subID := strings.TrimSpace(req.SubstituteResidentID)
	if subID == "" {
		return nil, ErrValidation
	}

	var item PatrolAssignment
	err := s.db.QueryRow(ctx, `
		UPDATE patrol_assignments
		SET substitute_resident_id = $1, status = 'substituted', updated_at = NOW()
		WHERE id = $2 AND organization_id = $3
		RETURNING id, organization_id, schedule_id, resident_id, substitute_resident_id, status, created_at, updated_at`,
		subID, assignmentID, principal.OrganizationID,
	).Scan(&item.ID, &item.OrganizationID, &item.ScheduleID, &item.ResidentID, &item.SubstituteResidentID, &item.Status, &item.CreatedAt, &item.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAssignmentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("swap patrol: %w", err)
	}

	s.audit(ctx, principal, "patrol_assignment.swap", "patrol_assignments", item.ID)
	return &item, nil
}

func (s *Service) CheckInPatrol(ctx context.Context, principal *auth.Principal, req CheckInPatrolRequest) (*PatrolAttendance, error) {
	if strings.TrimSpace(req.AssignmentID) == "" {
		return nil, ErrValidation
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var residentID string
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(substitute_resident_id, resident_id)
		FROM patrol_assignments
		WHERE id = $1 AND organization_id = $2 FOR UPDATE`,
		req.AssignmentID, principal.OrganizationID,
	).Scan(&residentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAssignmentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get assignment resident: %w", err)
	}

	var att PatrolAttendance
	err = tx.QueryRow(ctx, `
		INSERT INTO patrol_attendances (organization_id, assignment_id, resident_id, notes)
		VALUES ($1, $2, $3, $4)
		RETURNING id, organization_id, assignment_id, resident_id, check_in_time, check_out_time, notes, created_at`,
		principal.OrganizationID, req.AssignmentID, residentID, req.Notes,
	).Scan(&att.ID, &att.OrganizationID, &att.AssignmentID, &att.ResidentID, &att.CheckInTime, &att.CheckOutTime, &att.Notes, &att.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("check in patrol: %w", err)
	}

	if _, err := tx.Exec(ctx, `UPDATE patrol_assignments SET status = 'attended', updated_at = NOW() WHERE id = $1`, req.AssignmentID); err != nil {
		return nil, fmt.Errorf("update assignment status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit patrol check in: %w", err)
	}

	s.audit(ctx, principal, "patrol_attendance.checkin", "patrol_attendances", att.ID)
	return &att, nil
}

func (s *Service) ReportIncident(ctx context.Context, principal *auth.Principal, req ReportIncidentRequest) (*PatrolIncident, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var item PatrolIncident
	err := s.db.QueryRow(ctx, `
		INSERT INTO patrol_incidents (organization_id, schedule_id, reporter_id, incident_time, title, description, severity)
		VALUES ($1, $2, $3, NOW(), $4, $5, $6)
		RETURNING id, organization_id, schedule_id, reporter_id, incident_time, title, description, severity, status, resolution_notes, created_at`,
		principal.OrganizationID, req.ScheduleID, principal.UserID, req.Title, req.Description, req.Severity,
	).Scan(&item.ID, &item.OrganizationID, &item.ScheduleID, &item.ReporterID, &item.IncidentTime, &item.Title, &item.Description, &item.Severity, &item.Status, &item.ResolutionNotes, &item.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("report incident: %w", err)
	}

	s.audit(ctx, principal, "patrol_incident.create", "patrol_incidents", item.ID)
	return &item, nil
}

func (s *Service) ListIncidents(ctx context.Context, principal *auth.Principal) ([]PatrolIncident, error) {
	rows, err := s.db.Query(ctx, `
		SELECT i.id, i.organization_id, i.schedule_id, i.reporter_id, u.email, i.incident_time, i.title, i.description, i.severity, i.status, i.resolution_notes, i.created_at
		FROM patrol_incidents i
		JOIN users u ON u.id = i.reporter_id
		WHERE i.organization_id = $1
		ORDER BY i.incident_time DESC`, principal.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("query incidents: %w", err)
	}
	defer rows.Close()

	items := []PatrolIncident{}
	for rows.Next() {
		var item PatrolIncident
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.ScheduleID, &item.ReporterID, &item.ReporterEmail, &item.IncidentTime, &item.Title, &item.Description, &item.Severity, &item.Status, &item.ResolutionNotes, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan incident: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) ListActivities(ctx context.Context, principal *auth.Principal) ([]CommunityActivity, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, organization_id, title, description, activity_date, start_time, end_time, location, target_type, is_mandatory, status, created_at, updated_at
		FROM community_activities
		WHERE organization_id = $1
		ORDER BY activity_date DESC`, principal.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("query activities: %w", err)
	}
	defer rows.Close()

	items := []CommunityActivity{}
	for rows.Next() {
		var item CommunityActivity
		var aDate time.Time
		var sTime string
		var eTime *string
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.Title, &item.Description, &aDate, &sTime, &eTime, &item.Location, &item.TargetType, &item.IsMandatory, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan activity: %w", err)
		}
		item.ActivityDate = aDate.Format("2006-01-02")
		item.StartTime = sTime
		item.EndTime = eTime
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) CreateActivity(ctx context.Context, principal *auth.Principal, req CreateActivityRequest) (*CommunityActivity, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	aDate, err := time.Parse("2006-01-02", req.ActivityDate)
	if err != nil {
		return nil, ErrValidation
	}

	var item CommunityActivity
	var sTime string
	var eTime *string
	err = s.db.QueryRow(ctx, `
		INSERT INTO community_activities (organization_id, title, description, activity_date, start_time, end_time, location, target_type, is_mandatory, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'scheduled')
		RETURNING id, organization_id, title, description, activity_date, start_time, end_time, location, target_type, is_mandatory, status, created_at, updated_at`,
		principal.OrganizationID, req.Title, req.Description, aDate, req.StartTime, req.EndTime, req.Location, req.TargetType, req.IsMandatory,
	).Scan(&item.ID, &item.OrganizationID, &item.Title, &item.Description, &aDate, &sTime, &eTime, &item.Location, &item.TargetType, &item.IsMandatory, &item.Status, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("create activity: %w", err)
	}

	item.ActivityDate = aDate.Format("2006-01-02")
	item.StartTime = sTime
	item.EndTime = eTime
	s.audit(ctx, principal, "activity.create", "community_activities", item.ID)
	return &item, nil
}

func (s *Service) CreateVisitorInvite(ctx context.Context, principal *auth.Principal, residentID string, req CreateVisitorInviteRequest) (*VisitorInvite, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	vFrom, err1 := time.Parse(time.RFC3339, req.ValidFrom)
	vUntil, err2 := time.Parse(time.RFC3339, req.ValidUntil)
	if err1 != nil || err2 != nil || vUntil.Before(vFrom) {
		return nil, ErrValidation
	}

	hashInput := fmt.Sprintf("%s-%s-%s-%d", principal.OrganizationID, residentID, req.VisitorName, time.Now().UnixNano())
	hashSum := sha256.Sum256([]byte(hashInput))
	qrHash := hex.EncodeToString(hashSum[:])

	var item VisitorInvite
	err := s.db.QueryRow(ctx, `
		INSERT INTO visitor_invites (organization_id, host_resident_id, visitor_name, purpose, valid_from, valid_until, qr_code_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, organization_id, host_resident_id, visitor_name, purpose, valid_from, valid_until, qr_code_hash, status, created_at`,
		principal.OrganizationID, residentID, req.VisitorName, req.Purpose, vFrom, vUntil, qrHash,
	).Scan(&item.ID, &item.OrganizationID, &item.HostResidentID, &item.VisitorName, &item.Purpose, &item.ValidFrom, &item.ValidUntil, &item.QRCodeHash, &item.Status, &item.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("create visitor invite: %w", err)
	}

	s.audit(ctx, principal, "visitor_invite.create", "visitor_invites", item.ID)
	return &item, nil
}

func (s *Service) CheckInVisitor(ctx context.Context, principal *auth.Principal, req CheckInVisitorRequest) (*VisitorLog, error) {
	if strings.TrimSpace(req.VisitorName) == "" {
		return nil, ErrValidation
	}

	var item VisitorLog
	err := s.db.QueryRow(ctx, `
		INSERT INTO visitor_logs (organization_id, invite_id, host_resident_id, visitor_name, identity_type, identity_number, vehicle_plate, purpose, guard_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, organization_id, invite_id, host_resident_id, visitor_name, identity_type, identity_number, vehicle_plate, purpose, check_in_time, check_out_time, guard_id, status, created_at`,
		principal.OrganizationID, req.InviteID, req.HostResidentID, req.VisitorName, req.IdentityType, req.IdentityNumber, req.VehiclePlate, req.Purpose, principal.UserID,
	).Scan(&item.ID, &item.OrganizationID, &item.InviteID, &item.HostResidentID, &item.VisitorName, &item.IdentityType, &item.IdentityNumber, &item.VehiclePlate, &item.Purpose, &item.CheckInTime, &item.CheckOutTime, &item.GuardID, &item.Status, &item.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("check in visitor: %w", err)
	}

	if req.InviteID != nil && *req.InviteID != "" {
		_, _ = s.db.Exec(ctx, `UPDATE visitor_invites SET status = 'used', updated_at = NOW() WHERE id = $1`, *req.InviteID)
	}

	s.audit(ctx, principal, "visitor_log.checkin", "visitor_logs", item.ID)
	return &item, nil
}

func (s *Service) ListVisitorLogs(ctx context.Context, principal *auth.Principal) ([]VisitorLog, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, organization_id, invite_id, host_resident_id, visitor_name, identity_type, identity_number, vehicle_plate, purpose, check_in_time, check_out_time, guard_id, status, created_at
		FROM visitor_logs
		WHERE organization_id = $1
		ORDER BY check_in_time DESC`, principal.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("query visitor logs: %w", err)
	}
	defer rows.Close()

	items := []VisitorLog{}
	for rows.Next() {
		var item VisitorLog
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.InviteID, &item.HostResidentID, &item.VisitorName, &item.IdentityType, &item.IdentityNumber, &item.VehiclePlate, &item.Purpose, &item.CheckInTime, &item.CheckOutTime, &item.GuardID, &item.Status, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan visitor log: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) CreateEmergencyAlert(ctx context.Context, principal *auth.Principal, residentID string, req CreateEmergencyAlertRequest) (*EmergencyAlert, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var item EmergencyAlert
	err := s.db.QueryRow(ctx, `
		INSERT INTO emergency_alerts (organization_id, reporter_id, category, latitude, longitude, location_details)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, organization_id, reporter_id, category, latitude, longitude, location_details, status, acknowledged_by, acknowledged_at, resolved_at, resolution_notes, created_at`,
		principal.OrganizationID, residentID, req.Category, req.Latitude, req.Longitude, req.LocationDetails,
	).Scan(&item.ID, &item.OrganizationID, &item.ReporterID, &item.Category, &item.Latitude, &item.Longitude, &item.LocationDetails, &item.Status, &item.AcknowledgedBy, &item.AcknowledgedAt, &item.ResolvedAt, &item.ResolutionNotes, &item.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("create emergency alert: %w", err)
	}

	s.audit(ctx, principal, "emergency_alert.create", "emergency_alerts", item.ID)
	return &item, nil
}

func (s *Service) ListEmergencyAlerts(ctx context.Context, principal *auth.Principal, status string) ([]EmergencyAlert, error) {
	query := `
		SELECT a.id, a.organization_id, a.reporter_id, r.name, a.category, a.latitude, a.longitude, a.location_details, a.status, a.acknowledged_by, a.acknowledged_at, a.resolved_at, a.resolution_notes, a.created_at
		FROM emergency_alerts a
		JOIN residents r ON r.id = a.reporter_id
		WHERE a.organization_id = $1`
	args := []any{principal.OrganizationID}

	if status != "" {
		args = append(args, status)
		query += fmt.Sprintf(" AND a.status = $%d", len(args))
	}
	query += " ORDER BY a.created_at DESC"

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query emergency alerts: %w", err)
	}
	defer rows.Close()

	items := []EmergencyAlert{}
	for rows.Next() {
		var item EmergencyAlert
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.ReporterID, &item.ReporterName, &item.Category, &item.Latitude, &item.Longitude, &item.LocationDetails, &item.Status, &item.AcknowledgedBy, &item.AcknowledgedAt, &item.ResolvedAt, &item.ResolutionNotes, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan emergency alert: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) AcknowledgeEmergencyAlert(ctx context.Context, principal *auth.Principal, id string) (*EmergencyAlert, error) {
	var item EmergencyAlert
	now := time.Now().UTC()
	err := s.db.QueryRow(ctx, `
		UPDATE emergency_alerts
		SET status = 'acknowledged', acknowledged_by = $1, acknowledged_at = $2, updated_at = NOW()
		WHERE id = $3 AND organization_id = $4 AND status = 'active'
		RETURNING id, organization_id, reporter_id, category, latitude, longitude, location_details, status, acknowledged_by, acknowledged_at, resolved_at, resolution_notes, created_at`,
		principal.UserID, now, id, principal.OrganizationID,
	).Scan(&item.ID, &item.OrganizationID, &item.ReporterID, &item.Category, &item.Latitude, &item.Longitude, &item.LocationDetails, &item.Status, &item.AcknowledgedBy, &item.AcknowledgedAt, &item.ResolvedAt, &item.ResolutionNotes, &item.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAlertNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("acknowledge emergency alert: %w", err)
	}

	s.audit(ctx, principal, "emergency_alert.acknowledge", "emergency_alerts", item.ID)
	return &item, nil
}

func (s *Service) audit(ctx context.Context, principal *auth.Principal, action, entityType, entityID string) {
	_, _ = s.db.Exec(ctx, `
		INSERT INTO audit_logs (organization_id, actor_user_id, action, entity_type, entity_id)
		VALUES ($1, $2, $3, $4, $5)`,
		principal.OrganizationID, principal.UserID, action, entityType, entityID,
	)
}

func isDuplicate(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "23505"))
}
