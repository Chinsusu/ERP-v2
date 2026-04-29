package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	qcdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/qc/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrInboundQCInspectionNotFound = errors.New("inbound qc inspection not found")
var ErrInboundQCReceivingNotFound = errors.New("inbound qc receiving not found")
var ErrInboundQCReceivingInvalidState = errors.New("inbound qc receiving status is invalid")
var ErrInboundQCReceivingLineNotFound = errors.New("inbound qc receiving line not found")
var ErrInboundQCDuplicateReceivingLine = errors.New("inbound qc inspection already exists for receiving line")

const (
	inboundQCAuditEntityType         = "qc.inbound_inspection"
	inboundQCStockMovementEntityType = "qc.inbound_inspection_stock_movement"
	inboundQCStockMovementAction     = "qc.inbound_inspection.stock_movement.recorded"
	inboundQCStockMovementSourceDoc  = "inbound_qc_inspection"
	defaultInboundQCOrgID            = "org-my-pham"
)

type InboundQCInspectionStore interface {
	List(ctx context.Context, filter InboundQCInspectionFilter) ([]qcdomain.InboundQCInspection, error)
	Get(ctx context.Context, id string) (qcdomain.InboundQCInspection, error)
	Save(ctx context.Context, inspection qcdomain.InboundQCInspection) error
}

type InboundQCReceivingReader interface {
	Get(ctx context.Context, id string) (inventorydomain.WarehouseReceiving, error)
}

type InboundQCStockMovementRecorder interface {
	Record(ctx context.Context, movement inventorydomain.StockMovement) error
}

type InboundQCInspectionService struct {
	store         InboundQCInspectionStore
	receivingRead InboundQCReceivingReader
	stockMovement InboundQCStockMovementRecorder
	auditLog      audit.LogStore
	clock         func() time.Time
}

type InboundQCInspectionFilter struct {
	Status             qcdomain.InboundQCInspectionStatus
	GoodsReceiptID     string
	GoodsReceiptLineID string
	WarehouseID        string
}

type InboundQCChecklistInput struct {
	ID       string
	Code     string
	Label    string
	Required bool
	Status   string
	Note     string
}

type CreateInboundQCInspectionInput struct {
	ID                 string
	OrgID              string
	GoodsReceiptID     string
	GoodsReceiptLineID string
	InspectorID        string
	Checklist          []InboundQCChecklistInput
	Note               string
	ActorID            string
	RequestID          string
}

type InboundQCActionInput struct {
	ID             string
	PassedQuantity string
	FailedQuantity string
	HoldQuantity   string
	Checklist      []InboundQCChecklistInput
	Reason         string
	Note           string
	ActorID        string
	RequestID      string
}

type InboundQCInspectionResult struct {
	Inspection     qcdomain.InboundQCInspection
	PreviousStatus qcdomain.InboundQCInspectionStatus
	CurrentStatus  qcdomain.InboundQCInspectionStatus
	PreviousResult qcdomain.InboundQCResult
	CurrentResult  qcdomain.InboundQCResult
	AuditLogID     string
}

type PrototypeInboundQCInspectionStore struct {
	mu          sync.RWMutex
	inspections map[string]qcdomain.InboundQCInspection
}

func NewInboundQCInspectionService(
	store InboundQCInspectionStore,
	receivingRead InboundQCReceivingReader,
	auditLog audit.LogStore,
) InboundQCInspectionService {
	return InboundQCInspectionService{
		store:         store,
		receivingRead: receivingRead,
		auditLog:      auditLog,
		clock:         func() time.Time { return time.Now().UTC() },
	}
}

func (s InboundQCInspectionService) WithStockMovementRecorder(
	recorder InboundQCStockMovementRecorder,
) InboundQCInspectionService {
	s.stockMovement = recorder

	return s
}

func NewPrototypeInboundQCInspectionStore(
	rows ...qcdomain.InboundQCInspection,
) *PrototypeInboundQCInspectionStore {
	store := &PrototypeInboundQCInspectionStore{inspections: make(map[string]qcdomain.InboundQCInspection)}
	for _, row := range rows {
		store.inspections[row.ID] = row.Clone()
	}

	return store
}

func NewInboundQCInspectionFilter(
	status qcdomain.InboundQCInspectionStatus,
	goodsReceiptID string,
	goodsReceiptLineID string,
	warehouseID string,
) InboundQCInspectionFilter {
	return InboundQCInspectionFilter{
		Status:             qcdomain.NormalizeInboundQCInspectionStatus(status),
		GoodsReceiptID:     strings.TrimSpace(goodsReceiptID),
		GoodsReceiptLineID: strings.TrimSpace(goodsReceiptLineID),
		WarehouseID:        strings.TrimSpace(warehouseID),
	}
}

func (s InboundQCInspectionService) ListInboundQCInspections(
	ctx context.Context,
	filter InboundQCInspectionFilter,
) ([]qcdomain.InboundQCInspection, error) {
	if s.store == nil {
		return nil, errors.New("inbound qc inspection store is required")
	}

	return s.store.List(ctx, filter)
}

func (s InboundQCInspectionService) GetInboundQCInspection(
	ctx context.Context,
	id string,
) (qcdomain.InboundQCInspection, error) {
	if s.store == nil {
		return qcdomain.InboundQCInspection{}, errors.New("inbound qc inspection store is required")
	}

	return s.store.Get(ctx, id)
}

func (s InboundQCInspectionService) CreateInboundQCInspection(
	ctx context.Context,
	input CreateInboundQCInspectionInput,
) (InboundQCInspectionResult, error) {
	if err := s.ensureReadyForWrite(); err != nil {
		return InboundQCInspectionResult{}, err
	}

	receipt, line, err := s.readInspectableReceivingLine(ctx, input.GoodsReceiptID, input.GoodsReceiptLineID)
	if err != nil {
		return InboundQCInspectionResult{}, err
	}
	if err := s.ensureNoOpenInspection(ctx, receipt.ID, line.ID); err != nil {
		return InboundQCInspectionResult{}, err
	}

	now := s.clock()
	actorID := strings.TrimSpace(input.ActorID)
	inspectorID := strings.TrimSpace(input.InspectorID)
	if inspectorID == "" {
		inspectorID = actorID
	}
	orgID := strings.TrimSpace(input.OrgID)
	if orgID == "" {
		orgID = receipt.OrgID
	}
	if orgID == "" {
		orgID = defaultInboundQCOrgID
	}
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = defaultInspectionID(receipt.ID, line.ID)
	}
	checklist := checklistInputs(input.Checklist)
	if len(checklist) == 0 {
		checklist = defaultInboundQCChecklist()
	}

	inspection, err := qcdomain.NewInboundQCInspection(qcdomain.NewInboundQCInspectionInput{
		ID:                  id,
		OrgID:               orgID,
		GoodsReceiptID:      receipt.ID,
		GoodsReceiptNo:      receipt.ReceiptNo,
		GoodsReceiptLineID:  line.ID,
		PurchaseOrderID:     receipt.ReferenceDocID,
		PurchaseOrderLineID: line.PurchaseOrderLineID,
		ItemID:              line.ItemID,
		SKU:                 line.SKU,
		ItemName:            line.ItemName,
		BatchID:             line.BatchID,
		BatchNo:             line.BatchNo,
		LotNo:               line.LotNo,
		ExpiryDate:          line.ExpiryDate,
		WarehouseID:         line.WarehouseID,
		LocationID:          line.LocationID,
		Quantity:            line.Quantity,
		UOMCode:             line.BaseUOMCode.String(),
		InspectorID:         inspectorID,
		Checklist:           checklist,
		Note:                input.Note,
		CreatedAt:           now,
		CreatedBy:           actorID,
	})
	if err != nil {
		return InboundQCInspectionResult{}, err
	}
	if err := s.store.Save(ctx, inspection); err != nil {
		return InboundQCInspectionResult{}, err
	}
	log, err := newInboundQCAuditLog(
		actorID,
		input.RequestID,
		"qc.inbound_inspection.created",
		inspection,
		nil,
		inboundQCAuditData(inspection),
		now,
	)
	if err != nil {
		return InboundQCInspectionResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return InboundQCInspectionResult{}, err
	}

	return InboundQCInspectionResult{
		Inspection:    inspection,
		CurrentStatus: inspection.Status,
		CurrentResult: inspection.Result,
		AuditLogID:    log.ID,
	}, nil
}

func (s InboundQCInspectionService) StartInboundQCInspection(
	ctx context.Context,
	input InboundQCActionInput,
) (InboundQCInspectionResult, error) {
	return s.transition(ctx, input, "qc.inbound_inspection.started", func(
		current qcdomain.InboundQCInspection,
		actorID string,
		now time.Time,
	) (qcdomain.InboundQCInspection, error) {
		return current.Start(actorID, now)
	})
}

func (s InboundQCInspectionService) PassInboundQCInspection(
	ctx context.Context,
	input InboundQCActionInput,
) (InboundQCInspectionResult, error) {
	return s.recordDecision(ctx, input, qcdomain.InboundQCResultPass)
}

func (s InboundQCInspectionService) FailInboundQCInspection(
	ctx context.Context,
	input InboundQCActionInput,
) (InboundQCInspectionResult, error) {
	return s.recordDecision(ctx, input, qcdomain.InboundQCResultFail)
}

func (s InboundQCInspectionService) HoldInboundQCInspection(
	ctx context.Context,
	input InboundQCActionInput,
) (InboundQCInspectionResult, error) {
	return s.recordDecision(ctx, input, qcdomain.InboundQCResultHold)
}

func (s InboundQCInspectionService) PartialInboundQCInspection(
	ctx context.Context,
	input InboundQCActionInput,
) (InboundQCInspectionResult, error) {
	return s.recordDecision(ctx, input, qcdomain.InboundQCResultPartial)
}

func (s InboundQCInspectionService) recordDecision(
	ctx context.Context,
	input InboundQCActionInput,
	result qcdomain.InboundQCResult,
) (InboundQCInspectionResult, error) {
	return s.transition(ctx, input, actionForInboundQCResult(result), func(
		current qcdomain.InboundQCInspection,
		actorID string,
		now time.Time,
	) (qcdomain.InboundQCInspection, error) {
		decision, err := newInboundQCDecisionInput(current, input, result, actorID, now)
		if err != nil {
			return qcdomain.InboundQCInspection{}, err
		}

		return current.RecordDecision(decision)
	})
}

func (s InboundQCInspectionService) transition(
	ctx context.Context,
	input InboundQCActionInput,
	action string,
	apply func(qcdomain.InboundQCInspection, string, time.Time) (qcdomain.InboundQCInspection, error),
) (InboundQCInspectionResult, error) {
	if err := s.ensureReadyForWrite(); err != nil {
		return InboundQCInspectionResult{}, err
	}
	current, err := s.store.Get(ctx, input.ID)
	if err != nil {
		return InboundQCInspectionResult{}, err
	}

	now := s.clock()
	updated, err := apply(current, input.ActorID, now)
	if err != nil {
		return InboundQCInspectionResult{}, err
	}
	stockMovements, err := s.recordPassStockMovement(ctx, updated, input.ActorID, now)
	if err != nil {
		return InboundQCInspectionResult{}, err
	}
	if err := s.store.Save(ctx, updated); err != nil {
		return InboundQCInspectionResult{}, err
	}
	afterData := inboundQCAuditData(updated)
	if len(stockMovements) > 0 {
		afterData["stock_movement_count"] = len(stockMovements)
		afterData["stock_movement_no"] = stockMovements[0].MovementNo
		afterData["stock_movement_type"] = string(stockMovements[0].MovementType)
		afterData["target_stock_status"] = string(stockMovements[0].StockStatus)
	}
	log, err := newInboundQCAuditLog(
		input.ActorID,
		input.RequestID,
		action,
		updated,
		inboundQCAuditData(current),
		afterData,
		now,
	)
	if err != nil {
		return InboundQCInspectionResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return InboundQCInspectionResult{}, err
	}
	for _, movement := range stockMovements {
		movementLog, err := newInboundQCStockMovementAuditLog(input.ActorID, input.RequestID, updated, movement, now)
		if err != nil {
			return InboundQCInspectionResult{}, err
		}
		if err := s.auditLog.Record(ctx, movementLog); err != nil {
			return InboundQCInspectionResult{}, err
		}
	}

	return InboundQCInspectionResult{
		Inspection:     updated,
		PreviousStatus: current.Status,
		CurrentStatus:  updated.Status,
		PreviousResult: current.Result,
		CurrentResult:  updated.Result,
		AuditLogID:     log.ID,
	}, nil
}

func (s InboundQCInspectionService) recordPassStockMovement(
	ctx context.Context,
	inspection qcdomain.InboundQCInspection,
	actorID string,
	movementAt time.Time,
) ([]inventorydomain.StockMovement, error) {
	if inspection.Result != qcdomain.InboundQCResultPass {
		return nil, nil
	}
	if s.stockMovement == nil {
		return nil, errors.New("stock movement store is required")
	}

	movement, err := inventorydomain.NewStockMovement(inventorydomain.NewStockMovementInput{
		MovementNo:       fmt.Sprintf("%s-PASS-001", strings.ToUpper(inspection.ID)),
		MovementType:     inventorydomain.MovementPurchaseReceipt,
		OrgID:            inspection.OrgID,
		ItemID:           inspection.ItemID,
		BatchID:          inspection.BatchID,
		WarehouseID:      inspection.WarehouseID,
		BinID:            inspection.LocationID,
		Quantity:         inspection.PassedQuantity,
		BaseUOMCode:      inspection.UOMCode.String(),
		SourceQuantity:   inspection.PassedQuantity,
		SourceUOMCode:    inspection.UOMCode.String(),
		ConversionFactor: decimal.MustQuantity("1"),
		StockStatus:      inventorydomain.StockStatusAvailable,
		SourceDocType:    inboundQCStockMovementSourceDoc,
		SourceDocID:      inspection.ID,
		SourceDocLineID:  inspection.GoodsReceiptLineID,
		Reason:           "inbound qc pass released to available",
		CreatedBy:        actorID,
		MovementAt:       movementAt,
	})
	if err != nil {
		return nil, err
	}
	if err := s.stockMovement.Record(ctx, movement); err != nil {
		return nil, err
	}

	return []inventorydomain.StockMovement{movement}, nil
}

func (s InboundQCInspectionService) ensureReadyForWrite() error {
	if s.store == nil {
		return errors.New("inbound qc inspection store is required")
	}
	if s.receivingRead == nil {
		return errors.New("inbound qc receiving reader is required")
	}
	if s.auditLog == nil {
		return errors.New("audit log store is required")
	}
	if s.clock == nil {
		return errors.New("inbound qc inspection clock is required")
	}

	return nil
}

func (s InboundQCInspectionService) readInspectableReceivingLine(
	ctx context.Context,
	goodsReceiptID string,
	goodsReceiptLineID string,
) (inventorydomain.WarehouseReceiving, inventorydomain.WarehouseReceivingLine, error) {
	receipt, err := s.receivingRead.Get(ctx, goodsReceiptID)
	if err != nil {
		return inventorydomain.WarehouseReceiving{}, inventorydomain.WarehouseReceivingLine{}, ErrInboundQCReceivingNotFound
	}
	if receipt.Status != inventorydomain.WarehouseReceivingStatusInspectReady {
		return inventorydomain.WarehouseReceiving{}, inventorydomain.WarehouseReceivingLine{}, ErrInboundQCReceivingInvalidState
	}
	for _, line := range receipt.Lines {
		if line.ID == strings.TrimSpace(goodsReceiptLineID) {
			return receipt, line, nil
		}
	}

	return inventorydomain.WarehouseReceiving{}, inventorydomain.WarehouseReceivingLine{}, ErrInboundQCReceivingLineNotFound
}

func (s InboundQCInspectionService) ensureNoOpenInspection(
	ctx context.Context,
	goodsReceiptID string,
	goodsReceiptLineID string,
) error {
	rows, err := s.store.List(ctx, NewInboundQCInspectionFilter("", goodsReceiptID, goodsReceiptLineID, ""))
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.Status != qcdomain.InboundQCInspectionStatusCancelled {
			return ErrInboundQCDuplicateReceivingLine
		}
	}

	return nil
}

func (f InboundQCInspectionFilter) matches(inspection qcdomain.InboundQCInspection) bool {
	if f.Status != "" && qcdomain.NormalizeInboundQCInspectionStatus(inspection.Status) != f.Status {
		return false
	}
	if f.GoodsReceiptID != "" && inspection.GoodsReceiptID != f.GoodsReceiptID {
		return false
	}
	if f.GoodsReceiptLineID != "" && inspection.GoodsReceiptLineID != f.GoodsReceiptLineID {
		return false
	}
	if f.WarehouseID != "" && inspection.WarehouseID != f.WarehouseID {
		return false
	}

	return true
}

func (s *PrototypeInboundQCInspectionStore) List(
	_ context.Context,
	filter InboundQCInspectionFilter,
) ([]qcdomain.InboundQCInspection, error) {
	if s == nil {
		return nil, errors.New("inbound qc inspection store is required")
	}

	filter = NewInboundQCInspectionFilter(filter.Status, filter.GoodsReceiptID, filter.GoodsReceiptLineID, filter.WarehouseID)
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows := make([]qcdomain.InboundQCInspection, 0, len(s.inspections))
	for _, inspection := range s.inspections {
		if filter.matches(inspection) {
			rows = append(rows, inspection.Clone())
		}
	}
	sortInboundQCInspections(rows)

	return rows, nil
}

func (s *PrototypeInboundQCInspectionStore) Get(
	_ context.Context,
	id string,
) (qcdomain.InboundQCInspection, error) {
	if s == nil {
		return qcdomain.InboundQCInspection{}, errors.New("inbound qc inspection store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	inspection, ok := s.inspections[strings.TrimSpace(id)]
	if !ok {
		return qcdomain.InboundQCInspection{}, ErrInboundQCInspectionNotFound
	}

	return inspection.Clone(), nil
}

func (s *PrototypeInboundQCInspectionStore) Save(
	_ context.Context,
	inspection qcdomain.InboundQCInspection,
) error {
	if s == nil {
		return errors.New("inbound qc inspection store is required")
	}
	if strings.TrimSpace(inspection.ID) == "" {
		return qcdomain.ErrInboundQCInspectionRequiredField
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.inspections[inspection.ID] = inspection.Clone()

	return nil
}

func newInboundQCDecisionInput(
	current qcdomain.InboundQCInspection,
	input InboundQCActionInput,
	result qcdomain.InboundQCResult,
	actorID string,
	changedAt time.Time,
) (qcdomain.InboundQCDecisionInput, error) {
	passedQty, failedQty, holdQty, err := decisionQuantities(current.Quantity, input, result)
	if err != nil {
		return qcdomain.InboundQCDecisionInput{}, err
	}

	return qcdomain.InboundQCDecisionInput{
		Result:         result,
		PassedQuantity: passedQty,
		FailedQuantity: failedQty,
		HoldQuantity:   holdQty,
		Checklist:      checklistInputs(input.Checklist),
		Reason:         input.Reason,
		Note:           input.Note,
		ActorID:        actorID,
		ChangedAt:      changedAt,
	}, nil
}

func decisionQuantities(
	total decimal.Decimal,
	input InboundQCActionInput,
	result qcdomain.InboundQCResult,
) (decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
	zero := decimal.MustQuantity("0")
	switch result {
	case qcdomain.InboundQCResultPass:
		passedQty, err := parseOrDefaultQuantity(input.PassedQuantity, total)
		if err != nil {
			return "", "", "", err
		}
		return passedQty, zero, zero, nil
	case qcdomain.InboundQCResultFail:
		failedQty, err := parseOrDefaultQuantity(input.FailedQuantity, total)
		if err != nil {
			return "", "", "", err
		}
		return zero, failedQty, zero, nil
	case qcdomain.InboundQCResultHold:
		holdQty, err := parseOrDefaultQuantity(input.HoldQuantity, total)
		if err != nil {
			return "", "", "", err
		}
		return zero, zero, holdQty, nil
	default:
		passedQty, err := parseOrDefaultQuantity(input.PassedQuantity, zero)
		if err != nil {
			return "", "", "", err
		}
		failedQty, err := parseOrDefaultQuantity(input.FailedQuantity, zero)
		if err != nil {
			return "", "", "", err
		}
		holdQty, err := parseOrDefaultQuantity(input.HoldQuantity, zero)
		if err != nil {
			return "", "", "", err
		}
		return passedQty, failedQty, holdQty, nil
	}
}

func parseOrDefaultQuantity(value string, defaultValue decimal.Decimal) (decimal.Decimal, error) {
	if strings.TrimSpace(value) == "" {
		return defaultValue, nil
	}

	return decimal.ParseQuantity(value)
}

func checklistInputs(inputs []InboundQCChecklistInput) []qcdomain.NewInboundQCChecklistItemInput {
	if inputs == nil {
		return nil
	}
	items := make([]qcdomain.NewInboundQCChecklistItemInput, 0, len(inputs))
	for _, input := range inputs {
		items = append(items, qcdomain.NewInboundQCChecklistItemInput{
			ID:       input.ID,
			Code:     input.Code,
			Label:    input.Label,
			Required: input.Required,
			Status:   qcdomain.InboundQCChecklistStatus(input.Status),
			Note:     input.Note,
		})
	}

	return items
}

func defaultInboundQCChecklist() []qcdomain.NewInboundQCChecklistItemInput {
	return []qcdomain.NewInboundQCChecklistItemInput{
		{ID: "check-packaging", Code: "PACKAGING", Label: "Packaging condition", Required: true},
		{ID: "check-lot-expiry", Code: "LOT_EXPIRY", Label: "Lot and expiry match delivery", Required: true},
		{ID: "check-sample", Code: "SAMPLE", Label: "Sample retained when required", Required: false},
	}
}

func newInboundQCAuditLog(
	actorID string,
	requestID string,
	action string,
	inspection qcdomain.InboundQCInspection,
	beforeData map[string]any,
	afterData map[string]any,
	createdAt time.Time,
) (audit.Log, error) {
	return audit.NewLog(audit.NewLogInput{
		OrgID:      inspection.OrgID,
		ActorID:    strings.TrimSpace(actorID),
		Action:     action,
		EntityType: inboundQCAuditEntityType,
		EntityID:   inspection.ID,
		RequestID:  strings.TrimSpace(requestID),
		BeforeData: beforeData,
		AfterData:  afterData,
		Metadata: map[string]any{
			"goods_receipt_id":      inspection.GoodsReceiptID,
			"goods_receipt_line_id": inspection.GoodsReceiptLineID,
			"purchase_order_id":     inspection.PurchaseOrderID,
			"item_id":               inspection.ItemID,
			"sku":                   inspection.SKU,
			"batch_id":              inspection.BatchID,
			"lot_no":                inspection.LotNo,
			"warehouse_id":          inspection.WarehouseID,
			"source":                "inbound qc inspection",
		},
		CreatedAt: createdAt,
	})
}

func newInboundQCStockMovementAuditLog(
	actorID string,
	requestID string,
	inspection qcdomain.InboundQCInspection,
	movement inventorydomain.StockMovement,
	createdAt time.Time,
) (audit.Log, error) {
	direction, err := movement.Direction()
	if err != nil {
		return audit.Log{}, err
	}
	delta, err := movement.BalanceDelta()
	if err != nil {
		return audit.Log{}, err
	}

	return audit.NewLog(audit.NewLogInput{
		OrgID:      movement.OrgID,
		ActorID:    strings.TrimSpace(actorID),
		Action:     inboundQCStockMovementAction,
		EntityType: inboundQCStockMovementEntityType,
		EntityID:   movement.MovementNo,
		RequestID:  strings.TrimSpace(requestID),
		AfterData: map[string]any{
			"inspection_id":         inspection.ID,
			"goods_receipt_id":      inspection.GoodsReceiptID,
			"goods_receipt_line_id": inspection.GoodsReceiptLineID,
			"movement_no":           movement.MovementNo,
			"movement_type":         string(movement.MovementType),
			"direction":             string(direction),
			"sku":                   inspection.SKU,
			"batch_id":              inspection.BatchID,
			"lot_no":                inspection.LotNo,
			"quantity":              movement.Quantity.String(),
			"base_uom_code":         movement.BaseUOMCode.String(),
			"stock_status":          string(movement.StockStatus),
			"source_doc_type":       movement.SourceDocType,
			"source_doc_id":         movement.SourceDocID,
			"source_doc_line_id":    movement.SourceDocLineID,
			"delta_on_hand":         delta.OnHand.String(),
			"delta_reserved":        delta.Reserved.String(),
			"delta_available":       delta.Available.String(),
		},
		Metadata: map[string]any{
			"source":       "inbound qc pass stock movement",
			"warehouse_id": movement.WarehouseID,
			"location_id":  movement.BinID,
			"reason":       movement.Reason,
		},
		CreatedAt: createdAt,
	})
}

func inboundQCAuditData(inspection qcdomain.InboundQCInspection) map[string]any {
	return map[string]any{
		"status":                 string(inspection.Status),
		"result":                 string(inspection.Result),
		"goods_receipt_id":       inspection.GoodsReceiptID,
		"goods_receipt_line_id":  inspection.GoodsReceiptLineID,
		"purchase_order_id":      inspection.PurchaseOrderID,
		"purchase_order_line_id": inspection.PurchaseOrderLineID,
		"item_id":                inspection.ItemID,
		"sku":                    inspection.SKU,
		"batch_id":               inspection.BatchID,
		"lot_no":                 inspection.LotNo,
		"quantity":               inspection.Quantity.String(),
		"uom_code":               inspection.UOMCode.String(),
		"passed_quantity":        inboundQCQuantityString(inspection.PassedQuantity),
		"failed_quantity":        inboundQCQuantityString(inspection.FailedQuantity),
		"hold_quantity":          inboundQCQuantityString(inspection.HoldQuantity),
		"reason":                 inspection.Reason,
		"checklist_count":        len(inspection.Checklist),
	}
}

func inboundQCQuantityString(value decimal.Decimal) string {
	if strings.TrimSpace(value.String()) == "" {
		return decimal.MustQuantity("0").String()
	}

	return value.String()
}

func actionForInboundQCResult(result qcdomain.InboundQCResult) string {
	switch result {
	case qcdomain.InboundQCResultPass:
		return "qc.inbound_inspection.passed"
	case qcdomain.InboundQCResultFail:
		return "qc.inbound_inspection.failed"
	case qcdomain.InboundQCResultHold:
		return "qc.inbound_inspection.held"
	default:
		return "qc.inbound_inspection.partial"
	}
}

func defaultInspectionID(goodsReceiptID string, goodsReceiptLineID string) string {
	return fmt.Sprintf("iqc-%s-%s", strings.TrimSpace(goodsReceiptID), strings.TrimSpace(goodsReceiptLineID))
}

func sortInboundQCInspections(rows []qcdomain.InboundQCInspection) {
	sort.SliceStable(rows, func(i, j int) bool {
		left := rows[i]
		right := rows[j]
		if !left.UpdatedAt.Equal(right.UpdatedAt) {
			return left.UpdatedAt.After(right.UpdatedAt)
		}

		return left.ID < right.ID
	})
}
