package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewCODRemittanceCalculatesTotalsAndLineStatus(t *testing.T) {
	remittance := mustCODRemittance(t, baseCODRemittanceInput())

	if remittance.Status != CODRemittanceStatusDraft {
		t.Fatalf("status = %q, want draft", remittance.Status)
	}
	if remittance.RemittanceNo != "COD-GHN-260430-0001" || remittance.CarrierCode != "GHN" {
		t.Fatalf("normalized refs = %q %q", remittance.RemittanceNo, remittance.CarrierCode)
	}
	if remittance.ExpectedAmount != "2000000.00" ||
		remittance.RemittedAmount != "1950000.00" ||
		remittance.DiscrepancyAmount != "-50000.00" {
		t.Fatalf(
			"amounts = expected %q remitted %q discrepancy %q",
			remittance.ExpectedAmount,
			remittance.RemittedAmount,
			remittance.DiscrepancyAmount,
		)
	}
	if remittance.Lines[0].MatchStatus != CODLineMatchStatusShortPaid ||
		remittance.Lines[1].MatchStatus != CODLineMatchStatusMatched {
		t.Fatalf("line statuses = %q %q", remittance.Lines[0].MatchStatus, remittance.Lines[1].MatchStatus)
	}
}

func TestNewCODRemittanceRejectsAmountMismatchAndUnsafeMoney(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*NewCODRemittanceInput)
		wantErr error
	}{
		{
			name: "header expected mismatch",
			mutate: func(input *NewCODRemittanceInput) {
				input.ExpectedAmount = "2000000.01"
			},
			wantErr: ErrCODRemittanceInvalidAmount,
		},
		{
			name: "too much money scale",
			mutate: func(input *NewCODRemittanceInput) {
				input.Lines[0].ExpectedAmount = "1250000.001"
			},
			wantErr: ErrCODRemittanceInvalidAmount,
		},
		{
			name: "non VND currency",
			mutate: func(input *NewCODRemittanceInput) {
				input.CurrencyCode = "USD"
			},
			wantErr: ErrCODRemittanceInvalidAmount,
		},
		{
			name: "negative remitted amount",
			mutate: func(input *NewCODRemittanceInput) {
				input.Lines[0].RemittedAmount = "-1"
			},
			wantErr: ErrCODRemittanceInvalidAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := baseCODRemittanceInput()
			tt.mutate(&input)
			_, err := NewCODRemittance(input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestCODRemittanceRequiresDiscrepancyTraceBeforeSubmit(t *testing.T) {
	remittance := mustCODRemittance(t, baseCODRemittanceInput())

	_, err := remittance.MarkMatching("finance-user", time.Now())
	if !errors.Is(err, ErrCODRemittanceInvalidDiscrepancy) {
		t.Fatalf("mark matching error = %v, want %v", err, ErrCODRemittanceInvalidDiscrepancy)
	}

	_, err = remittance.Submit("finance-user", time.Now())
	if !errors.Is(err, ErrCODRemittanceInvalidTransition) {
		t.Fatalf("submit draft error = %v, want %v", err, ErrCODRemittanceInvalidTransition)
	}

	withDiscrepancy, err := remittance.RecordDiscrepancy(RecordCODDiscrepancyInput{
		ID:      "cod-disc-1",
		LineID:  "cod-line-1",
		Reason:  "carrier remitted short",
		OwnerID: "finance-user",
	})
	if err != nil {
		t.Fatalf("record discrepancy: %v", err)
	}
	if withDiscrepancy.Status != CODRemittanceStatusDiscrepancy ||
		len(withDiscrepancy.Discrepancies) != 1 ||
		withDiscrepancy.Discrepancies[0].Amount != "-50000.00" {
		t.Fatalf("discrepancy remittance = %+v", withDiscrepancy)
	}

	submitted, err := withDiscrepancy.Submit("finance-lead", time.Date(2026, 4, 30, 11, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("submit with trace: %v", err)
	}
	if submitted.Status != CODRemittanceStatusSubmitted || submitted.SubmittedBy != "finance-lead" {
		t.Fatalf("submitted remittance = %+v", submitted)
	}
}

func TestCODRemittanceMatchedFlowSubmitApproveClose(t *testing.T) {
	input := baseCODRemittanceInput()
	input.Lines[0].RemittedAmount = "1250000.00"
	input.RemittedAmount = "2000000.00"
	remittance := mustCODRemittance(t, input)
	baseTime := time.Date(2026, 4, 30, 11, 30, 0, 0, time.UTC)

	matching, err := remittance.MarkMatching("finance-user", baseTime)
	if err != nil {
		t.Fatalf("mark matching: %v", err)
	}
	if matching.Status != CODRemittanceStatusMatching {
		t.Fatalf("status = %q, want matching", matching.Status)
	}

	submitted, err := matching.Submit("finance-user", baseTime.Add(time.Hour))
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	approved, err := submitted.Approve("finance-lead", baseTime.Add(2*time.Hour))
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	closed, err := approved.Close("finance-lead", baseTime.Add(3*time.Hour))
	if err != nil {
		t.Fatalf("close: %v", err)
	}
	if closed.Status != CODRemittanceStatusClosed ||
		closed.Version != 5 ||
		closed.ClosedBy != "finance-lead" {
		t.Fatalf("closed remittance = %+v", closed)
	}
}

func TestCODRemittanceRejectsInvalidTransitions(t *testing.T) {
	remittance := mustCODRemittance(t, baseCODRemittanceInput())

	_, err := remittance.Approve("finance-lead", time.Now())
	if !errors.Is(err, ErrCODRemittanceInvalidTransition) {
		t.Fatalf("approve draft error = %v, want %v", err, ErrCODRemittanceInvalidTransition)
	}

	voided, err := remittance.Void("finance-user", "duplicate carrier file", time.Now())
	if err != nil {
		t.Fatalf("void draft: %v", err)
	}
	if voided.Status != CODRemittanceStatusVoid || voided.VoidReason != "duplicate carrier file" {
		t.Fatalf("voided remittance = %+v", voided)
	}

	_, err = voided.Submit("finance-user", time.Now())
	if !errors.Is(err, ErrCODRemittanceInvalidTransition) {
		t.Fatalf("submit void error = %v, want %v", err, ErrCODRemittanceInvalidTransition)
	}
}

func mustCODRemittance(t *testing.T, input NewCODRemittanceInput) CODRemittance {
	t.Helper()
	remittance, err := NewCODRemittance(input)
	if err != nil {
		t.Fatalf("new cod remittance: %v", err)
	}

	return remittance
}

func baseCODRemittanceInput() NewCODRemittanceInput {
	return NewCODRemittanceInput{
		ID:             "cod-remit-260430-0001",
		OrgID:          "org-my-pham",
		RemittanceNo:   "cod-ghn-260430-0001",
		CarrierID:      "carrier-ghn",
		CarrierCode:    "ghn",
		CarrierName:    "GHN Express",
		BusinessDate:   time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
		ExpectedAmount: "2000000.00",
		RemittedAmount: "1950000.00",
		CurrencyCode:   "VND",
		CreatedAt:      time.Date(2026, 4, 30, 9, 0, 0, 0, time.UTC),
		CreatedBy:      "finance-user",
		Lines: []NewCODRemittanceLineInput{
			{
				ID:             "cod-line-1",
				ReceivableID:   "ar-cod-260430-0001",
				ReceivableNo:   "ar-cod-260430-0001",
				ShipmentID:     "ship-hcm-260430-0001",
				TrackingNo:     "ghn260430001",
				CustomerName:   "My Pham HCM Retail",
				ExpectedAmount: "1250000.00",
				RemittedAmount: "1200000.00",
			},
			{
				ID:             "cod-line-2",
				ReceivableID:   "ar-cod-260430-0002",
				ReceivableNo:   "ar-cod-260430-0002",
				ShipmentID:     "ship-hcm-260430-0002",
				TrackingNo:     "ghn260430002",
				CustomerName:   "Marketplace COD",
				ExpectedAmount: "750000.00",
				RemittedAmount: "750000.00",
			},
		},
	}
}
