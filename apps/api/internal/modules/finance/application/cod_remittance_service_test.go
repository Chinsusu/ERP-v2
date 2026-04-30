package application

import (
	"context"
	"errors"
	"testing"
	"time"

	financedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	apperrors "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/errors"
)

func TestCODRemittanceServiceCreatesListsAndAudits(t *testing.T) {
	store := NewPrototypeCODRemittanceStore()
	auditStore := audit.NewInMemoryLogStore()
	service := NewCODRemittanceService(store, auditStore).WithClock(fixedFinanceClock)

	result, err := service.CreateCODRemittance(context.Background(), CreateCODRemittanceInput{
		RemittanceNo:   "cod-vtp-260430-0002",
		CarrierID:      "carrier-vtp",
		CarrierCode:    "VTP",
		CarrierName:    "Viettel Post",
		BusinessDate:   "2026-04-30",
		ExpectedAmount: "1500000.00",
		RemittedAmount: "1500000.00",
		CurrencyCode:   "VND",
		ActorID:        "finance-user",
		RequestID:      "req-cod-create",
		Lines: []CODRemittanceLineInput{
			{
				ID:             "cod-vtp-line-1",
				ReceivableID:   "ar-cod-vtp-1",
				ReceivableNo:   "AR-COD-VTP-1",
				ShipmentID:     "ship-vtp-1",
				TrackingNo:     "VTP260430001",
				CustomerName:   "VTP Customer",
				ExpectedAmount: "1500000.00",
				RemittedAmount: "1500000.00",
			},
		},
	})
	if err != nil {
		t.Fatalf("create cod remittance: %v", err)
	}
	if result.CODRemittance.Status != financedomain.CODRemittanceStatusDraft ||
		result.CODRemittance.ExpectedAmount != "1500000.00" ||
		result.AuditLogID == "" {
		t.Fatalf("created remittance = %+v audit %q", result.CODRemittance, result.AuditLogID)
	}

	remittances, err := service.ListCODRemittances(context.Background(), CODRemittanceFilter{
		Search:    "viettel",
		CarrierID: "carrier-vtp",
	})
	if err != nil {
		t.Fatalf("list cod remittances: %v", err)
	}
	if len(remittances) != 1 || remittances[0].CarrierCode != "VTP" {
		t.Fatalf("filtered remittances = %+v", remittances)
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: string(financedomain.FinanceAuditActionCODRemittanceCreated)})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].EntityID != result.CODRemittance.ID {
		t.Fatalf("audit logs = %+v", logs)
	}
}

func TestCODRemittanceServiceDiscrepancySubmitApproveClose(t *testing.T) {
	store := NewPrototypeCODRemittanceStore()
	auditStore := audit.NewInMemoryLogStore()
	service := NewCODRemittanceService(store, auditStore).WithClock(fixedFinanceClock)

	discrepancy, err := service.RecordCODRemittanceDiscrepancy(context.Background(), CODRemittanceDiscrepancyInput{
		RemittanceID:  "cod-remit-260430-0001",
		DiscrepancyID: "disc-cod-1",
		LineID:        "cod-remit-260430-0001-line-1",
		Reason:        "carrier remitted short",
		OwnerID:       "finance-user",
		ActorID:       "finance-user",
		RequestID:     "req-cod-discrepancy",
	})
	if err != nil {
		t.Fatalf("record discrepancy: %v", err)
	}
	if discrepancy.CurrentStatus != financedomain.CODRemittanceStatusDiscrepancy {
		t.Fatalf("current status = %q, want discrepancy", discrepancy.CurrentStatus)
	}

	submitted, err := service.SubmitCODRemittance(context.Background(), CODRemittanceActionInput{
		ID:        "cod-remit-260430-0001",
		ActorID:   "finance-lead",
		RequestID: "req-cod-submit",
	})
	if err != nil {
		t.Fatalf("submit cod remittance: %v", err)
	}
	approved, err := service.ApproveCODRemittance(context.Background(), CODRemittanceActionInput{
		ID:        submitted.CODRemittance.ID,
		ActorID:   "finance-lead",
		RequestID: "req-cod-approve",
	})
	if err != nil {
		t.Fatalf("approve cod remittance: %v", err)
	}
	closed, err := service.CloseCODRemittance(context.Background(), CODRemittanceActionInput{
		ID:        approved.CODRemittance.ID,
		ActorID:   "finance-lead",
		RequestID: "req-cod-close",
	})
	if err != nil {
		t.Fatalf("close cod remittance: %v", err)
	}
	if closed.CurrentStatus != financedomain.CODRemittanceStatusClosed ||
		closed.CODRemittance.DiscrepancyAmount != "-50000.00" ||
		len(closed.CODRemittance.Discrepancies) != 1 {
		t.Fatalf("closed remittance = %+v", closed.CODRemittance)
	}
}

func TestCODRemittanceServiceMatchedFlow(t *testing.T) {
	store := NewPrototypeCODRemittanceStore()
	auditStore := audit.NewInMemoryLogStore()
	service := NewCODRemittanceService(store, auditStore).WithClock(fixedFinanceClock)
	created, err := service.CreateCODRemittance(context.Background(), CreateCODRemittanceInput{
		RemittanceNo:   "cod-clean-260430-0001",
		CarrierID:      "carrier-ghn",
		CarrierCode:    "GHN",
		CarrierName:    "GHN Express",
		BusinessDate:   "2026-04-30",
		ExpectedAmount: "1250000.00",
		RemittedAmount: "1250000.00",
		CurrencyCode:   "VND",
		ActorID:        "finance-user",
		Lines: []CODRemittanceLineInput{
			{
				ID:             "cod-clean-line-1",
				ReceivableID:   "ar-clean-1",
				ReceivableNo:   "AR-CLEAN-1",
				ShipmentID:     "ship-clean-1",
				TrackingNo:     "GHNCLEAN1",
				CustomerName:   "Clean Customer",
				ExpectedAmount: "1250000.00",
				RemittedAmount: "1250000.00",
			},
		},
	})
	if err != nil {
		t.Fatalf("create clean remittance: %v", err)
	}

	matched, err := service.MarkCODRemittanceMatching(context.Background(), CODRemittanceActionInput{
		ID:      created.CODRemittance.ID,
		ActorID: "finance-user",
	})
	if err != nil {
		t.Fatalf("match clean remittance: %v", err)
	}
	if matched.CurrentStatus != financedomain.CODRemittanceStatusMatching {
		t.Fatalf("current status = %q, want matching", matched.CurrentStatus)
	}
}

func TestCODRemittanceServiceMapsErrors(t *testing.T) {
	service := NewCODRemittanceService(NewPrototypeCODRemittanceStore(), audit.NewInMemoryLogStore()).WithClock(fixedFinanceClock)

	_, err := service.GetCODRemittance(context.Background(), "missing-cod")
	var appErr apperrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != ErrorCodeCODRemittanceNotFound {
		t.Fatalf("missing error = %v", err)
	}

	_, err = service.MarkCODRemittanceMatching(context.Background(), CODRemittanceActionInput{
		ID:      "cod-remit-260430-0001",
		ActorID: "finance-user",
	})
	if !errors.As(err, &appErr) || appErr.Code != ErrorCodeCODRemittanceValidation {
		t.Fatalf("discrepancy match error = %v", err)
	}
}

func fixedFinanceClock() time.Time {
	return time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC)
}
