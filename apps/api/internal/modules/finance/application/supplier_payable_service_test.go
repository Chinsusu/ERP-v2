package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	financedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	apperrors "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/errors"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestSupplierPayableServiceCreatesPayableAndAudit(t *testing.T) {
	service, auditStore := newTestSupplierPayableService()

	result, err := service.CreateSupplierPayable(context.Background(), baseCreateSupplierPayableInput())
	if err != nil {
		t.Fatalf("create payable: %v", err)
	}
	if result.SupplierPayable.Status != financedomain.PayableStatusOpen {
		t.Fatalf("status = %q, want open", result.SupplierPayable.Status)
	}
	if result.AuditLogID == "" {
		t.Fatal("audit log id is empty")
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: string(financedomain.FinanceAuditActionPayableCreated)})
	if err != nil {
		t.Fatalf("list audit: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
	if logs[0].EntityType != string(financedomain.FinanceEntityTypeSupplierPayable) ||
		logs[0].Metadata["source_type"] != "qc_inspection" {
		t.Fatalf("audit log = %+v, want finance AP source metadata", logs[0])
	}
}

func TestSupplierPayableServiceApprovesAndRecordsPayment(t *testing.T) {
	service, auditStore := newTestSupplierPayableService()
	created, err := service.CreateSupplierPayable(context.Background(), baseCreateSupplierPayableInput())
	if err != nil {
		t.Fatalf("create payable: %v", err)
	}

	approved, err := service.ApproveSupplierPayablePayment(context.Background(), SupplierPayableActionInput{
		ID:        created.SupplierPayable.ID,
		ActorID:   "finance-lead",
		RequestID: "req-ap-approve",
	})
	if err != nil {
		t.Fatalf("approve payment: %v", err)
	}
	if approved.PreviousStatus != financedomain.PayableStatusOpen ||
		approved.CurrentStatus != financedomain.PayableStatusPaymentApproved {
		t.Fatalf("status transition = %q -> %q", approved.PreviousStatus, approved.CurrentStatus)
	}

	paid, err := service.RecordSupplierPayablePayment(context.Background(), SupplierPayableActionInput{
		ID:        created.SupplierPayable.ID,
		Amount:    "1250000.00",
		ActorID:   "cashier",
		RequestID: "req-ap-pay",
	})
	if err != nil {
		t.Fatalf("record payment: %v", err)
	}
	if paid.CurrentStatus != financedomain.PayableStatusPartiallyPaid ||
		paid.SupplierPayable.PaidAmount != "1250000.00" ||
		paid.SupplierPayable.OutstandingAmount != "3000000.00" {
		t.Fatalf("paid result = %+v", paid)
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: string(financedomain.FinanceAuditActionPayablePaymentRecorded)})
	if err != nil {
		t.Fatalf("list audit: %v", err)
	}
	if len(logs) != 1 || logs[0].Metadata["payment_amount"] != "1250000.00" {
		t.Fatalf("audit logs = %+v, want payment amount metadata", logs)
	}
}

func TestSupplierPayableServiceRequestsAndRejectsPayment(t *testing.T) {
	service, auditStore := newTestSupplierPayableService()
	created, err := service.CreateSupplierPayable(context.Background(), baseCreateSupplierPayableInput())
	if err != nil {
		t.Fatalf("create payable: %v", err)
	}

	requested, err := service.RequestSupplierPayablePayment(context.Background(), SupplierPayableActionInput{
		ID:        created.SupplierPayable.ID,
		ActorID:   "finance-user",
		RequestID: "req-ap-request-payment",
	})
	if err != nil {
		t.Fatalf("request payment: %v", err)
	}
	if requested.PreviousStatus != financedomain.PayableStatusOpen ||
		requested.CurrentStatus != financedomain.PayableStatusPaymentRequested {
		t.Fatalf("status transition = %q -> %q", requested.PreviousStatus, requested.CurrentStatus)
	}

	rejected, err := service.RejectSupplierPayablePayment(context.Background(), SupplierPayableActionInput{
		ID:        created.SupplierPayable.ID,
		Reason:    "supplier invoice mismatch",
		ActorID:   "finance-lead",
		RequestID: "req-ap-reject-payment",
	})
	if err != nil {
		t.Fatalf("reject payment: %v", err)
	}
	if rejected.PreviousStatus != financedomain.PayableStatusPaymentRequested ||
		rejected.CurrentStatus != financedomain.PayableStatusOpen ||
		rejected.SupplierPayable.PaymentRejectReason != "supplier invoice mismatch" {
		t.Fatalf("rejected result = %+v", rejected)
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: string(financedomain.FinanceAuditActionPayablePaymentRejected)})
	if err != nil {
		t.Fatalf("list audit: %v", err)
	}
	if len(logs) != 1 || logs[0].Metadata["reason"] != "supplier invoice mismatch" {
		t.Fatalf("audit logs = %+v, want rejection reason metadata", logs)
	}
}

func TestSupplierPayableServiceBlocksPaymentRequestWithoutMatchedSupplierInvoice(t *testing.T) {
	service, _ := newTestSupplierPayableServiceWithInvoiceGate()
	created, err := service.CreateSupplierPayable(context.Background(), baseCreateSupplierPayableInput())
	if err != nil {
		t.Fatalf("create payable: %v", err)
	}

	_, err = service.RequestSupplierPayablePayment(context.Background(), SupplierPayableActionInput{
		ID:        created.SupplierPayable.ID,
		ActorID:   "finance-user",
		RequestID: "req-ap-request-without-invoice",
	})
	var appErr apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error = %T, want app error", err)
	}
	if appErr.Code != ErrorCodeSupplierPayableInvoiceNotMatched {
		t.Fatalf("code = %q, want %q", appErr.Code, ErrorCodeSupplierPayableInvoiceNotMatched)
	}
}

func TestSupplierPayableServiceAllowsPaymentRequestWithMatchedSupplierInvoice(t *testing.T) {
	service, invoiceStore := newTestSupplierPayableServiceWithInvoiceGate()
	created, err := service.CreateSupplierPayable(context.Background(), baseCreateSupplierPayableInput())
	if err != nil {
		t.Fatalf("create payable: %v", err)
	}
	saveMatchedSupplierInvoiceForPayable(t, invoiceStore, created.SupplierPayable)

	requested, err := service.RequestSupplierPayablePayment(context.Background(), SupplierPayableActionInput{
		ID:        created.SupplierPayable.ID,
		ActorID:   "finance-user",
		RequestID: "req-ap-request-with-invoice",
	})
	if err != nil {
		t.Fatalf("request payment: %v", err)
	}
	if requested.CurrentStatus != financedomain.PayableStatusPaymentRequested {
		t.Fatalf("status = %q, want payment_requested", requested.CurrentStatus)
	}
}

func TestSupplierPayableServiceMapsValidationErrors(t *testing.T) {
	service, _ := newTestSupplierPayableService()
	input := baseCreateSupplierPayableInput()
	input.TotalAmount = "4250000.001"

	_, err := service.CreateSupplierPayable(context.Background(), input)
	var appErr apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error = %T, want app error", err)
	}
	if appErr.Code != ErrorCodeSupplierPayableValidation {
		t.Fatalf("code = %q, want %q", appErr.Code, ErrorCodeSupplierPayableValidation)
	}
}

func TestSupplierPayableServiceReturnsNotFound(t *testing.T) {
	service, _ := newTestSupplierPayableService()

	_, err := service.GetSupplierPayable(context.Background(), "missing-ap")
	var appErr apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error = %T, want app error", err)
	}
	if appErr.Code != ErrorCodeSupplierPayableNotFound {
		t.Fatalf("code = %q, want not found", appErr.Code)
	}
}

func TestSupplierPayableServiceSearchMatchesLineSourceDocument(t *testing.T) {
	service, _ := newTestSupplierPayableService()
	if _, err := service.CreateSupplierPayable(context.Background(), baseCreateSupplierPayableInput()); err != nil {
		t.Fatalf("create payable: %v", err)
	}

	payables, err := service.ListSupplierPayables(context.Background(), SupplierPayableFilter{
		Search: "GR-260430-0001",
	})
	if err != nil {
		t.Fatalf("list payables: %v", err)
	}
	if len(payables) != 1 || payables[0].ID != "ap-260430-0001" {
		t.Fatalf("payables = %+v, want line source document match", payables)
	}
}

func TestPostgresSupplierPayableServicePersistsPaymentLifecycleAcrossFreshStores(t *testing.T) {
	databaseURL := os.Getenv("ERP_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("ERP_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := seedSupplierPayableSmokeOrg(ctx, db); err != nil {
		t.Fatalf("seed org: %v", err)
	}

	suffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	input := baseCreateSupplierPayableInput()
	input.ID = "ap-s15-03-02-" + suffix
	input.PayableNo = "AP-S15-03-02-" + suffix
	input.SourceDocument.ID = "qc-s15-03-02-" + suffix
	input.SourceDocument.No = "QC-S15-03-02-" + suffix
	input.RequestID = "req-ap-s15-03-02-create-" + suffix
	input.Lines[0].ID = "ap-line-s15-03-02-" + suffix
	input.Lines[0].SourceDocument.ID = "gr-s15-03-02-" + suffix
	input.Lines[0].SourceDocument.No = "GR-S15-03-02-" + suffix

	newService := func(now time.Time) SupplierPayableService {
		store := NewPostgresSupplierPayableStore(
			db,
			PostgresSupplierPayableStoreConfig{DefaultOrgID: testSupplierPayableOrgID},
		)
		auditStore := audit.NewPostgresLogStore(
			db,
			audit.PostgresLogStoreConfig{DefaultOrgID: testSupplierPayableOrgID},
		)
		return NewSupplierPayableService(store, auditStore).WithClock(func() time.Time {
			return now
		})
	}

	created, err := newService(time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)).
		CreateSupplierPayable(ctx, input)
	if err != nil {
		t.Fatalf("create payable: %v", err)
	}

	reloaded, err := newService(time.Date(2026, 5, 2, 10, 5, 0, 0, time.UTC)).
		GetSupplierPayable(ctx, created.SupplierPayable.ID)
	if err != nil {
		t.Fatalf("reload created payable: %v", err)
	}
	if reloaded.ID != input.ID ||
		reloaded.SourceDocument.ID != input.SourceDocument.ID ||
		len(reloaded.Lines) != 1 ||
		reloaded.Lines[0].ID != input.Lines[0].ID {
		t.Fatalf("reloaded payable = %+v, want fresh store reload with source and line refs", reloaded)
	}

	requested, err := newService(time.Date(2026, 5, 2, 10, 10, 0, 0, time.UTC)).
		RequestSupplierPayablePayment(ctx, SupplierPayableActionInput{
			ID:        input.ID,
			ActorID:   "finance-user",
			RequestID: "req-ap-s15-03-02-request-" + suffix,
		})
	if err != nil {
		t.Fatalf("request payment: %v", err)
	}
	if requested.CurrentStatus != financedomain.PayableStatusPaymentRequested {
		t.Fatalf("request status = %q, want payment_requested", requested.CurrentStatus)
	}

	rejected, err := newService(time.Date(2026, 5, 2, 10, 15, 0, 0, time.UTC)).
		RejectSupplierPayablePayment(ctx, SupplierPayableActionInput{
			ID:        input.ID,
			Reason:    "supplier invoice mismatch",
			ActorID:   "finance-lead",
			RequestID: "req-ap-s15-03-02-reject-" + suffix,
		})
	if err != nil {
		t.Fatalf("reject payment: %v", err)
	}
	if rejected.CurrentStatus != financedomain.PayableStatusOpen ||
		rejected.SupplierPayable.PaymentRejectReason != "supplier invoice mismatch" {
		t.Fatalf("rejected payable = %+v, want open with reject reason", rejected.SupplierPayable)
	}

	requested, err = newService(time.Date(2026, 5, 2, 10, 20, 0, 0, time.UTC)).
		RequestSupplierPayablePayment(ctx, SupplierPayableActionInput{
			ID:        input.ID,
			ActorID:   "finance-user",
			RequestID: "req-ap-s15-03-02-request-again-" + suffix,
		})
	if err != nil {
		t.Fatalf("request payment again: %v", err)
	}

	approved, err := newService(time.Date(2026, 5, 2, 10, 25, 0, 0, time.UTC)).
		ApproveSupplierPayablePayment(ctx, SupplierPayableActionInput{
			ID:        input.ID,
			ActorID:   "finance-lead",
			RequestID: "req-ap-s15-03-02-approve-" + suffix,
		})
	if err != nil {
		t.Fatalf("approve payment: %v", err)
	}
	if approved.PreviousStatus != requested.CurrentStatus ||
		approved.CurrentStatus != financedomain.PayableStatusPaymentApproved {
		t.Fatalf("approve transition = %q -> %q", approved.PreviousStatus, approved.CurrentStatus)
	}

	paid, err := newService(time.Date(2026, 5, 2, 10, 30, 0, 0, time.UTC)).
		RecordSupplierPayablePayment(ctx, SupplierPayableActionInput{
			ID:        input.ID,
			Amount:    "1250000.00",
			ActorID:   "cashier",
			RequestID: "req-ap-s15-03-02-payment-" + suffix,
		})
	if err != nil {
		t.Fatalf("record payment: %v", err)
	}
	if paid.CurrentStatus != financedomain.PayableStatusPartiallyPaid {
		t.Fatalf("payment status = %q, want partially_paid", paid.CurrentStatus)
	}

	final, err := newService(time.Date(2026, 5, 2, 10, 35, 0, 0, time.UTC)).
		GetSupplierPayable(ctx, input.PayableNo)
	if err != nil {
		t.Fatalf("reload paid payable: %v", err)
	}
	if final.Status != financedomain.PayableStatusPartiallyPaid ||
		final.PaidAmount.String() != "1250000.00" ||
		final.OutstandingAmount.String() != "3000000.00" ||
		final.PaymentRequestedBy != "finance-user" ||
		final.PaymentApprovedBy != "finance-lead" ||
		final.LastPaymentBy != "cashier" ||
		final.PaymentRejectReason != "" ||
		final.Version != 6 {
		t.Fatalf("final payable = %+v, want persisted request/reject/approve/payment lifecycle", final)
	}

	auditStore := audit.NewPostgresLogStore(
		db,
		audit.PostgresLogStoreConfig{DefaultOrgID: testSupplierPayableOrgID},
	)
	logs, err := auditStore.List(ctx, audit.Query{
		EntityType: string(financedomain.FinanceEntityTypeSupplierPayable),
		EntityID:   input.ID,
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 6 {
		t.Fatalf("audit logs = %d, want create, request, reject, request, approve, payment", len(logs))
	}
}

func newTestSupplierPayableService() (SupplierPayableService, *audit.InMemoryLogStore) {
	auditStore := audit.NewInMemoryLogStore()
	store := &PrototypeSupplierPayableStore{records: make(map[string]financedomain.SupplierPayable)}
	service := NewSupplierPayableService(store, auditStore).WithClock(func() time.Time {
		return time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC)
	})

	return service, auditStore
}

func newTestSupplierPayableServiceWithInvoiceGate() (SupplierPayableService, *PrototypeSupplierInvoiceStore) {
	auditStore := audit.NewInMemoryLogStore()
	payableStore := &PrototypeSupplierPayableStore{records: make(map[string]financedomain.SupplierPayable)}
	invoiceStore := &PrototypeSupplierInvoiceStore{records: make(map[string]financedomain.SupplierInvoice)}
	service := NewSupplierPayableService(payableStore, auditStore).
		WithClock(func() time.Time {
			return time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC)
		}).
		WithSupplierInvoiceStore(invoiceStore)

	return service, invoiceStore
}

func saveMatchedSupplierInvoiceForPayable(
	t *testing.T,
	store *PrototypeSupplierInvoiceStore,
	payable financedomain.SupplierPayable,
) {
	t.Helper()
	lines := make([]financedomain.NewSupplierInvoiceLineInput, 0, len(payable.Lines))
	for _, line := range payable.Lines {
		lines = append(lines, financedomain.NewSupplierInvoiceLineInput{
			ID:             "si-" + line.ID,
			Description:    line.Description,
			SourceDocument: line.SourceDocument,
			Amount:         line.Amount.String(),
		})
	}
	invoice, err := financedomain.NewSupplierInvoice(financedomain.NewSupplierInvoiceInput{
		ID:             "si-" + payable.ID,
		OrgID:          payable.OrgID,
		InvoiceNo:      "INV-" + payable.PayableNo,
		SupplierID:     payable.SupplierID,
		SupplierCode:   payable.SupplierCode,
		SupplierName:   payable.SupplierName,
		PayableID:      payable.ID,
		PayableNo:      payable.PayableNo,
		Status:         financedomain.SupplierInvoiceStatusMatched,
		MatchStatus:    financedomain.SupplierInvoiceMatchStatusMatched,
		SourceDocument: payable.SourceDocument,
		Lines:          lines,
		InvoiceAmount:  payable.TotalAmount.String(),
		ExpectedAmount: payable.TotalAmount.String(),
		VarianceAmount: "0.00",
		CurrencyCode:   payable.CurrencyCode.String(),
		InvoiceDate:    time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC),
		CreatedAt:      time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC),
		CreatedBy:      "finance-user",
		UpdatedAt:      time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC),
		UpdatedBy:      "finance-user",
	})
	if err != nil {
		t.Fatalf("new supplier invoice: %v", err)
	}
	if err := store.Save(context.Background(), invoice); err != nil {
		t.Fatalf("save supplier invoice: %v", err)
	}
}

func baseCreateSupplierPayableInput() CreateSupplierPayableInput {
	qcSource := SourceDocumentInput{
		Type: string(financedomain.SourceDocumentTypeQCInspection),
		ID:   "qc-inbound-260430-0001",
		No:   "QC-260430-0001",
	}
	receiptSource := SourceDocumentInput{
		Type: string(financedomain.SourceDocumentTypeWarehouseReceipt),
		ID:   "gr-260430-0001",
		No:   "GR-260430-0001",
	}

	return CreateSupplierPayableInput{
		ID:             "ap-260430-0001",
		PayableNo:      "AP-260430-0001",
		SupplierID:     "supplier-hcm-001",
		SupplierCode:   "SUP-HCM-001",
		SupplierName:   "Nguyen Lieu HCM",
		SourceDocument: qcSource,
		TotalAmount:    "4250000.00",
		CurrencyCode:   "VND",
		DueDate:        "2026-05-07",
		ActorID:        "finance-user",
		RequestID:      "req-ap-create",
		Lines: []SupplierPayableLineInput{
			{
				ID:             "ap-line-1",
				Description:    "Accepted raw material after inbound QC",
				SourceDocument: receiptSource,
				Amount:         "4250000.00",
			},
		},
	}
}
