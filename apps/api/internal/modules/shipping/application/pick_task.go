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
)

var ErrPickTaskNotFound = errors.New("pick task not found")
var ErrPickTaskDuplicate = errors.New("pick task already exists")
var ErrPickTaskSalesOrderNotReserved = errors.New("sales order must be reserved before generating pick task")
var ErrPickTaskReservationMissing = errors.New("active stock reservation is required for every sales order line")

type PickTaskStore interface {
	GetPickTask(ctx context.Context, id string) (domain.PickTask, error)
	GetPickTaskBySalesOrder(ctx context.Context, salesOrderID string) (domain.PickTask, error)
	SavePickTask(ctx context.Context, task domain.PickTask) error
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

func (s *PrototypePickTaskStore) ListPickTasks() []domain.PickTask {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]domain.PickTask, 0, len(s.records))
	for _, task := range s.records {
		tasks = append(tasks, task.Clone())
	}
	domain.SortPickTasks(tasks)

	return tasks
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

func newPickTaskID(salesOrderID string) string {
	return fmt.Sprintf("pick-%s", strings.TrimSpace(salesOrderID))
}

func newPickTaskNo(orderNo string) string {
	return fmt.Sprintf("PICK-%s", strings.ToUpper(strings.TrimSpace(orderNo)))
}
