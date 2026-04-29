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

var ErrSupplierRejectionNotFound = errors.New("supplier rejection not found")
var ErrSupplierRejectionDuplicate = errors.New("supplier rejection already exists")

const (
	supplierRejectionAuditEntityType = "inventory.supplier_rejection"
	supplierRejectionCreatedAction   = "inventory.supplier_rejection.created"
)

type SupplierRejectionStore interface {
	List(ctx context.Context, filter domain.SupplierRejectionFilter) ([]domain.SupplierRejection, error)
	Get(ctx context.Context, id string) (domain.SupplierRejection, error)
	Save(ctx context.Context, rejection domain.SupplierRejection) error
}

type ListSupplierRejections struct {
	store SupplierRejectionStore
}

type CreateSupplierRejection struct {
	store    SupplierRejectionStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type CreateSupplierRejectionInput struct {
	ID                    string
	OrgID                 string
	RejectionNo           string
	SupplierID            string
	SupplierCode          string
	SupplierName          string
	PurchaseOrderID       string
	PurchaseOrderNo       string
	GoodsReceiptID        string
	GoodsReceiptNo        string
	InboundQCInspectionID string
	WarehouseID           string
	WarehouseCode         string
	Reason                string
	Lines                 []CreateSupplierRejectionLineInput
	Attachments           []CreateSupplierRejectionAttachmentInput
	ActorID               string
	RequestID             string
}

type CreateSupplierRejectionLineInput struct {
	ID                    string
	PurchaseOrderLineID   string
	GoodsReceiptLineID    string
	InboundQCInspectionID string
	ItemID                string
	SKU                   string
	ItemName              string
	BatchID               string
	BatchNo               string
	LotNo                 string
	ExpiryDate            time.Time
	RejectedQuantity      string
	UOMCode               string
	BaseUOMCode           string
	Reason                string
}

type CreateSupplierRejectionAttachmentInput struct {
	ID          string
	LineID      string
	FileName    string
	ObjectKey   string
	ContentType string
	Source      string
}

type SupplierRejectionResult struct {
	Rejection  domain.SupplierRejection
	AuditLogID string
}

func NewListSupplierRejections(store SupplierRejectionStore) ListSupplierRejections {
	return ListSupplierRejections{store: store}
}

func NewCreateSupplierRejection(store SupplierRejectionStore, auditLog audit.LogStore) CreateSupplierRejection {
	return CreateSupplierRejection{
		store:    store,
		auditLog: auditLog,
		clock:    func() time.Time { return time.Now().UTC() },
	}
}

func (uc ListSupplierRejections) Execute(
	ctx context.Context,
	filter domain.SupplierRejectionFilter,
) ([]domain.SupplierRejection, error) {
	if uc.store == nil {
		return nil, errors.New("supplier rejection store is required")
	}

	return uc.store.List(ctx, filter)
}

func (uc CreateSupplierRejection) Execute(
	ctx context.Context,
	input CreateSupplierRejectionInput,
) (SupplierRejectionResult, error) {
	if uc.store == nil {
		return SupplierRejectionResult{}, errors.New("supplier rejection store is required")
	}
	if uc.auditLog == nil {
		return SupplierRejectionResult{}, errors.New("audit log store is required")
	}

	now := uc.clock()
	rejectionNo := strings.ToUpper(strings.TrimSpace(input.RejectionNo))
	if rejectionNo == "" {
		rejectionNo = defaultSupplierRejectionNo(now)
	}
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = strings.ToLower(rejectionNo)
	}
	orgID := strings.TrimSpace(input.OrgID)
	if orgID == "" {
		orgID = defaultReceivingOrgID
	}

	rejection, err := domain.NewSupplierRejection(domain.NewSupplierRejectionInput{
		ID:                    id,
		OrgID:                 orgID,
		RejectionNo:           rejectionNo,
		SupplierID:            input.SupplierID,
		SupplierCode:          input.SupplierCode,
		SupplierName:          input.SupplierName,
		PurchaseOrderID:       input.PurchaseOrderID,
		PurchaseOrderNo:       input.PurchaseOrderNo,
		GoodsReceiptID:        input.GoodsReceiptID,
		GoodsReceiptNo:        input.GoodsReceiptNo,
		InboundQCInspectionID: input.InboundQCInspectionID,
		WarehouseID:           input.WarehouseID,
		WarehouseCode:         input.WarehouseCode,
		Reason:                input.Reason,
		Lines:                 newSupplierRejectionLineInputs(input.Lines),
		Attachments:           newSupplierRejectionAttachmentInputs(input.Attachments),
		CreatedAt:             now,
		CreatedBy:             input.ActorID,
	})
	if err != nil {
		return SupplierRejectionResult{}, err
	}
	if err := uc.store.Save(ctx, rejection); err != nil {
		return SupplierRejectionResult{}, err
	}

	log, err := newSupplierRejectionAuditLog(input.ActorID, input.RequestID, rejection)
	if err != nil {
		return SupplierRejectionResult{}, err
	}
	if err := uc.auditLog.Record(ctx, log); err != nil {
		return SupplierRejectionResult{}, err
	}

	return SupplierRejectionResult{Rejection: rejection, AuditLogID: log.ID}, nil
}

type PrototypeSupplierRejectionStore struct {
	mu         sync.RWMutex
	rejections map[string]domain.SupplierRejection
}

func NewPrototypeSupplierRejectionStore(rows ...domain.SupplierRejection) *PrototypeSupplierRejectionStore {
	store := &PrototypeSupplierRejectionStore{rejections: make(map[string]domain.SupplierRejection)}
	for _, row := range rows {
		store.rejections[row.ID] = row.Clone()
	}

	return store
}

func (s *PrototypeSupplierRejectionStore) List(
	_ context.Context,
	filter domain.SupplierRejectionFilter,
) ([]domain.SupplierRejection, error) {
	if s == nil {
		return nil, errors.New("supplier rejection store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows := make([]domain.SupplierRejection, 0, len(s.rejections))
	for _, rejection := range s.rejections {
		if filter.Matches(rejection) {
			rows = append(rows, rejection.Clone())
		}
	}
	domain.SortSupplierRejections(rows)

	return rows, nil
}

func (s *PrototypeSupplierRejectionStore) Get(_ context.Context, id string) (domain.SupplierRejection, error) {
	if s == nil {
		return domain.SupplierRejection{}, errors.New("supplier rejection store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rejection, ok := s.rejections[strings.TrimSpace(id)]
	if !ok {
		return domain.SupplierRejection{}, ErrSupplierRejectionNotFound
	}

	return rejection.Clone(), nil
}

func (s *PrototypeSupplierRejectionStore) Save(
	_ context.Context,
	rejection domain.SupplierRejection,
) error {
	if s == nil {
		return errors.New("supplier rejection store is required")
	}
	if err := rejection.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for existingID, existing := range s.rejections {
		if strings.TrimSpace(existingID) == strings.TrimSpace(rejection.ID) {
			continue
		}
		if supplierRejectionDuplicates(existing, rejection) {
			return ErrSupplierRejectionDuplicate
		}
	}
	s.rejections[rejection.ID] = rejection.Clone()

	return nil
}

func newSupplierRejectionLineInputs(
	inputs []CreateSupplierRejectionLineInput,
) []domain.NewSupplierRejectionLineInput {
	lines := make([]domain.NewSupplierRejectionLineInput, 0, len(inputs))
	for index, input := range inputs {
		quantity, _ := decimal.ParseQuantity(input.RejectedQuantity)
		id := strings.TrimSpace(input.ID)
		if id == "" {
			id = fmt.Sprintf("line-%03d", index+1)
		}
		lines = append(lines, domain.NewSupplierRejectionLineInput{
			ID:                    id,
			PurchaseOrderLineID:   input.PurchaseOrderLineID,
			GoodsReceiptLineID:    input.GoodsReceiptLineID,
			InboundQCInspectionID: input.InboundQCInspectionID,
			ItemID:                input.ItemID,
			SKU:                   input.SKU,
			ItemName:              input.ItemName,
			BatchID:               input.BatchID,
			BatchNo:               input.BatchNo,
			LotNo:                 input.LotNo,
			ExpiryDate:            input.ExpiryDate,
			RejectedQuantity:      quantity,
			UOMCode:               input.UOMCode,
			BaseUOMCode:           input.BaseUOMCode,
			Reason:                input.Reason,
		})
	}

	return lines
}

func newSupplierRejectionAttachmentInputs(
	inputs []CreateSupplierRejectionAttachmentInput,
) []domain.NewSupplierRejectionAttachmentInput {
	attachments := make([]domain.NewSupplierRejectionAttachmentInput, 0, len(inputs))
	for _, input := range inputs {
		attachments = append(attachments, domain.NewSupplierRejectionAttachmentInput{
			ID:          input.ID,
			LineID:      input.LineID,
			FileName:    input.FileName,
			ObjectKey:   input.ObjectKey,
			ContentType: input.ContentType,
			Source:      input.Source,
		})
	}

	return attachments
}

func newSupplierRejectionAuditLog(
	actorID string,
	requestID string,
	rejection domain.SupplierRejection,
) (audit.Log, error) {
	return audit.NewLog(audit.NewLogInput{
		OrgID:      rejection.OrgID,
		ActorID:    strings.TrimSpace(actorID),
		Action:     supplierRejectionCreatedAction,
		EntityType: supplierRejectionAuditEntityType,
		EntityID:   rejection.ID,
		RequestID:  strings.TrimSpace(requestID),
		AfterData: map[string]any{
			"rejection_no":              rejection.RejectionNo,
			"status":                    string(rejection.Status),
			"supplier_id":               rejection.SupplierID,
			"purchase_order_id":         rejection.PurchaseOrderID,
			"goods_receipt_id":          rejection.GoodsReceiptID,
			"inbound_qc_inspection_id":  rejection.InboundQCInspectionID,
			"warehouse_id":              rejection.WarehouseID,
			"line_count":                len(rejection.Lines),
			"attachment_count":          len(rejection.Attachments),
			"total_rejected_base_qty":   supplierRejectionTotalBaseQty(rejection).String(),
			"first_rejected_batch_id":   firstSupplierRejectionBatchID(rejection),
			"first_rejected_batch_no":   firstSupplierRejectionBatchNo(rejection),
			"first_rejected_reason":     firstSupplierRejectionReason(rejection),
			"first_rejected_base_uom":   firstSupplierRejectionBaseUOM(rejection),
			"first_rejected_item_sku":   firstSupplierRejectionSKU(rejection),
			"first_rejected_item_id":    firstSupplierRejectionItemID(rejection),
			"first_rejected_line_id":    firstSupplierRejectionLineID(rejection),
			"first_receiving_line_id":   firstSupplierRejectionReceivingLineID(rejection),
			"first_purchase_order_line": firstSupplierRejectionPurchaseOrderLineID(rejection),
		},
		Metadata: map[string]any{
			"source":       "supplier rejection",
			"supplier_id":  rejection.SupplierID,
			"warehouse_id": rejection.WarehouseID,
			"reason":       rejection.Reason,
		},
		CreatedAt: rejection.CreatedAt,
	})
}

func supplierRejectionDuplicates(left domain.SupplierRejection, right domain.SupplierRejection) bool {
	if strings.TrimSpace(left.ID) != "" && strings.TrimSpace(left.ID) == strings.TrimSpace(right.ID) {
		return true
	}

	return strings.EqualFold(strings.TrimSpace(left.RejectionNo), strings.TrimSpace(right.RejectionNo))
}

func defaultSupplierRejectionNo(now time.Time) string {
	if now.IsZero() {
		now = time.Now().UTC()
	}

	return fmt.Sprintf("SRJ-%s-%06d", now.UTC().Format("060102"), now.UTC().UnixNano()%1000000)
}

func supplierRejectionTotalBaseQty(rejection domain.SupplierRejection) decimal.Decimal {
	total := decimal.MustQuantity("0")
	for _, line := range rejection.Lines {
		next, err := decimal.AddQuantity(total, line.RejectedQuantity)
		if err != nil {
			return total
		}
		total = next
	}

	return total
}

func firstSupplierRejectionLine(rejection domain.SupplierRejection) domain.SupplierRejectionLine {
	if len(rejection.Lines) == 0 {
		return domain.SupplierRejectionLine{}
	}

	return rejection.Lines[0]
}

func firstSupplierRejectionBatchID(rejection domain.SupplierRejection) string {
	return firstSupplierRejectionLine(rejection).BatchID
}

func firstSupplierRejectionBatchNo(rejection domain.SupplierRejection) string {
	return firstSupplierRejectionLine(rejection).BatchNo
}

func firstSupplierRejectionReason(rejection domain.SupplierRejection) string {
	return firstSupplierRejectionLine(rejection).Reason
}

func firstSupplierRejectionBaseUOM(rejection domain.SupplierRejection) string {
	return firstSupplierRejectionLine(rejection).BaseUOMCode.String()
}

func firstSupplierRejectionSKU(rejection domain.SupplierRejection) string {
	return firstSupplierRejectionLine(rejection).SKU
}

func firstSupplierRejectionItemID(rejection domain.SupplierRejection) string {
	return firstSupplierRejectionLine(rejection).ItemID
}

func firstSupplierRejectionLineID(rejection domain.SupplierRejection) string {
	return firstSupplierRejectionLine(rejection).ID
}

func firstSupplierRejectionReceivingLineID(rejection domain.SupplierRejection) string {
	return firstSupplierRejectionLine(rejection).GoodsReceiptLineID
}

func firstSupplierRejectionPurchaseOrderLineID(rejection domain.SupplierRejection) string {
	return firstSupplierRejectionLine(rejection).PurchaseOrderLineID
}
