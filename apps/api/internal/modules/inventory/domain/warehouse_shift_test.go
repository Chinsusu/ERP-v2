package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestNewWarehouseShiftCapturesOperationsMetricsAndPendingIssues(t *testing.T) {
	openedAt := time.Date(2026, 4, 28, 8, 0, 0, 0, time.UTC)
	shift, err := NewWarehouseShift(NewWarehouseShiftInput{
		ID:            "shift-hcm-260428-day",
		WarehouseID:   "wh-hcm",
		WarehouseCode: "hcm",
		BusinessDate:  "2026-04-28",
		ShiftCode:     "DAY",
		OpenedBy:      "user-warehouse-lead",
		OpenedAt:      openedAt,
		Metrics: WarehouseShiftMetrics{
			ReceivedOrders:      12,
			ProcessedOrders:     18,
			PackedOrders:        16,
			HandedOverOrders:    15,
			ReturnOrders:        3,
			MovementCount:       7,
			InboundMovementQty:  decimal.MustQuantity("22"),
			OutboundMovementQty: decimal.MustQuantity("18.5"),
			NetMovementQty:      decimal.MustQuantity("3.5"),
		},
		PendingIssues: []WarehouseShiftPendingIssue{
			{
				ID:          "issue-manifest-short",
				IssueType:   "manifest_mismatch",
				Severity:    WarehouseShiftIssueSeverityCritical,
				ReferenceID: "manifest-hcm-001",
				Description: "one package missing at carrier handover",
				Blocking:    true,
				Owner:       "Warehouse Lead",
			},
		},
	})
	if err != nil {
		t.Fatalf("new warehouse shift: %v", err)
	}

	if shift.OrgID != "org-my-pham" ||
		shift.WarehouseCode != "HCM" ||
		shift.ShiftCode != "day" ||
		shift.Status != WarehouseShiftStatusBlocked {
		t.Fatalf("shift = %+v, want normalized blocked shift", shift)
	}
	if shift.Metrics.ProcessedOrders != 18 ||
		shift.Metrics.ReturnOrders != 3 ||
		shift.Metrics.MovementCount != 7 ||
		shift.Metrics.InboundMovementQty != "22.000000" ||
		shift.Metrics.OutboundMovementQty != "18.500000" ||
		shift.Metrics.NetMovementQty != "3.500000" {
		t.Fatalf("metrics = %+v, want order/return/movement summary", shift.Metrics)
	}
	if len(shift.OpenBlockingIssues()) != 1 || shift.OpenBlockingIssues()[0].Status != WarehouseShiftIssueStatusOpen {
		t.Fatalf("open blocking issues = %+v, want one normalized open issue", shift.OpenBlockingIssues())
	}
}

func TestWarehouseShiftCloseRequiresNoOpenBlockingIssue(t *testing.T) {
	shift, err := NewWarehouseShift(NewWarehouseShiftInput{
		WarehouseID:  "wh-hcm",
		BusinessDate: "2026-04-28",
		ShiftCode:    "day",
		OpenedBy:     "user-warehouse-lead",
		PendingIssues: []WarehouseShiftPendingIssue{
			{IssueType: "return_pending", Blocking: true, Description: "return inspection pending"},
		},
	})
	if err != nil {
		t.Fatalf("new warehouse shift: %v", err)
	}

	_, err = shift.Close("user-warehouse-lead", time.Time{})
	if !errors.Is(err, ErrWarehouseShiftPendingIssue) {
		t.Fatalf("close err = %v, want pending issue error", err)
	}
}

func TestNewWarehouseShiftRejectsInvalidMovementQuantity(t *testing.T) {
	_, err := NewWarehouseShift(NewWarehouseShiftInput{
		WarehouseID:  "wh-hcm",
		BusinessDate: "2026-04-28",
		ShiftCode:    "day",
		OpenedBy:     "user-warehouse-lead",
		Metrics: WarehouseShiftMetrics{
			InboundMovementQty: decimal.Decimal("1.1234567"),
		},
	})
	if !errors.Is(err, ErrWarehouseShiftInvalidQuantity) {
		t.Fatalf("err = %v, want invalid quantity", err)
	}
}

func TestWarehouseShiftCloseCapturesActorAndTimestamp(t *testing.T) {
	shift, err := NewWarehouseShift(NewWarehouseShiftInput{
		ID:           "shift-hcm-close",
		WarehouseID:  "wh-hcm",
		BusinessDate: "2026-04-28",
		ShiftCode:    "day",
		OpenedBy:     "user-warehouse-lead",
		PendingIssues: []WarehouseShiftPendingIssue{
			{IssueType: "variance", Status: WarehouseShiftIssueStatusResolved, Blocking: true},
			{IssueType: "handover_note", Blocking: false},
		},
	})
	if err != nil {
		t.Fatalf("new warehouse shift: %v", err)
	}
	closedAt := time.Date(2026, 4, 28, 17, 30, 0, 0, time.UTC)

	closed, err := shift.Close("user-warehouse-lead", closedAt)
	if err != nil {
		t.Fatalf("close warehouse shift: %v", err)
	}

	if closed.Status != WarehouseShiftStatusClosed ||
		closed.ClosedBy != "user-warehouse-lead" ||
		!closed.ClosedAt.Equal(closedAt) ||
		len(closed.OpenBlockingIssues()) != 0 {
		t.Fatalf("closed shift = %+v, want closed with actor and timestamp", closed)
	}
}
