package settings

import (
	"encoding/json"
	"time"
)

type OrganizationSettings struct {
	ID                         string          `json:"id"`
	Name                       string          `json:"name"`
	RTNumber                   string          `json:"rt_number"`
	RWNumber                   string          `json:"rw_number"`
	Address                    *string         `json:"address"`
	Timezone                   string          `json:"timezone"`
	LogoFileID                 *string         `json:"logo_file_id"`
	BankName                   *string         `json:"bank_name"`
	BankAccountNumber          *string         `json:"bank_account_number"`
	BankAccountHolder          *string         `json:"bank_account_holder"`
	MaxUploadSizeBytes         int64           `json:"max_upload_size_bytes"`
	DefaultLetterNumberPattern string          `json:"default_letter_number_pattern"`
	Status                     string          `json:"status"`
	Settings                   json.RawMessage `json:"settings"`
	CreatedAt                  time.Time       `json:"created_at"`
	UpdatedAt                  time.Time       `json:"updated_at"`
}

type UpdateOrganizationSettingsRequest struct {
	Name                       *string          `json:"name"`
	RTNumber                   *string          `json:"rt_number"`
	RWNumber                   *string          `json:"rw_number"`
	Address                    *string          `json:"address"`
	Timezone                   *string          `json:"timezone"`
	LogoFileID                 *string          `json:"logo_file_id"`
	BankName                   *string          `json:"bank_name"`
	BankAccountNumber          *string          `json:"bank_account_number"`
	BankAccountHolder          *string          `json:"bank_account_holder"`
	MaxUploadSizeBytes         *int64           `json:"max_upload_size_bytes"`
	DefaultLetterNumberPattern *string          `json:"default_letter_number_pattern"`
	Settings                   *json.RawMessage `json:"settings"`
}