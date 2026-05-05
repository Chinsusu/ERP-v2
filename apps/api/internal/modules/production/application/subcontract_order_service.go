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
	masterdatadomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
	apperrors "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/errors"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

var ErrSubcontractOrderNotFound = errors.New("subcontract order not found")
var ErrSubcontractOrderVersionConflict = errors.New("subcontract order version conflict")
var ErrSubcontractSampleApprovalNotFound = errors.New("subcontract sample approval not found")

const (
	ErrorCodeSubcontractOrderNotFound        response.ErrorCode = "SUBCONTRACT_ORDER_NOT_FOUND"
	ErrorCodeSubcontractOrderValidation      response.ErrorCode = "SUBCONTRACT_ORDER_VALIDATION_ERROR"
	ErrorCodeSubcontractOrderInvalidState    response.ErrorCode = "SUBCONTRACT_ORDER_INVALID_STATE"
	ErrorCodeSubcontractOrderVersionConflict response.ErrorCode = "SUBCONTRACT_ORDER_VERSION_CONFLICT"

	defaultSubcontractOrderOrgID           = "org-my-pham"
	subcontractOrderEntityType             = "production.subcontract_order"
	subcontractMaterialsIssuedAction       = "subcontract.materials_issued"
	subcontractSampleSubmittedAction       = "subcontract.sample_submitted"
	subcontractSampleApprovedAction        = "subcontract.sample_approved"
	subcontractSampleRejectedAction        = "subcontract.sample_rejected"
	subcontractFinishedGoodsAction         = "subcontract.finished_goods_received"
	subcontractFinishedGoodsAcceptedAction = "subcontract.finished_goods_accepted"
	subcontractFactoryClaimAction          = "subcontract.factory_claim_opened"
	subcontractDepositRecordedAction       = "subcontract.deposit_recorded"
	subcontractFinalPaymentAction          = "subcontract.final_payment_ready"
)

type SubcontractOrderStore interface {
	List(ctx context.Context, filter SubcontractOrderFilter) ([]productiondomain.SubcontractOrder, error)
	Get(ctx context.Context, id string) (productiondomain.SubcontractOrder, error)
	WithinTx(ctx context.Context, fn func(context.Context, SubcontractOrderTx) error) error
}

type SubcontractOrderTx interface {
	GetForUpdate(ctx context.Context, id string) (productiondomain.SubcontractOrder, error)
	Save(ctx context.Context, order productiondomain.SubcontractOrder) error
	RecordAudit(ctx context.Context, log audit.Log) error
}

type SubcontractOrderFactoryReader interface {
	GetSupplier(ctx context.Context, id string) (masterdatadomain.Supplier, error)
}

type SubcontractOrderItemReader interface {
	Get(ctx context.Context, id string) (masterdatadomain.Item, error)
}

type SubcontractOrderUOMConverter interface {
	ConvertToBase(ctx context.Context, input ConvertSubcontractOrderLineToBaseInput) (ConvertSubcontractOrderLineToBaseResult, error)
}

type SubcontractMaterialIssueMovementRecorder interface {
	Record(ctx context.Context, movement inventorydomain.StockMovement) error
}

type SubcontractFinishedGoodsReceiptMovementRecorder interface {
	Record(ctx context.Context, movement inventorydomain.StockMovement) error
}

type SubcontractPayableCreator interface {
	CreateSubcontractPayable(
		ctx context.Context,
		input CreateSubcontractPayableInput,
	) (SubcontractPayableCreationResult, error)
}

type SubcontractSampleApprovalStore interface {
	Save(ctx context.Context, sampleApproval productiondomain.SubcontractSampleApproval) error
	Get(ctx context.Context, id string) (productiondomain.SubcontractSampleApproval, error)
	GetLatestBySubcontractOrder(ctx context.Context, subcontractOrderID string) (productiondomain.SubcontractSampleApproval, error)
}

type ConvertSubcontractOrderLineToBaseInput struct {
	ItemID      string
	SKU         string
	Quantity    decimal.Decimal
	FromUOMCode string
	BaseUOMCode string
}

type ConvertSubcontractOrderLineToBaseResult struct {
	Quantity         decimal.Decimal
	SourceUOMCode    decimal.UOMCode
	BaseQuantity     decimal.Decimal
	BaseUOMCode      decimal.UOMCode
	ConversionFactor decimal.Decimal
}

type SubcontractOrderService struct {
	store                 SubcontractOrderStore
	factoryRead           SubcontractOrderFactoryReader
	itemRead              SubcontractOrderItemReader
	uomConverter          SubcontractOrderUOMConverter
	materialTransferStore SubcontractMaterialTransferStore
	materialIssueRecorder SubcontractMaterialIssueMovementRecorder
	materialTransferBuild SubcontractMaterialTransferService
	sampleApprovalStore   SubcontractSampleApprovalStore
	sampleApprovalBuild   SubcontractSampleApprovalService
	finishedGoodsStore    SubcontractFinishedGoodsReceiptStore
	finishedGoodsRecorder SubcontractFinishedGoodsReceiptMovementRecorder
	finishedGoodsBuild    SubcontractFinishedGoodsReceiptService
	factoryClaimStore     SubcontractFactoryClaimStore
	factoryClaimBuild     SubcontractFactoryClaimService
	paymentMilestoneStore SubcontractPaymentMilestoneStore
	paymentMilestoneBuild SubcontractPaymentMilestoneService
	payableCreator        SubcontractPayableCreator
	clock                 func() time.Time
}

type SubcontractOrderFilter struct {
	Search                 string
	Statuses               []productiondomain.SubcontractOrderStatus
	FactoryID              string
	FinishedItemID         string
	SourceProductionPlanID string
	ExpectedReceiptFrom    string
	ExpectedReceiptTo      string
}

type CreateSubcontractOrderInput struct {
	ID                     string
	OrgID                  string
	OrderNo                string
	FactoryID              string
	FinishedItemID         string
	PlannedQty             string
	UOMCode                string
	CurrencyCode           string
	SpecSummary            string
	SourceProductionPlanID string
	SourceProductionPlanNo string
	SampleRequired         bool
	ClaimWindowDays        int
	TargetStartDate        string
	ExpectedReceiptDate    string
	MaterialLines          []SubcontractOrderMaterialLineInput
	ActorID                string
	RequestID              string
}

type UpdateSubcontractOrderInput struct {
	ID                     string
	FactoryID              string
	FinishedItemID         string
	PlannedQty             string
	UOMCode                string
	SpecSummary            string
	SourceProductionPlanID string
	SourceProductionPlanNo string
	SampleRequired         *bool
	ClaimWindowDays        int
	TargetStartDate        string
	ExpectedReceiptDate    string
	MaterialLines          []SubcontractOrderMaterialLineInput
	ExpectedVersion        int
	ActorID                string
	RequestID              string
}

type SubcontractOrderMaterialLineInput struct {
	ID               string
	LineNo           int
	ItemID           string
	PlannedQty       string
	UOMCode          string
	UnitCost         string
	CurrencyCode     string
	LotTraceRequired bool
	Note             string
}

type SubcontractOrderActionInput struct {
	ID              string
	ExpectedVersion int
	Reason          string
	ActorID         string
	RequestID       string
}

type IssueSubcontractMaterialsInput struct {
	ID                  string
	ExpectedVersion     int
	TransferID          string
	TransferNo          string
	SourceWarehouseID   string
	SourceWarehouseCode string
	Lines               []IssueSubcontractMaterialsLineInput
	Evidence            []IssueSubcontractMaterialsEvidenceInput
	HandoverBy          string
	HandoverAt          time.Time
	ReceivedBy          string
	ReceiverContact     string
	VehicleNo           string
	Note                string
	ActorID             string
	RequestID           string
}

type IssueSubcontractMaterialsLineInput struct {
	ID                  string
	LineNo              int
	OrderMaterialLineID string
	IssueQty            string
	UOMCode             string
	BaseIssueQty        string
	BaseUOMCode         string
	ConversionFactor    string
	BatchID             string
	BatchNo             string
	LotNo               string
	SourceBinID         string
	Note                string
}

type IssueSubcontractMaterialsEvidenceInput struct {
	ID           string
	EvidenceType string
	FileName     string
	ObjectKey    string
	ExternalURL  string
	Note         string
}

type SubmitSubcontractSampleInput struct {
	ID               string
	ExpectedVersion  int
	SampleApprovalID string
	SampleCode       string
	FormulaVersion   string
	SpecVersion      string
	Evidence         []SubcontractSampleEvidenceInput
	SubmittedBy      string
	SubmittedAt      time.Time
	Note             string
	ActorID          string
	RequestID        string
}

type DecideSubcontractSampleInput struct {
	ID               string
	ExpectedVersion  int
	SampleApprovalID string
	DecisionAt       time.Time
	Reason           string
	StorageStatus    string
	ActorID          string
	RequestID        string
}

type SubcontractSampleEvidenceInput struct {
	ID           string
	EvidenceType string
	FileName     string
	ObjectKey    string
	ExternalURL  string
	Note         string
}

type ReceiveSubcontractFinishedGoodsInput struct {
	ID              string
	ExpectedVersion int
	ReceiptID       string
	ReceiptNo       string
	WarehouseID     string
	WarehouseCode   string
	LocationID      string
	LocationCode    string
	DeliveryNoteNo  string
	Lines           []ReceiveSubcontractFinishedGoodsLineInput
	Evidence        []ReceiveSubcontractFinishedGoodsEvidenceInput
	ReceivedBy      string
	ReceivedAt      time.Time
	Note            string
	ActorID         string
	RequestID       string
}

type ReceiveSubcontractFinishedGoodsLineInput struct {
	ID               string
	LineNo           int
	ItemID           string
	SKUCode          string
	ItemName         string
	BatchID          string
	BatchNo          string
	LotNo            string
	ExpiryDate       string
	ReceiveQty       string
	UOMCode          string
	BaseReceiveQty   string
	BaseUOMCode      string
	ConversionFactor string
	PackagingStatus  string
	Note             string
}

type ReceiveSubcontractFinishedGoodsEvidenceInput struct {
	ID           string
	EvidenceType string
	FileName     string
	ObjectKey    string
	ExternalURL  string
	Note         string
}

type AcceptSubcontractFinishedGoodsInput struct {
	ID              string
	ExpectedVersion int
	AcceptedBy      string
	AcceptedAt      time.Time
	Note            string
	ActorID         string
	RequestID       string
}

type PartialAcceptSubcontractFinishedGoodsInput struct {
	ID              string
	ExpectedVersion int
	AcceptedQty     string
	UOMCode         string
	BaseAcceptedQty string
	BaseUOMCode     string
	RejectedQty     string
	BaseRejectedQty string
	ClaimID         string
	ClaimNo         string
	ReceiptID       string
	ReceiptNo       string
	ReasonCode      string
	Reason          string
	Severity        string
	Evidence        []CreateSubcontractFactoryClaimEvidenceInput
	OwnerID         string
	AcceptedBy      string
	AcceptedAt      time.Time
	OpenedBy        string
	OpenedAt        time.Time
	Note            string
	ActorID         string
	RequestID       string
}

type CreateSubcontractFactoryClaimInput struct {
	ID              string
	ExpectedVersion int
	ClaimID         string
	ClaimNo         string
	ReceiptID       string
	ReceiptNo       string
	ReasonCode      string
	Reason          string
	Severity        string
	AffectedQty     string
	UOMCode         string
	BaseAffectedQty string
	BaseUOMCode     string
	Evidence        []CreateSubcontractFactoryClaimEvidenceInput
	OwnerID         string
	OpenedBy        string
	OpenedAt        time.Time
	ActorID         string
	RequestID       string
}

type CreateSubcontractFactoryClaimEvidenceInput struct {
	ID           string
	EvidenceType string
	FileName     string
	ObjectKey    string
	ExternalURL  string
	Note         string
}

type RecordSubcontractDepositInput struct {
	ID              string
	ExpectedVersion int
	MilestoneID     string
	MilestoneNo     string
	Amount          string
	RecordedBy      string
	RecordedAt      time.Time
	Note            string
	ActorID         string
	RequestID       string
}

type MarkSubcontractFinalPaymentReadyInput struct {
	ID                  string
	ExpectedVersion     int
	MilestoneID         string
	MilestoneNo         string
	Amount              string
	ReadyBy             string
	ReadyAt             time.Time
	ApprovedExceptionID string
	Note                string
	ActorID             string
	RequestID           string
}

type CreateSubcontractPayableInput struct {
	SubcontractOrder productiondomain.SubcontractOrder
	Milestone        productiondomain.SubcontractPaymentMilestone
	ActorID          string
	RequestID        string
}

type SubcontractPayableCreationResult struct {
	PayableID  string
	PayableNo  string
	AuditLogID string
}

type SubcontractOrderResult struct {
	SubcontractOrder productiondomain.SubcontractOrder
	AuditLogID       string
}

type SubcontractOrderActionResult struct {
	SubcontractOrder productiondomain.SubcontractOrder
	PreviousStatus   productiondomain.SubcontractOrderStatus
	CurrentStatus    productiondomain.SubcontractOrderStatus
	AuditLogID       string
}

type IssueSubcontractMaterialsResult struct {
	SubcontractOrder productiondomain.SubcontractOrder
	Transfer         productiondomain.SubcontractMaterialTransfer
	StockMovements   []inventorydomain.StockMovement
	AuditLogID       string
}

type SubcontractSampleApprovalResult struct {
	SubcontractOrder productiondomain.SubcontractOrder
	SampleApproval   productiondomain.SubcontractSampleApproval
	PreviousStatus   productiondomain.SubcontractOrderStatus
	CurrentStatus    productiondomain.SubcontractOrderStatus
	AuditLogID       string
}

type ReceiveSubcontractFinishedGoodsResult struct {
	SubcontractOrder productiondomain.SubcontractOrder
	Receipt          productiondomain.SubcontractFinishedGoodsReceipt
	StockMovements   []inventorydomain.StockMovement
	AuditLogID       string
}

type AcceptSubcontractFinishedGoodsResult struct {
	SubcontractOrder productiondomain.SubcontractOrder
	StockMovements   []inventorydomain.StockMovement
	PreviousStatus   productiondomain.SubcontractOrderStatus
	CurrentStatus    productiondomain.SubcontractOrderStatus
	AuditLogID       string
}

type PartialAcceptSubcontractFinishedGoodsResult struct {
	SubcontractOrder productiondomain.SubcontractOrder
	Claim            productiondomain.SubcontractFactoryClaim
	StockMovements   []inventorydomain.StockMovement
	PreviousStatus   productiondomain.SubcontractOrderStatus
	CurrentStatus    productiondomain.SubcontractOrderStatus
	AcceptAuditLogID string
	ClaimAuditLogID  string
}

type CreateSubcontractFactoryClaimResult struct {
	SubcontractOrder productiondomain.SubcontractOrder
	Claim            productiondomain.SubcontractFactoryClaim
	PreviousStatus   productiondomain.SubcontractOrderStatus
	CurrentStatus    productiondomain.SubcontractOrderStatus
	AuditLogID       string
}

type SubcontractPaymentMilestoneResult struct {
	SubcontractOrder productiondomain.SubcontractOrder
	Milestone        productiondomain.SubcontractPaymentMilestone
	PreviousStatus   productiondomain.SubcontractOrderStatus
	CurrentStatus    productiondomain.SubcontractOrderStatus
	AuditLogID       string
	SupplierPayable  SubcontractPayableCreationResult
}

type PrototypeSubcontractOrderStore struct {
	mu       sync.RWMutex
	records  map[string]productiondomain.SubcontractOrder
	auditLog audit.LogStore
	txCount  int
}

type PrototypeSubcontractSampleApprovalStore struct {
	mu      sync.RWMutex
	records map[string]productiondomain.SubcontractSampleApproval
}

type prototypeSubcontractOrderTx struct {
	store     *PrototypeSubcontractOrderStore
	auditLogs []audit.Log
}

func NewSubcontractOrderService(
	store SubcontractOrderStore,
	factoryRead SubcontractOrderFactoryReader,
	itemRead SubcontractOrderItemReader,
	uomConverter SubcontractOrderUOMConverter,
) SubcontractOrderService {
	return SubcontractOrderService{
		store:                 store,
		factoryRead:           factoryRead,
		itemRead:              itemRead,
		uomConverter:          uomConverter,
		materialTransferBuild: NewSubcontractMaterialTransferService(),
		sampleApprovalBuild:   NewSubcontractSampleApprovalService(),
		finishedGoodsBuild:    NewSubcontractFinishedGoodsReceiptService(),
		factoryClaimBuild:     NewSubcontractFactoryClaimService(),
		paymentMilestoneBuild: NewSubcontractPaymentMilestoneService(),
		clock:                 func() time.Time { return time.Now().UTC() },
	}
}

func (s SubcontractOrderService) WithMaterialIssueStores(
	transferStore SubcontractMaterialTransferStore,
	movementRecorder SubcontractMaterialIssueMovementRecorder,
) SubcontractOrderService {
	s.materialTransferStore = transferStore
	s.materialIssueRecorder = movementRecorder

	return s
}

func (s SubcontractOrderService) WithSampleApprovalStore(
	sampleApprovalStore SubcontractSampleApprovalStore,
) SubcontractOrderService {
	s.sampleApprovalStore = sampleApprovalStore

	return s
}

func (s SubcontractOrderService) WithFinishedGoodsReceiptStores(
	receiptStore SubcontractFinishedGoodsReceiptStore,
	movementRecorder SubcontractFinishedGoodsReceiptMovementRecorder,
) SubcontractOrderService {
	s.finishedGoodsStore = receiptStore
	s.finishedGoodsRecorder = movementRecorder

	return s
}

func (s SubcontractOrderService) WithFactoryClaimStore(
	claimStore SubcontractFactoryClaimStore,
) SubcontractOrderService {
	s.factoryClaimStore = claimStore

	return s
}

func (s SubcontractOrderService) WithPaymentMilestoneStore(
	paymentMilestoneStore SubcontractPaymentMilestoneStore,
) SubcontractOrderService {
	s.paymentMilestoneStore = paymentMilestoneStore

	return s
}

func (s SubcontractOrderService) WithSubcontractPayableCreator(
	payableCreator SubcontractPayableCreator,
) SubcontractOrderService {
	s.payableCreator = payableCreator

	return s
}

func NewPrototypeSubcontractOrderStore(auditLog audit.LogStore) *PrototypeSubcontractOrderStore {
	return &PrototypeSubcontractOrderStore{
		records:  make(map[string]productiondomain.SubcontractOrder),
		auditLog: auditLog,
	}
}

func NewPrototypeSubcontractSampleApprovalStore() *PrototypeSubcontractSampleApprovalStore {
	return &PrototypeSubcontractSampleApprovalStore{records: make(map[string]productiondomain.SubcontractSampleApproval)}
}

func (s SubcontractOrderService) ListSubcontractOrders(
	ctx context.Context,
	filter SubcontractOrderFilter,
) ([]productiondomain.SubcontractOrder, error) {
	if s.store == nil {
		return nil, errors.New("subcontract order store is required")
	}

	return s.store.List(ctx, filter)
}

func (s SubcontractOrderService) GetSubcontractOrder(
	ctx context.Context,
	id string,
) (productiondomain.SubcontractOrder, error) {
	if s.store == nil {
		return productiondomain.SubcontractOrder{}, errors.New("subcontract order store is required")
	}
	order, err := s.store.Get(ctx, id)
	if err != nil {
		return productiondomain.SubcontractOrder{}, MapSubcontractOrderError(err, map[string]any{"subcontract_order_id": strings.TrimSpace(id)})
	}

	return order, nil
}

func (s SubcontractOrderService) CreateSubcontractOrder(
	ctx context.Context,
	input CreateSubcontractOrderInput,
) (SubcontractOrderResult, error) {
	if err := s.ensureReadyForWrite(); err != nil {
		return SubcontractOrderResult{}, err
	}
	if err := requireSubcontractOrderActor(input.ActorID); err != nil {
		return SubcontractOrderResult{}, err
	}

	now := s.now()
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = newSubcontractOrderID(now)
	}
	orderNo := strings.TrimSpace(input.OrderNo)
	if orderNo == "" {
		orderNo = newSubcontractOrderNo(now)
	}
	orgID := strings.TrimSpace(input.OrgID)
	if orgID == "" {
		orgID = defaultSubcontractOrderOrgID
	}
	currencyCode := firstNonBlankSubcontractOrder(input.CurrencyCode, decimal.CurrencyVND.String())

	order, err := s.newSubcontractOrderDocument(ctx, newSubcontractOrderDocumentServiceInput{
		ID:                     id,
		OrgID:                  orgID,
		OrderNo:                orderNo,
		FactoryID:              input.FactoryID,
		FinishedItemID:         input.FinishedItemID,
		PlannedQty:             input.PlannedQty,
		UOMCode:                input.UOMCode,
		CurrencyCode:           currencyCode,
		SpecSummary:            input.SpecSummary,
		SourceProductionPlanID: input.SourceProductionPlanID,
		SourceProductionPlanNo: input.SourceProductionPlanNo,
		SampleRequired:         input.SampleRequired,
		ClaimWindowDays:        input.ClaimWindowDays,
		TargetStartDate:        input.TargetStartDate,
		ExpectedReceiptDate:    input.ExpectedReceiptDate,
		MaterialLines:          input.MaterialLines,
		CreatedAt:              now,
		CreatedBy:              input.ActorID,
		UpdatedAt:              now,
		UpdatedBy:              input.ActorID,
	})
	if err != nil {
		return SubcontractOrderResult{}, err
	}

	var result SubcontractOrderResult
	err = s.store.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		if err := tx.Save(txCtx, order); err != nil {
			return err
		}
		log, err := newSubcontractOrderAuditLog(
			input.ActorID,
			input.RequestID,
			"subcontract.order.created",
			order,
			nil,
			subcontractOrderAuditData(order),
			now,
		)
		if err != nil {
			return err
		}
		if err := tx.RecordAudit(txCtx, log); err != nil {
			return err
		}
		result = SubcontractOrderResult{SubcontractOrder: order, AuditLogID: log.ID}

		return nil
	})
	if err != nil {
		return SubcontractOrderResult{}, MapSubcontractOrderError(err, map[string]any{"subcontract_order_id": id})
	}

	return result, nil
}

func (s SubcontractOrderService) UpdateSubcontractOrder(
	ctx context.Context,
	input UpdateSubcontractOrderInput,
) (SubcontractOrderResult, error) {
	if err := s.ensureReadyForWrite(); err != nil {
		return SubcontractOrderResult{}, err
	}
	if err := requireSubcontractOrderActor(input.ActorID); err != nil {
		return SubcontractOrderResult{}, err
	}

	var result SubcontractOrderResult
	err := s.store.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		current, err := tx.GetForUpdate(txCtx, input.ID)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{"subcontract_order_id": strings.TrimSpace(input.ID)})
		}
		if err := ensureSubcontractOrderExpectedVersion(current, input.ExpectedVersion); err != nil {
			return err
		}
		if productiondomain.NormalizeSubcontractOrderStatus(current.Status) != productiondomain.SubcontractOrderStatusDraft {
			return MapSubcontractOrderError(productiondomain.ErrSubcontractOrderInvalidTransition, map[string]any{
				"subcontract_order_id": current.ID,
				"status":               string(current.Status),
			})
		}

		sampleRequired := current.SampleRequired
		if input.SampleRequired != nil {
			sampleRequired = *input.SampleRequired
		}
		materialLines := subcontractOrderMaterialLineInputsFromDomain(current.MaterialLines)
		if input.MaterialLines != nil {
			materialLines = input.MaterialLines
		}

		now := s.now()
		updated, err := s.newSubcontractOrderDocument(txCtx, newSubcontractOrderDocumentServiceInput{
			ID:                     current.ID,
			OrgID:                  current.OrgID,
			OrderNo:                current.OrderNo,
			FactoryID:              firstNonBlankSubcontractOrder(input.FactoryID, current.FactoryID),
			FinishedItemID:         firstNonBlankSubcontractOrder(input.FinishedItemID, current.FinishedItemID),
			PlannedQty:             firstNonBlankSubcontractOrder(input.PlannedQty, current.PlannedQty.String()),
			UOMCode:                firstNonBlankSubcontractOrder(input.UOMCode, current.UOMCode.String()),
			CurrencyCode:           current.CurrencyCode.String(),
			SpecSummary:            firstNonBlankSubcontractOrder(input.SpecSummary, current.SpecSummary),
			SourceProductionPlanID: firstNonBlankSubcontractOrder(input.SourceProductionPlanID, current.SourceProductionPlanID),
			SourceProductionPlanNo: firstNonBlankSubcontractOrder(input.SourceProductionPlanNo, current.SourceProductionPlanNo),
			SampleRequired:         sampleRequired,
			ClaimWindowDays:        firstNonZeroSubcontractOrder(input.ClaimWindowDays, current.ClaimWindowDays),
			TargetStartDate:        firstNonBlankSubcontractOrder(input.TargetStartDate, current.TargetStartDate),
			ExpectedReceiptDate:    firstNonBlankSubcontractOrder(input.ExpectedReceiptDate, current.ExpectedReceiptDate),
			MaterialLines:          materialLines,
			CreatedAt:              current.CreatedAt,
			CreatedBy:              current.CreatedBy,
			UpdatedAt:              now,
			UpdatedBy:              input.ActorID,
		})
		if err != nil {
			return err
		}
		updated.Version = current.Version + 1

		if err := tx.Save(txCtx, updated); err != nil {
			return err
		}
		log, err := newSubcontractOrderAuditLog(
			input.ActorID,
			input.RequestID,
			"subcontract.order.updated",
			updated,
			subcontractOrderAuditData(current),
			subcontractOrderAuditData(updated),
			now,
		)
		if err != nil {
			return err
		}
		if err := tx.RecordAudit(txCtx, log); err != nil {
			return err
		}
		result = SubcontractOrderResult{SubcontractOrder: updated, AuditLogID: log.ID}

		return nil
	})
	if err != nil {
		return SubcontractOrderResult{}, err
	}

	return result, nil
}

func (s SubcontractOrderService) SubmitSubcontractOrder(
	ctx context.Context,
	input SubcontractOrderActionInput,
) (SubcontractOrderActionResult, error) {
	return s.transition(ctx, input, "subcontract.order.submitted", func(
		order productiondomain.SubcontractOrder,
		actorID string,
		changedAt time.Time,
	) (productiondomain.SubcontractOrder, error) {
		return order.Submit(actorID, changedAt)
	})
}

func (s SubcontractOrderService) ApproveSubcontractOrder(
	ctx context.Context,
	input SubcontractOrderActionInput,
) (SubcontractOrderActionResult, error) {
	return s.transition(ctx, input, "subcontract.order.approved", func(
		order productiondomain.SubcontractOrder,
		actorID string,
		changedAt time.Time,
	) (productiondomain.SubcontractOrder, error) {
		return order.Approve(actorID, changedAt)
	})
}

func (s SubcontractOrderService) ConfirmFactorySubcontractOrder(
	ctx context.Context,
	input SubcontractOrderActionInput,
) (SubcontractOrderActionResult, error) {
	return s.transition(ctx, input, "subcontract.order.factory_confirmed", func(
		order productiondomain.SubcontractOrder,
		actorID string,
		changedAt time.Time,
	) (productiondomain.SubcontractOrder, error) {
		return order.ConfirmFactory(actorID, changedAt)
	})
}

func (s SubcontractOrderService) StartMassProductionSubcontractOrder(
	ctx context.Context,
	input SubcontractOrderActionInput,
) (SubcontractOrderActionResult, error) {
	return s.transition(ctx, input, "subcontract.order.mass_production_started", func(
		order productiondomain.SubcontractOrder,
		actorID string,
		changedAt time.Time,
	) (productiondomain.SubcontractOrder, error) {
		return order.StartMassProduction(actorID, changedAt)
	})
}

func (s SubcontractOrderService) CancelSubcontractOrder(
	ctx context.Context,
	input SubcontractOrderActionInput,
) (SubcontractOrderActionResult, error) {
	if strings.TrimSpace(input.Reason) == "" {
		return SubcontractOrderActionResult{}, subcontractOrderValidationError(
			productiondomain.ErrSubcontractOrderRequiredField,
			map[string]any{"field": "reason"},
		)
	}

	return s.transition(ctx, input, "subcontract.order.cancelled", func(
		order productiondomain.SubcontractOrder,
		actorID string,
		changedAt time.Time,
	) (productiondomain.SubcontractOrder, error) {
		return order.CancelWithReason(actorID, input.Reason, changedAt)
	})
}

func (s SubcontractOrderService) CloseSubcontractOrder(
	ctx context.Context,
	input SubcontractOrderActionInput,
) (SubcontractOrderActionResult, error) {
	return s.transition(ctx, input, "subcontract.order.closed", func(
		order productiondomain.SubcontractOrder,
		actorID string,
		changedAt time.Time,
	) (productiondomain.SubcontractOrder, error) {
		return order.Close(actorID, changedAt)
	})
}

func (s SubcontractOrderService) IssueSubcontractMaterials(
	ctx context.Context,
	input IssueSubcontractMaterialsInput,
) (IssueSubcontractMaterialsResult, error) {
	if err := s.ensureReadyForMaterialIssue(); err != nil {
		return IssueSubcontractMaterialsResult{}, err
	}
	if err := requireSubcontractOrderActor(input.ActorID); err != nil {
		return IssueSubcontractMaterialsResult{}, err
	}

	var result IssueSubcontractMaterialsResult
	err := s.store.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		current, err := tx.GetForUpdate(txCtx, input.ID)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{"subcontract_order_id": strings.TrimSpace(input.ID)})
		}
		if err := ensureSubcontractOrderExpectedVersion(current, input.ExpectedVersion); err != nil {
			return err
		}

		buildResult, err := s.materialTransferBuild.BuildIssue(txCtx, BuildSubcontractMaterialTransferInput{
			ID:                  input.TransferID,
			TransferNo:          input.TransferNo,
			Order:               current,
			SourceWarehouseID:   input.SourceWarehouseID,
			SourceWarehouseCode: input.SourceWarehouseCode,
			Lines:               subcontractMaterialIssueLineInputs(input.Lines),
			Evidence:            subcontractMaterialIssueEvidenceInputs(input.Evidence),
			HandoverBy:          input.HandoverBy,
			HandoverAt:          input.HandoverAt,
			ReceivedBy:          input.ReceivedBy,
			ReceiverContact:     input.ReceiverContact,
			VehicleNo:           input.VehicleNo,
			Note:                input.Note,
			ActorID:             input.ActorID,
		})
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{
				"subcontract_order_id": current.ID,
				"status":               string(current.Status),
			})
		}
		if err := s.materialTransferStore.Save(txCtx, buildResult.Transfer); err != nil {
			return MapSubcontractOrderError(err, map[string]any{
				"subcontract_order_id": current.ID,
				"transfer_id":          buildResult.Transfer.ID,
			})
		}
		for _, movement := range buildResult.StockMovements {
			if err := s.materialIssueRecorder.Record(txCtx, movement); err != nil {
				return err
			}
		}
		if err := tx.Save(txCtx, buildResult.UpdatedOrder); err != nil {
			return err
		}
		afterData := subcontractOrderAuditData(buildResult.UpdatedOrder)
		for key, value := range subcontractMaterialIssueAuditData(buildResult) {
			afterData[key] = value
		}
		log, err := newSubcontractOrderAuditLog(
			input.ActorID,
			input.RequestID,
			subcontractMaterialsIssuedAction,
			buildResult.UpdatedOrder,
			subcontractOrderAuditData(current),
			afterData,
			buildResult.Transfer.HandoverAt,
		)
		if err != nil {
			return err
		}
		if err := tx.RecordAudit(txCtx, log); err != nil {
			return err
		}
		result = IssueSubcontractMaterialsResult{
			SubcontractOrder: buildResult.UpdatedOrder,
			Transfer:         buildResult.Transfer,
			StockMovements:   buildResult.StockMovements,
			AuditLogID:       log.ID,
		}

		return nil
	})
	if err != nil {
		return IssueSubcontractMaterialsResult{}, err
	}

	return result, nil
}

func (s SubcontractOrderService) SubmitSubcontractSample(
	ctx context.Context,
	input SubmitSubcontractSampleInput,
) (SubcontractSampleApprovalResult, error) {
	if err := s.ensureReadyForSampleApproval(); err != nil {
		return SubcontractSampleApprovalResult{}, err
	}
	if err := requireSubcontractOrderActor(input.ActorID); err != nil {
		return SubcontractSampleApprovalResult{}, err
	}

	var result SubcontractSampleApprovalResult
	err := s.store.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		current, err := tx.GetForUpdate(txCtx, input.ID)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{"subcontract_order_id": strings.TrimSpace(input.ID)})
		}
		if err := ensureSubcontractOrderExpectedVersion(current, input.ExpectedVersion); err != nil {
			return err
		}

		buildResult, err := s.sampleApprovalBuild.BuildSubmission(txCtx, BuildSubcontractSampleSubmissionInput{
			ID:             input.SampleApprovalID,
			Order:          current,
			SampleCode:     input.SampleCode,
			FormulaVersion: input.FormulaVersion,
			SpecVersion:    input.SpecVersion,
			Evidence:       subcontractSampleBuildEvidenceInputs(input.Evidence),
			SubmittedBy:    input.SubmittedBy,
			SubmittedAt:    input.SubmittedAt,
			Note:           input.Note,
			ActorID:        input.ActorID,
		})
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{
				"subcontract_order_id": current.ID,
				"status":               string(current.Status),
			})
		}
		if err := s.sampleApprovalStore.Save(txCtx, buildResult.SampleApproval); err != nil {
			return MapSubcontractOrderError(err, map[string]any{
				"subcontract_order_id": current.ID,
				"sample_approval_id":   buildResult.SampleApproval.ID,
			})
		}
		if err := tx.Save(txCtx, buildResult.UpdatedOrder); err != nil {
			return err
		}
		afterData := subcontractOrderAuditData(buildResult.UpdatedOrder)
		for key, value := range subcontractSampleApprovalAuditData(buildResult.SampleApproval) {
			afterData[key] = value
		}
		log, err := newSubcontractOrderAuditLog(
			input.ActorID,
			input.RequestID,
			subcontractSampleSubmittedAction,
			buildResult.UpdatedOrder,
			subcontractOrderAuditData(current),
			afterData,
			buildResult.SampleApproval.SubmittedAt,
		)
		if err != nil {
			return err
		}
		if err := tx.RecordAudit(txCtx, log); err != nil {
			return err
		}
		result = SubcontractSampleApprovalResult{
			SubcontractOrder: buildResult.UpdatedOrder,
			SampleApproval:   buildResult.SampleApproval,
			PreviousStatus:   buildResult.PreviousOrderStatus,
			CurrentStatus:    buildResult.CurrentOrderStatus,
			AuditLogID:       log.ID,
		}

		return nil
	})
	if err != nil {
		return SubcontractSampleApprovalResult{}, err
	}

	return result, nil
}

func (s SubcontractOrderService) ApproveSubcontractSample(
	ctx context.Context,
	input DecideSubcontractSampleInput,
) (SubcontractSampleApprovalResult, error) {
	return s.decideSubcontractSample(ctx, input, subcontractSampleApprovedAction, func(
		buildInput BuildSubcontractSampleDecisionInput,
	) (SubcontractSampleApprovalBuildResult, error) {
		return s.sampleApprovalBuild.BuildApproval(ctx, buildInput)
	})
}

func (s SubcontractOrderService) RejectSubcontractSample(
	ctx context.Context,
	input DecideSubcontractSampleInput,
) (SubcontractSampleApprovalResult, error) {
	return s.decideSubcontractSample(ctx, input, subcontractSampleRejectedAction, func(
		buildInput BuildSubcontractSampleDecisionInput,
	) (SubcontractSampleApprovalBuildResult, error) {
		return s.sampleApprovalBuild.BuildRejection(ctx, buildInput)
	})
}

func (s SubcontractOrderService) ReceiveSubcontractFinishedGoods(
	ctx context.Context,
	input ReceiveSubcontractFinishedGoodsInput,
) (ReceiveSubcontractFinishedGoodsResult, error) {
	if err := s.ensureReadyForFinishedGoodsReceipt(); err != nil {
		return ReceiveSubcontractFinishedGoodsResult{}, err
	}
	if err := requireSubcontractOrderActor(input.ActorID); err != nil {
		return ReceiveSubcontractFinishedGoodsResult{}, err
	}

	var result ReceiveSubcontractFinishedGoodsResult
	err := s.store.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		current, err := tx.GetForUpdate(txCtx, input.ID)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{"subcontract_order_id": strings.TrimSpace(input.ID)})
		}
		if err := ensureSubcontractOrderExpectedVersion(current, input.ExpectedVersion); err != nil {
			return err
		}

		buildResult, err := s.finishedGoodsBuild.BuildReceipt(txCtx, BuildSubcontractFinishedGoodsReceiptInput{
			ID:             input.ReceiptID,
			ReceiptNo:      input.ReceiptNo,
			Order:          current,
			WarehouseID:    input.WarehouseID,
			WarehouseCode:  input.WarehouseCode,
			LocationID:     input.LocationID,
			LocationCode:   input.LocationCode,
			DeliveryNoteNo: input.DeliveryNoteNo,
			Lines:          receiveSubcontractFinishedGoodsBuildLineInputs(input.Lines),
			Evidence:       receiveSubcontractFinishedGoodsBuildEvidenceInputs(input.Evidence),
			ReceivedBy:     input.ReceivedBy,
			ReceivedAt:     input.ReceivedAt,
			Note:           input.Note,
			ActorID:        input.ActorID,
		})
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{
				"subcontract_order_id": current.ID,
				"status":               string(current.Status),
			})
		}
		if err := s.finishedGoodsStore.Save(txCtx, buildResult.Receipt); err != nil {
			return MapSubcontractOrderError(err, map[string]any{
				"subcontract_order_id": current.ID,
				"receipt_id":           buildResult.Receipt.ID,
			})
		}
		for _, movement := range buildResult.StockMovements {
			if err := s.finishedGoodsRecorder.Record(txCtx, movement); err != nil {
				return err
			}
		}
		if err := tx.Save(txCtx, buildResult.UpdatedOrder); err != nil {
			return err
		}
		afterData := subcontractOrderAuditData(buildResult.UpdatedOrder)
		for key, value := range subcontractFinishedGoodsReceiptAuditData(buildResult) {
			afterData[key] = value
		}
		log, err := newSubcontractOrderAuditLog(
			input.ActorID,
			input.RequestID,
			subcontractFinishedGoodsAction,
			buildResult.UpdatedOrder,
			subcontractOrderAuditData(current),
			afterData,
			buildResult.Receipt.ReceivedAt,
		)
		if err != nil {
			return err
		}
		if err := tx.RecordAudit(txCtx, log); err != nil {
			return err
		}
		result = ReceiveSubcontractFinishedGoodsResult{
			SubcontractOrder: buildResult.UpdatedOrder,
			Receipt:          buildResult.Receipt,
			StockMovements:   buildResult.StockMovements,
			AuditLogID:       log.ID,
		}

		return nil
	})
	if err != nil {
		return ReceiveSubcontractFinishedGoodsResult{}, err
	}

	return result, nil
}

func (s SubcontractOrderService) AcceptSubcontractFinishedGoods(
	ctx context.Context,
	input AcceptSubcontractFinishedGoodsInput,
) (AcceptSubcontractFinishedGoodsResult, error) {
	if err := s.ensureReadyForFinishedGoodsReceipt(); err != nil {
		return AcceptSubcontractFinishedGoodsResult{}, err
	}
	if err := requireSubcontractOrderActor(input.ActorID); err != nil {
		return AcceptSubcontractFinishedGoodsResult{}, err
	}

	var result AcceptSubcontractFinishedGoodsResult
	err := s.store.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		current, err := tx.GetForUpdate(txCtx, input.ID)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{"subcontract_order_id": strings.TrimSpace(input.ID)})
		}
		if err := ensureSubcontractOrderExpectedVersion(current, input.ExpectedVersion); err != nil {
			return err
		}
		receipts, err := s.finishedGoodsStore.ListBySubcontractOrder(txCtx, current.ID)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{"subcontract_order_id": current.ID})
		}
		if len(receipts) == 0 {
			return MapSubcontractOrderError(productiondomain.ErrSubcontractFinishedGoodsReceiptRequiredField, map[string]any{
				"subcontract_order_id": current.ID,
			})
		}

		acceptedAt := input.AcceptedAt
		if acceptedAt.IsZero() {
			acceptedAt = s.now()
		}
		acceptedBy := firstNonBlankSubcontractOrder(input.AcceptedBy, input.ActorID)
		acceptedOrder, err := current.AcceptFinishedGoods(input.ActorID, acceptedAt)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{
				"subcontract_order_id": current.ID,
				"status":               string(current.Status),
			})
		}
		movements, err := buildSubcontractFinishedGoodsAcceptanceMovements(receipts, input.ActorID, acceptedAt)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{"subcontract_order_id": current.ID})
		}
		for _, movement := range movements {
			if err := s.finishedGoodsRecorder.Record(txCtx, movement); err != nil {
				return err
			}
		}
		if err := tx.Save(txCtx, acceptedOrder); err != nil {
			return err
		}
		afterData := subcontractOrderAuditData(acceptedOrder)
		afterData["accepted_by"] = acceptedBy
		afterData["accepted_at"] = acceptedAt.Format(time.RFC3339)
		afterData["accepted_movement_count"] = len(movements)
		if strings.TrimSpace(input.Note) != "" {
			afterData["note"] = strings.TrimSpace(input.Note)
		}
		log, err := newSubcontractOrderAuditLog(
			input.ActorID,
			input.RequestID,
			subcontractFinishedGoodsAcceptedAction,
			acceptedOrder,
			subcontractOrderAuditData(current),
			afterData,
			acceptedAt,
		)
		if err != nil {
			return err
		}
		if err := tx.RecordAudit(txCtx, log); err != nil {
			return err
		}
		result = AcceptSubcontractFinishedGoodsResult{
			SubcontractOrder: acceptedOrder,
			StockMovements:   movements,
			PreviousStatus:   current.Status,
			CurrentStatus:    acceptedOrder.Status,
			AuditLogID:       log.ID,
		}

		return nil
	})
	if err != nil {
		return AcceptSubcontractFinishedGoodsResult{}, err
	}

	return result, nil
}

func (s SubcontractOrderService) PartialAcceptSubcontractFinishedGoods(
	ctx context.Context,
	input PartialAcceptSubcontractFinishedGoodsInput,
) (PartialAcceptSubcontractFinishedGoodsResult, error) {
	if err := s.ensureReadyForFinishedGoodsReceipt(); err != nil {
		return PartialAcceptSubcontractFinishedGoodsResult{}, err
	}
	if err := s.ensureReadyForFactoryClaim(); err != nil {
		return PartialAcceptSubcontractFinishedGoodsResult{}, err
	}
	if err := requireSubcontractOrderActor(input.ActorID); err != nil {
		return PartialAcceptSubcontractFinishedGoodsResult{}, err
	}

	var result PartialAcceptSubcontractFinishedGoodsResult
	err := s.store.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		current, err := tx.GetForUpdate(txCtx, input.ID)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{"subcontract_order_id": strings.TrimSpace(input.ID)})
		}
		if err := ensureSubcontractOrderExpectedVersion(current, input.ExpectedVersion); err != nil {
			return err
		}
		receipts, err := s.finishedGoodsStore.ListBySubcontractOrder(txCtx, current.ID)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{"subcontract_order_id": current.ID})
		}
		if len(receipts) == 0 {
			return MapSubcontractOrderError(productiondomain.ErrSubcontractFinishedGoodsReceiptRequiredField, map[string]any{
				"subcontract_order_id": current.ID,
			})
		}

		acceptedAt := input.AcceptedAt
		if acceptedAt.IsZero() {
			acceptedAt = s.now()
		}
		acceptedQty, err := decimal.ParseQuantity(input.AcceptedQty)
		if err != nil {
			return MapSubcontractOrderError(productiondomain.ErrSubcontractOrderInvalidQuantity, map[string]any{"field": "accepted_qty"})
		}
		baseAcceptedQty := decimal.Decimal(strings.TrimSpace(input.BaseAcceptedQty))
		baseRejectedQty := decimal.Decimal(strings.TrimSpace(input.BaseRejectedQty))
		rejectedQty, err := decimal.ParseQuantity(input.RejectedQty)
		if err != nil {
			return MapSubcontractOrderError(productiondomain.ErrSubcontractOrderInvalidQuantity, map[string]any{"field": "rejected_qty"})
		}
		acceptedOrder, err := current.PartialAcceptFinishedGoods(productiondomain.PartialAcceptSubcontractFinishedGoodsInput{
			AcceptedQty:     acceptedQty,
			UOMCode:         input.UOMCode,
			BaseAcceptedQty: baseAcceptedQty,
			BaseUOMCode:     input.BaseUOMCode,
			RejectedQty:     rejectedQty,
			BaseRejectedQty: baseRejectedQty,
			ActorID:         input.ActorID,
			ChangedAt:       acceptedAt,
		})
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{
				"subcontract_order_id": current.ID,
				"status":               string(current.Status),
			})
		}
		openedAt := input.OpenedAt
		if openedAt.IsZero() {
			openedAt = acceptedAt
		}
		claim, err := s.factoryClaimBuild.BuildClaimOnly(txCtx, BuildSubcontractFactoryClaimInput{
			ID:              input.ClaimID,
			ClaimNo:         input.ClaimNo,
			Order:           acceptedOrder,
			ReceiptID:       input.ReceiptID,
			ReceiptNo:       input.ReceiptNo,
			ReasonCode:      input.ReasonCode,
			Reason:          input.Reason,
			Severity:        input.Severity,
			AffectedQty:     input.RejectedQty,
			UOMCode:         input.UOMCode,
			BaseAffectedQty: input.BaseRejectedQty,
			BaseUOMCode:     input.BaseUOMCode,
			Evidence:        createSubcontractFactoryClaimBuildEvidenceInputs(input.Evidence),
			OwnerID:         input.OwnerID,
			OpenedBy:        input.OpenedBy,
			OpenedAt:        openedAt,
			ActorID:         input.ActorID,
		})
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{
				"subcontract_order_id": current.ID,
				"status":               string(current.Status),
			})
		}
		movements, err := buildSubcontractFinishedGoodsAcceptanceMovementsForQuantity(
			receipts,
			acceptedOrder.AcceptedQty,
			acceptedOrder.BaseAcceptedQty,
			input.ActorID,
			acceptedAt,
		)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{"subcontract_order_id": current.ID})
		}
		for _, movement := range movements {
			if err := s.finishedGoodsRecorder.Record(txCtx, movement); err != nil {
				return err
			}
		}
		if err := s.factoryClaimStore.Save(txCtx, claim); err != nil {
			return MapSubcontractOrderError(err, map[string]any{
				"subcontract_order_id": current.ID,
				"factory_claim_id":     claim.ID,
			})
		}
		if err := tx.Save(txCtx, acceptedOrder); err != nil {
			return err
		}

		acceptedBy := firstNonBlankSubcontractOrder(input.AcceptedBy, input.ActorID)
		acceptAfterData := subcontractOrderAuditData(acceptedOrder)
		acceptAfterData["accepted_by"] = acceptedBy
		acceptAfterData["accepted_at"] = acceptedAt.Format(time.RFC3339)
		acceptAfterData["accepted_movement_count"] = len(movements)
		if strings.TrimSpace(input.Note) != "" {
			acceptAfterData["note"] = strings.TrimSpace(input.Note)
		}
		acceptLog, err := newSubcontractOrderAuditLog(
			input.ActorID,
			input.RequestID,
			subcontractFinishedGoodsAcceptedAction,
			acceptedOrder,
			subcontractOrderAuditData(current),
			acceptAfterData,
			acceptedAt,
		)
		if err != nil {
			return err
		}
		if err := tx.RecordAudit(txCtx, acceptLog); err != nil {
			return err
		}

		claimAfterData := subcontractOrderAuditData(acceptedOrder)
		for key, value := range subcontractFactoryClaimAuditData(claim) {
			claimAfterData[key] = value
		}
		claimLog, err := newSubcontractOrderAuditLog(
			input.ActorID,
			input.RequestID,
			subcontractFactoryClaimAction,
			acceptedOrder,
			subcontractOrderAuditData(current),
			claimAfterData,
			claim.OpenedAt,
		)
		if err != nil {
			return err
		}
		if err := tx.RecordAudit(txCtx, claimLog); err != nil {
			return err
		}
		result = PartialAcceptSubcontractFinishedGoodsResult{
			SubcontractOrder: acceptedOrder,
			Claim:            claim,
			StockMovements:   movements,
			PreviousStatus:   current.Status,
			CurrentStatus:    acceptedOrder.Status,
			AcceptAuditLogID: acceptLog.ID,
			ClaimAuditLogID:  claimLog.ID,
		}

		return nil
	})
	if err != nil {
		return PartialAcceptSubcontractFinishedGoodsResult{}, err
	}

	return result, nil
}

func (s SubcontractOrderService) CreateSubcontractFactoryClaim(
	ctx context.Context,
	input CreateSubcontractFactoryClaimInput,
) (CreateSubcontractFactoryClaimResult, error) {
	if err := s.ensureReadyForFactoryClaim(); err != nil {
		return CreateSubcontractFactoryClaimResult{}, err
	}
	if err := requireSubcontractOrderActor(input.ActorID); err != nil {
		return CreateSubcontractFactoryClaimResult{}, err
	}

	var result CreateSubcontractFactoryClaimResult
	err := s.store.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		current, err := tx.GetForUpdate(txCtx, input.ID)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{"subcontract_order_id": strings.TrimSpace(input.ID)})
		}
		if err := ensureSubcontractOrderExpectedVersion(current, input.ExpectedVersion); err != nil {
			return err
		}

		buildResult, err := s.factoryClaimBuild.BuildClaim(txCtx, BuildSubcontractFactoryClaimInput{
			ID:              input.ClaimID,
			ClaimNo:         input.ClaimNo,
			Order:           current,
			ReceiptID:       input.ReceiptID,
			ReceiptNo:       input.ReceiptNo,
			ReasonCode:      input.ReasonCode,
			Reason:          input.Reason,
			Severity:        input.Severity,
			AffectedQty:     input.AffectedQty,
			UOMCode:         input.UOMCode,
			BaseAffectedQty: input.BaseAffectedQty,
			BaseUOMCode:     input.BaseUOMCode,
			Evidence:        createSubcontractFactoryClaimBuildEvidenceInputs(input.Evidence),
			OwnerID:         input.OwnerID,
			OpenedBy:        input.OpenedBy,
			OpenedAt:        input.OpenedAt,
			ActorID:         input.ActorID,
		})
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{
				"subcontract_order_id": current.ID,
				"status":               string(current.Status),
			})
		}
		if err := s.factoryClaimStore.Save(txCtx, buildResult.Claim); err != nil {
			return MapSubcontractOrderError(err, map[string]any{
				"subcontract_order_id": current.ID,
				"factory_claim_id":     buildResult.Claim.ID,
			})
		}
		if err := tx.Save(txCtx, buildResult.UpdatedOrder); err != nil {
			return err
		}
		afterData := subcontractOrderAuditData(buildResult.UpdatedOrder)
		for key, value := range subcontractFactoryClaimAuditData(buildResult.Claim) {
			afterData[key] = value
		}
		log, err := newSubcontractOrderAuditLog(
			input.ActorID,
			input.RequestID,
			subcontractFactoryClaimAction,
			buildResult.UpdatedOrder,
			subcontractOrderAuditData(current),
			afterData,
			buildResult.Claim.OpenedAt,
		)
		if err != nil {
			return err
		}
		if err := tx.RecordAudit(txCtx, log); err != nil {
			return err
		}
		result = CreateSubcontractFactoryClaimResult{
			SubcontractOrder: buildResult.UpdatedOrder,
			Claim:            buildResult.Claim,
			PreviousStatus:   buildResult.PreviousStatus,
			CurrentStatus:    buildResult.CurrentStatus,
			AuditLogID:       log.ID,
		}

		return nil
	})
	if err != nil {
		return CreateSubcontractFactoryClaimResult{}, err
	}

	return result, nil
}

func (s SubcontractOrderService) RecordSubcontractDeposit(
	ctx context.Context,
	input RecordSubcontractDepositInput,
) (SubcontractPaymentMilestoneResult, error) {
	if err := s.ensureReadyForPaymentMilestone(); err != nil {
		return SubcontractPaymentMilestoneResult{}, err
	}
	if err := requireSubcontractOrderActor(input.ActorID); err != nil {
		return SubcontractPaymentMilestoneResult{}, err
	}

	var result SubcontractPaymentMilestoneResult
	err := s.store.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		current, err := tx.GetForUpdate(txCtx, input.ID)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{"subcontract_order_id": strings.TrimSpace(input.ID)})
		}
		if err := ensureSubcontractOrderExpectedVersion(current, input.ExpectedVersion); err != nil {
			return err
		}

		buildResult, err := s.paymentMilestoneBuild.BuildDepositMilestone(txCtx, BuildSubcontractDepositMilestoneInput{
			ID:          input.MilestoneID,
			MilestoneNo: input.MilestoneNo,
			Order:       current,
			Amount:      input.Amount,
			RecordedBy:  input.RecordedBy,
			RecordedAt:  input.RecordedAt,
			ActorID:     input.ActorID,
			Note:        input.Note,
		})
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{
				"subcontract_order_id": current.ID,
				"status":               string(current.Status),
			})
		}
		if err := s.paymentMilestoneStore.Save(txCtx, buildResult.Milestone); err != nil {
			return MapSubcontractOrderError(err, map[string]any{
				"subcontract_order_id":       current.ID,
				"payment_milestone_id":       buildResult.Milestone.ID,
				"payment_milestone_kind":     string(buildResult.Milestone.Kind),
				"payment_milestone_currency": buildResult.Milestone.CurrencyCode.String(),
			})
		}
		if err := tx.Save(txCtx, buildResult.UpdatedOrder); err != nil {
			return err
		}
		afterData := subcontractOrderAuditData(buildResult.UpdatedOrder)
		for key, value := range subcontractPaymentMilestoneAuditData(buildResult.Milestone) {
			afterData[key] = value
		}
		log, err := newSubcontractOrderAuditLog(
			input.ActorID,
			input.RequestID,
			subcontractDepositRecordedAction,
			buildResult.UpdatedOrder,
			subcontractOrderAuditData(current),
			afterData,
			buildResult.Milestone.RecordedAt,
		)
		if err != nil {
			return err
		}
		if err := tx.RecordAudit(txCtx, log); err != nil {
			return err
		}
		result = SubcontractPaymentMilestoneResult{
			SubcontractOrder: buildResult.UpdatedOrder,
			Milestone:        buildResult.Milestone,
			PreviousStatus:   buildResult.PreviousStatus,
			CurrentStatus:    buildResult.CurrentStatus,
			AuditLogID:       log.ID,
		}

		return nil
	})
	if err != nil {
		return SubcontractPaymentMilestoneResult{}, err
	}

	return result, nil
}

func (s SubcontractOrderService) MarkSubcontractFinalPaymentReady(
	ctx context.Context,
	input MarkSubcontractFinalPaymentReadyInput,
) (SubcontractPaymentMilestoneResult, error) {
	if err := s.ensureReadyForFinalPaymentMilestone(); err != nil {
		return SubcontractPaymentMilestoneResult{}, err
	}
	if err := requireSubcontractOrderActor(input.ActorID); err != nil {
		return SubcontractPaymentMilestoneResult{}, err
	}

	var result SubcontractPaymentMilestoneResult
	err := s.store.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		current, err := tx.GetForUpdate(txCtx, input.ID)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{"subcontract_order_id": strings.TrimSpace(input.ID)})
		}
		if err := ensureSubcontractOrderExpectedVersion(current, input.ExpectedVersion); err != nil {
			return err
		}

		claims, err := s.factoryClaimStore.ListBySubcontractOrder(txCtx, current.ID)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{"subcontract_order_id": current.ID})
		}
		buildResult, err := s.paymentMilestoneBuild.BuildFinalPaymentMilestone(txCtx, BuildSubcontractFinalPaymentMilestoneInput{
			ID:                  input.MilestoneID,
			MilestoneNo:         input.MilestoneNo,
			Order:               current,
			Amount:              input.Amount,
			ReadyBy:             input.ReadyBy,
			ReadyAt:             input.ReadyAt,
			BlockingClaims:      claims,
			ApprovedExceptionID: input.ApprovedExceptionID,
			ActorID:             input.ActorID,
			Note:                input.Note,
		})
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{
				"subcontract_order_id": current.ID,
				"status":               string(current.Status),
				"blocking_claim_count": countBlockingSubcontractFactoryClaims(claims),
			})
		}
		if err := s.paymentMilestoneStore.Save(txCtx, buildResult.Milestone); err != nil {
			return MapSubcontractOrderError(err, map[string]any{
				"subcontract_order_id":   current.ID,
				"payment_milestone_id":   buildResult.Milestone.ID,
				"payment_milestone_kind": string(buildResult.Milestone.Kind),
			})
		}
		if err := tx.Save(txCtx, buildResult.UpdatedOrder); err != nil {
			return err
		}
		payableResult, err := s.createSubcontractPayableForFinalPayment(txCtx, buildResult.UpdatedOrder, buildResult.Milestone, input)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{
				"subcontract_order_id":   current.ID,
				"payment_milestone_id":   buildResult.Milestone.ID,
				"payment_milestone_kind": string(buildResult.Milestone.Kind),
			})
		}
		afterData := subcontractOrderAuditData(buildResult.UpdatedOrder)
		for key, value := range subcontractPaymentMilestoneAuditData(buildResult.Milestone) {
			afterData[key] = value
		}
		if payableResult.PayableID != "" {
			afterData["supplier_payable_id"] = payableResult.PayableID
			afterData["supplier_payable_no"] = payableResult.PayableNo
		}
		afterData["blocking_claim_count"] = countBlockingSubcontractFactoryClaims(claims)
		log, err := newSubcontractOrderAuditLog(
			input.ActorID,
			input.RequestID,
			subcontractFinalPaymentAction,
			buildResult.UpdatedOrder,
			subcontractOrderAuditData(current),
			afterData,
			buildResult.Milestone.ReadyAt,
		)
		if err != nil {
			return err
		}
		if err := tx.RecordAudit(txCtx, log); err != nil {
			return err
		}
		result = SubcontractPaymentMilestoneResult{
			SubcontractOrder: buildResult.UpdatedOrder,
			Milestone:        buildResult.Milestone,
			PreviousStatus:   buildResult.PreviousStatus,
			CurrentStatus:    buildResult.CurrentStatus,
			AuditLogID:       log.ID,
			SupplierPayable:  payableResult,
		}

		return nil
	})
	if err != nil {
		return SubcontractPaymentMilestoneResult{}, err
	}

	return result, nil
}

func (s SubcontractOrderService) createSubcontractPayableForFinalPayment(
	ctx context.Context,
	order productiondomain.SubcontractOrder,
	milestone productiondomain.SubcontractPaymentMilestone,
	input MarkSubcontractFinalPaymentReadyInput,
) (SubcontractPayableCreationResult, error) {
	if s.payableCreator == nil {
		return SubcontractPayableCreationResult{}, nil
	}

	return s.payableCreator.CreateSubcontractPayable(ctx, CreateSubcontractPayableInput{
		SubcontractOrder: order,
		Milestone:        milestone,
		ActorID:          input.ActorID,
		RequestID:        input.RequestID,
	})
}

func (s SubcontractOrderService) decideSubcontractSample(
	ctx context.Context,
	input DecideSubcontractSampleInput,
	action string,
	decide func(BuildSubcontractSampleDecisionInput) (SubcontractSampleApprovalBuildResult, error),
) (SubcontractSampleApprovalResult, error) {
	if err := s.ensureReadyForSampleApproval(); err != nil {
		return SubcontractSampleApprovalResult{}, err
	}
	if err := requireSubcontractOrderActor(input.ActorID); err != nil {
		return SubcontractSampleApprovalResult{}, err
	}

	var result SubcontractSampleApprovalResult
	err := s.store.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		current, err := tx.GetForUpdate(txCtx, input.ID)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{"subcontract_order_id": strings.TrimSpace(input.ID)})
		}
		if err := ensureSubcontractOrderExpectedVersion(current, input.ExpectedVersion); err != nil {
			return err
		}
		sampleApproval, err := s.getSubcontractSampleApprovalForDecision(txCtx, input.SampleApprovalID, current.ID)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{
				"subcontract_order_id": current.ID,
				"sample_approval_id":   input.SampleApprovalID,
			})
		}

		buildResult, err := decide(BuildSubcontractSampleDecisionInput{
			Order:          current,
			SampleApproval: sampleApproval,
			DecisionBy:     input.ActorID,
			DecisionAt:     input.DecisionAt,
			Reason:         input.Reason,
			StorageStatus:  input.StorageStatus,
			ActorID:        input.ActorID,
		})
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{
				"subcontract_order_id": current.ID,
				"status":               string(current.Status),
				"sample_approval_id":   sampleApproval.ID,
			})
		}
		if err := s.sampleApprovalStore.Save(txCtx, buildResult.SampleApproval); err != nil {
			return MapSubcontractOrderError(err, map[string]any{
				"subcontract_order_id": current.ID,
				"sample_approval_id":   buildResult.SampleApproval.ID,
			})
		}
		if err := tx.Save(txCtx, buildResult.UpdatedOrder); err != nil {
			return err
		}
		afterData := subcontractOrderAuditData(buildResult.UpdatedOrder)
		for key, value := range subcontractSampleApprovalAuditData(buildResult.SampleApproval) {
			afterData[key] = value
		}
		log, err := newSubcontractOrderAuditLog(
			input.ActorID,
			input.RequestID,
			action,
			buildResult.UpdatedOrder,
			subcontractOrderAuditData(current),
			afterData,
			buildResult.SampleApproval.UpdatedAt,
		)
		if err != nil {
			return err
		}
		if err := tx.RecordAudit(txCtx, log); err != nil {
			return err
		}
		result = SubcontractSampleApprovalResult{
			SubcontractOrder: buildResult.UpdatedOrder,
			SampleApproval:   buildResult.SampleApproval,
			PreviousStatus:   buildResult.PreviousOrderStatus,
			CurrentStatus:    buildResult.CurrentOrderStatus,
			AuditLogID:       log.ID,
		}

		return nil
	})
	if err != nil {
		return SubcontractSampleApprovalResult{}, err
	}

	return result, nil
}

func (s SubcontractOrderService) transition(
	ctx context.Context,
	input SubcontractOrderActionInput,
	action string,
	transition func(productiondomain.SubcontractOrder, string, time.Time) (productiondomain.SubcontractOrder, error),
) (SubcontractOrderActionResult, error) {
	if err := s.ensureReadyForWrite(); err != nil {
		return SubcontractOrderActionResult{}, err
	}
	if err := requireSubcontractOrderActor(input.ActorID); err != nil {
		return SubcontractOrderActionResult{}, err
	}

	var result SubcontractOrderActionResult
	err := s.store.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		current, err := tx.GetForUpdate(txCtx, input.ID)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{"subcontract_order_id": strings.TrimSpace(input.ID)})
		}
		if err := ensureSubcontractOrderExpectedVersion(current, input.ExpectedVersion); err != nil {
			return err
		}

		now := s.now()
		updated, err := transition(current, input.ActorID, now)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{
				"subcontract_order_id": current.ID,
				"status":               string(current.Status),
			})
		}
		if err := tx.Save(txCtx, updated); err != nil {
			return err
		}
		log, err := newSubcontractOrderAuditLog(
			input.ActorID,
			input.RequestID,
			action,
			updated,
			subcontractOrderAuditData(current),
			subcontractOrderAuditData(updated),
			now,
		)
		if err != nil {
			return err
		}
		if err := tx.RecordAudit(txCtx, log); err != nil {
			return err
		}
		result = SubcontractOrderActionResult{
			SubcontractOrder: updated,
			PreviousStatus:   current.Status,
			CurrentStatus:    updated.Status,
			AuditLogID:       log.ID,
		}

		return nil
	})
	if err != nil {
		return SubcontractOrderActionResult{}, err
	}

	return result, nil
}

type newSubcontractOrderDocumentServiceInput struct {
	ID                     string
	OrgID                  string
	OrderNo                string
	FactoryID              string
	FinishedItemID         string
	PlannedQty             string
	UOMCode                string
	CurrencyCode           string
	SpecSummary            string
	SourceProductionPlanID string
	SourceProductionPlanNo string
	SampleRequired         bool
	ClaimWindowDays        int
	TargetStartDate        string
	ExpectedReceiptDate    string
	MaterialLines          []SubcontractOrderMaterialLineInput
	CreatedAt              time.Time
	CreatedBy              string
	UpdatedAt              time.Time
	UpdatedBy              string
}

func (s SubcontractOrderService) newSubcontractOrderDocument(
	ctx context.Context,
	input newSubcontractOrderDocumentServiceInput,
) (productiondomain.SubcontractOrder, error) {
	factory, err := s.factoryRead.GetSupplier(ctx, input.FactoryID)
	if err != nil {
		return productiondomain.SubcontractOrder{}, mapSubcontractOrderMasterDataError(err, map[string]any{"factory_id": strings.TrimSpace(input.FactoryID)})
	}
	finishedItem, err := s.itemRead.Get(ctx, input.FinishedItemID)
	if err != nil {
		return productiondomain.SubcontractOrder{}, mapSubcontractOrderMasterDataError(err, map[string]any{"finished_item_id": strings.TrimSpace(input.FinishedItemID)})
	}
	plannedQty, err := decimal.ParseQuantity(input.PlannedQty)
	if err != nil || plannedQty.IsNegative() || plannedQty.IsZero() {
		return productiondomain.SubcontractOrder{}, subcontractOrderValidationError(
			productiondomain.ErrSubcontractOrderInvalidQuantity,
			map[string]any{"field": "planned_qty"},
		)
	}
	uomCode := firstNonBlankSubcontractOrder(input.UOMCode, finishedItem.UOMIssue, finishedItem.UOMBase)
	conversion, err := s.convertSubcontractOrderLineToBase(ctx, finishedItem, plannedQty, uomCode, 0, "planned_qty")
	if err != nil {
		return productiondomain.SubcontractOrder{}, err
	}
	lines, err := s.newSubcontractOrderMaterialLineInputs(ctx, input.ID, input.MaterialLines, input.CurrencyCode)
	if err != nil {
		return productiondomain.SubcontractOrder{}, err
	}

	order, err := productiondomain.NewSubcontractOrderDocument(productiondomain.NewSubcontractOrderDocumentInput{
		ID:                     input.ID,
		OrgID:                  input.OrgID,
		OrderNo:                input.OrderNo,
		FactoryID:              factory.ID,
		FactoryCode:            factory.Code,
		FactoryName:            factory.Name,
		FinishedItemID:         finishedItem.ID,
		FinishedSKUCode:        finishedItem.SKUCode,
		FinishedItemName:       finishedItem.Name,
		PlannedQty:             plannedQty,
		UOMCode:                conversion.SourceUOMCode.String(),
		BasePlannedQty:         conversion.BaseQuantity,
		BaseUOMCode:            conversion.BaseUOMCode.String(),
		ConversionFactor:       conversion.ConversionFactor,
		CurrencyCode:           input.CurrencyCode,
		SpecSummary:            input.SpecSummary,
		SourceProductionPlanID: input.SourceProductionPlanID,
		SourceProductionPlanNo: input.SourceProductionPlanNo,
		SampleRequired:         input.SampleRequired,
		ClaimWindowDays:        input.ClaimWindowDays,
		TargetStartDate:        input.TargetStartDate,
		ExpectedReceiptDate:    input.ExpectedReceiptDate,
		MaterialLines:          lines,
		CreatedAt:              input.CreatedAt,
		CreatedBy:              input.CreatedBy,
		UpdatedAt:              input.UpdatedAt,
		UpdatedBy:              input.UpdatedBy,
	})
	if err != nil {
		return productiondomain.SubcontractOrder{}, MapSubcontractOrderError(err, nil)
	}

	return order, nil
}

func (s SubcontractOrderService) newSubcontractOrderMaterialLineInputs(
	ctx context.Context,
	orderID string,
	inputs []SubcontractOrderMaterialLineInput,
	currencyCode string,
) ([]productiondomain.NewSubcontractMaterialLineInput, error) {
	if len(inputs) == 0 {
		return nil, subcontractOrderValidationError(productiondomain.ErrSubcontractOrderRequiredField, map[string]any{"field": "material_lines"})
	}

	lines := make([]productiondomain.NewSubcontractMaterialLineInput, 0, len(inputs))
	for index, input := range inputs {
		item, err := s.itemRead.Get(ctx, input.ItemID)
		if err != nil {
			return nil, mapSubcontractOrderMasterDataError(err, map[string]any{"item_id": strings.TrimSpace(input.ItemID)})
		}
		lineNo := input.LineNo
		if lineNo == 0 {
			lineNo = index + 1
		}
		lineID := strings.TrimSpace(input.ID)
		if lineID == "" {
			lineID = fmt.Sprintf("%s-material-%02d", strings.TrimSpace(orderID), lineNo)
		}
		plannedQty, err := decimal.ParseQuantity(input.PlannedQty)
		if err != nil || plannedQty.IsNegative() || plannedQty.IsZero() {
			return nil, subcontractOrderValidationError(
				productiondomain.ErrSubcontractOrderInvalidQuantity,
				map[string]any{"line_no": lineNo, "field": "planned_qty"},
			)
		}
		uomCode := firstNonBlankSubcontractOrder(input.UOMCode, item.UOMIssue, item.UOMBase)
		conversion, err := s.convertSubcontractOrderLineToBase(ctx, item, plannedQty, uomCode, lineNo, "uom_code")
		if err != nil {
			return nil, err
		}
		unitCostValue := firstNonBlankSubcontractOrder(input.UnitCost, item.StandardCost.String(), "0")
		unitCost, err := decimal.ParseUnitCost(unitCostValue)
		if err != nil || unitCost.IsNegative() {
			return nil, subcontractOrderValidationError(
				productiondomain.ErrSubcontractOrderInvalidAmount,
				map[string]any{"line_no": lineNo, "field": "unit_cost"},
			)
		}
		lineCurrency := firstNonBlankSubcontractOrder(input.CurrencyCode, currencyCode)
		if !strings.EqualFold(strings.TrimSpace(lineCurrency), decimal.CurrencyVND.String()) {
			return nil, subcontractOrderValidationError(
				productiondomain.ErrSubcontractOrderInvalidCurrency,
				map[string]any{"line_no": lineNo, "field": "currency_code"},
			)
		}

		lines = append(lines, productiondomain.NewSubcontractMaterialLineInput{
			ID:               lineID,
			LineNo:           lineNo,
			ItemID:           item.ID,
			SKUCode:          item.SKUCode,
			ItemName:         item.Name,
			PlannedQty:       plannedQty,
			IssuedQty:        decimal.MustQuantity("0"),
			UOMCode:          conversion.SourceUOMCode.String(),
			BasePlannedQty:   conversion.BaseQuantity,
			BaseIssuedQty:    decimal.MustQuantity("0"),
			BaseUOMCode:      conversion.BaseUOMCode.String(),
			ConversionFactor: conversion.ConversionFactor,
			UnitCost:         unitCost,
			CurrencyCode:     lineCurrency,
			LotTraceRequired: input.LotTraceRequired,
			Note:             input.Note,
		})
	}

	return lines, nil
}

func (s SubcontractOrderService) convertSubcontractOrderLineToBase(
	ctx context.Context,
	item masterdatadomain.Item,
	quantity decimal.Decimal,
	uomCode string,
	lineNo int,
	field string,
) (ConvertSubcontractOrderLineToBaseResult, error) {
	if strings.EqualFold(strings.TrimSpace(uomCode), strings.TrimSpace(item.UOMBase)) {
		sourceUOM, err := decimal.NormalizeUOMCode(uomCode)
		if err != nil {
			return ConvertSubcontractOrderLineToBaseResult{}, subcontractOrderValidationError(
				productiondomain.ErrSubcontractOrderInvalidQuantity,
				map[string]any{"line_no": lineNo, "field": field},
			)
		}
		baseUOM, err := decimal.NormalizeUOMCode(item.UOMBase)
		if err != nil {
			return ConvertSubcontractOrderLineToBaseResult{}, subcontractOrderValidationError(
				productiondomain.ErrSubcontractOrderInvalidQuantity,
				map[string]any{"line_no": lineNo, "field": "base_uom_code"},
			)
		}

		return ConvertSubcontractOrderLineToBaseResult{
			Quantity:         quantity,
			SourceUOMCode:    sourceUOM,
			BaseQuantity:     quantity,
			BaseUOMCode:      baseUOM,
			ConversionFactor: decimal.MustQuantity("1"),
		}, nil
	}

	conversion, err := s.uomConverter.ConvertToBase(ctx, ConvertSubcontractOrderLineToBaseInput{
		ItemID:      item.ID,
		SKU:         item.SKUCode,
		Quantity:    quantity,
		FromUOMCode: uomCode,
		BaseUOMCode: item.UOMBase,
	})
	if err != nil {
		return ConvertSubcontractOrderLineToBaseResult{}, mapSubcontractOrderMasterDataError(err, map[string]any{
			"line_no":       lineNo,
			"item_id":       item.ID,
			"sku_code":      item.SKUCode,
			"from_uom_code": strings.ToUpper(strings.TrimSpace(uomCode)),
			"base_uom_code": strings.ToUpper(strings.TrimSpace(item.UOMBase)),
		})
	}

	return conversion, nil
}

func (s SubcontractOrderService) ensureReadyForWrite() error {
	if s.store == nil {
		return errors.New("subcontract order store is required")
	}
	if s.factoryRead == nil {
		return errors.New("subcontract order factory reader is required")
	}
	if s.itemRead == nil {
		return errors.New("subcontract order item reader is required")
	}
	if s.uomConverter == nil {
		return errors.New("subcontract order uom converter is required")
	}

	return nil
}

func (s SubcontractOrderService) ensureReadyForMaterialIssue() error {
	if s.store == nil {
		return errors.New("subcontract order store is required")
	}
	if s.materialTransferStore == nil {
		return errors.New("subcontract material transfer store is required")
	}
	if s.materialIssueRecorder == nil {
		return errors.New("subcontract material issue movement recorder is required")
	}
	if s.materialTransferBuild.clock == nil {
		s.materialTransferBuild = NewSubcontractMaterialTransferService()
	}

	return nil
}

func (s SubcontractOrderService) ensureReadyForSampleApproval() error {
	if s.store == nil {
		return errors.New("subcontract order store is required")
	}
	if s.sampleApprovalStore == nil {
		return errors.New("subcontract sample approval store is required")
	}
	if s.sampleApprovalBuild.clock == nil {
		s.sampleApprovalBuild = NewSubcontractSampleApprovalService()
	}

	return nil
}

func (s SubcontractOrderService) ensureReadyForFinishedGoodsReceipt() error {
	if s.store == nil {
		return errors.New("subcontract order store is required")
	}
	if s.finishedGoodsStore == nil {
		return errors.New("subcontract finished goods receipt store is required")
	}
	if s.finishedGoodsRecorder == nil {
		return errors.New("subcontract finished goods receipt movement recorder is required")
	}
	if s.finishedGoodsBuild.clock == nil {
		s.finishedGoodsBuild = NewSubcontractFinishedGoodsReceiptService()
	}

	return nil
}

func (s SubcontractOrderService) ensureReadyForFactoryClaim() error {
	if s.store == nil {
		return errors.New("subcontract order store is required")
	}
	if s.factoryClaimStore == nil {
		return errors.New("subcontract factory claim store is required")
	}
	if s.factoryClaimBuild.clock == nil {
		s.factoryClaimBuild = NewSubcontractFactoryClaimService()
	}

	return nil
}

func (s SubcontractOrderService) ensureReadyForPaymentMilestone() error {
	if s.store == nil {
		return errors.New("subcontract order store is required")
	}
	if s.paymentMilestoneStore == nil {
		return errors.New("subcontract payment milestone store is required")
	}
	if s.paymentMilestoneBuild.clock == nil {
		s.paymentMilestoneBuild = NewSubcontractPaymentMilestoneService()
	}

	return nil
}

func (s SubcontractOrderService) ensureReadyForFinalPaymentMilestone() error {
	if err := s.ensureReadyForPaymentMilestone(); err != nil {
		return err
	}
	if s.factoryClaimStore == nil {
		return errors.New("subcontract factory claim store is required")
	}

	return nil
}

func (s SubcontractOrderService) now() time.Time {
	if s.clock == nil {
		return time.Now().UTC()
	}

	return s.clock().UTC()
}

func (s SubcontractOrderService) getSubcontractSampleApprovalForDecision(
	ctx context.Context,
	sampleApprovalID string,
	subcontractOrderID string,
) (productiondomain.SubcontractSampleApproval, error) {
	if strings.TrimSpace(sampleApprovalID) != "" {
		return s.sampleApprovalStore.Get(ctx, sampleApprovalID)
	}

	return s.sampleApprovalStore.GetLatestBySubcontractOrder(ctx, subcontractOrderID)
}

func (s *PrototypeSubcontractOrderStore) List(
	_ context.Context,
	filter SubcontractOrderFilter,
) ([]productiondomain.SubcontractOrder, error) {
	if s == nil {
		return nil, errors.New("subcontract order store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows := make([]productiondomain.SubcontractOrder, 0, len(s.records))
	for _, order := range s.records {
		if subcontractOrderMatchesFilter(order, filter) {
			rows = append(rows, order.Clone())
		}
	}
	sortSubcontractOrders(rows)

	return rows, nil
}

func (s *PrototypeSubcontractOrderStore) Get(
	_ context.Context,
	id string,
) (productiondomain.SubcontractOrder, error) {
	if s == nil {
		return productiondomain.SubcontractOrder{}, errors.New("subcontract order store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	order, ok := s.records[strings.TrimSpace(id)]
	if !ok {
		return productiondomain.SubcontractOrder{}, ErrSubcontractOrderNotFound
	}

	return order.Clone(), nil
}

func (s *PrototypeSubcontractOrderStore) WithinTx(
	ctx context.Context,
	fn func(context.Context, SubcontractOrderTx) error,
) error {
	if s == nil {
		return errors.New("subcontract order store is required")
	}
	if fn == nil {
		return errors.New("subcontract order transaction function is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot := cloneSubcontractOrderMap(s.records)
	tx := &prototypeSubcontractOrderTx{store: s}
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

func (s *PrototypeSubcontractOrderStore) TransactionCount() int {
	if s == nil {
		return 0
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.txCount
}

func (tx *prototypeSubcontractOrderTx) GetForUpdate(
	_ context.Context,
	id string,
) (productiondomain.SubcontractOrder, error) {
	order, ok := tx.store.records[strings.TrimSpace(id)]
	if !ok {
		return productiondomain.SubcontractOrder{}, ErrSubcontractOrderNotFound
	}

	return order.Clone(), nil
}

func (tx *prototypeSubcontractOrderTx) Save(
	_ context.Context,
	order productiondomain.SubcontractOrder,
) error {
	if strings.TrimSpace(order.ID) == "" {
		return productiondomain.ErrSubcontractOrderRequiredField
	}
	tx.store.records[order.ID] = order.Clone()

	return nil
}

func (tx *prototypeSubcontractOrderTx) RecordAudit(
	_ context.Context,
	log audit.Log,
) error {
	tx.auditLogs = append(tx.auditLogs, log)

	return nil
}

func (s *PrototypeSubcontractSampleApprovalStore) Save(
	_ context.Context,
	sampleApproval productiondomain.SubcontractSampleApproval,
) error {
	if s == nil {
		return errors.New("subcontract sample approval store is required")
	}
	if strings.TrimSpace(sampleApproval.ID) == "" {
		return productiondomain.ErrSubcontractSampleApprovalRequiredField
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.records[sampleApproval.ID] = sampleApproval.Clone()

	return nil
}

func (s *PrototypeSubcontractSampleApprovalStore) Get(
	_ context.Context,
	id string,
) (productiondomain.SubcontractSampleApproval, error) {
	if s == nil {
		return productiondomain.SubcontractSampleApproval{}, errors.New("subcontract sample approval store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	sampleApproval, ok := s.records[strings.TrimSpace(id)]
	if !ok {
		return productiondomain.SubcontractSampleApproval{}, ErrSubcontractSampleApprovalNotFound
	}

	return sampleApproval.Clone(), nil
}

func (s *PrototypeSubcontractSampleApprovalStore) GetLatestBySubcontractOrder(
	_ context.Context,
	subcontractOrderID string,
) (productiondomain.SubcontractSampleApproval, error) {
	if s == nil {
		return productiondomain.SubcontractSampleApproval{}, errors.New("subcontract sample approval store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var latest productiondomain.SubcontractSampleApproval
	for _, sampleApproval := range s.records {
		if sampleApproval.SubcontractOrderID != strings.TrimSpace(subcontractOrderID) {
			continue
		}
		if latest.ID == "" || sampleApproval.CreatedAt.After(latest.CreatedAt) {
			latest = sampleApproval
		}
	}
	if latest.ID == "" {
		return productiondomain.SubcontractSampleApproval{}, ErrSubcontractSampleApprovalNotFound
	}

	return latest.Clone(), nil
}

func (s *PrototypeSubcontractSampleApprovalStore) Count() int {
	if s == nil {
		return 0
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.records)
}

func ensureSubcontractOrderExpectedVersion(order productiondomain.SubcontractOrder, expectedVersion int) error {
	if expectedVersion <= 0 || order.Version == expectedVersion {
		return nil
	}

	return apperrors.Conflict(
		ErrorCodeSubcontractOrderVersionConflict,
		"Subcontract order version changed",
		ErrSubcontractOrderVersionConflict,
		map[string]any{
			"subcontract_order_id": order.ID,
			"expected_version":     expectedVersion,
			"current_version":      order.Version,
		},
	)
}

func requireSubcontractOrderActor(actorID string) error {
	if strings.TrimSpace(actorID) == "" {
		return subcontractOrderValidationError(productiondomain.ErrSubcontractOrderTransitionActorRequired, map[string]any{"field": "actor_id"})
	}

	return nil
}

func subcontractMaterialIssueLineInputs(
	inputs []IssueSubcontractMaterialsLineInput,
) []BuildSubcontractMaterialTransferLineInput {
	lines := make([]BuildSubcontractMaterialTransferLineInput, 0, len(inputs))
	for _, input := range inputs {
		lines = append(lines, BuildSubcontractMaterialTransferLineInput{
			ID:                  input.ID,
			LineNo:              input.LineNo,
			OrderMaterialLineID: input.OrderMaterialLineID,
			IssueQty:            input.IssueQty,
			UOMCode:             input.UOMCode,
			BaseIssueQty:        input.BaseIssueQty,
			BaseUOMCode:         input.BaseUOMCode,
			ConversionFactor:    input.ConversionFactor,
			BatchID:             input.BatchID,
			BatchNo:             input.BatchNo,
			LotNo:               input.LotNo,
			SourceBinID:         input.SourceBinID,
			Note:                input.Note,
		})
	}

	return lines
}

func subcontractMaterialIssueEvidenceInputs(
	inputs []IssueSubcontractMaterialsEvidenceInput,
) []BuildSubcontractMaterialTransferEvidenceInput {
	evidence := make([]BuildSubcontractMaterialTransferEvidenceInput, 0, len(inputs))
	for _, input := range inputs {
		evidence = append(evidence, BuildSubcontractMaterialTransferEvidenceInput{
			ID:           input.ID,
			EvidenceType: input.EvidenceType,
			FileName:     input.FileName,
			ObjectKey:    input.ObjectKey,
			ExternalURL:  input.ExternalURL,
			Note:         input.Note,
		})
	}

	return evidence
}

func subcontractSampleBuildEvidenceInputs(
	inputs []SubcontractSampleEvidenceInput,
) []BuildSubcontractSampleEvidenceInput {
	evidence := make([]BuildSubcontractSampleEvidenceInput, 0, len(inputs))
	for _, input := range inputs {
		evidence = append(evidence, BuildSubcontractSampleEvidenceInput{
			ID:           input.ID,
			EvidenceType: input.EvidenceType,
			FileName:     input.FileName,
			ObjectKey:    input.ObjectKey,
			ExternalURL:  input.ExternalURL,
			Note:         input.Note,
		})
	}

	return evidence
}

func receiveSubcontractFinishedGoodsBuildLineInputs(
	inputs []ReceiveSubcontractFinishedGoodsLineInput,
) []BuildSubcontractFinishedGoodsReceiptLineInput {
	lines := make([]BuildSubcontractFinishedGoodsReceiptLineInput, 0, len(inputs))
	for _, input := range inputs {
		lines = append(lines, BuildSubcontractFinishedGoodsReceiptLineInput{
			ID:               input.ID,
			LineNo:           input.LineNo,
			ItemID:           input.ItemID,
			SKUCode:          input.SKUCode,
			ItemName:         input.ItemName,
			BatchID:          input.BatchID,
			BatchNo:          input.BatchNo,
			LotNo:            input.LotNo,
			ExpiryDate:       input.ExpiryDate,
			ReceiveQty:       input.ReceiveQty,
			UOMCode:          input.UOMCode,
			BaseReceiveQty:   input.BaseReceiveQty,
			BaseUOMCode:      input.BaseUOMCode,
			ConversionFactor: input.ConversionFactor,
			PackagingStatus:  input.PackagingStatus,
			Note:             input.Note,
		})
	}

	return lines
}

func receiveSubcontractFinishedGoodsBuildEvidenceInputs(
	inputs []ReceiveSubcontractFinishedGoodsEvidenceInput,
) []BuildSubcontractFinishedGoodsReceiptEvidenceInput {
	evidence := make([]BuildSubcontractFinishedGoodsReceiptEvidenceInput, 0, len(inputs))
	for _, input := range inputs {
		evidence = append(evidence, BuildSubcontractFinishedGoodsReceiptEvidenceInput{
			ID:           input.ID,
			EvidenceType: input.EvidenceType,
			FileName:     input.FileName,
			ObjectKey:    input.ObjectKey,
			ExternalURL:  input.ExternalURL,
			Note:         input.Note,
		})
	}

	return evidence
}

func countBlockingSubcontractFactoryClaims(claims []productiondomain.SubcontractFactoryClaim) int {
	count := 0
	for _, claim := range claims {
		if claim.BlocksFinalPayment() {
			count++
		}
	}

	return count
}

func createSubcontractFactoryClaimBuildEvidenceInputs(
	inputs []CreateSubcontractFactoryClaimEvidenceInput,
) []BuildSubcontractFactoryClaimEvidenceInput {
	evidence := make([]BuildSubcontractFactoryClaimEvidenceInput, 0, len(inputs))
	for _, input := range inputs {
		evidence = append(evidence, BuildSubcontractFactoryClaimEvidenceInput{
			ID:           input.ID,
			EvidenceType: input.EvidenceType,
			FileName:     input.FileName,
			ObjectKey:    input.ObjectKey,
			ExternalURL:  input.ExternalURL,
			Note:         input.Note,
		})
	}

	return evidence
}

func subcontractOrderValidationError(cause error, details map[string]any) error {
	return apperrors.BadRequest(ErrorCodeSubcontractOrderValidation, "Subcontract order request is invalid", cause, details)
}

func MapSubcontractOrderError(err error, details map[string]any) error {
	if err == nil {
		return nil
	}
	if _, ok := apperrors.As(err); ok {
		return err
	}
	if errors.Is(err, ErrSubcontractOrderNotFound) {
		return apperrors.NotFound(ErrorCodeSubcontractOrderNotFound, "Subcontract order not found", err, details)
	}
	if errors.Is(err, ErrSubcontractSampleApprovalNotFound) {
		return apperrors.NotFound(ErrorCodeSubcontractOrderNotFound, "Subcontract sample approval not found", err, details)
	}
	if errors.Is(err, ErrSubcontractFactoryClaimNotFound) {
		return apperrors.NotFound(ErrorCodeSubcontractOrderNotFound, "Subcontract factory claim not found", err, details)
	}
	if errors.Is(err, ErrSubcontractPaymentMilestoneNotFound) {
		return apperrors.NotFound(ErrorCodeSubcontractOrderNotFound, "Subcontract payment milestone not found", err, details)
	}
	if errors.Is(err, productiondomain.ErrSubcontractOrderInvalidTransition) ||
		errors.Is(err, productiondomain.ErrSubcontractOrderInvalidStatus) ||
		errors.Is(err, productiondomain.ErrSubcontractOrderSampleApprovalRequired) ||
		errors.Is(err, productiondomain.ErrSubcontractSampleApprovalInvalidStatus) ||
		errors.Is(err, productiondomain.ErrSubcontractSampleApprovalInvalidTransition) ||
		errors.Is(err, productiondomain.ErrSubcontractFactoryClaimInvalidStatus) ||
		errors.Is(err, productiondomain.ErrSubcontractPaymentMilestoneInvalidStatus) ||
		errors.Is(err, productiondomain.ErrSubcontractPaymentMilestoneInvalidTransition) ||
		errors.Is(err, productiondomain.ErrSubcontractPaymentMilestoneBlocked) {
		return apperrors.Conflict(ErrorCodeSubcontractOrderInvalidState, "Subcontract order state is invalid", err, details)
	}
	if errors.Is(err, productiondomain.ErrSubcontractMaterialTransferInvalidStatus) {
		return apperrors.Conflict(ErrorCodeSubcontractOrderInvalidState, "Subcontract material transfer state is invalid", err, details)
	}
	if errors.Is(err, productiondomain.ErrSubcontractOrderRequiredField) ||
		errors.Is(err, productiondomain.ErrSubcontractOrderTransitionActorRequired) ||
		errors.Is(err, productiondomain.ErrSubcontractOrderInvalidCurrency) ||
		errors.Is(err, productiondomain.ErrSubcontractOrderInvalidQuantity) ||
		errors.Is(err, productiondomain.ErrSubcontractOrderInvalidAmount) ||
		errors.Is(err, productiondomain.ErrSubcontractOrderMaterialLineNotFound) ||
		errors.Is(err, productiondomain.ErrSubcontractOrderDuplicateMaterialLine) ||
		errors.Is(err, productiondomain.ErrSubcontractSampleApprovalRequiredField) ||
		errors.Is(err, productiondomain.ErrSubcontractMaterialTransferRequiredField) ||
		errors.Is(err, productiondomain.ErrSubcontractMaterialTransferInvalidQuantity) ||
		errors.Is(err, productiondomain.ErrSubcontractMaterialTransferBatchRequired) ||
		errors.Is(err, productiondomain.ErrSubcontractFinishedGoodsReceiptRequiredField) ||
		errors.Is(err, productiondomain.ErrSubcontractFinishedGoodsReceiptInvalidQuantity) ||
		errors.Is(err, productiondomain.ErrSubcontractFinishedGoodsReceiptInvalidStatus) ||
		errors.Is(err, productiondomain.ErrSubcontractFactoryClaimRequiredField) ||
		errors.Is(err, productiondomain.ErrSubcontractFactoryClaimInvalidQuantity) ||
		errors.Is(err, productiondomain.ErrSubcontractFactoryClaimInvalidSLA) ||
		errors.Is(err, productiondomain.ErrSubcontractPaymentMilestoneRequiredField) ||
		errors.Is(err, productiondomain.ErrSubcontractPaymentMilestoneInvalidKind) ||
		errors.Is(err, productiondomain.ErrSubcontractPaymentMilestoneInvalidCurrency) ||
		errors.Is(err, productiondomain.ErrSubcontractPaymentMilestoneInvalidAmount) {
		return subcontractOrderValidationError(err, details)
	}

	return err
}

func mapSubcontractOrderMasterDataError(err error, details map[string]any) error {
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

func newSubcontractOrderAuditLog(
	actorID string,
	requestID string,
	action string,
	order productiondomain.SubcontractOrder,
	beforeData map[string]any,
	afterData map[string]any,
	createdAt time.Time,
) (audit.Log, error) {
	return audit.NewLog(audit.NewLogInput{
		OrgID:      firstNonBlankSubcontractOrder(order.OrgID, defaultSubcontractOrderOrgID),
		ActorID:    strings.TrimSpace(actorID),
		Action:     strings.TrimSpace(action),
		EntityType: subcontractOrderEntityType,
		EntityID:   order.ID,
		RequestID:  strings.TrimSpace(requestID),
		BeforeData: beforeData,
		AfterData:  afterData,
		Metadata: map[string]any{
			"source":        "subcontract order application service",
			"order_no":      order.OrderNo,
			"factory_code":  order.FactoryCode,
			"finished_sku":  order.FinishedSKUCode,
			"sample_needed": order.SampleRequired,
		},
		CreatedAt: createdAt,
	})
}

func subcontractOrderAuditData(order productiondomain.SubcontractOrder) map[string]any {
	data := map[string]any{
		"order_no":                  order.OrderNo,
		"factory_id":                order.FactoryID,
		"factory_code":              order.FactoryCode,
		"finished_item_id":          order.FinishedItemID,
		"finished_sku_code":         order.FinishedSKUCode,
		"planned_qty":               order.PlannedQty.String(),
		"received_qty":              order.ReceivedQty.String(),
		"accepted_qty":              order.AcceptedQty.String(),
		"rejected_qty":              order.RejectedQty.String(),
		"base_planned_qty":          order.BasePlannedQty.String(),
		"base_received_qty":         order.BaseReceivedQty.String(),
		"base_accepted_qty":         order.BaseAcceptedQty.String(),
		"base_rejected_qty":         order.BaseRejectedQty.String(),
		"uom_code":                  order.UOMCode.String(),
		"base_uom_code":             order.BaseUOMCode.String(),
		"expected_receipt_date":     order.ExpectedReceiptDate,
		"source_production_plan_id": order.SourceProductionPlanID,
		"source_production_plan_no": order.SourceProductionPlanNo,
		"status":                    string(order.Status),
		"currency_code":             order.CurrencyCode.String(),
		"estimated_cost_amount":     order.EstimatedCostAmount.String(),
		"deposit_amount":            order.DepositAmount.String(),
		"sample_required":           order.SampleRequired,
		"claim_window_days":         order.ClaimWindowDays,
		"material_line_count":       len(order.MaterialLines),
		"version":                   order.Version,
	}
	if strings.TrimSpace(order.CancelReason) != "" {
		data["cancel_reason"] = order.CancelReason
	}
	if strings.TrimSpace(order.SampleRejectReason) != "" {
		data["sample_reject_reason"] = order.SampleRejectReason
	}
	if strings.TrimSpace(order.FactoryIssueReason) != "" {
		data["factory_issue_reason"] = order.FactoryIssueReason
	}

	return data
}

func subcontractFinishedGoodsReceiptAuditData(result SubcontractFinishedGoodsReceiptBuildResult) map[string]any {
	return map[string]any{
		"receipt_id":           result.Receipt.ID,
		"receipt_no":           result.Receipt.ReceiptNo,
		"receipt_status":       string(result.Receipt.Status),
		"warehouse_id":         result.Receipt.WarehouseID,
		"location_id":          result.Receipt.LocationID,
		"delivery_note_no":     result.Receipt.DeliveryNoteNo,
		"received_by":          result.Receipt.ReceivedBy,
		"received_at":          result.Receipt.ReceivedAt.Format(time.RFC3339),
		"receipt_line_count":   len(result.Receipt.Lines),
		"stock_movement_count": len(result.StockMovements),
	}
}

func subcontractFactoryClaimAuditData(claim productiondomain.SubcontractFactoryClaim) map[string]any {
	return map[string]any{
		"factory_claim_id":     claim.ID,
		"factory_claim_no":     claim.ClaimNo,
		"factory_claim_status": string(claim.Status),
		"receipt_id":           claim.ReceiptID,
		"receipt_no":           claim.ReceiptNo,
		"reason_code":          claim.ReasonCode,
		"reason":               claim.Reason,
		"severity":             claim.Severity,
		"affected_qty":         claim.AffectedQty.String(),
		"base_affected_qty":    claim.BaseAffectedQty.String(),
		"uom_code":             claim.UOMCode.String(),
		"base_uom_code":        claim.BaseUOMCode.String(),
		"owner_id":             claim.OwnerID,
		"opened_by":            claim.OpenedBy,
		"opened_at":            claim.OpenedAt.Format(time.RFC3339),
		"due_at":               claim.DueAt.Format(time.RFC3339),
		"evidence_count":       len(claim.Evidence),
		"blocks_final_payment": claim.BlocksFinalPayment(),
	}
}

func subcontractPaymentMilestoneAuditData(milestone productiondomain.SubcontractPaymentMilestone) map[string]any {
	return map[string]any{
		"payment_milestone_id":           milestone.ID,
		"payment_milestone_no":           milestone.MilestoneNo,
		"payment_milestone_kind":         string(milestone.Kind),
		"payment_milestone_status":       string(milestone.Status),
		"payment_milestone_amount":       milestone.Amount.String(),
		"payment_milestone_currency":     milestone.CurrencyCode.String(),
		"approved_exception_id":          milestone.ApprovedExceptionID,
		"payment_milestone_blocks_final": milestone.BlocksFinalPayment(),
	}
}

func subcontractMaterialIssueAuditData(result SubcontractMaterialTransferBuildResult) map[string]any {
	return map[string]any{
		"transfer_id":           result.Transfer.ID,
		"transfer_no":           result.Transfer.TransferNo,
		"transfer_status":       string(result.Transfer.Status),
		"source_warehouse_id":   result.Transfer.SourceWarehouseID,
		"source_warehouse_code": result.Transfer.SourceWarehouseCode,
		"handover_by":           result.Transfer.HandoverBy,
		"received_by":           result.Transfer.ReceivedBy,
		"transfer_line_count":   len(result.Transfer.Lines),
		"stock_movement_count":  len(result.StockMovements),
	}
}

func subcontractSampleApprovalAuditData(sampleApproval productiondomain.SubcontractSampleApproval) map[string]any {
	data := map[string]any{
		"sample_approval_id": sampleApproval.ID,
		"sample_code":        sampleApproval.SampleCode,
		"sample_status":      string(sampleApproval.Status),
		"evidence_count":     len(sampleApproval.Evidence),
		"submitted_by":       sampleApproval.SubmittedBy,
	}
	if strings.TrimSpace(sampleApproval.DecisionBy) != "" {
		data["decision_by"] = sampleApproval.DecisionBy
	}
	if strings.TrimSpace(sampleApproval.DecisionReason) != "" {
		data["decision_reason"] = sampleApproval.DecisionReason
	}
	if strings.TrimSpace(sampleApproval.StorageStatus) != "" {
		data["storage_status"] = sampleApproval.StorageStatus
	}

	return data
}

func subcontractOrderMatchesFilter(order productiondomain.SubcontractOrder, filter SubcontractOrderFilter) bool {
	search := strings.ToLower(strings.TrimSpace(filter.Search))
	if search != "" {
		haystack := strings.ToLower(strings.Join([]string{
			order.OrderNo,
			order.FactoryCode,
			order.FactoryName,
			order.FinishedSKUCode,
			order.FinishedItemName,
			order.SourceProductionPlanNo,
			order.SpecSummary,
		}, " "))
		if !strings.Contains(haystack, search) {
			return false
		}
	}
	if len(filter.Statuses) > 0 {
		matched := false
		for _, status := range filter.Statuses {
			if order.Status == productiondomain.NormalizeSubcontractOrderStatus(status) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if strings.TrimSpace(filter.FactoryID) != "" && order.FactoryID != strings.TrimSpace(filter.FactoryID) {
		return false
	}
	if strings.TrimSpace(filter.FinishedItemID) != "" && order.FinishedItemID != strings.TrimSpace(filter.FinishedItemID) {
		return false
	}
	if strings.TrimSpace(filter.SourceProductionPlanID) != "" &&
		!strings.EqualFold(order.SourceProductionPlanID, strings.TrimSpace(filter.SourceProductionPlanID)) {
		return false
	}
	if strings.TrimSpace(filter.ExpectedReceiptFrom) != "" && order.ExpectedReceiptDate < strings.TrimSpace(filter.ExpectedReceiptFrom) {
		return false
	}
	if strings.TrimSpace(filter.ExpectedReceiptTo) != "" && order.ExpectedReceiptDate > strings.TrimSpace(filter.ExpectedReceiptTo) {
		return false
	}

	return true
}

func sortSubcontractOrders(orders []productiondomain.SubcontractOrder) {
	sort.SliceStable(orders, func(i, j int) bool {
		if orders[i].ExpectedReceiptDate == orders[j].ExpectedReceiptDate {
			return orders[i].OrderNo < orders[j].OrderNo
		}

		return orders[i].ExpectedReceiptDate > orders[j].ExpectedReceiptDate
	})
}

func cloneSubcontractOrderMap(records map[string]productiondomain.SubcontractOrder) map[string]productiondomain.SubcontractOrder {
	clone := make(map[string]productiondomain.SubcontractOrder, len(records))
	for id, order := range records {
		clone[id] = order.Clone()
	}

	return clone
}

func subcontractOrderMaterialLineInputsFromDomain(
	lines []productiondomain.SubcontractMaterialLine,
) []SubcontractOrderMaterialLineInput {
	inputs := make([]SubcontractOrderMaterialLineInput, 0, len(lines))
	for _, line := range lines {
		inputs = append(inputs, SubcontractOrderMaterialLineInput{
			ID:               line.ID,
			LineNo:           line.LineNo,
			ItemID:           line.ItemID,
			PlannedQty:       line.PlannedQty.String(),
			UOMCode:          line.UOMCode.String(),
			UnitCost:         line.UnitCost.String(),
			CurrencyCode:     line.CurrencyCode.String(),
			LotTraceRequired: line.LotTraceRequired,
			Note:             line.Note,
		})
	}

	return inputs
}

func newSubcontractOrderID(now time.Time) string {
	return fmt.Sprintf("sco-%s-%06d", now.UTC().Format("060102"), now.UTC().UnixNano()%1000000)
}

func newSubcontractOrderNo(now time.Time) string {
	return fmt.Sprintf("SCO-%s-%06d", now.UTC().Format("060102"), now.UTC().UnixNano()%1000000)
}

func firstNonBlankSubcontractOrder(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}

	return ""
}

func firstNonZeroSubcontractOrder(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}

	return 0
}
