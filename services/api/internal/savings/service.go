package savings

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

func (s *Service) CreateProduct(ctx context.Context, principal *auth.Principal, req CreateProductReq) (Product, error) {
	if principal == nil || req.Validate() != nil {
		return Product{}, ErrValidation
	}

	now := s.now()
	prod := Product{}
	err := s.db.QueryRow(ctx, `
		INSERT INTO savings_products (id, organization_id, code, name, description, minimum_deposit, withdrawal_rule, status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'active', $8, $9, $9)
		RETURNING id, code, name, description, minimum_deposit, withdrawal_rule, status, created_by, created_at, updated_at`,
		newUUID(), principal.OrganizationID, req.Code, req.Name, req.Description, req.MinimumDeposit, req.WithdrawalRule, principal.UserID, now,
	).Scan(
		&prod.ID, &prod.Code, &prod.Name, &prod.Description, &prod.MinimumDeposit, &prod.WithdrawalRule, &prod.Status, &prod.CreatedBy, &prod.CreatedAt, &prod.UpdatedAt,
	)
	if err != nil {
		return Product{}, mapDBErr(err)
	}
	return prod, nil
}

func (s *Service) ListProducts(ctx context.Context, principal *auth.Principal, status string) ([]Product, error) {
	if principal == nil {
		return nil, ErrValidation
	}

	query := `SELECT id, code, name, description, minimum_deposit, withdrawal_rule, status, created_by, created_at, updated_at FROM savings_products WHERE organization_id = $1`
	args := []any{principal.OrganizationID}

	if status != "" {
		query += ` AND status = $2`
		args = append(args, status)
	}
	query += ` ORDER BY name ASC`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.Description, &p.MinimumDeposit, &p.WithdrawalRule, &p.Status, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

func (s *Service) GetProduct(ctx context.Context, principal *auth.Principal, id string) (Product, error) {
	if principal == nil || id == "" {
		return Product{}, ErrValidation
	}

	var p Product
	err := s.db.QueryRow(ctx, `
		SELECT id, code, name, description, minimum_deposit, withdrawal_rule, status, created_by, created_at, updated_at
		FROM savings_products WHERE id = $1 AND organization_id = $2`,
		id, principal.OrganizationID,
	).Scan(&p.ID, &p.Code, &p.Name, &p.Description, &p.MinimumDeposit, &p.WithdrawalRule, &p.Status, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, ErrProductNotFound
	}
	return p, err
}

func (s *Service) UpdateProduct(ctx context.Context, principal *auth.Principal, id string, req UpdateProductReq) (Product, error) {
	if principal == nil || id == "" || req.Validate() != nil {
		return Product{}, ErrValidation
	}

	now := s.now()
	var p Product
	err := s.db.QueryRow(ctx, `
		UPDATE savings_products
		SET name = COALESCE($1, name),
		    description = COALESCE($2, description),
		    minimum_deposit = COALESCE($3, minimum_deposit),
		    withdrawal_rule = COALESCE($4, withdrawal_rule),
		    status = COALESCE($5, status),
		    updated_at = $6
		WHERE id = $7 AND organization_id = $8
		RETURNING id, code, name, description, minimum_deposit, withdrawal_rule, status, created_by, created_at, updated_at`,
		req.Name, req.Description, req.MinimumDeposit, req.WithdrawalRule, req.Status, now, id, principal.OrganizationID,
	).Scan(&p.ID, &p.Code, &p.Name, &p.Description, &p.MinimumDeposit, &p.WithdrawalRule, &p.Status, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, ErrProductNotFound
	}
	if err != nil {
		return Product{}, mapDBErr(err)
	}
	return p, nil
}

func (s *Service) CreateAccount(ctx context.Context, principal *auth.Principal, req CreateAccountReq) (Account, error) {
	if principal == nil || req.Validate() != nil {
		return Account{}, ErrValidation
	}

	var pStatus string
	err := s.db.QueryRow(ctx, `SELECT status FROM savings_products WHERE id = $1 AND organization_id = $2`, req.ProductID, principal.OrganizationID).Scan(&pStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrProductNotFound
	}
	if pStatus != "active" {
		return Account{}, ErrInvalidState
	}

	now := s.now()
	accNo := fmt.Sprintf("SAV-%d-%s", now.UnixNano()%1000000, randHex(3))
	var acc Account
	err = s.db.QueryRow(ctx, `
		INSERT INTO savings_accounts (id, organization_id, product_id, household_id, account_number, balance, status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 0, 'active', $6, $7, $7)
		RETURNING id, product_id, household_id, account_number, balance, status, created_by, created_at, updated_at`,
		newUUID(), principal.OrganizationID, req.ProductID, req.HouseholdID, accNo, principal.UserID, now,
	).Scan(&acc.ID, &acc.ProductID, &acc.HouseholdID, &acc.AccountNumber, &acc.Balance, &acc.Status, &acc.CreatedBy, &acc.CreatedAt, &acc.UpdatedAt)

	if err != nil {
		return Account{}, mapDBErr(err)
	}
	return acc, nil
}


func (s *Service) ListAccounts(ctx context.Context, principal *auth.Principal, filter AccountFilter) ([]Account, error) {
	if principal == nil {
		return nil, ErrValidation
	}

	query := `
		SELECT a.id, a.product_id, p.name, a.household_id, a.account_number, a.balance, a.status, a.created_by, a.created_at, a.updated_at
		FROM savings_accounts a
		JOIN savings_products p ON p.id = a.product_id AND p.organization_id = a.organization_id
		WHERE a.organization_id = $1`
	args := []any{principal.OrganizationID}

	if filter.ProductID != "" {
		args = append(args, filter.ProductID)
		query += fmt.Sprintf(" AND a.product_id = $%d", len(args))
	}
	if filter.HouseholdID != "" {
		args = append(args, filter.HouseholdID)
		query += fmt.Sprintf(" AND a.household_id = $%d", len(args))
	}
	if filter.Status != "" {
		args = append(args, filter.Status)
		query += fmt.Sprintf(" AND a.status = $%d", len(args))
	}
	query += " ORDER BY a.created_at DESC"

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var a Account
		if err := rows.Scan(&a.ID, &a.ProductID, &a.ProductName, &a.HouseholdID, &a.AccountNumber, &a.Balance, &a.Status, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

func (s *Service) GetAccount(ctx context.Context, principal *auth.Principal, id string) (Account, error) {
	if principal == nil || id == "" {
		return Account{}, ErrValidation
	}

	var a Account
	err := s.db.QueryRow(ctx, `
		SELECT a.id, a.product_id, p.name, a.household_id, a.account_number, a.balance, a.status, a.created_by, a.created_at, a.updated_at
		FROM savings_accounts a
		JOIN savings_products p ON p.id = a.product_id AND p.organization_id = a.organization_id
		WHERE a.id = $1 AND a.organization_id = $2`,
		id, principal.OrganizationID,
	).Scan(&a.ID, &a.ProductID, &a.ProductName, &a.HouseholdID, &a.AccountNumber, &a.Balance, &a.Status, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrAccountNotFound
	}
	return a, err
}

func (s *Service) CloseAccount(ctx context.Context, principal *auth.Principal, id string) (Account, error) {
	if principal == nil || id == "" {
		return Account{}, ErrValidation
	}

	var a Account
	now := s.now()
	err := s.db.QueryRow(ctx, `
		UPDATE savings_accounts
		SET status = 'closed', updated_at = $1
		WHERE id = $2 AND organization_id = $3 AND status = 'active' AND balance = 0
		RETURNING id, product_id, household_id, account_number, balance, status, created_by, created_at, updated_at`,
		now, id, principal.OrganizationID,
	).Scan(&a.ID, &a.ProductID, &a.HouseholdID, &a.AccountNumber, &a.Balance, &a.Status, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrConstraint
	}
	return a, err
}

func (s *Service) Deposit(ctx context.Context, principal *auth.Principal, req DepositReq) (Transaction, error) {
	if principal == nil || req.Validate() != nil {
		return Transaction{}, ErrValidation
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Transaction{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var accStatus string
	var minDep float64
	err = tx.QueryRow(ctx, `
		SELECT a.status, p.minimum_deposit
		FROM savings_accounts a
		JOIN savings_products p ON p.id = a.product_id AND p.organization_id = a.organization_id
		WHERE a.id = $1 AND a.organization_id = $2`,
		req.AccountID, principal.OrganizationID,
	).Scan(&accStatus, &minDep)

	if errors.Is(err, pgx.ErrNoRows) {
		return Transaction{}, ErrAccountNotFound
	}
	if accStatus != "active" {
		return Transaction{}, ErrInvalidState
	}
	if req.Amount < minDep {
		return Transaction{}, ErrValidation
	}

	now := s.now()
	txNo := fmt.Sprintf("TXS-%d-%s", now.UnixNano()%1000000, randHex(3))
	var resTx Transaction
	err = tx.QueryRow(ctx, `
		INSERT INTO savings_transactions (
			id, organization_id, account_id, transaction_number, type, amount, balance_after,
			transaction_date, description, proof_file_id, verification_status, idempotency_key,
			created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, 'deposit', $5, 0, $6, $7, $8, 'pending', $9, $10, $11, $11
		) RETURNING id, account_id, transaction_number, type, amount, balance_after, transaction_date::text,
		            description, proof_file_id, reversal_of_id, verification_status, verified_by, verified_at,
		            rejection_reason, idempotency_key, created_by, created_at, updated_at`,
		newUUID(), principal.OrganizationID, req.AccountID, txNo, req.Amount, req.TransactionDate,
		req.Description, req.ProofFileID, req.IdempotencyKey, principal.UserID, now,
	).Scan(
		&resTx.ID, &resTx.AccountID, &resTx.TransactionNumber, &resTx.Type, &resTx.Amount, &resTx.BalanceAfter,
		&resTx.TransactionDate, &resTx.Description, &resTx.ProofFileID, &resTx.ReversalOfID, &resTx.VerificationStatus,
		&resTx.VerifiedBy, &resTx.VerifiedAt, &resTx.RejectionReason, &resTx.IdempotencyKey, &resTx.CreatedBy,
		&resTx.CreatedAt, &resTx.UpdatedAt,
	)

	if err != nil {
		return Transaction{}, mapDBErr(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Transaction{}, err
	}
	return resTx, nil
}

func (s *Service) Withdraw(ctx context.Context, principal *auth.Principal, req WithdrawReq) (Transaction, error) {
	if principal == nil || req.Validate() != nil {
		return Transaction{}, ErrValidation
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Transaction{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var accStatus string
	var currentBalance float64
	err = tx.QueryRow(ctx, `
		SELECT status, balance
		FROM savings_accounts
		WHERE id = $1 AND organization_id = $2`,
		req.AccountID, principal.OrganizationID,
	).Scan(&accStatus, &currentBalance)

	if errors.Is(err, pgx.ErrNoRows) {
		return Transaction{}, ErrAccountNotFound
	}
	if accStatus != "active" {
		return Transaction{}, ErrInvalidState
	}
	if currentBalance < req.Amount {
		return Transaction{}, ErrInsufficientBalance
	}

	now := s.now()
	txNo := fmt.Sprintf("TXS-%d-%s", now.UnixNano()%1000000, randHex(3))
	var resTx Transaction
	err = tx.QueryRow(ctx, `
		INSERT INTO savings_transactions (
			id, organization_id, account_id, transaction_number, type, amount, balance_after,
			transaction_date, description, proof_file_id, verification_status, idempotency_key,
			created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, 'withdrawal', $5, 0, $6, $7, $8, 'pending', $9, $10, $11, $11
		) RETURNING id, account_id, transaction_number, type, amount, balance_after, transaction_date::text,
		            description, proof_file_id, reversal_of_id, verification_status, verified_by, verified_at,
		            rejection_reason, idempotency_key, created_by, created_at, updated_at`,
		newUUID(), principal.OrganizationID, req.AccountID, txNo, req.Amount, req.TransactionDate,
		req.Description, req.ProofFileID, req.IdempotencyKey, principal.UserID, now,
	).Scan(
		&resTx.ID, &resTx.AccountID, &resTx.TransactionNumber, &resTx.Type, &resTx.Amount, &resTx.BalanceAfter,
		&resTx.TransactionDate, &resTx.Description, &resTx.ProofFileID, &resTx.ReversalOfID, &resTx.VerificationStatus,
		&resTx.VerifiedBy, &resTx.VerifiedAt, &resTx.RejectionReason, &resTx.IdempotencyKey, &resTx.CreatedBy,
		&resTx.CreatedAt, &resTx.UpdatedAt,
	)

	if err != nil {
		return Transaction{}, mapDBErr(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Transaction{}, err
	}
	return resTx, nil
}

func (s *Service) VerifyTransaction(ctx context.Context, principal *auth.Principal, txID string, req VerifyTxReq) (Transaction, error) {
	if principal == nil || txID == "" || req.Validate() != nil {
		return Transaction{}, ErrValidation
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Transaction{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var existingTx struct {
		ID                 string
		AccountID          string
		Type               string
		Amount             float64
		VerificationStatus string
		CreatedBy          string
	}
	err = tx.QueryRow(ctx, `
		SELECT id, account_id, type, amount, verification_status, created_by
		FROM savings_transactions
		WHERE id = $1 AND organization_id = $2 FOR UPDATE`,
		txID, principal.OrganizationID,
	).Scan(&existingTx.ID, &existingTx.AccountID, &existingTx.Type, &existingTx.Amount, &existingTx.VerificationStatus, &existingTx.CreatedBy)

	if errors.Is(err, pgx.ErrNoRows) {
		return Transaction{}, ErrTransactionNotFound
	}
	if existingTx.VerificationStatus != "pending" {
		return Transaction{}, ErrInvalidState
	}
	if existingTx.CreatedBy == principal.UserID {
		return Transaction{}, ErrForbidden
	}

	now := s.now()
	var newBal float64

	if req.Status == "verified" {
		var curBal float64
		err = tx.QueryRow(ctx, `SELECT balance FROM savings_accounts WHERE id = $1 AND organization_id = $2 FOR UPDATE`, existingTx.AccountID, principal.OrganizationID).Scan(&curBal)
		if err != nil {
			return Transaction{}, err
		}

		if existingTx.Type == "deposit" {
			newBal = curBal + existingTx.Amount
		} else if existingTx.Type == "withdrawal" {
			if curBal < existingTx.Amount {
				return Transaction{}, ErrInsufficientBalance
			}
			newBal = curBal - existingTx.Amount
		}

		_, err = tx.Exec(ctx, `UPDATE savings_accounts SET balance = $1, updated_at = $2 WHERE id = $3 AND organization_id = $4`, newBal, now, existingTx.AccountID, principal.OrganizationID)
		if err != nil {
			return Transaction{}, err
		}
	}

	var resTx Transaction
	err = tx.QueryRow(ctx, `
		UPDATE savings_transactions
		SET verification_status = $1,
		    balance_after = $2,
		    verified_by = $3,
		    verified_at = $4,
		    rejection_reason = $5,
		    updated_at = $4
		WHERE id = $6 AND organization_id = $7
		RETURNING id, account_id, transaction_number, type, amount, balance_after, transaction_date::text,
		          description, proof_file_id, reversal_of_id, verification_status, verified_by, verified_at,
		          rejection_reason, idempotency_key, created_by, created_at, updated_at`,
		req.Status, newBal, principal.UserID, now, req.RejectionReason, txID, principal.OrganizationID,
	).Scan(
		&resTx.ID, &resTx.AccountID, &resTx.TransactionNumber, &resTx.Type, &resTx.Amount, &resTx.BalanceAfter,
		&resTx.TransactionDate, &resTx.Description, &resTx.ProofFileID, &resTx.ReversalOfID, &resTx.VerificationStatus,
		&resTx.VerifiedBy, &resTx.VerifiedAt, &resTx.RejectionReason, &resTx.IdempotencyKey, &resTx.CreatedBy,
		&resTx.CreatedAt, &resTx.UpdatedAt,
	)

	if err != nil {
		return Transaction{}, mapDBErr(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Transaction{}, err
	}
	return resTx, nil
}

func (s *Service) ReverseTransaction(ctx context.Context, principal *auth.Principal, targetTxID string, description string) (Transaction, error) {
	if principal == nil || targetTxID == "" || strings.TrimSpace(description) == "" {
		return Transaction{}, ErrValidation
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Transaction{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var target struct {
		ID                 string
		AccountID          string
		Type               string
		Amount             float64
		VerificationStatus string
	}
	err = tx.QueryRow(ctx, `
		SELECT id, account_id, type, amount, verification_status
		FROM savings_transactions
		WHERE id = $1 AND organization_id = $2 FOR UPDATE`,
		targetTxID, principal.OrganizationID,
	).Scan(&target.ID, &target.AccountID, &target.Type, &target.Amount, &target.VerificationStatus)

	if errors.Is(err, pgx.ErrNoRows) {
		return Transaction{}, ErrTransactionNotFound
	}
	if target.VerificationStatus != "verified" {
		return Transaction{}, ErrInvalidState
	}

	var revExists bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM savings_transactions WHERE reversal_of_id = $1 AND organization_id = $2)`, targetTxID, principal.OrganizationID).Scan(&revExists)
	if err != nil {
		return Transaction{}, err
	}
	if revExists {
		return Transaction{}, ErrConstraint
	}

	var curBal float64
	err = tx.QueryRow(ctx, `SELECT balance FROM savings_accounts WHERE id = $1 AND organization_id = $2 FOR UPDATE`, target.AccountID, principal.OrganizationID).Scan(&curBal)
	if err != nil {
		return Transaction{}, err
	}

	var newBal float64
	if target.Type == "deposit" {
		if curBal < target.Amount {
			return Transaction{}, ErrInsufficientBalance
		}
		newBal = curBal - target.Amount
	} else if target.Type == "withdrawal" {
		newBal = curBal + target.Amount
	} else {
		return Transaction{}, ErrInvalidState
	}

	now := s.now()
	_, err = tx.Exec(ctx, `UPDATE savings_accounts SET balance = $1, updated_at = $2 WHERE id = $3 AND organization_id = $4`, newBal, now, target.AccountID, principal.OrganizationID)
	if err != nil {
		return Transaction{}, err
	}

	txNo := fmt.Sprintf("REV-%d-%s", now.UnixNano()%1000000, randHex(3))
	var resTx Transaction
	err = tx.QueryRow(ctx, `
		INSERT INTO savings_transactions (
			id, organization_id, account_id, transaction_number, type, amount, balance_after,
			transaction_date, description, reversal_of_id, verification_status, verified_by, verified_at,
			created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, 'reversal', $5, $6, $7, $8, $9, 'verified', $10, $11, $10, $11, $11
		) RETURNING id, account_id, transaction_number, type, amount, balance_after, transaction_date::text,
		            description, proof_file_id, reversal_of_id, verification_status, verified_by, verified_at,
		            rejection_reason, idempotency_key, created_by, created_at, updated_at`,
		newUUID(), principal.OrganizationID, target.AccountID, txNo, target.Amount, newBal,
		now.Format("2006-01-02"), description, target.ID, principal.UserID, now,
	).Scan(
		&resTx.ID, &resTx.AccountID, &resTx.TransactionNumber, &resTx.Type, &resTx.Amount, &resTx.BalanceAfter,
		&resTx.TransactionDate, &resTx.Description, &resTx.ProofFileID, &resTx.ReversalOfID, &resTx.VerificationStatus,
		&resTx.VerifiedBy, &resTx.VerifiedAt, &resTx.RejectionReason, &resTx.IdempotencyKey, &resTx.CreatedBy,
		&resTx.CreatedAt, &resTx.UpdatedAt,
	)

	if err != nil {
		return Transaction{}, mapDBErr(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Transaction{}, err
	}
	return resTx, nil
}

func (s *Service) ListTransactions(ctx context.Context, principal *auth.Principal, filter TransactionFilter) ([]Transaction, error) {
	if principal == nil {
		return nil, ErrValidation
	}

	query := `
		SELECT t.id, t.account_id, t.transaction_number, t.type, t.amount, t.balance_after,
		       t.transaction_date::text, t.description, t.proof_file_id, t.reversal_of_id,
		       t.verification_status, t.verified_by, t.verified_at, t.rejection_reason,
		       t.idempotency_key, t.created_by, t.created_at, t.updated_at
		FROM savings_transactions t
		JOIN savings_accounts a ON a.id = t.account_id AND a.organization_id = t.organization_id
		WHERE t.organization_id = $1`
	args := []any{principal.OrganizationID}

	if filter.AccountID != "" {
		args = append(args, filter.AccountID)
		query += fmt.Sprintf(" AND t.account_id = $%d", len(args))
	}
	if filter.ProductID != "" {
		args = append(args, filter.ProductID)
		query += fmt.Sprintf(" AND a.product_id = $%d", len(args))
	}
	if filter.HouseholdID != "" {
		args = append(args, filter.HouseholdID)
		query += fmt.Sprintf(" AND a.household_id = $%d", len(args))
	}
	if filter.Type != "" {
		args = append(args, filter.Type)
		query += fmt.Sprintf(" AND t.type = $%d", len(args))
	}
	if filter.VerificationStatus != "" {
		args = append(args, filter.VerificationStatus)
		query += fmt.Sprintf(" AND t.verification_status = $%d", len(args))
	}
	if filter.StartDate != "" {
		args = append(args, filter.StartDate)
		query += fmt.Sprintf(" AND t.transaction_date >= $%d", len(args))
	}
	if filter.EndDate != "" {
		args = append(args, filter.EndDate)
		query += fmt.Sprintf(" AND t.transaction_date <= $%d", len(args))
	}
	query += " ORDER BY t.transaction_date DESC, t.created_at DESC"

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(
			&t.ID, &t.AccountID, &t.TransactionNumber, &t.Type, &t.Amount, &t.BalanceAfter,
			&t.TransactionDate, &t.Description, &t.ProofFileID, &t.ReversalOfID, &t.VerificationStatus,
			&t.VerifiedBy, &t.VerifiedAt, &t.RejectionReason, &t.IdempotencyKey, &t.CreatedBy,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}
	return transactions, rows.Err()
}

func (s *Service) GetReconciliationReport(ctx context.Context, principal *auth.Principal) ([]ReconciliationReport, error) {
	if principal == nil {
		return nil, ErrValidation
	}

	rows, err := s.db.Query(ctx, `
		SELECT p.id, p.name,
		       COUNT(DISTINCT a.id) as total_accounts,
		       COALESCE(SUM(a.balance), 0) as total_verified_balance,
		       COALESCE(SUM(CASE WHEN t.type = 'deposit' AND t.verification_status = 'pending' THEN t.amount ELSE 0 END), 0) as pending_deposits,
		       COALESCE(SUM(CASE WHEN t.type = 'withdrawal' AND t.verification_status = 'pending' THEN t.amount ELSE 0 END), 0) as pending_withdrawals
		FROM savings_products p
		LEFT JOIN savings_accounts a ON a.product_id = p.id AND a.organization_id = p.organization_id AND a.status = 'active'
		LEFT JOIN savings_transactions t ON t.account_id = a.id AND t.organization_id = a.organization_id
		WHERE p.organization_id = $1
		GROUP BY p.id, p.name
		ORDER BY p.name ASC`,
		principal.OrganizationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []ReconciliationReport
	for rows.Next() {
		var r ReconciliationReport
		if err := rows.Scan(&r.ProductID, &r.ProductName, &r.TotalAccounts, &r.TotalVerifiedBalance, &r.PendingDeposits, &r.PendingWithdrawals); err != nil {
			return nil, err
		}
		reports = append(reports, r)
	}
	return reports, rows.Err()
}

func newUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func randHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func mapDBErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrDuplicateData
		case "23514", "23503":
			return ErrConstraint
		}
	}
	return err
}

