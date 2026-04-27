package domain

import (
	"errors"
	"testing"
	"time"
)

func TestBatchDefaultsAndExpiryValidation(t *testing.T) {
	batch, err := NewBatch(validBatchInput(func(input *NewBatchInput) {
		input.QCStatus = ""
		input.Status = ""
		input.MfgDate = time.Date(2026, 4, 1, 8, 0, 0, 0, time.FixedZone("ICT", 7*60*60))
		input.ExpiryDate = time.Date(2027, 4, 1, 8, 0, 0, 0, time.FixedZone("ICT", 7*60*60))
	}))
	if err != nil {
		t.Fatalf("new batch: %v", err)
	}
	if batch.QCStatus != QCStatusHold {
		t.Fatalf("qc status = %q, want hold", batch.QCStatus)
	}
	if batch.Status != BatchStatusActive {
		t.Fatalf("status = %q, want active", batch.Status)
	}
	if batch.MfgDate.Hour() != 0 || batch.ExpiryDate.Hour() != 0 {
		t.Fatalf("dates should be normalized to date-only UTC, got mfg=%s expiry=%s", batch.MfgDate, batch.ExpiryDate)
	}

	_, err = NewBatch(validBatchInput(func(input *NewBatchInput) {
		input.MfgDate = time.Date(2026, 4, 2, 0, 0, 0, 0, time.UTC)
		input.ExpiryDate = time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	}))
	if !errors.Is(err, ErrBatchInvalidExpiry) {
		t.Fatalf("expiry error = %v, want invalid expiry", err)
	}
}

func TestBatchReleaseAndRejectedTransitions(t *testing.T) {
	batch, err := NewBatch(validBatchInput())
	if err != nil {
		t.Fatalf("new batch: %v", err)
	}

	released, err := batch.ChangeQCStatus(QCStatusPass, time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("release batch: %v", err)
	}
	if released.QCStatus != QCStatusPass {
		t.Fatalf("released qc status = %q, want pass", released.QCStatus)
	}
	if !released.IsAvailableForInventory(time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("released batch should be available before expiry")
	}

	rejected, err := batch.ChangeQCStatus(QCStatusFail, time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("reject batch: %v", err)
	}
	if rejected.QCStatus != QCStatusFail {
		t.Fatalf("rejected qc status = %q, want fail", rejected.QCStatus)
	}
	if rejected.IsAvailableForInventory(time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("rejected batch should not be available")
	}
}

func TestBatchRejectsInvalidQCTransitions(t *testing.T) {
	batch, err := NewBatch(validBatchInput(func(input *NewBatchInput) {
		input.QCStatus = QCStatusPass
	}))
	if err != nil {
		t.Fatalf("new batch: %v", err)
	}

	if _, err := batch.ChangeQCStatus(QCStatusFail, time.Now()); !errors.Is(err, ErrBatchInvalidQCTransition) {
		t.Fatalf("transition error = %v, want invalid transition", err)
	}
}

func TestBatchAvailabilityBlocksExpiredAndHoldBatches(t *testing.T) {
	holdBatch, err := NewBatch(validBatchInput())
	if err != nil {
		t.Fatalf("new hold batch: %v", err)
	}
	if !holdBatch.BlocksAvailableStock(time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("hold batch should block available stock")
	}

	expiredBatch, err := NewBatch(validBatchInput(func(input *NewBatchInput) {
		input.QCStatus = QCStatusPass
		input.ExpiryDate = time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	}))
	if err != nil {
		t.Fatalf("new expired batch: %v", err)
	}
	if !expiredBatch.BlocksAvailableStock(time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("expired batch should block available stock")
	}
}

func validBatchInput(mutators ...func(*NewBatchInput)) NewBatchInput {
	input := NewBatchInput{
		ID:         "batch-serum-2604a",
		OrgID:      "org-local",
		ItemID:     "item-serum-30ml",
		SKU:        "serum-30ml",
		ItemName:   "Vitamin C Serum",
		BatchNo:    " lot-2604a ",
		SupplierID: "supplier-local",
		MfgDate:    time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		ExpiryDate: time.Date(2027, 4, 1, 0, 0, 0, 0, time.UTC),
		QCStatus:   QCStatusHold,
		Status:     BatchStatusActive,
		CreatedAt:  time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC),
	}
	for _, mutate := range mutators {
		mutate(&input)
	}

	return input
}
