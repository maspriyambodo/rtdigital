package letters

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
	"github.com/maspriyambodo/rtdigital/services/api/internal/notifications"
)

type Service struct {
	db         *pgxpool.Pool
	dispatcher *notifications.Dispatcher
	now        func() time.Time
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) SetNotificationDispatcher(dispatcher *notifications.Dispatcher) {
	s.dispatcher = dispatcher
}

func (s *Service) notifySecretariesAndRT(organizationID string, job notifications.DispatchJob) {
	if s.dispatcher == nil {
		return
	}
	rows, err := s.db.Query(context.Background(), `
		SELECT DISTINCT u.id
		FROM users u
		JOIN user_roles ur ON ur.user_id = u.id
		JOIN roles r ON r.id = ur.role_id
		WHERE u.organization_id = $1
		  AND u.status = 'active'
		  AND r.code IN ('sekretaris', 'ketua_rt')`,
		organizationID,
	)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		if rows.Scan(&userID) == nil {
			job.OrganizationID = organizationID
			job.RecipientUserID = userID
			s.dispatcher.Dispatch(job)
		}
	}
}

func (s *Service) dispatchNotification(job notifications.DispatchJob) {
	if s.dispatcher != nil {
		s.dispatcher.Dispatch(job)
	}
}

func (s *Service) CreateLetterType(ctx context.Context, principal *auth.Principal, req CreateLetterTypeRequest) (LetterTypeItem, error) {
	if principal == nil {
		return LetterTypeItem{}, ErrForbidden
	}
	if err := req.Validate(); err != nil {
		return LetterTypeItem{}, err
	}
	id := newUUID()
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return LetterTypeItem{}, fmt.Errorf("begin letter type: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `
		INSERT INTO letter_types (id, organization_id, name, requirements, form_schema, template, number_pattern, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		id, principal.OrganizationID, req.Name, req.Requirements, req.FormSchema, req.Template, req.NumberPattern, req.Status,
	); err != nil {
		if isUniqueConstraintViolation(err) {
			return LetterTypeItem{}, ErrConflict
		}
		return LetterTypeItem{}, fmt.Errorf("insert letter type: %w", err)
	}
	if err := s.audit(ctx, tx, principal, "letter_type.create", "letter_type", id); err != nil {
		return LetterTypeItem{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return LetterTypeItem{}, fmt.Errorf("commit letter type: %w", err)
	}
	return s.GetLetterType(ctx, principal, id)
}

func (s *Service) ListLetterTypes(ctx context.Context, principal *auth.Principal, includeInactive bool) ([]LetterTypeItem, error) {
	if principal == nil {
		return nil, ErrForbidden
	}
	rows, err := s.db.Query(ctx, `
		SELECT id, name, requirements, form_schema, template, number_pattern, status, created_at, updated_at
		FROM letter_types
		WHERE organization_id = $1 AND ($2 OR status = 'active')
		ORDER BY name`,
		principal.OrganizationID, includeInactive,
	)
	if err != nil {
		return nil, fmt.Errorf("list letter types: %w", err)
	}
	defer rows.Close()
	items := make([]LetterTypeItem, 0)
	for rows.Next() {
		var item LetterTypeItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Requirements, &item.FormSchema, &item.Template, &item.NumberPattern, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan letter type: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) GetLetterType(ctx context.Context, principal *auth.Principal, id string) (LetterTypeItem, error) {
	if principal == nil || strings.TrimSpace(id) == "" {
		return LetterTypeItem{}, ErrValidation
	}
	var item LetterTypeItem
	err := s.db.QueryRow(ctx, `
		SELECT id, name, requirements, form_schema, template, number_pattern, status, created_at, updated_at
		FROM letter_types WHERE organization_id = $1 AND id = $2`,
		principal.OrganizationID, id,
	).Scan(&item.ID, &item.Name, &item.Requirements, &item.FormSchema, &item.Template, &item.NumberPattern, &item.Status, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return LetterTypeItem{}, ErrLetterTypeNotFound
	}
	if err != nil {
		return LetterTypeItem{}, fmt.Errorf("get letter type: %w", err)
	}
	return item, nil
}

func (s *Service) UpdateLetterType(ctx context.Context, principal *auth.Principal, id string, req UpdateLetterTypeRequest) (LetterTypeItem, error) {
	if principal == nil {
		return LetterTypeItem{}, ErrForbidden
	}
	if err := req.Validate(); err != nil {
		return LetterTypeItem{}, err
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return LetterTypeItem{}, fmt.Errorf("begin update letter type: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	result, err := tx.Exec(ctx, `
		UPDATE letter_types SET name = $1, requirements = $2, form_schema = $3, template = $4, number_pattern = $5, status = $6
		WHERE organization_id = $7 AND id = $8`,
		req.Name, req.Requirements, req.FormSchema, req.Template, req.NumberPattern, req.Status, principal.OrganizationID, id,
	)
	if err != nil {
		if isUniqueConstraintViolation(err) {
			return LetterTypeItem{}, ErrConflict
		}
		return LetterTypeItem{}, fmt.Errorf("update letter type: %w", err)
	}
	if result.RowsAffected() == 0 {
		return LetterTypeItem{}, ErrLetterTypeNotFound
	}
	if err := s.audit(ctx, tx, principal, "letter_type.update", "letter_type", id); err != nil {
		return LetterTypeItem{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return LetterTypeItem{}, fmt.Errorf("commit update letter type: %w", err)
	}
	return s.GetLetterType(ctx, principal, id)
}

func (s *Service) DeactivateLetterType(ctx context.Context, principal *auth.Principal, id string) (LetterTypeItem, error) {
	if principal == nil {
		return LetterTypeItem{}, ErrForbidden
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return LetterTypeItem{}, fmt.Errorf("begin deactivate letter type: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	result, err := tx.Exec(ctx, `UPDATE letter_types SET status = 'inactive' WHERE organization_id = $1 AND id = $2 AND status <> 'inactive'`, principal.OrganizationID, id)
	if err != nil {
		return LetterTypeItem{}, fmt.Errorf("deactivate letter type: %w", err)
	}
	if result.RowsAffected() == 0 {
		return LetterTypeItem{}, ErrLetterTypeNotFound
	}
	if err := s.audit(ctx, tx, principal, "letter_type.deactivate", "letter_type", id); err != nil {
		return LetterTypeItem{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return LetterTypeItem{}, fmt.Errorf("commit deactivate letter type: %w", err)
	}
	return s.GetLetterType(ctx, principal, id)
}

func (s *Service) SubmitLetterRequest(ctx context.Context, principal *auth.Principal, req SubmitLetterRequest) (LetterRequestItem, error) {
	if principal == nil {
		return LetterRequestItem{}, ErrForbidden
	}
	if err := req.Validate(); err != nil {
		return LetterRequestItem{}, err
	}
	if !principal.HasPermission("letter_request.process") {
		ok, err := s.ownsResident(ctx, principal, req.ResidentID)
		if err != nil || !ok {
			if err != nil {
				return LetterRequestItem{}, err
			}
			return LetterRequestItem{}, ErrForbidden
		}
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return LetterRequestItem{}, fmt.Errorf("begin submit letter: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var requirements json.RawMessage
	if err := tx.QueryRow(ctx, `SELECT requirements FROM letter_types WHERE organization_id = $1 AND id = $2 AND status = 'active'`, principal.OrganizationID, req.LetterTypeID).Scan(&requirements); errors.Is(err, pgx.ErrNoRows) {
		return LetterRequestItem{}, ErrLetterTypeNotFound
	} else if err != nil {
		return LetterRequestItem{}, fmt.Errorf("get letter type: %w", err)
	}
	if err := validateForm(req.FormData, requirements, len(req.AttachmentFileIDs)); err != nil {
		return LetterRequestItem{}, err
	}
	id, now := newUUID(), s.now()
	requestNumber := fmt.Sprintf("SR-%s-%s", now.Format("060102"), strings.ToUpper(id[:6]))
	if _, err := tx.Exec(ctx, `
		INSERT INTO letter_requests (id, organization_id, requester_user_id, resident_id, letter_type_id, request_number, form_data, status, resident_note, submitted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'submitted', $8, $9)`,
		id, principal.OrganizationID, principal.UserID, req.ResidentID, req.LetterTypeID, requestNumber, req.FormData, req.ResidentNote, now,
	); err != nil {
		return LetterRequestItem{}, fmt.Errorf("insert letter request: %w", err)
	}
	if err := s.syncAttachments(ctx, tx, principal, id, req.AttachmentFileIDs); err != nil {
		return LetterRequestItem{}, err
	}
	if err := s.audit(ctx, tx, principal, "letter_request.submit", "letter_request", id); err != nil {
		return LetterRequestItem{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return LetterRequestItem{}, fmt.Errorf("commit submit letter: %w", err)
	}
	return s.GetLetterRequest(ctx, principal, id)
}

func (s *Service) ListLetterRequests(ctx context.Context, principal *auth.Principal, filter LetterRequestFilter) ([]LetterRequestItem, error) {
	if principal == nil {
		return nil, ErrForbidden
	}
	manager := principal.HasPermission("letter_request.process") || principal.HasPermission("letter_request.approve")
	rows, err := s.db.Query(ctx, `
		SELECT id FROM letter_requests
		WHERE organization_id = $1
		  AND ($2 OR requester_user_id = $3 OR EXISTS (
			SELECT 1 FROM household_members mine JOIN users u ON u.organization_id = $1 AND u.resident_id = mine.resident_id
			JOIN household_members target ON target.household_id = mine.household_id AND target.is_active
			WHERE mine.is_active AND u.id = $3 AND target.resident_id = letter_requests.resident_id
		  ))
		  AND ($4 = '' OR status = $4)
		  AND ($5 = '' OR letter_type_id = $5)
		  AND ($6 = '' OR request_number ILIKE '%' || $6 || '%')
		ORDER BY created_at DESC LIMIT 100`,
		principal.OrganizationID, manager, principal.UserID, strings.TrimSpace(filter.Status), strings.TrimSpace(filter.LetterTypeID), strings.TrimSpace(filter.Search),
	)
	if err != nil {
		return nil, fmt.Errorf("list letter requests: %w", err)
	}
	defer rows.Close()
	items := make([]LetterRequestItem, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan letter request id: %w", err)
		}
		item, err := s.GetLetterRequest(ctx, principal, id)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) GetLetterRequest(ctx context.Context, principal *auth.Principal, id string) (LetterRequestItem, error) {
	if principal == nil {
		return LetterRequestItem{}, ErrForbidden
	}
	var item LetterRequestItem
	err := s.db.QueryRow(ctx, `
		SELECT lr.id, lr.requester_user_id, COALESCE(ru.email, ru.phone, 'Warga'), lr.resident_id, r.full_name,
		       lr.letter_type_id, lt.name, lr.request_number, lr.letter_number, lr.form_data, lr.status,
		       lr.resident_note, lr.internal_note, lr.submitted_at, lr.processed_by, lr.approved_by,
		       lr.approved_at, lr.issued_file_id, lr.issued_at, lr.created_at, lr.updated_at
		FROM letter_requests lr
		JOIN users ru ON ru.organization_id = lr.organization_id AND ru.id = lr.requester_user_id
		JOIN residents r ON r.organization_id = lr.organization_id AND r.id = lr.resident_id
		JOIN letter_types lt ON lt.organization_id = lr.organization_id AND lt.id = lr.letter_type_id
		WHERE lr.organization_id = $1 AND lr.id = $2`,
		principal.OrganizationID, id,
	).Scan(&item.ID, &item.RequesterUserID, &item.RequesterName, &item.ResidentID, &item.ResidentName,
		&item.LetterTypeID, &item.LetterTypeName, &item.RequestNumber, &item.LetterNumber, &item.FormData,
		&item.Status, &item.ResidentNote, &item.InternalNote, &item.SubmittedAt, &item.ProcessedBy,
		&item.ApprovedBy, &item.ApprovedAt, &item.IssuedFileID, &item.IssuedAt, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return LetterRequestItem{}, ErrLetterRequestNotFound
	}
	if err != nil {
		return LetterRequestItem{}, fmt.Errorf("get letter request: %w", err)
	}
	if !principal.HasPermission("letter_request.process") && !principal.HasPermission("letter_request.approve") && item.RequesterUserID != principal.UserID {
		ok, scopeErr := s.ownsResident(ctx, principal, item.ResidentID)
		if scopeErr != nil {
			return LetterRequestItem{}, scopeErr
		}
		if !ok {
			return LetterRequestItem{}, ErrForbidden
		}
	}
	item.Attachments, err = s.attachments(ctx, principal.OrganizationID, id)
	return item, err
}

func (s *Service) UpdateLetterRequest(ctx context.Context, principal *auth.Principal, id string, req UpdateLetterRequest) (LetterRequestItem, error) {
	if principal == nil {
		return LetterRequestItem{}, ErrForbidden
	}
	if err := req.Validate(); err != nil {
		return LetterRequestItem{}, err
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return LetterRequestItem{}, fmt.Errorf("begin update letter: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var requesterID, status string
	if err := tx.QueryRow(ctx, `
		SELECT requester_user_id, status FROM letter_requests
		WHERE organization_id = $1 AND id = $2 FOR UPDATE`,
		principal.OrganizationID, id,
	).Scan(&requesterID, &status); errors.Is(err, pgx.ErrNoRows) {
		return LetterRequestItem{}, ErrLetterRequestNotFound
	} else if err != nil {
		return LetterRequestItem{}, fmt.Errorf("lock letter request: %w", err)
	}
	if requesterID != principal.UserID || (status != "draft" && status != "needs_revision") {
		if requesterID != principal.UserID {
			return LetterRequestItem{}, ErrForbidden
		}
		return LetterRequestItem{}, ErrInvalidState
	}
	allowed, err := s.ownsResident(ctx, principal, req.ResidentID)
	if err != nil {
		return LetterRequestItem{}, err
	}
	if !allowed {
		return LetterRequestItem{}, ErrForbidden
	}

	var requirements json.RawMessage
	if err := tx.QueryRow(ctx, `
		SELECT requirements FROM letter_types
		WHERE organization_id = $1 AND id = $2 AND status = 'active'`,
		principal.OrganizationID, req.LetterTypeID,
	).Scan(&requirements); errors.Is(err, pgx.ErrNoRows) {
		return LetterRequestItem{}, ErrLetterTypeNotFound
	} else if err != nil {
		return LetterRequestItem{}, fmt.Errorf("get letter type: %w", err)
	}
	if err := validateForm(req.FormData, requirements, len(req.AttachmentFileIDs)); err != nil {
		return LetterRequestItem{}, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE letter_requests
		SET resident_id = $1, letter_type_id = $2, form_data = $3, resident_note = $4,
		    status = 'submitted', submitted_at = $5
		WHERE organization_id = $6 AND id = $7`,
		req.ResidentID, req.LetterTypeID, req.FormData, req.ResidentNote, s.now(),
		principal.OrganizationID, id,
	); err != nil {
		return LetterRequestItem{}, fmt.Errorf("update letter request: %w", err)
	}
	if err := s.syncAttachments(ctx, tx, principal, id, req.AttachmentFileIDs); err != nil {
		return LetterRequestItem{}, err
	}
	if err := s.audit(ctx, tx, principal, "letter_request.resubmit", "letter_request", id); err != nil {
		return LetterRequestItem{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return LetterRequestItem{}, fmt.Errorf("commit update letter: %w", err)
	}
	s.notifySecretariesAndRT(principal.OrganizationID, notifications.DispatchJob{
		Type:          "letter_request_submitted",
		Title:         "Pengajuan surat pengantar baru",
		Body:          "Pengajuan surat memerlukan pemeriksaan dan persetujuan.",
		ReferenceType: "letter_request",
		ReferenceID:   id,
	})
	return s.GetLetterRequest(ctx, principal, id)
}

func (s *Service) transitionLetterRequest(ctx context.Context, principal *auth.Principal, id, from, to, action string, req ReviewLetterRequest, requireResidentNote bool, approval bool) (LetterRequestItem, error) {
	if principal == nil {
		return LetterRequestItem{}, ErrForbidden
	}
	if err := req.Validate(requireResidentNote); err != nil {
		return LetterRequestItem{}, err
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return LetterRequestItem{}, fmt.Errorf("begin %s: %w", action, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var requesterID, status string
	if err := tx.QueryRow(ctx, `
		SELECT requester_user_id, status FROM letter_requests
		WHERE organization_id = $1 AND id = $2 FOR UPDATE`,
		principal.OrganizationID, id,
	).Scan(&requesterID, &status); errors.Is(err, pgx.ErrNoRows) {
		return LetterRequestItem{}, ErrLetterRequestNotFound
	} else if err != nil {
		return LetterRequestItem{}, fmt.Errorf("lock letter request: %w", err)
	}
	if !strings.Contains(","+from+",", ","+status+",") {
		return LetterRequestItem{}, ErrInvalidState
	}
	if approval && requesterID == principal.UserID {
		return LetterRequestItem{}, ErrForbidden
	}

	query := `UPDATE letter_requests SET status = $1, resident_note = $2, internal_note = $3, processed_by = $4`
	args := []any{to, req.ResidentNote, req.InternalNote, principal.UserID}
	if approval {
		query += `, approved_by = $5, approved_at = $6`
		args = append(args, principal.UserID, s.now())
	}
	query += ` WHERE organization_id = $` + fmt.Sprint(len(args)+1) + ` AND id = $` + fmt.Sprint(len(args)+2)
	args = append(args, principal.OrganizationID, id)
	if _, err := tx.Exec(ctx, query, args...); err != nil {
		return LetterRequestItem{}, fmt.Errorf("%s letter request: %w", action, err)
	}
	if err := s.audit(ctx, tx, principal, "letter_request."+action, "letter_request", id); err != nil {
		return LetterRequestItem{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return LetterRequestItem{}, fmt.Errorf("commit %s: %w", action, err)
	}
	var title, body string
	switch to {
	case "under_review":
		title = "Surat diproses"
		body = "Pengajuan surat pengantar Anda sedang diperiksa oleh pengurus."
	case "needs_revision":
		title = "Revisi pengajuan surat"
		if req.ResidentNote != nil {
			body = fmt.Sprintf("Pengajuan surat pengantar memerlukan revisi: %s", *req.ResidentNote)
		} else {
			body = "Pengajuan surat pengantar Anda memerlukan revisi."
		}
	case "approved":
		title = "Pengajuan surat disetujui"
		body = "Pengajuan surat pengantar Anda disetujui dan siap diterbitkan."
	case "rejected":
		title = "Pengajuan surat ditolak"
		if req.ResidentNote != nil {
			body = fmt.Sprintf("Pengajuan surat pengantar ditolak: %s", *req.ResidentNote)
		} else {
			body = "Pengajuan surat pengantar Anda ditolak."
		}
	}
	if title != "" {
		s.dispatchNotification(notifications.DispatchJob{
			OrganizationID:  principal.OrganizationID,
			RecipientUserID: requesterID,
			Type:            "letter_request_" + to,
			Title:           title,
			Body:            body,
			ReferenceType:   "letter_request",
			ReferenceID:     id,
		})
	}
	return s.GetLetterRequest(ctx, principal, id)
}

func (s *Service) ProcessLetterRequest(ctx context.Context, principal *auth.Principal, id string, req ReviewLetterRequest) (LetterRequestItem, error) {
	return s.transitionLetterRequest(ctx, principal, id, "submitted", "under_review", "process", req, false, false)
}

func (s *Service) RequestRevision(ctx context.Context, principal *auth.Principal, id string, req ReviewLetterRequest) (LetterRequestItem, error) {
	return s.transitionLetterRequest(ctx, principal, id, "submitted,under_review,awaiting_approval", "needs_revision", "request_revision", req, true, false)
}

func (s *Service) ApproveLetterRequest(ctx context.Context, principal *auth.Principal, id string, req ReviewLetterRequest) (LetterRequestItem, error) {
	return s.transitionLetterRequest(ctx, principal, id, "under_review,awaiting_approval", "approved", "approve", req, false, true)
}

func (s *Service) RejectLetterRequest(ctx context.Context, principal *auth.Principal, id string, req ReviewLetterRequest) (LetterRequestItem, error) {
	return s.transitionLetterRequest(ctx, principal, id, "submitted,under_review,awaiting_approval", "rejected", "reject", req, true, true)
}

func (s *Service) syncAttachments(ctx context.Context, tx pgx.Tx, principal *auth.Principal, requestID string, fileIDs []string) error {
	if _, err := tx.Exec(ctx, `DELETE FROM file_attachments WHERE organization_id = $1 AND entity_type = 'letter_request' AND entity_id = $2`, principal.OrganizationID, requestID); err != nil {
		return fmt.Errorf("clear letter attachments: %w", err)
	}
	seen := map[string]struct{}{}
	for _, fileID := range fileIDs {
		if _, duplicate := seen[fileID]; duplicate {
			return ErrValidation
		}
		seen[fileID] = struct{}{}
		var valid bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM file_objects WHERE id = $1 AND organization_id = $2 AND uploaded_by = $3
				AND confirmed_at IS NOT NULL AND deleted_at IS NULL)`,
			fileID, principal.OrganizationID, principal.UserID,
		).Scan(&valid); err != nil {
			return fmt.Errorf("validate letter attachment: %w", err)
		}
		if !valid {
			return ErrValidation
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO file_attachments (id, organization_id, file_id, entity_type, entity_id, purpose)
			VALUES ($1, $2, $3, 'letter_request', $4, 'attachment')`,
			newUUID(), principal.OrganizationID, fileID, requestID,
		); err != nil {
			return fmt.Errorf("attach letter file: %w", err)
		}
	}
	return nil
}

func (s *Service) attachments(ctx context.Context, organizationID, requestID string) ([]AttachmentInfo, error) {
	rows, err := s.db.Query(ctx, `
		SELECT a.id, a.file_id, f.original_name, f.mime_type, f.size_bytes, a.purpose
		FROM file_attachments a JOIN file_objects f ON f.organization_id = a.organization_id AND f.id = a.file_id
		WHERE a.organization_id = $1 AND a.entity_type = 'letter_request' AND a.entity_id = $2
		  AND f.confirmed_at IS NOT NULL AND f.deleted_at IS NULL ORDER BY a.created_at`,
		organizationID, requestID,
	)
	if err != nil {
		return nil, fmt.Errorf("list letter attachments: %w", err)
	}
	defer rows.Close()
	items := make([]AttachmentInfo, 0)
	for rows.Next() {
		var item AttachmentInfo
		if err := rows.Scan(&item.AttachmentID, &item.FileID, &item.OriginalName, &item.MIMEType, &item.SizeBytes, &item.Purpose); err != nil {
			return nil, fmt.Errorf("scan letter attachment: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) ownsResident(ctx context.Context, principal *auth.Principal, residentID string) (bool, error) {
	var allowed bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM users u
			LEFT JOIN household_members mine ON mine.resident_id = u.resident_id AND mine.is_active
			LEFT JOIN household_members target ON target.household_id = mine.household_id AND target.is_active
			WHERE u.organization_id = $1 AND u.id = $2 AND (u.resident_id = $3 OR target.resident_id = $3)
		)`, principal.OrganizationID, principal.UserID, residentID).Scan(&allowed)
	if err != nil {
		return false, fmt.Errorf("check letter resident scope: %w", err)
	}
	return allowed, nil
}

func validateForm(formData, requirements json.RawMessage, attachmentCount int) error {
	var data map[string]any
	if err := json.Unmarshal(formData, &data); err != nil {
		return ErrValidation
	}
	var requirementsList []struct {
		Required bool `json:"required"`
	}
	if err := json.Unmarshal(requirements, &requirementsList); err != nil {
		return ErrValidation
	}
	for _, requirement := range requirementsList {
		if requirement.Required && attachmentCount == 0 {
			return ErrValidation
		}
	}
	return nil
}

func (s *Service) audit(ctx context.Context, tx pgx.Tx, principal *auth.Principal, action, entityType, entityID string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (organization_id, actor_user_id, action, entity_type, entity_id)
		VALUES ($1, $2, $3, $4, $5)`,
		principal.OrganizationID, principal.UserID, action, entityType, entityID,
	)
	if err != nil {
		return fmt.Errorf("audit %s: %w", action, err)
	}
	return nil
}

func newUUID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic("secure random source unavailable")
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}

func isUniqueConstraintViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
