package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	masterdatadomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

var ErrProductionPlanNotFound = errors.New("production plan not found")
var ErrProductionPlanFormulaNotFound = errors.New("production plan formula not found")
var ErrProductionPlanFormulaInactive = errors.New("production plan formula must be active")
var ErrPurchaseRequestDraftNotFound = errors.New("purchase request draft not found")
var ErrProductionPlanLineNotFound = errors.New("production plan line not found")
var ErrProductionPlanMaterialIssueNotReady = errors.New("production plan material is not ready to issue")

const (
	ErrorCodeProductionPlanNotFound              response.ErrorCode = "PRODUCTION_PLAN_NOT_FOUND"
	ErrorCodeProductionPlanValidation            response.ErrorCode = "PRODUCTION_PLAN_VALIDATION_ERROR"
	ErrorCodePurchaseRequestNotFound             response.ErrorCode = "PURCHASE_REQUEST_NOT_FOUND"
	ErrorCodePurchaseRequestInvalid              response.ErrorCode = "PURCHASE_REQUEST_INVALID_STATE"
	ErrorCodeProductionPlanMaterialIssueNotReady response.ErrorCode = "PRODUCTION_PLAN_MATERIAL_ISSUE_NOT_READY"

	defaultProductionPlanOrgID = "org-my-pham"
	productionPlanEntityType   = "production.plan"
	purchaseRequestEntityType  = "purchase.request"
)

type ProductionPlanStore interface {
	List(ctx context.Context, filter ProductionPlanFilter) ([]productiondomain.ProductionPlan, error)
	Get(ctx context.Context, id string) (productiondomain.ProductionPlan, error)
	Save(ctx context.Context, plan productiondomain.ProductionPlan) error
	RecordAudit(ctx context.Context, log audit.Log) error
}

type ProductionPlanFormulaReader interface {
	List(ctx context.Context, filter masterdatadomain.FormulaFilter) ([]masterdatadomain.Formula, error)
	Get(ctx context.Context, id string) (masterdatadomain.Formula, error)
}

type ProductionPlanAvailableStockLister interface {
	Execute(ctx context.Context, filter inventorydomain.AvailableStockFilter) ([]inventorydomain.AvailableStockSnapshot, error)
}

type ProductionPlanWarehouseIssueService interface {
	ListWarehouseIssues(ctx context.Context) ([]inventorydomain.WarehouseIssue, error)
	CreateWarehouseIssue(ctx context.Context, input inventoryapp.CreateWarehouseIssueInput) (inventoryapp.WarehouseIssueResult, error)
}

type ProductionPlanService struct {
	store          ProductionPlanStore
	formulaRead    ProductionPlanFormulaReader
	availableStock ProductionPlanAvailableStockLister
	warehouseIssue ProductionPlanWarehouseIssueService
	clock          func() time.Time
}

type ProductionPlanFilter struct {
	Search       string
	Statuses     []productiondomain.ProductionPlanStatus
	OutputItemID string
}

type PurchaseRequestDraftFilter struct {
	Search                 string
	Statuses               []productiondomain.PurchaseRequestDraftStatus
	SourceProductionPlanID string
}

type CreateProductionPlanInput struct {
	ID               string
	OrgID            string
	PlanNo           string
	OutputItemID     string
	FormulaID        string
	PlannedQty       string
	UOMCode          string
	PlannedStartDate string
	PlannedEndDate   string
	ActorID          string
	RequestID        string
}

type ProductionPlanResult struct {
	ProductionPlan productiondomain.ProductionPlan
	AuditLogID     string
}

type PurchaseRequestDraftActionInput struct {
	ID             string
	ExpectedStatus productiondomain.PurchaseRequestDraftStatus
	ActorID        string
	RequestID      string
}

type ConvertPurchaseRequestDraftInput struct {
	ID              string
	PurchaseOrderID string
	PurchaseOrderNo string
	ActorID         string
	RequestID       string
}

type PurchaseRequestDraftResult struct {
	PurchaseRequestDraft productiondomain.PurchaseRequestDraft
	PreviousStatus       productiondomain.PurchaseRequestDraftStatus
	CurrentStatus        productiondomain.PurchaseRequestDraftStatus
	AuditLogID           string
}

type CreateProductionPlanWarehouseIssueInput struct {
	PlanID          string
	LineIDs         []string
	WarehouseID     string
	WarehouseCode   string
	DestinationType string
	DestinationName string
	ReasonCode      string
	ActorID         string
	RequestID       string
}

type ProductionPlanWarehouseIssueResult struct {
	WarehouseIssue inventorydomain.WarehouseIssue
	AuditLogID     string
}

type PrototypeProductionPlanStore struct {
	mu       sync.RWMutex
	records  map[string]productiondomain.ProductionPlan
	auditLog audit.LogStore
}

func NewProductionPlanService(
	store ProductionPlanStore,
	formulaRead ProductionPlanFormulaReader,
	availableStock ProductionPlanAvailableStockLister,
) ProductionPlanService {
	return ProductionPlanService{
		store:          store,
		formulaRead:    formulaRead,
		availableStock: availableStock,
		clock:          func() time.Time { return time.Now().UTC() },
	}
}

func (s ProductionPlanService) WithWarehouseIssueService(service ProductionPlanWarehouseIssueService) ProductionPlanService {
	s.warehouseIssue = service

	return s
}

func NewPrototypeProductionPlanStore(auditLog audit.LogStore) *PrototypeProductionPlanStore {
	return &PrototypeProductionPlanStore{
		records:  make(map[string]productiondomain.ProductionPlan),
		auditLog: auditLog,
	}
}

func (s ProductionPlanService) ListProductionPlans(ctx context.Context, filter ProductionPlanFilter) ([]productiondomain.ProductionPlan, error) {
	if s.store == nil {
		return nil, errors.New("production plan store is required")
	}

	plans, err := s.store.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	for index := range plans {
		enriched, err := s.enrichProductionPlanIssueReadiness(ctx, plans[index])
		if err != nil {
			return nil, err
		}
		plans[index] = enriched
	}

	return plans, nil
}

func (s ProductionPlanService) GetProductionPlan(ctx context.Context, id string) (productiondomain.ProductionPlan, error) {
	if s.store == nil {
		return productiondomain.ProductionPlan{}, errors.New("production plan store is required")
	}
	plan, err := s.store.Get(ctx, id)
	if err != nil {
		return productiondomain.ProductionPlan{}, err
	}

	return s.enrichProductionPlanIssueReadiness(ctx, plan)
}

func (s ProductionPlanService) ListPurchaseRequestDrafts(
	ctx context.Context,
	filter PurchaseRequestDraftFilter,
) ([]productiondomain.PurchaseRequestDraft, error) {
	if s.store == nil {
		return nil, errors.New("production plan store is required")
	}
	plans, err := s.store.List(ctx, ProductionPlanFilter{})
	if err != nil {
		return nil, err
	}
	drafts := make([]productiondomain.PurchaseRequestDraft, 0, len(plans))
	for _, plan := range plans {
		if len(plan.PurchaseDraft.Lines) == 0 {
			continue
		}
		if purchaseRequestDraftMatchesFilter(plan.PurchaseDraft, filter) {
			drafts = append(drafts, plan.PurchaseDraft.Clone())
		}
	}
	sort.SliceStable(drafts, func(i, j int) bool {
		return drafts[i].CreatedAt.After(drafts[j].CreatedAt)
	})

	return drafts, nil
}

func (s ProductionPlanService) GetPurchaseRequestDraft(
	ctx context.Context,
	id string,
) (productiondomain.PurchaseRequestDraft, error) {
	plan, err := s.findPlanByPurchaseRequestDraft(ctx, id)
	if err != nil {
		return productiondomain.PurchaseRequestDraft{}, err
	}

	return plan.PurchaseDraft.Clone(), nil
}

func (s ProductionPlanService) SubmitPurchaseRequestDraft(
	ctx context.Context,
	input PurchaseRequestDraftActionInput,
) (PurchaseRequestDraftResult, error) {
	return s.transitionPurchaseRequestDraft(ctx, input, "purchase.request.submitted", func(
		draft productiondomain.PurchaseRequestDraft,
		actorID string,
		changedAt time.Time,
	) (productiondomain.PurchaseRequestDraft, error) {
		return draft.Submit(actorID, changedAt)
	})
}

func (s ProductionPlanService) ApprovePurchaseRequestDraft(
	ctx context.Context,
	input PurchaseRequestDraftActionInput,
) (PurchaseRequestDraftResult, error) {
	return s.transitionPurchaseRequestDraft(ctx, input, "purchase.request.approved", func(
		draft productiondomain.PurchaseRequestDraft,
		actorID string,
		changedAt time.Time,
	) (productiondomain.PurchaseRequestDraft, error) {
		return draft.Approve(actorID, changedAt)
	})
}

func (s ProductionPlanService) MarkPurchaseRequestDraftConverted(
	ctx context.Context,
	input ConvertPurchaseRequestDraftInput,
) (PurchaseRequestDraftResult, error) {
	if strings.TrimSpace(input.PurchaseOrderID) == "" || strings.TrimSpace(input.PurchaseOrderNo) == "" {
		return PurchaseRequestDraftResult{}, productiondomain.ErrProductionPlanRequiredField
	}

	return s.transitionPurchaseRequestDraft(ctx, PurchaseRequestDraftActionInput{
		ID:        input.ID,
		ActorID:   input.ActorID,
		RequestID: input.RequestID,
	}, "purchase.request.converted_to_po", func(
		draft productiondomain.PurchaseRequestDraft,
		actorID string,
		changedAt time.Time,
	) (productiondomain.PurchaseRequestDraft, error) {
		return draft.MarkConvertedToPO(actorID, changedAt, input.PurchaseOrderID, input.PurchaseOrderNo)
	})
}

func (s ProductionPlanService) CreateProductionPlan(ctx context.Context, input CreateProductionPlanInput) (ProductionPlanResult, error) {
	if s.store == nil {
		return ProductionPlanResult{}, errors.New("production plan store is required")
	}
	if s.formulaRead == nil {
		return ProductionPlanResult{}, errors.New("production plan formula reader is required")
	}
	if s.availableStock == nil {
		return ProductionPlanResult{}, errors.New("production plan stock reader is required")
	}
	actorID := strings.TrimSpace(input.ActorID)
	if actorID == "" {
		return ProductionPlanResult{}, productiondomain.ErrProductionPlanRequiredField
	}

	now := s.now()
	orgID := strings.TrimSpace(input.OrgID)
	if orgID == "" {
		orgID = defaultProductionPlanOrgID
	}
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = newProductionPlanID(now)
	}
	planNo := strings.TrimSpace(input.PlanNo)
	if planNo == "" {
		planNo = newProductionPlanNo(now)
	}
	formula, err := s.resolveFormula(ctx, input)
	if err != nil {
		return ProductionPlanResult{}, err
	}
	plannedQty, err := decimal.ParseQuantity(input.PlannedQty)
	if err != nil {
		return ProductionPlanResult{}, productiondomain.ErrProductionPlanInvalidQuantity
	}
	requirements, err := formula.CalculateRequirement(plannedQty, input.UOMCode)
	if err != nil {
		return ProductionPlanResult{}, err
	}
	lines, err := s.productionPlanLinesFromRequirements(ctx, id, formula, requirements)
	if err != nil {
		return ProductionPlanResult{}, err
	}

	plan, err := productiondomain.NewProductionPlanDocument(productiondomain.NewProductionPlanDocumentInput{
		ID:                  id,
		OrgID:               orgID,
		PlanNo:              planNo,
		OutputItemID:        formula.FinishedItemID,
		OutputSKU:           formula.FinishedSKU,
		OutputItemName:      formula.FinishedItemName,
		OutputItemType:      string(formula.FinishedItemType),
		PlannedQty:          plannedQty,
		UOMCode:             input.UOMCode,
		FormulaID:           formula.ID,
		FormulaCode:         formula.FormulaCode,
		FormulaVersion:      formula.FormulaVersion,
		FormulaBatchQty:     formula.BatchQty,
		FormulaBatchUOMCode: formula.BatchUOMCode.String(),
		PlannedStartDate:    input.PlannedStartDate,
		PlannedEndDate:      input.PlannedEndDate,
		Lines:               lines,
		PurchaseDraftID:     newPurchaseRequestDraftID(now),
		PurchaseDraftNo:     newPurchaseRequestDraftNo(now),
		CreatedAt:           now,
		CreatedBy:           actorID,
	})
	if err != nil {
		return ProductionPlanResult{}, err
	}

	if err := s.store.Save(ctx, plan); err != nil {
		return ProductionPlanResult{}, err
	}
	log, err := newProductionPlanAuditLog(actorID, input.RequestID, "production.plan.created", plan, now)
	if err != nil {
		return ProductionPlanResult{}, err
	}
	if err := s.store.RecordAudit(ctx, log); err != nil {
		return ProductionPlanResult{}, err
	}

	return ProductionPlanResult{ProductionPlan: plan, AuditLogID: log.ID}, nil
}

func (s ProductionPlanService) CreateWarehouseIssueFromProductionPlan(
	ctx context.Context,
	input CreateProductionPlanWarehouseIssueInput,
) (ProductionPlanWarehouseIssueResult, error) {
	if s.store == nil {
		return ProductionPlanWarehouseIssueResult{}, errors.New("production plan store is required")
	}
	if s.warehouseIssue == nil {
		return ProductionPlanWarehouseIssueResult{}, errors.New("warehouse issue service is required")
	}
	actorID := strings.TrimSpace(input.ActorID)
	if actorID == "" || len(input.LineIDs) == 0 {
		return ProductionPlanWarehouseIssueResult{}, productiondomain.ErrProductionPlanRequiredField
	}
	rawPlan, err := s.store.Get(ctx, input.PlanID)
	if err != nil {
		return ProductionPlanWarehouseIssueResult{}, err
	}
	plan, err := s.enrichProductionPlanIssueReadiness(ctx, rawPlan)
	if err != nil {
		return ProductionPlanWarehouseIssueResult{}, err
	}

	linesByID := make(map[string]productiondomain.ProductionPlanLine, len(plan.Lines))
	for _, line := range plan.Lines {
		linesByID[line.ID] = line
	}
	issueLines := make([]inventoryapp.CreateWarehouseIssueLineInput, 0, len(input.LineIDs))
	var selectedWarehouseID string
	var selectedWarehouseCode string
	for _, requestedLineID := range input.LineIDs {
		line, ok := linesByID[strings.TrimSpace(requestedLineID)]
		if !ok {
			return ProductionPlanWarehouseIssueResult{}, ErrProductionPlanLineNotFound
		}
		if !isIssueReadyProductionPlanLine(line) {
			return ProductionPlanWarehouseIssueResult{}, ErrProductionPlanMaterialIssueNotReady
		}
		allocations, err := s.allocateProductionPlanIssueStock(ctx, line, firstNonBlankProductionPlan(input.WarehouseID, selectedWarehouseID))
		if err != nil {
			return ProductionPlanWarehouseIssueResult{}, err
		}
		if len(allocations) == 0 {
			return ProductionPlanWarehouseIssueResult{}, ErrProductionPlanMaterialIssueNotReady
		}
		if selectedWarehouseID == "" {
			selectedWarehouseID = allocations[0].WarehouseID
			selectedWarehouseCode = allocations[0].WarehouseCode
		}
		for _, allocation := range allocations {
			issueLines = append(issueLines, inventoryapp.CreateWarehouseIssueLineInput{
				ItemID:               firstNonBlankProductionPlan(line.ComponentItemID, allocation.ItemID),
				SKU:                  line.ComponentSKU,
				ItemName:             line.ComponentName,
				Category:             line.ComponentType,
				BatchID:              allocation.BatchID,
				BatchNo:              allocation.BatchNo,
				LocationID:           allocation.LocationID,
				LocationCode:         allocation.LocationCode,
				Quantity:             allocation.Quantity.String(),
				BaseUOMCode:          line.StockBaseUOMCode.String(),
				SourceDocumentType:   "production_plan",
				SourceDocumentID:     plan.ID,
				SourceDocumentLineID: line.ID,
				Note:                 fmt.Sprintf("From %s line %d", plan.PlanNo, line.LineNo),
			})
		}
	}
	if selectedWarehouseID == "" {
		return ProductionPlanWarehouseIssueResult{}, ErrProductionPlanMaterialIssueNotReady
	}

	destinationType := firstNonBlankProductionPlan(input.DestinationType, "factory")
	destinationName := firstNonBlankProductionPlan(input.DestinationName, "Factory")
	reasonCode := firstNonBlankProductionPlan(input.ReasonCode, "production_plan_issue")
	result, err := s.warehouseIssue.CreateWarehouseIssue(ctx, inventoryapp.CreateWarehouseIssueInput{
		OrgID:           plan.OrgID,
		WarehouseID:     selectedWarehouseID,
		WarehouseCode:   firstNonBlankProductionPlan(input.WarehouseCode, selectedWarehouseCode),
		DestinationType: destinationType,
		DestinationName: destinationName,
		ReasonCode:      reasonCode,
		RequestedBy:     actorID,
		RequestID:       input.RequestID,
		Lines:           issueLines,
	})
	if err != nil {
		return ProductionPlanWarehouseIssueResult{}, err
	}

	return ProductionPlanWarehouseIssueResult{
		WarehouseIssue: result.WarehouseIssue,
		AuditLogID:     result.AuditLogID,
	}, nil
}

func (s ProductionPlanService) transitionPurchaseRequestDraft(
	ctx context.Context,
	input PurchaseRequestDraftActionInput,
	action string,
	transition func(productiondomain.PurchaseRequestDraft, string, time.Time) (productiondomain.PurchaseRequestDraft, error),
) (PurchaseRequestDraftResult, error) {
	if s.store == nil {
		return PurchaseRequestDraftResult{}, errors.New("production plan store is required")
	}
	actorID := strings.TrimSpace(input.ActorID)
	if actorID == "" {
		return PurchaseRequestDraftResult{}, productiondomain.ErrProductionPlanRequiredField
	}
	plan, err := s.findPlanByPurchaseRequestDraft(ctx, input.ID)
	if err != nil {
		return PurchaseRequestDraftResult{}, err
	}
	before := plan.PurchaseDraft.Clone()
	if expected := productiondomain.NormalizePurchaseRequestDraftStatus(input.ExpectedStatus); expected != "" && before.Status != expected {
		return PurchaseRequestDraftResult{}, productiondomain.ErrProductionPlanInvalidPurchaseRequestTransition
	}
	now := s.now()
	after, err := transition(before, actorID, now)
	if err != nil {
		return PurchaseRequestDraftResult{}, err
	}
	updatedPlan := plan.Clone()
	updatedPlan.PurchaseDraft = after
	updatedPlan.UpdatedAt = now
	updatedPlan.UpdatedBy = actorID
	if updatedPlan.Version > 0 {
		updatedPlan.Version++
	}
	if err := s.store.Save(ctx, updatedPlan); err != nil {
		return PurchaseRequestDraftResult{}, err
	}
	log, err := newPurchaseRequestDraftAuditLog(actorID, input.RequestID, action, updatedPlan, before, after, updatedPlan.UpdatedAt)
	if err != nil {
		return PurchaseRequestDraftResult{}, err
	}
	if err := s.store.RecordAudit(ctx, log); err != nil {
		return PurchaseRequestDraftResult{}, err
	}

	return PurchaseRequestDraftResult{
		PurchaseRequestDraft: after,
		PreviousStatus:       before.Status,
		CurrentStatus:        after.Status,
		AuditLogID:           log.ID,
	}, nil
}

func (s ProductionPlanService) findPlanByPurchaseRequestDraft(ctx context.Context, id string) (productiondomain.ProductionPlan, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return productiondomain.ProductionPlan{}, ErrPurchaseRequestDraftNotFound
	}
	plans, err := s.store.List(ctx, ProductionPlanFilter{})
	if err != nil {
		return productiondomain.ProductionPlan{}, err
	}
	for _, plan := range plans {
		draft := plan.PurchaseDraft
		if strings.EqualFold(draft.ID, id) || strings.EqualFold(draft.RequestNo, id) {
			return plan.Clone(), nil
		}
	}

	return productiondomain.ProductionPlan{}, ErrPurchaseRequestDraftNotFound
}

func (s ProductionPlanService) resolveFormula(ctx context.Context, input CreateProductionPlanInput) (masterdatadomain.Formula, error) {
	formulaID := strings.TrimSpace(input.FormulaID)
	var (
		formula masterdatadomain.Formula
		err     error
	)
	if formulaID != "" {
		formula, err = s.formulaRead.Get(ctx, formulaID)
	} else {
		formulas, listErr := s.formulaRead.List(ctx, masterdatadomain.FormulaFilter{
			FinishedItemID: strings.TrimSpace(input.OutputItemID),
			Status:         masterdatadomain.FormulaStatusActive,
		})
		if listErr != nil {
			return masterdatadomain.Formula{}, listErr
		}
		if len(formulas) == 0 {
			return masterdatadomain.Formula{}, ErrProductionPlanFormulaNotFound
		}
		formula = formulas[0]
	}
	if err != nil {
		return masterdatadomain.Formula{}, err
	}
	if formula.Status != masterdatadomain.FormulaStatusActive {
		return masterdatadomain.Formula{}, ErrProductionPlanFormulaInactive
	}
	if formulaID != "" {
		matchesOutput, err := s.formulaMatchesRequestedOutput(ctx, formula, input.OutputItemID)
		if err != nil {
			return masterdatadomain.Formula{}, err
		}
		if !matchesOutput {
			return masterdatadomain.Formula{}, productiondomain.ErrProductionPlanInvalidOutputType
		}
	}

	return formula.Clone(), nil
}

func (s ProductionPlanService) formulaMatchesRequestedOutput(
	ctx context.Context,
	formula masterdatadomain.Formula,
	outputItemID string,
) (bool, error) {
	outputItemID = strings.TrimSpace(outputItemID)
	if outputItemID == "" || strings.EqualFold(outputItemID, formula.FinishedItemID) || strings.EqualFold(outputItemID, formula.FinishedSKU) {
		return true, nil
	}
	formulas, err := s.formulaRead.List(ctx, masterdatadomain.FormulaFilter{
		FinishedItemID: outputItemID,
		Status:         masterdatadomain.FormulaStatusActive,
	})
	if err != nil {
		return false, err
	}
	for _, candidate := range formulas {
		if strings.EqualFold(candidate.ID, formula.ID) {
			return true, nil
		}
	}

	return false, nil
}

func (s ProductionPlanService) productionPlanLinesFromRequirements(
	ctx context.Context,
	planID string,
	formula masterdatadomain.Formula,
	requirements []masterdatadomain.FormulaRequirement,
) ([]productiondomain.NewProductionPlanLineInput, error) {
	formulaLinesByID := make(map[string]masterdatadomain.FormulaLine, len(formula.Lines))
	for _, line := range formula.Lines {
		formulaLinesByID[line.ID] = line
	}
	lines := make([]productiondomain.NewProductionPlanLineInput, 0, len(requirements))
	for _, requirement := range requirements {
		formulaLine := formulaLinesByID[requirement.FormulaLineID]
		availableQty, err := s.availableQty(ctx, requirement)
		if err != nil {
			return nil, err
		}
		shortageQty, err := shortageQuantity(requirement.RequiredStockBaseQty, availableQty)
		if err != nil {
			return nil, err
		}
		needsPurchase := requirement.IsStockManaged && !shortageQty.IsZero()
		purchaseDraftQty := decimal.MustQuantity("0")
		if needsPurchase {
			purchaseDraftQty = shortageQty
		}
		lines = append(lines, productiondomain.NewProductionPlanLineInput{
			ID:                   newProductionPlanLineID(planID, requirement.LineNo),
			FormulaLineID:        requirement.FormulaLineID,
			LineNo:               requirement.LineNo,
			ComponentItemID:      requirement.ComponentItemID,
			ComponentSKU:         requirement.ComponentSKU,
			ComponentName:        requirement.ComponentName,
			ComponentType:        string(requirement.ComponentType),
			FormulaQty:           formulaLine.CalcQty,
			FormulaUOMCode:       formulaLine.CalcUOMCode.String(),
			RequiredQty:          requirement.RequiredCalcQty,
			RequiredUOMCode:      requirement.CalcUOMCode.String(),
			RequiredStockBaseQty: requirement.RequiredStockBaseQty,
			StockBaseUOMCode:     requirement.StockBaseUOMCode.String(),
			AvailableQty:         availableQty,
			ShortageQty:          shortageQty,
			PurchaseDraftQty:     purchaseDraftQty,
			PurchaseDraftUOMCode: requirement.StockBaseUOMCode.String(),
			IsStockManaged:       requirement.IsStockManaged,
			NeedsPurchase:        needsPurchase,
		})
	}

	return lines, nil
}

func (s ProductionPlanService) availableQty(ctx context.Context, requirement masterdatadomain.FormulaRequirement) (decimal.Decimal, error) {
	snapshots, err := s.availableStock.Execute(ctx, inventorydomain.NewAvailableStockFilter("", "", requirement.ComponentSKU, ""))
	if err != nil {
		return "", err
	}
	total := decimal.MustQuantity("0")
	for _, snapshot := range snapshots {
		if snapshot.BaseUOMCode != requirement.StockBaseUOMCode {
			continue
		}
		next, err := decimal.AddQuantity(total, snapshot.AvailableQty)
		if err != nil {
			return "", err
		}
		total = next
	}

	return total, nil
}

func (s ProductionPlanService) now() time.Time {
	if s.clock == nil {
		return time.Now().UTC()
	}

	return s.clock().UTC()
}

func (s *PrototypeProductionPlanStore) List(
	_ context.Context,
	filter ProductionPlanFilter,
) ([]productiondomain.ProductionPlan, error) {
	if s == nil {
		return nil, errors.New("production plan store is required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	plans := make([]productiondomain.ProductionPlan, 0, len(s.records))
	for _, plan := range s.records {
		if productionPlanMatchesFilter(plan, filter) {
			plans = append(plans, plan.Clone())
		}
	}
	sort.Slice(plans, func(i, j int) bool {
		return plans[i].CreatedAt.After(plans[j].CreatedAt)
	})

	return plans, nil
}

func (s *PrototypeProductionPlanStore) Get(_ context.Context, id string) (productiondomain.ProductionPlan, error) {
	if s == nil {
		return productiondomain.ProductionPlan{}, errors.New("production plan store is required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	plan, ok := s.records[strings.TrimSpace(id)]
	if !ok {
		return productiondomain.ProductionPlan{}, ErrProductionPlanNotFound
	}

	return plan.Clone(), nil
}

func (s *PrototypeProductionPlanStore) Save(_ context.Context, plan productiondomain.ProductionPlan) error {
	if s == nil {
		return errors.New("production plan store is required")
	}
	if err := plan.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records[plan.ID] = plan.Clone()

	return nil
}

func (s *PrototypeProductionPlanStore) RecordAudit(ctx context.Context, log audit.Log) error {
	if s == nil {
		return errors.New("production plan store is required")
	}
	if s.auditLog == nil {
		return errors.New("audit log store is required")
	}

	return s.auditLog.Record(ctx, log)
}

func (s *PrototypeProductionPlanStore) Count() int {
	if s == nil {
		return 0
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.records)
}

func productionPlanMatchesFilter(plan productiondomain.ProductionPlan, filter ProductionPlanFilter) bool {
	if strings.TrimSpace(filter.OutputItemID) != "" && !strings.EqualFold(plan.OutputItemID, filter.OutputItemID) {
		return false
	}
	if len(filter.Statuses) > 0 {
		matched := false
		for _, status := range filter.Statuses {
			if plan.Status == productiondomain.NormalizeProductionPlanStatus(status) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	search := strings.ToLower(strings.TrimSpace(filter.Search))
	if search == "" {
		return true
	}

	return strings.Contains(strings.ToLower(plan.PlanNo), search) ||
		strings.Contains(strings.ToLower(plan.OutputSKU), search) ||
		strings.Contains(strings.ToLower(plan.OutputItemName), search)
}

func purchaseRequestDraftMatchesFilter(draft productiondomain.PurchaseRequestDraft, filter PurchaseRequestDraftFilter) bool {
	if strings.TrimSpace(filter.SourceProductionPlanID) != "" &&
		!strings.EqualFold(draft.SourceProductionPlanID, strings.TrimSpace(filter.SourceProductionPlanID)) {
		return false
	}
	if len(filter.Statuses) > 0 {
		matched := false
		for _, status := range filter.Statuses {
			if draft.Status == productiondomain.NormalizePurchaseRequestDraftStatus(status) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	search := strings.ToLower(strings.TrimSpace(filter.Search))
	if search == "" {
		return true
	}

	return strings.Contains(strings.ToLower(draft.RequestNo), search) ||
		strings.Contains(strings.ToLower(draft.SourceProductionPlanNo), search) ||
		purchaseRequestDraftLinesContain(draft, search)
}

func purchaseRequestDraftLinesContain(draft productiondomain.PurchaseRequestDraft, search string) bool {
	for _, line := range draft.Lines {
		if strings.Contains(strings.ToLower(line.SKU), search) || strings.Contains(strings.ToLower(line.ItemName), search) {
			return true
		}
	}

	return false
}

func shortageQuantity(required decimal.Decimal, available decimal.Decimal) (decimal.Decimal, error) {
	diff, err := decimal.SubtractQuantity(required, available)
	if err != nil {
		return "", err
	}
	if diff.IsNegative() {
		return decimal.MustQuantity("0"), nil
	}

	return diff, nil
}

type productionPlanIssueAllocation struct {
	WarehouseID   string
	WarehouseCode string
	LocationID    string
	LocationCode  string
	ItemID        string
	BatchID       string
	BatchNo       string
	Quantity      decimal.Decimal
}

func (s ProductionPlanService) enrichProductionPlanIssueReadiness(
	ctx context.Context,
	plan productiondomain.ProductionPlan,
) (productiondomain.ProductionPlan, error) {
	enriched := plan.Clone()
	issues := []inventorydomain.WarehouseIssue{}
	if s.warehouseIssue != nil {
		rows, err := s.warehouseIssue.ListWarehouseIssues(ctx)
		if err != nil {
			return productiondomain.ProductionPlan{}, err
		}
		issues = rows
	}

	for index, line := range enriched.Lines {
		updated, err := s.enrichProductionPlanLineIssueReadiness(ctx, enriched, line, issues)
		if err != nil {
			return productiondomain.ProductionPlan{}, err
		}
		enriched.Lines[index] = updated
	}

	return enriched, nil
}

func (s ProductionPlanService) enrichProductionPlanLineIssueReadiness(
	ctx context.Context,
	plan productiondomain.ProductionPlan,
	line productiondomain.ProductionPlanLine,
	issues []inventorydomain.WarehouseIssue,
) (productiondomain.ProductionPlanLine, error) {
	updated := line
	if !line.IsStockManaged {
		updated.AvailableQty = decimal.MustQuantity("0")
		updated.ShortageQty = decimal.MustQuantity("0")
		updated.IssuedQty = decimal.MustQuantity("0")
		updated.RemainingIssueQty = decimal.MustQuantity("0")
		updated.IssueStatus = productiondomain.ProductionPlanIssueStatusIssued
		updated.NeedsPurchase = false
		return updated, nil
	}

	availableQty, err := s.currentAvailableQtyForProductionLine(ctx, line)
	if err != nil {
		return productiondomain.ProductionPlanLine{}, err
	}
	issuedQty := decimal.MustQuantity("0")
	refs := make([]productiondomain.ProductionPlanWarehouseIssueRef, 0)
	for _, issue := range issues {
		for _, issueLine := range issue.Lines {
			if !warehouseIssueLineMatchesProductionPlanLine(issueLine, plan.ID, line.ID) {
				continue
			}
			refs = append(refs, productiondomain.ProductionPlanWarehouseIssueRef{
				ID:       issue.ID,
				IssueNo:  issue.IssueNo,
				LineID:   issueLine.ID,
				Status:   string(issue.Status),
				Quantity: issueLine.Quantity,
			})
			if issue.Status != inventorydomain.WarehouseIssueStatusPosted {
				continue
			}
			issuedQty, err = decimal.AddQuantity(issuedQty, issueLine.Quantity)
			if err != nil {
				return productiondomain.ProductionPlanLine{}, err
			}
		}
	}
	remainingQty, err := shortageQuantity(line.RequiredStockBaseQty, issuedQty)
	if err != nil {
		return productiondomain.ProductionPlanLine{}, err
	}
	shortageQty, err := shortageQuantity(remainingQty, availableQty)
	if err != nil {
		return productiondomain.ProductionPlanLine{}, err
	}

	updated.AvailableQty = availableQty
	updated.ShortageQty = shortageQty
	updated.NeedsPurchase = !shortageQty.IsZero()
	updated.IssuedQty = issuedQty
	updated.RemainingIssueQty = remainingQty
	updated.WarehouseIssues = refs
	updated.IssueStatus = productionPlanIssueStatus(updated)

	return updated, nil
}

func (s ProductionPlanService) currentAvailableQtyForProductionLine(
	ctx context.Context,
	line productiondomain.ProductionPlanLine,
) (decimal.Decimal, error) {
	if s.availableStock == nil {
		return line.AvailableQty, nil
	}
	snapshots, err := s.availableStock.Execute(ctx, inventorydomain.NewAvailableStockFilter("", "", line.ComponentSKU, ""))
	if err != nil {
		return "", err
	}
	total := decimal.MustQuantity("0")
	for _, snapshot := range snapshots {
		if snapshot.BaseUOMCode != line.StockBaseUOMCode {
			continue
		}
		total, err = decimal.AddQuantity(total, snapshot.AvailableQty)
		if err != nil {
			return "", err
		}
	}

	return total, nil
}

func productionPlanIssueStatus(line productiondomain.ProductionPlanLine) productiondomain.ProductionPlanIssueStatus {
	if line.RemainingIssueQty.IsZero() {
		return productiondomain.ProductionPlanIssueStatusIssued
	}
	if hasWarehouseIssueStatus(line.WarehouseIssues, inventorydomain.WarehouseIssueStatusApproved) {
		return productiondomain.ProductionPlanIssueStatusIssueApproved
	}
	if hasWarehouseIssueStatus(line.WarehouseIssues, inventorydomain.WarehouseIssueStatusSubmitted) {
		return productiondomain.ProductionPlanIssueStatusIssueSubmitted
	}
	if hasWarehouseIssueStatus(line.WarehouseIssues, inventorydomain.WarehouseIssueStatusDraft) {
		return productiondomain.ProductionPlanIssueStatusIssueDraft
	}
	if !line.IssuedQty.IsZero() {
		return productiondomain.ProductionPlanIssueStatusPartiallyIssued
	}
	if line.ShortageQty.IsZero() {
		return productiondomain.ProductionPlanIssueStatusReadyToIssue
	}

	return productiondomain.ProductionPlanIssueStatusShortage
}

func hasWarehouseIssueStatus(
	refs []productiondomain.ProductionPlanWarehouseIssueRef,
	status inventorydomain.WarehouseIssueStatus,
) bool {
	for _, ref := range refs {
		if ref.Status == string(status) {
			return true
		}
	}

	return false
}

func warehouseIssueLineMatchesProductionPlanLine(
	line inventorydomain.WarehouseIssueLine,
	planID string,
	lineID string,
) bool {
	return strings.EqualFold(line.SourceDocumentType, "production_plan") &&
		strings.EqualFold(line.SourceDocumentID, strings.TrimSpace(planID)) &&
		strings.EqualFold(line.SourceDocumentLineID, strings.TrimSpace(lineID))
}

func isIssueReadyProductionPlanLine(line productiondomain.ProductionPlanLine) bool {
	if !line.IsStockManaged || line.RemainingIssueQty.IsZero() {
		return false
	}
	switch line.IssueStatus {
	case productiondomain.ProductionPlanIssueStatusReadyToIssue,
		productiondomain.ProductionPlanIssueStatusPartiallyIssued:
		return line.ShortageQty.IsZero()
	default:
		return false
	}
}

func (s ProductionPlanService) allocateProductionPlanIssueStock(
	ctx context.Context,
	line productiondomain.ProductionPlanLine,
	warehouseID string,
) ([]productionPlanIssueAllocation, error) {
	if s.availableStock == nil {
		return nil, errors.New("production plan stock reader is required")
	}
	snapshots, err := s.availableStock.Execute(ctx, inventorydomain.NewAvailableStockFilter(strings.TrimSpace(warehouseID), "", line.ComponentSKU, ""))
	if err != nil {
		return nil, err
	}
	groups := groupProductionPlanIssueStockByWarehouse(snapshots, line.StockBaseUOMCode)
	requiredQty := line.RemainingIssueQty
	for _, group := range groups {
		shortage, err := shortageQuantity(requiredQty, group.totalQty)
		if err != nil {
			return nil, err
		}
		if !shortage.IsZero() {
			continue
		}
		return allocateProductionPlanIssueStockFromGroup(group, requiredQty)
	}

	return nil, ErrProductionPlanMaterialIssueNotReady
}

type productionPlanIssueWarehouseGroup struct {
	warehouseID   string
	warehouseCode string
	totalQty      decimal.Decimal
	snapshots     []inventorydomain.AvailableStockSnapshot
}

func groupProductionPlanIssueStockByWarehouse(
	snapshots []inventorydomain.AvailableStockSnapshot,
	uom decimal.UOMCode,
) []productionPlanIssueWarehouseGroup {
	byID := make(map[string]int)
	groups := make([]productionPlanIssueWarehouseGroup, 0)
	for _, snapshot := range snapshots {
		if snapshot.BaseUOMCode != uom || snapshot.AvailableQty.IsZero() || snapshot.AvailableQty.IsNegative() {
			continue
		}
		warehouseID := strings.TrimSpace(snapshot.WarehouseID)
		if warehouseID == "" {
			continue
		}
		index, ok := byID[warehouseID]
		if !ok {
			index = len(groups)
			byID[warehouseID] = index
			groups = append(groups, productionPlanIssueWarehouseGroup{
				warehouseID:   warehouseID,
				warehouseCode: strings.TrimSpace(snapshot.WarehouseCode),
				totalQty:      decimal.MustQuantity("0"),
				snapshots:     make([]inventorydomain.AvailableStockSnapshot, 0),
			})
		}
		groups[index].snapshots = append(groups[index].snapshots, snapshot)
		groups[index].totalQty = mustAddProductionPlanQuantity(groups[index].totalQty, snapshot.AvailableQty)
	}

	return groups
}

func allocateProductionPlanIssueStockFromGroup(
	group productionPlanIssueWarehouseGroup,
	requiredQty decimal.Decimal,
) ([]productionPlanIssueAllocation, error) {
	remaining := requiredQty
	allocations := make([]productionPlanIssueAllocation, 0, len(group.snapshots))
	for _, snapshot := range group.snapshots {
		if remaining.IsZero() {
			break
		}
		takeQty, err := minProductionPlanQuantity(snapshot.AvailableQty, remaining)
		if err != nil {
			return nil, err
		}
		if takeQty.IsZero() {
			continue
		}
		allocations = append(allocations, productionPlanIssueAllocation{
			WarehouseID:   group.warehouseID,
			WarehouseCode: group.warehouseCode,
			LocationID:    snapshot.LocationID,
			LocationCode:  snapshot.LocationCode,
			ItemID:        snapshot.ItemID,
			BatchID:       snapshot.BatchID,
			BatchNo:       snapshot.BatchNo,
			Quantity:      takeQty,
		})
		remaining, err = decimal.SubtractQuantity(remaining, takeQty)
		if err != nil {
			return nil, err
		}
	}
	if !remaining.IsZero() {
		return nil, ErrProductionPlanMaterialIssueNotReady
	}

	return allocations, nil
}

func minProductionPlanQuantity(left decimal.Decimal, right decimal.Decimal) (decimal.Decimal, error) {
	diff, err := decimal.SubtractQuantity(left, right)
	if err != nil {
		return "", err
	}
	if diff.IsNegative() {
		return left, nil
	}

	return right, nil
}

func mustAddProductionPlanQuantity(left decimal.Decimal, right decimal.Decimal) decimal.Decimal {
	result, err := decimal.AddQuantity(left, right)
	if err != nil {
		panic(err)
	}

	return result
}

func firstNonBlankProductionPlan(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}

	return ""
}

func newProductionPlanAuditLog(actorID string, requestID string, action string, plan productiondomain.ProductionPlan, occurredAt time.Time) (audit.Log, error) {
	if strings.TrimSpace(actorID) == "" {
		actorID = "system"
	}
	if strings.TrimSpace(requestID) == "" {
		requestID = fmt.Sprintf("production-plan-%d", occurredAt.UnixNano())
	}

	return audit.NewLog(audit.NewLogInput{
		ID:         fmt.Sprintf("audit-production-plan-%d", occurredAt.UnixNano()),
		OrgID:      plan.OrgID,
		ActorID:    actorID,
		Action:     action,
		EntityType: productionPlanEntityType,
		EntityID:   plan.ID,
		RequestID:  requestID,
		AfterData: map[string]any{
			"plan_no":             plan.PlanNo,
			"output_sku":          plan.OutputSKU,
			"planned_qty":         plan.PlannedQty.String(),
			"formula_id":          plan.FormulaID,
			"purchase_draft_no":   plan.PurchaseDraft.RequestNo,
			"purchase_line_count": len(plan.PurchaseDraft.Lines),
		},
		Metadata: map[string]any{
			"source": "production planning",
		},
		CreatedAt: occurredAt.UTC(),
	})
}

func newPurchaseRequestDraftAuditLog(
	actorID string,
	requestID string,
	action string,
	plan productiondomain.ProductionPlan,
	before productiondomain.PurchaseRequestDraft,
	after productiondomain.PurchaseRequestDraft,
	occurredAt time.Time,
) (audit.Log, error) {
	if strings.TrimSpace(actorID) == "" {
		actorID = "system"
	}
	if strings.TrimSpace(requestID) == "" {
		requestID = fmt.Sprintf("purchase-request-%d", occurredAt.UnixNano())
	}

	return audit.NewLog(audit.NewLogInput{
		ID:         fmt.Sprintf("audit-purchase-request-%d", occurredAt.UnixNano()),
		OrgID:      plan.OrgID,
		ActorID:    actorID,
		Action:     action,
		EntityType: purchaseRequestEntityType,
		EntityID:   after.ID,
		RequestID:  requestID,
		BeforeData: map[string]any{
			"request_no": after.RequestNo,
			"status":     string(before.Status),
		},
		AfterData: map[string]any{
			"request_no":                  after.RequestNo,
			"status":                      string(after.Status),
			"source_production_plan_id":   after.SourceProductionPlanID,
			"source_production_plan_no":   after.SourceProductionPlanNo,
			"converted_purchase_order_id": after.ConvertedPurchaseOrderID,
			"converted_purchase_order_no": after.ConvertedPurchaseOrderNo,
		},
		Metadata: map[string]any{
			"source":             "production planning",
			"production_plan_id": plan.ID,
			"production_plan_no": plan.PlanNo,
		},
		CreatedAt: occurredAt.UTC(),
	})
}

func newProductionPlanID(now time.Time) string {
	return fmt.Sprintf("pp-%s-%06d", now.UTC().Format("060102"), now.UTC().UnixNano()%1000000)
}

func newProductionPlanNo(now time.Time) string {
	return fmt.Sprintf("PP-%s-%06d", now.UTC().Format("060102"), now.UTC().UnixNano()%1000000)
}

func newProductionPlanLineID(planID string, lineNo int) string {
	return fmt.Sprintf("pp-line-%s-%03d", strings.ToLower(strings.TrimSpace(planID)), lineNo)
}

func newPurchaseRequestDraftID(now time.Time) string {
	return fmt.Sprintf("pr-draft-%s-%06d", now.UTC().Format("060102"), now.UTC().UnixNano()%1000000)
}

func newPurchaseRequestDraftNo(now time.Time) string {
	return fmt.Sprintf("PR-DRAFT-%s-%06d", now.UTC().Format("060102"), now.UTC().UnixNano()%1000000)
}
