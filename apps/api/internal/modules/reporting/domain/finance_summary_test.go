package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	financedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestNewFinanceSummaryReportBuildsARAPCODAndCashSummary(t *testing.T) {
	filters := mustFinanceSummaryFilters(t, "2026-04-30", "2026-04-30", "2026-04-30")

	report, err := NewFinanceSummaryReport(
		filters,
		[]financedomain.CustomerReceivable{
			financeSummaryReceivable("ar-overdue", financedomain.ReceivableStatusOpen, "1250000.00", "0.00", "1250000.00", "2026-04-20"),
			financeSummaryReceivable("ar-current", financedomain.ReceivableStatusPartiallyPaid, "500000.00", "200000.00", "300000.00", "2026-05-05"),
			financeSummaryReceivable("ar-disputed", financedomain.ReceivableStatusDisputed, "1000000.00", "0.00", "1000000.00", "2026-03-10"),
			financeSummaryReceivable("ar-paid", financedomain.ReceivableStatusPaid, "900000.00", "900000.00", "0.00", "2026-03-01"),
		},
		[]financedomain.SupplierPayable{
			financeSummaryPayable("ap-open", financedomain.PayableStatusOpen, "4250000.00", "0.00", "4250000.00", "2026-04-30"),
			financeSummaryPayable("ap-requested", financedomain.PayableStatusPaymentRequested, "800000.00", "0.00", "800000.00", "2026-04-25"),
			financeSummaryPayable("ap-approved", financedomain.PayableStatusPaymentApproved, "1200000.00", "0.00", "1200000.00", "2026-03-20"),
			financeSummaryPayable("ap-paid", financedomain.PayableStatusPaid, "500000.00", "500000.00", "0.00", "2026-03-01"),
		},
		[]financedomain.CODRemittance{
			financeSummaryCODRemittance("cod-short", financedomain.CODRemittanceStatusDiscrepancy, "2026-04-30", "2000000.00", "1950000.00", "-50000.00", []financedomain.CODDiscrepancy{
				financeSummaryCODDiscrepancy("disc-short", "cod-short-line-1", "ar-overdue", financedomain.CODDiscrepancyTypeShortPaid, financedomain.CODDiscrepancyStatusOpen, "-50000.00"),
			}),
			financeSummaryCODRemittance("cod-closed", financedomain.CODRemittanceStatusClosed, "2026-04-30", "300000.00", "300000.00", "0.00", nil),
		},
		[]financedomain.CashTransaction{
			financeSummaryCashTransaction("cash-in", financedomain.CashTransactionDirectionIn, financedomain.CashTransactionStatusPosted, "2026-04-30", "1250000.00"),
			financeSummaryCashTransaction("cash-out", financedomain.CashTransactionDirectionOut, financedomain.CashTransactionStatusPosted, "2026-04-30", "4250000.00"),
			financeSummaryCashTransaction("cash-draft", financedomain.CashTransactionDirectionIn, financedomain.CashTransactionStatusDraft, "2026-04-30", "999999.00"),
		},
		FinanceSummaryOptions{GeneratedAt: time.Date(2026, 4, 30, 8, 0, 0, 0, time.UTC)},
	)
	if err != nil {
		t.Fatalf("NewFinanceSummaryReport returned error: %v", err)
	}

	if report.Metadata.SourceVersion != ReportingSourceVersion ||
		report.Metadata.Filters.BusinessDateString() != "2026-04-30" ||
		report.CurrencyCode != "VND" {
		t.Fatalf("metadata = %+v currency = %q", report.Metadata, report.CurrencyCode)
	}
	if report.AR.OpenCount != 2 ||
		report.AR.OverdueCount != 2 ||
		report.AR.DisputedCount != 1 ||
		report.AR.OpenAmount != "1550000.00" ||
		report.AR.OverdueAmount != "2250000.00" ||
		report.AR.OutstandingAmount != "2550000.00" {
		t.Fatalf("ar summary = %+v", report.AR)
	}
	requireAgingBucket(t, report.AR.AgingBuckets, "current", 1, "300000.00")
	requireAgingBucket(t, report.AR.AgingBuckets, "8_30", 1, "1250000.00")
	requireAgingBucket(t, report.AR.AgingBuckets, "31_plus", 1, "1000000.00")
	requireFinanceSourceReference(t, report.AR.SourceReferences, "customer_receivable")
	if report.AR.AgingBuckets[0].SourceReference.EntityType != "customer_receivable" ||
		report.AR.AgingBuckets[0].SourceReference.Href == "" {
		t.Fatalf("ar aging source reference = %+v", report.AR.AgingBuckets[0].SourceReference)
	}

	if report.AP.OpenCount != 3 ||
		report.AP.DueCount != 3 ||
		report.AP.PaymentRequestedCount != 1 ||
		report.AP.PaymentApprovedCount != 1 ||
		report.AP.OpenAmount != "6250000.00" ||
		report.AP.DueAmount != "6250000.00" ||
		report.AP.OutstandingAmount != "6250000.00" {
		t.Fatalf("ap summary = %+v", report.AP)
	}
	requireAgingBucket(t, report.AP.AgingBuckets, "current", 1, "4250000.00")
	requireAgingBucket(t, report.AP.AgingBuckets, "1_7", 1, "800000.00")
	requireAgingBucket(t, report.AP.AgingBuckets, "31_plus", 1, "1200000.00")
	requireFinanceSourceReference(t, report.AP.SourceReferences, "supplier_payable")
	requireFinanceSourceReference(t, report.AP.SourceReferences, "payment_approval")

	if report.COD.PendingCount != 1 ||
		report.COD.DiscrepancyCount != 1 ||
		report.COD.PendingAmount != "2000000.00" ||
		report.COD.DiscrepancyAmount != "-50000.00" {
		t.Fatalf("cod summary = %+v", report.COD)
	}
	if len(report.COD.DiscrepancyBuckets) != 1 ||
		report.COD.DiscrepancyBuckets[0].Type != "short_paid" ||
		report.COD.DiscrepancyBuckets[0].Status != "open" ||
		report.COD.DiscrepancyBuckets[0].Amount != "-50000.00" {
		t.Fatalf("cod discrepancy buckets = %+v", report.COD.DiscrepancyBuckets)
	}
	requireFinanceSourceReference(t, report.COD.SourceReferences, "cod_remittance")
	requireFinanceSourceReference(t, report.COD.SourceReferences, "cod_discrepancy")
	if report.COD.DiscrepancyBuckets[0].SourceReference.EntityType != "cod_discrepancy" ||
		report.COD.DiscrepancyBuckets[0].SourceReference.Href == "" {
		t.Fatalf("cod discrepancy source reference = %+v", report.COD.DiscrepancyBuckets[0].SourceReference)
	}
	if report.Cash.TransactionCount != 2 ||
		report.Cash.CashInAmount != "1250000.00" ||
		report.Cash.CashOutAmount != "4250000.00" ||
		report.Cash.NetCashAmount != "-3000000.00" {
		t.Fatalf("cash summary = %+v", report.Cash)
	}
	requireFinanceSourceReference(t, report.Cash.SourceReferences, "cash_transaction")
}

func TestNewFinanceSummaryReportFiltersCODAndCashByDateRange(t *testing.T) {
	filters := mustFinanceSummaryFilters(t, "2026-04-29", "2026-04-30", "2026-04-30")

	report, err := NewFinanceSummaryReport(
		filters,
		nil,
		nil,
		[]financedomain.CODRemittance{
			financeSummaryCODRemittance("cod-in-range", financedomain.CODRemittanceStatusDraft, "2026-04-29", "100000.00", "100000.00", "0.00", nil),
			financeSummaryCODRemittance("cod-out-range", financedomain.CODRemittanceStatusDraft, "2026-04-28", "900000.00", "900000.00", "0.00", nil),
		},
		[]financedomain.CashTransaction{
			financeSummaryCashTransaction("cash-in-range", financedomain.CashTransactionDirectionIn, financedomain.CashTransactionStatusPosted, "2026-04-30", "200000.00"),
			financeSummaryCashTransaction("cash-out-range", financedomain.CashTransactionDirectionIn, financedomain.CashTransactionStatusPosted, "2026-04-28", "700000.00"),
		},
		FinanceSummaryOptions{},
	)
	if err != nil {
		t.Fatalf("NewFinanceSummaryReport returned error: %v", err)
	}

	if report.COD.PendingCount != 1 || report.COD.PendingAmount != "100000.00" {
		t.Fatalf("cod summary = %+v, want only in-range remittance", report.COD)
	}
	if report.Cash.TransactionCount != 1 || report.Cash.CashInAmount != "200000.00" {
		t.Fatalf("cash summary = %+v, want only in-range transaction", report.Cash)
	}
}

func TestNewFinanceSummaryReportBuildsCODDiscrepancyBucketsFromUntracedLines(t *testing.T) {
	filters := mustFinanceSummaryFilters(t, "2026-04-30", "2026-04-30", "2026-04-30")
	remittance := financeSummaryCODRemittance(
		"cod-untraced",
		financedomain.CODRemittanceStatusDraft,
		"2026-04-30",
		"2000000.00",
		"1950000.00",
		"-50000.00",
		nil,
	)
	remittance.Lines = []financedomain.CODRemittanceLine{
		{
			ID:                "cod-untraced-line-1",
			ReceivableID:      "ar-cod-untraced",
			ReceivableNo:      "AR-COD-UNTRACED",
			TrackingNo:        "GHN-UNTRACED",
			ExpectedAmount:    decimal.MustMoneyAmount("2000000.00"),
			RemittedAmount:    decimal.MustMoneyAmount("1950000.00"),
			DiscrepancyAmount: decimal.MustMoneyAmount("-50000.00"),
			MatchStatus:       financedomain.CODLineMatchStatusShortPaid,
		},
	}

	report, err := NewFinanceSummaryReport(filters, nil, nil, []financedomain.CODRemittance{remittance}, nil, FinanceSummaryOptions{})
	if err != nil {
		t.Fatalf("NewFinanceSummaryReport returned error: %v", err)
	}

	if len(report.COD.DiscrepancyBuckets) != 1 {
		t.Fatalf("cod discrepancy buckets = %+v, want one untraced line bucket", report.COD.DiscrepancyBuckets)
	}
	bucket := report.COD.DiscrepancyBuckets[0]
	if bucket.Type != "short_paid" || bucket.Status != "open" || bucket.Count != 1 || bucket.Amount != "-50000.00" {
		t.Fatalf("bucket = %+v, want short paid open discrepancy", bucket)
	}
	if bucket.SourceReference.EntityType != "cod_discrepancy" ||
		!strings.Contains(bucket.SourceReference.ID, "cod-untraced:cod-untraced-line-1") ||
		!strings.Contains(bucket.SourceReference.Href, "source_ids=") ||
		!strings.Contains(bucket.SourceReference.Href, "cod-untraced") {
		t.Fatalf("source reference = %+v, want auditable remittance line source", bucket.SourceReference)
	}
}

func TestNewFinanceSummaryReportRejectsNonVNDRecords(t *testing.T) {
	filters := mustFinanceSummaryFilters(t, "2026-04-30", "2026-04-30", "2026-04-30")
	receivable := financeSummaryReceivable("ar-usd", financedomain.ReceivableStatusOpen, "100.00", "0.00", "100.00", "2026-04-30")
	receivable.CurrencyCode = decimal.CurrencyCode("USD")

	_, err := NewFinanceSummaryReport(filters, []financedomain.CustomerReceivable{receivable}, nil, nil, nil, FinanceSummaryOptions{})
	if !errors.Is(err, ErrInvalidFinanceSummaryReport) {
		t.Fatalf("error = %v, want ErrInvalidFinanceSummaryReport", err)
	}
}

func requireAgingBucket(t *testing.T, buckets []FinanceSummaryAgingBucket, bucket string, count int, amount string) {
	t.Helper()
	for _, current := range buckets {
		if current.Bucket == bucket {
			if current.Count != count || current.Amount != amount {
				t.Fatalf("bucket %s = %+v, want count %d amount %s", bucket, current, count, amount)
			}
			return
		}
	}
	t.Fatalf("bucket %s not found in %+v", bucket, buckets)
}

func requireFinanceSourceReference(t *testing.T, references []ReportSourceReference, entityType string) {
	t.Helper()
	for _, reference := range references {
		if reference.EntityType == entityType {
			if reference.ID == "" || reference.Label == "" || reference.Href == "" || reference.Unavailable {
				t.Fatalf("reference = %+v, want available %s reference", reference, entityType)
			}
			return
		}
	}
	t.Fatalf("references = %+v, missing %s", references, entityType)
}

func mustFinanceSummaryFilters(t *testing.T, fromDate string, toDate string, businessDate string) ReportFilters {
	t.Helper()
	filters, err := NewReportFilters(ReportFilterInput{
		FromDate:     fromDate,
		ToDate:       toDate,
		BusinessDate: businessDate,
	})
	if err != nil {
		t.Fatalf("NewReportFilters returned error: %v", err)
	}

	return filters
}

func financeSummaryReceivable(
	id string,
	status financedomain.ReceivableStatus,
	total string,
	paid string,
	outstanding string,
	dueDate string,
) financedomain.CustomerReceivable {
	return financedomain.CustomerReceivable{
		ID:                id,
		OrgID:             "org-my-pham",
		ReceivableNo:      strings.ToUpper(id),
		CustomerID:        "customer-" + id,
		CustomerName:      "Customer " + id,
		Status:            status,
		TotalAmount:       decimal.MustMoneyAmount(total),
		PaidAmount:        decimal.MustMoneyAmount(paid),
		OutstandingAmount: decimal.MustMoneyAmount(outstanding),
		CurrencyCode:      decimal.CurrencyVND,
		DueDate:           financeSummaryTestDate(dueDate),
		CreatedAt:         financeSummaryTestDate("2026-04-01"),
		CreatedBy:         "finance-user",
		Lines: []financedomain.CustomerReceivableLine{
			{
				ID:          id + "-line",
				Description: "Report fixture",
				Amount:      decimal.MustMoneyAmount(total),
				SourceDocument: financedomain.SourceDocumentRef{
					Type: financedomain.SourceDocumentTypeShipment,
					ID:   "shipment-" + id,
					No:   "SHP-" + strings.ToUpper(id),
				},
			},
		},
	}
}

func financeSummaryPayable(
	id string,
	status financedomain.PayableStatus,
	total string,
	paid string,
	outstanding string,
	dueDate string,
) financedomain.SupplierPayable {
	return financedomain.SupplierPayable{
		ID:                id,
		OrgID:             "org-my-pham",
		PayableNo:         strings.ToUpper(id),
		SupplierID:        "supplier-" + id,
		SupplierName:      "Supplier " + id,
		Status:            status,
		TotalAmount:       decimal.MustMoneyAmount(total),
		PaidAmount:        decimal.MustMoneyAmount(paid),
		OutstandingAmount: decimal.MustMoneyAmount(outstanding),
		CurrencyCode:      decimal.CurrencyVND,
		DueDate:           financeSummaryTestDate(dueDate),
		CreatedAt:         financeSummaryTestDate("2026-04-01"),
		CreatedBy:         "finance-user",
		Lines: []financedomain.SupplierPayableLine{
			{
				ID:          id + "-line",
				Description: "Report fixture",
				Amount:      decimal.MustMoneyAmount(total),
				SourceDocument: financedomain.SourceDocumentRef{
					Type: financedomain.SourceDocumentTypePurchaseOrder,
					ID:   "po-" + id,
					No:   "PO-" + strings.ToUpper(id),
				},
			},
		},
	}
}

func financeSummaryCODRemittance(
	id string,
	status financedomain.CODRemittanceStatus,
	businessDate string,
	expectedAmount string,
	remittedAmount string,
	discrepancyAmount string,
	discrepancies []financedomain.CODDiscrepancy,
) financedomain.CODRemittance {
	return financedomain.CODRemittance{
		ID:                id,
		OrgID:             "org-my-pham",
		RemittanceNo:      strings.ToUpper(id),
		CarrierID:         "carrier-ghn",
		CarrierName:       "GHN",
		Status:            status,
		BusinessDate:      financeSummaryTestDate(businessDate),
		ExpectedAmount:    decimal.MustMoneyAmount(expectedAmount),
		RemittedAmount:    decimal.MustMoneyAmount(remittedAmount),
		DiscrepancyAmount: decimal.MustMoneyAmount(discrepancyAmount),
		CurrencyCode:      decimal.CurrencyVND,
		Discrepancies:     discrepancies,
	}
}

func financeSummaryCODDiscrepancy(
	id string,
	lineID string,
	receivableID string,
	discrepancyType financedomain.CODDiscrepancyType,
	status financedomain.CODDiscrepancyStatus,
	amount string,
) financedomain.CODDiscrepancy {
	return financedomain.CODDiscrepancy{
		ID:           id,
		LineID:       lineID,
		ReceivableID: receivableID,
		Type:         discrepancyType,
		Status:       status,
		Amount:       decimal.MustMoneyAmount(amount),
		Reason:       "Carrier short paid",
		OwnerID:      "finance-user",
		RecordedBy:   "finance-user",
		RecordedAt:   financeSummaryTestDate("2026-04-30"),
	}
}

func financeSummaryCashTransaction(
	id string,
	direction financedomain.CashTransactionDirection,
	status financedomain.CashTransactionStatus,
	businessDate string,
	amount string,
) financedomain.CashTransaction {
	return financedomain.CashTransaction{
		ID:               id,
		OrgID:            "org-my-pham",
		TransactionNo:    strings.ToUpper(id),
		Direction:        direction,
		Status:           status,
		BusinessDate:     financeSummaryTestDate(businessDate),
		CounterpartyName: "Counterparty " + id,
		PaymentMethod:    "bank_transfer",
		TotalAmount:      decimal.MustMoneyAmount(amount),
		CurrencyCode:     decimal.CurrencyVND,
		CreatedAt:        financeSummaryTestDate("2026-04-01"),
		CreatedBy:        "finance-user",
	}
}

func financeSummaryTestDate(value string) time.Time {
	parsed, err := time.ParseInLocation(ReportDateLayout, value, HoChiMinhLocation())
	if err != nil {
		panic(err)
	}

	return parsed
}
