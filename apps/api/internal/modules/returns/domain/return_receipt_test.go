package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewReturnReceiptCreatesPendingInspectionForKnownReturn(t *testing.T) {
	receipt, err := NewReturnReceipt(NewReturnReceiptInput{
		WarehouseID:      "wh-hcm",
		WarehouseCode:    "HCM",
		Source:           ReturnSourceCarrier,
		ReceivedBy:       "user-return-inspector",
		ScanCode:         " ghn260426001 ",
		PackageCondition: "sealed",
		Disposition:      ReturnDispositionReusable,
		ExpectedReturn: &ExpectedReturn{
			OrderNo:       "SO-260426-001",
			TrackingNo:    "GHN260426001",
			ReturnCode:    "RET-260426-001",
			CustomerName:  "Nguyen An",
			SKU:           "SERUM-30ML",
			ProductName:   "Serum 30ml",
			Quantity:      2,
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
		},
		CreatedAt: time.Date(2026, 4, 26, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("new return receipt: %v", err)
	}

	if receipt.Status != ReturnStatusPendingInspection {
		t.Fatalf("status = %q, want pending inspection", receipt.Status)
	}
	if receipt.UnknownCase {
		t.Fatal("unknown case = true, want false")
	}
	if receipt.OriginalOrderNo != "SO-260426-001" || receipt.TrackingNo != "GHN260426001" {
		t.Fatalf("receipt identifiers = %+v, want linked order and tracking", receipt)
	}
	if receipt.TargetLocation != "return-area-pending-inspection" {
		t.Fatalf("target location = %q, want return-area-pending-inspection", receipt.TargetLocation)
	}
	if receipt.StockMovement == nil {
		t.Fatal("stock movement is nil, want RETURN_RECEIPT movement")
	}
	if receipt.StockMovement.MovementType != ReturnReceiptMovementType || receipt.StockMovement.TargetStockStatus != "return_pending" {
		t.Fatalf("stock movement = %+v, want RETURN_RECEIPT to return_pending", receipt.StockMovement)
	}
}

func TestNewReturnReceiptCreatesUnknownCase(t *testing.T) {
	receipt, err := NewReturnReceipt(NewReturnReceiptInput{
		WarehouseID:      "wh-hcm",
		ScanCode:         "UNKNOWN-TRACKING",
		PackageCondition: "damaged box",
		Disposition:      ReturnDispositionNeedsInspection,
	})
	if err != nil {
		t.Fatalf("new return receipt: %v", err)
	}

	if !receipt.UnknownCase {
		t.Fatal("unknown case = false, want true")
	}
	if receipt.InvestigationNote == "" {
		t.Fatal("investigation note is empty")
	}
	if receipt.TargetLocation != "return-inspection-queue" {
		t.Fatalf("target location = %q, want return-inspection-queue", receipt.TargetLocation)
	}
	if receipt.StockMovement != nil {
		t.Fatalf("stock movement = %+v, want nil for needs inspection", receipt.StockMovement)
	}
}

func TestNewReturnReceiptRoutesNotReusableToLabPlaceholder(t *testing.T) {
	receipt, err := NewReturnReceipt(NewReturnReceiptInput{
		WarehouseID:      "wh-hcm",
		ScanCode:         "SO-260426-002",
		PackageCondition: "broken seal",
		Disposition:      ReturnDispositionNotReusable,
		ExpectedReturn: &ExpectedReturn{
			OrderNo:     "SO-260426-002",
			TrackingNo:  "GHN260426002",
			SKU:         "TONER-100ML",
			ProductName: "Toner 100ml",
			Quantity:    1,
		},
	})
	if err != nil {
		t.Fatalf("new return receipt: %v", err)
	}

	if receipt.TargetLocation != "lab-damaged-placeholder" {
		t.Fatalf("target location = %q, want lab-damaged-placeholder", receipt.TargetLocation)
	}
	if receipt.StockMovement != nil {
		t.Fatalf("stock movement = %+v, want nil for not reusable", receipt.StockMovement)
	}
}

func TestNewReturnReceiptValidatesScanAndDisposition(t *testing.T) {
	if _, err := NewReturnReceipt(NewReturnReceiptInput{
		WarehouseID: "wh-hcm",
		Disposition: ReturnDispositionReusable,
	}); !errors.Is(err, ErrReturnReceiptScanCodeRequired) {
		t.Fatalf("err = %v, want scan code required", err)
	}

	if _, err := NewReturnReceipt(NewReturnReceiptInput{
		WarehouseID: "wh-hcm",
		ScanCode:    "SO-260426-001",
		Disposition: "available",
	}); !errors.Is(err, ErrReturnReceiptInvalidDisposition) {
		t.Fatalf("err = %v, want invalid disposition", err)
	}
}

func TestExpectedReturnMatchesOrderTrackingReturnCodeOrShipment(t *testing.T) {
	expected := ExpectedReturn{
		OrderNo:    "SO-260426-001",
		TrackingNo: "GHN260426001",
		ReturnCode: "RET-260426-001",
		ShipmentID: "ship-hcm-260426-001",
	}

	for _, code := range []string{"so-260426-001", "GHN260426001", "RET-260426-001", "ship-hcm-260426-001"} {
		if !expected.MatchesScanCode(code) {
			t.Fatalf("expected return did not match %q", code)
		}
	}
}
