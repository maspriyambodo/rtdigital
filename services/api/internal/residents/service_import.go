package residents

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

type ImportResult struct {
	TotalRows     int      `json:"total_rows"`
	ValidRows     int      `json:"valid_rows"`
	InvalidRows   int      `json:"invalid_rows"`
	DuplicateRows int      `json:"duplicate_rows"`
	ImportedRows  int      `json:"imported_rows"`
	Errors        []string `json:"errors"`
}

type importResidentRow struct {
	line int
	req  CreateResidentRequest
}

func (s *Service) ImportCSV(ctx context.Context, principal *auth.Principal, source io.Reader, dryRun bool) (ImportResult, error) {
	reader := csv.NewReader(source)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return ImportResult{}, fmt.Errorf("%w: invalid CSV", ErrValidation)
	}
	if len(records) < 2 {
		return ImportResult{}, fmt.Errorf("%w: CSV must contain a header and at least one row", ErrValidation)
	}
	headers := importHeaderIndexes(records[0])
	if headers["full_name"] < 0 || headers["resident_status"] < 0 {
		return ImportResult{}, fmt.Errorf("%w: required headers are full_name,resident_status", ErrValidation)
	}

	result := ImportResult{TotalRows: len(records) - 1}
	rows := make([]importResidentRow, 0, result.TotalRows)
	seenNIK := map[string]int{}

	for index, record := range records[1:] {
		line := index + 2
		value := func(name string) string {
			position := headers[name]
			if position < 0 || position >= len(record) {
				return ""
			}
			return strings.TrimSpace(record[position])
		}
		req := CreateResidentRequest{
			FullName:       value("full_name"),
			ResidentStatus: value("resident_status"),
		}
		if nationalID := value("national_id"); nationalID != "" {
			req.NationalID = &nationalID
		}
		if phone := value("phone"); phone != "" {
			req.Phone = &phone
		}
		if educationCode := value("education_level_code"); educationCode != "" {
			req.EducationLevelID = &educationCode
		}
		if maritalCode := value("marital_status_code"); maritalCode != "" {
			req.MaritalStatusID = &maritalCode
		}
		if req.FullName == "" || !validResidentStatus(req.ResidentStatus) {
			result.InvalidRows++
			result.Errors = append(result.Errors, fmt.Sprintf("Baris %d: nama atau status warga tidak valid.", line))
			continue
		}
		if req.NationalID != nil {
			index := auth.GenerateBlindIndex(*req.NationalID, s.blindKey)
			if previous, duplicate := seenNIK[index]; duplicate {
				result.DuplicateRows++
				result.Errors = append(result.Errors, fmt.Sprintf("Baris %d: NIK duplikat dengan baris %d.", line, previous))
				continue
			}
			seenNIK[index] = line
		}
		rows = append(rows, importResidentRow{line: line, req: req})
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return result, fmt.Errorf("begin resident import: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	educationIDs, maritalIDs, err := importLookupIDs(ctx, tx)
	if err != nil {
		return result, err
	}
	validRows := rows[:0]
	for _, row := range rows {
		if row.req.EducationLevelID != nil {
			id, exists := educationIDs[strings.ToLower(*row.req.EducationLevelID)]
			if !exists {
				result.InvalidRows++
				result.Errors = append(result.Errors, fmt.Sprintf("Baris %d: kode pendidikan tidak valid.", row.line))
				continue
			}
			row.req.EducationLevelID = &id
		}
		if row.req.MaritalStatusID != nil {
			id, exists := maritalIDs[strings.ToLower(*row.req.MaritalStatusID)]
			if !exists {
				result.InvalidRows++
				result.Errors = append(result.Errors, fmt.Sprintf("Baris %d: kode status perkawinan tidak valid.", row.line))
				continue
			}
			row.req.MaritalStatusID = &id
		}
		if row.req.NationalID == nil {
			validRows = append(validRows, row)
			continue
		}
		blindIndex := auth.GenerateBlindIndex(*row.req.NationalID, s.blindKey)
		var exists bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM residents
				WHERE organization_id = $1 AND national_id_blind_index = $2
			)`, principal.OrganizationID, blindIndex).Scan(&exists); err != nil {
			return result, fmt.Errorf("check resident duplicate: %w", err)
		}
		if exists {
			result.DuplicateRows++
			result.Errors = append(result.Errors, fmt.Sprintf("Baris %d: NIK sudah terdaftar.", row.line))
			continue
		}
		validRows = append(validRows, row)
	}

	result.ValidRows = len(validRows)
	if dryRun || result.InvalidRows > 0 || result.DuplicateRows > 0 {
		return result, nil
	}

	for _, row := range validRows {
		encryptedID, blindIndex, err := s.encryptIndexedValue(row.req.NationalID)
		if err != nil {
			return result, err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO residents (
				id, organization_id, national_id_encrypted, national_id_blind_index,
				full_name, phone, education_level_id, marital_status_id,
				resident_status, verification_status
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'verified')`,
			newUUID(), principal.OrganizationID, encryptedID, blindIndex,
			row.req.FullName, row.req.Phone, row.req.EducationLevelID,
			row.req.MaritalStatusID, row.req.ResidentStatus,
		); err != nil {
			return result, mapDatabaseError(err, "import resident")
		}
		result.ImportedRows++
	}
	if err := s.audit(ctx, tx, principal, "resident.import", "organizations", principal.OrganizationID); err != nil {
		return result, err
	}
	if err := tx.Commit(ctx); err != nil {
		return result, fmt.Errorf("commit resident import: %w", err)
	}
	return result, nil
}

func importHeaderIndexes(header []string) map[string]int {
	indexes := map[string]int{
		"full_name": -1, "resident_status": -1, "national_id": -1, "phone": -1,
		"education_level_code": -1, "marital_status_code": -1,
	}
	for index, value := range header {
		if _, supported := indexes[strings.ToLower(strings.TrimSpace(value))]; supported {
			indexes[strings.ToLower(strings.TrimSpace(value))] = index
		}
	}
	return indexes
}

func importLookupIDs(ctx context.Context, tx pgx.Tx) (map[string]string, map[string]string, error) {
	load := func(table string) (map[string]string, error) {
		rows, err := tx.Query(ctx, `SELECT id, code FROM `+table)
		if err != nil {
			return nil, fmt.Errorf("load import lookup: %w", err)
		}
		defer rows.Close()
		items := map[string]string{}
		for rows.Next() {
			var id, code string
			if err := rows.Scan(&id, &code); err != nil {
				return nil, fmt.Errorf("scan import lookup: %w", err)
			}
			items[strings.ToLower(code)] = id
		}
		return items, rows.Err()
	}
	education, err := load("education_levels")
	if err != nil {
		return nil, nil, err
	}
	marital, err := load("marital_statuses")
	if err != nil {
		return nil, nil, err
	}
	return education, marital, nil
}
