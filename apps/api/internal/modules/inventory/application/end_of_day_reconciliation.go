package application

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

var ErrEndOfDayReconciliationNotFound = errors.New("end-of-day reconciliation not found")

type EndOfDayReconciliationStore interface {
	List(ctx context.Context, filter domain.EndOfDayReconciliationFilter) ([]domain.EndOfDayReconciliation, error)
	Get(ctx context.Context, id string) (domain.EndOfDayReconciliation, error)
	Save(ctx context.Context, reconciliation domain.EndOfDayReconciliation) error
}

type ListEndOfDayReconciliations struct {
	store EndOfDayReconciliationStore
}

type CloseEndOfDayReconciliation struct {
	store    EndOfDayReconciliationStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type CloseEndOfDayReconciliationInput struct {
	ID            string
	ActorID       string
	RequestID     string
	ExceptionNote string
}

type CloseEndOfDayReconciliationResult struct {
	Reconciliation domain.EndOfDayReconciliation
	AuditLogID     string
}

func NewListEndOfDayReconciliations(store EndOfDayReconciliationStore) ListEndOfDayReconciliations {
	return ListEndOfDayReconciliations{store: store}
}

func NewCloseEndOfDayReconciliation(
	store EndOfDayReconciliationStore,
	auditLog audit.LogStore,
) CloseEndOfDayReconciliation {
	return CloseEndOfDayReconciliation{
		store:    store,
		auditLog: auditLog,
		clock:    func() time.Time { return time.Now().UTC() },
	}
}

func (uc ListEndOfDayReconciliations) Execute(
	ctx context.Context,
	filter domain.EndOfDayReconciliationFilter,
) ([]domain.EndOfDayReconciliation, error) {
	if uc.store == nil {
		return nil, errors.New("end-of-day reconciliation store is required")
	}

	return uc.store.List(ctx, filter)
}

func (uc CloseEndOfDayReconciliation) Execute(
	ctx context.Context,
	input CloseEndOfDayReconciliationInput,
) (CloseEndOfDayReconciliationResult, error) {
	if uc.store == nil {
		return CloseEndOfDayReconciliationResult{}, errors.New("end-of-day reconciliation store is required")
	}
	if uc.auditLog == nil {
		return CloseEndOfDayReconciliationResult{}, errors.New("audit log store is required")
	}

	reconciliation, err := uc.store.Get(ctx, input.ID)
	if err != nil {
		return CloseEndOfDayReconciliationResult{}, err
	}

	closed, err := reconciliation.Close(input.ActorID, input.ExceptionNote, uc.clock())
	if err != nil {
		return CloseEndOfDayReconciliationResult{}, err
	}
	if err := uc.store.Save(ctx, closed); err != nil {
		return CloseEndOfDayReconciliationResult{}, err
	}

	summary := closed.Summary(input.ExceptionNote)
	log, err := audit.NewLog(audit.NewLogInput{
		OrgID:      "org-my-pham",
		ActorID:    strings.TrimSpace(input.ActorID),
		Action:     "warehouse.shift.closed",
		EntityType: "inventory.warehouse_daily_closing",
		EntityID:   closed.ID,
		RequestID:  strings.TrimSpace(input.RequestID),
		AfterData: map[string]any{
			"status":              string(closed.Status),
			"warehouse_id":        closed.WarehouseID,
			"warehouse_code":      closed.WarehouseCode,
			"date":                closed.Date,
			"shift_code":          closed.ShiftCode,
			"variance_count":      summary.VarianceCount,
			"variance_quantity":   summary.VarianceQuantity,
			"checklist_total":     summary.ChecklistTotal,
			"checklist_complete":  summary.ChecklistCompleted,
			"order_count":         closed.Operations.OrderCount,
			"handover_count":      closed.Operations.HandoverOrderCount,
			"return_count":        closed.Operations.ReturnOrderCount,
			"movement_count":      closed.Operations.StockMovementCount,
			"stock_count_count":   closed.Operations.StockCountSessionCount,
			"pending_issue_count": closed.Operations.PendingIssueCount,
		},
		Metadata: map[string]any{
			"exception_note": strings.TrimSpace(input.ExceptionNote),
			"source":         "end-of-day reconciliation",
		},
		CreatedAt: closed.ClosedAt,
	})
	if err != nil {
		return CloseEndOfDayReconciliationResult{}, err
	}
	if err := uc.auditLog.Record(ctx, log); err != nil {
		return CloseEndOfDayReconciliationResult{}, err
	}

	return CloseEndOfDayReconciliationResult{
		Reconciliation: closed,
		AuditLogID:     log.ID,
	}, nil
}

type PrototypeEndOfDayReconciliationStore struct {
	mu      sync.RWMutex
	records map[string]domain.EndOfDayReconciliation
}

func NewPrototypeEndOfDayReconciliationStore() *PrototypeEndOfDayReconciliationStore {
	store := &PrototypeEndOfDayReconciliationStore{
		records: make(map[string]domain.EndOfDayReconciliation),
	}
	for _, record := range prototypeEndOfDayReconciliations() {
		store.records[record.ID] = record.Clone()
	}

	return store
}

func (s *PrototypeEndOfDayReconciliationStore) List(
	_ context.Context,
	filter domain.EndOfDayReconciliationFilter,
) ([]domain.EndOfDayReconciliation, error) {
	if s == nil {
		return nil, errors.New("end-of-day reconciliation store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows := make([]domain.EndOfDayReconciliation, 0, len(s.records))
	for _, record := range s.records {
		if filter.WarehouseID != "" && record.WarehouseID != filter.WarehouseID {
			continue
		}
		if filter.Date != "" && record.Date != filter.Date {
			continue
		}
		if filter.ShiftCode != "" && record.ShiftCode != filter.ShiftCode {
			continue
		}
		if filter.Status != "" && record.Status != filter.Status {
			continue
		}

		rows = append(rows, record.Clone())
	}
	domain.SortEndOfDayReconciliations(rows)

	return rows, nil
}

func (s *PrototypeEndOfDayReconciliationStore) Get(
	_ context.Context,
	id string,
) (domain.EndOfDayReconciliation, error) {
	if s == nil {
		return domain.EndOfDayReconciliation{}, errors.New("end-of-day reconciliation store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.records[strings.TrimSpace(id)]
	if !ok {
		return domain.EndOfDayReconciliation{}, ErrEndOfDayReconciliationNotFound
	}

	return record.Clone(), nil
}

func (s *PrototypeEndOfDayReconciliationStore) Save(
	_ context.Context,
	reconciliation domain.EndOfDayReconciliation,
) error {
	if s == nil {
		return errors.New("end-of-day reconciliation store is required")
	}
	if strings.TrimSpace(reconciliation.ID) == "" {
		return errors.New("end-of-day reconciliation id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[reconciliation.ID] = reconciliation.Clone()

	return nil
}

func prototypeEndOfDayReconciliations() []domain.EndOfDayReconciliation {
	return []domain.EndOfDayReconciliation{
		{
			ID:            "rec-hcm-260426-day",
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			Date:          "2026-04-26",
			ShiftCode:     "day",
			Status:        domain.ReconciliationStatusInReview,
			Owner:         "Warehouse Lead",
			Operations: domain.ReconciliationOperations{
				OrderCount:             42,
				HandoverOrderCount:     27,
				ReturnOrderCount:       3,
				StockMovementCount:     6,
				StockCountSessionCount: 1,
				PendingIssueCount:      2,
			},
			Checklist: []domain.ReconciliationChecklistItem{
				{Key: "shipments", Label: "Shipments reconciled", Complete: true, Blocking: true},
				{Key: "inbound", Label: "Inbound and QC checked", Complete: true, Blocking: true},
				{Key: "returns", Label: "Returns triaged", Complete: true, Blocking: true},
				{Key: "variance", Label: "Stock variance reviewed", Complete: false, Blocking: true, Note: "LOT-2604A short by 2"},
				{Key: "pending_tasks", Label: "P0 tasks cleared or noted", Complete: false, Blocking: true, Note: "One mismatch pending lead sign-off"},
			},
			Lines: []domain.ReconciliationLine{
				{
					ID:              "line-hcm-serum-2604a",
					SKU:             "SERUM-30ML",
					BatchNo:         "LOT-2604A",
					BinCode:         "A-01",
					SystemQuantity:  120,
					CountedQuantity: 118,
					Reason:          "cycle count variance",
					Owner:           "Warehouse Lead",
				},
				{
					ID:              "line-hcm-cream-2603b",
					SKU:             "CREAM-50G",
					BatchNo:         "LOT-2603B",
					BinCode:         "A-02",
					SystemQuantity:  44,
					CountedQuantity: 44,
					Owner:           "Inventory",
				},
			},
		},
		{
			ID:            "rec-hn-260426-day",
			WarehouseID:   "wh-hn",
			WarehouseCode: "HN",
			Date:          "2026-04-26",
			ShiftCode:     "day",
			Status:        domain.ReconciliationStatusOpen,
			Owner:         "HN Lead",
			Operations: domain.ReconciliationOperations{
				OrderCount:             18,
				HandoverOrderCount:     14,
				ReturnOrderCount:       1,
				StockMovementCount:     3,
				StockCountSessionCount: 1,
				PendingIssueCount:      0,
			},
			Checklist: []domain.ReconciliationChecklistItem{
				{Key: "shipments", Label: "Shipments reconciled", Complete: true, Blocking: true},
				{Key: "returns", Label: "Returns triaged", Complete: true, Blocking: true},
				{Key: "variance", Label: "Stock variance reviewed", Complete: true, Blocking: true},
				{Key: "pending_tasks", Label: "P0 tasks cleared or noted", Complete: true, Blocking: true},
			},
			Lines: []domain.ReconciliationLine{
				{
					ID:              "line-hn-toner-2604c",
					SKU:             "TONER-100ML",
					BatchNo:         "LOT-2604C",
					BinCode:         "HN-B-04",
					SystemQuantity:  85,
					CountedQuantity: 85,
					Owner:           "HN Lead",
				},
			},
		},
		{
			ID:            "rec-hcm-260425-day",
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			Date:          "2026-04-25",
			ShiftCode:     "day",
			Status:        domain.ReconciliationStatusClosed,
			Owner:         "Warehouse Lead",
			ClosedAt:      time.Date(2026, 4, 25, 17, 42, 0, 0, time.UTC),
			ClosedBy:      "user-warehouse-lead",
			Operations: domain.ReconciliationOperations{
				OrderCount:             39,
				HandoverOrderCount:     32,
				ReturnOrderCount:       2,
				StockMovementCount:     5,
				StockCountSessionCount: 1,
				PendingIssueCount:      0,
			},
			Checklist: []domain.ReconciliationChecklistItem{
				{Key: "shipments", Label: "Shipments reconciled", Complete: true, Blocking: true},
				{Key: "returns", Label: "Returns triaged", Complete: true, Blocking: true},
				{Key: "variance", Label: "Stock variance reviewed", Complete: true, Blocking: true},
				{Key: "pending_tasks", Label: "P0 tasks cleared or noted", Complete: true, Blocking: true},
			},
			Lines: []domain.ReconciliationLine{
				{
					ID:              "line-hcm-serum-2603z",
					SKU:             "SERUM-30ML",
					BatchNo:         "LOT-2603Z",
					BinCode:         "A-01",
					SystemQuantity:  96,
					CountedQuantity: 96,
					Owner:           "Warehouse Lead",
				},
			},
		},
	}
}
