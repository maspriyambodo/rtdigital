package complaints

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

func (s *Service) dispatchNotification(job notifications.DispatchJob) {
	if s.dispatcher != nil {
		s.dispatcher.Dispatch(job)
	}
}

func (s *Service) CreateComplaint(ctx context.Context, principal *auth.Principal, req CreateComplaintRequest) (ComplaintItem, error) {
	if principal == nil {
		return ComplaintItem{}, ErrForbidden
	}
	if err := req.Validate(); err != nil {
		return ComplaintItem{}, err
	}

	id, now := newUUID(), s.now()
	ticketNumber := fmt.Sprintf("TKT-%s-%s", now.Format("060102"), strings.ToUpper(id[:6]))
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return ComplaintItem{}, fmt.Errorf("begin create complaint: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var categoryExists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM complaint_categories
			WHERE id = $1 AND organization_id = $2 AND status = 'active'
		)`,
		req.ComplaintCategoryID, principal.OrganizationID,
	).Scan(&categoryExists); err != nil {
		return ComplaintItem{}, fmt.Errorf("validate complaint category: %w", err)
	}
	if !categoryExists {
		return ComplaintItem{}, ErrCategoryNotFound
	}

	var responseHours, resolutionHours *int
	if err := tx.QueryRow(ctx, `
		SELECT target_response_hours, target_resolution_hours
		FROM complaint_categories
		WHERE id = $1 AND organization_id = $2`,
		req.ComplaintCategoryID, principal.OrganizationID,
	).Scan(&responseHours, &resolutionHours); err != nil {
		return ComplaintItem{}, fmt.Errorf("get complaint category SLA: %w", err)
	}

	var responseDueAt, resolutionDueAt *time.Time
	if responseHours != nil {
		due := now.Add(time.Duration(*responseHours) * time.Hour)
		responseDueAt = &due
	}
	if resolutionHours != nil {
		due := now.Add(time.Duration(*resolutionHours) * time.Hour)
		resolutionDueAt = &due
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO complaints (
			id, organization_id, reporter_user_id, ticket_number, complaint_category_id, title,
			description, location_description, priority, status, response_due_at, resolution_due_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'new', $10, $11)`,
		id, principal.OrganizationID, principal.UserID, ticketNumber, req.ComplaintCategoryID,
		req.Title, req.Description, req.LocationDescription, req.Priority, responseDueAt, resolutionDueAt,
	); err != nil {
		if isUniqueConstraintViolation(err) {
			return ComplaintItem{}, ErrConflict
		}
		return ComplaintItem{}, fmt.Errorf("insert complaint: %w", err)
	}
	if err := s.syncAttachments(ctx, tx, principal, id, req.AttachmentFileIDs); err != nil {
		return ComplaintItem{}, err
	}
	if err := s.addEvent(ctx, tx, principal.OrganizationID, id, &principal.UserID, "submitted", map[string]string{
		"status": "new",
	}); err != nil {
		return ComplaintItem{}, err
	}
	if err := s.audit(ctx, tx, principal, "complaint.submit", "complaint", id); err != nil {
		return ComplaintItem{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ComplaintItem{}, fmt.Errorf("commit create complaint: %w", err)
	}
	return s.GetComplaint(ctx, principal, id)
}

func (s *Service) ListComplaints(ctx context.Context, principal *auth.Principal, filter ComplaintFilter) ([]ComplaintItem, error) {
	if principal == nil {
		return nil, ErrForbidden
	}
	manager := s.isManager(principal)
	rows, err := s.db.Query(ctx, `
		SELECT id
		FROM complaints
		WHERE organization_id = $1
		  AND ($2 OR reporter_user_id = $3 OR assigned_to = $3)
		  AND ($4 = '' OR status = $4)
		  AND (NULLIF($5, '')::uuid IS NULL OR complaint_category_id = NULLIF($5, '')::uuid)
		  AND (NULLIF($6, '')::uuid IS NULL OR assigned_to = NULLIF($6, '')::uuid)
		  AND ($7 = '' OR ticket_number ILIKE '%' || $7 || '%' OR title ILIKE '%' || $7 || '%')
		ORDER BY created_at DESC
		LIMIT 100`,
		principal.OrganizationID, manager, principal.UserID,
		strings.TrimSpace(filter.Status), strings.TrimSpace(filter.ComplaintCategoryID),
		strings.TrimSpace(filter.AssignedTo), strings.TrimSpace(filter.Search),
	)
	if err != nil {
		return nil, fmt.Errorf("list complaints: %w", err)
	}
	defer rows.Close()

	items := make([]ComplaintItem, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan complaint id: %w", err)
		}
		item, err := s.GetComplaint(ctx, principal, id)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) GetComplaint(ctx context.Context, principal *auth.Principal, id string) (ComplaintItem, error) {
	if principal == nil {
		return ComplaintItem{}, ErrForbidden
	}
	if strings.TrimSpace(id) == "" {
		return ComplaintItem{}, ErrValidation
	}

	var item ComplaintItem
	err := s.db.QueryRow(ctx, `
		SELECT c.id, c.reporter_user_id, COALESCE(ru.email, ru.phone, 'Warga'),
		       c.ticket_number, c.complaint_category_id, cc.name, c.title, c.description,
		       c.location_description, c.priority, c.status, c.assigned_to,
		       NULLIF(COALESCE(au.email, au.phone, ''), ''), c.resolution_note, c.resolved_at,
		       c.closed_at, c.response_due_at, c.responded_at, c.resolution_due_at,
		       c.reporter_confirmation_due_at, c.reporter_confirmed_at, c.closure_reason,
		       c.created_at, c.updated_at
		FROM complaints c
		JOIN complaint_categories cc ON cc.organization_id = c.organization_id AND cc.id = c.complaint_category_id
		JOIN users ru ON ru.organization_id = c.organization_id AND ru.id = c.reporter_user_id
		LEFT JOIN users au ON au.organization_id = c.organization_id AND au.id = c.assigned_to
		WHERE c.organization_id = $1 AND c.id = $2`,
		principal.OrganizationID, id,
	).Scan(
		&item.ID, &item.ReporterUserID, &item.ReporterName, &item.TicketNumber,
		&item.ComplaintCategoryID, &item.CategoryName, &item.Title, &item.Description,
		&item.LocationDescription, &item.Priority, &item.Status, &item.AssignedTo,
		&item.AssignedToName, &item.ResolutionNote, &item.ResolvedAt, &item.ClosedAt,
		&item.ResponseDueAt, &item.RespondedAt, &item.ResolutionDueAt,
		&item.ReporterConfirmationDueAt, &item.ReporterConfirmedAt, &item.ClosureReason,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ComplaintItem{}, ErrComplaintNotFound
	}
	if err != nil {
		return ComplaintItem{}, fmt.Errorf("get complaint: %w", err)
	}

	manager := s.isManager(principal)
	assigned := item.AssignedTo != nil && *item.AssignedTo == principal.UserID
	if !manager && item.ReporterUserID != principal.UserID && !assigned {
		return ComplaintItem{}, ErrForbidden
	}

	item.Attachments, err = s.attachments(ctx, principal.OrganizationID, id)
	if err != nil {
		return ComplaintItem{}, err
	}
	item.Comments, err = s.comments(ctx, principal, id, manager || assigned)
	if err != nil {
		return ComplaintItem{}, err
	}
	item.Events, err = s.events(ctx, principal.OrganizationID, id)
	if err != nil {
		return ComplaintItem{}, err
	}
	return item, nil
}

func (s *Service) UpdateComplaint(ctx context.Context, principal *auth.Principal, id string, req UpdateComplaintRequest) (ComplaintItem, error) {
	if principal == nil {
		return ComplaintItem{}, ErrForbidden
	}
	if err := req.Validate(); err != nil {
		return ComplaintItem{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return ComplaintItem{}, fmt.Errorf("begin update complaint: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var reporterID, status string
	err = tx.QueryRow(ctx, `
		SELECT reporter_user_id, status FROM complaints
		WHERE organization_id = $1 AND id = $2 FOR UPDATE`,
		principal.OrganizationID, id,
	).Scan(&reporterID, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return ComplaintItem{}, ErrComplaintNotFound
	}
	if err != nil {
		return ComplaintItem{}, fmt.Errorf("lock complaint: %w", err)
	}
	if reporterID != principal.UserID {
		return ComplaintItem{}, ErrForbidden
	}
	if status != "new" {
		return ComplaintItem{}, ErrInvalidState
	}

	var categoryExists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM complaint_categories
			WHERE id = $1 AND organization_id = $2 AND status = 'active'
		)`,
		req.ComplaintCategoryID, principal.OrganizationID,
	).Scan(&categoryExists); err != nil {
		return ComplaintItem{}, fmt.Errorf("validate complaint category: %w", err)
	}
	if !categoryExists {
		return ComplaintItem{}, ErrCategoryNotFound
	}

	if _, err := tx.Exec(ctx, `
		UPDATE complaints
		SET complaint_category_id = $1, title = $2, description = $3, location_description = $4, priority = $5
		WHERE organization_id = $6 AND id = $7`,
		req.ComplaintCategoryID, req.Title, req.Description, req.LocationDescription, req.Priority,
		principal.OrganizationID, id,
	); err != nil {
		return ComplaintItem{}, fmt.Errorf("update complaint: %w", err)
	}
	if err := s.syncAttachments(ctx, tx, principal, id, req.AttachmentFileIDs); err != nil {
		return ComplaintItem{}, err
	}
	if err := s.audit(ctx, tx, principal, "complaint.update", "complaint", id); err != nil {
		return ComplaintItem{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ComplaintItem{}, fmt.Errorf("commit update complaint: %w", err)
	}
	return s.GetComplaint(ctx, principal, id)
}

func (s *Service) AssignComplaint(ctx context.Context, principal *auth.Principal, id string, req AssignComplaintRequest) (ComplaintItem, error) {
	if principal == nil || !principal.HasPermission("complaint.assign") {
		return ComplaintItem{}, ErrForbidden
	}
	if err := req.Validate(); err != nil {
		return ComplaintItem{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return ComplaintItem{}, fmt.Errorf("begin assign complaint: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var status string
	err = tx.QueryRow(ctx, `
		SELECT status FROM complaints
		WHERE organization_id = $1 AND id = $2 FOR UPDATE`,
		principal.OrganizationID, id,
	).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return ComplaintItem{}, ErrComplaintNotFound
	}
	if err != nil {
		return ComplaintItem{}, fmt.Errorf("lock complaint: %w", err)
	}
	if status == "resolved" || status == "rejected" || status == "closed" {
		return ComplaintItem{}, ErrInvalidState
	}

	var targetExists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM users
			WHERE organization_id = $1 AND id = $2 AND status = 'active'
		)`, principal.OrganizationID, req.AssignedTo,
	).Scan(&targetExists); err != nil {
		return ComplaintItem{}, fmt.Errorf("validate assignee: %w", err)
	}
	if !targetExists {
		return ComplaintItem{}, ErrValidation
	}

	nextStatus := status
	if status == "new" {
		nextStatus = "reviewed"
	}
	if _, err := tx.Exec(ctx, `
		UPDATE complaints
		SET assigned_to = $1,
		    status = $2,
		    responded_at = CASE
		        WHEN responded_at IS NULL AND $2 <> 'new' THEN $3
		        ELSE responded_at
		    END
		WHERE organization_id = $4 AND id = $5`,
		req.AssignedTo, nextStatus, s.now(), principal.OrganizationID, id,
	); err != nil {
		return ComplaintItem{}, fmt.Errorf("assign complaint: %w", err)
	}
	if err := s.addEvent(ctx, tx, principal.OrganizationID, id, &principal.UserID, "assigned", map[string]string{
		"assigned_to": req.AssignedTo,
		"status":      nextStatus,
	}); err != nil {
		return ComplaintItem{}, err
	}
	if err := s.audit(ctx, tx, principal, "complaint.assign", "complaint", id); err != nil {
		return ComplaintItem{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ComplaintItem{}, fmt.Errorf("commit assign complaint: %w", err)
	}
	s.dispatchNotification(notifications.DispatchJob{
		OrganizationID:  principal.OrganizationID,
		RecipientUserID: req.AssignedTo,
		Type:            "complaint_assigned",
		Title:           "Penugasan aduan baru",
		Body:            "Anda ditugaskan menangani aduan warga.",
		ReferenceType:   "complaint",
		ReferenceID:     id,
	})
	return s.GetComplaint(ctx, principal, id)
}

func (s *Service) ConfirmResolution(ctx context.Context, principal *auth.Principal, id string) (ComplaintItem, error) {
	if principal == nil {
		return ComplaintItem{}, ErrForbidden
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return ComplaintItem{}, fmt.Errorf("begin confirm complaint resolution: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var reporterID, status string
	err = tx.QueryRow(ctx, `
		SELECT reporter_user_id, status
		FROM complaints
		WHERE organization_id = $1 AND id = $2
		FOR UPDATE`,
		principal.OrganizationID, id,
	).Scan(&reporterID, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return ComplaintItem{}, ErrComplaintNotFound
	}
	if err != nil {
		return ComplaintItem{}, fmt.Errorf("lock complaint confirmation: %w", err)
	}
	if reporterID != principal.UserID {
		return ComplaintItem{}, ErrForbidden
	}
	if status != "resolved" {
		return ComplaintItem{}, ErrInvalidState
	}

	now := s.now()
	if _, err := tx.Exec(ctx, `
		UPDATE complaints
		SET status = 'closed',
		    closed_at = $1,
		    closed_by = $2,
		    closure_reason = 'confirmed_by_reporter',
		    reporter_confirmed_at = $1
		WHERE organization_id = $3 AND id = $4`,
		now, principal.UserID, principal.OrganizationID, id,
	); err != nil {
		return ComplaintItem{}, fmt.Errorf("confirm complaint resolution: %w", err)
	}
	if err := s.addEvent(ctx, tx, principal.OrganizationID, id, &principal.UserID, "reporter_confirmed", map[string]string{
		"status": "closed",
	}); err != nil {
		return ComplaintItem{}, err
	}
	if err := s.audit(ctx, tx, principal, "complaint.confirm_resolution", "complaint", id); err != nil {
		return ComplaintItem{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ComplaintItem{}, fmt.Errorf("commit complaint confirmation: %w", err)
	}
	return s.GetComplaint(ctx, principal, id)
}

func (s *Service) UpdateStatus(ctx context.Context, principal *auth.Principal, id string, req UpdateStatusRequest) (ComplaintItem, error) {
	if principal == nil {
		return ComplaintItem{}, ErrForbidden
	}
	if err := req.Validate(); err != nil {
		return ComplaintItem{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return ComplaintItem{}, fmt.Errorf("begin update complaint status: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var reporterID, assignedTo, current string
	err = tx.QueryRow(ctx, `
		SELECT reporter_user_id, COALESCE(assigned_to::text, ''), status
		FROM complaints
		WHERE organization_id = $1 AND id = $2 FOR UPDATE`,
		principal.OrganizationID, id,
	).Scan(&reporterID, &assignedTo, &current)
	if errors.Is(err, pgx.ErrNoRows) {
		return ComplaintItem{}, ErrComplaintNotFound
	}
	if err != nil {
		return ComplaintItem{}, fmt.Errorf("lock complaint: %w", err)
	}

	assigned := assignedTo == principal.UserID
	manager := principal.HasPermission("complaint.update_status")
	if req.Status == "closed" {
		if reporterID != principal.UserID || current != "resolved" {
			return ComplaintItem{}, ErrForbidden
		}
	} else if !manager && !assigned {
		return ComplaintItem{}, ErrForbidden
	}
	if !validTransition(current, req.Status) {
		return ComplaintItem{}, ErrInvalidState
	}

	now := s.now()
	var query string
	var args []any
	switch req.Status {
	case "resolved":
		var confirmationHours *int
		if err := tx.QueryRow(ctx, `
			SELECT cc.target_reporter_confirmation_hours
			FROM complaints c
			JOIN complaint_categories cc
			  ON cc.organization_id = c.organization_id
			 AND cc.id = c.complaint_category_id
			WHERE c.organization_id = $1 AND c.id = $2`,
			principal.OrganizationID, id,
		).Scan(&confirmationHours); err != nil {
			return ComplaintItem{}, fmt.Errorf("get complaint confirmation SLA: %w", err)
		}

		var confirmationDueAt *time.Time
		if confirmationHours != nil {
			due := now.Add(time.Duration(*confirmationHours) * time.Hour)
			confirmationDueAt = &due
		}
		query = `
			UPDATE complaints
			SET status = $1,
			    responded_at = COALESCE(responded_at, $2),
			    resolution_note = $3,
			    resolved_at = $2,
			    reporter_confirmation_due_at = $4
			WHERE organization_id = $5 AND id = $6`
		args = []any{req.Status, now, req.ResolutionNote, confirmationDueAt, principal.OrganizationID, id}
	case "rejected":
		query = `
			UPDATE complaints
			SET status = $1,
			    responded_at = COALESCE(responded_at, $2),
			    resolution_note = $3,
			    resolved_at = $2
			WHERE organization_id = $4 AND id = $5`
		args = []any{req.Status, now, req.ResolutionNote, principal.OrganizationID, id}
	case "closed":
		query = `
			UPDATE complaints
			SET status = $1,
			    responded_at = COALESCE(responded_at, $2),
			    closed_at = $2,
			    closed_by = $3,
			    closure_reason = 'closed_by_reporter'
			WHERE organization_id = $4 AND id = $5`
		args = []any{req.Status, now, principal.UserID, principal.OrganizationID, id}
	default:
		query = `
			UPDATE complaints
			SET status = $1,
			    responded_at = COALESCE(responded_at, $2)
			WHERE organization_id = $3 AND id = $4`
		args = []any{req.Status, now, principal.OrganizationID, id}
	}

	if _, err := tx.Exec(ctx, query, args...); err != nil {
		return ComplaintItem{}, fmt.Errorf("update complaint status: %w", err)
	}
	if err := s.addEvent(ctx, tx, principal.OrganizationID, id, &principal.UserID, "status_changed", map[string]string{
		"from": current,
		"to":   req.Status,
	}); err != nil {
		return ComplaintItem{}, err
	}
	if err := s.audit(ctx, tx, principal, "complaint.update_status", "complaint", id); err != nil {
		return ComplaintItem{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ComplaintItem{}, fmt.Errorf("commit update complaint status: %w", err)
	}
	var title, body string
	switch req.Status {
	case "reviewed":
		title = "Aduan ditinjau"
		body = "Aduan Anda sedang ditinjau oleh pengurus."
	case "in_progress":
		title = "Aduan diproses"
		body = "Aduan Anda sedang ditangani petugas."
	case "waiting_information":
		title = "Aduan memerlukan informasi"
		body = "Petugas membutuhkan informasi tambahan untuk aduan Anda."
	case "resolved":
		title = "Aduan selesai ditangani"
		if req.ResolutionNote != nil {
			body = fmt.Sprintf("Aduan Anda telah diselesaikan: %s", *req.ResolutionNote)
		} else {
			body = "Aduan Anda telah diselesaikan."
		}
	case "rejected":
		title = "Aduan ditolak"
		if req.ResolutionNote != nil {
			body = fmt.Sprintf("Aduan Anda ditolak: %s", *req.ResolutionNote)
		} else {
			body = "Aduan Anda ditolak."
		}
	}
	if title != "" && reporterID != principal.UserID {
		s.dispatchNotification(notifications.DispatchJob{
			OrganizationID:  principal.OrganizationID,
			RecipientUserID: reporterID,
			Type:            "complaint_" + req.Status,
			Title:           title,
			Body:            body,
			ReferenceType:   "complaint",
			ReferenceID:     id,
		})
	}
	return s.GetComplaint(ctx, principal, id)
}

func (s *Service) AddComment(ctx context.Context, principal *auth.Principal, complaintID string, req AddCommentRequest) (CommentItem, error) {
	if principal == nil || !principal.HasPermission("complaint.comment") {
		return CommentItem{}, ErrForbidden
	}
	if err := req.Validate(); err != nil {
		return CommentItem{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return CommentItem{}, fmt.Errorf("begin complaint comment: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var reporterID, assignedTo string
	err = tx.QueryRow(ctx, `
		SELECT reporter_user_id, COALESCE(assigned_to::text, '')
		FROM complaints
		WHERE organization_id = $1 AND id = $2 FOR UPDATE`,
		principal.OrganizationID, complaintID,
	).Scan(&reporterID, &assignedTo)
	if errors.Is(err, pgx.ErrNoRows) {
		return CommentItem{}, ErrComplaintNotFound
	}
	if err != nil {
		return CommentItem{}, fmt.Errorf("lock complaint: %w", err)
	}

	manager := s.isManager(principal)
	assigned := assignedTo == principal.UserID
	reporter := reporterID == principal.UserID
	if !manager && !assigned && !reporter {
		return CommentItem{}, ErrForbidden
	}
	if req.IsInternal && !manager && !assigned {
		return CommentItem{}, ErrForbidden
	}

	commentID := newUUID()
	if _, err := tx.Exec(ctx, `
		INSERT INTO complaint_comments (
			id, organization_id, complaint_id, author_user_id, body, is_internal
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		commentID, principal.OrganizationID, complaintID, principal.UserID, req.Body, req.IsInternal,
	); err != nil {
		return CommentItem{}, fmt.Errorf("insert complaint comment: %w", err)
	}
	if err := s.audit(ctx, tx, principal, "complaint.comment", "complaint_comment", commentID); err != nil {
		return CommentItem{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CommentItem{}, fmt.Errorf("commit complaint comment: %w", err)
	}
	if !req.IsInternal {
		recipientUserID := reporterID
		if principal.UserID == reporterID {
			recipientUserID = assignedTo
		}
		if recipientUserID != "" && recipientUserID != principal.UserID {
			s.dispatchNotification(notifications.DispatchJob{
				OrganizationID:  principal.OrganizationID,
				RecipientUserID: recipientUserID,
				Type:            "complaint_comment_added",
				Title:           "Komentar aduan baru",
				Body:            "Ada pembaruan pada tiket aduan terkait.",
				ReferenceType:   "complaint",
				ReferenceID:     complaintID,
			})
		}
	}
	return s.getComment(ctx, principal, commentID)
}

func (s *Service) getComment(ctx context.Context, principal *auth.Principal, id string) (CommentItem, error) {
	var item CommentItem
	err := s.db.QueryRow(ctx, `
		SELECT cc.id, cc.complaint_id, cc.author_user_id, COALESCE(u.email, u.phone, 'Warga'),
		       cc.body, cc.is_internal, cc.created_at
		FROM complaint_comments cc
		JOIN users u ON u.organization_id = cc.organization_id AND u.id = cc.author_user_id
		WHERE cc.organization_id = $1 AND cc.id = $2`,
		principal.OrganizationID, id,
	).Scan(&item.ID, &item.ComplaintID, &item.AuthorUserID, &item.AuthorName, &item.Body, &item.IsInternal, &item.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return CommentItem{}, ErrComplaintNotFound
	}
	if err != nil {
		return CommentItem{}, fmt.Errorf("get complaint comment: %w", err)
	}
	return item, nil
}

func (s *Service) comments(ctx context.Context, principal *auth.Principal, complaintID string, includeInternal bool) ([]CommentItem, error) {
	rows, err := s.db.Query(ctx, `
		SELECT cc.id, cc.complaint_id, cc.author_user_id, COALESCE(u.email, u.phone, 'Warga'),
		       cc.body, cc.is_internal, cc.created_at
		FROM complaint_comments cc
		JOIN users u ON u.organization_id = cc.organization_id AND u.id = cc.author_user_id
		WHERE cc.organization_id = $1 AND cc.complaint_id = $2
		  AND ($3 OR cc.is_internal = false)
		ORDER BY cc.created_at ASC`,
		principal.OrganizationID, complaintID, includeInternal,
	)
	if err != nil {
		return nil, fmt.Errorf("list complaint comments: %w", err)
	}
	defer rows.Close()

	items := make([]CommentItem, 0)
	for rows.Next() {
		var item CommentItem
		if err := rows.Scan(
			&item.ID, &item.ComplaintID, &item.AuthorUserID, &item.AuthorName,
			&item.Body, &item.IsInternal, &item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan complaint comment: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) addEvent(ctx context.Context, tx pgx.Tx, organizationID, complaintID string, actorUserID *string, eventType string, data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("encode complaint event: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO complaint_events (id, organization_id, complaint_id, actor_user_id, event_type, data)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		newUUID(), organizationID, complaintID, actorUserID, eventType, payload,
	); err != nil {
		return fmt.Errorf("insert complaint event: %w", err)
	}
	return nil
}

func (s *Service) events(ctx context.Context, organizationID, complaintID string) ([]ComplaintEvent, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, event_type, data, created_at
		FROM complaint_events
		WHERE organization_id = $1 AND complaint_id = $2
		ORDER BY created_at ASC, id ASC`,
		organizationID, complaintID,
	)
	if err != nil {
		return nil, fmt.Errorf("list complaint events: %w", err)
	}
	defer rows.Close()

	items := make([]ComplaintEvent, 0)
	for rows.Next() {
		var item ComplaintEvent
		var data json.RawMessage
		if err := rows.Scan(&item.ID, &item.EventType, &data, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan complaint event: %w", err)
		}
		if err := json.Unmarshal(data, &item.Data); err != nil {
			return nil, fmt.Errorf("decode complaint event: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) syncAttachments(ctx context.Context, tx pgx.Tx, principal *auth.Principal, complaintID string, fileIDs []string) error {
	if _, err := tx.Exec(ctx, `
		DELETE FROM file_attachments
		WHERE organization_id = $1 AND entity_type = 'complaint' AND entity_id = $2`,
		principal.OrganizationID, complaintID,
	); err != nil {
		return fmt.Errorf("clear complaint attachments: %w", err)
	}

	seen := make(map[string]struct{}, len(fileIDs))
	for _, fileID := range fileIDs {
		fileID = strings.TrimSpace(fileID)
		if fileID == "" {
			return ErrValidation
		}
		if _, duplicate := seen[fileID]; duplicate {
			return ErrValidation
		}
		seen[fileID] = struct{}{}

		var valid bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM file_objects
				WHERE id = $1 AND organization_id = $2 AND uploaded_by = $3
				  AND confirmed_at IS NOT NULL AND deleted_at IS NULL
			)`, fileID, principal.OrganizationID, principal.UserID,
		).Scan(&valid); err != nil {
			return fmt.Errorf("validate complaint attachment: %w", err)
		}
		if !valid {
			return ErrValidation
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO file_attachments (
				id, organization_id, file_id, entity_type, entity_id, purpose
			) VALUES ($1, $2, $3, 'complaint', $4, 'attachment')`,
			newUUID(), principal.OrganizationID, fileID, complaintID,
		); err != nil {
			return fmt.Errorf("attach complaint file: %w", err)
		}
	}
	return nil
}

func (s *Service) attachments(ctx context.Context, organizationID, complaintID string) ([]AttachmentInfo, error) {
	rows, err := s.db.Query(ctx, `
		SELECT a.id, a.file_id, f.original_name, f.mime_type, f.size_bytes, a.purpose
		FROM file_attachments a
		JOIN file_objects f ON f.organization_id = a.organization_id AND f.id = a.file_id
		WHERE a.organization_id = $1 AND a.entity_type = 'complaint' AND a.entity_id = $2
		  AND f.confirmed_at IS NOT NULL AND f.deleted_at IS NULL
		ORDER BY a.created_at`,
		organizationID, complaintID,
	)
	if err != nil {
		return nil, fmt.Errorf("list complaint attachments: %w", err)
	}
	defer rows.Close()

	items := make([]AttachmentInfo, 0)
	for rows.Next() {
		var item AttachmentInfo
		if err := rows.Scan(&item.AttachmentID, &item.FileID, &item.OriginalName, &item.MIMEType, &item.SizeBytes, &item.Purpose); err != nil {
			return nil, fmt.Errorf("scan complaint attachment: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
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

// AutoCloseComplaints closes resolved complaints whose reporter-confirmation
// deadline has elapsed. It is safe to invoke repeatedly from a scheduler.
func (s *Service) AutoCloseComplaints(ctx context.Context) (int, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin complaint auto-close: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `
		UPDATE complaints
		SET status = 'closed',
		    closed_at = now(),
		    closed_by = NULL,
		    closure_reason = 'auto_closed_confirmation_timeout'
		WHERE status = 'resolved'
		  AND reporter_confirmation_due_at IS NOT NULL
		  AND reporter_confirmation_due_at <= now()
		RETURNING organization_id, id`,
	)
	if err != nil {
		return 0, fmt.Errorf("auto-close expired complaints: %w", err)
	}

	count := 0
	for rows.Next() {
		var organizationID, complaintID string
		if err := rows.Scan(&organizationID, &complaintID); err != nil {
			rows.Close()
			return 0, fmt.Errorf("scan auto-closed complaint: %w", err)
		}
		if err := s.addEvent(ctx, tx, organizationID, complaintID, nil, "auto_closed", map[string]string{
			"status": "closed",
			"reason": "reporter_confirmation_timeout",
		}); err != nil {
			rows.Close()
			return 0, err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO audit_logs (organization_id, actor_user_id, action, entity_type, entity_id)
			VALUES ($1, NULL, 'complaint.auto_close', 'complaint', $2)`,
			organizationID, complaintID,
		); err != nil {
			rows.Close()
			return 0, fmt.Errorf("audit complaint auto-close: %w", err)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, fmt.Errorf("iterate auto-closed complaints: %w", err)
	}
	rows.Close()

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit complaint auto-close: %w", err)
	}
	return count, nil
}

func (s *Service) isManager(principal *auth.Principal) bool {
	return principal.HasPermission("complaint.assign") || principal.HasPermission("complaint.update_status")
}

func (s *Service) ListComplaintCategories(ctx context.Context, principal *auth.Principal, onlyActive bool) ([]ComplaintCategory, error) {
	if principal == nil {
		return nil, ErrForbidden
	}
	rows, err := s.db.Query(ctx, `
		SELECT id, code, name, status, target_response_hours, target_resolution_hours,
		       target_reporter_confirmation_hours, created_at, updated_at
		FROM complaint_categories
		WHERE organization_id = $1
		  AND (NOT $2 OR status = 'active')
		ORDER BY name`,
		principal.OrganizationID, onlyActive,
	)
	if err != nil {
		return nil, fmt.Errorf("list complaint categories: %w", err)
	}
	defer rows.Close()

	items := make([]ComplaintCategory, 0)
	for rows.Next() {
		var item ComplaintCategory
		if err := rows.Scan(
			&item.ID, &item.Code, &item.Name, &item.Status,
			&item.TargetResponseHours, &item.TargetResolutionHours,
			&item.TargetReporterConfirmationHours, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan complaint category: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) GetComplaintCategory(ctx context.Context, principal *auth.Principal, id string) (ComplaintCategory, error) {
	if principal == nil || strings.TrimSpace(id) == "" {
		return ComplaintCategory{}, ErrValidation
	}
	var item ComplaintCategory
	err := s.db.QueryRow(ctx, `
		SELECT id, code, name, status, target_response_hours, target_resolution_hours,
		       target_reporter_confirmation_hours, created_at, updated_at
		FROM complaint_categories
		WHERE organization_id = $1 AND id = $2`,
		principal.OrganizationID, id,
	).Scan(
		&item.ID, &item.Code, &item.Name, &item.Status,
		&item.TargetResponseHours, &item.TargetResolutionHours,
		&item.TargetReporterConfirmationHours, &item.CreatedAt, &item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ComplaintCategory{}, ErrCategoryNotFound
	}
	if err != nil {
		return ComplaintCategory{}, fmt.Errorf("get complaint category: %w", err)
	}
	return item, nil
}

func (s *Service) CreateComplaintCategory(ctx context.Context, principal *auth.Principal, req CreateComplaintCategoryRequest) (ComplaintCategory, error) {
	if principal == nil {
		return ComplaintCategory{}, ErrForbidden
	}
	if err := req.Validate(); err != nil {
		return ComplaintCategory{}, err
	}
	id := newUUID()
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return ComplaintCategory{}, fmt.Errorf("begin create complaint category: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
		INSERT INTO complaint_categories (
			id, organization_id, code, name, target_response_hours, target_resolution_hours,
			target_reporter_confirmation_hours
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, principal.OrganizationID, req.Code, req.Name, req.TargetResponseHours,
		req.TargetResolutionHours, req.TargetReporterConfirmationHours,
	); err != nil {
		if isUniqueConstraintViolation(err) {
			return ComplaintCategory{}, ErrConflict
		}
		return ComplaintCategory{}, fmt.Errorf("insert complaint category: %w", err)
	}
	if err := s.audit(ctx, tx, principal, "complaint_category.create", "complaint_category", id); err != nil {
		return ComplaintCategory{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ComplaintCategory{}, fmt.Errorf("commit create complaint category: %w", err)
	}
	return s.GetComplaintCategory(ctx, principal, id)
}

func (s *Service) UpdateComplaintCategory(ctx context.Context, principal *auth.Principal, id string, req UpdateComplaintCategoryRequest) (ComplaintCategory, error) {
	if principal == nil {
		return ComplaintCategory{}, ErrForbidden
	}
	if err := req.Validate(); err != nil {
		return ComplaintCategory{}, err
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return ComplaintCategory{}, fmt.Errorf("begin update complaint category: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx, `
		UPDATE complaint_categories
		SET name = $1,
		    status = COALESCE(NULLIF($2, ''), status),
		    target_response_hours = COALESCE($3, target_response_hours),
		    target_resolution_hours = COALESCE($4, target_resolution_hours),
		    target_reporter_confirmation_hours = COALESCE($5, target_reporter_confirmation_hours)
		WHERE organization_id = $6 AND id = $7`,
		req.Name, req.Status, req.TargetResponseHours, req.TargetResolutionHours,
		req.TargetReporterConfirmationHours, principal.OrganizationID, id,
	)
	if err != nil {
		return ComplaintCategory{}, fmt.Errorf("update complaint category: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ComplaintCategory{}, ErrCategoryNotFound
	}
	if err := s.audit(ctx, tx, principal, "complaint_category.update", "complaint_category", id); err != nil {
		return ComplaintCategory{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ComplaintCategory{}, fmt.Errorf("commit update complaint category: %w", err)
	}
	return s.GetComplaintCategory(ctx, principal, id)
}

func validTransition(from, to string) bool {
	switch from {
	case "new":
		return to == "reviewed" || to == "in_progress" || to == "waiting_information" || to == "rejected"
	case "reviewed":
		return to == "in_progress" || to == "waiting_information" || to == "rejected"
	case "in_progress":
		return to == "waiting_information" || to == "resolved" || to == "rejected"
	case "waiting_information":
		return to == "in_progress" || to == "rejected"
	case "resolved":
		return to == "closed"
	default:
		return false
	}
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
