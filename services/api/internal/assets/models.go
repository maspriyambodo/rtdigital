package assets

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrValidation        = errors.New("invalid request data")
	ErrCategoryNotFound  = errors.New("asset category not found")
	ErrLocationNotFound  = errors.New("asset location not found")
	ErrAssetNotFound     = errors.New("asset not found")
	ErrLoanNotFound      = errors.New("asset loan not found")
	ErrDuplicateCode     = errors.New("asset code already exists")
	ErrConstraint        = errors.New("business constraint violation")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidState      = errors.New("invalid asset state")
)

type AssetCategory struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Code           string    `json:"code"`
	Name           string    `json:"name"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateCategoryRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func (r *CreateCategoryRequest) Validate() error {
	r.Code = strings.TrimSpace(r.Code)
	r.Name = strings.TrimSpace(r.Name)
	if r.Code == "" || r.Name == "" {
		return ErrValidation
	}
	return nil
}

type UpdateCategoryRequest struct {
	Name   *string `json:"name,omitempty"`
	Status *string `json:"status,omitempty"`
}

func (r *UpdateCategoryRequest) Validate() error {
	if r.Name == nil && r.Status == nil {
		return ErrValidation
	}
	if r.Name != nil {
		name := strings.TrimSpace(*r.Name)
		if name == "" {
			return ErrValidation
		}
		r.Name = &name
	}
	if r.Status != nil && *r.Status != "active" && *r.Status != "inactive" {
		return ErrValidation
	}
	return nil
}

type AssetLocation struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Code           string    `json:"code"`
	Name           string    `json:"name"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateLocationRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func (r *CreateLocationRequest) Validate() error {
	r.Code = strings.TrimSpace(r.Code)
	r.Name = strings.TrimSpace(r.Name)
	if r.Code == "" || r.Name == "" {
		return ErrValidation
	}
	return nil
}

type UpdateLocationRequest struct {
	Name   *string `json:"name,omitempty"`
	Status *string `json:"status,omitempty"`
}

func (r *UpdateLocationRequest) Validate() error {
	if r.Name == nil && r.Status == nil {
		return ErrValidation
	}
	if r.Name != nil {
		name := strings.TrimSpace(*r.Name)
		if name == "" {
			return ErrValidation
		}
		r.Name = &name
	}
	if r.Status != nil && *r.Status != "active" && *r.Status != "inactive" {
		return ErrValidation
	}
	return nil
}


type Asset struct {
	ID               string    `json:"id"`
	OrganizationID   string    `json:"organization_id"`
	CategoryID       string    `json:"category_id"`
	CategoryName     string    `json:"category_name,omitempty"`
	LocationID       string    `json:"location_id"`
	LocationName     string    `json:"location_name,omitempty"`
	Code             string    `json:"code"`
	Name             string    `json:"name"`
	Description      *string   `json:"description,omitempty"`
	Condition        string    `json:"condition"` // good, fair, poor, broken
	Status           string    `json:"status"`    // available, borrowed, maintenance, inactive, disposed
	AcquisitionDate  *string   `json:"acquisition_date,omitempty"`
	AcquisitionValue *float64  `json:"acquisition_value,omitempty"`
	PICID            *string   `json:"pic_id,omitempty"`
	PICName          *string   `json:"pic_name,omitempty"`
	FileObjectID     *string   `json:"file_object_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CreateAssetRequest struct {
	CategoryID       string   `json:"category_id"`
	LocationID       string   `json:"location_id"`
	Code             string   `json:"code"`
	Name             string   `json:"name"`
	Description      *string  `json:"description,omitempty"`
	Condition        string   `json:"condition"`
	AcquisitionDate  *string  `json:"acquisition_date,omitempty"`
	AcquisitionValue *float64 `json:"acquisition_value,omitempty"`
	PICID            *string  `json:"pic_id,omitempty"`
	FileObjectID     *string  `json:"file_object_id,omitempty"`
}

func (r *CreateAssetRequest) Validate() error {
	r.CategoryID = strings.TrimSpace(r.CategoryID)
	r.LocationID = strings.TrimSpace(r.LocationID)
	r.Code = strings.TrimSpace(r.Code)
	r.Name = strings.TrimSpace(r.Name)
	r.Condition = strings.TrimSpace(r.Condition)
	if r.CategoryID == "" || r.LocationID == "" || r.Code == "" || r.Name == "" {
		return ErrValidation
	}
	switch r.Condition {
	case "good", "fair", "poor", "broken":
	default:
		return ErrValidation
	}
	return nil
}

type UpdateAssetRequest struct {
	CategoryID       *string  `json:"category_id,omitempty"`
	LocationID       *string  `json:"location_id,omitempty"`
	Name             *string  `json:"name,omitempty"`
	Description      *string  `json:"description,omitempty"`
	Condition        *string  `json:"condition,omitempty"`
	Status           *string  `json:"status,omitempty"`
	AcquisitionDate  *string  `json:"acquisition_date,omitempty"`
	AcquisitionValue *float64 `json:"acquisition_value,omitempty"`
	PICID            *string  `json:"pic_id,omitempty"`
	FileObjectID     *string  `json:"file_object_id,omitempty"`
}

func (r *UpdateAssetRequest) Validate() error {
	if r.Condition != nil {
		switch *r.Condition {
		case "good", "fair", "poor", "broken":
		default:
			return ErrValidation
		}
	}
	if r.Status != nil {
		switch *r.Status {
		case "available", "borrowed", "maintenance", "inactive", "disposed":
		default:
			return ErrValidation
		}
	}
	return nil
}


type AssetLoan struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	AssetID        string    `json:"asset_id"`
	AssetName      string    `json:"asset_name,omitempty"`
	BorrowerID     string    `json:"borrower_id"`
	BorrowerName   string    `json:"borrower_name,omitempty"`
	ApproverID     *string   `json:"approver_id,omitempty"`
	ApproverName   *string   `json:"approver_name,omitempty"`
	LoanDate       string    `json:"loan_date"`
	DueDate        string    `json:"due_date"`
	ReturnDate     *string   `json:"return_date,omitempty"`
	ConditionOut   string    `json:"condition_out"`
	ConditionIn    *string   `json:"condition_in,omitempty"`
	Status         string    `json:"status"` // pending, approved, rejected, returned
	Notes          *string   `json:"notes,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateLoanRequest struct {
	AssetID      string  `json:"asset_id"`
	LoanDate     string  `json:"loan_date"`
	DueDate      string  `json:"due_date"`
	ConditionOut string  `json:"condition_out"`
	Notes        *string `json:"notes,omitempty"`
}

func (r *CreateLoanRequest) Validate() error {
	r.AssetID = strings.TrimSpace(r.AssetID)
	r.ConditionOut = strings.TrimSpace(r.ConditionOut)
	if r.AssetID == "" || r.LoanDate == "" || r.DueDate == "" || r.ConditionOut == "" {
		return ErrValidation
	}
	return nil
}

type ReviewLoanRequest struct {
	Action string  `json:"action"` // approve, reject
	Notes  *string `json:"notes,omitempty"`
}

type ReturnLoanRequest struct {
	ConditionIn string  `json:"condition_in"`
	Notes       *string `json:"notes,omitempty"`
}

type AssetMaintenanceLog struct {
	ID              string    `json:"id"`
	OrganizationID  string    `json:"organization_id"`
	AssetID         string    `json:"asset_id"`
	AssetName       string    `json:"asset_name,omitempty"`
	MaintenanceDate string    `json:"maintenance_date"`
	MaintenanceType string    `json:"maintenance_type"`
	Cost            *float64  `json:"cost,omitempty"`
	Technician      *string   `json:"technician,omitempty"`
	Notes           *string   `json:"notes,omitempty"`
	FileObjectID    *string   `json:"file_object_id,omitempty"`
	ConditionAfter  string    `json:"condition_after"`
	StatusAfter     string    `json:"status_after"`
	CreatedBy       string    `json:"created_by"`
	CreatedAt       time.Time `json:"created_at"`
}

type CreateMaintenanceLogRequest struct {
	AssetID         string   `json:"asset_id"`
	MaintenanceDate string   `json:"maintenance_date"`
	MaintenanceType string   `json:"maintenance_type"`
	Cost            *float64 `json:"cost,omitempty"`
	Technician      *string  `json:"technician,omitempty"`
	Notes           *string  `json:"notes,omitempty"`
	FileObjectID    *string  `json:"file_object_id,omitempty"`
	ConditionAfter  string   `json:"condition_after"`
	StatusAfter     string   `json:"status_after"`
}

func (r *CreateMaintenanceLogRequest) Validate() error {
	r.AssetID = strings.TrimSpace(r.AssetID)
	r.MaintenanceType = strings.TrimSpace(r.MaintenanceType)
	if r.AssetID == "" || r.MaintenanceDate == "" || r.MaintenanceType == "" || r.ConditionAfter == "" || r.StatusAfter == "" {
		return ErrValidation
	}
	return nil
}
