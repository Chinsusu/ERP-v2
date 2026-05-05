package application

import (
	"context"
	"errors"
	"testing"
	"time"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	masterdatadomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestProductionPlanServiceCreatesFormulaSnapshotDemandAndPurchaseRequestDraft(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)
	formula := activeProductionPlanFormula(t)
	store := NewPrototypeProductionPlanStore(audit.NewInMemoryLogStore())
	service := NewProductionPlanService(
		store,
		fakeProductionPlanFormulaReader{formula: formula},
		fakeProductionPlanAvailableStock{
			rows: []inventorydomain.AvailableStockSnapshot{
				{
					ItemID:       "item-act-baicapil",
					SKU:          "ACT_BAICAPIL",
					BaseUOMCode:  decimal.MustUOMCode("KG"),
					AvailableQty: decimal.MustQuantity("0.000500"),
				},
			},
		},
	)
	service.clock = func() time.Time { return now }

	result, err := service.CreateProductionPlan(ctx, CreateProductionPlanInput{
		ID:               "plan-001",
		PlanNo:           "PP-260504-0001",
		OutputItemID:     "item-xff-150",
		PlannedQty:       "162",
		UOMCode:          "PCS",
		PlannedStartDate: "2026-05-10",
		PlannedEndDate:   "2026-05-11",
		ActorID:          "user-production",
		RequestID:        "req-production-plan",
	})
	if err != nil {
		t.Fatalf("CreateProductionPlan() error = %v", err)
	}

	plan := result.ProductionPlan
	if plan.FormulaID != formula.ID || plan.FormulaVersion != "v1" || len(plan.Lines) != 1 {
		t.Fatalf("plan snapshot = %+v, want active formula snapshot with one line", plan)
	}
	line := plan.Lines[0]
	if line.RequiredStockBaseQty != "0.162000" || line.AvailableQty != "0.000500" || line.ShortageQty != "0.161500" {
		t.Fatalf("line = %+v, want shortage from formula demand minus available stock", line)
	}
	if len(plan.PurchaseDraft.Lines) != 1 {
		t.Fatalf("purchase draft lines = %d, want 1", len(plan.PurchaseDraft.Lines))
	}
	prLine := plan.PurchaseDraft.Lines[0]
	if prLine.SKU != "ACT_BAICAPIL" || prLine.RequestedQty != "0.161500" || prLine.UOMCode != "KG" {
		t.Fatalf("purchase draft line = %+v, want shortage draft only", prLine)
	}
	if result.AuditLogID == "" {
		t.Fatalf("audit id is empty")
	}
	if store.Count() != 1 {
		t.Fatalf("stored plans = %d, want 1", store.Count())
	}
}

func TestProductionPlanServiceDoesNotCreatePurchaseDraftWhenEnoughStock(t *testing.T) {
	ctx := context.Background()
	formula := activeProductionPlanFormula(t)
	store := NewPrototypeProductionPlanStore(audit.NewInMemoryLogStore())
	service := NewProductionPlanService(
		store,
		fakeProductionPlanFormulaReader{formula: formula},
		fakeProductionPlanAvailableStock{
			rows: []inventorydomain.AvailableStockSnapshot{
				{
					ItemID:       "item-act-baicapil",
					SKU:          "ACT_BAICAPIL",
					BaseUOMCode:  decimal.MustUOMCode("KG"),
					AvailableQty: decimal.MustQuantity("1.000000"),
				},
			},
		},
	)

	result, err := service.CreateProductionPlan(ctx, CreateProductionPlanInput{
		ID:           "plan-002",
		PlanNo:       "PP-260504-0002",
		OutputItemID: "item-xff-150",
		PlannedQty:   "162",
		UOMCode:      "PCS",
		ActorID:      "user-production",
	})
	if err != nil {
		t.Fatalf("CreateProductionPlan() error = %v", err)
	}
	if len(result.ProductionPlan.PurchaseDraft.Lines) != 0 {
		t.Fatalf("purchase draft = %+v, want no draft when stock is enough", result.ProductionPlan.PurchaseDraft)
	}
	if result.ProductionPlan.Lines[0].NeedsPurchase || result.ProductionPlan.Lines[0].ShortageQty != "0.000000" {
		t.Fatalf("line = %+v, want no shortage", result.ProductionPlan.Lines[0])
	}
}

func TestProductionPlanServiceCreatesWarehouseIssueFromReadyDemandLineAndRollsUpPostedIssue(t *testing.T) {
	ctx := context.Background()
	formula := activeProductionPlanFormula(t)
	auditStore := audit.NewInMemoryLogStore()
	planStore := NewPrototypeProductionPlanStore(auditStore)
	issueStore := inventoryapp.NewPrototypeWarehouseIssueStore()
	issueService := inventoryapp.NewWarehouseIssueService(
		issueStore,
		inventoryapp.NewInMemoryStockMovementStore(),
		auditStore,
	).WithClock(func() time.Time {
		return time.Date(2026, 5, 5, 11, 0, 0, 0, time.UTC)
	})
	service := NewProductionPlanService(
		planStore,
		fakeProductionPlanFormulaReader{formula: formula},
		fakeProductionPlanAvailableStock{
			rows: []inventorydomain.AvailableStockSnapshot{
				{
					WarehouseID:   "wh-hcm-rm",
					WarehouseCode: "WH-HCM-RM",
					LocationID:    "bin-rm-a01",
					LocationCode:  "RM-A01",
					ItemID:        "item-act-baicapil",
					SKU:           "ACT_BAICAPIL",
					BatchID:       "batch-act-baicapil-001",
					BatchNo:       "LOT-ACT-001",
					BaseUOMCode:   decimal.MustUOMCode("KG"),
					AvailableQty:  decimal.MustQuantity("1.000000"),
				},
			},
		},
	).WithWarehouseIssueService(issueService)

	created, err := service.CreateProductionPlan(ctx, CreateProductionPlanInput{
		ID:           "plan-issue-001",
		PlanNo:       "PP-260505-ISSUE-001",
		OutputItemID: "item-xff-150",
		PlannedQty:   "162",
		UOMCode:      "PCS",
		ActorID:      "user-production",
	})
	if err != nil {
		t.Fatalf("CreateProductionPlan() error = %v", err)
	}
	lineID := created.ProductionPlan.Lines[0].ID

	issueResult, err := service.CreateWarehouseIssueFromProductionPlan(ctx, CreateProductionPlanWarehouseIssueInput{
		PlanID:          created.ProductionPlan.ID,
		LineIDs:         []string{lineID},
		DestinationType: "factory",
		DestinationName: "Factory A",
		ReasonCode:      "production_plan_issue",
		ActorID:         "user-production",
		RequestID:       "req-plan-issue-create",
	})
	if err != nil {
		t.Fatalf("CreateWarehouseIssueFromProductionPlan() error = %v", err)
	}
	issue := issueResult.WarehouseIssue
	if issue.WarehouseID != "wh-hcm-rm" || issue.Lines[0].LocationID != "bin-rm-a01" {
		t.Fatalf("issue location = %s/%s, want selected available stock warehouse/bin", issue.WarehouseID, issue.Lines[0].LocationID)
	}
	if issue.Lines[0].SourceDocumentType != "production_plan" ||
		issue.Lines[0].SourceDocumentID != created.ProductionPlan.ID ||
		issue.Lines[0].SourceDocumentLineID != lineID {
		t.Fatalf("issue source = %+v, want production plan and demand line source", issue.Lines[0])
	}
	if issue.Lines[0].Quantity != created.ProductionPlan.Lines[0].RequiredStockBaseQty {
		t.Fatalf("issue qty = %s, want required stock base qty %s", issue.Lines[0].Quantity, created.ProductionPlan.Lines[0].RequiredStockBaseQty)
	}

	for _, action := range []string{"submit", "approve", "post"} {
		if _, err := issueService.TransitionWarehouseIssue(ctx, inventoryapp.WarehouseIssueTransitionInput{
			ID:      issue.ID,
			ActorID: "warehouse-lead",
			Action:  action,
		}); err != nil {
			t.Fatalf("%s issue: %v", action, err)
		}
	}

	updated, err := service.GetProductionPlan(ctx, created.ProductionPlan.ID)
	if err != nil {
		t.Fatalf("GetProductionPlan() error = %v", err)
	}
	updatedLine := updated.Lines[0]
	if updatedLine.IssuedQty != updatedLine.RequiredStockBaseQty ||
		updatedLine.RemainingIssueQty != "0.000000" ||
		updatedLine.IssueStatus != productiondomain.ProductionPlanIssueStatusIssued {
		t.Fatalf("updated line = %+v, want issued readiness after posted Warehouse Issue", updatedLine)
	}
	if len(updatedLine.WarehouseIssues) != 1 || updatedLine.WarehouseIssues[0].IssueNo != issue.IssueNo {
		t.Fatalf("warehouse issue refs = %+v, want posted issue ref", updatedLine.WarehouseIssues)
	}
}

func TestProductionPlanServiceRejectsWarehouseIssueFromShortageLine(t *testing.T) {
	ctx := context.Background()
	auditStore := audit.NewInMemoryLogStore()
	service := NewProductionPlanService(
		NewPrototypeProductionPlanStore(auditStore),
		fakeProductionPlanFormulaReader{formula: activeProductionPlanFormula(t)},
		fakeProductionPlanAvailableStock{
			rows: []inventorydomain.AvailableStockSnapshot{
				{
					WarehouseID:   "wh-hcm-rm",
					WarehouseCode: "WH-HCM-RM",
					ItemID:        "item-act-baicapil",
					SKU:           "ACT_BAICAPIL",
					BaseUOMCode:   decimal.MustUOMCode("KG"),
					AvailableQty:  decimal.MustQuantity("0.000500"),
				},
			},
		},
	).WithWarehouseIssueService(inventoryapp.NewWarehouseIssueService(
		inventoryapp.NewPrototypeWarehouseIssueStore(),
		inventoryapp.NewInMemoryStockMovementStore(),
		auditStore,
	))

	created, err := service.CreateProductionPlan(ctx, CreateProductionPlanInput{
		ID:           "plan-shortage-001",
		PlanNo:       "PP-260505-SHORTAGE-001",
		OutputItemID: "item-xff-150",
		PlannedQty:   "162",
		UOMCode:      "PCS",
		ActorID:      "user-production",
	})
	if err != nil {
		t.Fatalf("CreateProductionPlan() error = %v", err)
	}

	_, err = service.CreateWarehouseIssueFromProductionPlan(ctx, CreateProductionPlanWarehouseIssueInput{
		PlanID:          created.ProductionPlan.ID,
		LineIDs:         []string{created.ProductionPlan.Lines[0].ID},
		DestinationType: "factory",
		DestinationName: "Factory A",
		ReasonCode:      "production_plan_issue",
		ActorID:         "user-production",
	})
	if !errors.Is(err, ErrProductionPlanMaterialIssueNotReady) {
		t.Fatalf("CreateWarehouseIssueFromProductionPlan() error = %v, want ErrProductionPlanMaterialIssueNotReady", err)
	}
}

func TestProductionPlanServiceSubmitsApprovesAndConvertsPurchaseRequestDraft(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 5, 9, 0, 0, 0, time.UTC)
	store := NewPrototypeProductionPlanStore(audit.NewInMemoryLogStore())
	service := newProductionPlanServiceWithShortage(t, store)
	service.clock = func() time.Time { return now }

	created, err := service.CreateProductionPlan(ctx, CreateProductionPlanInput{
		ID:           "plan-001",
		PlanNo:       "PP-260505-0001",
		OutputItemID: "item-xff-150",
		PlannedQty:   "162",
		UOMCode:      "PCS",
		ActorID:      "user-production",
	})
	if err != nil {
		t.Fatalf("CreateProductionPlan() error = %v", err)
	}
	draftID := created.ProductionPlan.PurchaseDraft.ID

	_, err = service.SubmitPurchaseRequestDraft(ctx, PurchaseRequestDraftActionInput{
		ID:             draftID,
		ExpectedStatus: productiondomain.PurchaseRequestDraftStatusSubmitted,
		ActorID:        "user-purchase",
		RequestID:      "req-pr-submit-stale",
	})
	if !errors.Is(err, productiondomain.ErrProductionPlanInvalidPurchaseRequestTransition) {
		t.Fatalf("stale submit error = %v, want ErrProductionPlanInvalidPurchaseRequestTransition", err)
	}

	submitted, err := service.SubmitPurchaseRequestDraft(ctx, PurchaseRequestDraftActionInput{
		ID:        draftID,
		ActorID:   "user-purchase",
		RequestID: "req-pr-submit",
	})
	if err != nil {
		t.Fatalf("SubmitPurchaseRequestDraft() error = %v", err)
	}
	if submitted.CurrentStatus != productiondomain.PurchaseRequestDraftStatusSubmitted {
		t.Fatalf("submitted status = %q, want submitted", submitted.CurrentStatus)
	}
	service.clock = func() time.Time { return now.Add(time.Hour) }
	approved, err := service.ApprovePurchaseRequestDraft(ctx, PurchaseRequestDraftActionInput{
		ID:        draftID,
		ActorID:   "user-manager",
		RequestID: "req-pr-approve",
	})
	if err != nil {
		t.Fatalf("ApprovePurchaseRequestDraft() error = %v", err)
	}
	if approved.CurrentStatus != productiondomain.PurchaseRequestDraftStatusApproved {
		t.Fatalf("approved status = %q, want approved", approved.CurrentStatus)
	}
	service.clock = func() time.Time { return now.Add(2 * time.Hour) }
	converted, err := service.MarkPurchaseRequestDraftConverted(ctx, ConvertPurchaseRequestDraftInput{
		ID:              draftID,
		PurchaseOrderID: "po-260505-0001",
		PurchaseOrderNo: "PO-260505-0001",
		ActorID:         "user-purchase",
		RequestID:       "req-pr-convert",
	})
	if err != nil {
		t.Fatalf("MarkPurchaseRequestDraftConverted() error = %v", err)
	}
	if converted.PurchaseRequestDraft.Status != productiondomain.PurchaseRequestDraftStatusConvertedToPO {
		t.Fatalf("converted status = %q, want converted_to_po", converted.PurchaseRequestDraft.Status)
	}
	if converted.PurchaseRequestDraft.ConvertedPurchaseOrderID != "po-260505-0001" {
		t.Fatalf("converted PO id = %q, want po-260505-0001", converted.PurchaseRequestDraft.ConvertedPurchaseOrderID)
	}

	drafts, err := service.ListPurchaseRequestDrafts(ctx, PurchaseRequestDraftFilter{Search: "PP-260505"})
	if err != nil {
		t.Fatalf("ListPurchaseRequestDrafts() error = %v", err)
	}
	if len(drafts) != 1 || drafts[0].Status != productiondomain.PurchaseRequestDraftStatusConvertedToPO {
		t.Fatalf("drafts = %+v, want one converted draft", drafts)
	}
}

func TestProductionPlanServiceRejectsPurchaseRequestApprovalBeforeSubmit(t *testing.T) {
	ctx := context.Background()
	store := NewPrototypeProductionPlanStore(audit.NewInMemoryLogStore())
	service := newProductionPlanServiceWithShortage(t, store)
	created, err := service.CreateProductionPlan(ctx, CreateProductionPlanInput{
		ID:           "plan-001",
		PlanNo:       "PP-260505-0001",
		OutputItemID: "item-xff-150",
		PlannedQty:   "162",
		UOMCode:      "PCS",
		ActorID:      "user-production",
	})
	if err != nil {
		t.Fatalf("CreateProductionPlan() error = %v", err)
	}

	_, err = service.ApprovePurchaseRequestDraft(ctx, PurchaseRequestDraftActionInput{
		ID:      created.ProductionPlan.PurchaseDraft.ID,
		ActorID: "user-manager",
	})
	if !errors.Is(err, productiondomain.ErrProductionPlanInvalidPurchaseRequestTransition) {
		t.Fatalf("error = %v, want ErrProductionPlanInvalidPurchaseRequestTransition", err)
	}
}

func TestProductionPlanServiceAcceptsPublicItemReferenceResolvedByFormulaList(t *testing.T) {
	ctx := context.Background()
	formula := activeProductionPlanFormula(t)
	formula.FinishedItemID = "2e2f71b4-a502-43e8-a448-04d875a04cb5"
	store := NewPrototypeProductionPlanStore(audit.NewInMemoryLogStore())
	service := NewProductionPlanService(
		store,
		fakeProductionPlanFormulaReader{formula: formula, outputAliases: []string{"item-xff-150"}},
		fakeProductionPlanAvailableStock{},
	)

	result, err := service.CreateProductionPlan(ctx, CreateProductionPlanInput{
		ID:           "plan-public-ref-list",
		PlanNo:       "PP-260504-PUBLIC-LIST",
		OutputItemID: "item-xff-150",
		PlannedQty:   "10",
		UOMCode:      "PCS",
		ActorID:      "user-production",
	})
	if err != nil {
		t.Fatalf("CreateProductionPlan() error = %v", err)
	}
	if result.ProductionPlan.OutputItemID != formula.FinishedItemID {
		t.Fatalf("OutputItemID = %q, want formula finished item UUID %q", result.ProductionPlan.OutputItemID, formula.FinishedItemID)
	}
}

func newProductionPlanServiceWithShortage(t *testing.T, store *PrototypeProductionPlanStore) ProductionPlanService {
	t.Helper()
	return NewProductionPlanService(
		store,
		fakeProductionPlanFormulaReader{formula: activeProductionPlanFormula(t)},
		fakeProductionPlanAvailableStock{
			rows: []inventorydomain.AvailableStockSnapshot{
				{
					ItemID:       "item-act-baicapil",
					SKU:          "ACT_BAICAPIL",
					BaseUOMCode:  decimal.MustUOMCode("KG"),
					AvailableQty: decimal.MustQuantity("0.000500"),
				},
			},
		},
	)
}

func TestProductionPlanServiceAcceptsPublicItemReferenceWithExplicitFormulaID(t *testing.T) {
	ctx := context.Background()
	formula := activeProductionPlanFormula(t)
	formula.FinishedItemID = "2e2f71b4-a502-43e8-a448-04d875a04cb5"
	store := NewPrototypeProductionPlanStore(audit.NewInMemoryLogStore())
	service := NewProductionPlanService(
		store,
		fakeProductionPlanFormulaReader{formula: formula, outputAliases: []string{"item-xff-150"}},
		fakeProductionPlanAvailableStock{},
	)

	result, err := service.CreateProductionPlan(ctx, CreateProductionPlanInput{
		ID:           "plan-public-ref-get",
		PlanNo:       "PP-260504-PUBLIC-GET",
		OutputItemID: "item-xff-150",
		FormulaID:    formula.ID,
		PlannedQty:   "10",
		UOMCode:      "PCS",
		ActorID:      "user-production",
	})
	if err != nil {
		t.Fatalf("CreateProductionPlan() error = %v", err)
	}
	if result.ProductionPlan.FormulaID != formula.ID {
		t.Fatalf("FormulaID = %q, want %q", result.ProductionPlan.FormulaID, formula.ID)
	}
}

func activeProductionPlanFormula(t *testing.T) masterdatadomain.Formula {
	t.Helper()
	formula, err := masterdatadomain.NewFormula(masterdatadomain.NewFormulaInput{
		ID:               "formula-xff-v1",
		FormulaCode:      "XFF-150ML",
		FinishedItemID:   "item-xff-150",
		FinishedSKU:      "XFF",
		FinishedItemName: "Tinh chat buoi Fast & Furious 150ML",
		FinishedItemType: masterdatadomain.ItemTypeFinishedGood,
		FormulaVersion:   "v1",
		BatchQty:         decimal.MustQuantity("81"),
		BatchUOMCode:     "PCS",
		BaseBatchQty:     decimal.MustQuantity("81"),
		BaseBatchUOMCode: "PCS",
		Status:           masterdatadomain.FormulaStatusActive,
		ApprovalStatus:   masterdatadomain.FormulaApprovalApproved,
		Lines: []masterdatadomain.NewFormulaLineInput{
			{
				ID:               "formula-line-001",
				LineNo:           1,
				ComponentItemID:  "item-act-baicapil",
				ComponentSKU:     "ACT_BAICAPIL",
				ComponentName:    "BAICAPIL",
				ComponentType:    masterdatadomain.FormulaComponentRawMaterial,
				EnteredQty:       decimal.MustQuantity("0.001"),
				EnteredUOMCode:   "KG",
				CalcQty:          decimal.MustQuantity("1"),
				CalcUOMCode:      "G",
				StockBaseQty:     decimal.MustQuantity("0.001"),
				StockBaseUOMCode: "KG",
				WastePercent:     decimal.MustRate("0"),
				IsRequired:       true,
				IsStockManaged:   true,
				LineStatus:       masterdatadomain.FormulaLineStatusActive,
			},
		},
	})
	if err != nil {
		t.Fatalf("formula fixture: %v", err)
	}

	return formula
}

type fakeProductionPlanFormulaReader struct {
	formula       masterdatadomain.Formula
	err           error
	outputAliases []string
}

func (r fakeProductionPlanFormulaReader) Get(ctx context.Context, id string) (masterdatadomain.Formula, error) {
	if r.err != nil {
		return masterdatadomain.Formula{}, r.err
	}
	if id != "" && id != r.formula.ID {
		return masterdatadomain.Formula{}, errors.New("formula not found")
	}

	return r.formula.Clone(), nil
}

func (r fakeProductionPlanFormulaReader) List(ctx context.Context, filter masterdatadomain.FormulaFilter) ([]masterdatadomain.Formula, error) {
	if r.err != nil {
		return nil, r.err
	}
	if filter.FinishedItemID != "" && filter.FinishedItemID != r.formula.FinishedItemID && !r.matchesOutputAlias(filter.FinishedItemID) {
		return nil, nil
	}
	if filter.Status != "" && filter.Status != r.formula.Status {
		return nil, nil
	}

	return []masterdatadomain.Formula{r.formula.Clone()}, nil
}

func (r fakeProductionPlanFormulaReader) matchesOutputAlias(outputItemID string) bool {
	for _, alias := range r.outputAliases {
		if alias == outputItemID {
			return true
		}
	}

	return false
}

type fakeProductionPlanAvailableStock struct {
	rows []inventorydomain.AvailableStockSnapshot
}

func (s fakeProductionPlanAvailableStock) Execute(ctx context.Context, filter inventorydomain.AvailableStockFilter) ([]inventorydomain.AvailableStockSnapshot, error) {
	matches := make([]inventorydomain.AvailableStockSnapshot, 0, len(s.rows))
	for _, row := range s.rows {
		if filter.SKU != "" && row.SKU != filter.SKU {
			continue
		}
		matches = append(matches, row)
	}

	return matches, nil
}

var _ ProductionPlanStore = (*PrototypeProductionPlanStore)(nil)
