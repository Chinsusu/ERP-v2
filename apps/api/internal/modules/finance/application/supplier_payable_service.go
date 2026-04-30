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

var ErrSupplierPayableNotFound = errors.New("supplier payable not found")

const (
	ErrorCodeSupplierPayableNotFound     response.ErrorCode = "SUPPLIER_PAYABLE_NOT_FOUND"
	ErrorCodeSupplierPayableValidation   response.ErrorCode = "SUPPLIER_PAYABLE_VALIDATION_ERROR"
	ErrorCodeSupplierPayableInvalidState response.ErrorCode = "SUPPLIER_PAYABLE_INVALID_STATE"
)

type SupplierPayableStore interface {
	List(ctx context.Context, filter SupplierPayableFilter) ([]financedomain.SupplierPayable, error)
	Get(ctx context.Context, id string) (financedomain.SupplierPayable, error)
	Save(ctx context.Context, payable financedomain.SupplierPayable) error
}

type SupplierPayableService struct {
	store    SupplierPayableStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type SupplierPayableFilter struct {
	Search     string
	Statuses   []financedomain.PayableStatus
	SupplierID string
}

type SupplierPayableLineInput struct {
	ID             string
	Description    string
	SourceDocument SourceDocumentInput
	Amount         string
}

type CreateSupplierPayableInput struct {
	ID             string
	OrgID          string
	PayableNo      string
	SupplierID     string
	SupplierCode   string
	SupplierName   string
	Status         string
	SourceDocument SourceDocumentInput
	Lines          []SupplierPayableLineInput
	TotalAmount    string
	CurrencyCode   string
	DueDate        string
	ActorID        string
	RequestID      string
}

type SupplierPayableActionInput struct {
	ID        string
	Amount    string
	Reason    string
	ActorID   string
	RequestID string
}

type SupplierPayableResult struct {
	SupplierPayable financedomain.SupplierPayable
	AuditLogID      string
}

type SupplierPayableActionResult struct {
	SupplierPayable financedomain.SupplierPayable
	PreviousStatus  financedomain.PayableStatus
	CurrentStatus   financedomain.PayableStatus
	AuditLogID      string
}

type PrototypeSupplierPayableStore struct {
	mu      sync.RWMutex
	records map[string]financedomain.SupplierPayable
}

func NewSupplierPayableService(
	store SupplierPayableStore,
	auditLog audit.LogStore,
) SupplierPayableService {
	return SupplierPayableService{
		store:    store,
		auditLog: auditLog,
		clock:    func() time.Time { return time.Now().UTC() },
	}
}

func (s SupplierPayableService) WithClock(clock func() time.Time) SupplierPayableService {
	if clock != nil {
		s.clock = clock
	}

	return s
}

func NewPrototypeSupplierPayableStore() *PrototypeSupplierPayableStore {
	store := &PrototypeSupplierPayableStore{records: make(map[string]financedomain.SupplierPayable)}
	for _, payable := range prototypeSupplierPayables() {
		store.records[payable.ID] = payable.Clone()
	}

	return store
}

func (s SupplierPayableService) ListSupplierPayables(
	ctx context.Context,
	filter SupplierPayableFilter,
) ([]financedomain.SupplierPayable, error) {
	if s.store == nil {
		return nil, errors.New("supplier payable store is required")
	}

	return s.store.List(ctx, filter)
}

func (s SupplierPayableService) GetSupplierPayable(
	ctx context.Context,
	id string,
) (financedomain.SupplierPayable, error) {
	if s.store == nil {
		return financedomain.SupplierPayable{}, errors.New("supplier payable store is required")
	}
	payable, err := s.store.Get(ctx, id)
	if err != nil {
		return financedomain.SupplierPayable{}, mapSupplierPayableError(
			err,
			map[string]any{"supplier_payable_id": strings.TrimSpace(id)},
		)
	}

	return payable, nil
}

func (s SupplierPayableService) CreateSupplierPayable(
	ctx context.Context,
	input CreateSupplierPayableInput,
) (SupplierPayableResult, error) {
	if s.store == nil {
		return SupplierPayableResult{}, errors.New("supplier payable store is required")
	}
	if s.auditLog == nil {
		return SupplierPayableResult{}, errors.New("audit log store is required")
	}
	now := s.clock().UTC()
	id := firstNonBlank(input.ID, newSupplierPayableID(now))
	payableNo := firstNonBlank(input.PayableNo, newSupplierPayableNo(now))
	source, err := sourceDocumentFromInput(input.SourceDocument)
	if err != nil {
		return SupplierPayableResult{}, mapSupplierPayableError(err, nil)
	}
	lines, err := supplierPayableLinesFromInput(input.Lines)
	if err != nil {
		return SupplierPayableResult{}, mapSupplierPayableError(err, nil)
	}
	dueDate, err := parseOptionalFinanceDate(input.DueDate)
	if err != nil {
		return SupplierPayableResult{}, supplierPayableValidationError(err, map[string]any{"field": "due_date"})
	}

	payable, err := financedomain.NewSupplierPayable(financedomain.NewSupplierPayableInput{
		ID:             id,
		OrgID:          firstNonBlank(input.OrgID, defaultFinanceOrgID),
		PayableNo:      payableNo,
		SupplierID:     input.SupplierID,
		SupplierCode:   input.SupplierCode,
		SupplierName:   input.SupplierName,
		Status:         financedomain.PayableStatus(input.Status),
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
		return SupplierPayableResult{}, mapSupplierPayableError(err, map[string]any{"supplier_payable_id": id})
	}
	if err := s.store.Save(ctx, payable); err != nil {
		return SupplierPayableResult{}, err
	}
	auditLogID, err := s.recordSupplierPayableAudit(
		ctx,
		payable,
		financedomain.FinanceAuditActionPayableCreated,
		input.ActorID,
		input.RequestID,
		nil,
		supplierPayableAuditData(payable),
		payable.SourceDocument.Metadata(),
		now,
	)
	if err != nil {
		return SupplierPayableResult{}, err
	}

	return SupplierPayableResult{SupplierPayable: payable, AuditLogID: auditLogID}, nil
}

func (s SupplierPayableService) ApproveSupplierPayablePayment(
	ctx context.Context,
	input SupplierPayableActionInput,
) (SupplierPayableActionResult, error) {
	return s.applySupplierPayableAction(ctx, input, "approve-payment")
}

func (s SupplierPayableService) RequestSupplierPayablePayment(
	ctx context.Context,
	input SupplierPayableActionInput,
) (SupplierPayableActionResult, error) {
	return s.applySupplierPayableAction(ctx, input, "request-payment")
}

func (s SupplierPayableService) RejectSupplierPayablePayment(
	ctx context.Context,
	input SupplierPayableActionInput,
) (SupplierPayableActionResult, error) {
	return s.applySupplierPayableAction(ctx, input, "reject-payment")
}

func (s SupplierPayableService) RecordSupplierPayablePayment(
	ctx context.Context,
	input SupplierPayableActionInput,
) (SupplierPayableActionResult, error) {
	return s.applySupplierPayableAction(ctx, input, "record-payment")
}

func (s SupplierPayableService) VoidSupplierPayable(
	ctx context.Context,
	input SupplierPayableActionInput,
) (SupplierPayableActionResult, error) {
	return s.applySupplierPayableAction(ctx, input, "void")
}

func (s SupplierPayableService) applySupplierPayableAction(
	ctx context.Context,
	input SupplierPayableActionInput,
	action string,
) (SupplierPayableActionResult, error) {
	if s.store == nil {
		return SupplierPayableActionResult{}, errors.New("supplier payable store is required")
	}
	if s.auditLog == nil {
		return SupplierPayableActionResult{}, errors.New("audit log store is required")
	}
	current, err := s.store.Get(ctx, input.ID)
	if err != nil {
		return SupplierPayableActionResult{}, mapSupplierPayableError(
			err,
			map[string]any{"supplier_payable_id": strings.TrimSpace(input.ID)},
		)
	}

	now := s.clock().UTC()
	before := supplierPayableAuditData(current)
	var (
		updated     financedomain.SupplierPayable
		auditAction financedomain.FinanceAuditAction
		metadata    map[string]any
	)
	switch action {
	case "request-payment":
		updated, err = current.RequestPayment(input.ActorID, now)
		auditAction = financedomain.FinanceAuditActionPayablePaymentRequested
		metadata = map[string]any{"requested_amount": current.OutstandingAmount.String()}
	case "approve-payment":
		updated, err = approveSupplierPayableFromCurrent(current, input.ActorID, now)
		auditAction = financedomain.FinanceAuditActionPayablePaymentApproved
		metadata = map[string]any{"approved_amount": current.OutstandingAmount.String()}
	case "reject-payment":
		updated, err = current.RejectPayment(input.ActorID, input.Reason, now)
		auditAction = financedomain.FinanceAuditActionPayablePaymentRejected
		metadata = map[string]any{"reason": strings.TrimSpace(input.Reason)}
	case "record-payment":
		updated, err = current.RecordPayment(input.Amount, input.ActorID, now)
		auditAction = financedomain.FinanceAuditActionPayablePaymentRecorded
		metadata = map[string]any{"payment_amount": input.Amount}
	case "void":
		updated, err = current.Void(input.ActorID, input.Reason, now)
		auditAction = financedomain.FinanceAuditActionPayableVoided
		metadata = map[string]any{"reason": strings.TrimSpace(input.Reason)}
	default:
		return SupplierPayableActionResult{}, errors.New("supplier payable action is unknown")
	}
	if err != nil {
		return SupplierPayableActionResult{}, mapSupplierPayableError(
			err,
			map[string]any{"supplier_payable_id": current.ID},
		)
	}
	if err := s.store.Save(ctx, updated); err != nil {
		return SupplierPayableActionResult{}, err
	}
	for key, value := range updated.SourceDocument.Metadata() {
		metadata[key] = value
	}
	auditLogID, err := s.recordSupplierPayableAudit(
		ctx,
		updated,
		auditAction,
		input.ActorID,
		input.RequestID,
		before,
		supplierPayableAuditData(updated),
		metadata,
		now,
	)
	if err != nil {
		return SupplierPayableActionResult{}, err
	}

	return SupplierPayableActionResult{
		SupplierPayable: updated,
		PreviousStatus:  current.Status,
		CurrentStatus:   updated.Status,
		AuditLogID:      auditLogID,
	}, nil
}

func (s *PrototypeSupplierPayableStore) List(
	_ context.Context,
	filter SupplierPayableFilter,
) ([]financedomain.SupplierPayable, error) {
	if s == nil {
		return nil, errors.New("supplier payable store is required")
	}
	filter = normalizeSupplierPayableFilter(filter)
	s.mu.RLock()
	defer s.mu.RUnlock()

	payables := make([]financedomain.SupplierPayable, 0, len(s.records))
	for _, payable := range s.records {
		if !matchesSupplierPayableFilter(payable, filter) {
			continue
		}
		payables = append(payables, payable.Clone())
	}
	sort.Slice(payables, func(i, j int) bool {
		return payables[i].CreatedAt.After(payables[j].CreatedAt)
	})

	return payables, nil
}

func (s *PrototypeSupplierPayableStore) Get(
	_ context.Context,
	id string,
) (financedomain.SupplierPayable, error) {
	if s == nil {
		return financedomain.SupplierPayable{}, errors.New("supplier payable store is required")
	}
	id = strings.TrimSpace(id)
	s.mu.RLock()
	defer s.mu.RUnlock()
	payable, ok := s.records[id]
	if !ok {
		return financedomain.SupplierPayable{}, ErrSupplierPayableNotFound
	}

	return payable.Clone(), nil
}

func (s *PrototypeSupplierPayableStore) Save(
	_ context.Context,
	payable financedomain.SupplierPayable,
) error {
	if s == nil {
		return errors.New("supplier payable store is required")
	}
	if err := payable.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[payable.ID] = payable.Clone()

	return nil
}

func (s SupplierPayableService) recordSupplierPayableAudit(
	ctx context.Context,
	payable financedomain.SupplierPayable,
	action financedomain.FinanceAuditAction,
	actorID string,
	requestID string,
	before map[string]any,
	after map[string]any,
	metadata map[string]any,
	createdAt time.Time,
) (string, error) {
	log, err := audit.NewLog(audit.NewLogInput{
		ID:         fmt.Sprintf("audit_%s_%d", strings.ReplaceAll(payable.ID, "-", "_"), createdAt.UnixNano()),
		OrgID:      payable.OrgID,
		ActorID:    strings.TrimSpace(actorID),
		Action:     string(action),
		EntityType: string(financedomain.FinanceEntityTypeSupplierPayable),
		EntityID:   payable.ID,
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

func supplierPayableLinesFromInput(
	inputs []SupplierPayableLineInput,
) ([]financedomain.NewSupplierPayableLineInput, error) {
	lines := make([]financedomain.NewSupplierPayableLineInput, 0, len(inputs))
	for _, input := range inputs {
		source, err := sourceDocumentFromInput(input.SourceDocument)
		if err != nil {
			return nil, err
		}
		lines = append(lines, financedomain.NewSupplierPayableLineInput{
			ID:             input.ID,
			Description:    input.Description,
			SourceDocument: source,
			Amount:         input.Amount,
		})
	}

	return lines, nil
}

func approveSupplierPayableFromCurrent(
	current financedomain.SupplierPayable,
	actorID string,
	approvedAt time.Time,
) (financedomain.SupplierPayable, error) {
	status := financedomain.NormalizePayableStatus(current.Status)
	if status == financedomain.PayableStatusOpen {
		requested, err := current.RequestPayment(actorID, approvedAt)
		if err != nil {
			return financedomain.SupplierPayable{}, err
		}
		return requested.ApprovePayment(actorID, approvedAt)
	}

	return current.ApprovePayment(actorID, approvedAt)
}

func normalizeSupplierPayableFilter(filter SupplierPayableFilter) SupplierPayableFilter {
	filter.Search = strings.ToLower(strings.TrimSpace(filter.Search))
	filter.SupplierID = strings.TrimSpace(filter.SupplierID)
	statuses := make([]financedomain.PayableStatus, 0, len(filter.Statuses))
	for _, status := range filter.Statuses {
		normalized := financedomain.NormalizePayableStatus(status)
		if normalized != "" {
			statuses = append(statuses, normalized)
		}
	}
	filter.Statuses = statuses

	return filter
}

func matchesSupplierPayableFilter(
	payable financedomain.SupplierPayable,
	filter SupplierPayableFilter,
) bool {
	if filter.SupplierID != "" && payable.SupplierID != filter.SupplierID {
		return false
	}
	if len(filter.Statuses) > 0 {
		matched := false
		for _, status := range filter.Statuses {
			if financedomain.NormalizePayableStatus(payable.Status) == status {
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
		payable.ID,
		payable.PayableNo,
		payable.SupplierID,
		payable.SupplierCode,
		payable.SupplierName,
		payable.SourceDocument.ID,
		payable.SourceDocument.No,
	}, " "))

	return strings.Contains(haystack, filter.Search)
}

func supplierPayableAuditData(payable financedomain.SupplierPayable) map[string]any {
	return map[string]any{
		"status":             string(payable.Status),
		"supplier_id":        payable.SupplierID,
		"supplier_code":      payable.SupplierCode,
		"supplier_name":      payable.SupplierName,
		"total_amount":       payable.TotalAmount.String(),
		"paid_amount":        payable.PaidAmount.String(),
		"outstanding_amount": payable.OutstandingAmount.String(),
		"currency_code":      payable.CurrencyCode.String(),
		"version":            payable.Version,
	}
}

func newSupplierPayableID(now time.Time) string {
	return fmt.Sprintf("ap-%s-%d", now.Format("060102150405"), now.UnixNano()%100000)
}

func newSupplierPayableNo(now time.Time) string {
	return fmt.Sprintf("AP-%s-%05d", now.Format("060102-150405"), now.UnixNano()%100000)
}

func supplierPayableValidationError(cause error, details map[string]any) error {
	return apperrors.BadRequest(
		ErrorCodeSupplierPayableValidation,
		"Supplier payable request is invalid",
		cause,
		details,
	)
}

func mapSupplierPayableError(err error, details map[string]any) error {
	if err == nil {
		return nil
	}
	if _, ok := apperrors.As(err); ok {
		return err
	}
	if errors.Is(err, ErrSupplierPayableNotFound) {
		return apperrors.NotFound(
			ErrorCodeSupplierPayableNotFound,
			"Supplier payable not found",
			err,
			details,
		)
	}
	if errors.Is(err, financedomain.ErrSupplierPayableInvalidTransition) ||
		errors.Is(err, financedomain.ErrSupplierPayableInvalidStatus) {
		return apperrors.Conflict(
			ErrorCodeSupplierPayableInvalidState,
			"Supplier payable state is invalid",
			err,
			details,
		)
	}
	if errors.Is(err, financedomain.ErrSupplierPayableRequiredField) ||
		errors.Is(err, financedomain.ErrSupplierPayableInvalidAmount) ||
		errors.Is(err, financedomain.ErrSupplierPayableInvalidSource) ||
		errors.Is(err, financedomain.ErrFinanceInvalidSourceDocument) ||
		errors.Is(err, financedomain.ErrFinanceInvalidCurrency) ||
		errors.Is(err, financedomain.ErrFinanceInvalidMoneyAmount) {
		return supplierPayableValidationError(err, details)
	}

	return err
}

func prototypeSupplierPayables() []financedomain.SupplierPayable {
	qcSource := financedomain.SourceDocumentRef{
		Type: financedomain.SourceDocumentTypeQCInspection,
		ID:   "qc-inbound-260430-0001",
		No:   "QC-260430-0001",
	}
	receiptSource := financedomain.SourceDocumentRef{
		Type: financedomain.SourceDocumentTypeWarehouseReceipt,
		ID:   "gr-260430-0001",
		No:   "GR-260430-0001",
	}
	poSource := financedomain.SourceDocumentRef{
		Type: financedomain.SourceDocumentTypePurchaseOrder,
		ID:   "po-260430-0001",
		No:   "PO-260430-0001",
	}
	payable, err := financedomain.NewSupplierPayable(financedomain.NewSupplierPayableInput{
		ID:             "ap-supplier-260430-0001",
		OrgID:          defaultFinanceOrgID,
		PayableNo:      "AP-SUP-260430-0001",
		SupplierID:     "supplier-hcm-001",
		SupplierCode:   "SUP-HCM-001",
		SupplierName:   "Nguyen Lieu HCM",
		SourceDocument: qcSource,
		TotalAmount:    "4250000.00",
		CurrencyCode:   "VND",
		DueDate:        time.Date(2026, 5, 7, 0, 0, 0, 0, time.UTC),
		CreatedAt:      time.Date(2026, 4, 30, 8, 30, 0, 0, time.UTC),
		CreatedBy:      "system-seed",
		Lines: []financedomain.NewSupplierPayableLineInput{
			{
				ID:             "ap-supplier-260430-0001-line-1",
				Description:    "Accepted raw material after inbound QC",
				SourceDocument: receiptSource,
				Amount:         "3000000.00",
			},
			{
				ID:             "ap-supplier-260430-0001-line-2",
				Description:    "Accepted packaging after inbound QC",
				SourceDocument: poSource,
				Amount:         "1250000.00",
			},
		},
	})
	if err != nil {
		return nil
	}

	return []financedomain.SupplierPayable{payable}
}
