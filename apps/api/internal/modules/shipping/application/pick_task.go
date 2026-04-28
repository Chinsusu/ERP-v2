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
	salesdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrPickTaskNotFound = errors.New("pick task not found")
var ErrPickTaskDuplicate = errors.New("pick task already exists")
var ErrPickTaskSalesOrderNotReserved = errors.New("sales order must be reserved before generating pick task")
var ErrPickTaskReservationMissing = errors.New("active stock reservation is required for every sales order line")

type PickTaskStore interface {
	ListPickTasks(ctx context.Context) ([]domain.PickTask, error)
	GetPickTask(ctx context.Context, id string) (domain.PickTask, error)
	GetPickTaskBySalesOrder(ctx context.Context, salesOrderID string) (domain.PickTask, error)
	SavePickTask(ctx context.Context, task domain.PickTask) error
}

type PickTaskFilter struct {
	WarehouseID string
	Status      domain.PickTaskStatus
	AssignedTo  string
}

type ListPickTasks struct {
	store PickTaskStore
}

type GetPickTask struct {
	store PickTaskStore
}

type GeneratePickTaskFromReservedOrder struct {
	store    PickTaskStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type GeneratePickTaskFromReservedOrderInput struct {
	SalesOrder   salesdomain.SalesOrder
	Reservations []inventorydomain.StockReservation
	AssignedTo   string
	ActorID      string
	RequestID    string
}

type PickTaskResult struct {
	PickTask   domain.PickTask
	AuditLogID string
}

type PickTaskActionInput struct {
	PickTaskID string
	ActorID    string
	RequestID  string
}

type ConfirmPickTaskLineInput struct {
	PickTaskID string
	LineID     string
	PickedQty  string
	ActorID    string
	RequestID  string
}

type ReportPickTaskExceptionInput struct {
	PickTaskID    string
	LineID        string
	ExceptionCode string
	ActorID       string
	RequestID     string
	Investigation string
}

type StartPickTask struct {
	store    PickTaskStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type ConfirmPickTaskLine struct {
	store    PickTaskStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type CompletePickTask struct {
	store    PickTaskStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type ReportPickTaskException struct {
	store    PickTaskStore
	auditLog audit.LogStore
	clock    func() time.Time
}

func NewGeneratePickTaskFromReservedOrder(
	store PickTaskStore,
	auditLog audit.LogStore,
) GeneratePickTaskFromReservedOrder {
	return GeneratePickTaskFromReservedOrder{
		store:    store,
		auditLog: auditLog,
		clock:    func() time.Time { return time.Now().UTC() },
	}
}

func NewListPickTasks(store PickTaskStore) ListPickTasks {
	return ListPickTasks{store: store}
}

func NewGetPickTask(store PickTaskStore) GetPickTask {
	return GetPickTask{store: store}
}

func NewStartPickTask(store PickTaskStore, auditLog audit.LogStore) StartPickTask {
	return StartPickTask{store: store, auditLog: auditLog, clock: func() time.Time { return time.Now().UTC() }}
}

func NewConfirmPickTaskLine(store PickTaskStore, auditLog audit.LogStore) ConfirmPickTaskLine {
	return ConfirmPickTaskLine{store: store, auditLog: auditLog, clock: func() time.Time { return time.Now().UTC() }}
}

func NewCompletePickTask(store PickTaskStore, auditLog audit.LogStore) CompletePickTask {
	return CompletePickTask{store: store, auditLog: auditLog, clock: func() time.Time { return time.Now().UTC() }}
}

func NewReportPickTaskException(store PickTaskStore, auditLog audit.LogStore) ReportPickTaskException {
	return ReportPickTaskException{store: store, auditLog: auditLog, clock: func() time.Time { return time.Now().UTC() }}
}

func (uc ListPickTasks) Execute(ctx context.Context, filter PickTaskFilter) ([]domain.PickTask, error) {
	if uc.store == nil {
		return nil, errors.New("pick task store is required")
	}
	tasks, err := uc.store.ListPickTasks(ctx)
	if err != nil {
		return nil, err
	}

	filter.WarehouseID = strings.TrimSpace(filter.WarehouseID)
	filter.AssignedTo = strings.TrimSpace(filter.AssignedTo)
	filter.Status = domain.NormalizePickTaskStatus(filter.Status)
	if filter.WarehouseID == "" && filter.AssignedTo == "" && filter.Status == "" {
		return tasks, nil
	}

	filtered := make([]domain.PickTask, 0, len(tasks))
	for _, task := range tasks {
		if filter.WarehouseID != "" && strings.TrimSpace(task.WarehouseID) != filter.WarehouseID {
			continue
		}
		if filter.AssignedTo != "" && strings.TrimSpace(task.AssignedTo) != filter.AssignedTo {
			continue
		}
		if filter.Status != "" && task.Status != filter.Status {
			continue
		}
		filtered = append(filtered, task.Clone())
	}

	return filtered, nil
}

func (uc GetPickTask) Execute(ctx context.Context, id string) (domain.PickTask, error) {
	if uc.store == nil {
		return domain.PickTask{}, errors.New("pick task store is required")
	}

	return uc.store.GetPickTask(ctx, id)
}

func (uc GeneratePickTaskFromReservedOrder) Execute(
	ctx context.Context,
	input GeneratePickTaskFromReservedOrderInput,
) (PickTaskResult, error) {
	if uc.store == nil {
		return PickTaskResult{}, errors.New("pick task store is required")
	}
	if uc.auditLog == nil {
		return PickTaskResult{}, errors.New("audit log store is required")
	}
	if strings.TrimSpace(input.ActorID) == "" {
		return PickTaskResult{}, domain.ErrPickTaskActorRequired
	}
	order := input.SalesOrder.Clone()
	if salesdomain.NormalizeSalesOrderStatus(order.Status) != salesdomain.SalesOrderStatusReserved {
		return PickTaskResult{}, ErrPickTaskSalesOrderNotReserved
	}
	if _, err := uc.store.GetPickTaskBySalesOrder(ctx, order.ID); err == nil {
		return PickTaskResult{}, ErrPickTaskDuplicate
	} else if !errors.Is(err, ErrPickTaskNotFound) {
		return PickTaskResult{}, err
	}

	now := uc.clock()
	lines, err := newPickTaskLinesFromReservations(order, input.Reservations)
	if err != nil {
		return PickTaskResult{}, err
	}
	task, err := domain.NewPickTask(domain.NewPickTaskInput{
		ID:            newPickTaskID(order.ID),
		OrgID:         order.OrgID,
		PickTaskNo:    newPickTaskNo(order.OrderNo),
		SalesOrderID:  order.ID,
		OrderNo:       order.OrderNo,
		WarehouseID:   order.WarehouseID,
		WarehouseCode: order.WarehouseCode,
		CreatedAt:     now,
		Lines:         lines,
	})
	if err != nil {
		return PickTaskResult{}, err
	}
	if strings.TrimSpace(input.AssignedTo) != "" {
		task, err = task.Assign(input.AssignedTo, now)
		if err != nil {
			return PickTaskResult{}, err
		}
	}
	if err := uc.store.SavePickTask(ctx, task); err != nil {
		return PickTaskResult{}, err
	}
	log, err := newPickTaskAuditLog(input.ActorID, input.RequestID, "shipping.pick_task.created", task, now)
	if err != nil {
		return PickTaskResult{}, err
	}
	if err := uc.auditLog.Record(ctx, log); err != nil {
		return PickTaskResult{}, err
	}

	return PickTaskResult{PickTask: task, AuditLogID: log.ID}, nil
}

func (uc StartPickTask) Execute(ctx context.Context, input PickTaskActionInput) (PickTaskResult, error) {
	if err := ensurePickTaskActionReady(uc.store, uc.auditLog, input.ActorID); err != nil {
		return PickTaskResult{}, err
	}
	current, err := uc.store.GetPickTask(ctx, input.PickTaskID)
	if err != nil {
		return PickTaskResult{}, err
	}
	if current.Status == domain.PickTaskStatusInProgress || current.Status == domain.PickTaskStatusCompleted {
		return PickTaskResult{PickTask: current}, nil
	}

	now := uc.clock()
	updated := current
	if updated.Status == domain.PickTaskStatusCreated {
		updated, err = updated.Assign(input.ActorID, now)
		if err != nil {
			return PickTaskResult{}, err
		}
	}
	updated, err = updated.Start(input.ActorID, now)
	if err != nil {
		return PickTaskResult{}, err
	}

	return saveAndAuditPickTaskAction(ctx, uc.store, uc.auditLog, input.ActorID, input.RequestID, "shipping.pick_task.started", current, updated, nil, now)
}

func (uc ConfirmPickTaskLine) Execute(ctx context.Context, input ConfirmPickTaskLineInput) (PickTaskResult, error) {
	if err := ensurePickTaskActionReady(uc.store, uc.auditLog, input.ActorID); err != nil {
		return PickTaskResult{}, err
	}
	current, err := uc.store.GetPickTask(ctx, input.PickTaskID)
	if err != nil {
		return PickTaskResult{}, err
	}
	if lineAlreadyPicked(current, input.LineID, input.PickedQty) {
		return PickTaskResult{PickTask: current}, nil
	}

	now := uc.clock()
	updated, err := current.MarkLinePicked(input.LineID, input.PickedQty, input.ActorID, now)
	if err != nil {
		return PickTaskResult{}, err
	}

	return saveAndAuditPickTaskAction(ctx, uc.store, uc.auditLog, input.ActorID, input.RequestID, "shipping.pick_task.line_confirmed", current, updated, map[string]any{"line_id": strings.TrimSpace(input.LineID)}, now)
}

func (uc CompletePickTask) Execute(ctx context.Context, input PickTaskActionInput) (PickTaskResult, error) {
	if err := ensurePickTaskActionReady(uc.store, uc.auditLog, input.ActorID); err != nil {
		return PickTaskResult{}, err
	}
	current, err := uc.store.GetPickTask(ctx, input.PickTaskID)
	if err != nil {
		return PickTaskResult{}, err
	}
	if current.Status == domain.PickTaskStatusCompleted {
		return PickTaskResult{PickTask: current}, nil
	}

	now := uc.clock()
	updated, err := current.Complete(input.ActorID, now)
	if err != nil {
		return PickTaskResult{}, err
	}

	return saveAndAuditPickTaskAction(ctx, uc.store, uc.auditLog, input.ActorID, input.RequestID, "shipping.pick_task.completed", current, updated, nil, now)
}

func (uc ReportPickTaskException) Execute(
	ctx context.Context,
	input ReportPickTaskExceptionInput,
) (PickTaskResult, error) {
	if err := ensurePickTaskActionReady(uc.store, uc.auditLog, input.ActorID); err != nil {
		return PickTaskResult{}, err
	}
	current, err := uc.store.GetPickTask(ctx, input.PickTaskID)
	if err != nil {
		return PickTaskResult{}, err
	}
	exceptionStatus := normalizePickTaskExceptionStatus(input.ExceptionCode)
	if exceptionStatus == "" {
		return PickTaskResult{}, domain.ErrPickTaskInvalidStatus
	}
	lineStatus := normalizePickTaskExceptionLineStatus(input.ExceptionCode)
	if strings.TrimSpace(input.LineID) != "" && lineStatus == "" {
		return PickTaskResult{}, domain.ErrPickTaskInvalidStatus
	}
	if current.Status == exceptionStatus && (strings.TrimSpace(input.LineID) == "" || lineAlreadyInException(current, input.LineID, lineStatus)) {
		return PickTaskResult{PickTask: current}, nil
	}
	if current.Status == domain.PickTaskStatusCompleted || (isPickTaskExceptionStatus(current.Status) && current.Status != exceptionStatus) {
		return PickTaskResult{}, domain.ErrPickTaskInvalidTransition
	}

	now := uc.clock()
	updated := domain.PickTask{}
	if strings.TrimSpace(input.LineID) == "" {
		updated = current.Clone()
		updated.Status = exceptionStatus
		updated.UpdatedAt = now
		if err := updated.Validate(); err != nil {
			return PickTaskResult{}, err
		}
	} else {
		updated, err = current.ReportLineException(input.LineID, lineStatus, input.ActorID, now)
		if err != nil {
			return PickTaskResult{}, err
		}
	}
	metadata := map[string]any{
		"exception_code": string(exceptionStatus),
		"investigation":  strings.TrimSpace(input.Investigation),
	}
	if strings.TrimSpace(input.LineID) != "" {
		metadata["line_id"] = strings.TrimSpace(input.LineID)
	}

	return saveAndAuditPickTaskAction(
		ctx,
		uc.store,
		uc.auditLog,
		input.ActorID,
		input.RequestID,
		"shipping.pick_task.exception_reported",
		current,
		updated,
		metadata,
		now,
	)
}

type PrototypePickTaskStore struct {
	mu           sync.RWMutex
	records      map[string]domain.PickTask
	bySalesOrder map[string]string
}

func NewPrototypePickTaskStore(tasks ...domain.PickTask) *PrototypePickTaskStore {
	store := &PrototypePickTaskStore{
		records:      make(map[string]domain.PickTask),
		bySalesOrder: make(map[string]string),
	}
	for _, task := range tasks {
		_ = store.SavePickTask(context.Background(), task)
	}

	return store
}

func (s *PrototypePickTaskStore) GetPickTask(_ context.Context, id string) (domain.PickTask, error) {
	if s == nil {
		return domain.PickTask{}, errors.New("pick task store is required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.records[strings.TrimSpace(id)]
	if !ok {
		return domain.PickTask{}, ErrPickTaskNotFound
	}

	return task.Clone(), nil
}

func (s *PrototypePickTaskStore) GetPickTaskBySalesOrder(_ context.Context, salesOrderID string) (domain.PickTask, error) {
	if s == nil {
		return domain.PickTask{}, errors.New("pick task store is required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	taskID, ok := s.bySalesOrder[strings.TrimSpace(salesOrderID)]
	if !ok {
		return domain.PickTask{}, ErrPickTaskNotFound
	}

	return s.records[taskID].Clone(), nil
}

func (s *PrototypePickTaskStore) SavePickTask(_ context.Context, task domain.PickTask) error {
	if s == nil {
		return errors.New("pick task store is required")
	}
	if err := task.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if existingID, ok := s.bySalesOrder[task.SalesOrderID]; ok && existingID != task.ID {
		return ErrPickTaskDuplicate
	}
	s.records[task.ID] = task.Clone()
	s.bySalesOrder[task.SalesOrderID] = task.ID

	return nil
}

func (s *PrototypePickTaskStore) ListPickTasks(_ context.Context) ([]domain.PickTask, error) {
	if s == nil {
		return nil, errors.New("pick task store is required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]domain.PickTask, 0, len(s.records))
	for _, task := range s.records {
		tasks = append(tasks, task.Clone())
	}
	domain.SortPickTasks(tasks)

	return tasks, nil
}

func saveAndAuditPickTaskAction(
	ctx context.Context,
	store PickTaskStore,
	auditLog audit.LogStore,
	actorID string,
	requestID string,
	action string,
	before domain.PickTask,
	after domain.PickTask,
	metadata map[string]any,
	createdAt time.Time,
) (PickTaskResult, error) {
	if err := store.SavePickTask(ctx, after); err != nil {
		return PickTaskResult{}, err
	}
	log, err := newPickTaskActionAuditLog(actorID, requestID, action, before, after, metadata, createdAt)
	if err != nil {
		return PickTaskResult{}, err
	}
	if err := auditLog.Record(ctx, log); err != nil {
		return PickTaskResult{}, err
	}

	return PickTaskResult{PickTask: after, AuditLogID: log.ID}, nil
}

func ensurePickTaskActionReady(store PickTaskStore, auditLog audit.LogStore, actorID string) error {
	if store == nil {
		return errors.New("pick task store is required")
	}
	if auditLog == nil {
		return errors.New("audit log store is required")
	}
	if strings.TrimSpace(actorID) == "" {
		return domain.ErrPickTaskActorRequired
	}

	return nil
}

func newPickTaskLinesFromReservations(
	order salesdomain.SalesOrder,
	reservations []inventorydomain.StockReservation,
) ([]domain.NewPickTaskLineInput, error) {
	activeByLineID := activeReservationsByLineID(order.ID, reservations)
	lines := make([]domain.NewPickTaskLineInput, 0, len(order.Lines))
	for _, orderLine := range order.Lines {
		reservation, ok := activeByLineID[orderLine.ID]
		if !ok {
			return nil, ErrPickTaskReservationMissing
		}
		lines = append(lines, domain.NewPickTaskLineInput{
			ID:                 fmt.Sprintf("%s-line-%02d", newPickTaskID(order.ID), orderLine.LineNo),
			LineNo:             orderLine.LineNo,
			SalesOrderLineID:   orderLine.ID,
			StockReservationID: reservation.ID,
			ItemID:             reservation.ItemID,
			SKUCode:            reservation.SKUCode,
			BatchID:            reservation.BatchID,
			BatchNo:            reservation.BatchNo,
			WarehouseID:        reservation.WarehouseID,
			BinID:              reservation.BinID,
			BinCode:            reservation.BinCode,
			QtyToPick:          reservation.ReservedQty.String(),
			BaseUOMCode:        reservation.BaseUOMCode.String(),
		})
	}

	return lines, nil
}

func activeReservationsByLineID(
	salesOrderID string,
	reservations []inventorydomain.StockReservation,
) map[string]inventorydomain.StockReservation {
	rows := make([]inventorydomain.StockReservation, 0, len(reservations))
	for _, reservation := range reservations {
		if strings.TrimSpace(reservation.SalesOrderID) == strings.TrimSpace(salesOrderID) &&
			reservation.IsActive() {
			rows = append(rows, reservation)
		}
	}
	sort.SliceStable(rows, func(i int, j int) bool {
		return rows[i].ReservedAt.Before(rows[j].ReservedAt)
	})

	byLineID := make(map[string]inventorydomain.StockReservation, len(rows))
	for _, reservation := range rows {
		byLineID[reservation.SalesOrderLineID] = reservation
	}

	return byLineID
}

func lineAlreadyPicked(task domain.PickTask, lineID string, pickedQty string) bool {
	qty := strings.TrimSpace(pickedQty)
	for _, line := range task.Lines {
		if strings.TrimSpace(line.ID) == strings.TrimSpace(lineID) &&
			line.Status == domain.PickTaskLineStatusPicked &&
			line.QtyPicked.String() == normalizeQuantityText(qty) {
			return true
		}
	}

	return false
}

func normalizePickTaskExceptionStatus(value string) domain.PickTaskStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(domain.PickTaskStatusMissingStock):
		return domain.PickTaskStatusMissingStock
	case string(domain.PickTaskStatusWrongSKU):
		return domain.PickTaskStatusWrongSKU
	case string(domain.PickTaskStatusWrongBatch):
		return domain.PickTaskStatusWrongBatch
	case string(domain.PickTaskStatusWrongLocation):
		return domain.PickTaskStatusWrongLocation
	case string(domain.PickTaskStatusCancelled):
		return domain.PickTaskStatusCancelled
	default:
		return ""
	}
}

func normalizePickTaskExceptionLineStatus(value string) domain.PickTaskLineStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(domain.PickTaskLineStatusMissingStock):
		return domain.PickTaskLineStatusMissingStock
	case string(domain.PickTaskLineStatusWrongSKU):
		return domain.PickTaskLineStatusWrongSKU
	case string(domain.PickTaskLineStatusWrongBatch):
		return domain.PickTaskLineStatusWrongBatch
	case string(domain.PickTaskLineStatusWrongLocation):
		return domain.PickTaskLineStatusWrongLocation
	case string(domain.PickTaskLineStatusCancelled):
		return domain.PickTaskLineStatusCancelled
	default:
		return ""
	}
}

func isPickTaskExceptionStatus(status domain.PickTaskStatus) bool {
	switch status {
	case domain.PickTaskStatusMissingStock,
		domain.PickTaskStatusWrongSKU,
		domain.PickTaskStatusWrongBatch,
		domain.PickTaskStatusWrongLocation,
		domain.PickTaskStatusCancelled:
		return true
	default:
		return false
	}
}

func lineAlreadyInException(task domain.PickTask, lineID string, status domain.PickTaskLineStatus) bool {
	for _, line := range task.Lines {
		if strings.TrimSpace(line.ID) == strings.TrimSpace(lineID) && line.Status == status {
			return true
		}
	}

	return false
}

func normalizeQuantityText(value string) string {
	quantity, err := decimal.ParseQuantity(value)
	if err != nil {
		return strings.TrimSpace(value)
	}

	return quantity.String()
}

func newPickTaskAuditLog(
	actorID string,
	requestID string,
	action string,
	task domain.PickTask,
	createdAt time.Time,
) (audit.Log, error) {
	return audit.NewLog(audit.NewLogInput{
		OrgID:      task.OrgID,
		ActorID:    strings.TrimSpace(actorID),
		Action:     action,
		EntityType: "shipping.pick_task",
		EntityID:   task.ID,
		RequestID:  strings.TrimSpace(requestID),
		AfterData: map[string]any{
			"pick_task_no":   task.PickTaskNo,
			"sales_order_id": task.SalesOrderID,
			"order_no":       task.OrderNo,
			"warehouse_id":   task.WarehouseID,
			"status":         string(task.Status),
			"line_count":     len(task.Lines),
			"assigned_to":    task.AssignedTo,
		},
		Metadata: map[string]any{
			"source": "reserved sales order",
		},
		CreatedAt: createdAt,
	})
}

func newPickTaskActionAuditLog(
	actorID string,
	requestID string,
	action string,
	before domain.PickTask,
	after domain.PickTask,
	metadata map[string]any,
	createdAt time.Time,
) (audit.Log, error) {
	baseMetadata := map[string]any{"source": "pick task action"}
	for key, value := range metadata {
		baseMetadata[key] = value
	}

	return audit.NewLog(audit.NewLogInput{
		OrgID:      after.OrgID,
		ActorID:    strings.TrimSpace(actorID),
		Action:     action,
		EntityType: "shipping.pick_task",
		EntityID:   after.ID,
		RequestID:  strings.TrimSpace(requestID),
		BeforeData: pickTaskAuditData(before),
		AfterData:  pickTaskAuditData(after),
		Metadata:   baseMetadata,
		CreatedAt:  createdAt,
	})
}

func pickTaskAuditData(task domain.PickTask) map[string]any {
	return map[string]any{
		"pick_task_no":   task.PickTaskNo,
		"sales_order_id": task.SalesOrderID,
		"warehouse_id":   task.WarehouseID,
		"status":         string(task.Status),
		"assigned_to":    task.AssignedTo,
		"line_count":     len(task.Lines),
		"picked_count":   countPickedLines(task),
	}
}

func countPickedLines(task domain.PickTask) int {
	count := 0
	for _, line := range task.Lines {
		if line.Status == domain.PickTaskLineStatusPicked {
			count++
		}
	}

	return count
}

func newPickTaskID(salesOrderID string) string {
	return fmt.Sprintf("pick-%s", strings.TrimSpace(salesOrderID))
}

func newPickTaskNo(orderNo string) string {
	return fmt.Sprintf("PICK-%s", strings.ToUpper(strings.TrimSpace(orderNo)))
}
