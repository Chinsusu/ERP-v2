package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	financedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	apperrors "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/errors"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

var ErrCashTransactionNotFound = errors.New("cash transaction not found")

const (
	ErrorCodeCashTransactionNotFound     response.ErrorCode = "CASH_TRANSACTION_NOT_FOUND"
	ErrorCodeCashTransactionValidation   response.ErrorCode = "CASH_TRANSACTION_VALIDATION_ERROR"
	ErrorCodeCashTransactionInvalidState response.ErrorCode = "CASH_TRANSACTION_INVALID_STATE"
)

type CashTransactionStore interface {
	List(ctx context.Context, filter CashTransactionFilter) ([]financedomain.CashTransaction, error)
	Get(ctx context.Context, id string) (financedomain.CashTransaction, error)
	Save(ctx context.Context, transaction financedomain.CashTransaction) error
}

type CashTransactionService struct {
	store    CashTransactionStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type CashTransactionFilter struct {
	Search         string
	Directions     []financedomain.CashTransactionDirection
	Statuses       []financedomain.CashTransactionStatus
	CounterpartyID string
}

type CashTransactionAllocationInput struct {
	ID         string
	TargetType string
	TargetID   string
	TargetNo   string
	Amount     string
}

type CreateCashTransactionInput struct {
	ID               string
	OrgID            string
	TransactionNo    string
	Direction        string
	BusinessDate     string
	CounterpartyID   string
	CounterpartyName string
	PaymentMethod    string
	ReferenceNo      string
	Allocations      []CashTransactionAllocationInput
	TotalAmount      string
	CurrencyCode     string
	Memo             string
	ActorID          string
	RequestID        string
}

type CashTransactionResult struct {
	CashTransaction financedomain.CashTransaction
	AuditLogID      string
}

type PrototypeCashTransactionStore struct {
	mu      sync.RWMutex
	records map[string]financedomain.CashTransaction
}

func NewCashTransactionService(
	store CashTransactionStore,
	auditLog audit.LogStore,
) CashTransactionService {
	return CashTransactionService{
		store:    store,
		auditLog: auditLog,
		clock:    func() time.Time { return time.Now().UTC() },
	}
}

func (s CashTransactionService) WithClock(clock func() time.Time) CashTransactionService {
	if clock != nil {
		s.clock = clock
	}

	return s
}

func NewPrototypeCashTransactionStore() *PrototypeCashTransactionStore {
	store := &PrototypeCashTransactionStore{records: make(map[string]financedomain.CashTransaction)}
	for _, transaction := range prototypeCashTransactions() {
		store.records[transaction.ID] = transaction.Clone()
	}

	return store
}

func (s CashTransactionService) ListCashTransactions(
	ctx context.Context,
	filter CashTransactionFilter,
) ([]financedomain.CashTransaction, error) {
	if s.store == nil {
		return nil, errors.New("cash transaction store is required")
	}

	return s.store.List(ctx, filter)
}

func (s CashTransactionService) GetCashTransaction(
	ctx context.Context,
	id string,
) (financedomain.CashTransaction, error) {
	if s.store == nil {
		return financedomain.CashTransaction{}, errors.New("cash transaction store is required")
	}
	transaction, err := s.store.Get(ctx, id)
	if err != nil {
		return financedomain.CashTransaction{}, mapCashTransactionError(
			err,
			map[string]any{"cash_transaction_id": strings.TrimSpace(id)},
		)
	}

	return transaction, nil
}

func (s CashTransactionService) CreateCashTransaction(
	ctx context.Context,
	input CreateCashTransactionInput,
) (CashTransactionResult, error) {
	if s.store == nil {
		return CashTransactionResult{}, errors.New("cash transaction store is required")
	}
	if s.auditLog == nil {
		return CashTransactionResult{}, errors.New("audit log store is required")
	}
	now := s.clock().UTC()
	id := firstNonBlank(input.ID, newCashTransactionID(now))
	transactionNo := firstNonBlank(input.TransactionNo, newCashTransactionNo(input.Direction, now))
	businessDate, err := parseRequiredCashTransactionDate(input.BusinessDate)
	if err != nil {
		return CashTransactionResult{}, cashTransactionValidationError(err, map[string]any{"field": "business_date"})
	}

	transaction, err := financedomain.NewCashTransaction(financedomain.NewCashTransactionInput{
		ID:               id,
		OrgID:            firstNonBlank(input.OrgID, defaultFinanceOrgID),
		TransactionNo:    transactionNo,
		Direction:        financedomain.CashTransactionDirection(input.Direction),
		BusinessDate:     businessDate,
		CounterpartyID:   input.CounterpartyID,
		CounterpartyName: input.CounterpartyName,
		PaymentMethod:    input.PaymentMethod,
		ReferenceNo:      input.ReferenceNo,
		Allocations:      cashTransactionAllocationsFromInput(input.Allocations),
		TotalAmount:      input.TotalAmount,
		CurrencyCode:     input.CurrencyCode,
		Memo:             input.Memo,
		CreatedAt:        now,
		CreatedBy:        input.ActorID,
		UpdatedAt:        now,
		UpdatedBy:        input.ActorID,
	})
	if err != nil {
		return CashTransactionResult{}, mapCashTransactionError(err, map[string]any{"cash_transaction_id": id})
	}
	transaction, err = transaction.Post(input.ActorID, now)
	if err != nil {
		return CashTransactionResult{}, mapCashTransactionError(err, map[string]any{"cash_transaction_id": id})
	}
	if err := s.store.Save(ctx, transaction); err != nil {
		return CashTransactionResult{}, err
	}
	auditLogID, err := s.recordCashTransactionAudit(
		ctx,
		transaction,
		financedomain.FinanceAuditActionCashTransactionRecorded,
		input.ActorID,
		input.RequestID,
		nil,
		cashTransactionAuditData(transaction),
		map[string]any{"direction": string(transaction.Direction), "allocation_count": len(transaction.Allocations)},
		now,
	)
	if err != nil {
		return CashTransactionResult{}, err
	}

	return CashTransactionResult{CashTransaction: transaction, AuditLogID: auditLogID}, nil
}

func (s *PrototypeCashTransactionStore) List(
	_ context.Context,
	filter CashTransactionFilter,
) ([]financedomain.CashTransaction, error) {
	if s == nil {
		return nil, errors.New("cash transaction store is required")
	}
	filter = normalizeCashTransactionFilter(filter)
	s.mu.RLock()
	defer s.mu.RUnlock()

	transactions := make([]financedomain.CashTransaction, 0, len(s.records))
	for _, transaction := range s.records {
		if !matchesCashTransactionFilter(transaction, filter) {
			continue
		}
		transactions = append(transactions, transaction.Clone())
	}
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].CreatedAt.After(transactions[j].CreatedAt)
	})

	return transactions, nil
}

func (s *PrototypeCashTransactionStore) Get(
	_ context.Context,
	id string,
) (financedomain.CashTransaction, error) {
	if s == nil {
		return financedomain.CashTransaction{}, errors.New("cash transaction store is required")
	}
	id = strings.TrimSpace(id)
	s.mu.RLock()
	defer s.mu.RUnlock()
	transaction, ok := s.records[id]
	if !ok {
		return financedomain.CashTransaction{}, ErrCashTransactionNotFound
	}

	return transaction.Clone(), nil
}

func (s *PrototypeCashTransactionStore) Save(
	_ context.Context,
	transaction financedomain.CashTransaction,
) error {
	if s == nil {
		return errors.New("cash transaction store is required")
	}
	if err := transaction.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[transaction.ID] = transaction.Clone()

	return nil
}

func (s CashTransactionService) recordCashTransactionAudit(
	ctx context.Context,
	transaction financedomain.CashTransaction,
	action financedomain.FinanceAuditAction,
	actorID string,
	requestID string,
	before map[string]any,
	after map[string]any,
	metadata map[string]any,
	createdAt time.Time,
) (string, error) {
	log, err := audit.NewLog(audit.NewLogInput{
		ID:         fmt.Sprintf("audit_%s_%d", strings.ReplaceAll(transaction.ID, "-", "_"), createdAt.UnixNano()),
		OrgID:      transaction.OrgID,
		ActorID:    strings.TrimSpace(actorID),
		Action:     string(action),
		EntityType: string(financedomain.FinanceEntityTypeCashTransaction),
		EntityID:   transaction.ID,
		RequestID:  strings.TrimSpace(requestID),
		BeforeData: before,
		AfterData:  after,
		Metadata:   metadata,
		CreatedAt:  createdAt,
	})
	if err != nil {
		return "", err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return "", err
	}

	return log.ID, nil
}

func cashTransactionAllocationsFromInput(
	inputs []CashTransactionAllocationInput,
) []financedomain.NewCashTransactionAllocationInput {
	allocations := make([]financedomain.NewCashTransactionAllocationInput, 0, len(inputs))
	for _, input := range inputs {
		allocations = append(allocations, financedomain.NewCashTransactionAllocationInput{
			ID:         input.ID,
			TargetType: financedomain.CashAllocationTargetType(input.TargetType),
			TargetID:   input.TargetID,
			TargetNo:   input.TargetNo,
			Amount:     input.Amount,
		})
	}

	return allocations
}

func normalizeCashTransactionFilter(filter CashTransactionFilter) CashTransactionFilter {
	filter.Search = strings.ToLower(strings.TrimSpace(filter.Search))
	filter.CounterpartyID = strings.TrimSpace(filter.CounterpartyID)
	directions := make([]financedomain.CashTransactionDirection, 0, len(filter.Directions))
	for _, direction := range filter.Directions {
		normalized := financedomain.NormalizeCashTransactionDirection(direction)
		if normalized != "" {
			directions = append(directions, normalized)
		}
	}
	filter.Directions = directions
	statuses := make([]financedomain.CashTransactionStatus, 0, len(filter.Statuses))
	for _, status := range filter.Statuses {
		normalized := financedomain.NormalizeCashTransactionStatus(status)
		if normalized != "" {
			statuses = append(statuses, normalized)
		}
	}
	filter.Statuses = statuses

	return filter
}

func matchesCashTransactionFilter(
	transaction financedomain.CashTransaction,
	filter CashTransactionFilter,
) bool {
	if len(filter.Directions) > 0 {
		matched := false
		for _, direction := range filter.Directions {
			if financedomain.NormalizeCashTransactionDirection(transaction.Direction) == direction {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if len(filter.Statuses) > 0 {
		matched := false
		for _, status := range filter.Statuses {
			if financedomain.NormalizeCashTransactionStatus(transaction.Status) == status {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if filter.CounterpartyID != "" && transaction.CounterpartyID != filter.CounterpartyID {
		return false
	}
	if filter.Search == "" {
		return true
	}
	haystack := strings.ToLower(strings.Join([]string{
		transaction.ID,
		transaction.TransactionNo,
		transaction.CounterpartyID,
		transaction.CounterpartyName,
		transaction.ReferenceNo,
	}, " "))
	if strings.Contains(haystack, filter.Search) {
		return true
	}
	for _, allocation := range transaction.Allocations {
		if strings.Contains(strings.ToLower(strings.Join([]string{
			string(allocation.TargetType),
			allocation.TargetID,
			allocation.TargetNo,
		}, " ")), filter.Search) {
			return true
		}
	}

	return false
}

func cashTransactionAuditData(transaction financedomain.CashTransaction) map[string]any {
	return map[string]any{
		"status":            string(transaction.Status),
		"direction":         string(transaction.Direction),
		"counterparty_id":   transaction.CounterpartyID,
		"counterparty_name": transaction.CounterpartyName,
		"total_amount":      transaction.TotalAmount.String(),
		"currency_code":     transaction.CurrencyCode.String(),
		"payment_method":    transaction.PaymentMethod,
		"reference_no":      transaction.ReferenceNo,
		"allocation_count":  len(transaction.Allocations),
		"version":           transaction.Version,
	}
}

func parseRequiredCashTransactionDate(value string) (time.Time, error) {
	parsed, err := parseOptionalFinanceDate(value)
	if err != nil {
		return time.Time{}, err
	}
	if parsed.IsZero() {
		return time.Time{}, financedomain.ErrCashTransactionRequiredField
	}

	return parsed, nil
}

func newCashTransactionID(now time.Time) string {
	return fmt.Sprintf("cash-%s-%d", now.Format("060102150405"), now.UnixNano()%100000)
}

func newCashTransactionNo(direction string, now time.Time) string {
	prefix := "CASH"
	switch financedomain.NormalizeCashTransactionDirection(financedomain.CashTransactionDirection(direction)) {
	case financedomain.CashTransactionDirectionIn:
		prefix = "CASH-IN"
	case financedomain.CashTransactionDirectionOut:
		prefix = "CASH-OUT"
	}

	return fmt.Sprintf("%s-%s-%05d", prefix, now.Format("060102-150405"), now.UnixNano()%100000)
}

func cashTransactionValidationError(cause error, details map[string]any) error {
	return apperrors.BadRequest(
		ErrorCodeCashTransactionValidation,
		"Cash transaction request is invalid",
		cause,
		details,
	)
}

func mapCashTransactionError(err error, details map[string]any) error {
	if err == nil {
		return nil
	}
	if _, ok := apperrors.As(err); ok {
		return err
	}
	if errors.Is(err, ErrCashTransactionNotFound) {
		return apperrors.NotFound(
			ErrorCodeCashTransactionNotFound,
			"Cash transaction not found",
			err,
			details,
		)
	}
	if errors.Is(err, financedomain.ErrCashTransactionInvalidTransition) ||
		errors.Is(err, financedomain.ErrCashTransactionInvalidStatus) {
		return apperrors.Conflict(
			ErrorCodeCashTransactionInvalidState,
			"Cash transaction state is invalid",
			err,
			details,
		)
	}
	if errors.Is(err, financedomain.ErrCashTransactionRequiredField) ||
		errors.Is(err, financedomain.ErrCashTransactionInvalidDirection) ||
		errors.Is(err, financedomain.ErrCashTransactionInvalidAmount) ||
		errors.Is(err, financedomain.ErrCashTransactionInvalidAllocation) ||
		errors.Is(err, financedomain.ErrFinanceInvalidCurrency) ||
		errors.Is(err, financedomain.ErrFinanceInvalidMoneyAmount) {
		return cashTransactionValidationError(err, details)
	}

	return err
}

func prototypeCashTransactions() []financedomain.CashTransaction {
	receipt, err := financedomain.NewCashTransaction(financedomain.NewCashTransactionInput{
		ID:               "cash-in-260430-0001",
		OrgID:            defaultFinanceOrgID,
		TransactionNo:    "CASH-IN-260430-0001",
		Direction:        financedomain.CashTransactionDirectionIn,
		BusinessDate:     time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
		CounterpartyID:   "carrier-ghn",
		CounterpartyName: "GHN COD",
		PaymentMethod:    "bank_transfer",
		ReferenceNo:      "BANK-COD-260430-0001",
		TotalAmount:      "1250000.00",
		CurrencyCode:     "VND",
		CreatedAt:        time.Date(2026, 4, 30, 9, 30, 0, 0, time.UTC),
		CreatedBy:        "finance-seed",
		Allocations: []financedomain.NewCashTransactionAllocationInput{
			{
				ID:         "cash-in-260430-0001-line-1",
				TargetType: financedomain.CashAllocationTargetCustomerReceivable,
				TargetID:   "ar-cod-260430-0001",
				TargetNo:   "AR-COD-260430-0001",
				Amount:     "1000000.00",
			},
			{
				ID:         "cash-in-260430-0001-line-2",
				TargetType: financedomain.CashAllocationTargetCODRemittance,
				TargetID:   "cod-remit-260430-0001",
				TargetNo:   "COD-REMIT-260430-0001",
				Amount:     "250000.00",
			},
		},
	})
	if err != nil {
		return nil
	}
	receipt, err = receipt.Post("finance-seed", receipt.CreatedAt)
	if err != nil {
		return nil
	}
	payment, err := financedomain.NewCashTransaction(financedomain.NewCashTransactionInput{
		ID:               "cash-out-260430-0002",
		OrgID:            defaultFinanceOrgID,
		TransactionNo:    "CASH-OUT-260430-0002",
		Direction:        financedomain.CashTransactionDirectionOut,
		BusinessDate:     time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
		CounterpartyID:   "supplier-hcm-001",
		CounterpartyName: "Nguyen Lieu HCM",
		PaymentMethod:    "bank_transfer",
		ReferenceNo:      "BANK-AP-260430-0002",
		TotalAmount:      "4250000.00",
		CurrencyCode:     "VND",
		CreatedAt:        time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC),
		CreatedBy:        "finance-seed",
		Allocations: []financedomain.NewCashTransactionAllocationInput{
			{
				ID:         "cash-out-260430-0002-line-1",
				TargetType: financedomain.CashAllocationTargetSupplierPayable,
				TargetID:   "ap-supplier-260430-0001",
				TargetNo:   "AP-SUP-260430-0001",
				Amount:     "4250000.00",
			},
		},
	})
	if err != nil {
		return nil
	}
	payment, err = payment.Post("finance-seed", payment.CreatedAt)
	if err != nil {
		return nil
	}

	return []financedomain.CashTransaction{receipt, payment}
}
