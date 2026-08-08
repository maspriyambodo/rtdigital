package assets

import (
	"context"
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

func (s *Service) ListCategories(ctx context.Context, principal *auth.Principal) ([]AssetCategory, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, organization_id, code, name, status, created_at, updated_at
		FROM asset_categories
		WHERE organization_id = $1
		ORDER BY name ASC`, principal.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("query categories: %w", err)
	}
	defer rows.Close()

	items := []AssetCategory{}
	for rows.Next() {
		var item AssetCategory
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.Code, &item.Name, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) CreateCategory(ctx context.Context, principal *auth.Principal, req CreateCategoryRequest) (*AssetCategory, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var item AssetCategory
	err := s.db.QueryRow(ctx, `
		INSERT INTO asset_categories (organization_id, code, name, status)
		VALUES ($1, $2, $3, 'active')
		RETURNING id, organization_id, code, name, status, created_at, updated_at`,
		principal.OrganizationID, req.Code, req.Name,
	).Scan(&item.ID, &item.OrganizationID, &item.Code, &item.Name, &item.Status, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		if isDuplicate(err) {
			return nil, ErrDuplicateCode
		}
		return nil, fmt.Errorf("create category: %w", err)
	}

	s.audit(ctx, principal, "asset_category.create", "asset_categories", item.ID)
	return &item, nil
}

func (s *Service) UpdateCategory(ctx context.Context, principal *auth.Principal, id string, req UpdateCategoryRequest) (*AssetCategory, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var item AssetCategory
	err := s.db.QueryRow(ctx, `
		UPDATE asset_categories
		SET name = COALESCE($1, name),
		    status = COALESCE($2, status),
		    updated_at = NOW()
		WHERE id = $3 AND organization_id = $4
		RETURNING id, organization_id, code, name, status, created_at, updated_at`,
		req.Name, req.Status, id, principal.OrganizationID,
	).Scan(&item.ID, &item.OrganizationID, &item.Code, &item.Name, &item.Status, &item.CreatedAt, &item.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCategoryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update category: %w", err)
	}

	s.audit(ctx, principal, "asset_category.update", "asset_categories", item.ID)
	return &item, nil
}

func (s *Service) ListLocations(ctx context.Context, principal *auth.Principal) ([]AssetLocation, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, organization_id, code, name, status, created_at, updated_at
		FROM asset_locations
		WHERE organization_id = $1
		ORDER BY name ASC`, principal.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("query locations: %w", err)
	}
	defer rows.Close()

	items := []AssetLocation{}
	for rows.Next() {
		var item AssetLocation
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.Code, &item.Name, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan location: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) CreateLocation(ctx context.Context, principal *auth.Principal, req CreateLocationRequest) (*AssetLocation, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var item AssetLocation
	err := s.db.QueryRow(ctx, `
		INSERT INTO asset_locations (organization_id, code, name, status)
		VALUES ($1, $2, $3, 'active')
		RETURNING id, organization_id, code, name, status, created_at, updated_at`,
		principal.OrganizationID, req.Code, req.Name,
	).Scan(&item.ID, &item.OrganizationID, &item.Code, &item.Name, &item.Status, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		if isDuplicate(err) {
			return nil, ErrDuplicateCode
		}
		return nil, fmt.Errorf("create location: %w", err)
	}

	s.audit(ctx, principal, "asset_location.create", "asset_locations", item.ID)
	return &item, nil
}

func (s *Service) UpdateLocation(ctx context.Context, principal *auth.Principal, id string, req UpdateLocationRequest) (*AssetLocation, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var item AssetLocation
	err := s.db.QueryRow(ctx, `
		UPDATE asset_locations
		SET name = COALESCE($1, name),
		    status = COALESCE($2, status),
		    updated_at = NOW()
		WHERE id = $3 AND organization_id = $4
		RETURNING id, organization_id, code, name, status, created_at, updated_at`,
		req.Name, req.Status, id, principal.OrganizationID,
	).Scan(&item.ID, &item.OrganizationID, &item.Code, &item.Name, &item.Status, &item.CreatedAt, &item.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrLocationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update location: %w", err)
	}

	s.audit(ctx, principal, "asset_location.update", "asset_locations", item.ID)
	return &item, nil
}

func (s *Service) ListAssets(ctx context.Context, principal *auth.Principal, categoryID, locationID, status string) ([]Asset, error) {
	query := `
		SELECT a.id, a.organization_id, a.category_id, c.name, a.location_id, l.name,
		       a.code, a.name, a.description, a.condition, a.status,
		       a.acquisition_date, a.acquisition_value, a.pic_id, u.email,
		       a.file_object_id, a.created_at, a.updated_at
		FROM assets a
		JOIN asset_categories c ON c.id = a.category_id
		JOIN asset_locations l ON l.id = a.location_id
		LEFT JOIN users u ON u.id = a.pic_id
		WHERE a.organization_id = $1`
	args := []any{principal.OrganizationID}

	if categoryID != "" {
		args = append(args, categoryID)
		query += fmt.Sprintf(" AND a.category_id = $%d", len(args))
	}
	if locationID != "" {
		args = append(args, locationID)
		query += fmt.Sprintf(" AND a.location_id = $%d", len(args))
	}
	if status != "" {
		args = append(args, status)
		query += fmt.Sprintf(" AND a.status = $%d", len(args))
	}
	query += " ORDER BY a.name ASC"

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query assets: %w", err)
	}
	defer rows.Close()

	items := []Asset{}
	for rows.Next() {
		var item Asset
		var acqDate *time.Time
		if err := rows.Scan(
			&item.ID, &item.OrganizationID, &item.CategoryID, &item.CategoryName,
			&item.LocationID, &item.LocationName, &item.Code, &item.Name,
			&item.Description, &item.Condition, &item.Status, &acqDate,
			&item.AcquisitionValue, &item.PICID, &item.PICName,
			&item.FileObjectID, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan asset: %w", err)
		}
		if acqDate != nil {
			str := acqDate.Format("2006-01-02")
			item.AcquisitionDate = &str
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) GetAsset(ctx context.Context, principal *auth.Principal, id string) (*Asset, error) {
	var item Asset
	var acqDate *time.Time
	err := s.db.QueryRow(ctx, `
		SELECT a.id, a.organization_id, a.category_id, c.name, a.location_id, l.name,
		       a.code, a.name, a.description, a.condition, a.status,
		       a.acquisition_date, a.acquisition_value, a.pic_id, u.email,
		       a.file_object_id, a.created_at, a.updated_at
		FROM assets a
		JOIN asset_categories c ON c.id = a.category_id
		JOIN asset_locations l ON l.id = a.location_id
		LEFT JOIN users u ON u.id = a.pic_id
		WHERE a.id = $1 AND a.organization_id = $2`, id, principal.OrganizationID,
	).Scan(
		&item.ID, &item.OrganizationID, &item.CategoryID, &item.CategoryName,
		&item.LocationID, &item.LocationName, &item.Code, &item.Name,
		&item.Description, &item.Condition, &item.Status, &acqDate,
		&item.AcquisitionValue, &item.PICID, &item.PICName,
		&item.FileObjectID, &item.CreatedAt, &item.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAssetNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get asset: %w", err)
	}
	if acqDate != nil {
		str := acqDate.Format("2006-01-02")
		item.AcquisitionDate = &str
	}
	return &item, nil
}

func (s *Service) CreateAsset(ctx context.Context, principal *auth.Principal, req CreateAssetRequest) (*Asset, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var acqDate *time.Time
	if req.AcquisitionDate != nil && *req.AcquisitionDate != "" {
		t, err := time.Parse("2006-01-02", *req.AcquisitionDate)
		if err != nil {
			return nil, ErrValidation
		}
		acqDate = &t
	}

	var item Asset
	err := s.db.QueryRow(ctx, `
		INSERT INTO assets (
			organization_id, category_id, location_id, code, name, description,
			condition, status, acquisition_date, acquisition_value, pic_id, file_object_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, 'available', $8, $9, $10, $11)
		RETURNING id, organization_id, category_id, location_id, code, name, description,
		          condition, status, acquisition_date, acquisition_value, pic_id, file_object_id, created_at, updated_at`,
		principal.OrganizationID, req.CategoryID, req.LocationID, req.Code, req.Name,
		req.Description, req.Condition, acqDate, req.AcquisitionValue, req.PICID, req.FileObjectID,
	).Scan(
		&item.ID, &item.OrganizationID, &item.CategoryID, &item.LocationID, &item.Code, &item.Name,
		&item.Description, &item.Condition, &item.Status, &acqDate, &item.AcquisitionValue,
		&item.PICID, &item.FileObjectID, &item.CreatedAt, &item.UpdatedAt,
	)

	if err != nil {
		if isDuplicate(err) {
			return nil, ErrDuplicateCode
		}
		return nil, fmt.Errorf("create asset: %w", err)
	}

	if acqDate != nil {
		str := acqDate.Format("2006-01-02")
		item.AcquisitionDate = &str
	}
	s.audit(ctx, principal, "asset.create", "assets", item.ID)
	return &item, nil
}

func (s *Service) UpdateAsset(ctx context.Context, principal *auth.Principal, id string, req UpdateAssetRequest) (*Asset, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var acqDate *time.Time
	if req.AcquisitionDate != nil && *req.AcquisitionDate != "" {
		t, err := time.Parse("2006-01-02", *req.AcquisitionDate)
		if err != nil {
			return nil, ErrValidation
		}
		acqDate = &t
	}

	var item Asset
	err := s.db.QueryRow(ctx, `
		UPDATE assets
		SET category_id = COALESCE($1, category_id),
		    location_id = COALESCE($2, location_id),
		    name = COALESCE($3, name),
		    description = COALESCE($4, description),
		    condition = COALESCE($5, condition),
		    status = COALESCE($6, status),
		    acquisition_date = COALESCE($7, acquisition_date),
		    acquisition_value = COALESCE($8, acquisition_value),
		    pic_id = COALESCE($9, pic_id),
		    file_object_id = COALESCE($10, file_object_id),
		    updated_at = NOW()
		WHERE id = $11 AND organization_id = $12
		RETURNING id, organization_id, category_id, location_id, code, name, description,
		          condition, status, acquisition_date, acquisition_value, pic_id, file_object_id, created_at, updated_at`,
		req.CategoryID, req.LocationID, req.Name, req.Description, req.Condition,
		req.Status, acqDate, req.AcquisitionValue, req.PICID, req.FileObjectID, id, principal.OrganizationID,
	).Scan(
		&item.ID, &item.OrganizationID, &item.CategoryID, &item.LocationID, &item.Code, &item.Name,
		&item.Description, &item.Condition, &item.Status, &acqDate, &item.AcquisitionValue,
		&item.PICID, &item.FileObjectID, &item.CreatedAt, &item.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAssetNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update asset: %w", err)
	}

	if acqDate != nil {
		str := acqDate.Format("2006-01-02")
		item.AcquisitionDate = &str
	}
	s.audit(ctx, principal, "asset.update", "assets", item.ID)
	return &item, nil
}

func (s *Service) CreateLoan(ctx context.Context, principal *auth.Principal, req CreateLoanRequest) (*AssetLoan, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	loanDate, err1 := time.Parse("2006-01-02", req.LoanDate)
	dueDate, err2 := time.Parse("2006-01-02", req.DueDate)
	if err1 != nil || err2 != nil || dueDate.Before(loanDate) {
		return nil, ErrValidation
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var assetStatus string
	err = tx.QueryRow(ctx, `
		SELECT status FROM assets
		WHERE id = $1 AND organization_id = $2 FOR UPDATE`,
		req.AssetID, principal.OrganizationID,
	).Scan(&assetStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAssetNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("check asset status: %w", err)
	}
	if assetStatus != "available" {
		return nil, ErrInvalidState
	}

	var item AssetLoan
	err = tx.QueryRow(ctx, `
		INSERT INTO asset_loans (
			organization_id, asset_id, borrower_id, loan_date, due_date, condition_out, status, notes
		) VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7)
		RETURNING id, organization_id, asset_id, borrower_id, approver_id, loan_date, due_date,
		          return_date, condition_out, condition_in, status, notes, created_at, updated_at`,
		principal.OrganizationID, req.AssetID, principal.UserID, loanDate, dueDate, req.ConditionOut, req.Notes,
	).Scan(
		&item.ID, &item.OrganizationID, &item.AssetID, &item.BorrowerID, &item.ApproverID,
		&loanDate, &dueDate, &item.ReturnDate, &item.ConditionOut, &item.ConditionIn,
		&item.Status, &item.Notes, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create loan: %w", err)
	}

	item.LoanDate = loanDate.Format("2006-01-02")
	item.DueDate = dueDate.Format("2006-01-02")

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit loan: %w", err)
	}

	s.audit(ctx, principal, "asset_loan.create", "asset_loans", item.ID)
	return &item, nil
}

func (s *Service) ReviewLoan(ctx context.Context, principal *auth.Principal, id string, req ReviewLoanRequest) (*AssetLoan, error) {
	if req.Action != "approve" && req.Action != "reject" {
		return nil, ErrValidation
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var loan AssetLoan
	var loanDate, dueDate time.Time
	err = tx.QueryRow(ctx, `
		SELECT id, asset_id, status FROM asset_loans
		WHERE id = $1 AND organization_id = $2 FOR UPDATE`,
		id, principal.OrganizationID,
	).Scan(&loan.ID, &loan.AssetID, &loan.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrLoanNotFound
	}
	if err != nil || loan.Status != "pending" {
		return nil, ErrInvalidState
	}

	newStatus := "approved"
	if req.Action == "reject" {
		newStatus = "rejected"
	}

	err = tx.QueryRow(ctx, `
		UPDATE asset_loans
		SET status = $1, approver_id = $2, notes = COALESCE($3, notes), updated_at = NOW()
		WHERE id = $4
		RETURNING id, organization_id, asset_id, borrower_id, approver_id, loan_date, due_date,
		          return_date, condition_out, condition_in, status, notes, created_at, updated_at`,
		newStatus, principal.UserID, req.Notes, id,
	).Scan(
		&loan.ID, &loan.OrganizationID, &loan.AssetID, &loan.BorrowerID, &loan.ApproverID,
		&loanDate, &dueDate, &loan.ReturnDate, &loan.ConditionOut, &loan.ConditionIn,
		&loan.Status, &loan.Notes, &loan.CreatedAt, &loan.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update loan: %w", err)
	}

	if req.Action == "approve" {
		if _, err := tx.Exec(ctx, `UPDATE assets SET status = 'borrowed', updated_at = NOW() WHERE id = $1`, loan.AssetID); err != nil {
			return nil, fmt.Errorf("update asset borrowed: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit review loan: %w", err)
	}

	loan.LoanDate = loanDate.Format("2006-01-02")
	loan.DueDate = dueDate.Format("2006-01-02")
	s.audit(ctx, principal, "asset_loan.review", "asset_loans", loan.ID)
	return &loan, nil
}

func (s *Service) ReturnLoan(ctx context.Context, principal *auth.Principal, id string, req ReturnLoanRequest) (*AssetLoan, error) {
	if strings.TrimSpace(req.ConditionIn) == "" {
		return nil, ErrValidation
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var loan AssetLoan
	var loanDate, dueDate time.Time
	var returnDate *time.Time
	err = tx.QueryRow(ctx, `
		SELECT id, asset_id, status FROM asset_loans
		WHERE id = $1 AND organization_id = $2 FOR UPDATE`,
		id, principal.OrganizationID,
	).Scan(&loan.ID, &loan.AssetID, &loan.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrLoanNotFound
	}
	if err != nil || loan.Status != "approved" {
		return nil, ErrInvalidState
	}

	now := time.Now().UTC()
	err = tx.QueryRow(ctx, `
		UPDATE asset_loans
		SET status = 'returned', return_date = $1, condition_in = $2, notes = COALESCE($3, notes), updated_at = NOW()
		WHERE id = $4
		RETURNING id, organization_id, asset_id, borrower_id, approver_id, loan_date, due_date,
		          return_date, condition_out, condition_in, status, notes, created_at, updated_at`,
		now, req.ConditionIn, req.Notes, id,
	).Scan(
		&loan.ID, &loan.OrganizationID, &loan.AssetID, &loan.BorrowerID, &loan.ApproverID,
		&loanDate, &dueDate, &returnDate, &loan.ConditionOut, &loan.ConditionIn,
		&loan.Status, &loan.Notes, &loan.CreatedAt, &loan.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update loan returned: %w", err)
	}

	if _, err := tx.Exec(ctx, `UPDATE assets SET status = 'available', condition = $1, updated_at = NOW() WHERE id = $2`, req.ConditionIn, loan.AssetID); err != nil {
		return nil, fmt.Errorf("update asset returned: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit return loan: %w", err)
	}

	loan.LoanDate = loanDate.Format("2006-01-02")
	loan.DueDate = dueDate.Format("2006-01-02")
	if returnDate != nil {
		str := returnDate.Format("2006-01-02")
		loan.ReturnDate = &str
	}
	s.audit(ctx, principal, "asset_loan.return", "asset_loans", loan.ID)
	return &loan, nil
}

func (s *Service) ListLoans(ctx context.Context, principal *auth.Principal, assetID, status string) ([]AssetLoan, error) {
	query := `
		SELECT l.id, l.organization_id, l.asset_id, a.name, l.borrower_id, u1.email,
		       l.approver_id, COALESCE(u2.email, ''), l.loan_date, l.due_date, l.return_date,
		       l.condition_out, COALESCE(l.condition_in, ''), l.status, COALESCE(l.notes, ''), l.created_at, l.updated_at
		FROM asset_loans l
		JOIN assets a ON a.id = l.asset_id
		JOIN users u1 ON u1.id = l.borrower_id
		LEFT JOIN users u2 ON u2.id = l.approver_id
		WHERE l.organization_id = $1`
	args := []any{principal.OrganizationID}

	if assetID != "" {
		args = append(args, assetID)
		query += fmt.Sprintf(" AND l.asset_id = $%d", len(args))
	}
	if status != "" {
		args = append(args, status)
		query += fmt.Sprintf(" AND l.status = $%d", len(args))
	}
	query += " ORDER BY l.created_at DESC"

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query loans: %w", err)
	}
	defer rows.Close()

	items := []AssetLoan{}
	for rows.Next() {
		var item AssetLoan
		var lDate, dDate time.Time
		var rDate *time.Time
		if err := rows.Scan(
			&item.ID, &item.OrganizationID, &item.AssetID, &item.AssetName,
			&item.BorrowerID, &item.BorrowerName, &item.ApproverID, &item.ApproverName,
			&lDate, &dDate, &rDate, &item.ConditionOut, &item.ConditionIn,
			&item.Status, &item.Notes, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan loan: %w", err)
		}
		item.LoanDate = lDate.Format("2006-01-02")
		item.DueDate = dDate.Format("2006-01-02")
		if rDate != nil {
			str := rDate.Format("2006-01-02")
			item.ReturnDate = &str
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) CreateMaintenanceLog(ctx context.Context, principal *auth.Principal, req CreateMaintenanceLogRequest) (*AssetMaintenanceLog, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	mDate, err := time.Parse("2006-01-02", req.MaintenanceDate)
	if err != nil {
		return nil, ErrValidation
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var log AssetMaintenanceLog
	err = tx.QueryRow(ctx, `
		INSERT INTO asset_maintenance_logs (
			organization_id, asset_id, maintenance_date, maintenance_type, cost,
			technician, notes, file_object_id, condition_after, status_after, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, organization_id, asset_id, maintenance_date, maintenance_type, cost,
		          technician, notes, file_object_id, condition_after, status_after, created_by, created_at`,
		principal.OrganizationID, req.AssetID, mDate, req.MaintenanceType, req.Cost,
		req.Technician, req.Notes, req.FileObjectID, req.ConditionAfter, req.StatusAfter, principal.UserID,
	).Scan(
		&log.ID, &log.OrganizationID, &log.AssetID, &mDate, &log.MaintenanceType,
		&log.Cost, &log.Technician, &log.Notes, &log.FileObjectID,
		&log.ConditionAfter, &log.StatusAfter, &log.CreatedBy, &log.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create maintenance log: %w", err)
	}

	if _, err := tx.Exec(ctx, `UPDATE assets SET condition = $1, status = $2, updated_at = NOW() WHERE id = $3`, req.ConditionAfter, req.StatusAfter, req.AssetID); err != nil {
		return nil, fmt.Errorf("update asset after maintenance: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit maintenance log: %w", err)
	}

	log.MaintenanceDate = mDate.Format("2006-01-02")
	s.audit(ctx, principal, "asset_maintenance.create", "asset_maintenance_logs", log.ID)
	return &log, nil
}

func (s *Service) ListMaintenanceLogs(ctx context.Context, principal *auth.Principal, assetID string) ([]AssetMaintenanceLog, error) {
	rows, err := s.db.Query(ctx, `
		SELECT m.id, m.organization_id, m.asset_id, a.name, m.maintenance_date, m.maintenance_type,
		       m.cost, COALESCE(m.technician, ''), COALESCE(m.notes, ''), m.file_object_id, m.condition_after, m.status_after,
		       m.created_by, m.created_at
		FROM asset_maintenance_logs m
		JOIN assets a ON a.id = m.asset_id
		WHERE m.organization_id = $1 AND ($2 = '' OR m.asset_id = $2)
		ORDER BY m.maintenance_date DESC`, principal.OrganizationID, assetID)
	if err != nil {
		return nil, fmt.Errorf("query maintenance logs: %w", err)
	}
	defer rows.Close()

	items := []AssetMaintenanceLog{}
	for rows.Next() {
		var item AssetMaintenanceLog
		var mDate time.Time
		if err := rows.Scan(
			&item.ID, &item.OrganizationID, &item.AssetID, &item.AssetName, &mDate,
			&item.MaintenanceType, &item.Cost, &item.Technician, &item.Notes,
			&item.FileObjectID, &item.ConditionAfter, &item.StatusAfter, &item.CreatedBy, &item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan maintenance log: %w", err)
		}
		item.MaintenanceDate = mDate.Format("2006-01-02")
		items = append(items, item)
	}
	return items, rows.Err()
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
