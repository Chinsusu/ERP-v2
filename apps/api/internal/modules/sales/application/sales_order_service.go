package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	salesdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
	apperrors "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/errors"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

var ErrSalesOrderNotFound = errors.New("sales order not found")
var ErrSalesOrderVersionConflict = errors.New("sales order version conflict")

const (
	ErrorCodeSalesOrderNotFound        response.ErrorCode = "SALES_ORDER_NOT_FOUND"
	ErrorCodeSalesOrderValidation      response.ErrorCode = "SALES_ORDER_VALIDATION_ERROR"
	ErrorCodeSalesOrderInvalidState    response.ErrorCode = "SALES_ORDER_INVALID_STATE"
	ErrorCodeSalesOrderVersionConflict response.ErrorCode = "SALES_ORDER_VERSION_CONFLICT"

	defaultSalesOrderOrgID = "org-my-pham"
	salesOrderEntityType   = "sales.sales_order"
)

type SalesOrderStore interface {
	List(ctx context.Context, filter SalesOrderFilter) ([]salesdomain.SalesOrder, error)
	Get(ctx context.Context, id string) (salesdomain.SalesOrder, error)
	WithinTx(ctx context.Context, fn func(context.Context, SalesOrderTx) error) error
}

type SalesOrderTx interface {
	GetForUpdate(ctx context.Context, id string) (salesdomain.SalesOrder, error)
	Save(ctx context.Context, order salesdomain.SalesOrder) error
	RecordAudit(ctx context.Context, log audit.Log) error
}

type SalesOrderCustomerReader interface {
	GetCustomer(ctx context.Context, id string) (domain.Customer, error)
}

type SalesOrderItemReader interface {
	Get(ctx context.Context, id string) (domain.Item, error)
}

type SalesOrderWarehouseReader interface {
	GetWarehouse(ctx context.Context, id string) (domain.Warehouse, error)
}

type SalesOrderStockReserver interface {
	ReserveSalesOrder(ctx context.Context, input SalesOrderStockReservationInput) (SalesOrderStockReservationResult, error)
	ReleaseSalesOrder(ctx context.Context, input SalesOrderStockReleaseInput) (SalesOrderStockReleaseResult, error)
}

type SalesOrderService struct {
	store         SalesOrderStore
	customerRead  SalesOrderCustomerReader
	itemRead      SalesOrderItemReader
	warehouseRead SalesOrderWarehouseReader
	stockReserver SalesOrderStockReserver
	clock         func() time.Time
}

type SalesOrderFilter struct {
	Search      string
	Statuses    []salesdomain.SalesOrderStatus
	CustomerID  string
	Channel     string
	WarehouseID string
	DateFrom    string
	DateTo      string
}

type CreateSalesOrderInput struct {
	ID           string
	OrgID        string
	OrderNo      string
	CustomerID   string
	Channel      string
	WarehouseID  string
	OrderDate    string
	CurrencyCode string
	Note         string
	Lines        []SalesOrderLineInput
	ActorID      string
	RequestID    string
}

type UpdateSalesOrderInput struct {
	ID              string
	CustomerID      string
	Channel         string
	WarehouseID     string
	OrderDate       string
	Note            string
	Lines           []SalesOrderLineInput
	ExpectedVersion int
	ActorID         string
	RequestID       string
}

type SalesOrderLineInput struct {
	ID                 string
	LineNo             int
	ItemID             string
	OrderedQty         string
	UOMCode            string
	UnitPrice          string
	CurrencyCode       string
	LineDiscountAmount string
	BatchID            string
	BatchNo            string
}

type SalesOrderActionInput struct {
	ID              string
	ExpectedVersion int
	Reason          string
	Note            string
	ActorID         string
	RequestID       string
}

type SalesOrderResult struct {
	SalesOrder salesdomain.SalesOrder
	AuditLogID string
}

type SalesOrderActionResult struct {
	SalesOrder     salesdomain.SalesOrder
	PreviousStatus salesdomain.SalesOrderStatus
	CurrentStatus  salesdomain.SalesOrderStatus
	AuditLogID     string
}

type SalesOrderStockReservationInput struct {
	OrgID         string
	SalesOrderID  string
	OrderNo       string
	WarehouseID   string
	WarehouseCode string
	ActorID       string
	Reason        string
	RequestID     string
	ReservedAt    time.Time
	Lines         []SalesOrderStockReservationLineInput
}

type SalesOrderStockReservationLineInput struct {
	SalesOrderLineID string
	LineNo           int
	ItemID           string
	SKUCode          string
	OrderedQty       decimal.Decimal
	BaseOrderedQty   decimal.Decimal
	BaseUOMCode      decimal.UOMCode
}

type SalesOrderStockReservationResult struct {
	Lines []SalesOrderReservedLine
}

type SalesOrderReservedLine struct {
	SalesOrderLineID string
	ReservedQty      decimal.Decimal
	BatchID          string
	BatchNo          string
	BinID            string
	BinCode          string
}

type SalesOrderStockReleaseInput struct {
	OrgID        string
	SalesOrderID string
	OrderNo      string
	ActorID      string
	Reason       string
	RequestID    string
	ReleasedAt   time.Time
}

type SalesOrderStockReleaseResult struct {
	ReleasedReservationCount int
}

type PrototypeSalesOrderStore struct {
	mu       sync.RWMutex
	records  map[string]salesdomain.SalesOrder
	auditLog audit.LogStore
	txCount  int
}

type prototypeSalesOrderTx struct {
	store     *PrototypeSalesOrderStore
	auditLogs []audit.Log
}

func NewSalesOrderService(
	store SalesOrderStore,
	customerRead SalesOrderCustomerReader,
	itemRead SalesOrderItemReader,
	warehouseRead SalesOrderWarehouseReader,
) SalesOrderService {
	return SalesOrderService{
		store:         store,
		customerRead:  customerRead,
		itemRead:      itemRead,
		warehouseRead: warehouseRead,
		clock:         func() time.Time { return time.Now().UTC() },
	}
}

func (s SalesOrderService) WithStockReserver(reserver SalesOrderStockReserver) SalesOrderService {
	s.stockReserver = reserver
	return s
}

func NewPrototypeSalesOrderStore(auditLog audit.LogStore) *PrototypeSalesOrderStore {
	store := &PrototypeSalesOrderStore{
		records:  make(map[string]salesdomain.SalesOrder),
		auditLog: auditLog,
	}
	for _, order := range prototypeSalesOrders() {
		store.records[order.ID] = order.Clone()
	}

	return store
}

func (s SalesOrderService) ListSalesOrders(
	ctx context.Context,
	filter SalesOrderFilter,
) ([]salesdomain.SalesOrder, error) {
	if s.store == nil {
		return nil, errors.New("sales order store is required")
	}

	return s.store.List(ctx, filter)
}

func (s SalesOrderService) GetSalesOrder(
	ctx context.Context,
	id string,
) (salesdomain.SalesOrder, error) {
	if s.store == nil {
		return salesdomain.SalesOrder{}, errors.New("sales order store is required")
	}
	order, err := s.store.Get(ctx, id)
	if err != nil {
		return salesdomain.SalesOrder{}, mapSalesOrderError(err, map[string]any{"sales_order_id": strings.TrimSpace(id)})
	}

	return order, nil
}

func (s SalesOrderService) CreateSalesOrder(
	ctx context.Context,
	input CreateSalesOrderInput,
) (SalesOrderResult, error) {
	if err := s.ensureReadyForWrite(); err != nil {
		return SalesOrderResult{}, err
	}
	if err := requireActor(input.ActorID); err != nil {
		return SalesOrderResult{}, err
	}

	now := s.now()
	orgID := strings.TrimSpace(input.OrgID)
	if orgID == "" {
		orgID = defaultSalesOrderOrgID
	}
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = newSalesOrderID(now)
	}
	orderNo := strings.TrimSpace(input.OrderNo)
	if orderNo == "" {
		orderNo = newSalesOrderNo(now)
	}
	currencyCode := strings.TrimSpace(input.CurrencyCode)
	if currencyCode == "" {
		currencyCode = decimal.CurrencyVND.String()
	}

	customer, err := s.customerRead.GetCustomer(ctx, input.CustomerID)
	if err != nil {
		return SalesOrderResult{}, mapMasterDataReadError(err, map[string]any{"customer_id": strings.TrimSpace(input.CustomerID)})
	}
	warehouse, err := s.getOptionalWarehouse(ctx, input.WarehouseID)
	if err != nil {
		return SalesOrderResult{}, err
	}
	lines, err := s.newSalesOrderLineInputs(ctx, id, input.Lines, currencyCode)
	if err != nil {
		return SalesOrderResult{}, err
	}
	order, err := salesdomain.NewSalesOrderDocument(salesdomain.NewSalesOrderDocumentInput{
		ID:            id,
		OrgID:         orgID,
		OrderNo:       orderNo,
		CustomerID:    customer.ID,
		CustomerCode:  customer.Code,
		CustomerName:  customer.Name,
		Channel:       input.Channel,
		WarehouseID:   warehouse.ID,
		WarehouseCode: warehouse.Code,
		OrderDate:     input.OrderDate,
		CurrencyCode:  currencyCode,
		Note:          input.Note,
		Lines:         lines,
		CreatedAt:     now,
		CreatedBy:     input.ActorID,
		UpdatedAt:     now,
	})
	if err != nil {
		return SalesOrderResult{}, mapSalesOrderError(err, nil)
	}

	var result SalesOrderResult
	err = s.store.WithinTx(ctx, func(txCtx context.Context, tx SalesOrderTx) error {
		if err := tx.Save(txCtx, order); err != nil {
			return err
		}
		log, err := newSalesOrderAuditLog(
			input.ActorID,
			input.RequestID,
			"sales.order.created",
			order,
			nil,
			salesOrderAuditData(order),
			now,
		)
		if err != nil {
			return err
		}
		if err := tx.RecordAudit(txCtx, log); err != nil {
			return err
		}
		result = SalesOrderResult{SalesOrder: order, AuditLogID: log.ID}

		return nil
	})
	if err != nil {
		return SalesOrderResult{}, mapSalesOrderError(err, map[string]any{"sales_order_id": id})
	}

	return result, nil
}

func (s SalesOrderService) UpdateSalesOrder(
	ctx context.Context,
	input UpdateSalesOrderInput,
) (SalesOrderResult, error) {
	if err := s.ensureReadyForWrite(); err != nil {
		return SalesOrderResult{}, err
	}
	if err := requireActor(input.ActorID); err != nil {
		return SalesOrderResult{}, err
	}

	var result SalesOrderResult
	err := s.store.WithinTx(ctx, func(txCtx context.Context, tx SalesOrderTx) error {
		current, err := tx.GetForUpdate(txCtx, input.ID)
		if err != nil {
			return mapSalesOrderError(err, map[string]any{"sales_order_id": strings.TrimSpace(input.ID)})
		}
		if err := ensureExpectedVersion(current, input.ExpectedVersion); err != nil {
			return err
		}

		customerID := firstNonBlank(input.CustomerID, current.CustomerID)
		customer, err := s.customerRead.GetCustomer(txCtx, customerID)
		if err != nil {
			return mapMasterDataReadError(err, map[string]any{"customer_id": customerID})
		}
		warehouseID := firstNonBlank(input.WarehouseID, current.WarehouseID)
		warehouse, err := s.getOptionalWarehouse(txCtx, warehouseID)
		if err != nil {
			return err
		}

		var lines []salesdomain.NewSalesOrderLineInput
		if input.Lines != nil {
			lines, err = s.newSalesOrderLineInputs(txCtx, current.ID, input.Lines, current.CurrencyCode.String())
			if err != nil {
				return err
			}
		}
		updated, err := current.ReplaceDraftDetails(salesdomain.UpdateSalesOrderDocumentInput{
			CustomerID:    customer.ID,
			CustomerCode:  customer.Code,
			CustomerName:  customer.Name,
			Channel:       firstNonBlank(input.Channel, current.Channel),
			WarehouseID:   warehouse.ID,
			WarehouseCode: warehouse.Code,
			OrderDate:     firstNonBlank(input.OrderDate, current.OrderDate),
			Note:          firstNonBlank(input.Note, current.Note),
			Lines:         lines,
			UpdatedAt:     s.now(),
			UpdatedBy:     input.ActorID,
		})
		if err != nil {
			return mapSalesOrderError(err, map[string]any{"sales_order_id": current.ID})
		}
		if err := tx.Save(txCtx, updated); err != nil {
			return err
		}
		log, err := newSalesOrderAuditLog(
			input.ActorID,
			input.RequestID,
			"sales.order.updated",
			updated,
			salesOrderAuditData(current),
			salesOrderAuditData(updated),
			updated.UpdatedAt,
		)
		if err != nil {
			return err
		}
		if err := tx.RecordAudit(txCtx, log); err != nil {
			return err
		}
		result = SalesOrderResult{SalesOrder: updated, AuditLogID: log.ID}

		return nil
	})
	if err != nil {
		return SalesOrderResult{}, err
	}

	return result, nil
}

func (s SalesOrderService) ConfirmSalesOrder(
	ctx context.Context,
	input SalesOrderActionInput,
) (SalesOrderActionResult, error) {
	if s.stockReserver != nil {
		return s.confirmAndReserveSalesOrder(ctx, input)
	}

	return s.transition(ctx, input, "sales.order.confirmed", func(
		order salesdomain.SalesOrder,
		actorID string,
		changedAt time.Time,
	) (salesdomain.SalesOrder, error) {
		return order.Confirm(actorID, changedAt)
	})
}

func (s SalesOrderService) confirmAndReserveSalesOrder(
	ctx context.Context,
	input SalesOrderActionInput,
) (SalesOrderActionResult, error) {
	if err := s.ensureReadyForWrite(); err != nil {
		return SalesOrderActionResult{}, err
	}
	if err := requireActor(input.ActorID); err != nil {
		return SalesOrderActionResult{}, err
	}

	var result SalesOrderActionResult
	err := s.store.WithinTx(ctx, func(txCtx context.Context, tx SalesOrderTx) error {
		current, err := tx.GetForUpdate(txCtx, input.ID)
		if err != nil {
			return mapSalesOrderError(err, map[string]any{"sales_order_id": strings.TrimSpace(input.ID)})
		}
		if err := ensureExpectedVersion(current, input.ExpectedVersion); err != nil {
			return err
		}

		now := s.now()
		confirmed, err := current.Confirm(input.ActorID, now)
		if err != nil {
			return mapSalesOrderError(err, map[string]any{
				"sales_order_id": current.ID,
				"status":         string(current.Status),
			})
		}

		reservationResult, err := s.stockReserver.ReserveSalesOrder(
			txCtx,
			newSalesOrderStockReservationInput(confirmed, input, now),
		)
		if err != nil {
			return mapSalesOrderError(err, map[string]any{"sales_order_id": current.ID})
		}
		reserved, err := applySalesOrderReservations(confirmed, reservationResult, input.ActorID, now)
		if err != nil {
			return err
		}
		if strings.TrimSpace(input.Note) != "" {
			reserved.Note = strings.TrimSpace(input.Note)
		}
		if err := tx.Save(txCtx, reserved); err != nil {
			return err
		}
		log, err := newSalesOrderAuditLog(
			input.ActorID,
			input.RequestID,
			"sales.order.reserved",
			reserved,
			salesOrderAuditData(current),
			salesOrderAuditData(reserved),
			now,
		)
		if err != nil {
			return err
		}
		if err := tx.RecordAudit(txCtx, log); err != nil {
			return err
		}
		result = SalesOrderActionResult{
			SalesOrder:     reserved,
			PreviousStatus: current.Status,
			CurrentStatus:  reserved.Status,
			AuditLogID:     log.ID,
		}

		return nil
	})
	if err != nil {
		return SalesOrderActionResult{}, err
	}

	return result, nil
}

func (s SalesOrderService) CancelSalesOrder(
	ctx context.Context,
	input SalesOrderActionInput,
) (SalesOrderActionResult, error) {
	if strings.TrimSpace(input.Reason) == "" {
		return SalesOrderActionResult{}, validationError(
			salesdomain.ErrSalesOrderRequiredField,
			map[string]any{"field": "reason"},
		)
	}
	if s.stockReserver != nil {
		return s.cancelAndReleaseSalesOrder(ctx, input)
	}

	return s.transition(ctx, input, "sales.order.cancelled", func(
		order salesdomain.SalesOrder,
		actorID string,
		changedAt time.Time,
	) (salesdomain.SalesOrder, error) {
		return order.CancelWithReason(actorID, input.Reason, changedAt)
	})
}

func (s SalesOrderService) MarkSalesOrderPacked(
	ctx context.Context,
	input SalesOrderActionInput,
) (SalesOrderActionResult, error) {
	return s.transition(ctx, input, "sales.order.packed", func(
		order salesdomain.SalesOrder,
		actorID string,
		changedAt time.Time,
	) (salesdomain.SalesOrder, error) {
		return order.MarkPacked(actorID, changedAt)
	})
}

func (s SalesOrderService) MarkSalesOrderHandedOver(
	ctx context.Context,
	input SalesOrderActionInput,
) (SalesOrderActionResult, error) {
	return s.transition(ctx, input, "sales.order.handed_over", func(
		order salesdomain.SalesOrder,
		actorID string,
		changedAt time.Time,
	) (salesdomain.SalesOrder, error) {
		if order.Status == salesdomain.SalesOrderStatusPacked {
			waitingHandover, err := order.MarkWaitingHandover(actorID, changedAt)
			if err != nil {
				return salesdomain.SalesOrder{}, err
			}

			return waitingHandover.MarkHandedOver(actorID, changedAt)
		}

		return order.MarkHandedOver(actorID, changedAt)
	})
}

func (s SalesOrderService) cancelAndReleaseSalesOrder(
	ctx context.Context,
	input SalesOrderActionInput,
) (SalesOrderActionResult, error) {
	if err := s.ensureReadyForWrite(); err != nil {
		return SalesOrderActionResult{}, err
	}
	if err := requireActor(input.ActorID); err != nil {
		return SalesOrderActionResult{}, err
	}

	var result SalesOrderActionResult
	err := s.store.WithinTx(ctx, func(txCtx context.Context, tx SalesOrderTx) error {
		current, err := tx.GetForUpdate(txCtx, input.ID)
		if err != nil {
			return mapSalesOrderError(err, map[string]any{"sales_order_id": strings.TrimSpace(input.ID)})
		}
		if err := ensureExpectedVersion(current, input.ExpectedVersion); err != nil {
			return err
		}

		now := s.now()
		if current.Status == salesdomain.SalesOrderStatusReserved {
			releaseResult, err := s.stockReserver.ReleaseSalesOrder(
				txCtx,
				newSalesOrderStockReleaseInput(current, input, now),
			)
			if err != nil {
				return mapSalesOrderError(err, map[string]any{"sales_order_id": current.ID})
			}
			if releaseResult.ReleasedReservationCount == 0 {
				return apperrors.Conflict(
					response.ErrorCodeConflict,
					"Reserved stock could not be released",
					errors.New("no active stock reservation found for sales order"),
					map[string]any{"sales_order_id": current.ID},
				)
			}
		}
		cancelled, err := current.CancelWithReason(input.ActorID, input.Reason, now)
		if err != nil {
			return mapSalesOrderError(err, map[string]any{
				"sales_order_id": current.ID,
				"status":         string(current.Status),
			})
		}
		if strings.TrimSpace(input.Note) != "" {
			cancelled.Note = strings.TrimSpace(input.Note)
		}
		if err := tx.Save(txCtx, cancelled); err != nil {
			return err
		}
		log, err := newSalesOrderAuditLog(
			input.ActorID,
			input.RequestID,
			"sales.order.cancelled",
			cancelled,
			salesOrderAuditData(current),
			salesOrderAuditData(cancelled),
			now,
		)
		if err != nil {
			return err
		}
		if err := tx.RecordAudit(txCtx, log); err != nil {
			return err
		}
		result = SalesOrderActionResult{
			SalesOrder:     cancelled,
			PreviousStatus: current.Status,
			CurrentStatus:  cancelled.Status,
			AuditLogID:     log.ID,
		}

		return nil
	})
	if err != nil {
		return SalesOrderActionResult{}, err
	}

	return result, nil
}

func (s SalesOrderService) transition(
	ctx context.Context,
	input SalesOrderActionInput,
	action string,
	transition func(salesdomain.SalesOrder, string, time.Time) (salesdomain.SalesOrder, error),
) (SalesOrderActionResult, error) {
	if err := s.ensureReadyForWrite(); err != nil {
		return SalesOrderActionResult{}, err
	}
	if err := requireActor(input.ActorID); err != nil {
		return SalesOrderActionResult{}, err
	}

	var result SalesOrderActionResult
	err := s.store.WithinTx(ctx, func(txCtx context.Context, tx SalesOrderTx) error {
		current, err := tx.GetForUpdate(txCtx, input.ID)
		if err != nil {
			return mapSalesOrderError(err, map[string]any{"sales_order_id": strings.TrimSpace(input.ID)})
		}
		if err := ensureExpectedVersion(current, input.ExpectedVersion); err != nil {
			return err
		}

		now := s.now()
		updated, err := transition(current, input.ActorID, now)
		if err != nil {
			return mapSalesOrderError(err, map[string]any{
				"sales_order_id": current.ID,
				"status":         string(current.Status),
			})
		}
		if strings.TrimSpace(input.Note) != "" {
			updated.Note = strings.TrimSpace(input.Note)
		}
		if err := tx.Save(txCtx, updated); err != nil {
			return err
		}
		log, err := newSalesOrderAuditLog(
			input.ActorID,
			input.RequestID,
			action,
			updated,
			salesOrderAuditData(current),
			salesOrderAuditData(updated),
			now,
		)
		if err != nil {
			return err
		}
		if err := tx.RecordAudit(txCtx, log); err != nil {
			return err
		}
		result = SalesOrderActionResult{
			SalesOrder:     updated,
			PreviousStatus: current.Status,
			CurrentStatus:  updated.Status,
			AuditLogID:     log.ID,
		}

		return nil
	})
	if err != nil {
		return SalesOrderActionResult{}, err
	}

	return result, nil
}

func (s SalesOrderService) newSalesOrderLineInputs(
	ctx context.Context,
	orderID string,
	inputs []SalesOrderLineInput,
	currencyCode string,
) ([]salesdomain.NewSalesOrderLineInput, error) {
	if len(inputs) == 0 {
		return nil, validationError(salesdomain.ErrSalesOrderRequiredField, map[string]any{"field": "lines"})
	}

	lines := make([]salesdomain.NewSalesOrderLineInput, 0, len(inputs))
	for index, input := range inputs {
		item, err := s.itemRead.Get(ctx, input.ItemID)
		if err != nil {
			return nil, mapMasterDataReadError(err, map[string]any{"item_id": strings.TrimSpace(input.ItemID)})
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
			return nil, validationError(salesdomain.ErrSalesOrderInvalidQuantity, map[string]any{"line_no": lineNo, "field": "ordered_qty"})
		}
		uomCode := firstNonBlank(input.UOMCode, item.UOMBase)
		if !strings.EqualFold(strings.TrimSpace(uomCode), strings.TrimSpace(item.UOMBase)) {
			return nil, apperrors.Unprocessable(
				response.ErrorCodeUOMConversionNotFound,
				"Sales order line UOM must be the item base UOM until conversion service is wired",
				domain.ErrUOMConversionMissing,
				map[string]any{
					"line_no":       lineNo,
					"item_id":       item.ID,
					"sku_code":      item.SKUCode,
					"from_uom_code": strings.ToUpper(strings.TrimSpace(uomCode)),
					"base_uom_code": strings.ToUpper(strings.TrimSpace(item.UOMBase)),
				},
			)
		}
		unitPrice, err := decimal.ParseUnitPrice(input.UnitPrice)
		if err != nil || unitPrice.IsNegative() {
			return nil, validationError(salesdomain.ErrSalesOrderInvalidAmount, map[string]any{"line_no": lineNo, "field": "unit_price"})
		}
		lineDiscount := decimal.MustMoneyAmount("0")
		if strings.TrimSpace(input.LineDiscountAmount) != "" {
			lineDiscount, err = decimal.ParseMoneyAmount(input.LineDiscountAmount)
			if err != nil || lineDiscount.IsNegative() {
				return nil, validationError(salesdomain.ErrSalesOrderInvalidAmount, map[string]any{"line_no": lineNo, "field": "line_discount_amount"})
			}
		}
		lineCurrency := firstNonBlank(input.CurrencyCode, currencyCode)
		if !strings.EqualFold(strings.TrimSpace(lineCurrency), decimal.CurrencyVND.String()) {
			return nil, validationError(salesdomain.ErrSalesOrderInvalidCurrency, map[string]any{"line_no": lineNo, "field": "currency_code"})
		}
		lines = append(lines, salesdomain.NewSalesOrderLineInput{
			ID:                 lineID,
			LineNo:             lineNo,
			ItemID:             item.ID,
			SKUCode:            item.SKUCode,
			ItemName:           item.Name,
			OrderedQty:         orderedQty,
			UOMCode:            item.UOMBase,
			BaseOrderedQty:     orderedQty,
			BaseUOMCode:        item.UOMBase,
			ConversionFactor:   decimal.MustQuantity("1"),
			UnitPrice:          unitPrice,
			CurrencyCode:       lineCurrency,
			LineDiscountAmount: lineDiscount,
			ReservedQty:        decimal.MustQuantity("0"),
			ShippedQty:         decimal.MustQuantity("0"),
			BatchID:            input.BatchID,
			BatchNo:            input.BatchNo,
		})
	}

	return lines, nil
}

func (s SalesOrderService) getOptionalWarehouse(ctx context.Context, id string) (domain.Warehouse, error) {
	if strings.TrimSpace(id) == "" {
		return domain.Warehouse{}, nil
	}
	warehouse, err := s.warehouseRead.GetWarehouse(ctx, id)
	if err != nil {
		return domain.Warehouse{}, mapMasterDataReadError(err, map[string]any{"warehouse_id": strings.TrimSpace(id)})
	}

	return warehouse, nil
}

func (s SalesOrderService) ensureReadyForWrite() error {
	if s.store == nil {
		return errors.New("sales order store is required")
	}
	if s.customerRead == nil {
		return errors.New("sales order customer reader is required")
	}
	if s.itemRead == nil {
		return errors.New("sales order item reader is required")
	}
	if s.warehouseRead == nil {
		return errors.New("sales order warehouse reader is required")
	}

	return nil
}

func (s SalesOrderService) now() time.Time {
	if s.clock == nil {
		return time.Now().UTC()
	}

	return s.clock().UTC()
}

func (s *PrototypeSalesOrderStore) List(
	_ context.Context,
	filter SalesOrderFilter,
) ([]salesdomain.SalesOrder, error) {
	if s == nil {
		return nil, errors.New("sales order store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows := make([]salesdomain.SalesOrder, 0, len(s.records))
	for _, order := range s.records {
		if salesOrderMatchesFilter(order, filter) {
			rows = append(rows, order.Clone())
		}
	}
	sortSalesOrders(rows)

	return rows, nil
}

func (s *PrototypeSalesOrderStore) Get(
	_ context.Context,
	id string,
) (salesdomain.SalesOrder, error) {
	if s == nil {
		return salesdomain.SalesOrder{}, errors.New("sales order store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	order, ok := s.records[strings.TrimSpace(id)]
	if !ok {
		return salesdomain.SalesOrder{}, ErrSalesOrderNotFound
	}

	return order.Clone(), nil
}

func (s *PrototypeSalesOrderStore) WithinTx(
	ctx context.Context,
	fn func(context.Context, SalesOrderTx) error,
) error {
	if s == nil {
		return errors.New("sales order store is required")
	}
	if fn == nil {
		return errors.New("sales order transaction function is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot := cloneSalesOrderMap(s.records)
	tx := &prototypeSalesOrderTx{store: s}
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

func (s *PrototypeSalesOrderStore) TransactionCount() int {
	if s == nil {
		return 0
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.txCount
}

func (tx *prototypeSalesOrderTx) GetForUpdate(
	_ context.Context,
	id string,
) (salesdomain.SalesOrder, error) {
	order, ok := tx.store.records[strings.TrimSpace(id)]
	if !ok {
		return salesdomain.SalesOrder{}, ErrSalesOrderNotFound
	}

	return order.Clone(), nil
}

func (tx *prototypeSalesOrderTx) Save(
	_ context.Context,
	order salesdomain.SalesOrder,
) error {
	if strings.TrimSpace(order.ID) == "" {
		return salesdomain.ErrSalesOrderRequiredField
	}
	tx.store.records[order.ID] = order.Clone()

	return nil
}

func (tx *prototypeSalesOrderTx) RecordAudit(
	_ context.Context,
	log audit.Log,
) error {
	tx.auditLogs = append(tx.auditLogs, log)

	return nil
}

func ensureExpectedVersion(order salesdomain.SalesOrder, expectedVersion int) error {
	if expectedVersion <= 0 || order.Version == expectedVersion {
		return nil
	}

	return apperrors.Conflict(
		ErrorCodeSalesOrderVersionConflict,
		"Sales order version changed",
		ErrSalesOrderVersionConflict,
		map[string]any{
			"sales_order_id":   order.ID,
			"expected_version": expectedVersion,
			"current_version":  order.Version,
		},
	)
}

func requireActor(actorID string) error {
	if strings.TrimSpace(actorID) == "" {
		return validationError(salesdomain.ErrSalesOrderTransitionActorRequired, map[string]any{"field": "actor_id"})
	}

	return nil
}

func validationError(cause error, details map[string]any) error {
	return apperrors.BadRequest(ErrorCodeSalesOrderValidation, "Sales order request is invalid", cause, details)
}

func mapSalesOrderError(err error, details map[string]any) error {
	if err == nil {
		return nil
	}
	if _, ok := apperrors.As(err); ok {
		return err
	}
	if errors.Is(err, ErrSalesOrderNotFound) {
		return apperrors.NotFound(ErrorCodeSalesOrderNotFound, "Sales order not found", err, details)
	}
	if errors.Is(err, salesdomain.ErrSalesOrderInvalidTransition) ||
		errors.Is(err, salesdomain.ErrSalesOrderInvalidStatus) {
		return apperrors.Conflict(ErrorCodeSalesOrderInvalidState, "Sales order state is invalid", err, details)
	}
	if errors.Is(err, salesdomain.ErrSalesOrderRequiredField) ||
		errors.Is(err, salesdomain.ErrSalesOrderTransitionActorRequired) ||
		errors.Is(err, salesdomain.ErrSalesOrderInvalidCurrency) ||
		errors.Is(err, salesdomain.ErrSalesOrderInvalidQuantity) ||
		errors.Is(err, salesdomain.ErrSalesOrderInvalidAmount) {
		return validationError(err, details)
	}

	return err
}

func mapMasterDataReadError(err error, details map[string]any) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, domain.ErrUOMConversionMissing) ||
		errors.Is(err, domain.ErrUOMConversionInactive) ||
		errors.Is(err, domain.ErrUOMInvalid) {
		return apperrors.Unprocessable(response.ErrorCodeUOMConversionNotFound, "UOM conversion is not available", err, details)
	}

	return apperrors.NotFound(response.ErrorCodeNotFound, "Referenced master data was not found", err, details)
}

func newSalesOrderStockReservationInput(
	order salesdomain.SalesOrder,
	action SalesOrderActionInput,
	reservedAt time.Time,
) SalesOrderStockReservationInput {
	lines := make([]SalesOrderStockReservationLineInput, 0, len(order.Lines))
	for _, line := range order.Lines {
		lines = append(lines, SalesOrderStockReservationLineInput{
			SalesOrderLineID: line.ID,
			LineNo:           line.LineNo,
			ItemID:           line.ItemID,
			SKUCode:          line.SKUCode,
			OrderedQty:       line.OrderedQty,
			BaseOrderedQty:   line.BaseOrderedQty,
			BaseUOMCode:      line.BaseUOMCode,
		})
	}

	return SalesOrderStockReservationInput{
		OrgID:         order.OrgID,
		SalesOrderID:  order.ID,
		OrderNo:       order.OrderNo,
		WarehouseID:   order.WarehouseID,
		WarehouseCode: order.WarehouseCode,
		ActorID:       action.ActorID,
		Reason:        firstNonBlank(action.Reason, "sales order confirm"),
		RequestID:     action.RequestID,
		ReservedAt:    reservedAt.UTC(),
		Lines:         lines,
	}
}

func newSalesOrderStockReleaseInput(
	order salesdomain.SalesOrder,
	action SalesOrderActionInput,
	releasedAt time.Time,
) SalesOrderStockReleaseInput {
	return SalesOrderStockReleaseInput{
		OrgID:        order.OrgID,
		SalesOrderID: order.ID,
		OrderNo:      order.OrderNo,
		ActorID:      action.ActorID,
		Reason:       action.Reason,
		RequestID:    action.RequestID,
		ReleasedAt:   releasedAt.UTC(),
	}
}

func applySalesOrderReservations(
	order salesdomain.SalesOrder,
	result SalesOrderStockReservationResult,
	actorID string,
	reservedAt time.Time,
) (salesdomain.SalesOrder, error) {
	reservationsByLine := make(map[string]SalesOrderReservedLine, len(result.Lines))
	for _, line := range result.Lines {
		reservationsByLine[strings.TrimSpace(line.SalesOrderLineID)] = line
	}

	updated := order.Clone()
	for index, line := range updated.Lines {
		reservation, ok := reservationsByLine[line.ID]
		if !ok {
			return salesdomain.SalesOrder{}, apperrors.Conflict(
				response.ErrorCodeInsufficientStock,
				"Sales order reservation did not cover every line",
				nil,
				map[string]any{"sales_order_line_id": line.ID},
			)
		}
		reservedQty, err := decimal.ParseQuantity(reservation.ReservedQty.String())
		if err != nil || reservedQty != line.BaseOrderedQty {
			return salesdomain.SalesOrder{}, apperrors.Conflict(
				response.ErrorCodeInsufficientStock,
				"Sales order reservation quantity does not match ordered quantity",
				err,
				map[string]any{
					"sales_order_line_id": line.ID,
					"ordered_qty":         line.BaseOrderedQty.String(),
					"reserved_qty":        reservation.ReservedQty.String(),
				},
			)
		}
		line.ReservedQty = reservedQty
		line.BatchID = reservation.BatchID
		line.BatchNo = reservation.BatchNo
		updated.Lines[index] = line
	}

	return updated.MarkReserved(actorID, reservedAt)
}

func newSalesOrderAuditLog(
	actorID string,
	requestID string,
	action string,
	order salesdomain.SalesOrder,
	beforeData map[string]any,
	afterData map[string]any,
	createdAt time.Time,
) (audit.Log, error) {
	return audit.NewLog(audit.NewLogInput{
		OrgID:      firstNonBlank(order.OrgID, defaultSalesOrderOrgID),
		ActorID:    strings.TrimSpace(actorID),
		Action:     strings.TrimSpace(action),
		EntityType: salesOrderEntityType,
		EntityID:   order.ID,
		RequestID:  strings.TrimSpace(requestID),
		BeforeData: beforeData,
		AfterData:  afterData,
		Metadata: map[string]any{
			"source":   "sales order application service",
			"order_no": order.OrderNo,
			"channel":  order.Channel,
		},
		CreatedAt: createdAt,
	})
}

func salesOrderAuditData(order salesdomain.SalesOrder) map[string]any {
	data := map[string]any{
		"order_no":      order.OrderNo,
		"customer_id":   order.CustomerID,
		"customer_code": order.CustomerCode,
		"channel":       order.Channel,
		"warehouse_id":  order.WarehouseID,
		"order_date":    order.OrderDate,
		"status":        string(order.Status),
		"currency_code": order.CurrencyCode.String(),
		"subtotal":      order.SubtotalAmount.String(),
		"discount":      order.DiscountAmount.String(),
		"total":         order.TotalAmount.String(),
		"line_count":    len(order.Lines),
		"version":       order.Version,
	}
	if strings.TrimSpace(order.CancelReason) != "" {
		data["cancel_reason"] = order.CancelReason
	}

	return data
}

func salesOrderMatchesFilter(order salesdomain.SalesOrder, filter SalesOrderFilter) bool {
	search := strings.ToLower(strings.TrimSpace(filter.Search))
	if search != "" {
		haystack := strings.ToLower(strings.Join([]string{
			order.OrderNo,
			order.CustomerCode,
			order.CustomerName,
			order.Channel,
		}, " "))
		if !strings.Contains(haystack, search) {
			return false
		}
	}
	if len(filter.Statuses) > 0 {
		matched := false
		for _, status := range filter.Statuses {
			if order.Status == salesdomain.NormalizeSalesOrderStatus(status) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if strings.TrimSpace(filter.CustomerID) != "" && order.CustomerID != strings.TrimSpace(filter.CustomerID) {
		return false
	}
	if strings.TrimSpace(filter.Channel) != "" && !strings.EqualFold(order.Channel, filter.Channel) {
		return false
	}
	if strings.TrimSpace(filter.WarehouseID) != "" && order.WarehouseID != strings.TrimSpace(filter.WarehouseID) {
		return false
	}
	if strings.TrimSpace(filter.DateFrom) != "" && order.OrderDate < strings.TrimSpace(filter.DateFrom) {
		return false
	}
	if strings.TrimSpace(filter.DateTo) != "" && order.OrderDate > strings.TrimSpace(filter.DateTo) {
		return false
	}

	return true
}

func sortSalesOrders(orders []salesdomain.SalesOrder) {
	sort.SliceStable(orders, func(i, j int) bool {
		if orders[i].OrderDate == orders[j].OrderDate {
			return orders[i].OrderNo < orders[j].OrderNo
		}

		return orders[i].OrderDate > orders[j].OrderDate
	})
}

func cloneSalesOrderMap(records map[string]salesdomain.SalesOrder) map[string]salesdomain.SalesOrder {
	clone := make(map[string]salesdomain.SalesOrder, len(records))
	for id, order := range records {
		clone[id] = order.Clone()
	}

	return clone
}

func newSalesOrderID(now time.Time) string {
	return fmt.Sprintf("so-%s-%06d", now.UTC().Format("060102"), now.UTC().UnixNano()%1000000)
}

func newSalesOrderNo(now time.Time) string {
	return fmt.Sprintf("SO-%s-%06d", now.UTC().Format("060102"), now.UTC().UnixNano()%1000000)
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}

	return ""
}

func prototypeSalesOrders() []salesdomain.SalesOrder {
	baseTime := time.Date(2026, 4, 28, 8, 0, 0, 0, time.UTC)
	draft, err := salesdomain.NewSalesOrderDocument(salesdomain.NewSalesOrderDocumentInput{
		ID:            "so-260428-0001",
		OrgID:         defaultSalesOrderOrgID,
		OrderNo:       "SO-260428-0001",
		CustomerID:    "cus-dl-minh-anh",
		CustomerCode:  "CUS-DL-MINHANH",
		CustomerName:  "Minh Anh Distributor",
		Channel:       "B2B",
		WarehouseID:   "wh-hcm-fg",
		WarehouseCode: "WH-HCM-FG",
		OrderDate:     "2026-04-28",
		CurrencyCode:  decimal.CurrencyVND.String(),
		Lines: []salesdomain.NewSalesOrderLineInput{
			{
				ID:               "so-260428-0001-line-01",
				LineNo:           1,
				ItemID:           "item-serum-30ml",
				SKUCode:          "SERUM-30ML",
				ItemName:         "Hydrating Serum 30ml",
				OrderedQty:       decimal.MustQuantity("12"),
				UOMCode:          "EA",
				BaseOrderedQty:   decimal.MustQuantity("12"),
				BaseUOMCode:      "EA",
				ConversionFactor: decimal.MustQuantity("1"),
				UnitPrice:        decimal.MustUnitPrice("125000"),
				CurrencyCode:     decimal.CurrencyVND.String(),
			},
		},
		CreatedAt: baseTime,
		CreatedBy: "user-sales",
		UpdatedAt: baseTime,
	})
	if err != nil {
		panic(fmt.Sprintf("invalid prototype draft sales order: %v", err))
	}
	confirmed, err := salesdomain.NewSalesOrderDocument(salesdomain.NewSalesOrderDocumentInput{
		ID:            "so-260428-0002",
		OrgID:         defaultSalesOrderOrgID,
		OrderNo:       "SO-260428-0002",
		CustomerID:    "cus-mp-shopee",
		CustomerCode:  "CUS-MP-SHOPEE",
		CustomerName:  "Shopee Marketplace",
		Channel:       "MP",
		WarehouseID:   "wh-hcm-fg",
		WarehouseCode: "WH-HCM-FG",
		OrderDate:     "2026-04-28",
		CurrencyCode:  decimal.CurrencyVND.String(),
		Lines: []salesdomain.NewSalesOrderLineInput{
			{
				ID:               "so-260428-0002-line-01",
				LineNo:           1,
				ItemID:           "item-cream-50g",
				SKUCode:          "CREAM-50G",
				ItemName:         "Repair Cream 50g",
				OrderedQty:       decimal.MustQuantity("3"),
				UOMCode:          "EA",
				BaseOrderedQty:   decimal.MustQuantity("3"),
				BaseUOMCode:      "EA",
				ConversionFactor: decimal.MustQuantity("1"),
				UnitPrice:        decimal.MustUnitPrice("95000"),
				CurrencyCode:     decimal.CurrencyVND.String(),
			},
		},
		CreatedAt: baseTime.Add(30 * time.Minute),
		CreatedBy: "user-sales",
		UpdatedAt: baseTime.Add(30 * time.Minute),
	})
	if err != nil {
		panic(fmt.Sprintf("invalid prototype confirmed sales order: %v", err))
	}
	confirmed, err = confirmed.Confirm("user-sales", baseTime.Add(35*time.Minute))
	if err != nil {
		panic(fmt.Sprintf("invalid prototype sales order transition: %v", err))
	}
	packing, err := salesdomain.NewSalesOrderDocument(salesdomain.NewSalesOrderDocumentInput{
		ID:            "so-260428-0003",
		OrgID:         defaultSalesOrderOrgID,
		OrderNo:       "SO-260428-0003",
		CustomerID:    "cus-mp-tiktok",
		CustomerCode:  "CUS-MP-TIKTOK",
		CustomerName:  "TikTok Shop",
		Channel:       "MP",
		WarehouseID:   "wh-hcm-fg",
		WarehouseCode: "WH-HCM-FG",
		OrderDate:     "2026-04-28",
		CurrencyCode:  decimal.CurrencyVND.String(),
		Lines: []salesdomain.NewSalesOrderLineInput{
			{
				ID:               "so-260428-0003-line-01",
				LineNo:           1,
				ItemID:           "item-serum-30ml",
				SKUCode:          "SERUM-30ML",
				ItemName:         "Hydrating Serum 30ml",
				OrderedQty:       decimal.MustQuantity("3"),
				UOMCode:          "EA",
				BaseOrderedQty:   decimal.MustQuantity("3"),
				BaseUOMCode:      "EA",
				ConversionFactor: decimal.MustQuantity("1"),
				UnitPrice:        decimal.MustUnitPrice("125000"),
				CurrencyCode:     decimal.CurrencyVND.String(),
				ReservedQty:      decimal.MustQuantity("3"),
				BatchID:          "batch-serum-2604a",
				BatchNo:          "LOT-2604A",
			},
		},
		CreatedAt: baseTime.Add(45 * time.Minute),
		CreatedBy: "user-sales",
		UpdatedAt: baseTime.Add(45 * time.Minute),
	})
	if err != nil {
		panic(fmt.Sprintf("invalid prototype packing sales order: %v", err))
	}
	packing, err = packing.Confirm("user-sales", baseTime.Add(50*time.Minute))
	if err != nil {
		panic(fmt.Sprintf("invalid prototype packing order confirm: %v", err))
	}
	packing, err = packing.MarkReserved("user-sales", baseTime.Add(55*time.Minute))
	if err != nil {
		panic(fmt.Sprintf("invalid prototype packing order reserve: %v", err))
	}
	packing, err = packing.StartPicking("user-picker", baseTime.Add(60*time.Minute))
	if err != nil {
		panic(fmt.Sprintf("invalid prototype packing order start pick: %v", err))
	}
	packing, err = packing.MarkPicked("user-picker", baseTime.Add(65*time.Minute))
	if err != nil {
		panic(fmt.Sprintf("invalid prototype packing order picked: %v", err))
	}
	packing, err = packing.StartPacking("user-packer", baseTime.Add(70*time.Minute))
	if err != nil {
		panic(fmt.Sprintf("invalid prototype packing order start pack: %v", err))
	}

	orders := []salesdomain.SalesOrder{draft, confirmed, packing}
	orders = append(orders, prototypeHandoverSalesOrders(baseTime)...)
	orders = append(orders, prototypeSprint2QASalesOrders(baseTime)...)

	return orders
}

func prototypeHandoverSalesOrders(baseTime time.Time) []salesdomain.SalesOrder {
	return []salesdomain.SalesOrder{
		mustPrototypeWaitingHandoverSalesOrder(
			"so-260426-001",
			"SO-260426-001",
			"cus-mp-shopee",
			"CUS-MP-SHOPEE",
			"Shopee Marketplace",
			"item-serum-30ml",
			"SERUM-30ML",
			"Hydrating Serum 30ml",
			"12",
			baseTime.Add(-48*time.Hour),
		),
		mustPrototypeWaitingHandoverSalesOrder(
			"so-260426-002",
			"SO-260426-002",
			"cus-mp-tiktok",
			"CUS-MP-TIKTOK",
			"TikTok Shop",
			"item-cream-50g",
			"CREAM-50G",
			"Repair Cream 50g",
			"3",
			baseTime.Add(-47*time.Hour),
		),
		mustPrototypeWaitingHandoverSalesOrder(
			"so-260426-003",
			"SO-260426-003",
			"cus-dl-minh-anh",
			"CUS-DL-MINHANH",
			"Minh Anh Distributor",
			"item-serum-30ml",
			"SERUM-30ML",
			"Hydrating Serum 30ml",
			"2",
			baseTime.Add(-46*time.Hour),
		),
	}
}

func mustPrototypeWaitingHandoverSalesOrder(
	id string,
	orderNo string,
	customerID string,
	customerCode string,
	customerName string,
	itemID string,
	skuCode string,
	itemName string,
	qty string,
	createdAt time.Time,
) salesdomain.SalesOrder {
	order, err := salesdomain.NewSalesOrderDocument(salesdomain.NewSalesOrderDocumentInput{
		ID:            id,
		OrgID:         defaultSalesOrderOrgID,
		OrderNo:       orderNo,
		CustomerID:    customerID,
		CustomerCode:  customerCode,
		CustomerName:  customerName,
		Channel:       "MP",
		WarehouseID:   "wh-hcm",
		WarehouseCode: "HCM",
		OrderDate:     "2026-04-26",
		CurrencyCode:  decimal.CurrencyVND.String(),
		Lines: []salesdomain.NewSalesOrderLineInput{
			{
				ID:               id + "-line-01",
				LineNo:           1,
				ItemID:           itemID,
				SKUCode:          skuCode,
				ItemName:         itemName,
				OrderedQty:       decimal.MustQuantity(qty),
				UOMCode:          "EA",
				BaseOrderedQty:   decimal.MustQuantity(qty),
				BaseUOMCode:      "EA",
				ConversionFactor: decimal.MustQuantity("1"),
				UnitPrice:        decimal.MustUnitPrice("125000"),
				CurrencyCode:     decimal.CurrencyVND.String(),
				ReservedQty:      decimal.MustQuantity(qty),
			},
		},
		CreatedAt: createdAt,
		CreatedBy: "user-sales",
		UpdatedAt: createdAt,
	})
	if err != nil {
		panic(fmt.Sprintf("invalid prototype handover sales order: %v", err))
	}
	order, err = order.Confirm("user-sales", createdAt.Add(5*time.Minute))
	if err != nil {
		panic(fmt.Sprintf("invalid prototype handover order confirm: %v", err))
	}
	order, err = order.MarkReserved("user-sales", createdAt.Add(10*time.Minute))
	if err != nil {
		panic(fmt.Sprintf("invalid prototype handover order reserve: %v", err))
	}
	order, err = order.StartPicking("user-picker", createdAt.Add(15*time.Minute))
	if err != nil {
		panic(fmt.Sprintf("invalid prototype handover order start pick: %v", err))
	}
	order, err = order.MarkPicked("user-picker", createdAt.Add(20*time.Minute))
	if err != nil {
		panic(fmt.Sprintf("invalid prototype handover order picked: %v", err))
	}
	order, err = order.StartPacking("user-packer", createdAt.Add(25*time.Minute))
	if err != nil {
		panic(fmt.Sprintf("invalid prototype handover order start pack: %v", err))
	}
	order, err = order.MarkPacked("user-packer", createdAt.Add(30*time.Minute))
	if err != nil {
		panic(fmt.Sprintf("invalid prototype handover order packed: %v", err))
	}
	order, err = order.MarkWaitingHandover("user-handover-operator", createdAt.Add(35*time.Minute))
	if err != nil {
		panic(fmt.Sprintf("invalid prototype handover order waiting: %v", err))
	}

	return order
}

type prototypeSalesOrderParty struct {
	customerID   string
	customerCode string
	customerName string
	channel      string
}

type prototypeSalesOrderProduct struct {
	itemID    string
	skuCode   string
	itemName  string
	batchID   string
	batchNo   string
	unitPrice string
}

type prototypeSalesOrderSeed struct {
	id           string
	orderNo      string
	orderDate    string
	partyKey     string
	productKey   string
	qty          string
	targetStatus salesdomain.SalesOrderStatus
	createdAt    time.Time
}

var prototypeSalesOrderParties = map[string]prototypeSalesOrderParty{
	"minh-anh": {
		customerID:   "cus-dl-minh-anh",
		customerCode: "CUS-DL-MINHANH",
		customerName: "Minh Anh Distributor",
		channel:      "B2B",
	},
	"linh-chi": {
		customerID:   "cus-dealer-linh-chi",
		customerCode: "CUS-DL-LINHCHI",
		customerName: "Linh Chi Dealer",
		channel:      "DEALER",
	},
	"shopee": {
		customerID:   "cus-mp-shopee",
		customerCode: "CUS-MP-SHOPEE",
		customerName: "Shopee Marketplace",
		channel:      "MP",
	},
	"tiktok": {
		customerID:   "cus-mp-tiktok",
		customerCode: "CUS-MP-TIKTOK",
		customerName: "TikTok Shop",
		channel:      "MP",
	},
}

var prototypeSalesOrderProducts = map[string]prototypeSalesOrderProduct{
	"serum": {
		itemID:    "item-serum-30ml",
		skuCode:   "SERUM-30ML",
		itemName:  "Hydrating Serum 30ml",
		batchID:   "batch-serum-2604a",
		batchNo:   "LOT-2604A",
		unitPrice: "125000",
	},
	"cream": {
		itemID:    "item-cream-50g",
		skuCode:   "CREAM-50G",
		itemName:  "Repair Cream 50g",
		batchID:   "batch-cream-2603b",
		batchNo:   "LOT-2603B",
		unitPrice: "95000",
	},
}

func prototypeSprint2QASalesOrders(baseTime time.Time) []salesdomain.SalesOrder {
	seeds := []prototypeSalesOrderSeed{
		{
			id:           "so-260427-0001",
			orderNo:      "SO-260427-0001",
			orderDate:    "2026-04-27",
			partyKey:     "linh-chi",
			productKey:   "serum",
			qty:          "4",
			targetStatus: salesdomain.SalesOrderStatusDraft,
			createdAt:    baseTime.Add(-24 * time.Hour),
		},
		{id: "so-260427-0002", orderNo: "SO-260427-0002", orderDate: "2026-04-27", partyKey: "shopee", productKey: "cream", qty: "2", targetStatus: salesdomain.SalesOrderStatusConfirmed, createdAt: baseTime.Add(-23 * time.Hour)},
		{id: "so-260427-0003", orderNo: "SO-260427-0003", orderDate: "2026-04-27", partyKey: "minh-anh", productKey: "serum", qty: "6", targetStatus: salesdomain.SalesOrderStatusReserved, createdAt: baseTime.Add(-22 * time.Hour)},
		{id: "so-260427-0004", orderNo: "SO-260427-0004", orderDate: "2026-04-27", partyKey: "tiktok", productKey: "cream", qty: "5", targetStatus: salesdomain.SalesOrderStatusPicking, createdAt: baseTime.Add(-21 * time.Hour)},
		{id: "so-260427-0005", orderNo: "SO-260427-0005", orderDate: "2026-04-27", partyKey: "linh-chi", productKey: "serum", qty: "1", targetStatus: salesdomain.SalesOrderStatusPicked, createdAt: baseTime.Add(-20 * time.Hour)},
		{id: "so-260427-0006", orderNo: "SO-260427-0006", orderDate: "2026-04-27", partyKey: "shopee", productKey: "cream", qty: "7", targetStatus: salesdomain.SalesOrderStatusPacking, createdAt: baseTime.Add(-19 * time.Hour)},
		{id: "so-260427-0007", orderNo: "SO-260427-0007", orderDate: "2026-04-27", partyKey: "minh-anh", productKey: "serum", qty: "8", targetStatus: salesdomain.SalesOrderStatusPacked, createdAt: baseTime.Add(-18 * time.Hour)},
		{id: "so-260427-0008", orderNo: "SO-260427-0008", orderDate: "2026-04-27", partyKey: "tiktok", productKey: "cream", qty: "3", targetStatus: salesdomain.SalesOrderStatusWaitingHandover, createdAt: baseTime.Add(-17 * time.Hour)},
		{id: "so-260427-0009", orderNo: "SO-260427-0009", orderDate: "2026-04-27", partyKey: "linh-chi", productKey: "serum", qty: "9", targetStatus: salesdomain.SalesOrderStatusHandedOver, createdAt: baseTime.Add(-16 * time.Hour)},
		{id: "so-260427-0010", orderNo: "SO-260427-0010", orderDate: "2026-04-27", partyKey: "shopee", productKey: "cream", qty: "4", targetStatus: salesdomain.SalesOrderStatusCancelled, createdAt: baseTime.Add(-15 * time.Hour)},
		{id: "so-260427-0011", orderNo: "SO-260427-0011", orderDate: "2026-04-27", partyKey: "minh-anh", productKey: "serum", qty: "120", targetStatus: salesdomain.SalesOrderStatusReservationFailed, createdAt: baseTime.Add(-14 * time.Hour)},
		{id: "so-260427-0012", orderNo: "SO-260427-0012", orderDate: "2026-04-27", partyKey: "tiktok", productKey: "cream", qty: "2", targetStatus: salesdomain.SalesOrderStatusPickException, createdAt: baseTime.Add(-13 * time.Hour)},
		{id: "so-260427-0013", orderNo: "SO-260427-0013", orderDate: "2026-04-27", partyKey: "linh-chi", productKey: "serum", qty: "5", targetStatus: salesdomain.SalesOrderStatusPackException, createdAt: baseTime.Add(-12 * time.Hour)},
		{id: "so-260427-0014", orderNo: "SO-260427-0014", orderDate: "2026-04-27", partyKey: "shopee", productKey: "cream", qty: "6", targetStatus: salesdomain.SalesOrderStatusHandoverException, createdAt: baseTime.Add(-11 * time.Hour)},
	}

	orders := make([]salesdomain.SalesOrder, 0, len(seeds))
	for _, seed := range seeds {
		orders = append(orders, mustPrototypeSalesOrder(seed))
	}

	return orders
}

func mustPrototypeSalesOrder(seed prototypeSalesOrderSeed) salesdomain.SalesOrder {
	party, ok := prototypeSalesOrderParties[seed.partyKey]
	if !ok {
		panic(fmt.Sprintf("unknown sprint 2 prototype sales order party %q", seed.partyKey))
	}
	product, ok := prototypeSalesOrderProducts[seed.productKey]
	if !ok {
		panic(fmt.Sprintf("unknown sprint 2 prototype sales order product %q", seed.productKey))
	}

	reserveLine := prototypeSalesOrderStatusNeedsReservation(seed.targetStatus)
	line := salesdomain.NewSalesOrderLineInput{
		ID:               seed.id + "-line-01",
		LineNo:           1,
		ItemID:           product.itemID,
		SKUCode:          product.skuCode,
		ItemName:         product.itemName,
		OrderedQty:       decimal.MustQuantity(seed.qty),
		UOMCode:          "EA",
		BaseOrderedQty:   decimal.MustQuantity(seed.qty),
		BaseUOMCode:      "EA",
		ConversionFactor: decimal.MustQuantity("1"),
		UnitPrice:        decimal.MustUnitPrice(product.unitPrice),
		CurrencyCode:     decimal.CurrencyVND.String(),
	}
	if reserveLine {
		line.ReservedQty = decimal.MustQuantity(seed.qty)
		line.BatchID = product.batchID
		line.BatchNo = product.batchNo
	}

	order, err := salesdomain.NewSalesOrderDocument(salesdomain.NewSalesOrderDocumentInput{
		ID:            seed.id,
		OrgID:         defaultSalesOrderOrgID,
		OrderNo:       seed.orderNo,
		CustomerID:    party.customerID,
		CustomerCode:  party.customerCode,
		CustomerName:  party.customerName,
		Channel:       party.channel,
		WarehouseID:   "wh-hcm-fg",
		WarehouseCode: "WH-HCM-FG",
		OrderDate:     seed.orderDate,
		CurrencyCode:  decimal.CurrencyVND.String(),
		Lines:         []salesdomain.NewSalesOrderLineInput{line},
		CreatedAt:     seed.createdAt,
		CreatedBy:     "user-sales",
		UpdatedAt:     seed.createdAt,
	})
	if err != nil {
		panic(fmt.Sprintf("invalid sprint 2 prototype sales order %s: %v", seed.orderNo, err))
	}

	return mustPrototypeSalesOrderStatus(order, seed.targetStatus, seed.createdAt)
}

func prototypeSalesOrderStatusNeedsReservation(status salesdomain.SalesOrderStatus) bool {
	switch status {
	case salesdomain.SalesOrderStatusReserved,
		salesdomain.SalesOrderStatusPicking,
		salesdomain.SalesOrderStatusPicked,
		salesdomain.SalesOrderStatusPacking,
		salesdomain.SalesOrderStatusPacked,
		salesdomain.SalesOrderStatusWaitingHandover,
		salesdomain.SalesOrderStatusHandedOver,
		salesdomain.SalesOrderStatusPickException,
		salesdomain.SalesOrderStatusPackException,
		salesdomain.SalesOrderStatusHandoverException:
		return true
	default:
		return false
	}
}

func mustPrototypeSalesOrderStatus(
	order salesdomain.SalesOrder,
	targetStatus salesdomain.SalesOrderStatus,
	createdAt time.Time,
) salesdomain.SalesOrder {
	switch targetStatus {
	case salesdomain.SalesOrderStatusDraft:
		return order
	case salesdomain.SalesOrderStatusConfirmed:
		return mustPrototypeSalesOrderTransition(order, targetStatus, func() (salesdomain.SalesOrder, error) {
			return order.Confirm("user-sales", createdAt.Add(5*time.Minute))
		})
	case salesdomain.SalesOrderStatusReserved:
		order = mustPrototypeSalesOrderStatus(order, salesdomain.SalesOrderStatusConfirmed, createdAt)
		return mustPrototypeSalesOrderTransition(order, targetStatus, func() (salesdomain.SalesOrder, error) {
			return order.MarkReserved("user-sales", createdAt.Add(10*time.Minute))
		})
	case salesdomain.SalesOrderStatusPicking:
		order = mustPrototypeSalesOrderStatus(order, salesdomain.SalesOrderStatusReserved, createdAt)
		return mustPrototypeSalesOrderTransition(order, targetStatus, func() (salesdomain.SalesOrder, error) {
			return order.StartPicking("user-picker", createdAt.Add(15*time.Minute))
		})
	case salesdomain.SalesOrderStatusPicked:
		order = mustPrototypeSalesOrderStatus(order, salesdomain.SalesOrderStatusPicking, createdAt)
		return mustPrototypeSalesOrderTransition(order, targetStatus, func() (salesdomain.SalesOrder, error) {
			return order.MarkPicked("user-picker", createdAt.Add(20*time.Minute))
		})
	case salesdomain.SalesOrderStatusPacking:
		order = mustPrototypeSalesOrderStatus(order, salesdomain.SalesOrderStatusPicked, createdAt)
		return mustPrototypeSalesOrderTransition(order, targetStatus, func() (salesdomain.SalesOrder, error) {
			return order.StartPacking("user-packer", createdAt.Add(25*time.Minute))
		})
	case salesdomain.SalesOrderStatusPacked:
		order = mustPrototypeSalesOrderStatus(order, salesdomain.SalesOrderStatusPacking, createdAt)
		return mustPrototypeSalesOrderTransition(order, targetStatus, func() (salesdomain.SalesOrder, error) {
			return order.MarkPacked("user-packer", createdAt.Add(30*time.Minute))
		})
	case salesdomain.SalesOrderStatusWaitingHandover:
		order = mustPrototypeSalesOrderStatus(order, salesdomain.SalesOrderStatusPacked, createdAt)
		return mustPrototypeSalesOrderTransition(order, targetStatus, func() (salesdomain.SalesOrder, error) {
			return order.MarkWaitingHandover("user-handover-operator", createdAt.Add(35*time.Minute))
		})
	case salesdomain.SalesOrderStatusHandedOver:
		order = mustPrototypeSalesOrderStatus(order, salesdomain.SalesOrderStatusWaitingHandover, createdAt)
		return mustPrototypeSalesOrderTransition(order, targetStatus, func() (salesdomain.SalesOrder, error) {
			return order.MarkHandedOver("user-handover-operator", createdAt.Add(40*time.Minute))
		})
	case salesdomain.SalesOrderStatusCancelled:
		order = mustPrototypeSalesOrderStatus(order, salesdomain.SalesOrderStatusConfirmed, createdAt)
		return mustPrototypeSalesOrderTransition(order, targetStatus, func() (salesdomain.SalesOrder, error) {
			return order.CancelWithReason("user-sales", "QA cancelled order seed", createdAt.Add(10*time.Minute))
		})
	case salesdomain.SalesOrderStatusReservationFailed:
		order = mustPrototypeSalesOrderStatus(order, salesdomain.SalesOrderStatusConfirmed, createdAt)
		return mustPrototypeSalesOrderTransition(order, targetStatus, func() (salesdomain.SalesOrder, error) {
			return order.MarkReservationFailed("user-sales", createdAt.Add(10*time.Minute))
		})
	case salesdomain.SalesOrderStatusPickException:
		order = mustPrototypeSalesOrderStatus(order, salesdomain.SalesOrderStatusPicking, createdAt)
		return mustPrototypeSalesOrderTransition(order, targetStatus, func() (salesdomain.SalesOrder, error) {
			return order.MarkPickException("user-picker", createdAt.Add(20*time.Minute))
		})
	case salesdomain.SalesOrderStatusPackException:
		order = mustPrototypeSalesOrderStatus(order, salesdomain.SalesOrderStatusPacking, createdAt)
		return mustPrototypeSalesOrderTransition(order, targetStatus, func() (salesdomain.SalesOrder, error) {
			return order.MarkPackException("user-packer", createdAt.Add(30*time.Minute))
		})
	case salesdomain.SalesOrderStatusHandoverException:
		order = mustPrototypeSalesOrderStatus(order, salesdomain.SalesOrderStatusWaitingHandover, createdAt)
		return mustPrototypeSalesOrderTransition(order, targetStatus, func() (salesdomain.SalesOrder, error) {
			return order.MarkHandoverException("user-handover-operator", createdAt.Add(40*time.Minute))
		})
	default:
		panic(fmt.Sprintf("unsupported sprint 2 prototype sales order status %q", targetStatus))
	}
}

func mustPrototypeSalesOrderTransition(
	order salesdomain.SalesOrder,
	targetStatus salesdomain.SalesOrderStatus,
	transition func() (salesdomain.SalesOrder, error),
) salesdomain.SalesOrder {
	updated, err := transition()
	if err != nil {
		panic(fmt.Sprintf("invalid prototype sales order transition %s -> %s: %v", order.Status, targetStatus, err))
	}

	return updated
}
