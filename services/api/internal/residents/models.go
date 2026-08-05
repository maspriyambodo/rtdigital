package residents

import "time"

type EducationLevel struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type MaritalStatus struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type HouseUnit struct {
	ID              string    `json:"id"`
	Code            string    `json:"code"`
	AddressDetail   *string   `json:"address_detail"`
	OccupancyStatus string    `json:"occupancy_status"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Resident struct {
	ID                 string    `json:"id"`
	NationalID         *string   `json:"national_id,omitempty"`
	FullName           string    `json:"full_name"`
	BirthPlace         *string   `json:"birth_place"`
	BirthDate          *string   `json:"birth_date"`
	Gender             *string   `json:"gender"`
	MaritalStatusID    *string   `json:"marital_status_id"`
	MaritalStatusName  *string   `json:"marital_status_name"`
	Occupation         *string   `json:"occupation"`
	EducationLevelID   *string   `json:"education_level_id"`
	EducationLevelName *string   `json:"education_level_name"`
	Phone              *string   `json:"phone"`
	Email              *string   `json:"email"`
	ResidentStatus     string    `json:"resident_status"`
	VerificationStatus string    `json:"verification_status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type Household struct {
	ID                      string            `json:"id"`
	HouseUnitID             string            `json:"house_unit_id"`
	InternalNumber          string            `json:"internal_number"`
	FamilyCardNumber        *string           `json:"family_card_number,omitempty"`
	HeadResidentID          *string           `json:"head_resident_id"`
	HeadResidentName        *string           `json:"head_resident_name,omitempty"`
	DomicileStatus          string            `json:"domicile_status"`
	MoveInDate              *string           `json:"move_in_date"`
	MoveOutDate             *string           `json:"move_out_date"`
	DomicileReviewDueAt     *string           `json:"domicile_review_due_at,omitempty"`
	DomicileLastConfirmedAt *time.Time        `json:"domicile_last_confirmed_at,omitempty"`
	VerificationStatus      string            `json:"verification_status"`
	Members                 []HouseholdMember `json:"members,omitempty"`
	CreatedAt               time.Time         `json:"created_at"`
	UpdatedAt               time.Time         `json:"updated_at"`
}

type HouseholdMember struct {
	ID           string  `json:"id"`
	ResidentID   string  `json:"resident_id"`
	ResidentName string  `json:"resident_name"`
	Relationship string  `json:"relationship"`
	IsActive     bool    `json:"is_active"`
	StartedAt    string  `json:"started_at"`
	EndedAt      *string `json:"ended_at"`
}

// HouseholdHealthScore is an operational completeness score, never a
// sanction or eligibility score. It deliberately excludes sensitive values.
type HouseholdHealthScore struct {
	HouseholdID    string    `json:"household_id"`
	InternalNumber string    `json:"internal_number"`
	Score          int       `json:"score"`
	MissingItems   []string  `json:"missing_items"`
	UpdatedAt      time.Time `json:"updated_at"`
	DomicileDueAt  *string   `json:"domicile_due_at,omitempty"`
}

type CreateHouseUnitRequest struct {
	Code            string  `json:"code"`
	AddressDetail   *string `json:"address_detail"`
	OccupancyStatus string  `json:"occupancy_status"`
}

type UpdateHouseUnitRequest struct {
	Code            *string `json:"code"`
	AddressDetail   *string `json:"address_detail"`
	OccupancyStatus *string `json:"occupancy_status"`
}

type CreateResidentRequest struct {
	NationalID       *string `json:"national_id"`
	FullName         string  `json:"full_name"`
	BirthPlace       *string `json:"birth_place"`
	BirthDate        *string `json:"birth_date"`
	Gender           *string `json:"gender"`
	MaritalStatusID  *string `json:"marital_status_id"`
	Occupation       *string `json:"occupation"`
	EducationLevelID *string `json:"education_level_id"`
	Phone            *string `json:"phone"`
	Email            *string `json:"email"`
	ResidentStatus   string  `json:"resident_status"`
}

type UpdateResidentRequest struct {
	NationalID       *string `json:"national_id"`
	FullName         *string `json:"full_name"`
	BirthPlace       *string `json:"birth_place"`
	BirthDate        *string `json:"birth_date"`
	Gender           *string `json:"gender"`
	MaritalStatusID  *string `json:"marital_status_id"`
	Occupation       *string `json:"occupation"`
	EducationLevelID *string `json:"education_level_id"`
	Phone            *string `json:"phone"`
	Email            *string `json:"email"`
	ResidentStatus   *string `json:"resident_status"`
}

type CreateHouseholdRequest struct {
	HouseUnitID      string  `json:"house_unit_id"`
	InternalNumber   string  `json:"internal_number"`
	FamilyCardNumber *string `json:"family_card_number"`
	DomicileStatus   string  `json:"domicile_status"`
	MoveInDate       *string `json:"move_in_date"`
}

type UpdateHouseholdRequest struct {
	HouseUnitID      *string `json:"house_unit_id"`
	InternalNumber   *string `json:"internal_number"`
	FamilyCardNumber *string `json:"family_card_number"`
	DomicileStatus   *string `json:"domicile_status"`
	MoveInDate       *string `json:"move_in_date"`
	MoveOutDate      *string `json:"move_out_date"`
}

type HouseholdMemberRequest struct {
	ResidentID   string `json:"resident_id"`
	Relationship string `json:"relationship"`
}
