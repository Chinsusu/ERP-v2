package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrStockTransferNotFound = errors.New("stock transfer not found")
var ErrWarehouseIssueNotFound = errors.New("warehouse issue not found")

const (
	stockTransferCreatedAction   = "inventory.stock_transfer.created"
	stockTransferSubmittedAction = "inventory.stock_transfer.submitted"
	stockTransferApprovedAction  = "inventory.stock_transfer.approved"
	stockTransferPostedAction    = "inventory.stock_transfer.posted"
	stockTransferEntityType      = "inventory.stock_transfer"

	warehouseIssueCreatedAction   = "inventory.warehouse_issue.created"
	warehouseIssueSubmittedAction = "inventory.warehouse_issue.submitted"
	warehouseIssueApprovedAction  = "inventory.warehouse_issue.approved"
	warehouseIssuePostedAction    = "inventory.warehouse_issue.posted"
	warehouseIssueEntityType      = "inventory.warehouse_issue"
)

type StockTransferStore interface {
	ListStockTransfers(ctx context.Context) ([]domain.StockTransfer, error)
	FindStockTransferByID(ctx context.Context, id string) (domain.StockTransfer, error)
	SaveStockTransfer(ctx context.Context, transfer domain.StockTransfer) error
}

type WarehouseIssueStore interface {
	ListWarehouseIssues(ctx context.Context) ([]domain.WarehouseIssue, error)
	FindWarehouseIssueByID(ctx context.Context, id string) (domain.WarehouseIssue, error)
	SaveWarehouseIssue(ctx context.Context, issue domain.WarehouseIssue) error
}

type PrototypeStockTransferStore struct {
	mu      sync.RWMutex
	records map[string]domain.StockTransfer
}

type PrototypeWarehouseIssueStore struct {
	mu      sync.RWMutex
	records map[string]domain.WarehouseIssue
}

type StockTransferService struct {
	store    StockTransferStore
	movement StockMovementStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type WarehouseIssueService struct {
	store    WarehouseIssueStore
	movement StockMovementStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type CreateStockTransferInput struct {
	ID                       string
	TransferNo               string
	OrgID                    string
	SourceWarehouseID        string
	SourceWarehouseCode      string
	DestinationWarehouseID   string
	DestinationWarehouseCode string
	ReasonCode               string
	RequestedBy              string
	RequestID                string
	Lines                    []CreateStockTransferLineInput
}

type CreateStockTransferLineInput struct {
	ID                      string
	ItemID                  string
	SKU                     string
	BatchID                 string
	BatchNo                 string
	SourceLocationID        string
	SourceLocationCode      string
	DestinationLocationID   string
	DestinationLocationCode string
	Quantity                string
	BaseUOMCode             string
	Note                    string
}

type StockTransferTransitionInput struct {
	ID        string
	ActorID   string
	Action    string
	RequestID string
}

type StockTransferResult struct {
	StockTransfer domain.StockTransfer
	AuditLogID    string
}

type CreateWarehouseIssueInput struct {
	ID              string
	IssueNo         string
	OrgID           string
	WarehouseID     string
	WarehouseCode   string
	DestinationType string
	DestinationName string
	ReasonCode      string
	RequestedBy     string
	RequestID       string
	Lines           []CreateWarehouseIssueLineInput
}

type CreateWarehouseIssueLineInput struct {
	ID                 string
	ItemID             string
	SKU                string
	ItemName           string
	Category           string
	BatchID            string
	BatchNo            string
	LocationID         string
	LocationCode       string
	Quantity           string
	BaseUOMCode        string
	Specification      string
	SourceDocumentType string
	SourceDocumentID   string
	Note               string
}

type WarehouseIssueTransitionInput struct {
	ID        string
	ActorID   string
	Action    string
	RequestID string
}

type WarehouseIssueResult struct {
	WarehouseIssue domain.WarehouseIssue
	AuditLogID     string
}

func NewPrototypeStockTransferStore() *PrototypeStockTransferStore {
	return &PrototypeStockTransferStore{records: make(map[string]domain.StockTransfer)}
}

func NewPrototypeWarehouseIssueStore() *PrototypeWarehouseIssueStore {
	return &PrototypeWarehouseIssueStore{records: make(map[string]domain.WarehouseIssue)}
}

func NewStockTransferService(
	store StockTransferStore,
	movement StockMovementStore,
	auditLog audit.LogStore,
) StockTransferService {
	return StockTransferService{
		store:    store,
		movement: movement,
		auditLog: auditLog,
		clock:    func() time.Time { return time.Now().UTC() },
	}
}

func (s StockTransferService) WithClock(clock func() time.Time) StockTransferService {
	s.clock = clock

	return s
}

func NewWarehouseIssueService(
	store WarehouseIssueStore,
	movement StockMovementStore,
	auditLog audit.LogStore,
) WarehouseIssueService {
	return WarehouseIssueService{
		store:    store,
		movement: movement,
		auditLog: auditLog,
		clock:    func() time.Time { return time.Now().UTC() },
	}
}

func (s WarehouseIssueService) WithClock(clock func() time.Time) WarehouseIssueService {
	s.clock = clock

	return s
}

func (s StockTransferService) ListStockTransfers(ctx context.Context) ([]domain.StockTransfer, error) {
	if s.store == nil {
		return nil, errors.New("stock transfer store is required")
	}

	return s.store.ListStockTransfers(ctx)
}

func (s StockTransferService) CreateStockTransfer(
	ctx context.Context,
	input CreateStockTransferInput,
) (StockTransferResult, error) {
	if s.store == nil {
		return StockTransferResult{}, errors.New("stock transfer store is required")
	}
	if s.auditLog == nil {
		return StockTransferResult{}, errors.New("audit log store is required")
	}
	lines, err := newStockTransferLines(input.Lines)
	if err != nil {
		return StockTransferResult{}, err
	}
	transfer, err := domain.NewStockTransfer(domain.NewStockTransferInput{
		ID:                       input.ID,
		TransferNo:               input.TransferNo,
		OrgID:                    input.OrgID,
		SourceWarehouseID:        input.SourceWarehouseID,
		SourceWarehouseCode:      input.SourceWarehouseCode,
		DestinationWarehouseID:   input.DestinationWarehouseID,
		DestinationWarehouseCode: input.DestinationWarehouseCode,
		ReasonCode:               input.ReasonCode,
		RequestedBy:              input.RequestedBy,
		Lines:                    lines,
		CreatedAt:                s.clock(),
	})
	if err != nil {
		return StockTransferResult{}, err
	}
	if err := s.store.SaveStockTransfer(ctx, transfer); err != nil {
		return StockTransferResult{}, err
	}
	log, err := newStockTransferAuditLog(input.RequestedBy, input.RequestID, stockTransferCreatedAction, transfer)
	if err != nil {
		return StockTransferResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return StockTransferResult{}, err
	}

	return StockTransferResult{StockTransfer: transfer, AuditLogID: log.ID}, nil
}

func (s StockTransferService) TransitionStockTransfer(
	ctx context.Context,
	input StockTransferTransitionInput,
) (StockTransferResult, error) {
	switch strings.TrimSpace(input.Action) {
	case "submit":
		return s.transition(ctx, input, stockTransferSubmittedAction, func(transfer domain.StockTransfer, actorID string, at time.Time) (domain.StockTransfer, error) {
			return transfer.Submit(actorID, at)
		})
	case "approve":
		return s.transition(ctx, input, stockTransferApprovedAction, func(transfer domain.StockTransfer, actorID string, at time.Time) (domain.StockTransfer, error) {
			return transfer.Approve(actorID, at)
		})
	case "post":
		if s.movement == nil {
			return StockTransferResult{}, errors.New("stock movement store is required")
		}
		return s.transition(ctx, input, stockTransferPostedAction, func(transfer domain.StockTransfer, actorID string, at time.Time) (domain.StockTransfer, error) {
			posted, err := transfer.MarkPosted(actorID, at)
			if err != nil {
				return domain.StockTransfer{}, err
			}
			movements, err := newStockTransferMovements(posted, actorID, at)
			if err != nil {
				return domain.StockTransfer{}, err
			}
			for _, movement := range movements {
				if err := s.movement.Record(ctx, movement); err != nil {
					return domain.StockTransfer{}, err
				}
			}

			return posted, nil
		})
	default:
		return StockTransferResult{}, domain.ErrStockTransferInvalidStatus
	}
}

func (s StockTransferService) transition(
	ctx context.Context,
	input StockTransferTransitionInput,
	action string,
	apply func(transfer domain.StockTransfer, actorID string, at time.Time) (domain.StockTransfer, error),
) (StockTransferResult, error) {
	if s.store == nil {
		return StockTransferResult{}, errors.New("stock transfer store is required")
	}
	if s.auditLog == nil {
		return StockTransferResult{}, errors.New("audit log store is required")
	}
	current, err := s.store.FindStockTransferByID(ctx, input.ID)
	if err != nil {
		return StockTransferResult{}, err
	}
	updated, err := apply(current, input.ActorID, s.clock())
	if err != nil {
		return StockTransferResult{}, err
	}
	if err := s.store.SaveStockTransfer(ctx, updated); err != nil {
		return StockTransferResult{}, err
	}
	log, err := newStockTransferAuditLog(input.ActorID, input.RequestID, action, updated)
	if err != nil {
		return StockTransferResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return StockTransferResult{}, err
	}

	return StockTransferResult{StockTransfer: updated, AuditLogID: log.ID}, nil
}

func (s WarehouseIssueService) ListWarehouseIssues(ctx context.Context) ([]domain.WarehouseIssue, error) {
	if s.store == nil {
		return nil, errors.New("warehouse issue store is required")
	}

	return s.store.ListWarehouseIssues(ctx)
}

func (s WarehouseIssueService) CreateWarehouseIssue(
	ctx context.Context,
	input CreateWarehouseIssueInput,
) (WarehouseIssueResult, error) {
	if s.store == nil {
		return WarehouseIssueResult{}, errors.New("warehouse issue store is required")
	}
	if s.auditLog == nil {
		return WarehouseIssueResult{}, errors.New("audit log store is required")
	}
	lines, err := newWarehouseIssueLines(input.Lines)
	if err != nil {
		return WarehouseIssueResult{}, err
	}
	issue, err := domain.NewWarehouseIssue(domain.NewWarehouseIssueInput{
		ID:              input.ID,
		IssueNo:         input.IssueNo,
		OrgID:           input.OrgID,
		WarehouseID:     input.WarehouseID,
		WarehouseCode:   input.WarehouseCode,
		DestinationType: input.DestinationType,
		DestinationName: input.DestinationName,
		ReasonCode:      input.ReasonCode,
		RequestedBy:     input.RequestedBy,
		Lines:           lines,
		CreatedAt:       s.clock(),
	})
	if err != nil {
		return WarehouseIssueResult{}, err
	}
	if err := s.store.SaveWarehouseIssue(ctx, issue); err != nil {
		return WarehouseIssueResult{}, err
	}
	log, err := newWarehouseIssueAuditLog(input.RequestedBy, input.RequestID, warehouseIssueCreatedAction, issue)
	if err != nil {
		return WarehouseIssueResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return WarehouseIssueResult{}, err
	}

	return WarehouseIssueResult{WarehouseIssue: issue, AuditLogID: log.ID}, nil
}

func (s WarehouseIssueService) TransitionWarehouseIssue(
	ctx context.Context,
	input WarehouseIssueTransitionInput,
) (WarehouseIssueResult, error) {
	switch strings.TrimSpace(input.Action) {
	case "submit":
		return s.transition(ctx, input, warehouseIssueSubmittedAction, func(issue domain.WarehouseIssue, actorID string, at time.Time) (domain.WarehouseIssue, error) {
			return issue.Submit(actorID, at)
		})
	case "approve":
		return s.transition(ctx, input, warehouseIssueApprovedAction, func(issue domain.WarehouseIssue, actorID string, at time.Time) (domain.WarehouseIssue, error) {
			return issue.Approve(actorID, at)
		})
	case "post":
		if s.movement == nil {
			return WarehouseIssueResult{}, errors.New("stock movement store is required")
		}
		return s.transition(ctx, input, warehouseIssuePostedAction, func(issue domain.WarehouseIssue, actorID string, at time.Time) (domain.WarehouseIssue, error) {
			posted, err := issue.MarkPosted(actorID, at)
			if err != nil {
				return domain.WarehouseIssue{}, err
			}
			movements, err := newWarehouseIssueMovements(posted, actorID, at)
			if err != nil {
				return domain.WarehouseIssue{}, err
			}
			for _, movement := range movements {
				if err := s.movement.Record(ctx, movement); err != nil {
					return domain.WarehouseIssue{}, err
				}
			}

			return posted, nil
		})
	default:
		return WarehouseIssueResult{}, domain.ErrWarehouseIssueInvalidStatus
	}
}

func (s WarehouseIssueService) transition(
	ctx context.Context,
	input WarehouseIssueTransitionInput,
	action string,
	apply func(issue domain.WarehouseIssue, actorID string, at time.Time) (domain.WarehouseIssue, error),
) (WarehouseIssueResult, error) {
	if s.store == nil {
		return WarehouseIssueResult{}, errors.New("warehouse issue store is required")
	}
	if s.auditLog == nil {
		return WarehouseIssueResult{}, errors.New("audit log store is required")
	}
	current, err := s.store.FindWarehouseIssueByID(ctx, input.ID)
	if err != nil {
		return WarehouseIssueResult{}, err
	}
	updated, err := apply(current, input.ActorID, s.clock())
	if err != nil {
		return WarehouseIssueResult{}, err
	}
	if err := s.store.SaveWarehouseIssue(ctx, updated); err != nil {
		return WarehouseIssueResult{}, err
	}
	log, err := newWarehouseIssueAuditLog(input.ActorID, input.RequestID, action, updated)
	if err != nil {
		return WarehouseIssueResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return WarehouseIssueResult{}, err
	}

	return WarehouseIssueResult{WarehouseIssue: updated, AuditLogID: log.ID}, nil
}

func newStockTransferLines(inputs []CreateStockTransferLineInput) ([]domain.NewStockTransferLineInput, error) {
	lines := make([]domain.NewStockTransferLineInput, 0, len(inputs))
	for _, input := range inputs {
		quantity, err := decimal.ParseQuantity(input.Quantity)
		if err != nil {
			return nil, domain.ErrStockTransferInvalidQuantity
		}
		lines = append(lines, domain.NewStockTransferLineInput{
			ID:                      input.ID,
			ItemID:                  input.ItemID,
			SKU:                     input.SKU,
			BatchID:                 input.BatchID,
			BatchNo:                 input.BatchNo,
			SourceLocationID:        input.SourceLocationID,
			SourceLocationCode:      input.SourceLocationCode,
			DestinationLocationID:   input.DestinationLocationID,
			DestinationLocationCode: input.DestinationLocationCode,
			Quantity:                quantity,
			BaseUOMCode:             input.BaseUOMCode,
			Note:                    input.Note,
		})
	}

	return lines, nil
}

func newWarehouseIssueLines(inputs []CreateWarehouseIssueLineInput) ([]domain.NewWarehouseIssueLineInput, error) {
	lines := make([]domain.NewWarehouseIssueLineInput, 0, len(inputs))
	for _, input := range inputs {
		quantity, err := decimal.ParseQuantity(input.Quantity)
		if err != nil {
			return nil, domain.ErrWarehouseIssueInvalidQuantity
		}
		lines = append(lines, domain.NewWarehouseIssueLineInput{
			ID:                 input.ID,
			ItemID:             input.ItemID,
			SKU:                input.SKU,
			ItemName:           input.ItemName,
			Category:           input.Category,
			BatchID:            input.BatchID,
			BatchNo:            input.BatchNo,
			LocationID:         input.LocationID,
			LocationCode:       input.LocationCode,
			Quantity:           quantity,
			BaseUOMCode:        input.BaseUOMCode,
			Specification:      input.Specification,
			SourceDocumentType: input.SourceDocumentType,
			SourceDocumentID:   input.SourceDocumentID,
			Note:               input.Note,
		})
	}

	return lines, nil
}

func newStockTransferMovements(
	transfer domain.StockTransfer,
	actorID string,
	movementAt time.Time,
) ([]domain.StockMovement, error) {
	movements := make([]domain.StockMovement, 0, len(transfer.Lines)*2)
	for index, line := range transfer.Lines {
		out, err := domain.NewStockMovement(domain.NewStockMovementInput{
			MovementNo:       fmt.Sprintf("%s-OUT-%03d", transfer.TransferNo, index+1),
			MovementType:     domain.MovementTransferOut,
			OrgID:            transfer.OrgID,
			ItemID:           inventoryDocumentItemID(line.ItemID, line.SKU),
			BatchID:          line.BatchID,
			WarehouseID:      transfer.SourceWarehouseID,
			BinID:            line.SourceLocationID,
			Quantity:         line.Quantity,
			BaseUOMCode:      line.BaseUOMCode.String(),
			SourceQuantity:   line.Quantity,
			SourceUOMCode:    line.BaseUOMCode.String(),
			ConversionFactor: decimal.MustQuantity("1"),
			StockStatus:      domain.StockStatusAvailable,
			SourceDocType:    "stock_transfer",
			SourceDocID:      transfer.ID,
			SourceDocLineID:  line.ID,
			Reason:           transfer.ReasonCode,
			CreatedBy:        actorID,
			MovementAt:       movementAt,
		})
		if err != nil {
			return nil, err
		}
		in, err := domain.NewStockMovement(domain.NewStockMovementInput{
			MovementNo:       fmt.Sprintf("%s-IN-%03d", transfer.TransferNo, index+1),
			MovementType:     domain.MovementTransferIn,
			OrgID:            transfer.OrgID,
			ItemID:           inventoryDocumentItemID(line.ItemID, line.SKU),
			BatchID:          line.BatchID,
			WarehouseID:      transfer.DestinationWarehouseID,
			BinID:            line.DestinationLocationID,
			Quantity:         line.Quantity,
			BaseUOMCode:      line.BaseUOMCode.String(),
			SourceQuantity:   line.Quantity,
			SourceUOMCode:    line.BaseUOMCode.String(),
			ConversionFactor: decimal.MustQuantity("1"),
			StockStatus:      domain.StockStatusAvailable,
			SourceDocType:    "stock_transfer",
			SourceDocID:      transfer.ID,
			SourceDocLineID:  line.ID,
			Reason:           transfer.ReasonCode,
			CreatedBy:        actorID,
			MovementAt:       movementAt,
		})
		if err != nil {
			return nil, err
		}
		movements = append(movements, out, in)
	}

	return movements, nil
}

func newWarehouseIssueMovements(
	issue domain.WarehouseIssue,
	actorID string,
	movementAt time.Time,
) ([]domain.StockMovement, error) {
	movements := make([]domain.StockMovement, 0, len(issue.Lines))
	for index, line := range issue.Lines {
		movement, err := domain.NewStockMovement(domain.NewStockMovementInput{
			MovementNo:       fmt.Sprintf("%s-OUT-%03d", issue.IssueNo, index+1),
			MovementType:     domain.MovementWarehouseIssue,
			OrgID:            issue.OrgID,
			ItemID:           inventoryDocumentItemID(line.ItemID, line.SKU),
			BatchID:          line.BatchID,
			WarehouseID:      issue.WarehouseID,
			BinID:            line.LocationID,
			Quantity:         line.Quantity,
			BaseUOMCode:      line.BaseUOMCode.String(),
			SourceQuantity:   line.Quantity,
			SourceUOMCode:    line.BaseUOMCode.String(),
			ConversionFactor: decimal.MustQuantity("1"),
			StockStatus:      domain.StockStatusAvailable,
			SourceDocType:    "warehouse_issue",
			SourceDocID:      issue.ID,
			SourceDocLineID:  line.ID,
			Reason:           issue.ReasonCode,
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

func inventoryDocumentItemID(itemID string, sku string) string {
	itemID = strings.TrimSpace(itemID)
	if itemID != "" {
		return itemID
	}
	sku = strings.ToLower(strings.TrimSpace(sku))
	if sku == "" {
		return "item-unknown-sku"
	}

	return "item-" + sku
}

func (s *PrototypeStockTransferStore) ListStockTransfers(_ context.Context) ([]domain.StockTransfer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := make([]domain.StockTransfer, 0, len(s.records))
	for _, transfer := range s.records {
		rows = append(rows, transfer.Clone())
	}
	domain.SortStockTransfers(rows)

	return rows, nil
}

func (s *PrototypeStockTransferStore) FindStockTransferByID(_ context.Context, id string) (domain.StockTransfer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	transfer, ok := s.records[strings.TrimSpace(id)]
	if !ok {
		return domain.StockTransfer{}, ErrStockTransferNotFound
	}

	return transfer.Clone(), nil
}

func (s *PrototypeStockTransferStore) SaveStockTransfer(_ context.Context, transfer domain.StockTransfer) error {
	if err := transfer.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[transfer.ID] = transfer.Clone()

	return nil
}

func (s *PrototypeWarehouseIssueStore) ListWarehouseIssues(_ context.Context) ([]domain.WarehouseIssue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := make([]domain.WarehouseIssue, 0, len(s.records))
	for _, issue := range s.records {
		rows = append(rows, issue.Clone())
	}
	domain.SortWarehouseIssues(rows)

	return rows, nil
}

func (s *PrototypeWarehouseIssueStore) FindWarehouseIssueByID(_ context.Context, id string) (domain.WarehouseIssue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	issue, ok := s.records[strings.TrimSpace(id)]
	if !ok {
		return domain.WarehouseIssue{}, ErrWarehouseIssueNotFound
	}

	return issue.Clone(), nil
}

func (s *PrototypeWarehouseIssueStore) SaveWarehouseIssue(_ context.Context, issue domain.WarehouseIssue) error {
	if err := issue.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[issue.ID] = issue.Clone()

	return nil
}

func newStockTransferAuditLog(actorID string, requestID string, action string, transfer domain.StockTransfer) (audit.Log, error) {
	return audit.NewLog(audit.NewLogInput{
		OrgID:      transfer.OrgID,
		ActorID:    strings.TrimSpace(actorID),
		Action:     action,
		EntityType: stockTransferEntityType,
		EntityID:   transfer.ID,
		RequestID:  strings.TrimSpace(requestID),
		AfterData: map[string]any{
			"transfer_no":              transfer.TransferNo,
			"status":                   string(transfer.Status),
			"source_warehouse_id":      transfer.SourceWarehouseID,
			"destination_warehouse_id": transfer.DestinationWarehouseID,
			"reason_code":              transfer.ReasonCode,
			"line_count":               len(transfer.Lines),
			"submitted_by":             transfer.SubmittedBy,
			"approved_by":              transfer.ApprovedBy,
			"posted_by":                transfer.PostedBy,
		},
		Metadata:  map[string]any{"source": "stock transfer"},
		CreatedAt: transfer.UpdatedAt,
	})
}

func newWarehouseIssueAuditLog(actorID string, requestID string, action string, issue domain.WarehouseIssue) (audit.Log, error) {
	return audit.NewLog(audit.NewLogInput{
		OrgID:      issue.OrgID,
		ActorID:    strings.TrimSpace(actorID),
		Action:     action,
		EntityType: warehouseIssueEntityType,
		EntityID:   issue.ID,
		RequestID:  strings.TrimSpace(requestID),
		AfterData: map[string]any{
			"issue_no":         issue.IssueNo,
			"status":           string(issue.Status),
			"warehouse_id":     issue.WarehouseID,
			"destination_type": issue.DestinationType,
			"destination_name": issue.DestinationName,
			"reason_code":      issue.ReasonCode,
			"line_count":       len(issue.Lines),
			"submitted_by":     issue.SubmittedBy,
			"approved_by":      issue.ApprovedBy,
			"posted_by":        issue.PostedBy,
		},
		Metadata:  map[string]any{"source": "warehouse issue"},
		CreatedAt: issue.UpdatedAt,
	})
}
