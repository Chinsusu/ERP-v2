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

type PackTaskStore interface {
	ListPackTasks(ctx context.Context) ([]domain.PackTask, error)
	GetPackTask(ctx context.Context, id string) (domain.PackTask, error)
	GetPackTaskBySalesOrder(ctx context.Context, salesOrderID string) (domain.PackTask, error)
	GetPackTaskByPickTask(ctx context.Context, pickTaskID string) (domain.PackTask, error)
	SavePackTask(ctx context.Context, task domain.PackTask) error
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

func NewGeneratePackTaskAfterPick(store PackTaskStore, auditLog audit.LogStore) GeneratePackTaskAfterPick {
	return GeneratePackTaskAfterPick{
		store:    store,
		auditLog: auditLog,
		clock:    func() time.Time { return time.Now().UTC() },
	}
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

func newPackTaskID(salesOrderID string) string {
	return fmt.Sprintf("pack-%s", strings.TrimSpace(salesOrderID))
}

func newPackTaskNo(orderNo string) string {
	return fmt.Sprintf("PACK-%s", strings.ToUpper(strings.TrimSpace(orderNo)))
}
