package cash

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

type Service struct {
	db  *pgxpool.Pool
	now func() time.Time
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{
		db:  db,
		now: func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) CreateCategory(ctx context.Context, principal *auth.Principal, request CreateCategoryRequest) (Category, error) {
	if principal == nil || request.Validate() != nil {
		return Category{}, ErrValidation
	}

	now := s.now()
	category := Category{}
	err := s.db.QueryRow(ctx, `
		INSERT INTO cash_categories (id, organization_id, name, type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING id, name, type, status, created_at, updated_at`,
		newUUID(), principal.OrganizationID, request.Name, request.Type, now,
	).Scan(
		&category.ID, &category.Name, &category.Type, &category.Status,
		&category.CreatedAt, &category.UpdatedAt,
	)
	if err != nil {
		return Category{}, mapDatabaseError(err, "create cash category")
	}
	return category, nil
}

func (s *Service) ListCategories(ctx context.Context, principal *auth.Principal) ([]Category, error) {
	if principal == nil {
		return nil, ErrValidation
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, name, type, status, created_at, updated_at
		FROM cash_categories
		WHERE organization_id = $1
		ORDER BY type, name`,
		principal.OrganizationID,
	)
	if err != nil {
		return nil, fmt.Errorf("list cash categories: %w", err)
	}
	defer rows.Close()

	categories := []Category{}
	for rows.Next() {
		var category Category
		if err := rows.Scan(
			&category.ID, &category.Name, &category.Type, &category.Status,
			&category.CreatedAt, &category.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan cash category: %w", err)
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cash categories: %w", err)
	}
	return categories, nil
}

func (s *Service) UpdateCategory(ctx context.Context, principal *auth.Principal, categoryID string, request UpdateCategoryRequest) (Category, error) {
	if principal == nil || request.Validate() != nil {
		return Category{}, ErrValidation
	}

	category := Category{}
	err := s.db.QueryRow(ctx, `
		UPDATE cash_categories
		SET name = COALESCE($1, name),
		    status = COALESCE($2, status)
		WHERE id = $3
		  AND organization_id = $4
		RETURNING id, name, type, status, created_at, updated_at`,
		request.Name, request.Status, categoryID, principal.OrganizationID,
	).Scan(
		&category.ID, &category.Name, &category.Type, &category.Status,
		&category.CreatedAt, &category.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Category{}, ErrCategoryNotFound
	}
	if err != nil {
		return Category{}, mapDatabaseError(err, "update cash category")
	}
	return category, nil
}

func (s *Service) RecordManual(ctx context.Context, principal *auth.Principal, request RecordTransactionRequest) (Transaction, error) {
	if principal == nil || request.Validate() != nil {
		return Transaction{}, ErrValidation
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Transaction{}, fmt.Errorf("begin cash transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var categoryName string
	err = tx.QueryRow(ctx, `
		SELECT name
		FROM cash_categories
		WHERE id = $1
		  AND organization_id = $2
		  AND type = $3
		  AND status = 'active'`,
		request.CategoryID, principal.OrganizationID, request.Type,
	).Scan(&categoryName)
	if errors.Is(err, pgx.ErrNoRows) {
		return Transaction{}, ErrCategoryNotFound
	}
	if err != nil {
		return Transaction{}, fmt.Errorf("find cash category: %w", err)
	}

	if request.ProofFileID != nil {
		var confirmed bool
		err := tx.QueryRow(ctx, `
			SELECT confirmed_at IS NOT NULL
			FROM file_objects
			WHERE id = $1
			  AND organization_id = $2
			  AND deleted_at IS NULL
			FOR UPDATE`,
			*request.ProofFileID, principal.OrganizationID,
		).Scan(&confirmed)
		if errors.Is(err, pgx.ErrNoRows) {
			return Transaction{}, ErrConstraint
		}
		if err != nil {
			return Transaction{}, fmt.Errorf("lock cash proof: %w", err)
		}
		if !confirmed {
			return Transaction{}, ErrConstraint
		}
	}

	now := s.now()
	transactionID := newUUID()
	transactionNumber := fmt.Sprintf("KAS-%s-%s", now.Format("0601"), transactionID[:8])
	if _, err := tx.Exec(ctx, `
		INSERT INTO cash_transactions (
			id, organization_id, transaction_number, type, category_id, amount,
			transaction_date, description, proof_file_id, created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11)`,
		transactionID, principal.OrganizationID, transactionNumber, request.Type,
		request.CategoryID, request.Amount, request.TransactionDate, request.Description,
		request.ProofFileID, principal.UserID, now,
	); err != nil {
		return Transaction{}, mapDatabaseError(err, "insert cash transaction")
	}

	if request.ProofFileID != nil {
		if _, err := tx.Exec(ctx, `
			INSERT INTO file_attachments (
				id, organization_id, file_id, entity_type, entity_id, purpose
			) VALUES ($1, $2, $3, 'cash_transaction', $4, 'proof')`,
			newUUID(), principal.OrganizationID, *request.ProofFileID, transactionID,
		); err != nil {
			return Transaction{}, mapDatabaseError(err, "attach cash proof")
		}
	}
	if err := s.audit(ctx, tx, principal, "cash.record", transactionID); err != nil {
		return Transaction{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Transaction{}, fmt.Errorf("commit cash transaction: %w", err)
	}

	return Transaction{
		ID: transactionID, TransactionNumber: transactionNumber, Type: request.Type,
		CategoryID: &request.CategoryID, CategoryName: &categoryName, Amount: request.Amount,
		TransactionDate: request.TransactionDate, Description: request.Description,
		ProofFileID: request.ProofFileID, Status: "active", CreatedBy: principal.UserID,
		CreatedAt: now, UpdatedAt: now,
	}, nil
}

func (s *Service) Reverse(ctx context.Context, principal *auth.Principal, transactionID string, request ReverseTransactionRequest) (Transaction, error) {
	if principal == nil || request.Validate() != nil {
		return Transaction{}, ErrValidation
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Transaction{}, fmt.Errorf("begin cash reversal: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var original Transaction
	var transactionDate time.Time
	err = tx.QueryRow(ctx, `
		SELECT id, transaction_number, type, category_id, amount, transaction_date,
		       description, status, reversal_of_id
		FROM cash_transactions
		WHERE id = $1 AND organization_id = $2
		FOR UPDATE`,
		transactionID, principal.OrganizationID,
	).Scan(
		&original.ID, &original.TransactionNumber, &original.Type, &original.CategoryID,
		&original.Amount, &transactionDate, &original.Description, &original.Status,
		&original.ReversalOfID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Transaction{}, ErrTransactionNotFound
	}
	if err != nil {
		return Transaction{}, fmt.Errorf("lock cash transaction: %w", err)
	}
	if original.Status != "active" || original.ReversalOfID != nil {
		return Transaction{}, ErrInvalidState
	}

	now := s.now()
	reversalID := newUUID()
	reversalType := "expense"
	if original.Type == "expense" {
		reversalType = "income"
	}
	reversalNumber := fmt.Sprintf("KAS-%s-%s", now.Format("0601"), reversalID[:8])
	reversalDescription := fmt.Sprintf("Pembalikan %s. Alasan: %s", original.TransactionNumber, request.Reason)

	if _, err := tx.Exec(ctx, `
		INSERT INTO cash_transactions (
			id, organization_id, transaction_number, type, category_id, amount,
			transaction_date, description, reversal_of_id, created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11)`,
		reversalID, principal.OrganizationID, reversalNumber, reversalType,
		original.CategoryID, original.Amount, now.Format(time.DateOnly),
		reversalDescription, transactionID, principal.UserID, now,
	); err != nil {
		return Transaction{}, mapDatabaseError(err, "insert cash reversal")
	}
	if _, err := tx.Exec(ctx, `
		UPDATE cash_transactions
		SET status = 'reversed'
		WHERE id = $1 AND organization_id = $2`,
		transactionID, principal.OrganizationID,
	); err != nil {
		return Transaction{}, fmt.Errorf("mark cash transaction reversed: %w", err)
	}
	if err := s.audit(ctx, tx, principal, "cash.reverse", reversalID); err != nil {
		return Transaction{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Transaction{}, fmt.Errorf("commit cash reversal: %w", err)
	}

	return Transaction{
		ID: reversalID, TransactionNumber: reversalNumber, Type: reversalType,
		CategoryID: original.CategoryID, Amount: original.Amount,
		TransactionDate: now.Format(time.DateOnly), Description: reversalDescription,
		ReversalOfID: &transactionID, Status: "active", CreatedBy: principal.UserID,
		CreatedAt: now, UpdatedAt: now,
	}, nil
}

func (s *Service) GetBook(ctx context.Context, principal *auth.Principal, filter TransactionFilter) (CashBook, error) {
	if principal == nil || filter.Validate() != nil {
		return CashBook{}, ErrValidation
	}

	var openingBalance float64
	if filter.StartDate != "" {
		if err := s.db.QueryRow(ctx, `
			SELECT COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE -amount END), 0)
			FROM cash_transactions
			WHERE organization_id = $1
			  AND transaction_date < $2`,
			principal.OrganizationID, filter.StartDate,
		).Scan(&openingBalance); err != nil {
			return CashBook{}, fmt.Errorf("get opening balance: %w", err)
		}
	}

	conditions := []string{"t.organization_id = $1"}
	args := []any{principal.OrganizationID}
	argument := 2
	if filter.StartDate != "" {
		conditions = append(conditions, fmt.Sprintf("t.transaction_date >= $%d", argument))
		args = append(args, filter.StartDate)
		argument++
	}
	if filter.EndDate != "" {
		conditions = append(conditions, fmt.Sprintf("t.transaction_date <= $%d", argument))
		args = append(args, filter.EndDate)
		argument++
	}
	if filter.Type != "" {
		conditions = append(conditions, fmt.Sprintf("t.type = $%d", argument))
		args = append(args, filter.Type)
		argument++
	}
	if filter.CategoryID != "" {
		conditions = append(conditions, fmt.Sprintf("t.category_id = $%d", argument))
		args = append(args, filter.CategoryID)
	}

	rows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT t.id, t.transaction_number, t.type, t.category_id, c.name, t.amount,
		       t.transaction_date, t.description, t.proof_file_id, t.reference_type,
		       t.reference_id, t.reversal_of_id, t.status, t.created_by,
		       t.created_at, t.updated_at
		FROM cash_transactions t
		LEFT JOIN cash_categories c
		  ON c.organization_id = t.organization_id AND c.id = t.category_id
		WHERE %s
		ORDER BY t.transaction_date ASC, t.created_at ASC`, strings.Join(conditions, " AND ")),
		args...,
	)
	if err != nil {
		return CashBook{}, fmt.Errorf("list cash transactions: %w", err)
	}
	defer rows.Close()

	book := CashBook{Transactions: []Transaction{}}
	runningBalance := openingBalance
	for rows.Next() {
		var transaction Transaction
		var transactionDate time.Time
		if err := rows.Scan(
			&transaction.ID, &transaction.TransactionNumber, &transaction.Type,
			&transaction.CategoryID, &transaction.CategoryName, &transaction.Amount,
			&transactionDate, &transaction.Description, &transaction.ProofFileID,
			&transaction.ReferenceType, &transaction.ReferenceID, &transaction.ReversalOfID,
			&transaction.Status, &transaction.CreatedBy, &transaction.CreatedAt,
			&transaction.UpdatedAt,
		); err != nil {
			return CashBook{}, fmt.Errorf("scan cash transaction: %w", err)
		}
		transaction.TransactionDate = transactionDate.Format(time.DateOnly)
		if transaction.Type == "income" {
			book.TotalIncome += transaction.Amount
			runningBalance += transaction.Amount
		} else {
			book.TotalExpense += transaction.Amount
			runningBalance -= transaction.Amount
		}
		transaction.RunningBalance = runningBalance
		book.Transactions = append(book.Transactions, transaction)
	}
	if err := rows.Err(); err != nil {
		return CashBook{}, fmt.Errorf("iterate cash transactions: %w", err)
	}
	book.Balance = runningBalance
	return book, nil
}

func (s *Service) RecordVerifiedPayment(ctx context.Context, tx pgx.Tx, principal *auth.Principal, paymentID, paymentNumber string, amount float64, paidAt time.Time) error {
	transactionID := newUUID()
	transactionNumber := fmt.Sprintf("KAS-%s-%s", s.now().Format("0601"), transactionID[:8])
	_, err := tx.Exec(ctx, `
		INSERT INTO cash_transactions (
			id, organization_id, transaction_number, type, amount, transaction_date,
			description, reference_type, reference_id, created_by, created_at, updated_at
		) VALUES ($1, $2, $3, 'income', $4, $5, $6, 'payment', $7, $8, $9, $9)`,
		transactionID, principal.OrganizationID, transactionNumber, amount,
		paidAt.Format(time.DateOnly), fmt.Sprintf("Pembayaran %s", paymentNumber),
		paymentID, principal.UserID, s.now(),
	)
	if err != nil {
		return mapDatabaseError(err, "record verified payment in cash book")
	}
	return s.audit(ctx, tx, principal, "cash.payment.record", transactionID)
}

func (s *Service) audit(ctx context.Context, tx pgx.Tx, principal *auth.Principal, action, entityID string) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (organization_id, actor_user_id, action, entity_type, entity_id)
		VALUES ($1, $2, $3, 'cash_transaction', $4)`,
		principal.OrganizationID, principal.UserID, action, entityID,
	); err != nil {
		return fmt.Errorf("audit %s: %w", action, err)
	}
	return nil
}

func mapDatabaseError(err error, operation string) error {
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "unique"):
		return ErrDuplicateData
	case strings.Contains(message, "check"), strings.Contains(message, "foreign key"):
		return ErrConstraint
	default:
		return fmt.Errorf("%s: %w", operation, err)
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
