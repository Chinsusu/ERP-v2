package application

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

var ErrExpectedReturnNotFound = errors.New("expected return not found")
var ErrExpectedReturnOrderNotReturnable = errors.New("expected return order status is not returnable")
var ErrReturnReceiptDuplicate = errors.New("return receipt already exists")
var ErrReturnReceiptNotFound = errors.New("return receipt not found")

type ReturnReceiptStore interface {
	List(ctx context.Context, filter domain.ReturnReceiptFilter) ([]domain.ReturnReceipt, error)
	Save(ctx context.Context, receipt domain.ReturnReceipt) error
	FindExpectedReturnByCode(ctx context.Context, code string) (domain.ExpectedReturn, error)
}

type ListReturnReceipts struct {
	store ReturnReceiptStore
}

type ReceiveReturn struct {
	store    ReturnReceiptStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type ReceiveReturnInput struct {
	WarehouseID       string
	WarehouseCode     string
	Source            string
	ScanCode          string
	PackageCondition  string
	Disposition       string
	InvestigationNote string
	ActorID           string
	RequestID         string
}

type ReturnReceiptResult struct {
	Receipt    domain.ReturnReceipt
	AuditLogID string
}

func NewListReturnReceipts(store ReturnReceiptStore) ListReturnReceipts {
	return ListReturnReceipts{store: store}
}

func NewReceiveReturn(store ReturnReceiptStore, auditLog audit.LogStore) ReceiveReturn {
	return ReceiveReturn{
		store:    store,
		auditLog: auditLog,
		clock:    func() time.Time { return time.Now().UTC() },
	}
}

func (uc ListReturnReceipts) Execute(
	ctx context.Context,
	filter domain.ReturnReceiptFilter,
) ([]domain.ReturnReceipt, error) {
	if uc.store == nil {
		return nil, errors.New("return receipt store is required")
	}

	return uc.store.List(ctx, filter)
}

func (uc ReceiveReturn) Execute(ctx context.Context, input ReceiveReturnInput) (ReturnReceiptResult, error) {
	if uc.store == nil {
		return ReturnReceiptResult{}, errors.New("return receipt store is required")
	}
	if uc.auditLog == nil {
		return ReturnReceiptResult{}, errors.New("audit log store is required")
	}

	var expected *domain.ExpectedReturn
	found, err := uc.store.FindExpectedReturnByCode(ctx, input.ScanCode)
	if err == nil {
		if !domain.IsReturnReceivableOrderStatus(found.OrderStatus) {
			return ReturnReceiptResult{}, ErrExpectedReturnOrderNotReturnable
		}
		expected = &found
	} else if !errors.Is(err, ErrExpectedReturnNotFound) {
		return ReturnReceiptResult{}, err
	}

	receipt, err := domain.NewReturnReceipt(domain.NewReturnReceiptInput{
		WarehouseID:       input.WarehouseID,
		WarehouseCode:     input.WarehouseCode,
		Source:            domain.ReturnSource(input.Source),
		ReceivedBy:        input.ActorID,
		ScanCode:          input.ScanCode,
		PackageCondition:  input.PackageCondition,
		Disposition:       domain.ReturnDisposition(input.Disposition),
		ExpectedReturn:    expected,
		InvestigationNote: input.InvestigationNote,
		CreatedAt:         uc.clock(),
	})
	if err != nil {
		return ReturnReceiptResult{}, err
	}
	if err := uc.store.Save(ctx, receipt); err != nil {
		return ReturnReceiptResult{}, err
	}

	log, err := newReturnReceiptAuditLog(input.ActorID, input.RequestID, receipt)
	if err != nil {
		return ReturnReceiptResult{}, err
	}
	if err := uc.auditLog.Record(ctx, log); err != nil {
		return ReturnReceiptResult{}, err
	}

	return ReturnReceiptResult{Receipt: receipt, AuditLogID: log.ID}, nil
}

type PrototypeReturnReceiptStore struct {
	mu           sync.RWMutex
	records      map[string]domain.ReturnReceipt
	inspections  map[string]domain.ReturnInspection
	dispositions map[string]domain.ReturnDispositionAction
	expected     []domain.ExpectedReturn
}

func NewPrototypeReturnReceiptStore() *PrototypeReturnReceiptStore {
	store := &PrototypeReturnReceiptStore{
		records:      make(map[string]domain.ReturnReceipt),
		inspections:  make(map[string]domain.ReturnInspection),
		dispositions: make(map[string]domain.ReturnDispositionAction),
	}
	store.expected = prototypeExpectedReturns()
	for _, receipt := range prototypeReturnReceipts() {
		store.records[receipt.ID] = receipt.Clone()
	}

	return store
}

func (s *PrototypeReturnReceiptStore) List(
	_ context.Context,
	filter domain.ReturnReceiptFilter,
) ([]domain.ReturnReceipt, error) {
	if s == nil {
		return nil, errors.New("return receipt store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows := make([]domain.ReturnReceipt, 0, len(s.records))
	for _, record := range s.records {
		if filter.WarehouseID != "" && record.WarehouseID != filter.WarehouseID {
			continue
		}
		if filter.Status != "" && record.Status != filter.Status {
			continue
		}
		rows = append(rows, record.Clone())
	}
	domain.SortReturnReceipts(rows)

	return rows, nil
}

func (s *PrototypeReturnReceiptStore) Save(_ context.Context, receipt domain.ReturnReceipt) error {
	if s == nil {
		return errors.New("return receipt store is required")
	}
	if strings.TrimSpace(receipt.ID) == "" {
		return errors.New("return receipt id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.records {
		if returnReceiptDuplicates(existing, receipt) {
			return ErrReturnReceiptDuplicate
		}
	}
	s.records[receipt.ID] = receipt.Clone()

	return nil
}

func (s *PrototypeReturnReceiptStore) FindReceiptByID(
	_ context.Context,
	id string,
) (domain.ReturnReceipt, error) {
	if s == nil {
		return domain.ReturnReceipt{}, errors.New("return receipt store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	receipt, ok := s.records[strings.TrimSpace(id)]
	if !ok {
		return domain.ReturnReceipt{}, ErrReturnReceiptNotFound
	}

	return receipt.Clone(), nil
}

func (s *PrototypeReturnReceiptStore) SaveInspection(
	_ context.Context,
	receipt domain.ReturnReceipt,
	inspection domain.ReturnInspection,
) error {
	if s == nil {
		return errors.New("return receipt store is required")
	}
	if strings.TrimSpace(receipt.ID) == "" || strings.TrimSpace(inspection.ID) == "" {
		return domain.ErrReturnInspectionRequiredField
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.records[receipt.ID]; !ok {
		return ErrReturnReceiptNotFound
	}
	s.records[receipt.ID] = receipt.Clone()
	s.inspections[inspection.ID] = inspection.Clone()

	return nil
}

func (s *PrototypeReturnReceiptStore) SaveDisposition(
	_ context.Context,
	receipt domain.ReturnReceipt,
	action domain.ReturnDispositionAction,
) error {
	if s == nil {
		return errors.New("return receipt store is required")
	}
	if strings.TrimSpace(receipt.ID) == "" || strings.TrimSpace(action.ID) == "" {
		return domain.ErrReturnDispositionRequiredField
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.records[receipt.ID]; !ok {
		return ErrReturnReceiptNotFound
	}
	s.records[receipt.ID] = receipt.Clone()
	s.dispositions[action.ID] = action.Clone()

	return nil
}

func returnReceiptDuplicates(existing domain.ReturnReceipt, receipt domain.ReturnReceipt) bool {
	if strings.TrimSpace(existing.ID) != "" && strings.TrimSpace(existing.ID) == strings.TrimSpace(receipt.ID) {
		return true
	}
	if sameReturnIdentifier(existing.ScanCode, receipt.ScanCode) {
		return true
	}

	for _, left := range []string{existing.OriginalOrderNo, existing.TrackingNo, existing.ReturnCode} {
		for _, right := range []string{receipt.OriginalOrderNo, receipt.TrackingNo, receipt.ReturnCode} {
			if sameReturnIdentifier(left, right) {
				return true
			}
		}
	}

	return false
}

func sameReturnIdentifier(left string, right string) bool {
	normalizedLeft := domain.NormalizeReturnScanCode(left)
	normalizedRight := domain.NormalizeReturnScanCode(right)

	return normalizedLeft != "" && normalizedLeft == normalizedRight
}

func (s *PrototypeReturnReceiptStore) FindExpectedReturnByCode(
	_ context.Context,
	code string,
) (domain.ExpectedReturn, error) {
	if s == nil {
		return domain.ExpectedReturn{}, errors.New("return receipt store is required")
	}

	normalizedCode := domain.NormalizeReturnScanCode(code)
	if normalizedCode == "" {
		return domain.ExpectedReturn{}, ErrExpectedReturnNotFound
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, expected := range s.expected {
		if expected.MatchesScanCode(normalizedCode) {
			return expected, nil
		}
	}

	return domain.ExpectedReturn{}, ErrExpectedReturnNotFound
}

func newReturnReceiptAuditLog(actorID string, requestID string, receipt domain.ReturnReceipt) (audit.Log, error) {
	afterData := map[string]any{
		"receipt_no":        receipt.ReceiptNo,
		"warehouse_id":      receipt.WarehouseID,
		"source":            string(receipt.Source),
		"status":            string(receipt.Status),
		"disposition":       string(receipt.Disposition),
		"target_location":   receipt.TargetLocation,
		"scan_code":         receipt.ScanCode,
		"original_order_no": receipt.OriginalOrderNo,
		"tracking_no":       receipt.TrackingNo,
		"unknown_case":      receipt.UnknownCase,
	}
	if receipt.StockMovement != nil {
		afterData["stock_movement_id"] = receipt.StockMovement.ID
		afterData["movement_type"] = receipt.StockMovement.MovementType
		afterData["target_stock_status"] = receipt.StockMovement.TargetStockStatus
	}

	return audit.NewLog(audit.NewLogInput{
		OrgID:      "org-my-pham",
		ActorID:    strings.TrimSpace(actorID),
		Action:     "returns.receipt.created",
		EntityType: "returns.return_receipt",
		EntityID:   receipt.ID,
		RequestID:  strings.TrimSpace(requestID),
		AfterData:  afterData,
		Metadata: map[string]any{
			"source": "return receiving",
		},
		CreatedAt: receipt.CreatedAt,
	})
}

func prototypeReturnReceipts() []domain.ReturnReceipt {
	receipt, err := domain.NewReturnReceipt(domain.NewReturnReceiptInput{
		ID:                "rr-260426-0001",
		ReceiptNo:         "RR-260426-0001",
		WarehouseID:       "wh-hcm",
		WarehouseCode:     "HCM",
		Source:            domain.ReturnSourceCarrier,
		ReceivedBy:        "user-return-inspector",
		ScanCode:          "GHN260425099",
		PackageCondition:  "sealed bag",
		Disposition:       domain.ReturnDispositionNeedsInspection,
		InvestigationNote: "Customer reported wrong shade",
		ExpectedReturn: &domain.ExpectedReturn{
			OrderNo:       "SO-260425-099",
			OrderStatus:   "delivered",
			TrackingNo:    "GHN260425099",
			ReturnCode:    "RET-260425-099",
			CustomerName:  "Tran Binh",
			SKU:           "CREAM-50ML",
			ProductName:   "Repair Cream 50ml",
			Quantity:      1,
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			Source:        domain.ReturnSourceCarrier,
		},
		CreatedAt: time.Date(2026, 4, 26, 8, 30, 0, 0, time.UTC),
	})
	if err != nil {
		return nil
	}

	return []domain.ReturnReceipt{receipt}
}

func prototypeExpectedReturns() []domain.ExpectedReturn {
	return []domain.ExpectedReturn{
		{
			OrderNo:       "SO-260426-001",
			OrderStatus:   "handed_over",
			TrackingNo:    "GHN260426001",
			ReturnCode:    "RET-260426-001",
			ShipmentID:    "ship-hcm-260426-001",
			CustomerName:  "Nguyen An",
			SKU:           "SERUM-30ML",
			ProductName:   "Hydrating Serum 30ml",
			Quantity:      1,
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			Source:        domain.ReturnSourceCarrier,
		},
		{
			OrderNo:       "SO-260426-004",
			OrderStatus:   "delivered",
			TrackingNo:    "GHN260426004",
			ReturnCode:    "RET-260426-004",
			ShipmentID:    "ship-hcm-260426-004",
			CustomerName:  "Le Chi",
			SKU:           "TONER-100ML",
			ProductName:   "Balancing Toner 100ml",
			Quantity:      2,
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			Source:        domain.ReturnSourceShipper,
		},
		{
			OrderNo:       "SO-260426-HN-011",
			OrderStatus:   "handed_over",
			TrackingNo:    "GHNHN260426001",
			ReturnCode:    "RET-HN-260426-011",
			ShipmentID:    "ship-hn-260426-001",
			CustomerName:  "Pham Ha",
			SKU:           "MASK-SET-05",
			ProductName:   "Sheet Mask Set",
			Quantity:      1,
			WarehouseID:   "wh-hn",
			WarehouseCode: "HN",
			Source:        domain.ReturnSourceMarketplace,
		},
		{
			OrderNo:       "SO-260426-009",
			OrderStatus:   "waiting_handover",
			TrackingNo:    "GHN260426009",
			ReturnCode:    "RET-260426-009",
			ShipmentID:    "ship-hcm-260426-009",
			CustomerName:  "Vu Nhi",
			SKU:           "SERUM-30ML",
			ProductName:   "Hydrating Serum 30ml",
			Quantity:      1,
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			Source:        domain.ReturnSourceCarrier,
		},
	}
}

func NewPrototypeReceiveReturnAt(store ReturnReceiptStore, auditLog audit.LogStore, now time.Time) ReceiveReturn {
	service := NewReceiveReturn(store, auditLog)
	service.clock = func() time.Time { return now.UTC() }

	return service
}
