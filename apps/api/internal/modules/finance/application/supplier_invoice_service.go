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
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
	apperrors "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/errors"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

var ErrSupplierInvoiceNotFound = errors.New("supplier invoice not found")

const (
	ErrorCodeSupplierInvoiceNotFound     response.ErrorCode = "SUPPLIER_INVOICE_NOT_FOUND"
	ErrorCodeSupplierInvoiceValidation   response.ErrorCode = "SUPPLIER_INVOICE_VALIDATION_ERROR"
	ErrorCodeSupplierInvoiceInvalidState response.ErrorCode = "SUPPLIER_INVOICE_INVALID_STATE"
)

type SupplierInvoiceStore interface {
	List(ctx context.Context, filter SupplierInvoiceFilter) ([]financedomain.SupplierInvoice, error)
	Get(ctx context.Context, id string) (financedomain.SupplierInvoice, error)
	Save(ctx context.Context, invoice financedomain.SupplierInvoice) error
}

type SupplierInvoiceService struct {
	store        SupplierInvoiceStore
	payableStore SupplierPayableStore
	auditLog     audit.LogStore
	clock        func() time.Time
}

type SupplierInvoiceFilter struct {
	Search     string
	Statuses   []financedomain.SupplierInvoiceStatus
	SupplierID string
	PayableID  string
}

type SupplierInvoiceLineInput struct {
	ID             string
	Description    string
	SourceDocument SourceDocumentInput
	Amount         string
}

type CreateSupplierInvoiceInput struct {
	ID            string
	OrgID         string
	InvoiceNo     string
	SupplierID    string
	SupplierCode  string
	SupplierName  string
	PayableID     string
	InvoiceDate   string
	InvoiceAmount string
	CurrencyCode  string
	Lines         []SupplierInvoiceLineInput
	ActorID       string
	RequestID     string
}

type SupplierInvoiceActionInput struct {
	ID        string
	Reason    string
	ActorID   string
	RequestID string
}

type SupplierInvoiceResult struct {
	SupplierInvoice financedomain.SupplierInvoice
	AuditLogID      string
}

type SupplierInvoiceActionResult struct {
	SupplierInvoice financedomain.SupplierInvoice
	PreviousStatus  financedomain.SupplierInvoiceStatus
	CurrentStatus   financedomain.SupplierInvoiceStatus
	AuditLogID      string
}

type PrototypeSupplierInvoiceStore struct {
	mu      sync.RWMutex
	records map[string]financedomain.SupplierInvoice
}

func NewSupplierInvoiceService(
	store SupplierInvoiceStore,
	payableStore SupplierPayableStore,
	auditLog audit.LogStore,
) SupplierInvoiceService {
	return SupplierInvoiceService{
		store:        store,
		payableStore: payableStore,
		auditLog:     auditLog,
		clock:        func() time.Time { return time.Now().UTC() },
	}
}

func (s SupplierInvoiceService) WithClock(clock func() time.Time) SupplierInvoiceService {
	if clock != nil {
		s.clock = clock
	}

	return s
}

func NewPrototypeSupplierInvoiceStore() *PrototypeSupplierInvoiceStore {
	store := &PrototypeSupplierInvoiceStore{records: make(map[string]financedomain.SupplierInvoice)}
	for _, invoice := range prototypeSupplierInvoices() {
		store.records[invoice.ID] = invoice.Clone()
	}

	return store
}

func (s SupplierInvoiceService) ListSupplierInvoices(
	ctx context.Context,
	filter SupplierInvoiceFilter,
) ([]financedomain.SupplierInvoice, error) {
	if s.store == nil {
		return nil, errors.New("supplier invoice store is required")
	}

	return s.store.List(ctx, filter)
}

func (s SupplierInvoiceService) GetSupplierInvoice(
	ctx context.Context,
	id string,
) (financedomain.SupplierInvoice, error) {
	if s.store == nil {
		return financedomain.SupplierInvoice{}, errors.New("supplier invoice store is required")
	}
	invoice, err := s.store.Get(ctx, id)
	if err != nil {
		return financedomain.SupplierInvoice{}, mapSupplierInvoiceError(
			err,
			map[string]any{"supplier_invoice_id": strings.TrimSpace(id)},
		)
	}

	return invoice, nil
}

func (s SupplierInvoiceService) CreateSupplierInvoice(
	ctx context.Context,
	input CreateSupplierInvoiceInput,
) (SupplierInvoiceResult, error) {
	if s.store == nil {
		return SupplierInvoiceResult{}, errors.New("supplier invoice store is required")
	}
	if s.payableStore == nil {
		return SupplierInvoiceResult{}, errors.New("supplier payable store is required")
	}
	if s.auditLog == nil {
		return SupplierInvoiceResult{}, errors.New("audit log store is required")
	}
	payable, err := s.payableStore.Get(ctx, input.PayableID)
	if err != nil {
		return SupplierInvoiceResult{}, mapSupplierInvoiceError(
			err,
			map[string]any{"supplier_payable_id": strings.TrimSpace(input.PayableID)},
		)
	}
	now := s.clock().UTC()
	id := firstNonBlank(input.ID, newSupplierInvoiceID(now))
	invoiceNo := firstNonBlank(input.InvoiceNo, newSupplierInvoiceNo(now))
	invoiceDate, err := parseOptionalFinanceDate(input.InvoiceDate)
	if err != nil {
		return SupplierInvoiceResult{}, supplierInvoiceValidationError(err, map[string]any{"field": "invoice_date"})
	}
	if invoiceDate.IsZero() {
		invoiceDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	}
	currencyCode := firstNonBlank(input.CurrencyCode, payable.CurrencyCode.String())
	invoiceAmount, err := financedomain.NewMoneyAmount(input.InvoiceAmount, currencyCode)
	if err != nil || invoiceAmount.Amount.IsZero() {
		return SupplierInvoiceResult{}, supplierInvoiceValidationError(
			financedomain.ErrSupplierInvoiceInvalidAmount,
			map[string]any{"field": "invoice_amount"},
		)
	}
	varianceAmount, err := subtractSupplierInvoiceMoney(invoiceAmount.Amount, payable.TotalAmount, currencyCode)
	if err != nil {
		return SupplierInvoiceResult{}, supplierInvoiceValidationError(err, map[string]any{"field": "invoice_amount"})
	}
	status := financedomain.SupplierInvoiceStatusMatched
	matchStatus := financedomain.SupplierInvoiceMatchStatusMatched
	if !varianceAmount.IsZero() {
		status = financedomain.SupplierInvoiceStatusMismatch
		matchStatus = financedomain.SupplierInvoiceMatchStatusMismatch
	}
	lines, err := supplierInvoiceLinesFromInputOrPayable(id, payable, varianceAmount, input.Lines)
	if err != nil {
		return SupplierInvoiceResult{}, mapSupplierInvoiceError(err, nil)
	}

	invoice, err := financedomain.NewSupplierInvoice(financedomain.NewSupplierInvoiceInput{
		ID:             id,
		OrgID:          firstNonBlank(input.OrgID, payable.OrgID, defaultFinanceOrgID),
		InvoiceNo:      invoiceNo,
		SupplierID:     firstNonBlank(input.SupplierID, payable.SupplierID),
		SupplierCode:   firstNonBlank(input.SupplierCode, payable.SupplierCode),
		SupplierName:   firstNonBlank(input.SupplierName, payable.SupplierName),
		PayableID:      payable.ID,
		PayableNo:      payable.PayableNo,
		Status:         status,
		MatchStatus:    matchStatus,
		SourceDocument: payable.SourceDocument,
		Lines:          lines,
		InvoiceAmount:  invoiceAmount.Amount.String(),
		ExpectedAmount: payable.TotalAmount.String(),
		VarianceAmount: varianceAmount.String(),
		CurrencyCode:   currencyCode,
		InvoiceDate:    invoiceDate,
		CreatedAt:      now,
		CreatedBy:      input.ActorID,
		UpdatedAt:      now,
		UpdatedBy:      input.ActorID,
	})
	if err != nil {
		return SupplierInvoiceResult{}, mapSupplierInvoiceError(err, map[string]any{"supplier_invoice_id": id})
	}
	if err := s.store.Save(ctx, invoice); err != nil {
		return SupplierInvoiceResult{}, err
	}
	auditLogID, err := s.recordSupplierInvoiceAudit(
		ctx,
		invoice,
		financedomain.FinanceAuditActionSupplierInvoiceCreated,
		input.ActorID,
		input.RequestID,
		nil,
		supplierInvoiceAuditData(invoice),
		supplierInvoiceAuditMetadata(invoice),
		now,
	)
	if err != nil {
		return SupplierInvoiceResult{}, err
	}

	return SupplierInvoiceResult{SupplierInvoice: invoice, AuditLogID: auditLogID}, nil
}

func (s SupplierInvoiceService) VoidSupplierInvoice(
	ctx context.Context,
	input SupplierInvoiceActionInput,
) (SupplierInvoiceActionResult, error) {
	if s.store == nil {
		return SupplierInvoiceActionResult{}, errors.New("supplier invoice store is required")
	}
	if s.auditLog == nil {
		return SupplierInvoiceActionResult{}, errors.New("audit log store is required")
	}
	current, err := s.store.Get(ctx, input.ID)
	if err != nil {
		return SupplierInvoiceActionResult{}, mapSupplierInvoiceError(
			err,
			map[string]any{"supplier_invoice_id": strings.TrimSpace(input.ID)},
		)
	}
	now := s.clock().UTC()
	before := supplierInvoiceAuditData(current)
	updated, err := current.Void(input.ActorID, input.Reason, now)
	if err != nil {
		return SupplierInvoiceActionResult{}, mapSupplierInvoiceError(
			err,
			map[string]any{"supplier_invoice_id": current.ID},
		)
	}
	if err := s.store.Save(ctx, updated); err != nil {
		return SupplierInvoiceActionResult{}, err
	}
	auditLogID, err := s.recordSupplierInvoiceAudit(
		ctx,
		updated,
		financedomain.FinanceAuditActionSupplierInvoiceVoided,
		input.ActorID,
		input.RequestID,
		before,
		supplierInvoiceAuditData(updated),
		map[string]any{"reason": strings.TrimSpace(input.Reason)},
		now,
	)
	if err != nil {
		return SupplierInvoiceActionResult{}, err
	}

	return SupplierInvoiceActionResult{
		SupplierInvoice: updated,
		PreviousStatus:  current.Status,
		CurrentStatus:   updated.Status,
		AuditLogID:      auditLogID,
	}, nil
}

func (s *PrototypeSupplierInvoiceStore) List(
	_ context.Context,
	filter SupplierInvoiceFilter,
) ([]financedomain.SupplierInvoice, error) {
	if s == nil {
		return nil, errors.New("supplier invoice store is required")
	}
	filter = normalizeSupplierInvoiceFilter(filter)
	s.mu.RLock()
	defer s.mu.RUnlock()

	invoices := make([]financedomain.SupplierInvoice, 0, len(s.records))
	for _, invoice := range s.records {
		if !matchesSupplierInvoiceFilter(invoice, filter) {
			continue
		}
		invoices = append(invoices, invoice.Clone())
	}
	sort.Slice(invoices, func(i, j int) bool {
		return invoices[i].CreatedAt.After(invoices[j].CreatedAt)
	})

	return invoices, nil
}

func (s *PrototypeSupplierInvoiceStore) Get(
	_ context.Context,
	id string,
) (financedomain.SupplierInvoice, error) {
	if s == nil {
		return financedomain.SupplierInvoice{}, errors.New("supplier invoice store is required")
	}
	id = strings.TrimSpace(id)
	s.mu.RLock()
	defer s.mu.RUnlock()
	invoice, ok := s.records[id]
	if !ok {
		return financedomain.SupplierInvoice{}, ErrSupplierInvoiceNotFound
	}

	return invoice.Clone(), nil
}

func (s *PrototypeSupplierInvoiceStore) Save(
	_ context.Context,
	invoice financedomain.SupplierInvoice,
) error {
	if s == nil {
		return errors.New("supplier invoice store is required")
	}
	if err := invoice.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[invoice.ID] = invoice.Clone()

	return nil
}

func (s SupplierInvoiceService) recordSupplierInvoiceAudit(
	ctx context.Context,
	invoice financedomain.SupplierInvoice,
	action financedomain.FinanceAuditAction,
	actorID string,
	requestID string,
	before map[string]any,
	after map[string]any,
	metadata map[string]any,
	createdAt time.Time,
) (string, error) {
	log, err := audit.NewLog(audit.NewLogInput{
		ID:         fmt.Sprintf("audit_%s_%d", strings.ReplaceAll(invoice.ID, "-", "_"), createdAt.UnixNano()),
		OrgID:      invoice.OrgID,
		ActorID:    strings.TrimSpace(actorID),
		Action:     string(action),
		EntityType: string(financedomain.FinanceEntityTypeSupplierInvoice),
		EntityID:   invoice.ID,
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

func supplierInvoiceLinesFromInputOrPayable(
	invoiceID string,
	payable financedomain.SupplierPayable,
	varianceAmount decimal.Decimal,
	inputs []SupplierInvoiceLineInput,
) ([]financedomain.NewSupplierInvoiceLineInput, error) {
	if len(inputs) > 0 {
		return supplierInvoiceLinesFromInput(inputs)
	}

	lines := make([]financedomain.NewSupplierInvoiceLineInput, 0, len(payable.Lines)+1)
	for _, line := range payable.Lines {
		lines = append(lines, financedomain.NewSupplierInvoiceLineInput{
			ID:             fmt.Sprintf("%s-%s", invoiceID, line.ID),
			Description:    line.Description,
			SourceDocument: line.SourceDocument,
			Amount:         line.Amount.String(),
		})
	}
	if !varianceAmount.IsZero() {
		lines = append(lines, financedomain.NewSupplierInvoiceLineInput{
			ID:          fmt.Sprintf("%s-variance", invoiceID),
			Description: "Supplier invoice variance against AP",
			SourceDocument: financedomain.SourceDocumentRef{
				Type: financedomain.SourceDocumentTypeSupplierPayable,
				ID:   payable.ID,
				No:   payable.PayableNo,
			},
			Amount: varianceAmount.String(),
		})
	}

	return lines, nil
}

func supplierInvoiceLinesFromInput(
	inputs []SupplierInvoiceLineInput,
) ([]financedomain.NewSupplierInvoiceLineInput, error) {
	lines := make([]financedomain.NewSupplierInvoiceLineInput, 0, len(inputs))
	for _, input := range inputs {
		source, err := sourceDocumentFromInput(input.SourceDocument)
		if err != nil {
			return nil, err
		}
		lines = append(lines, financedomain.NewSupplierInvoiceLineInput{
			ID:             input.ID,
			Description:    input.Description,
			SourceDocument: source,
			Amount:         input.Amount,
		})
	}

	return lines, nil
}

func subtractSupplierInvoiceMoney(
	left decimal.Decimal,
	right decimal.Decimal,
	currencyCode string,
) (decimal.Decimal, error) {
	accumulator := newMoneyAccumulator()
	if err := accumulator.Add(left); err != nil {
		return "", err
	}
	if err := accumulator.Subtract(right); err != nil {
		return "", err
	}

	money, err := financedomain.NewSignedMoneyAmount(accumulator.String(), currencyCode)
	if err != nil {
		return "", err
	}

	return money.Amount, nil
}

func normalizeSupplierInvoiceFilter(filter SupplierInvoiceFilter) SupplierInvoiceFilter {
	filter.Search = strings.ToLower(strings.TrimSpace(filter.Search))
	filter.SupplierID = strings.TrimSpace(filter.SupplierID)
	filter.PayableID = strings.TrimSpace(filter.PayableID)
	statuses := make([]financedomain.SupplierInvoiceStatus, 0, len(filter.Statuses))
	for _, status := range filter.Statuses {
		normalized := financedomain.NormalizeSupplierInvoiceStatus(status)
		if normalized != "" {
			statuses = append(statuses, normalized)
		}
	}
	filter.Statuses = statuses

	return filter
}

func matchesSupplierInvoiceFilter(
	invoice financedomain.SupplierInvoice,
	filter SupplierInvoiceFilter,
) bool {
	if filter.SupplierID != "" && invoice.SupplierID != filter.SupplierID {
		return false
	}
	if filter.PayableID != "" && invoice.PayableID != filter.PayableID {
		return false
	}
	if len(filter.Statuses) > 0 {
		matched := false
		for _, status := range filter.Statuses {
			if financedomain.NormalizeSupplierInvoiceStatus(invoice.Status) == status {
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
		invoice.ID,
		invoice.InvoiceNo,
		invoice.SupplierID,
		invoice.SupplierCode,
		invoice.SupplierName,
		invoice.PayableID,
		invoice.PayableNo,
		invoice.SourceDocument.ID,
		invoice.SourceDocument.No,
	}, " "))
	for _, line := range invoice.Lines {
		haystack += " " + strings.ToLower(strings.Join([]string{
			line.Description,
			line.SourceDocument.ID,
			line.SourceDocument.No,
		}, " "))
	}

	return strings.Contains(haystack, filter.Search)
}

func supplierInvoiceAuditData(invoice financedomain.SupplierInvoice) map[string]any {
	return map[string]any{
		"status":          string(invoice.Status),
		"match_status":    string(invoice.MatchStatus),
		"supplier_id":     invoice.SupplierID,
		"supplier_code":   invoice.SupplierCode,
		"supplier_name":   invoice.SupplierName,
		"payable_id":      invoice.PayableID,
		"payable_no":      invoice.PayableNo,
		"invoice_amount":  invoice.InvoiceAmount.String(),
		"expected_amount": invoice.ExpectedAmount.String(),
		"variance_amount": invoice.VarianceAmount.String(),
		"currency_code":   invoice.CurrencyCode.String(),
		"version":         invoice.Version,
	}
}

func supplierInvoiceAuditMetadata(invoice financedomain.SupplierInvoice) map[string]any {
	metadata := invoice.SourceDocument.Metadata()
	metadata["payable_id"] = invoice.PayableID
	metadata["payable_no"] = invoice.PayableNo
	metadata["match_status"] = string(invoice.MatchStatus)
	metadata["variance_amount"] = invoice.VarianceAmount.String()

	return metadata
}

func newSupplierInvoiceID(now time.Time) string {
	return fmt.Sprintf("si-%s-%d", now.Format("060102150405"), now.UnixNano()%100000)
}

func newSupplierInvoiceNo(now time.Time) string {
	return fmt.Sprintf("SI-%s-%05d", now.Format("060102-150405"), now.UnixNano()%100000)
}

func supplierInvoiceValidationError(cause error, details map[string]any) error {
	return apperrors.BadRequest(
		ErrorCodeSupplierInvoiceValidation,
		"Supplier invoice request is invalid",
		cause,
		details,
	)
}

func mapSupplierInvoiceError(err error, details map[string]any) error {
	if err == nil {
		return nil
	}
	if _, ok := apperrors.As(err); ok {
		return err
	}
	if errors.Is(err, ErrSupplierInvoiceNotFound) {
		return apperrors.NotFound(
			ErrorCodeSupplierInvoiceNotFound,
			"Supplier invoice not found",
			err,
			details,
		)
	}
	if errors.Is(err, ErrSupplierPayableNotFound) {
		return apperrors.NotFound(
			ErrorCodeSupplierPayableNotFound,
			"Supplier payable not found",
			err,
			details,
		)
	}
	if errors.Is(err, financedomain.ErrSupplierInvoiceInvalidTransition) ||
		errors.Is(err, financedomain.ErrSupplierInvoiceInvalidStatus) {
		return apperrors.Conflict(
			ErrorCodeSupplierInvoiceInvalidState,
			"Supplier invoice state is invalid",
			err,
			details,
		)
	}
	if errors.Is(err, financedomain.ErrSupplierInvoiceRequiredField) ||
		errors.Is(err, financedomain.ErrSupplierInvoiceInvalidAmount) ||
		errors.Is(err, financedomain.ErrSupplierInvoiceInvalidSource) ||
		errors.Is(err, financedomain.ErrFinanceInvalidSourceDocument) ||
		errors.Is(err, financedomain.ErrFinanceInvalidCurrency) ||
		errors.Is(err, financedomain.ErrFinanceInvalidMoneyAmount) {
		return supplierInvoiceValidationError(err, details)
	}

	return err
}

func prototypeSupplierInvoices() []financedomain.SupplierInvoice {
	input := financedomain.NewSupplierInvoiceInput{
		ID:             "si-supplier-260430-0001",
		OrgID:          defaultFinanceOrgID,
		InvoiceNo:      "INV-SUP-260430-0001",
		SupplierID:     "supplier-hcm-001",
		SupplierCode:   "SUP-HCM-001",
		SupplierName:   "Nguyen Lieu HCM",
		PayableID:      "ap-supplier-260430-0001",
		PayableNo:      "AP-SUP-260430-0001",
		Status:         financedomain.SupplierInvoiceStatusMatched,
		MatchStatus:    financedomain.SupplierInvoiceMatchStatusMatched,
		SourceDocument: financedomain.SourceDocumentRef{Type: financedomain.SourceDocumentTypeQCInspection, ID: "qc-inbound-260430-0001", No: "QC-260430-0001"},
		InvoiceAmount:  "4250000.00",
		ExpectedAmount: "4250000.00",
		VarianceAmount: "0.00",
		CurrencyCode:   "VND",
		InvoiceDate:    time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC),
		CreatedAt:      time.Date(2026, 5, 5, 9, 0, 0, 0, time.UTC),
		CreatedBy:      "system-seed",
		UpdatedAt:      time.Date(2026, 5, 5, 9, 0, 0, 0, time.UTC),
		UpdatedBy:      "system-seed",
		Lines: []financedomain.NewSupplierInvoiceLineInput{
			{
				ID:             "si-supplier-260430-0001-line-1",
				Description:    "Accepted raw material after inbound QC",
				SourceDocument: financedomain.SourceDocumentRef{Type: financedomain.SourceDocumentTypeWarehouseReceipt, ID: "gr-260430-0001", No: "GR-260430-0001"},
				Amount:         "3000000.00",
			},
			{
				ID:             "si-supplier-260430-0001-line-2",
				Description:    "Accepted packaging after inbound QC",
				SourceDocument: financedomain.SourceDocumentRef{Type: financedomain.SourceDocumentTypePurchaseOrder, ID: "po-260430-0001", No: "PO-260430-0001"},
				Amount:         "1250000.00",
			},
		},
	}
	invoice, err := financedomain.NewSupplierInvoice(input)
	if err != nil {
		return nil
	}

	return []financedomain.SupplierInvoice{invoice}
}
