package application

import (
	"context"
	"testing"
	"time"

	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestSubcontractOrderServiceFactoryDispatchConfirmTransitionsOrder(t *testing.T) {
	ctx := context.Background()
	auditStore := audit.NewInMemoryLogStore()
	orderStore := NewPrototypeSubcontractOrderStore(auditStore)
	dispatchStore := NewPrototypeSubcontractFactoryDispatchStore()
	service := SubcontractOrderService{
		store:                orderStore,
		factoryDispatchStore: dispatchStore,
		factoryDispatchBuild: NewSubcontractFactoryDispatchService(),
	}
	order := subcontractFactoryDispatchApprovedOrder(t)
	if err := orderStore.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		return tx.Save(txCtx, order)
	}); err != nil {
		t.Fatalf("seed order: %v", err)
	}

	created, err := service.CreateFactoryDispatch(ctx, CreateFactoryDispatchInput{
		ID:                 "dispatch-001",
		DispatchNo:         "FDP-260506-001",
		SubcontractOrderID: order.ID,
		ExpectedVersion:    order.Version,
		ActorID:            "production-user",
		RequestID:          "req-dispatch-create",
		Note:               "Manual send pack for factory.",
	})
	if err != nil {
		t.Fatalf("create dispatch: %v", err)
	}
	if created.Dispatch.Status != productiondomain.SubcontractFactoryDispatchStatusDraft ||
		created.Dispatch.FinishedSKUCode != order.FinishedSKUCode ||
		len(created.Dispatch.Lines) != len(order.MaterialLines) {
		t.Fatalf("created dispatch = %+v, want draft snapshot from order", created.Dispatch)
	}

	ready, err := service.MarkFactoryDispatchReady(ctx, FactoryDispatchActionInput{
		SubcontractOrderID: order.ID,
		DispatchID:         created.Dispatch.ID,
		ExpectedVersion:    created.Dispatch.Version,
		ActorID:            "production-lead",
		RequestID:          "req-dispatch-ready",
	})
	if err != nil {
		t.Fatalf("mark dispatch ready: %v", err)
	}
	sent, err := service.MarkFactoryDispatchSent(ctx, MarkFactoryDispatchSentInput{
		SubcontractOrderID: order.ID,
		DispatchID:         ready.Dispatch.ID,
		ExpectedVersion:    ready.Dispatch.Version,
		SentBy:             "production-user",
		SentAt:             time.Date(2026, 5, 6, 10, 30, 0, 0, time.UTC),
		ActorID:            "production-user",
		RequestID:          "req-dispatch-sent",
		Evidence: []FactoryDispatchEvidenceInput{
			{
				ID:           "dispatch-evidence-001",
				EvidenceType: "manual_send",
				FileName:     "factory-dispatch-screenshot.png",
				ObjectKey:    "subcontract/dispatch-001/factory-dispatch-screenshot.png",
			},
		},
	})
	if err != nil {
		t.Fatalf("mark dispatch sent: %v", err)
	}

	confirmed, err := service.RecordFactoryDispatchResponse(ctx, RecordFactoryDispatchResponseInput{
		SubcontractOrderID: order.ID,
		DispatchID:         sent.Dispatch.ID,
		ExpectedVersion:    sent.Dispatch.Version,
		ResponseStatus:     "confirmed",
		ResponseBy:         "factory-coordinator",
		RespondedAt:        time.Date(2026, 5, 6, 11, 0, 0, 0, time.UTC),
		ResponseNote:       "Factory confirmed quantity, spec, and delivery date.",
		ActorID:            "production-user",
		RequestID:          "req-dispatch-confirm",
	})
	if err != nil {
		t.Fatalf("record dispatch response: %v", err)
	}

	if confirmed.Dispatch.Status != productiondomain.SubcontractFactoryDispatchStatusConfirmed ||
		confirmed.SubcontractOrder.Status != productiondomain.SubcontractOrderStatusFactoryConfirmed {
		t.Fatalf("confirmed result = %+v, want dispatch confirmed and order factory_confirmed", confirmed)
	}
	logs, err := auditStore.List(ctx, audit.Query{EntityID: order.ID})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if !subcontractAuditActionsContain(logs, subcontractFactoryDispatchCreatedAction) ||
		!subcontractAuditActionsContain(logs, subcontractFactoryDispatchSentAction) ||
		!subcontractAuditActionsContain(logs, subcontractFactoryDispatchConfirmedAction) {
		t.Fatalf("audit logs = %+v, want factory dispatch create/sent/confirmed actions", logs)
	}
}

func TestSubcontractOrderServiceFactoryDispatchRevisionDoesNotConfirmOrder(t *testing.T) {
	ctx := context.Background()
	auditStore := audit.NewInMemoryLogStore()
	orderStore := NewPrototypeSubcontractOrderStore(auditStore)
	dispatchStore := NewPrototypeSubcontractFactoryDispatchStore()
	service := SubcontractOrderService{
		store:                orderStore,
		factoryDispatchStore: dispatchStore,
		factoryDispatchBuild: NewSubcontractFactoryDispatchService(),
	}
	order := subcontractFactoryDispatchApprovedOrder(t)
	if err := orderStore.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		return tx.Save(txCtx, order)
	}); err != nil {
		t.Fatalf("seed order: %v", err)
	}
	created, err := service.CreateFactoryDispatch(ctx, CreateFactoryDispatchInput{
		ID:                 "dispatch-001",
		SubcontractOrderID: order.ID,
		ExpectedVersion:    order.Version,
		ActorID:            "production-user",
		RequestID:          "req-dispatch-create",
	})
	if err != nil {
		t.Fatalf("create dispatch: %v", err)
	}
	ready, err := service.MarkFactoryDispatchReady(ctx, FactoryDispatchActionInput{
		SubcontractOrderID: order.ID,
		DispatchID:         created.Dispatch.ID,
		ExpectedVersion:    created.Dispatch.Version,
		ActorID:            "production-lead",
		RequestID:          "req-dispatch-ready",
	})
	if err != nil {
		t.Fatalf("mark ready: %v", err)
	}
	sent, err := service.MarkFactoryDispatchSent(ctx, MarkFactoryDispatchSentInput{
		SubcontractOrderID: order.ID,
		DispatchID:         ready.Dispatch.ID,
		ExpectedVersion:    ready.Dispatch.Version,
		SentBy:             "production-user",
		ActorID:            "production-user",
		RequestID:          "req-dispatch-sent",
		Evidence: []FactoryDispatchEvidenceInput{
			{
				ID:           "dispatch-evidence-001",
				EvidenceType: "manual_send",
				ObjectKey:    "subcontract/dispatch-001/factory-dispatch-screenshot.png",
			},
		},
	})
	if err != nil {
		t.Fatalf("mark sent: %v", err)
	}

	revision, err := service.RecordFactoryDispatchResponse(ctx, RecordFactoryDispatchResponseInput{
		SubcontractOrderID: order.ID,
		DispatchID:         sent.Dispatch.ID,
		ExpectedVersion:    sent.Dispatch.Version,
		ResponseStatus:     "revision_requested",
		ResponseBy:         "factory-coordinator",
		ResponseNote:       "Factory requests carton spec revision.",
		ActorID:            "production-user",
		RequestID:          "req-dispatch-revision",
	})
	if err != nil {
		t.Fatalf("record revision response: %v", err)
	}

	if revision.Dispatch.Status != productiondomain.SubcontractFactoryDispatchStatusRevisionRequested ||
		revision.SubcontractOrder.Status != productiondomain.SubcontractOrderStatusApproved {
		t.Fatalf("revision result = %+v, want order to stay approved", revision)
	}
}

func subcontractFactoryDispatchApprovedOrder(t *testing.T) productiondomain.SubcontractOrder {
	t.Helper()

	order, err := productiondomain.NewSubcontractOrderDocument(productiondomain.NewSubcontractOrderDocumentInput{
		ID:                     "sco-001",
		OrgID:                  "org-my-pham",
		OrderNo:                "SCO-260506-001",
		FactoryID:              "fac-001",
		FactoryCode:            "FAC-HCM",
		FactoryName:            "HCM Cosmetics Factory",
		FinishedItemID:         "fg-serum",
		FinishedSKUCode:        "FG-SERUM-001",
		FinishedItemName:       "Brightening Serum",
		PlannedQty:             decimal.MustQuantity("1000"),
		UOMCode:                "PCS",
		BasePlannedQty:         decimal.MustQuantity("1000"),
		BaseUOMCode:            "PCS",
		ConversionFactor:       decimal.MustQuantity("1"),
		CurrencyCode:           "VND",
		SpecSummary:            "formula v2026.05, box spec approved",
		SourceProductionPlanID: "plan-001",
		SourceProductionPlanNo: "PP-260506-001",
		SampleRequired:         true,
		TargetStartDate:        "2026-05-08",
		ExpectedReceiptDate:    "2026-05-20",
		CreatedAt:              time.Date(2026, 5, 6, 9, 0, 0, 0, time.UTC),
		CreatedBy:              "production-user",
		MaterialLines: []productiondomain.NewSubcontractMaterialLineInput{
			{
				ID:               "sco-mat-001",
				LineNo:           1,
				ItemID:           "rm-base",
				SKUCode:          "RM-BASE",
				ItemName:         "Serum Base",
				PlannedQty:       decimal.MustQuantity("10"),
				UOMCode:          "KG",
				BasePlannedQty:   decimal.MustQuantity("10000"),
				BaseUOMCode:      "G",
				ConversionFactor: decimal.MustQuantity("1000"),
				UnitCost:         decimal.MustUnitCost("150000"),
				CurrencyCode:     "VND",
				LotTraceRequired: true,
			},
		},
	})
	if err != nil {
		t.Fatalf("new subcontract order: %v", err)
	}
	order, err = order.Submit("production-user", time.Now())
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	order, err = order.Approve("production-lead", time.Now())
	if err != nil {
		t.Fatalf("approve: %v", err)
	}

	return order
}
