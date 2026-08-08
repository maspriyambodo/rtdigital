package securityops

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrValidation        = errors.New("invalid request data")
	ErrPostNotFound      = errors.New("patrol post not found")
	ErrScheduleNotFound  = errors.New("patrol schedule not found")
	ErrAssignmentNotFound = errors.New("patrol assignment not found")
	ErrActivityNotFound  = errors.New("community activity not found")
	ErrInviteNotFound    = errors.New("visitor invite not found")
	ErrVisitorLogNotFound = errors.New("visitor log not found")
	ErrAlertNotFound     = errors.New("emergency alert not found")
	ErrDuplicateCode     = errors.New("code already exists")
	ErrInvalidState      = errors.New("invalid state")
)

type PatrolPost struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Code           string    `json:"code"`
	Name           string    `json:"name"`
	Location       *string   `json:"location,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreatePatrolPostRequest struct {
	Code     string  `json:"code"`
	Name     string  `json:"name"`
	Location *string `json:"location,omitempty"`
}

func (r *CreatePatrolPostRequest) Validate() error {
	r.Code = strings.TrimSpace(r.Code)
	r.Name = strings.TrimSpace(r.Name)
	if r.Code == "" || r.Name == "" {
		return ErrValidation
	}
	return nil
}

type PatrolSchedule struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	PostID         string    `json:"post_id"`
	PostName       string    `json:"post_name,omitempty"`
	ShiftDate      string    `json:"shift_date"`
	ShiftStartTime string    `json:"shift_start_time"`
	ShiftEndTime   string    `json:"shift_end_time"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreatePatrolScheduleRequest struct {
	PostID         string `json:"post_id"`
	ShiftDate      string `json:"shift_date"`
	ShiftStartTime string `json:"shift_start_time"`
	ShiftEndTime   string `json:"shift_end_time"`
}

func (r *CreatePatrolScheduleRequest) Validate() error {
	r.PostID = strings.TrimSpace(r.PostID)
	if r.PostID == "" || r.ShiftDate == "" || r.ShiftStartTime == "" || r.ShiftEndTime == "" {
		return ErrValidation
	}
	return nil
}

type PatrolAssignment struct {
	ID                   string    `json:"id"`
	OrganizationID       string    `json:"organization_id"`
	ScheduleID           string    `json:"schedule_id"`
	ResidentID           string    `json:"resident_id"`
	ResidentName         string    `json:"resident_name,omitempty"`
	SubstituteResidentID *string   `json:"substitute_resident_id,omitempty"`
	SubstituteName       *string   `json:"substitute_name,omitempty"`
	Status               string    `json:"status"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type AssignPatrolRequest struct {
	ScheduleID string `json:"schedule_id"`
	ResidentID string `json:"resident_id"`
}

func (r *AssignPatrolRequest) Validate() error {
	r.ScheduleID = strings.TrimSpace(r.ScheduleID)
	r.ResidentID = strings.TrimSpace(r.ResidentID)
	if r.ScheduleID == "" || r.ResidentID == "" {
		return ErrValidation
	}
	return nil
}

type SwapPatrolRequest struct {
	SubstituteResidentID string `json:"substitute_resident_id"`
}

type PatrolAttendance struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organization_id"`
	AssignmentID   string     `json:"assignment_id"`
	ResidentID     string     `json:"resident_id"`
	CheckInTime    time.Time  `json:"check_in_time"`
	CheckOutTime   *time.Time `json:"check_out_time,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type CheckInPatrolRequest struct {
	AssignmentID string  `json:"assignment_id"`
	Notes        *string `json:"notes,omitempty"`
}

type PatrolIncident struct {
	ID              string    `json:"id"`
	OrganizationID  string    `json:"organization_id"`
	ScheduleID      *string   `json:"schedule_id,omitempty"`
	ReporterID      string    `json:"reporter_id"`
	ReporterEmail   string    `json:"reporter_email,omitempty"`
	IncidentTime    time.Time `json:"incident_time"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Severity        string    `json:"severity"`
	Status          string    `json:"status"`
	ResolutionNotes *string   `json:"resolution_notes,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type ReportIncidentRequest struct {
	ScheduleID  *string `json:"schedule_id,omitempty"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"`
}

func (r *ReportIncidentRequest) Validate() error {
	r.Title = strings.TrimSpace(r.Title)
	r.Description = strings.TrimSpace(r.Description)
	if r.Title == "" || r.Description == "" {
		return ErrValidation
	}
	switch r.Severity {
	case "low", "medium", "high", "critical":
	default:
		r.Severity = "medium"
	}
	return nil
}

type CommunityActivity struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Title          string    `json:"title"`
	Description    *string   `json:"description,omitempty"`
	ActivityDate   string    `json:"activity_date"`
	StartTime      string    `json:"start_time"`
	EndTime        *string   `json:"end_time,omitempty"`
	Location       *string   `json:"location,omitempty"`
	TargetType     string    `json:"target_type"`
	IsMandatory    bool      `json:"is_mandatory"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateActivityRequest struct {
	Title        string  `json:"title"`
	Description  *string `json:"description,omitempty"`
	ActivityDate string  `json:"activity_date"`
	StartTime    string  `json:"start_time"`
	EndTime      *string `json:"end_time,omitempty"`
	Location     *string `json:"location,omitempty"`
	TargetType   string  `json:"target_type"`
	IsMandatory  bool    `json:"is_mandatory"`
}

func (r *CreateActivityRequest) Validate() error {
	r.Title = strings.TrimSpace(r.Title)
	if r.Title == "" || r.ActivityDate == "" || r.StartTime == "" {
		return ErrValidation
	}
	if r.TargetType == "" {
		r.TargetType = "all"
	}
	return nil
}

type ActivityAttendance struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	ActivityID     string    `json:"activity_id"`
	HouseholdID    *string   `json:"household_id,omitempty"`
	ResidentID     *string   `json:"resident_id,omitempty"`
	Status         string    `json:"status"`
	Notes          *string   `json:"notes,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type RecordActivityAttendanceRequest struct {
	ActivityID  string  `json:"activity_id"`
	HouseholdID *string `json:"household_id,omitempty"`
	ResidentID  *string `json:"resident_id,omitempty"`
	Status      string  `json:"status"`
	Notes       *string `json:"notes,omitempty"`
}

type VisitorInvite struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	HostResidentID string    `json:"host_resident_id"`
	HostName       string    `json:"host_name,omitempty"`
	VisitorName    string    `json:"visitor_name"`
	Purpose        *string   `json:"purpose,omitempty"`
	ValidFrom      time.Time `json:"valid_from"`
	ValidUntil     time.Time `json:"valid_until"`
	QRCodeHash     string    `json:"qr_code_hash"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

type CreateVisitorInviteRequest struct {
	VisitorName string  `json:"visitor_name"`
	Purpose     *string `json:"purpose,omitempty"`
	ValidFrom   string  `json:"valid_from"`
	ValidUntil  string  `json:"valid_until"`
}

func (r *CreateVisitorInviteRequest) Validate() error {
	r.VisitorName = strings.TrimSpace(r.VisitorName)
	if r.VisitorName == "" || r.ValidFrom == "" || r.ValidUntil == "" {
		return ErrValidation
	}
	return nil
}

type VisitorLog struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organization_id"`
	InviteID       *string    `json:"invite_id,omitempty"`
	HostResidentID *string    `json:"host_resident_id,omitempty"`
	VisitorName    string     `json:"visitor_name"`
	IdentityType   *string    `json:"identity_type,omitempty"`
	IdentityNumber *string    `json:"identity_number,omitempty"`
	VehiclePlate   *string    `json:"vehicle_plate,omitempty"`
	Purpose        *string    `json:"purpose,omitempty"`
	CheckInTime    time.Time  `json:"check_in_time"`
	CheckOutTime   *time.Time `json:"check_out_time,omitempty"`
	GuardID        string     `json:"guard_id"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
}

type CheckInVisitorRequest struct {
	InviteID       *string `json:"invite_id,omitempty"`
	HostResidentID *string `json:"host_resident_id,omitempty"`
	VisitorName    string  `json:"visitor_name"`
	IdentityType   *string `json:"identity_type,omitempty"`
	IdentityNumber *string `json:"identity_number,omitempty"`
	VehiclePlate   *string `json:"vehicle_plate,omitempty"`
	Purpose        *string `json:"purpose,omitempty"`
}

type EmergencyAlert struct {
	ID              string     `json:"id"`
	OrganizationID  string     `json:"organization_id"`
	ReporterID      string     `json:"reporter_id"`
	ReporterName    string     `json:"reporter_name,omitempty"`
	Category        string     `json:"category"`
	Latitude        *float64   `json:"latitude,omitempty"`
	Longitude       *float64   `json:"longitude,omitempty"`
	LocationDetails *string    `json:"location_details,omitempty"`
	Status          string     `json:"status"`
	AcknowledgedBy  *string    `json:"acknowledged_by,omitempty"`
	AcknowledgedAt  *time.Time `json:"acknowledged_at,omitempty"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
	ResolutionNotes *string    `json:"resolution_notes,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type CreateEmergencyAlertRequest struct {
	Category        string   `json:"category"`
	Latitude        *float64 `json:"latitude,omitempty"`
	Longitude       *float64 `json:"longitude,omitempty"`
	LocationDetails *string  `json:"location_details,omitempty"`
}

func (r *CreateEmergencyAlertRequest) Validate() error {
	r.Category = strings.TrimSpace(r.Category)
	switch r.Category {
	case "fire", "medical", "crime", "accident", "other":
	default:
		return ErrValidation
	}
	return nil
}
