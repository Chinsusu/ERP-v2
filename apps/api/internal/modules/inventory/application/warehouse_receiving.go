package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	masterdatadomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	purchasedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/purchase/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrWarehouseReceivingNotFound = errors.New("warehouse receiving not found")
var ErrReceivingInvalidLocation = errors.New("warehouse receiving location is invalid")
var ErrReceivingBatchMismatch = errors.New("warehouse receiving batch does not match line")
var ErrReceivingPurchaseOrderRequired = errors.New("warehouse receiving purchase order reader is required")
var ErrReceivingPurchaseOrderInvalidState = errors.New("warehouse receiving purchase order status is invalid")
var ErrReceivingPurchaseOrderMismatch = errors.New("warehouse receiving purchase order data mismatch")
var ErrReceivingQuantityExceedsPurchaseOrder = errors.New("warehouse receiving quantity exceeds purchase order remaining quantity")

const (
	receivingAuditEntityType = "inventory.receiving"
	receivingSourceDocType   = "warehouse_receiving"
	defaultReceivingOrgID    = "org-my-pham"
)

type WarehouseReceivingStore interface {
	List(ctx context.Context, filter domain.WarehouseReceivingFilter) ([]domain.WarehouseReceiving, error)
	Get(ctx context.Context, id string) (domain.WarehouseReceiving, error)
	Save(ctx context.Context, receipt domain.WarehouseReceiving) error
}

type WarehouseReceivingLocationReader interface {
	GetLocation(ctx context.Context, id string) (masterdatadomain.Location, error)
}

type WarehouseReceivingBatchReader interface {
	GetBatch(ctx context.Context, id string) (domain.Batch, error)
}

type WarehouseReceivingPurchaseOrderReader interface {
	GetPurchaseOrder(ctx context.Context, id string) (purchasedomain.PurchaseOrder, error)
}

type WarehouseReceivingService struct {
	store             WarehouseReceivingStore
	locationRead      WarehouseReceivingLocationReader
	batchRead         WarehouseReceivingBatchReader
	purchaseOrderRead WarehouseReceivingPurchaseOrderReader
	movement          StockMovementStore
	auditLog          audit.LogStore
	clock             func() time.Time
}

type CreateWarehouseReceivingInput struct {
	ID               string
	OrgID            string
	ReceiptNo        string
	WarehouseID      string
	LocationID       string
	ReferenceDocType string
	ReferenceDocID   string
	SupplierID       string
	DeliveryNoteNo   string
	Lines            []CreateWarehouseReceivingLineInput
	ActorID          string
	RequestID        string
}

type CreateWarehouseReceivingLineInput struct {
	ID                  string
	PurchaseOrderLineID string
	ItemID              string
	SKU                 string
	ItemName            string
	BatchID             string
	BatchNo             string
	LotNo               string
	ExpiryDate          string
	Quantity            string
	UOMCode             string
	BaseUOMCode         string
	PackagingStatus     string
	QCStatus            string
}

type WarehouseReceivingTransitionInput struct {
	ID        string
	ActorID   string
	RequestID string
}

type WarehouseReceivingResult struct {
	Receipt    domain.WarehouseReceiving
	AuditLogID string
}

type PrototypeWarehouseReceivingStore struct {
	mu       sync.RWMutex
	receipts map[string]domain.WarehouseReceiving
}

func NewWarehouseReceivingService(
	store WarehouseReceivingStore,
	locationRead WarehouseReceivingLocationReader,
	batchRead WarehouseReceivingBatchReader,
	movement StockMovementStore,
	auditLog audit.LogStore,
) WarehouseReceivingService {
	return WarehouseReceivingService{
		store:        store,
		locationRead: locationRead,
		batchRead:    batchRead,
		movement:     movement,
		auditLog:     auditLog,
		clock:        func() time.Time { return time.Now().UTC() },
	}
}

func (s WarehouseReceivingService) WithPurchaseOrderReader(
	reader WarehouseReceivingPurchaseOrderReader,
) WarehouseReceivingService {
	s.purchaseOrderRead = reader

	return s
}

func NewPrototypeWarehouseReceivingStore() *PrototypeWarehouseReceivingStore {
	store := &PrototypeWarehouseReceivingStore{receipts: make(map[string]domain.WarehouseReceiving)}
	for _, receipt := range prototypeWarehouseReceivings() {
		store.receipts[receipt.ID] = receipt.Clone()
	}

	return store
}

func (s WarehouseReceivingService) ListWarehouseReceivings(
	ctx context.Context,
	filter domain.WarehouseReceivingFilter,
) ([]domain.WarehouseReceiving, error) {
	if s.store == nil {
		return nil, errors.New("warehouse receiving store is required")
	}

	return s.store.List(ctx, filter)
}

func (s WarehouseReceivingService) GetWarehouseReceiving(
	ctx context.Context,
	id string,
) (domain.WarehouseReceiving, error) {
	if s.store == nil {
		return domain.WarehouseReceiving{}, errors.New("warehouse receiving store is required")
	}

	return s.store.Get(ctx, id)
}

func (s WarehouseReceivingService) CreateWarehouseReceiving(
	ctx context.Context,
	input CreateWarehouseReceivingInput,
) (WarehouseReceivingResult, error) {
	if err := s.ensureReadyForWrite(); err != nil {
		return WarehouseReceivingResult{}, err
	}

	now := s.clock()
	location, err := s.validateReceivingLocation(ctx, input.WarehouseID, input.LocationID)
	if err != nil {
		return WarehouseReceivingResult{}, err
	}
	lines, err := s.newReceivingLines(ctx, input.Lines)
	if err != nil {
		return WarehouseReceivingResult{}, err
	}
	if err := s.validatePurchaseOrderReceiving(ctx, input, lines, location); err != nil {
		return WarehouseReceivingResult{}, err
	}

	orgID := strings.TrimSpace(input.OrgID)
	if orgID == "" {
		orgID = defaultReceivingOrgID
	}
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = newReceivingID(now)
	}
	receiptNo := strings.TrimSpace(input.ReceiptNo)
	if receiptNo == "" {
		receiptNo = newReceivingNo(now)
	}
	receipt, err := domain.NewWarehouseReceiving(domain.NewWarehouseReceivingInput{
		ID:               id,
		OrgID:            orgID,
		ReceiptNo:        receiptNo,
		WarehouseID:      location.WarehouseID,
		WarehouseCode:    location.WarehouseCode,
		LocationID:       location.ID,
		LocationCode:     location.Code,
		ReferenceDocType: input.ReferenceDocType,
		ReferenceDocID:   input.ReferenceDocID,
		SupplierID:       input.SupplierID,
		DeliveryNoteNo:   input.DeliveryNoteNo,
		Lines:            lines,
		CreatedBy:        input.ActorID,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		return WarehouseReceivingResult{}, err
	}
	if err := s.store.Save(ctx, receipt); err != nil {
		return WarehouseReceivingResult{}, err
	}
	log, err := newWarehouseReceivingAuditLog(
		input.ActorID,
		input.RequestID,
		"inventory.receiving.created",
		receipt,
		nil,
		receivingAuditData(receipt),
		now,
	)
	if err != nil {
		return WarehouseReceivingResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return WarehouseReceivingResult{}, err
	}

	return WarehouseReceivingResult{Receipt: receipt, AuditLogID: log.ID}, nil
}

func (s WarehouseReceivingService) SubmitWarehouseReceiving(
	ctx context.Context,
	input WarehouseReceivingTransitionInput,
) (WarehouseReceivingResult, error) {
	return s.transition(ctx, input, "inventory.receiving.submitted", func(receipt domain.WarehouseReceiving, actorID string, at time.Time) (domain.WarehouseReceiving, error) {
		return receipt.Submit(actorID, at)
	})
}

func (s WarehouseReceivingService) MarkWarehouseReceivingInspectReady(
	ctx context.Context,
	input WarehouseReceivingTransitionInput,
) (WarehouseReceivingResult, error) {
	return s.transition(ctx, input, "inventory.receiving.inspect_ready", func(receipt domain.WarehouseReceiving, actorID string, at time.Time) (domain.WarehouseReceiving, error) {
		return receipt.MarkInspectReady(actorID, at)
	})
}

func (s WarehouseReceivingService) PostWarehouseReceiving(
	ctx context.Context,
	input WarehouseReceivingTransitionInput,
) (WarehouseReceivingResult, error) {
	if err := s.ensureReadyForWrite(); err != nil {
		return WarehouseReceivingResult{}, err
	}

	current, err := s.store.Get(ctx, input.ID)
	if err != nil {
		return WarehouseReceivingResult{}, err
	}

	now := s.clock()
	posted, err := current.Post(input.ActorID, now)
	if err != nil {
		return WarehouseReceivingResult{}, err
	}
	if err := s.validatePostingBatches(ctx, posted); err != nil {
		return WarehouseReceivingResult{}, err
	}
	movements, err := s.newReceivingMovements(posted, input.ActorID, now)
	if err != nil {
		return WarehouseReceivingResult{}, err
	}
	for _, movement := range movements {
		if err := s.movement.Record(ctx, movement); err != nil {
			return WarehouseReceivingResult{}, err
		}
	}
	posted = posted.AttachStockMovements(movements)
	if err := s.store.Save(ctx, posted); err != nil {
		return WarehouseReceivingResult{}, err
	}
	afterData := receivingAuditData(posted)
	afterData["stock_movement_count"] = len(posted.StockMovements)
	log, err := newWarehouseReceivingAuditLog(
		input.ActorID,
		input.RequestID,
		"inventory.receiving.posted",
		posted,
		receivingAuditData(current),
		afterData,
		now,
	)
	if err != nil {
		return WarehouseReceivingResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return WarehouseReceivingResult{}, err
	}

	return WarehouseReceivingResult{Receipt: posted, AuditLogID: log.ID}, nil
}

func (s WarehouseReceivingService) transition(
	ctx context.Context,
	input WarehouseReceivingTransitionInput,
	action string,
	apply func(domain.WarehouseReceiving, string, time.Time) (domain.WarehouseReceiving, error),
) (WarehouseReceivingResult, error) {
	if err := s.ensureReadyForWrite(); err != nil {
		return WarehouseReceivingResult{}, err
	}
	current, err := s.store.Get(ctx, input.ID)
	if err != nil {
		return WarehouseReceivingResult{}, err
	}

	now := s.clock()
	updated, err := apply(current, input.ActorID, now)
	if err != nil {
		return WarehouseReceivingResult{}, err
	}
	if err := s.store.Save(ctx, updated); err != nil {
		return WarehouseReceivingResult{}, err
	}
	log, err := newWarehouseReceivingAuditLog(
		input.ActorID,
		input.RequestID,
		action,
		updated,
		receivingAuditData(current),
		receivingAuditData(updated),
		now,
	)
	if err != nil {
		return WarehouseReceivingResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return WarehouseReceivingResult{}, err
	}

	return WarehouseReceivingResult{Receipt: updated, AuditLogID: log.ID}, nil
}

func (s WarehouseReceivingService) ensureReadyForWrite() error {
	if s.store == nil {
		return errors.New("warehouse receiving store is required")
	}
	if s.auditLog == nil {
		return errors.New("audit log store is required")
	}
	if s.clock == nil {
		return errors.New("warehouse receiving clock is required")
	}

	return nil
}

func (s WarehouseReceivingService) validateReceivingLocation(
	ctx context.Context,
	warehouseID string,
	locationID string,
) (masterdatadomain.Location, error) {
	if s.locationRead == nil {
		return masterdatadomain.Location{}, errors.New("warehouse location reader is required")
	}
	location, err := s.locationRead.GetLocation(ctx, locationID)
	if err != nil {
		return masterdatadomain.Location{}, ErrReceivingInvalidLocation
	}
	if location.WarehouseID != strings.TrimSpace(warehouseID) ||
		location.Status != masterdatadomain.LocationStatusActive ||
		!location.AllowReceive {
		return masterdatadomain.Location{}, ErrReceivingInvalidLocation
	}

	return location, nil
}

func (s WarehouseReceivingService) newReceivingLines(
	ctx context.Context,
	inputs []CreateWarehouseReceivingLineInput,
) ([]domain.NewWarehouseReceivingLineInput, error) {
	lines := make([]domain.NewWarehouseReceivingLineInput, 0, len(inputs))
	for index, input := range inputs {
		expiryDate, err := parseReceivingDate(input.ExpiryDate)
		if err != nil {
			return nil, err
		}
		line := domain.NewWarehouseReceivingLineInput{
			ID:                  strings.TrimSpace(input.ID),
			PurchaseOrderLineID: strings.TrimSpace(input.PurchaseOrderLineID),
			ItemID:              strings.TrimSpace(input.ItemID),
			SKU:                 strings.ToUpper(strings.TrimSpace(input.SKU)),
			ItemName:            strings.TrimSpace(input.ItemName),
			BatchID:             strings.TrimSpace(input.BatchID),
			BatchNo:             domain.NormalizeBatchNo(input.BatchNo),
			LotNo:               domain.NormalizeBatchNo(input.LotNo),
			ExpiryDate:          expiryDate,
			Quantity:            decimal.Decimal(strings.TrimSpace(input.Quantity)),
			UOMCode:             input.UOMCode,
			BaseUOMCode:         input.BaseUOMCode,
			PackagingStatus:     domain.ReceivingPackagingStatus(input.PackagingStatus),
			QCStatus:            domain.QCStatus(input.QCStatus),
		}
		if line.ID == "" {
			line.ID = fmt.Sprintf("line-%03d", index+1)
		}
		if line.BatchID != "" {
			batch, err := s.readBatch(ctx, line.BatchID)
			if err != nil {
				return nil, err
			}
			if err := hydrateLineFromBatch(&line, batch); err != nil {
				return nil, err
			}
		}
		lines = append(lines, line)
	}

	return lines, nil
}

func (s WarehouseReceivingService) readBatch(ctx context.Context, id string) (domain.Batch, error) {
	if s.batchRead == nil {
		return domain.Batch{}, errors.New("batch reader is required")
	}

	return s.batchRead.GetBatch(ctx, id)
}

func (s WarehouseReceivingService) validatePurchaseOrderReceiving(
	ctx context.Context,
	input CreateWarehouseReceivingInput,
	lines []domain.NewWarehouseReceivingLineInput,
	location masterdatadomain.Location,
) error {
	if domain.NormalizeReceivingReferenceDocType(input.ReferenceDocType) != domain.ReceivingReferenceDocTypePurchaseOrder {
		return nil
	}
	if s.purchaseOrderRead == nil {
		return ErrReceivingPurchaseOrderRequired
	}
	order, err := s.purchaseOrderRead.GetPurchaseOrder(ctx, input.ReferenceDocID)
	if err != nil {
		return ErrReceivingPurchaseOrderMismatch
	}
	if !isReceivingPurchaseOrderOpen(order.Status) {
		return ErrReceivingPurchaseOrderInvalidState
	}
	if order.SupplierID != strings.TrimSpace(input.SupplierID) ||
		order.WarehouseID != location.WarehouseID {
		return ErrReceivingPurchaseOrderMismatch
	}
	for _, line := range lines {
		if err := s.validatePurchaseOrderReceivingLine(ctx, order, line); err != nil {
			return err
		}
	}

	return nil
}

func (s WarehouseReceivingService) validatePurchaseOrderReceivingLine(
	ctx context.Context,
	order purchasedomain.PurchaseOrder,
	line domain.NewWarehouseReceivingLineInput,
) error {
	poLine, ok := purchaseOrderLineByID(order, line.PurchaseOrderLineID)
	if !ok {
		return ErrReceivingPurchaseOrderMismatch
	}
	if strings.TrimSpace(line.BatchID) == "" ||
		strings.TrimSpace(line.LotNo) == "" ||
		line.ExpiryDate.IsZero() {
		return domain.ErrReceivingRequiredField
	}
	if line.ItemID != poLine.ItemID ||
		strings.ToUpper(strings.TrimSpace(line.SKU)) != poLine.SKUCode {
		return ErrReceivingPurchaseOrderMismatch
	}
	uomCode, err := decimal.NormalizeUOMCode(line.UOMCode)
	if err != nil || uomCode != poLine.UOMCode {
		return ErrReceivingPurchaseOrderMismatch
	}
	baseUOMCode, err := decimal.NormalizeUOMCode(line.BaseUOMCode)
	if err != nil || baseUOMCode != poLine.BaseUOMCode {
		return ErrReceivingPurchaseOrderMismatch
	}
	batch, err := s.readBatch(ctx, line.BatchID)
	if err != nil {
		return err
	}
	if batch.SupplierID != "" && batch.SupplierID != order.SupplierID {
		return ErrReceivingPurchaseOrderMismatch
	}
	if batch.BatchNo != line.BatchNo || !sameDate(batch.ExpiryDate, line.ExpiryDate) {
		return ErrReceivingPurchaseOrderMismatch
	}
	remainingQty, err := decimal.SubtractQuantity(poLine.BaseOrderedQty, poLine.BaseReceivedQty)
	if err != nil || remainingQty.IsNegative() {
		return ErrReceivingPurchaseOrderMismatch
	}
	afterReceive, err := decimal.SubtractQuantity(remainingQty, line.Quantity)
	if err != nil {
		return err
	}
	if afterReceive.IsNegative() {
		return ErrReceivingQuantityExceedsPurchaseOrder
	}

	return nil
}

func isReceivingPurchaseOrderOpen(status purchasedomain.PurchaseOrderStatus) bool {
	switch purchasedomain.NormalizePurchaseOrderStatus(status) {
	case purchasedomain.PurchaseOrderStatusApproved,
		purchasedomain.PurchaseOrderStatusPartiallyReceived:
		return true
	default:
		return false
	}
}

func purchaseOrderLineByID(
	order purchasedomain.PurchaseOrder,
	id string,
) (purchasedomain.PurchaseOrderLine, bool) {
	id = strings.TrimSpace(id)
	for _, line := range order.Lines {
		if line.ID == id {
			return line, true
		}
	}

	return purchasedomain.PurchaseOrderLine{}, false
}

func sameDate(left time.Time, right time.Time) bool {
	if left.IsZero() || right.IsZero() {
		return left.IsZero() && right.IsZero()
	}

	return left.UTC().Format("2006-01-02") == right.UTC().Format("2006-01-02")
}

func hydrateLineFromBatch(line *domain.NewWarehouseReceivingLineInput, batch domain.Batch) error {
	if strings.TrimSpace(line.ItemID) != "" && strings.TrimSpace(line.ItemID) != batch.ItemID {
		return ErrReceivingBatchMismatch
	}
	if strings.TrimSpace(line.SKU) != "" && strings.ToUpper(strings.TrimSpace(line.SKU)) != batch.SKU {
		return ErrReceivingBatchMismatch
	}
	line.ItemID = batch.ItemID
	line.SKU = batch.SKU
	if strings.TrimSpace(line.ItemName) == "" {
		line.ItemName = batch.ItemName
	}
	if strings.TrimSpace(line.BatchNo) == "" {
		line.BatchNo = batch.BatchNo
	}
	if strings.TrimSpace(line.LotNo) == "" {
		line.LotNo = batch.BatchNo
	}
	if line.ExpiryDate.IsZero() {
		line.ExpiryDate = batch.ExpiryDate
	}
	if strings.TrimSpace(string(line.QCStatus)) == "" {
		line.QCStatus = batch.QCStatus
	}

	return nil
}

func (s WarehouseReceivingService) validatePostingBatches(ctx context.Context, receipt domain.WarehouseReceiving) error {
	for _, line := range receipt.Lines {
		batch, err := s.readBatch(ctx, line.BatchID)
		if err != nil {
			return err
		}
		if line.ItemID != batch.ItemID || line.SKU != batch.SKU {
			return ErrReceivingBatchMismatch
		}
	}

	return nil
}

func (s WarehouseReceivingService) newReceivingMovements(
	receipt domain.WarehouseReceiving,
	actorID string,
	movementAt time.Time,
) ([]domain.StockMovement, error) {
	if s.movement == nil {
		return nil, errors.New("stock movement store is required")
	}
	movements := make([]domain.StockMovement, 0, len(receipt.Lines))
	for index, line := range receipt.Lines {
		movement, err := domain.NewStockMovement(domain.NewStockMovementInput{
			MovementNo:       fmt.Sprintf("%s-MV-%03d", receipt.ReceiptNo, index+1),
			MovementType:     domain.MovementPurchaseReceipt,
			OrgID:            receipt.OrgID,
			ItemID:           line.ItemID,
			BatchID:          line.BatchID,
			WarehouseID:      line.WarehouseID,
			BinID:            line.LocationID,
			Quantity:         line.Quantity,
			BaseUOMCode:      line.BaseUOMCode.String(),
			SourceQuantity:   line.Quantity,
			SourceUOMCode:    line.UOMCode.String(),
			ConversionFactor: decimal.MustQuantity("1"),
			StockStatus:      stockStatusForReceivingLine(line),
			SourceDocType:    receivingSourceDocType,
			SourceDocID:      receipt.ID,
			SourceDocLineID:  line.ID,
			Reason:           "warehouse receiving posted",
			CreatedBy:        actorID,
			MovementAt:       movementAt,
		})
		if err != nil {
			return nil, err
		}
		movements = append(movements, movement)
	}

	return movements, nil
}

func (s *PrototypeWarehouseReceivingStore) List(
	_ context.Context,
	filter domain.WarehouseReceivingFilter,
) ([]domain.WarehouseReceiving, error) {
	if s == nil {
		return nil, errors.New("warehouse receiving store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows := make([]domain.WarehouseReceiving, 0, len(s.receipts))
	for _, receipt := range s.receipts {
		if filter.Matches(receipt) {
			rows = append(rows, receipt.Clone())
		}
	}
	domain.SortWarehouseReceivings(rows)

	return rows, nil
}

func (s *PrototypeWarehouseReceivingStore) Get(_ context.Context, id string) (domain.WarehouseReceiving, error) {
	if s == nil {
		return domain.WarehouseReceiving{}, errors.New("warehouse receiving store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	receipt, ok := s.receipts[strings.TrimSpace(id)]
	if !ok {
		return domain.WarehouseReceiving{}, ErrWarehouseReceivingNotFound
	}

	return receipt.Clone(), nil
}

func (s *PrototypeWarehouseReceivingStore) Save(_ context.Context, receipt domain.WarehouseReceiving) error {
	if s == nil {
		return errors.New("warehouse receiving store is required")
	}
	if strings.TrimSpace(receipt.ID) == "" {
		return domain.ErrReceivingRequiredField
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.receipts[receipt.ID] = receipt.Clone()

	return nil
}

func stockStatusForReceivingLine(line domain.WarehouseReceivingLine) domain.StockStatus {
	if line.QCStatus == domain.QCStatusPass {
		return domain.StockStatusAvailable
	}

	return domain.StockStatusQCHold
}

func newWarehouseReceivingAuditLog(
	actorID string,
	requestID string,
	action string,
	receipt domain.WarehouseReceiving,
	beforeData map[string]any,
	afterData map[string]any,
	createdAt time.Time,
) (audit.Log, error) {
	return audit.NewLog(audit.NewLogInput{
		OrgID:      receipt.OrgID,
		ActorID:    strings.TrimSpace(actorID),
		Action:     action,
		EntityType: receivingAuditEntityType,
		EntityID:   receipt.ID,
		RequestID:  strings.TrimSpace(requestID),
		BeforeData: beforeData,
		AfterData:  afterData,
		Metadata: map[string]any{
			"receipt_no":         receipt.ReceiptNo,
			"reference_doc_type": receipt.ReferenceDocType,
			"reference_doc_id":   receipt.ReferenceDocID,
			"delivery_note_no":   receipt.DeliveryNoteNo,
			"warehouse_id":       receipt.WarehouseID,
			"location_id":        receipt.LocationID,
			"source":             "warehouse receiving",
		},
		CreatedAt: createdAt,
	})
}

func receivingAuditData(receipt domain.WarehouseReceiving) map[string]any {
	return map[string]any{
		"receipt_no":         receipt.ReceiptNo,
		"warehouse_id":       receipt.WarehouseID,
		"location_id":        receipt.LocationID,
		"status":             string(receipt.Status),
		"reference_doc_type": receipt.ReferenceDocType,
		"reference_doc_id":   receipt.ReferenceDocID,
		"supplier_id":        receipt.SupplierID,
		"delivery_note_no":   receipt.DeliveryNoteNo,
		"line_count":         len(receipt.Lines),
	}
}

func parseReceivingDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, domain.ErrReceivingRequiredField
	}

	return parsed, nil
}

func newReceivingID(now time.Time) string {
	return fmt.Sprintf("grn_%d", now.UTC().UnixNano())
}

func newReceivingNo(now time.Time) string {
	return fmt.Sprintf("GRN-%s-%06d", now.UTC().Format("060102"), now.UTC().UnixNano()%1000000)
}

func prototypeWarehouseReceivings() []domain.WarehouseReceiving {
	baseTime := time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC)
	draft := mustWarehouseReceiving(domain.NewWarehouseReceiving(domain.NewWarehouseReceivingInput{
		ID:               "grn-hcm-260427-draft",
		OrgID:            defaultReceivingOrgID,
		ReceiptNo:        "GRN-260427-0001",
		WarehouseID:      "wh-hcm-fg",
		WarehouseCode:    "WH-HCM-FG",
		LocationID:       "loc-hcm-fg-recv-01",
		LocationCode:     "FG-RECV-01",
		ReferenceDocType: "purchase_order",
		ReferenceDocID:   "PO-260427-0001",
		SupplierID:       "supplier-local",
		DeliveryNoteNo:   "DN-260427-0001",
		Lines: []domain.NewWarehouseReceivingLineInput{
			{
				ID:                  "grn-line-draft-001",
				PurchaseOrderLineID: "po-line-260427-0001-001",
				ItemID:              "item-serum-30ml",
				SKU:                 "SERUM-30ML",
				ItemName:            "Vitamin C Serum",
				BatchID:             "batch-serum-2604a",
				BatchNo:             "LOT-2604A",
				LotNo:               "LOT-2604A",
				ExpiryDate:          time.Date(2027, 4, 1, 0, 0, 0, 0, time.UTC),
				Quantity:            decimal.MustQuantity("24"),
				UOMCode:             "EA",
				BaseUOMCode:         "EA",
				PackagingStatus:     domain.ReceivingPackagingStatusIntact,
				QCStatus:            domain.QCStatusHold,
			},
		},
		CreatedBy: "user-warehouse-lead",
		CreatedAt: baseTime,
		UpdatedAt: baseTime,
	}))
	submitted := mustWarehouseReceiving(draft.Submit("user-warehouse-lead", baseTime.Add(30*time.Minute)))
	submitted.ID = "grn-hcm-260427-submitted"
	submitted.ReceiptNo = "GRN-260427-0002"
	submitted.ReferenceDocID = "PO-260427-0002"
	submitted.DeliveryNoteNo = "DN-260427-0002"
	submitted.Lines[0].PurchaseOrderLineID = "po-line-260427-0002-001"
	inspectReady := mustWarehouseReceiving(submitted.MarkInspectReady("user-qa", baseTime.Add(60*time.Minute)))
	inspectReady.ID = "grn-hcm-260427-inspect"
	inspectReady.ReceiptNo = "GRN-260427-0003"
	inspectReady.ReferenceDocID = "PO-260427-0003"
	inspectReady.DeliveryNoteNo = "DN-260427-0003"
	inspectReady.Lines[0].PurchaseOrderLineID = "po-line-260427-0003-001"
	inspectReady.Lines[0].ItemID = "item-cream-50g"
	inspectReady.Lines[0].SKU = "CREAM-50G"
	inspectReady.Lines[0].ItemName = "Moisturizing Cream"
	inspectReady.Lines[0].BatchID = "batch-cream-2603b"
	inspectReady.Lines[0].BatchNo = "LOT-2603B"
	inspectReady.Lines[0].LotNo = "LOT-2603B"
	inspectReady.Lines[0].ExpiryDate = time.Date(2028, 3, 1, 0, 0, 0, 0, time.UTC)
	inspectReady.Lines[0].QCStatus = domain.QCStatusPass

	return []domain.WarehouseReceiving{draft, submitted, inspectReady}
}

func mustWarehouseReceiving(receipt domain.WarehouseReceiving, err error) domain.WarehouseReceiving {
	if err != nil {
		panic(err)
	}

	return receipt
}
