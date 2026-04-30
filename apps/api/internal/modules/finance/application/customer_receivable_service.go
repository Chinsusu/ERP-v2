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

var ErrCustomerReceivableNotFound = errors.New("customer receivable not found")

const (
	ErrorCodeCustomerReceivableNotFound     response.ErrorCode = "CUSTOMER_RECEIVABLE_NOT_FOUND"
	ErrorCodeCustomerReceivableValidation   response.ErrorCode = "CUSTOMER_RECEIVABLE_VALIDATION_ERROR"
	ErrorCodeCustomerReceivableInvalidState response.ErrorCode = "CUSTOMER_RECEIVABLE_INVALID_STATE"

	defaultFinanceOrgID = "org-my-pham"
)

type CustomerReceivableStore interface {
	List(ctx context.Context, filter CustomerReceivableFilter) ([]financedomain.CustomerReceivable, error)
	Get(ctx context.Context, id string) (financedomain.CustomerReceivable, error)
	Save(ctx context.Context, receivable financedomain.CustomerReceivable) error
}

type CustomerReceivableService struct {
	store    CustomerReceivableStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type CustomerReceivableFilter struct {
	Search     string
	Statuses   []financedomain.ReceivableStatus
	CustomerID string
}

type SourceDocumentInput struct {
	Type string
	ID   string
	No   string
}

type CustomerReceivableLineInput struct {
	ID             string
	Description    string
	SourceDocument SourceDocumentInput
	Amount         string
}

type CreateCustomerReceivableInput struct {
	ID             string
	OrgID          string
	ReceivableNo   string
	CustomerID     string
	CustomerCode   string
	CustomerName   string
	Status         string
	SourceDocument SourceDocumentInput
	Lines          []CustomerReceivableLineInput
	TotalAmount    string
	CurrencyCode   string
	DueDate        string
	ActorID        string
	RequestID      string
}

type CustomerReceivableActionInput struct {
	ID        string
	Amount    string
	Reason    string
	ActorID   string
	RequestID string
}

type CustomerReceivableResult struct {
	CustomerReceivable financedomain.CustomerReceivable
	AuditLogID         string
}

type CustomerReceivableActionResult struct {
	CustomerReceivable financedomain.CustomerReceivable
	PreviousStatus     financedomain.ReceivableStatus
	CurrentStatus      financedomain.ReceivableStatus
	AuditLogID         string
}

type PrototypeCustomerReceivableStore struct {
	mu      sync.RWMutex
	records map[string]financedomain.CustomerReceivable
}

func NewCustomerReceivableService(
	store CustomerReceivableStore,
	auditLog audit.LogStore,
) CustomerReceivableService {
	return CustomerReceivableService{
		store:    store,
		auditLog: auditLog,
		clock:    func() time.Time { return time.Now().UTC() },
	}
}

func (s CustomerReceivableService) WithClock(clock func() time.Time) CustomerReceivableService {
	if clock != nil {
		s.clock = clock
	}

	return s
}

func NewPrototypeCustomerReceivableStore() *PrototypeCustomerReceivableStore {
	store := &PrototypeCustomerReceivableStore{records: make(map[string]financedomain.CustomerReceivable)}
	for _, receivable := range prototypeCustomerReceivables() {
		store.records[receivable.ID] = receivable.Clone()
	}

	return store
}

func (s CustomerReceivableService) ListCustomerReceivables(
	ctx context.Context,
	filter CustomerReceivableFilter,
) ([]financedomain.CustomerReceivable, error) {
	if s.store == nil {
		return nil, errors.New("customer receivable store is required")
	}

	return s.store.List(ctx, filter)
}

func (s CustomerReceivableService) GetCustomerReceivable(
	ctx context.Context,
	id string,
) (financedomain.CustomerReceivable, error) {
	if s.store == nil {
		return financedomain.CustomerReceivable{}, errors.New("customer receivable store is required")
	}
	receivable, err := s.store.Get(ctx, id)
	if err != nil {
		return financedomain.CustomerReceivable{}, mapCustomerReceivableError(
			err,
			map[string]any{"customer_receivable_id": strings.TrimSpace(id)},
		)
	}

	return receivable, nil
}

func (s CustomerReceivableService) CreateCustomerReceivable(
	ctx context.Context,
	input CreateCustomerReceivableInput,
) (CustomerReceivableResult, error) {
	if s.store == nil {
		return CustomerReceivableResult{}, errors.New("customer receivable store is required")
	}
	if s.auditLog == nil {
		return CustomerReceivableResult{}, errors.New("audit log store is required")
	}
	now := s.clock().UTC()
	id := firstNonBlank(input.ID, newCustomerReceivableID(now))
	receivableNo := firstNonBlank(input.ReceivableNo, newCustomerReceivableNo(now))
	source, err := sourceDocumentFromInput(input.SourceDocument)
	if err != nil {
		return CustomerReceivableResult{}, mapCustomerReceivableError(err, nil)
	}
	lines, err := customerReceivableLinesFromInput(input.Lines)
	if err != nil {
		return CustomerReceivableResult{}, mapCustomerReceivableError(err, nil)
	}
	dueDate, err := parseOptionalFinanceDate(input.DueDate)
	if err != nil {
		return CustomerReceivableResult{}, customerReceivableValidationError(err, map[string]any{"field": "due_date"})
	}

	receivable, err := financedomain.NewCustomerReceivable(financedomain.NewCustomerReceivableInput{
		ID:             id,
		OrgID:          firstNonBlank(input.OrgID, defaultFinanceOrgID),
		ReceivableNo:   receivableNo,
		CustomerID:     input.CustomerID,
		CustomerCode:   input.CustomerCode,
		CustomerName:   input.CustomerName,
		Status:         financedomain.ReceivableStatus(input.Status),
		SourceDocument: source,
		Lines:          lines,
		TotalAmount:    input.TotalAmount,
		CurrencyCode:   input.CurrencyCode,
		DueDate:        dueDate,
		CreatedAt:      now,
		CreatedBy:      input.ActorID,
		UpdatedAt:      now,
		UpdatedBy:      input.ActorID,
	})
	if err != nil {
		return CustomerReceivableResult{}, mapCustomerReceivableError(err, map[string]any{"customer_receivable_id": id})
	}
	if err := s.store.Save(ctx, receivable); err != nil {
		return CustomerReceivableResult{}, err
	}
	auditLogID, err := s.recordCustomerReceivableAudit(
		ctx,
		receivable,
		financedomain.FinanceAuditActionReceivableCreated,
		input.ActorID,
		input.RequestID,
		nil,
		customerReceivableAuditData(receivable),
		receivable.SourceDocument.Metadata(),
		now,
	)
	if err != nil {
		return CustomerReceivableResult{}, err
	}

	return CustomerReceivableResult{CustomerReceivable: receivable, AuditLogID: auditLogID}, nil
}

func (s CustomerReceivableService) RecordCustomerReceivableReceipt(
	ctx context.Context,
	input CustomerReceivableActionInput,
) (CustomerReceivableActionResult, error) {
	return s.applyCustomerReceivableAction(ctx, input, "record-receipt")
}

func (s CustomerReceivableService) MarkCustomerReceivableDisputed(
	ctx context.Context,
	input CustomerReceivableActionInput,
) (CustomerReceivableActionResult, error) {
	return s.applyCustomerReceivableAction(ctx, input, "mark-disputed")
}

func (s CustomerReceivableService) VoidCustomerReceivable(
	ctx context.Context,
	input CustomerReceivableActionInput,
) (CustomerReceivableActionResult, error) {
	return s.applyCustomerReceivableAction(ctx, input, "void")
}

func (s CustomerReceivableService) applyCustomerReceivableAction(
	ctx context.Context,
	input CustomerReceivableActionInput,
	action string,
) (CustomerReceivableActionResult, error) {
	if s.store == nil {
		return CustomerReceivableActionResult{}, errors.New("customer receivable store is required")
	}
	if s.auditLog == nil {
		return CustomerReceivableActionResult{}, errors.New("audit log store is required")
	}
	current, err := s.store.Get(ctx, input.ID)
	if err != nil {
		return CustomerReceivableActionResult{}, mapCustomerReceivableError(
			err,
			map[string]any{"customer_receivable_id": strings.TrimSpace(input.ID)},
		)
	}

	now := s.clock().UTC()
	before := customerReceivableAuditData(current)
	var (
		updated     financedomain.CustomerReceivable
		auditAction financedomain.FinanceAuditAction
		metadata    map[string]any
	)
	switch action {
	case "record-receipt":
		updated, err = current.RecordReceipt(input.Amount, input.ActorID, now)
		auditAction = financedomain.FinanceAuditActionReceivableReceiptRecorded
		metadata = map[string]any{"receipt_amount": input.Amount}
	case "mark-disputed":
		updated, err = current.MarkDisputed(input.ActorID, input.Reason, now)
		auditAction = financedomain.FinanceAuditActionReceivableDisputed
		metadata = map[string]any{"reason": strings.TrimSpace(input.Reason)}
	case "void":
		updated, err = current.Void(input.ActorID, input.Reason, now)
		auditAction = financedomain.FinanceAuditActionReceivableVoided
		metadata = map[string]any{"reason": strings.TrimSpace(input.Reason)}
	default:
		return CustomerReceivableActionResult{}, errors.New("customer receivable action is unknown")
	}
	if err != nil {
		return CustomerReceivableActionResult{}, mapCustomerReceivableError(
			err,
			map[string]any{"customer_receivable_id": current.ID},
		)
	}
	if err := s.store.Save(ctx, updated); err != nil {
		return CustomerReceivableActionResult{}, err
	}
	for key, value := range updated.SourceDocument.Metadata() {
		metadata[key] = value
	}
	auditLogID, err := s.recordCustomerReceivableAudit(
		ctx,
		updated,
		auditAction,
		input.ActorID,
		input.RequestID,
		before,
		customerReceivableAuditData(updated),
		metadata,
		now,
	)
	if err != nil {
		return CustomerReceivableActionResult{}, err
	}

	return CustomerReceivableActionResult{
		CustomerReceivable: updated,
		PreviousStatus:     current.Status,
		CurrentStatus:      updated.Status,
		AuditLogID:         auditLogID,
	}, nil
}

func (s *PrototypeCustomerReceivableStore) List(
	_ context.Context,
	filter CustomerReceivableFilter,
) ([]financedomain.CustomerReceivable, error) {
	if s == nil {
		return nil, errors.New("customer receivable store is required")
	}
	filter = normalizeCustomerReceivableFilter(filter)
	s.mu.RLock()
	defer s.mu.RUnlock()

	receivables := make([]financedomain.CustomerReceivable, 0, len(s.records))
	for _, receivable := range s.records {
		if !matchesCustomerReceivableFilter(receivable, filter) {
			continue
		}
		receivables = append(receivables, receivable.Clone())
	}
	sort.Slice(receivables, func(i, j int) bool {
		return receivables[i].CreatedAt.After(receivables[j].CreatedAt)
	})

	return receivables, nil
}

func (s *PrototypeCustomerReceivableStore) Get(
	_ context.Context,
	id string,
) (financedomain.CustomerReceivable, error) {
	if s == nil {
		return financedomain.CustomerReceivable{}, errors.New("customer receivable store is required")
	}
	id = strings.TrimSpace(id)
	s.mu.RLock()
	defer s.mu.RUnlock()
	receivable, ok := s.records[id]
	if !ok {
		return financedomain.CustomerReceivable{}, ErrCustomerReceivableNotFound
	}

	return receivable.Clone(), nil
}

func (s *PrototypeCustomerReceivableStore) Save(
	_ context.Context,
	receivable financedomain.CustomerReceivable,
) error {
	if s == nil {
		return errors.New("customer receivable store is required")
	}
	if err := receivable.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[receivable.ID] = receivable.Clone()

	return nil
}

func (s CustomerReceivableService) recordCustomerReceivableAudit(
	ctx context.Context,
	receivable financedomain.CustomerReceivable,
	action financedomain.FinanceAuditAction,
	actorID string,
	requestID string,
	before map[string]any,
	after map[string]any,
	metadata map[string]any,
	createdAt time.Time,
) (string, error) {
	log, err := audit.NewLog(audit.NewLogInput{
		ID:         fmt.Sprintf("audit_%s_%d", strings.ReplaceAll(receivable.ID, "-", "_"), createdAt.UnixNano()),
		OrgID:      receivable.OrgID,
		ActorID:    strings.TrimSpace(actorID),
		Action:     string(action),
		EntityType: string(financedomain.FinanceEntityTypeCustomerReceivable),
		EntityID:   receivable.ID,
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

func sourceDocumentFromInput(input SourceDocumentInput) (financedomain.SourceDocumentRef, error) {
	return financedomain.NewSourceDocumentRef(
		financedomain.SourceDocumentType(input.Type),
		input.ID,
		input.No,
	)
}

func customerReceivableLinesFromInput(
	inputs []CustomerReceivableLineInput,
) ([]financedomain.NewCustomerReceivableLineInput, error) {
	lines := make([]financedomain.NewCustomerReceivableLineInput, 0, len(inputs))
	for _, input := range inputs {
		source, err := sourceDocumentFromInput(input.SourceDocument)
		if err != nil {
			return nil, err
		}
		lines = append(lines, financedomain.NewCustomerReceivableLineInput{
			ID:             input.ID,
			Description:    input.Description,
			SourceDocument: source,
			Amount:         input.Amount,
		})
	}

	return lines, nil
}

func normalizeCustomerReceivableFilter(filter CustomerReceivableFilter) CustomerReceivableFilter {
	filter.Search = strings.ToLower(strings.TrimSpace(filter.Search))
	filter.CustomerID = strings.TrimSpace(filter.CustomerID)
	statuses := make([]financedomain.ReceivableStatus, 0, len(filter.Statuses))
	for _, status := range filter.Statuses {
		normalized := financedomain.NormalizeReceivableStatus(status)
		if normalized != "" {
			statuses = append(statuses, normalized)
		}
	}
	filter.Statuses = statuses

	return filter
}

func matchesCustomerReceivableFilter(
	receivable financedomain.CustomerReceivable,
	filter CustomerReceivableFilter,
) bool {
	if filter.CustomerID != "" && receivable.CustomerID != filter.CustomerID {
		return false
	}
	if len(filter.Statuses) > 0 {
		matched := false
		for _, status := range filter.Statuses {
			if financedomain.NormalizeReceivableStatus(receivable.Status) == status {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if filter.Search == "" {
		return true
	}
	haystack := strings.ToLower(strings.Join([]string{
		receivable.ID,
		receivable.ReceivableNo,
		receivable.CustomerID,
		receivable.CustomerCode,
		receivable.CustomerName,
		receivable.SourceDocument.ID,
		receivable.SourceDocument.No,
	}, " "))

	return strings.Contains(haystack, filter.Search)
}

func customerReceivableAuditData(receivable financedomain.CustomerReceivable) map[string]any {
	return map[string]any{
		"status":             string(receivable.Status),
		"customer_id":        receivable.CustomerID,
		"customer_code":      receivable.CustomerCode,
		"customer_name":      receivable.CustomerName,
		"total_amount":       receivable.TotalAmount.String(),
		"paid_amount":        receivable.PaidAmount.String(),
		"outstanding_amount": receivable.OutstandingAmount.String(),
		"currency_code":      receivable.CurrencyCode.String(),
		"version":            receivable.Version,
	}
}

func parseOptionalFinanceDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return time.Time{}, err
	}

	return parsed, nil
}

func newCustomerReceivableID(now time.Time) string {
	return fmt.Sprintf("ar-%s-%d", now.Format("060102150405"), now.UnixNano()%100000)
}

func newCustomerReceivableNo(now time.Time) string {
	return fmt.Sprintf("AR-%s-%05d", now.Format("060102-150405"), now.UnixNano()%100000)
}

func customerReceivableValidationError(cause error, details map[string]any) error {
	return apperrors.BadRequest(
		ErrorCodeCustomerReceivableValidation,
		"Customer receivable request is invalid",
		cause,
		details,
	)
}

func mapCustomerReceivableError(err error, details map[string]any) error {
	if err == nil {
		return nil
	}
	if _, ok := apperrors.As(err); ok {
		return err
	}
	if errors.Is(err, ErrCustomerReceivableNotFound) {
		return apperrors.NotFound(
			ErrorCodeCustomerReceivableNotFound,
			"Customer receivable not found",
			err,
			details,
		)
	}
	if errors.Is(err, financedomain.ErrCustomerReceivableInvalidTransition) ||
		errors.Is(err, financedomain.ErrCustomerReceivableInvalidStatus) {
		return apperrors.Conflict(
			ErrorCodeCustomerReceivableInvalidState,
			"Customer receivable state is invalid",
			err,
			details,
		)
	}
	if errors.Is(err, financedomain.ErrCustomerReceivableRequiredField) ||
		errors.Is(err, financedomain.ErrCustomerReceivableInvalidAmount) ||
		errors.Is(err, financedomain.ErrCustomerReceivableInvalidSource) ||
		errors.Is(err, financedomain.ErrFinanceInvalidSourceDocument) ||
		errors.Is(err, financedomain.ErrFinanceInvalidCurrency) ||
		errors.Is(err, financedomain.ErrFinanceInvalidMoneyAmount) {
		return customerReceivableValidationError(err, details)
	}

	return err
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}

	return ""
}

func prototypeCustomerReceivables() []financedomain.CustomerReceivable {
	source := financedomain.SourceDocumentRef{
		Type: financedomain.SourceDocumentTypeShipment,
		ID:   "shipment-cod-260430-0001",
		No:   "SHP-COD-260430-0001",
	}
	receivable, err := financedomain.NewCustomerReceivable(financedomain.NewCustomerReceivableInput{
		ID:             "ar-cod-260430-0001",
		OrgID:          defaultFinanceOrgID,
		ReceivableNo:   "AR-COD-260430-0001",
		CustomerID:     "customer-hcm-001",
		CustomerCode:   "KH-HCM-001",
		CustomerName:   "My Pham HCM Retail",
		SourceDocument: source,
		TotalAmount:    "1250000.00",
		CurrencyCode:   "VND",
		DueDate:        time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC),
		CreatedAt:      time.Date(2026, 4, 30, 8, 0, 0, 0, time.UTC),
		CreatedBy:      "system-seed",
		Lines: []financedomain.NewCustomerReceivableLineInput{
			{
				ID:             "ar-cod-260430-0001-line-1",
				Description:    "COD delivered goods",
				SourceDocument: source,
				Amount:         "1250000.00",
			},
		},
	})
	if err != nil {
		return nil
	}

	return []financedomain.CustomerReceivable{receivable}
}
