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

var ErrCODRemittanceNotFound = errors.New("cod remittance not found")

const (
	ErrorCodeCODRemittanceNotFound     response.ErrorCode = "COD_REMITTANCE_NOT_FOUND"
	ErrorCodeCODRemittanceValidation   response.ErrorCode = "COD_REMITTANCE_VALIDATION_ERROR"
	ErrorCodeCODRemittanceInvalidState response.ErrorCode = "COD_REMITTANCE_INVALID_STATE"
)

type CODRemittanceStore interface {
	List(ctx context.Context, filter CODRemittanceFilter) ([]financedomain.CODRemittance, error)
	Get(ctx context.Context, id string) (financedomain.CODRemittance, error)
	Save(ctx context.Context, remittance financedomain.CODRemittance) error
}

type CODRemittanceService struct {
	store    CODRemittanceStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type CODRemittanceFilter struct {
	Search    string
	Statuses  []financedomain.CODRemittanceStatus
	CarrierID string
}

type CODRemittanceLineInput struct {
	ID             string
	ReceivableID   string
	ReceivableNo   string
	ShipmentID     string
	TrackingNo     string
	CustomerName   string
	ExpectedAmount string
	RemittedAmount string
}

type CreateCODRemittanceInput struct {
	ID             string
	OrgID          string
	RemittanceNo   string
	CarrierID      string
	CarrierCode    string
	CarrierName    string
	BusinessDate   string
	ExpectedAmount string
	RemittedAmount string
	CurrencyCode   string
	Lines          []CODRemittanceLineInput
	ActorID        string
	RequestID      string
}

type CODRemittanceActionInput struct {
	ID        string
	ActorID   string
	RequestID string
}

type CODRemittanceDiscrepancyInput struct {
	RemittanceID  string
	DiscrepancyID string
	LineID        string
	Type          string
	Status        string
	Reason        string
	OwnerID       string
	ActorID       string
	RequestID     string
}

type CODRemittanceResult struct {
	CODRemittance financedomain.CODRemittance
	AuditLogID    string
}

type CODRemittanceActionResult struct {
	CODRemittance  financedomain.CODRemittance
	PreviousStatus financedomain.CODRemittanceStatus
	CurrentStatus  financedomain.CODRemittanceStatus
	AuditLogID     string
}

type PrototypeCODRemittanceStore struct {
	mu      sync.RWMutex
	records map[string]financedomain.CODRemittance
}

func NewCODRemittanceService(store CODRemittanceStore, auditLog audit.LogStore) CODRemittanceService {
	return CODRemittanceService{
		store:    store,
		auditLog: auditLog,
		clock:    func() time.Time { return time.Now().UTC() },
	}
}

func (s CODRemittanceService) WithClock(clock func() time.Time) CODRemittanceService {
	if clock != nil {
		s.clock = clock
	}

	return s
}

func NewPrototypeCODRemittanceStore() *PrototypeCODRemittanceStore {
	store := &PrototypeCODRemittanceStore{records: make(map[string]financedomain.CODRemittance)}
	for _, remittance := range prototypeCODRemittances() {
		store.records[remittance.ID] = remittance.Clone()
	}

	return store
}

func (s CODRemittanceService) ListCODRemittances(
	ctx context.Context,
	filter CODRemittanceFilter,
) ([]financedomain.CODRemittance, error) {
	if s.store == nil {
		return nil, errors.New("cod remittance store is required")
	}

	return s.store.List(ctx, filter)
}

func (s CODRemittanceService) GetCODRemittance(
	ctx context.Context,
	id string,
) (financedomain.CODRemittance, error) {
	if s.store == nil {
		return financedomain.CODRemittance{}, errors.New("cod remittance store is required")
	}
	remittance, err := s.store.Get(ctx, id)
	if err != nil {
		return financedomain.CODRemittance{}, mapCODRemittanceError(
			err,
			map[string]any{"cod_remittance_id": strings.TrimSpace(id)},
		)
	}

	return remittance, nil
}

func (s CODRemittanceService) CreateCODRemittance(
	ctx context.Context,
	input CreateCODRemittanceInput,
) (CODRemittanceResult, error) {
	if s.store == nil {
		return CODRemittanceResult{}, errors.New("cod remittance store is required")
	}
	if s.auditLog == nil {
		return CODRemittanceResult{}, errors.New("audit log store is required")
	}
	now := s.clock().UTC()
	businessDate, err := parseRequiredFinanceDate(input.BusinessDate)
	if err != nil {
		return CODRemittanceResult{}, codRemittanceValidationError(err, map[string]any{"field": "business_date"})
	}

	remittance, err := financedomain.NewCODRemittance(financedomain.NewCODRemittanceInput{
		ID:             firstNonBlank(input.ID, newCODRemittanceID(now)),
		OrgID:          firstNonBlank(input.OrgID, defaultFinanceOrgID),
		RemittanceNo:   firstNonBlank(input.RemittanceNo, newCODRemittanceNo(now)),
		CarrierID:      input.CarrierID,
		CarrierCode:    input.CarrierCode,
		CarrierName:    input.CarrierName,
		BusinessDate:   businessDate,
		ExpectedAmount: input.ExpectedAmount,
		RemittedAmount: input.RemittedAmount,
		CurrencyCode:   input.CurrencyCode,
		Lines:          codRemittanceLinesFromInput(input.Lines),
		CreatedAt:      now,
		CreatedBy:      input.ActorID,
		UpdatedAt:      now,
		UpdatedBy:      input.ActorID,
	})
	if err != nil {
		return CODRemittanceResult{}, mapCODRemittanceError(err, nil)
	}
	if err := s.store.Save(ctx, remittance); err != nil {
		return CODRemittanceResult{}, mapCODRemittanceError(err, nil)
	}
	auditLogID, err := s.recordCODRemittanceAudit(
		ctx,
		remittance,
		financedomain.FinanceAuditActionCODRemittanceCreated,
		input.ActorID,
		input.RequestID,
		nil,
		codRemittanceAuditData(remittance),
		map[string]any{"carrier_id": remittance.CarrierID, "carrier_code": remittance.CarrierCode},
		now,
	)
	if err != nil {
		return CODRemittanceResult{}, err
	}

	return CODRemittanceResult{CODRemittance: remittance, AuditLogID: auditLogID}, nil
}

func (s CODRemittanceService) MarkCODRemittanceMatching(
	ctx context.Context,
	input CODRemittanceActionInput,
) (CODRemittanceActionResult, error) {
	return s.applyCODRemittanceAction(ctx, input, "match")
}

func (s CODRemittanceService) SubmitCODRemittance(
	ctx context.Context,
	input CODRemittanceActionInput,
) (CODRemittanceActionResult, error) {
	return s.applyCODRemittanceAction(ctx, input, "submit")
}

func (s CODRemittanceService) ApproveCODRemittance(
	ctx context.Context,
	input CODRemittanceActionInput,
) (CODRemittanceActionResult, error) {
	return s.applyCODRemittanceAction(ctx, input, "approve")
}

func (s CODRemittanceService) CloseCODRemittance(
	ctx context.Context,
	input CODRemittanceActionInput,
) (CODRemittanceActionResult, error) {
	return s.applyCODRemittanceAction(ctx, input, "close")
}

func (s CODRemittanceService) RecordCODRemittanceDiscrepancy(
	ctx context.Context,
	input CODRemittanceDiscrepancyInput,
) (CODRemittanceActionResult, error) {
	if s.store == nil {
		return CODRemittanceActionResult{}, errors.New("cod remittance store is required")
	}
	if s.auditLog == nil {
		return CODRemittanceActionResult{}, errors.New("audit log store is required")
	}
	current, err := s.store.Get(ctx, input.RemittanceID)
	if err != nil {
		return CODRemittanceActionResult{}, mapCODRemittanceError(err, map[string]any{"cod_remittance_id": input.RemittanceID})
	}
	now := s.clock().UTC()
	before := codRemittanceAuditData(current)
	updated, err := current.RecordDiscrepancy(financedomain.RecordCODDiscrepancyInput{
		ID:         firstNonBlank(input.DiscrepancyID, input.LineID+"-discrepancy"),
		LineID:     input.LineID,
		Type:       financedomain.CODDiscrepancyType(input.Type),
		Status:     financedomain.CODDiscrepancyStatus(input.Status),
		Reason:     input.Reason,
		OwnerID:    input.OwnerID,
		RecordedBy: input.ActorID,
		RecordedAt: now,
	})
	if err != nil {
		return CODRemittanceActionResult{}, mapCODRemittanceError(err, map[string]any{"cod_remittance_id": current.ID})
	}
	if err := s.store.Save(ctx, updated); err != nil {
		return CODRemittanceActionResult{}, mapCODRemittanceError(err, nil)
	}
	auditLogID, err := s.recordCODRemittanceAudit(
		ctx,
		updated,
		financedomain.FinanceAuditActionCODRemittanceDiscrepancyRecorded,
		input.ActorID,
		input.RequestID,
		before,
		codRemittanceAuditData(updated),
		map[string]any{"line_id": input.LineID, "reason": strings.TrimSpace(input.Reason), "owner_id": strings.TrimSpace(input.OwnerID)},
		now,
	)
	if err != nil {
		return CODRemittanceActionResult{}, err
	}

	return CODRemittanceActionResult{
		CODRemittance:  updated,
		PreviousStatus: current.Status,
		CurrentStatus:  updated.Status,
		AuditLogID:     auditLogID,
	}, nil
}

func (s CODRemittanceService) applyCODRemittanceAction(
	ctx context.Context,
	input CODRemittanceActionInput,
	action string,
) (CODRemittanceActionResult, error) {
	if s.store == nil {
		return CODRemittanceActionResult{}, errors.New("cod remittance store is required")
	}
	if s.auditLog == nil {
		return CODRemittanceActionResult{}, errors.New("audit log store is required")
	}
	current, err := s.store.Get(ctx, input.ID)
	if err != nil {
		return CODRemittanceActionResult{}, mapCODRemittanceError(err, map[string]any{"cod_remittance_id": input.ID})
	}

	now := s.clock().UTC()
	before := codRemittanceAuditData(current)
	var (
		updated     financedomain.CODRemittance
		auditAction financedomain.FinanceAuditAction
	)
	switch action {
	case "match":
		updated, err = current.MarkMatching(input.ActorID, now)
		auditAction = financedomain.FinanceAuditActionCODRemittanceMatched
	case "submit":
		updated, err = current.Submit(input.ActorID, now)
		auditAction = financedomain.FinanceAuditActionCODRemittanceSubmitted
	case "approve":
		updated, err = current.Approve(input.ActorID, now)
		auditAction = financedomain.FinanceAuditActionCODRemittanceApproved
	case "close":
		updated, err = current.Close(input.ActorID, now)
		auditAction = financedomain.FinanceAuditActionCODRemittanceClosed
	default:
		return CODRemittanceActionResult{}, errors.New("cod remittance action is unknown")
	}
	if err != nil {
		return CODRemittanceActionResult{}, mapCODRemittanceError(err, map[string]any{"cod_remittance_id": current.ID})
	}
	if err := s.store.Save(ctx, updated); err != nil {
		return CODRemittanceActionResult{}, mapCODRemittanceError(err, nil)
	}
	auditLogID, err := s.recordCODRemittanceAudit(
		ctx,
		updated,
		auditAction,
		input.ActorID,
		input.RequestID,
		before,
		codRemittanceAuditData(updated),
		map[string]any{"carrier_id": updated.CarrierID, "carrier_code": updated.CarrierCode},
		now,
	)
	if err != nil {
		return CODRemittanceActionResult{}, err
	}

	return CODRemittanceActionResult{
		CODRemittance:  updated,
		PreviousStatus: current.Status,
		CurrentStatus:  updated.Status,
		AuditLogID:     auditLogID,
	}, nil
}

func (s *PrototypeCODRemittanceStore) List(
	_ context.Context,
	filter CODRemittanceFilter,
) ([]financedomain.CODRemittance, error) {
	if s == nil {
		return nil, errors.New("cod remittance store is required")
	}
	filter = normalizeCODRemittanceFilter(filter)
	s.mu.RLock()
	defer s.mu.RUnlock()

	remittances := make([]financedomain.CODRemittance, 0, len(s.records))
	for _, remittance := range s.records {
		if !matchesCODRemittanceFilter(remittance, filter) {
			continue
		}
		remittances = append(remittances, remittance.Clone())
	}
	sort.Slice(remittances, func(i, j int) bool {
		return remittances[i].CreatedAt.After(remittances[j].CreatedAt)
	})

	return remittances, nil
}

func (s *PrototypeCODRemittanceStore) Get(_ context.Context, id string) (financedomain.CODRemittance, error) {
	if s == nil {
		return financedomain.CODRemittance{}, errors.New("cod remittance store is required")
	}
	id = strings.TrimSpace(id)
	s.mu.RLock()
	defer s.mu.RUnlock()
	remittance, ok := s.records[id]
	if !ok {
		return financedomain.CODRemittance{}, ErrCODRemittanceNotFound
	}

	return remittance.Clone(), nil
}

func (s *PrototypeCODRemittanceStore) Save(_ context.Context, remittance financedomain.CODRemittance) error {
	if s == nil {
		return errors.New("cod remittance store is required")
	}
	if err := remittance.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[remittance.ID] = remittance.Clone()

	return nil
}

func (s CODRemittanceService) recordCODRemittanceAudit(
	ctx context.Context,
	remittance financedomain.CODRemittance,
	action financedomain.FinanceAuditAction,
	actorID string,
	requestID string,
	before map[string]any,
	after map[string]any,
	metadata map[string]any,
	createdAt time.Time,
) (string, error) {
	log, err := audit.NewLog(audit.NewLogInput{
		ID:         fmt.Sprintf("audit_%s_%d", strings.ReplaceAll(remittance.ID, "-", "_"), createdAt.UnixNano()),
		OrgID:      remittance.OrgID,
		ActorID:    strings.TrimSpace(actorID),
		Action:     string(action),
		EntityType: string(financedomain.FinanceEntityTypeCODRemittance),
		EntityID:   remittance.ID,
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

func codRemittanceLinesFromInput(inputs []CODRemittanceLineInput) []financedomain.NewCODRemittanceLineInput {
	lines := make([]financedomain.NewCODRemittanceLineInput, 0, len(inputs))
	for _, input := range inputs {
		lines = append(lines, financedomain.NewCODRemittanceLineInput{
			ID:             input.ID,
			ReceivableID:   input.ReceivableID,
			ReceivableNo:   input.ReceivableNo,
			ShipmentID:     input.ShipmentID,
			TrackingNo:     input.TrackingNo,
			CustomerName:   input.CustomerName,
			ExpectedAmount: input.ExpectedAmount,
			RemittedAmount: input.RemittedAmount,
		})
	}

	return lines
}

func normalizeCODRemittanceFilter(filter CODRemittanceFilter) CODRemittanceFilter {
	filter.Search = strings.ToLower(strings.TrimSpace(filter.Search))
	filter.CarrierID = strings.TrimSpace(filter.CarrierID)
	statuses := make([]financedomain.CODRemittanceStatus, 0, len(filter.Statuses))
	for _, status := range filter.Statuses {
		normalized := financedomain.NormalizeCODRemittanceStatus(status)
		if normalized != "" {
			statuses = append(statuses, normalized)
		}
	}
	filter.Statuses = statuses

	return filter
}

func matchesCODRemittanceFilter(remittance financedomain.CODRemittance, filter CODRemittanceFilter) bool {
	if filter.CarrierID != "" && remittance.CarrierID != filter.CarrierID {
		return false
	}
	if len(filter.Statuses) > 0 {
		matched := false
		for _, status := range filter.Statuses {
			if financedomain.NormalizeCODRemittanceStatus(remittance.Status) == status {
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
		remittance.ID,
		remittance.RemittanceNo,
		remittance.CarrierID,
		remittance.CarrierCode,
		remittance.CarrierName,
	}, " "))
	for _, line := range remittance.Lines {
		haystack += " " + strings.ToLower(strings.Join([]string{
			line.ReceivableID,
			line.ReceivableNo,
			line.ShipmentID,
			line.TrackingNo,
			line.CustomerName,
		}, " "))
	}

	return strings.Contains(haystack, filter.Search)
}

func codRemittanceAuditData(remittance financedomain.CODRemittance) map[string]any {
	return map[string]any{
		"status":             string(remittance.Status),
		"carrier_id":         remittance.CarrierID,
		"carrier_code":       remittance.CarrierCode,
		"carrier_name":       remittance.CarrierName,
		"expected_amount":    remittance.ExpectedAmount.String(),
		"remitted_amount":    remittance.RemittedAmount.String(),
		"discrepancy_amount": remittance.DiscrepancyAmount.String(),
		"currency_code":      remittance.CurrencyCode.String(),
		"line_count":         len(remittance.Lines),
		"discrepancy_count":  len(remittance.Discrepancies),
		"version":            remittance.Version,
	}
}

func parseRequiredFinanceDate(value string) (time.Time, error) {
	parsed, err := parseOptionalFinanceDate(value)
	if err != nil {
		return time.Time{}, err
	}
	if parsed.IsZero() {
		return time.Time{}, financedomain.ErrCODRemittanceRequiredField
	}

	return parsed, nil
}

func newCODRemittanceID(now time.Time) string {
	return fmt.Sprintf("cod-remit-%s-%d", now.Format("060102150405"), now.UnixNano()%100000)
}

func newCODRemittanceNo(now time.Time) string {
	return fmt.Sprintf("COD-%s-%05d", now.Format("060102-150405"), now.UnixNano()%100000)
}

func codRemittanceValidationError(cause error, details map[string]any) error {
	return apperrors.BadRequest(
		ErrorCodeCODRemittanceValidation,
		"COD remittance request is invalid",
		cause,
		details,
	)
}

func mapCODRemittanceError(err error, details map[string]any) error {
	if err == nil {
		return nil
	}
	if _, ok := apperrors.As(err); ok {
		return err
	}
	if errors.Is(err, ErrCODRemittanceNotFound) {
		return apperrors.NotFound(ErrorCodeCODRemittanceNotFound, "COD remittance not found", err, details)
	}
	if errors.Is(err, financedomain.ErrCODRemittanceInvalidTransition) ||
		errors.Is(err, financedomain.ErrCODRemittanceInvalidStatus) {
		return apperrors.Conflict(ErrorCodeCODRemittanceInvalidState, "COD remittance state is invalid", err, details)
	}
	if errors.Is(err, financedomain.ErrCODRemittanceRequiredField) ||
		errors.Is(err, financedomain.ErrCODRemittanceInvalidAmount) ||
		errors.Is(err, financedomain.ErrCODRemittanceInvalidDiscrepancy) ||
		errors.Is(err, financedomain.ErrFinanceInvalidCurrency) ||
		errors.Is(err, financedomain.ErrFinanceInvalidMoneyAmount) {
		return codRemittanceValidationError(err, details)
	}

	return err
}

func prototypeCODRemittances() []financedomain.CODRemittance {
	remittance, err := financedomain.NewCODRemittance(financedomain.NewCODRemittanceInput{
		ID:             "cod-remit-260430-0001",
		OrgID:          defaultFinanceOrgID,
		RemittanceNo:   "COD-GHN-260430-0001",
		CarrierID:      "carrier-ghn",
		CarrierCode:    "GHN",
		CarrierName:    "GHN Express",
		BusinessDate:   time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
		ExpectedAmount: "2000000.00",
		RemittedAmount: "1950000.00",
		CurrencyCode:   "VND",
		CreatedAt:      time.Date(2026, 4, 30, 9, 0, 0, 0, time.UTC),
		CreatedBy:      "system-seed",
		Lines: []financedomain.NewCODRemittanceLineInput{
			{
				ID:             "cod-remit-260430-0001-line-1",
				ReceivableID:   "ar-cod-260430-0001",
				ReceivableNo:   "AR-COD-260430-0001",
				ShipmentID:     "shipment-cod-260430-0001",
				TrackingNo:     "GHN260430001",
				CustomerName:   "My Pham HCM Retail",
				ExpectedAmount: "1250000.00",
				RemittedAmount: "1200000.00",
			},
			{
				ID:             "cod-remit-260430-0001-line-2",
				ReceivableID:   "ar-cod-260430-0002",
				ReceivableNo:   "AR-COD-260430-0002",
				ShipmentID:     "shipment-cod-260430-0002",
				TrackingNo:     "GHN260430002",
				CustomerName:   "Marketplace COD",
				ExpectedAmount: "750000.00",
				RemittedAmount: "750000.00",
			},
		},
	})
	if err != nil {
		return nil
	}

	return []financedomain.CODRemittance{remittance}
}
