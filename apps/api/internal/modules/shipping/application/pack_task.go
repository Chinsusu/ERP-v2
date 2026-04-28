package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	salesdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

var ErrPackTaskNotFound = errors.New("pack task not found")
var ErrPackTaskDuplicate = errors.New("pack task already exists")
var ErrPackTaskSalesOrderNotPicked = errors.New("sales order must be picked before generating pack task")
var ErrPackTaskPickTaskNotCompleted = errors.New("pick task must be completed before generating pack task")
var ErrPackTaskPickTaskLineNotPicked = errors.New("every pick task line must be picked before generating pack task")
var ErrPackTaskSalesOrderPackerRequired = errors.New("sales order packer is required")

type PackTaskStore interface {
	ListPackTasks(ctx context.Context) ([]domain.PackTask, error)
	GetPackTask(ctx context.Context, id string) (domain.PackTask, error)
	GetPackTaskBySalesOrder(ctx context.Context, salesOrderID string) (domain.PackTask, error)
	GetPackTaskByPickTask(ctx context.Context, pickTaskID string) (domain.PackTask, error)
	SavePackTask(ctx context.Context, task domain.PackTask) error
}

type PackTaskSalesOrderPacker interface {
	MarkSalesOrderPacked(ctx context.Context, input PackTaskSalesOrderPackedInput) (salesdomain.SalesOrder, error)
}

type PackTaskSalesOrderPackedInput struct {
	SalesOrderID string
	ActorID      string
	RequestID    string
}

type PackTaskFilter struct {
	WarehouseID string
	Status      domain.PackTaskStatus
	AssignedTo  string
}

type ListPackTasks struct {
	store PackTaskStore
}

type GetPackTask struct {
	store PackTaskStore
}

type GeneratePackTaskAfterPick struct {
	store    PackTaskStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type GeneratePackTaskAfterPickInput struct {
	SalesOrder salesdomain.SalesOrder
	PickTask   domain.PickTask
	AssignedTo string
	ActorID    string
	RequestID  string
}

type PackTaskResult struct {
	PackTask   domain.PackTask
	SalesOrder salesdomain.SalesOrder
	AuditLogID string
}

type PackTaskActionInput struct {
	PackTaskID string
	ActorID    string
	RequestID  string
}

type ConfirmPackTaskInput struct {
	PackTaskID string
	Lines      []ConfirmPackTaskLineInput
	ActorID    string
	RequestID  string
}

type ConfirmPackTaskLineInput struct {
	LineID    string
	PackedQty string
}

type ReportPackTaskExceptionInput struct {
	PackTaskID    string
	LineID        string
	ExceptionCode string
	ActorID       string
	RequestID     string
	Investigation string
}

type StartPackTask struct {
	store    PackTaskStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type ConfirmPackTask struct {
	store       PackTaskStore
	auditLog    audit.LogStore
	salesOrders PackTaskSalesOrderPacker
	clock       func() time.Time
}

type ReportPackTaskException struct {
	store    PackTaskStore
	auditLog audit.LogStore
	clock    func() time.Time
}

func NewGeneratePackTaskAfterPick(store PackTaskStore, auditLog audit.LogStore) GeneratePackTaskAfterPick {
	return GeneratePackTaskAfterPick{
		store:    store,
		auditLog: auditLog,
		clock:    func() time.Time { return time.Now().UTC() },
	}
}

func NewListPackTasks(store PackTaskStore) ListPackTasks {
	return ListPackTasks{store: store}
}

func NewGetPackTask(store PackTaskStore) GetPackTask {
	return GetPackTask{store: store}
}

func NewStartPackTask(store PackTaskStore, auditLog audit.LogStore) StartPackTask {
	return StartPackTask{store: store, auditLog: auditLog, clock: func() time.Time { return time.Now().UTC() }}
}

func NewConfirmPackTask(
	store PackTaskStore,
	auditLog audit.LogStore,
	salesOrders PackTaskSalesOrderPacker,
) ConfirmPackTask {
	return ConfirmPackTask{
		store:       store,
		auditLog:    auditLog,
		salesOrders: salesOrders,
		clock:       func() time.Time { return time.Now().UTC() },
	}
}

func NewReportPackTaskException(store PackTaskStore, auditLog audit.LogStore) ReportPackTaskException {
	return ReportPackTaskException{store: store, auditLog: auditLog, clock: func() time.Time { return time.Now().UTC() }}
}

func (uc ListPackTasks) Execute(ctx context.Context, filter PackTaskFilter) ([]domain.PackTask, error) {
	if uc.store == nil {
		return nil, errors.New("pack task store is required")
	}
	tasks, err := uc.store.ListPackTasks(ctx)
	if err != nil {
		return nil, err
	}

	filter.WarehouseID = strings.TrimSpace(filter.WarehouseID)
	filter.AssignedTo = strings.TrimSpace(filter.AssignedTo)
	filter.Status = domain.NormalizePackTaskStatus(filter.Status)
	if filter.WarehouseID == "" && filter.AssignedTo == "" && filter.Status == "" {
		return tasks, nil
	}

	filtered := make([]domain.PackTask, 0, len(tasks))
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

func (uc GetPackTask) Execute(ctx context.Context, id string) (domain.PackTask, error) {
	if uc.store == nil {
		return domain.PackTask{}, errors.New("pack task store is required")
	}

	return uc.store.GetPackTask(ctx, id)
}

func (uc GeneratePackTaskAfterPick) Execute(
	ctx context.Context,
	input GeneratePackTaskAfterPickInput,
) (PackTaskResult, error) {
	if uc.store == nil {
		return PackTaskResult{}, errors.New("pack task store is required")
	}
	if uc.auditLog == nil {
		return PackTaskResult{}, errors.New("audit log store is required")
	}
	if strings.TrimSpace(input.ActorID) == "" {
		return PackTaskResult{}, domain.ErrPackTaskActorRequired
	}
	order := input.SalesOrder.Clone()
	if salesdomain.NormalizeSalesOrderStatus(order.Status) != salesdomain.SalesOrderStatusPicked {
		return PackTaskResult{}, ErrPackTaskSalesOrderNotPicked
	}
	pickTask := input.PickTask.Clone()
	if pickTask.Status != domain.PickTaskStatusCompleted || strings.TrimSpace(pickTask.SalesOrderID) != strings.TrimSpace(order.ID) {
		return PackTaskResult{}, ErrPackTaskPickTaskNotCompleted
	}
	if _, err := uc.store.GetPackTaskBySalesOrder(ctx, order.ID); err == nil {
		return PackTaskResult{}, ErrPackTaskDuplicate
	} else if !errors.Is(err, ErrPackTaskNotFound) {
		return PackTaskResult{}, err
	}
	if _, err := uc.store.GetPackTaskByPickTask(ctx, pickTask.ID); err == nil {
		return PackTaskResult{}, ErrPackTaskDuplicate
	} else if !errors.Is(err, ErrPackTaskNotFound) {
		return PackTaskResult{}, err
	}

	now := uc.clock()
	lines, err := newPackTaskLinesFromPickTask(order.ID, pickTask)
	if err != nil {
		return PackTaskResult{}, err
	}
	task, err := domain.NewPackTask(domain.NewPackTaskInput{
		ID:            newPackTaskID(order.ID),
		OrgID:         order.OrgID,
		PackTaskNo:    newPackTaskNo(order.OrderNo),
		SalesOrderID:  order.ID,
		OrderNo:       order.OrderNo,
		PickTaskID:    pickTask.ID,
		PickTaskNo:    pickTask.PickTaskNo,
		WarehouseID:   pickTask.WarehouseID,
		WarehouseCode: pickTask.WarehouseCode,
		AssignedTo:    input.AssignedTo,
		AssignedAt:    assignedAt(input.AssignedTo, now),
		CreatedAt:     now,
		Lines:         lines,
	})
	if err != nil {
		return PackTaskResult{}, err
	}
	updatedOrder, err := order.StartPacking(input.ActorID, now)
	if err != nil {
		return PackTaskResult{}, err
	}
	if err := uc.store.SavePackTask(ctx, task); err != nil {
		return PackTaskResult{}, err
	}
	log, err := newPackTaskAuditLog(input.ActorID, input.RequestID, "shipping.pack_task.created", task, updatedOrder, now)
	if err != nil {
		return PackTaskResult{}, err
	}
	if err := uc.auditLog.Record(ctx, log); err != nil {
		return PackTaskResult{}, err
	}

	return PackTaskResult{PackTask: task, SalesOrder: updatedOrder, AuditLogID: log.ID}, nil
}

func (uc StartPackTask) Execute(ctx context.Context, input PackTaskActionInput) (PackTaskResult, error) {
	if err := ensurePackTaskActionReady(uc.store, uc.auditLog, input.ActorID); err != nil {
		return PackTaskResult{}, err
	}
	current, err := uc.store.GetPackTask(ctx, input.PackTaskID)
	if err != nil {
		return PackTaskResult{}, err
	}
	if current.Status == domain.PackTaskStatusInProgress || current.Status == domain.PackTaskStatusPacked {
		return PackTaskResult{PackTask: current}, nil
	}

	now := uc.clock()
	updated, err := current.Start(input.ActorID, now)
	if err != nil {
		return PackTaskResult{}, err
	}

	return saveAndAuditPackTaskAction(ctx, uc.store, uc.auditLog, input.ActorID, input.RequestID, "shipping.pack_task.started", current, updated, salesdomain.SalesOrder{}, nil, now)
}

func (uc ConfirmPackTask) Execute(ctx context.Context, input ConfirmPackTaskInput) (PackTaskResult, error) {
	if err := ensurePackTaskActionReady(uc.store, uc.auditLog, input.ActorID); err != nil {
		return PackTaskResult{}, err
	}
	if uc.salesOrders == nil {
		return PackTaskResult{}, ErrPackTaskSalesOrderPackerRequired
	}
	current, err := uc.store.GetPackTask(ctx, input.PackTaskID)
	if err != nil {
		return PackTaskResult{}, err
	}
	if current.Status == domain.PackTaskStatusPacked {
		return PackTaskResult{PackTask: current}, nil
	}

	now := uc.clock()
	updated := current.Clone()
	lines, err := packTaskConfirmLines(current, input.Lines)
	if err != nil {
		return PackTaskResult{}, err
	}
	for _, line := range lines {
		if lineAlreadyPacked(updated, line.LineID, line.PackedQty) {
			continue
		}
		updated, err = updated.MarkLinePacked(line.LineID, line.PackedQty, input.ActorID, now)
		if err != nil {
			return PackTaskResult{}, err
		}
	}
	updated, err = updated.Complete(input.ActorID, now)
	if err != nil {
		return PackTaskResult{}, err
	}
	packedOrder, err := uc.salesOrders.MarkSalesOrderPacked(ctx, PackTaskSalesOrderPackedInput{
		SalesOrderID: current.SalesOrderID,
		ActorID:      input.ActorID,
		RequestID:    input.RequestID,
	})
	if err != nil {
		return PackTaskResult{}, err
	}

	return saveAndAuditPackTaskAction(ctx, uc.store, uc.auditLog, input.ActorID, input.RequestID, "shipping.pack_task.confirmed", current, updated, packedOrder, nil, now)
}

func (uc ReportPackTaskException) Execute(
	ctx context.Context,
	input ReportPackTaskExceptionInput,
) (PackTaskResult, error) {
	if err := ensurePackTaskActionReady(uc.store, uc.auditLog, input.ActorID); err != nil {
		return PackTaskResult{}, err
	}
	current, err := uc.store.GetPackTask(ctx, input.PackTaskID)
	if err != nil {
		return PackTaskResult{}, err
	}
	exceptionCode := normalizePackTaskExceptionCode(input.ExceptionCode)
	exceptionStatus := normalizePackTaskExceptionStatus(input.ExceptionCode)
	if exceptionStatus == "" {
		return PackTaskResult{}, domain.ErrPackTaskInvalidStatus
	}
	if current.Status == domain.PackTaskStatusPackException && (strings.TrimSpace(input.LineID) == "" || packLineAlreadyInException(current, input.LineID)) {
		return PackTaskResult{PackTask: current}, nil
	}
	if current.Status == domain.PackTaskStatusPacked || current.Status == domain.PackTaskStatusCancelled {
		return PackTaskResult{}, domain.ErrPackTaskInvalidTransition
	}

	now := uc.clock()
	updated := domain.PackTask{}
	if strings.TrimSpace(input.LineID) == "" {
		updated = current.Clone()
		updated.Status = domain.PackTaskStatusPackException
		updated.UpdatedAt = now
		if err := updated.Validate(); err != nil {
			return PackTaskResult{}, err
		}
	} else {
		updated, err = current.ReportLineException(input.LineID, input.ActorID, now)
		if err != nil {
			return PackTaskResult{}, err
		}
	}
	metadata := map[string]any{
		"exception_code":   exceptionCode,
		"exception_status": string(exceptionStatus),
		"investigation":    strings.TrimSpace(input.Investigation),
	}
	if strings.TrimSpace(input.LineID) != "" {
		metadata["line_id"] = strings.TrimSpace(input.LineID)
	}

	return saveAndAuditPackTaskAction(
		ctx,
		uc.store,
		uc.auditLog,
		input.ActorID,
		input.RequestID,
		"shipping.pack_task.exception_reported",
		current,
		updated,
		salesdomain.SalesOrder{},
		metadata,
		now,
	)
}

type PrototypePackTaskStore struct {
	mu           sync.RWMutex
	records      map[string]domain.PackTask
	bySalesOrder map[string]string
	byPickTask   map[string]string
}

func NewPrototypePackTaskStore(tasks ...domain.PackTask) *PrototypePackTaskStore {
	store := &PrototypePackTaskStore{
		records:      make(map[string]domain.PackTask),
		bySalesOrder: make(map[string]string),
		byPickTask:   make(map[string]string),
	}
	for _, task := range tasks {
		_ = store.SavePackTask(context.Background(), task)
	}

	return store
}

func (s *PrototypePackTaskStore) GetPackTask(_ context.Context, id string) (domain.PackTask, error) {
	if s == nil {
		return domain.PackTask{}, errors.New("pack task store is required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.records[strings.TrimSpace(id)]
	if !ok {
		return domain.PackTask{}, ErrPackTaskNotFound
	}

	return task.Clone(), nil
}

func (s *PrototypePackTaskStore) GetPackTaskBySalesOrder(_ context.Context, salesOrderID string) (domain.PackTask, error) {
	if s == nil {
		return domain.PackTask{}, errors.New("pack task store is required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	taskID, ok := s.bySalesOrder[strings.TrimSpace(salesOrderID)]
	if !ok {
		return domain.PackTask{}, ErrPackTaskNotFound
	}

	return s.records[taskID].Clone(), nil
}

func (s *PrototypePackTaskStore) GetPackTaskByPickTask(_ context.Context, pickTaskID string) (domain.PackTask, error) {
	if s == nil {
		return domain.PackTask{}, errors.New("pack task store is required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	taskID, ok := s.byPickTask[strings.TrimSpace(pickTaskID)]
	if !ok {
		return domain.PackTask{}, ErrPackTaskNotFound
	}

	return s.records[taskID].Clone(), nil
}

func (s *PrototypePackTaskStore) SavePackTask(_ context.Context, task domain.PackTask) error {
	if s == nil {
		return errors.New("pack task store is required")
	}
	if err := task.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if existingID, ok := s.bySalesOrder[task.SalesOrderID]; ok && existingID != task.ID {
		return ErrPackTaskDuplicate
	}
	if existingID, ok := s.byPickTask[task.PickTaskID]; ok && existingID != task.ID {
		return ErrPackTaskDuplicate
	}
	s.records[task.ID] = task.Clone()
	s.bySalesOrder[task.SalesOrderID] = task.ID
	s.byPickTask[task.PickTaskID] = task.ID

	return nil
}

func (s *PrototypePackTaskStore) ListPackTasks(_ context.Context) ([]domain.PackTask, error) {
	if s == nil {
		return nil, errors.New("pack task store is required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]domain.PackTask, 0, len(s.records))
	for _, task := range s.records {
		tasks = append(tasks, task.Clone())
	}
	domain.SortPackTasks(tasks)

	return tasks, nil
}

func saveAndAuditPackTaskAction(
	ctx context.Context,
	store PackTaskStore,
	auditLog audit.LogStore,
	actorID string,
	requestID string,
	action string,
	before domain.PackTask,
	after domain.PackTask,
	salesOrder salesdomain.SalesOrder,
	metadata map[string]any,
	createdAt time.Time,
) (PackTaskResult, error) {
	if err := store.SavePackTask(ctx, after); err != nil {
		return PackTaskResult{}, err
	}
	log, err := newPackTaskActionAuditLog(actorID, requestID, action, before, after, salesOrder, metadata, createdAt)
	if err != nil {
		return PackTaskResult{}, err
	}
	if err := auditLog.Record(ctx, log); err != nil {
		return PackTaskResult{}, err
	}

	return PackTaskResult{PackTask: after, SalesOrder: salesOrder, AuditLogID: log.ID}, nil
}

func ensurePackTaskActionReady(store PackTaskStore, auditLog audit.LogStore, actorID string) error {
	if store == nil {
		return errors.New("pack task store is required")
	}
	if auditLog == nil {
		return errors.New("audit log store is required")
	}
	if strings.TrimSpace(actorID) == "" {
		return domain.ErrPackTaskActorRequired
	}

	return nil
}

func packTaskConfirmLines(task domain.PackTask, input []ConfirmPackTaskLineInput) ([]ConfirmPackTaskLineInput, error) {
	if len(input) == 0 {
		lines := make([]ConfirmPackTaskLineInput, 0, len(task.Lines))
		for _, line := range task.Lines {
			lines = append(lines, ConfirmPackTaskLineInput{
				LineID:    line.ID,
				PackedQty: line.QtyToPack.String(),
			})
		}

		return lines, nil
	}

	byLineID := make(map[string]ConfirmPackTaskLineInput, len(input))
	for _, line := range input {
		line.LineID = strings.TrimSpace(line.LineID)
		line.PackedQty = strings.TrimSpace(line.PackedQty)
		if line.LineID == "" || line.PackedQty == "" {
			return nil, domain.ErrPackTaskRequiredField
		}
		if _, ok := byLineID[line.LineID]; ok {
			return nil, domain.ErrPackTaskDuplicateLine
		}
		byLineID[line.LineID] = line
	}
	lines := make([]ConfirmPackTaskLineInput, 0, len(task.Lines))
	for _, taskLine := range task.Lines {
		line, ok := byLineID[taskLine.ID]
		if !ok {
			return nil, domain.ErrPackTaskInvalidTransition
		}
		lines = append(lines, line)
		delete(byLineID, taskLine.ID)
	}
	if len(byLineID) > 0 {
		return nil, domain.ErrPackTaskRequiredField
	}

	return lines, nil
}

func newPackTaskLinesFromPickTask(salesOrderID string, pickTask domain.PickTask) ([]domain.NewPackTaskLineInput, error) {
	lines := make([]domain.NewPackTaskLineInput, 0, len(pickTask.Lines))
	for _, pickLine := range pickTask.Lines {
		if pickLine.Status != domain.PickTaskLineStatusPicked {
			return nil, ErrPackTaskPickTaskLineNotPicked
		}
		lines = append(lines, domain.NewPackTaskLineInput{
			ID:               fmt.Sprintf("%s-line-%02d", newPackTaskID(salesOrderID), pickLine.LineNo),
			LineNo:           pickLine.LineNo,
			PickTaskLineID:   pickLine.ID,
			SalesOrderLineID: pickLine.SalesOrderLineID,
			ItemID:           pickLine.ItemID,
			SKUCode:          pickLine.SKUCode,
			BatchID:          pickLine.BatchID,
			BatchNo:          pickLine.BatchNo,
			WarehouseID:      pickLine.WarehouseID,
			QtyToPack:        pickLine.QtyPicked.String(),
			BaseUOMCode:      pickLine.BaseUOMCode.String(),
		})
	}

	return lines, nil
}

func lineAlreadyPacked(task domain.PackTask, lineID string, packedQty string) bool {
	qty := normalizeQuantityText(packedQty)
	for _, line := range task.Lines {
		if strings.TrimSpace(line.ID) == strings.TrimSpace(lineID) &&
			line.Status == domain.PackTaskLineStatusPacked &&
			line.QtyPacked.String() == qty {
			return true
		}
	}

	return false
}

func normalizePackTaskExceptionStatus(value string) domain.PackTaskStatus {
	if normalizePackTaskExceptionCode(value) != "" {
		return domain.PackTaskStatusPackException
	}

	return ""
}

func normalizePackTaskExceptionCode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(domain.PackTaskStatusPackException):
		return string(domain.PackTaskStatusPackException)
	case "missing_stock", "missing_item", "short_pack", "shortage":
		return "missing_stock"
	case "wrong_sku":
		return "wrong_sku"
	case "wrong_batch":
		return "wrong_batch"
	default:
		return ""
	}
}

func packLineAlreadyInException(task domain.PackTask, lineID string) bool {
	for _, line := range task.Lines {
		if strings.TrimSpace(line.ID) == strings.TrimSpace(lineID) && line.Status == domain.PackTaskLineStatusPackException {
			return true
		}
	}

	return false
}

func assignedAt(assignedTo string, at time.Time) time.Time {
	if strings.TrimSpace(assignedTo) == "" {
		return time.Time{}
	}

	return at
}

func newPackTaskAuditLog(
	actorID string,
	requestID string,
	action string,
	task domain.PackTask,
	order salesdomain.SalesOrder,
	createdAt time.Time,
) (audit.Log, error) {
	return audit.NewLog(audit.NewLogInput{
		OrgID:      task.OrgID,
		ActorID:    strings.TrimSpace(actorID),
		Action:     action,
		EntityType: "shipping.pack_task",
		EntityID:   task.ID,
		RequestID:  strings.TrimSpace(requestID),
		AfterData: map[string]any{
			"pack_task_no":      task.PackTaskNo,
			"sales_order_id":    task.SalesOrderID,
			"order_no":          task.OrderNo,
			"sales_order_state": string(order.Status),
			"pick_task_id":      task.PickTaskID,
			"warehouse_id":      task.WarehouseID,
			"status":            string(task.Status),
			"line_count":        len(task.Lines),
			"assigned_to":       task.AssignedTo,
		},
		Metadata: map[string]any{
			"source": "completed pick task",
		},
		CreatedAt: createdAt,
	})
}

func newPackTaskActionAuditLog(
	actorID string,
	requestID string,
	action string,
	before domain.PackTask,
	after domain.PackTask,
	salesOrder salesdomain.SalesOrder,
	metadata map[string]any,
	createdAt time.Time,
) (audit.Log, error) {
	baseMetadata := map[string]any{"source": "pack task action"}
	for key, value := range metadata {
		baseMetadata[key] = value
	}
	afterData := packTaskAuditData(after)
	if strings.TrimSpace(salesOrder.ID) != "" {
		afterData["sales_order_state"] = string(salesOrder.Status)
	}

	return audit.NewLog(audit.NewLogInput{
		OrgID:      after.OrgID,
		ActorID:    strings.TrimSpace(actorID),
		Action:     action,
		EntityType: "shipping.pack_task",
		EntityID:   after.ID,
		RequestID:  strings.TrimSpace(requestID),
		BeforeData: packTaskAuditData(before),
		AfterData:  afterData,
		Metadata:   baseMetadata,
		CreatedAt:  createdAt,
	})
}

func packTaskAuditData(task domain.PackTask) map[string]any {
	return map[string]any{
		"pack_task_no":   task.PackTaskNo,
		"sales_order_id": task.SalesOrderID,
		"pick_task_id":   task.PickTaskID,
		"warehouse_id":   task.WarehouseID,
		"status":         string(task.Status),
		"assigned_to":    task.AssignedTo,
		"line_count":     len(task.Lines),
		"packed_count":   countPackedLines(task),
	}
}

func countPackedLines(task domain.PackTask) int {
	count := 0
	for _, line := range task.Lines {
		if line.Status == domain.PackTaskLineStatusPacked {
			count++
		}
	}

	return count
}

func newPackTaskID(salesOrderID string) string {
	return fmt.Sprintf("pack-%s", strings.TrimSpace(salesOrderID))
}

func newPackTaskNo(orderNo string) string {
	return fmt.Sprintf("PACK-%s", strings.ToUpper(strings.TrimSpace(orderNo)))
}
