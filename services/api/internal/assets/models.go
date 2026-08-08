package assets

import (
	"time"

	"github.com/google/uuid"
)

type AssetCategory struct {
	ID             uuid.UUID `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	Code           string    `json:"code" db:"code"`
	Name           string    `json:"name" db:"name"`
	Status         string    `json:"status" db:"status"` 
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type AssetLocation struct {
	ID             uuid.UUID `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	Code           string    `json:"code" db:"code"`
	Name           string    `json:"name" db:"name"`
	Status         string    `json:"status" db:"status"` 
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type Asset struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	OrganizationID   uuid.UUID  `json:"organization_id" db:"organization_id"`
	CategoryID       uuid.UUID  `json:"category_id" db:"category_id"`
	LocationID       uuid.UUID  `json:"location_id" db:"location_id"`
	Code             string     `json:"code" db:"code"`
	Name             string     `json:"name" db:"name"`
	Description      *string    `json:"description,omitempty" db:"description"`
	Condition        string     `json:"condition" db:"condition"` 
	Status           string     `json:"status" db:"status"`       
	AcquisitionDate  *time.Time `json:"acquisition_date,omitempty" db:"acquisition_date"`
	AcquisitionValue *float64   `json:"acquisition_value,omitempty" db:"acquisition_value"`
	PICID            *uuid.UUID `json:"pic_id,omitempty" db:"pic_id"`
	FileObjectID     *uuid.UUID `json:"file_object_id,omitempty" db:"file_object_id"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

type AssetLoan struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	AssetID        uuid.UUID  `json:"asset_id" db:"asset_id"`
	BorrowerID     uuid.UUID  `json:"borrower_id" db:"borrower_id"`
	ApproverID     *uuid.UUID `json:"approver_id,omitempty" db:"approver_id"`
	LoanDate       time.Time  `json:"loan_date" db:"loan_date"`
	DueDate        time.Time  `json:"due_date" db:"due_date"`
	ReturnDate     *time.Time `json:"return_date,omitempty" db:"return_date"`
	ConditionOut   string     `json:"condition_out" db:"condition_out"`
	ConditionIn    *string    `json:"condition_in,omitempty" db:"condition_in"`
	Status         string     `json:"status" db:"status"` 
	Notes          *string    `json:"notes,omitempty" db:"notes"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

type AssetMaintenanceLog struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	AssetID        uuid.UUID  `json:"asset_id" db:"asset_id"`
	MaintenanceDate time.Time `json:"maintenance_date" db:"maintenance_date"`
	MaintenanceType string    `json:"maintenance_type" db:"maintenance_type"`
	Cost           *float64   `json:"cost,omitempty" db:"cost"`
	Technician     *string    `json:"technician,omitempty" db:"technician"`
	Notes          *string    `json:"notes,omitempty" db:"notes"`
	FileObjectID   *uuid.UUID `json:"file_object_id,omitempty" db:"file_object_id"`
	ConditionAfter string     `json:"condition_after" db:"condition_after"`
	StatusAfter    string     `json:"status_after" db:"status_after"`
	CreatedBy      uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}