package audit

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Event struct {
	ActorID   string
	Action    string
	Resource  string
	CreatedAt time.Time
}

type Log struct {
	ID         string
	OrgID      string
	ActorID    string
	Action     string
	EntityType string
	EntityID   string
	RequestID  string
	BeforeData map[string]any
	AfterData  map[string]any
	Metadata   map[string]any
	CreatedAt  time.Time
}

type NewLogInput struct {
	ID         string
	OrgID      string
	ActorID    string
	Action     string
	EntityType string
	EntityID   string
	RequestID  string
	BeforeData map[string]any
	AfterData  map[string]any
	Metadata   map[string]any
	CreatedAt  time.Time
}

type Query struct {
	ActorID    string
	Action     string
	EntityType string
	EntityID   string
	Limit      int
}

type LogStore interface {
	Record(ctx context.Context, log Log) error
	List(ctx context.Context, query Query) ([]Log, error)
}

type InMemoryLogStore struct {
	mu   sync.RWMutex
	logs []Log
}

var generatedLogSequence atomic.Uint64

func NewEvent(actorID string, action string, resource string) Event {
	return Event{
		ActorID:   actorID,
		Action:    action,
		Resource:  resource,
		CreatedAt: time.Now().UTC(),
	}
}

func NewLog(input NewLogInput) (Log, error) {
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	log := Log{
		ID:         strings.TrimSpace(input.ID),
		OrgID:      strings.TrimSpace(input.OrgID),
		ActorID:    strings.TrimSpace(input.ActorID),
		Action:     strings.TrimSpace(input.Action),
		EntityType: strings.TrimSpace(input.EntityType),
		EntityID:   strings.TrimSpace(input.EntityID),
		RequestID:  strings.TrimSpace(input.RequestID),
		BeforeData: cloneMap(input.BeforeData),
		AfterData:  cloneMap(input.AfterData),
		Metadata:   cloneMap(input.Metadata),
		CreatedAt:  createdAt.UTC(),
	}
	if log.ID == "" {
		log.ID = fmt.Sprintf("audit_%d_%d", log.CreatedAt.UnixNano(), generatedLogSequence.Add(1))
	}
	if log.ActorID == "" {
		return Log{}, errors.New("audit actor id is required")
	}
	if log.Action == "" {
		return Log{}, errors.New("audit action is required")
	}
	if log.EntityType == "" {
		return Log{}, errors.New("audit entity type is required")
	}
	if log.EntityID == "" {
		return Log{}, errors.New("audit entity id is required")
	}
	if log.Metadata == nil {
		log.Metadata = map[string]any{}
	}

	return log, nil
}

func NewInMemoryLogStore(logs ...Log) *InMemoryLogStore {
	store := &InMemoryLogStore{}
	for _, log := range logs {
		store.logs = append(store.logs, cloneLog(log))
	}
	sortLogs(store.logs)

	return store
}

func NewPrototypeLogStore() *InMemoryLogStore {
	baseTime := time.Date(2026, 4, 26, 8, 30, 0, 0, time.UTC)
	logs := make([]Log, 0, 4)
	for _, input := range []NewLogInput{
		{
			ID:         "audit-adjust-260426-0001",
			OrgID:      "org-my-pham",
			ActorID:    "user-erp-admin",
			Action:     "inventory.stock_movement.adjusted",
			EntityType: "inventory.stock_movement",
			EntityID:   "mov-adjust-260426-0001",
			RequestID:  "req_adjust_260426",
			AfterData: map[string]any{
				"movement_type":   "ADJUST",
				"quantity":        8,
				"warehouse_id":    "wh-hcm",
				"sku":             "SERUM-30ML",
				"available_delta": -8,
			},
			Metadata: map[string]any{
				"reason": "cycle count variance",
				"source": "inventory stock movement",
			},
			CreatedAt: baseTime,
		},
		{
			ID:         "audit-rbac-260426-0002",
			OrgID:      "org-my-pham",
			ActorID:    "user-erp-admin",
			Action:     "security.role.assigned",
			EntityType: "core.user_role",
			EntityID:   "role-assignment-warehouse-lead",
			RequestID:  "req_rbac_260426",
			AfterData: map[string]any{
				"user_id": "user-warehouse-lead",
				"role":    "WAREHOUSE_LEAD",
			},
			Metadata: map[string]any{
				"scope_type": "warehouse",
				"warehouse":  "HCM",
			},
			CreatedAt: baseTime.Add(-40 * time.Minute),
		},
		{
			ID:         "audit-qc-260426-0003",
			OrgID:      "org-my-pham",
			ActorID:    "user-qa",
			Action:     "qc.lot.released",
			EntityType: "qc.inspection",
			EntityID:   "qc-inspection-260426-0007",
			RequestID:  "req_qc_260426",
			BeforeData: map[string]any{
				"status": "hold",
			},
			AfterData: map[string]any{
				"status": "released",
			},
			Metadata: map[string]any{
				"lot_no": "LOT-2604A",
				"sku":    "SERUM-30ML",
			},
			CreatedAt: baseTime.Add(-95 * time.Minute),
		},
		{
			ID:         "audit-batch-qc-260426-0004",
			OrgID:      "org-local",
			ActorID:    "user-qa",
			Action:     "inventory.batch.qc_status_changed",
			EntityType: "inventory.batch",
			EntityID:   "batch-cream-2603b",
			RequestID:  "req_batch_qc_260426",
			BeforeData: map[string]any{
				"qc_status": "hold",
			},
			AfterData: map[string]any{
				"qc_status": "pass",
			},
			Metadata: map[string]any{
				"batch_no":     "LOT-2603B",
				"business_ref": "QC-260426-0004",
				"reason":       "incoming inspection passed",
				"sku":          "CREAM-50G",
			},
			CreatedAt: baseTime.Add(-50 * time.Minute),
		},
	} {
		log, err := NewLog(input)
		if err != nil {
			continue
		}
		logs = append(logs, log)
	}

	return NewInMemoryLogStore(logs...)
}

func (s *InMemoryLogStore) Record(_ context.Context, log Log) error {
	if s == nil {
		return errors.New("audit log store is required")
	}
	normalizedLog, err := NewLog(NewLogInput{
		ID:         log.ID,
		OrgID:      log.OrgID,
		ActorID:    log.ActorID,
		Action:     log.Action,
		EntityType: log.EntityType,
		EntityID:   log.EntityID,
		RequestID:  log.RequestID,
		BeforeData: log.BeforeData,
		AfterData:  log.AfterData,
		Metadata:   log.Metadata,
		CreatedAt:  log.CreatedAt,
	})
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs = append(s.logs, normalizedLog)
	sortLogs(s.logs)

	return nil
}

func (s *InMemoryLogStore) List(_ context.Context, query Query) ([]Log, error) {
	if s == nil {
		return nil, errors.New("audit log store is required")
	}

	query = normalizeQuery(query)
	s.mu.RLock()
	defer s.mu.RUnlock()

	logs := make([]Log, 0, len(s.logs))
	for _, log := range s.logs {
		if !matchesQuery(log, query) {
			continue
		}
		logs = append(logs, cloneLog(log))
		if len(logs) == query.Limit {
			break
		}
	}

	return logs, nil
}

func normalizeQuery(query Query) Query {
	query.ActorID = strings.TrimSpace(query.ActorID)
	query.Action = strings.TrimSpace(query.Action)
	query.EntityType = strings.TrimSpace(query.EntityType)
	query.EntityID = strings.TrimSpace(query.EntityID)
	if query.Limit <= 0 {
		query.Limit = 50
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	return query
}

func matchesQuery(log Log, query Query) bool {
	if query.ActorID != "" && !strings.EqualFold(log.ActorID, query.ActorID) {
		return false
	}
	if query.Action != "" && !strings.EqualFold(log.Action, query.Action) {
		return false
	}
	if query.EntityType != "" && !strings.EqualFold(log.EntityType, query.EntityType) {
		return false
	}
	if query.EntityID != "" && !strings.EqualFold(log.EntityID, query.EntityID) {
		return false
	}

	return true
}

func sortLogs(logs []Log) {
	sort.SliceStable(logs, func(i, j int) bool {
		if logs[i].CreatedAt.Equal(logs[j].CreatedAt) {
			return false
		}

		return logs[i].CreatedAt.After(logs[j].CreatedAt)
	})
}

func cloneLog(log Log) Log {
	log.BeforeData = cloneMap(log.BeforeData)
	log.AfterData = cloneMap(log.AfterData)
	log.Metadata = cloneMap(log.Metadata)
	return log
}

func cloneMap(value map[string]any) map[string]any {
	if value == nil {
		return nil
	}

	copyValue := make(map[string]any, len(value))
	for key, item := range value {
		copyValue[key] = item
	}

	return copyValue
}
