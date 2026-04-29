package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	masterdatadomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	purchasedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/purchase/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
	apperrors "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/errors"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

var ErrPurchaseOrderNotFound = errors.New("purchase order not found")
var ErrPurchaseOrderVersionConflict = errors.New("purchase order version conflict")

const (
	ErrorCodePurchaseOrderNotFound        response.ErrorCode = "PURCHASE_ORDER_NOT_FOUND"
	ErrorCodePurchaseOrderValidation      response.ErrorCode = "PURCHASE_ORDER_VALIDATION_ERROR"
	ErrorCodePurchaseOrderInvalidState    response.ErrorCode = "PURCHASE_ORDER_INVALID_STATE"
	ErrorCodePurchaseOrderVersionConflict response.ErrorCode = "PURCHASE_ORDER_VERSION_CONFLICT"

	defaultPurchaseOrderOrgID = "org-my-pham"
	purchaseOrderEntityType   = "purchase.purchase_order"
)

type PurchaseOrderStore interface {
	List(ctx context.Context, filter PurchaseOrderFilter) ([]purchasedomain.PurchaseOrder, error)
	Get(ctx context.Context, id string) (purchasedomain.PurchaseOrder, error)
	WithinTx(ctx context.Context, fn func(context.Context, PurchaseOrderTx) error) error
}

type PurchaseOrderTx interface {
	GetForUpdate(ctx context.Context, id string) (purchasedomain.PurchaseOrder, error)
	Save(ctx context.Context, order purchasedomain.PurchaseOrder) error
	RecordAudit(ctx context.Context, log audit.Log) error
}

type PurchaseOrderSupplierReader interface {
	GetSupplier(ctx context.Context, id string) (masterdatadomain.Supplier, error)
}

type PurchaseOrderItemReader interface {
	Get(ctx context.Context, id string) (masterdatadomain.Item, error)
}

type PurchaseOrderWarehouseReader interface {
	GetWarehouse(ctx context.Context, id string) (masterdatadomain.Warehouse, error)
}

type PurchaseOrderUOMConverter interface {
	ConvertToBase(ctx context.Context, input ConvertPurchaseOrderLineToBaseInput) (ConvertPurchaseOrderLineToBaseResult, error)
}

type ConvertPurchaseOrderLineToBaseInput struct {
	ItemID      string
	SKU         string
	Quantity    decimal.Decimal
	FromUOMCode string
	BaseUOMCode string
}

type ConvertPurchaseOrderLineToBaseResult struct {
	Quantity         decimal.Decimal
	SourceUOMCode    decimal.UOMCode
	BaseQuantity     decimal.Decimal
	BaseUOMCode      decimal.UOMCode
	ConversionFactor decimal.Decimal
}

type PurchaseOrderService struct {
	store         PurchaseOrderStore
	supplierRead  PurchaseOrderSupplierReader
	itemRead      PurchaseOrderItemReader
	warehouseRead PurchaseOrderWarehouseReader
	uomConverter  PurchaseOrderUOMConverter
	clock         func() time.Time
}

type PurchaseOrderFilter struct {
	Search       string
	Statuses     []purchasedomain.PurchaseOrderStatus
	SupplierID   string
	WarehouseID  string
	ExpectedFrom string
	ExpectedTo   string
}

type CreatePurchaseOrderInput struct {
	ID           string
	OrgID        string
	PONo         string
	SupplierID   string
	WarehouseID  string
	ExpectedDate string
	CurrencyCode string
	Note         string
	Lines        []PurchaseOrderLineInput
	ActorID      string
	RequestID    string
}

type UpdatePurchaseOrderInput struct {
	ID              string
	SupplierID      string
	WarehouseID     string
	ExpectedDate    string
	Note            string
	Lines           []PurchaseOrderLineInput
	ExpectedVersion int
	ActorID         string
	RequestID       string
}

type PurchaseOrderLineInput struct {
	ID           string
	LineNo       int
	ItemID       string
	OrderedQty   string
	UOMCode      string
	UnitPrice    string
	CurrencyCode string
	ExpectedDate string
	Note         string
}

type PurchaseOrderActionInput struct {
	ID              string
	ExpectedVersion int
	Reason          string
	Note            string
	ActorID         string
	RequestID       string
}

type PurchaseOrderResult struct {
	PurchaseOrder purchasedomain.PurchaseOrder
	AuditLogID    string
}

type PurchaseOrderActionResult struct {
	PurchaseOrder  purchasedomain.PurchaseOrder
	PreviousStatus purchasedomain.PurchaseOrderStatus
	CurrentStatus  purchasedomain.PurchaseOrderStatus
	AuditLogID     string
}

type PrototypePurchaseOrderStore struct {
	mu       sync.RWMutex
	records  map[string]purchasedomain.PurchaseOrder
	auditLog audit.LogStore
	txCount  int
}

type prototypePurchaseOrderTx struct {
	store     *PrototypePurchaseOrderStore
	auditLogs []audit.Log
}

func NewPurchaseOrderService(
	store PurchaseOrderStore,
	supplierRead PurchaseOrderSupplierReader,
	itemRead PurchaseOrderItemReader,
	warehouseRead PurchaseOrderWarehouseReader,
	uomConverter PurchaseOrderUOMConverter,
) PurchaseOrderService {
	return PurchaseOrderService{
		store:         store,
		supplierRead:  supplierRead,
		itemRead:      itemRead,
		warehouseRead: warehouseRead,
		uomConverter:  uomConverter,
		clock:         func() time.Time { return time.Now().UTC() },
	}
}

func NewPrototypePurchaseOrderStore(auditLog audit.LogStore) *PrototypePurchaseOrderStore {
	return &PrototypePurchaseOrderStore{
		records:  make(map[string]purchasedomain.PurchaseOrder),
		auditLog: auditLog,
	}
}

func (s PurchaseOrderService) ListPurchaseOrders(
	ctx context.Context,
	filter PurchaseOrderFilter,
) ([]purchasedomain.PurchaseOrder, error) {
	if s.store == nil {
		return nil, errors.New("purchase order store is required")
	}

	return s.store.List(ctx, filter)
}

func (s PurchaseOrderService) GetPurchaseOrder(
	ctx context.Context,
	id string,
) (purchasedomain.PurchaseOrder, error) {
	if s.store == nil {
		return purchasedomain.PurchaseOrder{}, errors.New("purchase order store is required")
	}
	order, err := s.store.Get(ctx, id)
	if err != nil {
		return purchasedomain.PurchaseOrder{}, MapPurchaseOrderError(err, map[string]any{"purchase_order_id": strings.TrimSpace(id)})
	}

	return order, nil
}

func (s PurchaseOrderService) CreatePurchaseOrder(
	ctx context.Context,
	input CreatePurchaseOrderInput,
) (PurchaseOrderResult, error) {
	if err := s.ensureReadyForWrite(); err != nil {
		return PurchaseOrderResult{}, err
	}
	if err := requirePurchaseOrderActor(input.ActorID); err != nil {
		return PurchaseOrderResult{}, err
	}

	now := s.now()
	orgID := strings.TrimSpace(input.OrgID)
	if orgID == "" {
		orgID = defaultPurchaseOrderOrgID
	}
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = newPurchaseOrderID(now)
	}
	poNo := strings.TrimSpace(input.PONo)
	if poNo == "" {
		poNo = newPurchaseOrderNo(now)
	}
	currencyCode := firstNonBlankPurchaseOrder(input.CurrencyCode, decimal.CurrencyVND.String())

	supplier, err := s.supplierRead.GetSupplier(ctx, input.SupplierID)
	if err != nil {
		return PurchaseOrderResult{}, mapPurchaseOrderMasterDataError(err, map[string]any{"supplier_id": strings.TrimSpace(input.SupplierID)})
	}
	warehouse, err := s.warehouseRead.GetWarehouse(ctx, input.WarehouseID)
	if err != nil {
		return PurchaseOrderResult{}, mapPurchaseOrderMasterDataError(err, map[string]any{"warehouse_id": strings.TrimSpace(input.WarehouseID)})
	}
	lines, err := s.newPurchaseOrderLineInputs(ctx, id, input.ExpectedDate, input.Lines, currencyCode)
	if err != nil {
		return PurchaseOrderResult{}, err
	}
	order, err := purchasedomain.NewPurchaseOrderDocument(purchasedomain.NewPurchaseOrderDocumentInput{
		ID:            id,
		OrgID:         orgID,
		PONo:          poNo,
		SupplierID:    supplier.ID,
		SupplierCode:  supplier.Code,
		SupplierName:  supplier.Name,
		WarehouseID:   warehouse.ID,
		WarehouseCode: warehouse.Code,
		ExpectedDate:  input.ExpectedDate,
		CurrencyCode:  currencyCode,
		Note:          input.Note,
		Lines:         lines,
		CreatedAt:     now,
		CreatedBy:     input.ActorID,
		UpdatedAt:     now,
	})
	if err != nil {
		return PurchaseOrderResult{}, MapPurchaseOrderError(err, nil)
	}

	var result PurchaseOrderResult
	err = s.store.WithinTx(ctx, func(txCtx context.Context, tx PurchaseOrderTx) error {
		if err := tx.Save(txCtx, order); err != nil {
			return err
		}
		log, err := newPurchaseOrderAuditLog(
			input.ActorID,
			input.RequestID,
			"purchase.order.created",
			order,
			nil,
			purchaseOrderAuditData(order),
			now,
		)
		if err != nil {
			return err
		}
		if err := tx.RecordAudit(txCtx, log); err != nil {
			return err
		}
		result = PurchaseOrderResult{PurchaseOrder: order, AuditLogID: log.ID}

		return nil
	})
	if err != nil {
		return PurchaseOrderResult{}, MapPurchaseOrderError(err, map[string]any{"purchase_order_id": id})
	}

	return result, nil
}

func (s PurchaseOrderService) UpdatePurchaseOrder(
	ctx context.Context,
	input UpdatePurchaseOrderInput,
) (PurchaseOrderResult, error) {
	if err := s.ensureReadyForWrite(); err != nil {
		return PurchaseOrderResult{}, err
	}
	if err := requirePurchaseOrderActor(input.ActorID); err != nil {
		return PurchaseOrderResult{}, err
	}

	var result PurchaseOrderResult
	err := s.store.WithinTx(ctx, func(txCtx context.Context, tx PurchaseOrderTx) error {
		current, err := tx.GetForUpdate(txCtx, input.ID)
		if err != nil {
			return MapPurchaseOrderError(err, map[string]any{"purchase_order_id": strings.TrimSpace(input.ID)})
		}
		if err := ensurePurchaseOrderExpectedVersion(current, input.ExpectedVersion); err != nil {
			return err
		}

		supplierID := firstNonBlankPurchaseOrder(input.SupplierID, current.SupplierID)
		supplier, err := s.supplierRead.GetSupplier(txCtx, supplierID)
		if err != nil {
			return mapPurchaseOrderMasterDataError(err, map[string]any{"supplier_id": supplierID})
		}
		warehouseID := firstNonBlankPurchaseOrder(input.WarehouseID, current.WarehouseID)
		warehouse, err := s.warehouseRead.GetWarehouse(txCtx, warehouseID)
		if err != nil {
			return mapPurchaseOrderMasterDataError(err, map[string]any{"warehouse_id": warehouseID})
		}

		expectedDate := firstNonBlankPurchaseOrder(input.ExpectedDate, current.ExpectedDate)
		var lines []purchasedomain.NewPurchaseOrderLineInput
		if input.Lines != nil {
			lines, err = s.newPurchaseOrderLineInputs(txCtx, current.ID, expectedDate, input.Lines, current.CurrencyCode.String())
			if err != nil {
				return err
			}
		}
		updated, err := current.ReplaceDraftDetails(purchasedomain.UpdatePurchaseOrderDocumentInput{
			SupplierID:    supplier.ID,
			SupplierCode:  supplier.Code,
			SupplierName:  supplier.Name,
			WarehouseID:   warehouse.ID,
			WarehouseCode: warehouse.Code,
			ExpectedDate:  expectedDate,
			Note:          firstNonBlankPurchaseOrder(input.Note, current.Note),
			Lines:         lines,
			UpdatedAt:     s.now(),
			UpdatedBy:     input.ActorID,
		})
		if err != nil {
			return MapPurchaseOrderError(err, map[string]any{"purchase_order_id": current.ID})
		}
		if err := tx.Save(txCtx, updated); err != nil {
			return err
		}
		log, err := newPurchaseOrderAuditLog(
			input.ActorID,
			input.RequestID,
			"purchase.order.updated",
			updated,
			purchaseOrderAuditData(current),
			purchaseOrderAuditData(updated),
			updated.UpdatedAt,
		)
		if err != nil {
			return err
		}
		if err := tx.RecordAudit(txCtx, log); err != nil {
			return err
		}
		result = PurchaseOrderResult{PurchaseOrder: updated, AuditLogID: log.ID}

		return nil
	})
	if err != nil {
		return PurchaseOrderResult{}, err
	}

	return result, nil
}

func (s PurchaseOrderService) SubmitPurchaseOrder(
	ctx context.Context,
	input PurchaseOrderActionInput,
) (PurchaseOrderActionResult, error) {
	return s.transition(ctx, input, "purchase.order.submitted", func(
		order purchasedomain.PurchaseOrder,
		actorID string,
		changedAt time.Time,
	) (purchasedomain.PurchaseOrder, error) {
		return order.Submit(actorID, changedAt)
	})
}

func (s PurchaseOrderService) ApprovePurchaseOrder(
	ctx context.Context,
	input PurchaseOrderActionInput,
) (PurchaseOrderActionResult, error) {
	return s.transition(ctx, input, "purchase.order.approved", func(
		order purchasedomain.PurchaseOrder,
		actorID string,
		changedAt time.Time,
	) (purchasedomain.PurchaseOrder, error) {
		return order.Approve(actorID, changedAt)
	})
}

func (s PurchaseOrderService) CancelPurchaseOrder(
	ctx context.Context,
	input PurchaseOrderActionInput,
) (PurchaseOrderActionResult, error) {
	if strings.TrimSpace(input.Reason) == "" {
		return PurchaseOrderActionResult{}, purchaseOrderValidationError(
			purchasedomain.ErrPurchaseOrderRequiredField,
			map[string]any{"field": "reason"},
		)
	}

	return s.transition(ctx, input, "purchase.order.cancelled", func(
		order purchasedomain.PurchaseOrder,
		actorID string,
		changedAt time.Time,
	) (purchasedomain.PurchaseOrder, error) {
		return order.CancelWithReason(actorID, input.Reason, changedAt)
	})
}

func (s PurchaseOrderService) ClosePurchaseOrder(
	ctx context.Context,
	input PurchaseOrderActionInput,
) (PurchaseOrderActionResult, error) {
	return s.transition(ctx, input, "purchase.order.closed", func(
		order purchasedomain.PurchaseOrder,
		actorID string,
		changedAt time.Time,
	) (purchasedomain.PurchaseOrder, error) {
		return order.Close(actorID, changedAt)
	})
}

func (s PurchaseOrderService) transition(
	ctx context.Context,
	input PurchaseOrderActionInput,
	action string,
	transition func(purchasedomain.PurchaseOrder, string, time.Time) (purchasedomain.PurchaseOrder, error),
) (PurchaseOrderActionResult, error) {
	if err := s.ensureReadyForWrite(); err != nil {
		return PurchaseOrderActionResult{}, err
	}
	if err := requirePurchaseOrderActor(input.ActorID); err != nil {
		return PurchaseOrderActionResult{}, err
	}

	var result PurchaseOrderActionResult
	err := s.store.WithinTx(ctx, func(txCtx context.Context, tx PurchaseOrderTx) error {
		current, err := tx.GetForUpdate(txCtx, input.ID)
		if err != nil {
			return MapPurchaseOrderError(err, map[string]any{"purchase_order_id": strings.TrimSpace(input.ID)})
		}
		if err := ensurePurchaseOrderExpectedVersion(current, input.ExpectedVersion); err != nil {
			return err
		}

		now := s.now()
		updated, err := transition(current, input.ActorID, now)
		if err != nil {
			return MapPurchaseOrderError(err, map[string]any{
				"purchase_order_id": current.ID,
				"status":            string(current.Status),
			})
		}
		if strings.TrimSpace(input.Note) != "" {
			updated.Note = strings.TrimSpace(input.Note)
		}
		if err := tx.Save(txCtx, updated); err != nil {
			return err
		}
		log, err := newPurchaseOrderAuditLog(
			input.ActorID,
			input.RequestID,
			action,
			updated,
			purchaseOrderAuditData(current),
			purchaseOrderAuditData(updated),
			now,
		)
		if err != nil {
			return err
		}
		if err := tx.RecordAudit(txCtx, log); err != nil {
			return err
		}
		result = PurchaseOrderActionResult{
			PurchaseOrder:  updated,
			PreviousStatus: current.Status,
			CurrentStatus:  updated.Status,
			AuditLogID:     log.ID,
		}

		return nil
	})
	if err != nil {
		return PurchaseOrderActionResult{}, err
	}

	return result, nil
}

func (s PurchaseOrderService) newPurchaseOrderLineInputs(
	ctx context.Context,
	orderID string,
	expectedDate string,
	inputs []PurchaseOrderLineInput,
	currencyCode string,
) ([]purchasedomain.NewPurchaseOrderLineInput, error) {
	if len(inputs) == 0 {
		return nil, purchaseOrderValidationError(purchasedomain.ErrPurchaseOrderRequiredField, map[string]any{"field": "lines"})
	}

	lines := make([]purchasedomain.NewPurchaseOrderLineInput, 0, len(inputs))
	for index, input := range inputs {
		item, err := s.itemRead.Get(ctx, input.ItemID)
		if err != nil {
			return nil, mapPurchaseOrderMasterDataError(err, map[string]any{"item_id": strings.TrimSpace(input.ItemID)})
		}
		lineNo := input.LineNo
		if lineNo == 0 {
			lineNo = index + 1
		}
		lineID := strings.TrimSpace(input.ID)
		if lineID == "" {
			lineID = fmt.Sprintf("%s-line-%02d", strings.TrimSpace(orderID), lineNo)
		}
		orderedQty, err := decimal.ParseQuantity(input.OrderedQty)
		if err != nil || orderedQty.IsNegative() || orderedQty.IsZero() {
			return nil, purchaseOrderValidationError(purchasedomain.ErrPurchaseOrderInvalidQuantity, map[string]any{"line_no": lineNo, "field": "ordered_qty"})
		}
		uomCode := firstNonBlankPurchaseOrder(input.UOMCode, item.UOMPurchase, item.UOMBase)
		conversion, err := s.convertPurchaseOrderLineToBase(ctx, item, orderedQty, uomCode, lineNo)
		if err != nil {
			return nil, err
		}
		unitPrice, err := decimal.ParseUnitPrice(input.UnitPrice)
		if err != nil || unitPrice.IsNegative() {
			return nil, purchaseOrderValidationError(purchasedomain.ErrPurchaseOrderInvalidAmount, map[string]any{"line_no": lineNo, "field": "unit_price"})
		}
		lineCurrency := firstNonBlankPurchaseOrder(input.CurrencyCode, currencyCode)
		if !strings.EqualFold(strings.TrimSpace(lineCurrency), decimal.CurrencyVND.String()) {
			return nil, purchaseOrderValidationError(purchasedomain.ErrPurchaseOrderInvalidCurrency, map[string]any{"line_no": lineNo, "field": "currency_code"})
		}
		lines = append(lines, purchasedomain.NewPurchaseOrderLineInput{
			ID:               lineID,
			LineNo:           lineNo,
			ItemID:           item.ID,
			SKUCode:          item.SKUCode,
			ItemName:         item.Name,
			OrderedQty:       orderedQty,
			ReceivedQty:      decimal.MustQuantity("0"),
			UOMCode:          conversion.SourceUOMCode.String(),
			BaseOrderedQty:   conversion.BaseQuantity,
			BaseReceivedQty:  decimal.MustQuantity("0"),
			BaseUOMCode:      conversion.BaseUOMCode.String(),
			ConversionFactor: conversion.ConversionFactor,
			UnitPrice:        unitPrice,
			CurrencyCode:     lineCurrency,
			ExpectedDate:     firstNonBlankPurchaseOrder(input.ExpectedDate, expectedDate),
			Note:             input.Note,
		})
	}

	return lines, nil
}

func (s PurchaseOrderService) convertPurchaseOrderLineToBase(
	ctx context.Context,
	item masterdatadomain.Item,
	orderedQty decimal.Decimal,
	uomCode string,
	lineNo int,
) (ConvertPurchaseOrderLineToBaseResult, error) {
	if strings.EqualFold(strings.TrimSpace(uomCode), strings.TrimSpace(item.UOMBase)) {
		sourceUOM, err := decimal.NormalizeUOMCode(uomCode)
		if err != nil {
			return ConvertPurchaseOrderLineToBaseResult{}, purchaseOrderValidationError(purchasedomain.ErrPurchaseOrderInvalidQuantity, map[string]any{"line_no": lineNo, "field": "uom_code"})
		}
		baseUOM, err := decimal.NormalizeUOMCode(item.UOMBase)
		if err != nil {
			return ConvertPurchaseOrderLineToBaseResult{}, purchaseOrderValidationError(purchasedomain.ErrPurchaseOrderInvalidQuantity, map[string]any{"line_no": lineNo, "field": "base_uom_code"})
		}

		return ConvertPurchaseOrderLineToBaseResult{
			Quantity:         orderedQty,
			SourceUOMCode:    sourceUOM,
			BaseQuantity:     orderedQty,
			BaseUOMCode:      baseUOM,
			ConversionFactor: decimal.MustQuantity("1"),
		}, nil
	}

	conversion, err := s.uomConverter.ConvertToBase(ctx, ConvertPurchaseOrderLineToBaseInput{
		ItemID:      item.ID,
		SKU:         item.SKUCode,
		Quantity:    orderedQty,
		FromUOMCode: uomCode,
		BaseUOMCode: item.UOMBase,
	})
	if err != nil {
		return ConvertPurchaseOrderLineToBaseResult{}, mapPurchaseOrderMasterDataError(err, map[string]any{
			"line_no":       lineNo,
			"item_id":       item.ID,
			"sku_code":      item.SKUCode,
			"from_uom_code": strings.ToUpper(strings.TrimSpace(uomCode)),
			"base_uom_code": strings.ToUpper(strings.TrimSpace(item.UOMBase)),
		})
	}

	return conversion, nil
}

func (s PurchaseOrderService) ensureReadyForWrite() error {
	if s.store == nil {
		return errors.New("purchase order store is required")
	}
	if s.supplierRead == nil {
		return errors.New("purchase order supplier reader is required")
	}
	if s.itemRead == nil {
		return errors.New("purchase order item reader is required")
	}
	if s.warehouseRead == nil {
		return errors.New("purchase order warehouse reader is required")
	}
	if s.uomConverter == nil {
		return errors.New("purchase order uom converter is required")
	}

	return nil
}

func (s PurchaseOrderService) now() time.Time {
	if s.clock == nil {
		return time.Now().UTC()
	}

	return s.clock().UTC()
}

func (s *PrototypePurchaseOrderStore) List(
	_ context.Context,
	filter PurchaseOrderFilter,
) ([]purchasedomain.PurchaseOrder, error) {
	if s == nil {
		return nil, errors.New("purchase order store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows := make([]purchasedomain.PurchaseOrder, 0, len(s.records))
	for _, order := range s.records {
		if purchaseOrderMatchesFilter(order, filter) {
			rows = append(rows, order.Clone())
		}
	}
	sortPurchaseOrders(rows)

	return rows, nil
}

func (s *PrototypePurchaseOrderStore) Get(
	_ context.Context,
	id string,
) (purchasedomain.PurchaseOrder, error) {
	if s == nil {
		return purchasedomain.PurchaseOrder{}, errors.New("purchase order store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	order, ok := s.records[strings.TrimSpace(id)]
	if !ok {
		return purchasedomain.PurchaseOrder{}, ErrPurchaseOrderNotFound
	}

	return order.Clone(), nil
}

func (s *PrototypePurchaseOrderStore) WithinTx(
	ctx context.Context,
	fn func(context.Context, PurchaseOrderTx) error,
) error {
	if s == nil {
		return errors.New("purchase order store is required")
	}
	if fn == nil {
		return errors.New("purchase order transaction function is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot := clonePurchaseOrderMap(s.records)
	tx := &prototypePurchaseOrderTx{store: s}
	if err := fn(ctx, tx); err != nil {
		s.records = snapshot
		return err
	}
	if s.auditLog == nil {
		s.records = snapshot
		return errors.New("audit log store is required")
	}
	for _, log := range tx.auditLogs {
		if err := s.auditLog.Record(ctx, log); err != nil {
			s.records = snapshot
			return err
		}
	}
	s.txCount++

	return nil
}

func (s *PrototypePurchaseOrderStore) TransactionCount() int {
	if s == nil {
		return 0
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.txCount
}

func (tx *prototypePurchaseOrderTx) GetForUpdate(
	_ context.Context,
	id string,
) (purchasedomain.PurchaseOrder, error) {
	order, ok := tx.store.records[strings.TrimSpace(id)]
	if !ok {
		return purchasedomain.PurchaseOrder{}, ErrPurchaseOrderNotFound
	}

	return order.Clone(), nil
}

func (tx *prototypePurchaseOrderTx) Save(
	_ context.Context,
	order purchasedomain.PurchaseOrder,
) error {
	if strings.TrimSpace(order.ID) == "" {
		return purchasedomain.ErrPurchaseOrderRequiredField
	}
	tx.store.records[order.ID] = order.Clone()

	return nil
}

func (tx *prototypePurchaseOrderTx) RecordAudit(
	_ context.Context,
	log audit.Log,
) error {
	tx.auditLogs = append(tx.auditLogs, log)

	return nil
}

func ensurePurchaseOrderExpectedVersion(order purchasedomain.PurchaseOrder, expectedVersion int) error {
	if expectedVersion <= 0 || order.Version == expectedVersion {
		return nil
	}

	return apperrors.Conflict(
		ErrorCodePurchaseOrderVersionConflict,
		"Purchase order version changed",
		ErrPurchaseOrderVersionConflict,
		map[string]any{
			"purchase_order_id": order.ID,
			"expected_version":  expectedVersion,
			"current_version":   order.Version,
		},
	)
}

func requirePurchaseOrderActor(actorID string) error {
	if strings.TrimSpace(actorID) == "" {
		return purchaseOrderValidationError(purchasedomain.ErrPurchaseOrderTransitionActorRequired, map[string]any{"field": "actor_id"})
	}

	return nil
}

func purchaseOrderValidationError(cause error, details map[string]any) error {
	return apperrors.BadRequest(ErrorCodePurchaseOrderValidation, "Purchase order request is invalid", cause, details)
}

func MapPurchaseOrderError(err error, details map[string]any) error {
	if err == nil {
		return nil
	}
	if _, ok := apperrors.As(err); ok {
		return err
	}
	if errors.Is(err, ErrPurchaseOrderNotFound) {
		return apperrors.NotFound(ErrorCodePurchaseOrderNotFound, "Purchase order not found", err, details)
	}
	if errors.Is(err, purchasedomain.ErrPurchaseOrderInvalidTransition) ||
		errors.Is(err, purchasedomain.ErrPurchaseOrderInvalidStatus) {
		return apperrors.Conflict(ErrorCodePurchaseOrderInvalidState, "Purchase order state is invalid", err, details)
	}
	if errors.Is(err, purchasedomain.ErrPurchaseOrderRequiredField) ||
		errors.Is(err, purchasedomain.ErrPurchaseOrderTransitionActorRequired) ||
		errors.Is(err, purchasedomain.ErrPurchaseOrderInvalidCurrency) ||
		errors.Is(err, purchasedomain.ErrPurchaseOrderInvalidQuantity) ||
		errors.Is(err, purchasedomain.ErrPurchaseOrderInvalidAmount) {
		return purchaseOrderValidationError(err, details)
	}

	return err
}

func mapPurchaseOrderMasterDataError(err error, details map[string]any) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, masterdatadomain.ErrUOMConversionMissing) ||
		errors.Is(err, masterdatadomain.ErrUOMConversionInactive) ||
		errors.Is(err, masterdatadomain.ErrUOMInvalid) {
		return apperrors.Unprocessable(response.ErrorCodeUOMConversionNotFound, "UOM conversion is not available", err, details)
	}

	return apperrors.NotFound(response.ErrorCodeNotFound, "Referenced master data was not found", err, details)
}

func newPurchaseOrderAuditLog(
	actorID string,
	requestID string,
	action string,
	order purchasedomain.PurchaseOrder,
	beforeData map[string]any,
	afterData map[string]any,
	createdAt time.Time,
) (audit.Log, error) {
	return audit.NewLog(audit.NewLogInput{
		OrgID:      firstNonBlankPurchaseOrder(order.OrgID, defaultPurchaseOrderOrgID),
		ActorID:    strings.TrimSpace(actorID),
		Action:     strings.TrimSpace(action),
		EntityType: purchaseOrderEntityType,
		EntityID:   order.ID,
		RequestID:  strings.TrimSpace(requestID),
		BeforeData: beforeData,
		AfterData:  afterData,
		Metadata: map[string]any{
			"source":        "purchase order application service",
			"po_no":         order.PONo,
			"supplier_code": order.SupplierCode,
		},
		CreatedAt: createdAt,
	})
}

func purchaseOrderAuditData(order purchasedomain.PurchaseOrder) map[string]any {
	data := map[string]any{
		"po_no":          order.PONo,
		"supplier_id":    order.SupplierID,
		"supplier_code":  order.SupplierCode,
		"warehouse_id":   order.WarehouseID,
		"warehouse_code": order.WarehouseCode,
		"expected_date":  order.ExpectedDate,
		"status":         string(order.Status),
		"currency_code":  order.CurrencyCode.String(),
		"subtotal":       order.SubtotalAmount.String(),
		"total":          order.TotalAmount.String(),
		"line_count":     len(order.Lines),
		"version":        order.Version,
	}
	if strings.TrimSpace(order.CancelReason) != "" {
		data["cancel_reason"] = order.CancelReason
	}
	if strings.TrimSpace(order.RejectReason) != "" {
		data["reject_reason"] = order.RejectReason
	}

	return data
}

func purchaseOrderMatchesFilter(order purchasedomain.PurchaseOrder, filter PurchaseOrderFilter) bool {
	search := strings.ToLower(strings.TrimSpace(filter.Search))
	if search != "" {
		haystack := strings.ToLower(strings.Join([]string{
			order.PONo,
			order.SupplierCode,
			order.SupplierName,
			order.WarehouseCode,
		}, " "))
		if !strings.Contains(haystack, search) {
			return false
		}
	}
	if len(filter.Statuses) > 0 {
		matched := false
		for _, status := range filter.Statuses {
			if order.Status == purchasedomain.NormalizePurchaseOrderStatus(status) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if strings.TrimSpace(filter.SupplierID) != "" && order.SupplierID != strings.TrimSpace(filter.SupplierID) {
		return false
	}
	if strings.TrimSpace(filter.WarehouseID) != "" && order.WarehouseID != strings.TrimSpace(filter.WarehouseID) {
		return false
	}
	if strings.TrimSpace(filter.ExpectedFrom) != "" && order.ExpectedDate < strings.TrimSpace(filter.ExpectedFrom) {
		return false
	}
	if strings.TrimSpace(filter.ExpectedTo) != "" && order.ExpectedDate > strings.TrimSpace(filter.ExpectedTo) {
		return false
	}

	return true
}

func sortPurchaseOrders(orders []purchasedomain.PurchaseOrder) {
	sort.SliceStable(orders, func(i, j int) bool {
		if orders[i].ExpectedDate == orders[j].ExpectedDate {
			return orders[i].PONo < orders[j].PONo
		}

		return orders[i].ExpectedDate > orders[j].ExpectedDate
	})
}

func clonePurchaseOrderMap(records map[string]purchasedomain.PurchaseOrder) map[string]purchasedomain.PurchaseOrder {
	clone := make(map[string]purchasedomain.PurchaseOrder, len(records))
	for id, order := range records {
		clone[id] = order.Clone()
	}

	return clone
}

func newPurchaseOrderID(now time.Time) string {
	return fmt.Sprintf("po-%s-%06d", now.UTC().Format("060102"), now.UTC().UnixNano()%1000000)
}

func newPurchaseOrderNo(now time.Time) string {
	return fmt.Sprintf("PO-%s-%06d", now.UTC().Format("060102"), now.UTC().UnixNano()%1000000)
}

func firstNonBlankPurchaseOrder(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}

	return ""
}
