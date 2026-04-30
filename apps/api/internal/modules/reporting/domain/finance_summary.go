package domain

import (
	"errors"
	"math/big"
	"net/url"
	"sort"
	"strings"
	"time"

	financedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrInvalidFinanceSummaryReport = errors.New("finance summary report is invalid")

type FinanceSummaryOptions struct {
	GeneratedAt time.Time
}

type FinanceSummaryReport struct {
	Metadata     ReportMetadata
	CurrencyCode string
	AR           FinanceSummaryReceivable
	AP           FinanceSummaryPayable
	COD          FinanceSummaryCOD
	Cash         FinanceSummaryCash
}

type FinanceSummaryReceivable struct {
	OpenCount         int
	OverdueCount      int
	DisputedCount     int
	OpenAmount        string
	OverdueAmount     string
	OutstandingAmount string
	AgingBuckets      []FinanceSummaryAgingBucket
	SourceReferences  []ReportSourceReference
}

type FinanceSummaryPayable struct {
	OpenCount             int
	DueCount              int
	PaymentRequestedCount int
	PaymentApprovedCount  int
	OpenAmount            string
	DueAmount             string
	OutstandingAmount     string
	AgingBuckets          []FinanceSummaryAgingBucket
	SourceReferences      []ReportSourceReference
}

type FinanceSummaryCOD struct {
	PendingCount       int
	DiscrepancyCount   int
	PendingAmount      string
	DiscrepancyAmount  string
	DiscrepancyBuckets []FinanceSummaryDiscrepancyBucket
	SourceReferences   []ReportSourceReference
}

type FinanceSummaryCash struct {
	TransactionCount int
	CashInAmount     string
	CashOutAmount    string
	NetCashAmount    string
	SourceReferences []ReportSourceReference
}

type FinanceSummaryAgingBucket struct {
	Bucket          string
	Count           int
	Amount          string
	SourceReference ReportSourceReference
}

type FinanceSummaryDiscrepancyBucket struct {
	Type            string
	Status          string
	Count           int
	Amount          string
	SourceReference ReportSourceReference
	sourceIDs       []string
}

func NewFinanceSummaryReport(
	filters ReportFilters,
	receivables []financedomain.CustomerReceivable,
	payables []financedomain.SupplierPayable,
	remittances []financedomain.CODRemittance,
	cashTransactions []financedomain.CashTransaction,
	options FinanceSummaryOptions,
) (FinanceSummaryReport, error) {
	generatedAt := options.GeneratedAt
	if generatedAt.IsZero() {
		generatedAt = time.Now().UTC()
	}

	ar, err := summarizeFinanceSummaryReceivables(filters, receivables)
	if err != nil {
		return FinanceSummaryReport{}, err
	}
	ap, err := summarizeFinanceSummaryPayables(filters, payables)
	if err != nil {
		return FinanceSummaryReport{}, err
	}
	cod, err := summarizeFinanceSummaryCOD(filters, remittances)
	if err != nil {
		return FinanceSummaryReport{}, err
	}
	cash, err := summarizeFinanceSummaryCash(filters, cashTransactions)
	if err != nil {
		return FinanceSummaryReport{}, err
	}

	return FinanceSummaryReport{
		Metadata:     NewReportMetadata(filters, generatedAt),
		CurrencyCode: decimal.CurrencyVND.String(),
		AR:           ar,
		AP:           ap,
		COD:          cod,
		Cash:         cash,
	}, nil
}

func summarizeFinanceSummaryReceivables(
	filters ReportFilters,
	receivables []financedomain.CustomerReceivable,
) (FinanceSummaryReceivable, error) {
	openAmount := newFinanceSummaryMoneyTotal()
	overdueAmount := newFinanceSummaryMoneyTotal()
	outstandingAmount := newFinanceSummaryMoneyTotal()
	aging := newFinanceSummaryAgingBuckets()
	var summary FinanceSummaryReceivable

	for _, receivable := range receivables {
		if receivable.CurrencyCode != decimal.CurrencyVND {
			return FinanceSummaryReceivable{}, ErrInvalidFinanceSummaryReport
		}
		status := financedomain.NormalizeReceivableStatus(receivable.Status)
		if status == financedomain.ReceivableStatusPaid || status == financedomain.ReceivableStatusVoid {
			continue
		}
		if err := outstandingAmount.Add(receivable.OutstandingAmount); err != nil {
			return FinanceSummaryReceivable{}, ErrInvalidFinanceSummaryReport
		}
		if err := aging.Add(receivable.DueDate, filters.BusinessDate, receivable.OutstandingAmount); err != nil {
			return FinanceSummaryReceivable{}, ErrInvalidFinanceSummaryReport
		}
		if status == financedomain.ReceivableStatusOpen || status == financedomain.ReceivableStatusPartiallyPaid {
			summary.OpenCount++
			if err := openAmount.Add(receivable.OutstandingAmount); err != nil {
				return FinanceSummaryReceivable{}, ErrInvalidFinanceSummaryReport
			}
		}
		if status == financedomain.ReceivableStatusDisputed {
			summary.DisputedCount++
		}
		if isFinanceSummaryPastDue(receivable.DueDate, filters.BusinessDate) {
			summary.OverdueCount++
			if err := overdueAmount.Add(receivable.OutstandingAmount); err != nil {
				return FinanceSummaryReceivable{}, ErrInvalidFinanceSummaryReport
			}
		}
	}

	summary.OpenAmount = openAmount.String()
	summary.OverdueAmount = overdueAmount.String()
	summary.OutstandingAmount = outstandingAmount.String()
	summary.AgingBuckets = financeSummaryAgingBucketReferences("customer_receivable", filters, aging.Rows())
	summary.SourceReferences = financeSummarySectionReferences(filters, []financeSummaryReferenceSeed{
		{entityType: "customer_receivable", label: "customer_receivables"},
	})

	return summary, nil
}

func summarizeFinanceSummaryPayables(
	filters ReportFilters,
	payables []financedomain.SupplierPayable,
) (FinanceSummaryPayable, error) {
	openAmount := newFinanceSummaryMoneyTotal()
	dueAmount := newFinanceSummaryMoneyTotal()
	outstandingAmount := newFinanceSummaryMoneyTotal()
	aging := newFinanceSummaryAgingBuckets()
	var summary FinanceSummaryPayable

	for _, payable := range payables {
		if payable.CurrencyCode != decimal.CurrencyVND {
			return FinanceSummaryPayable{}, ErrInvalidFinanceSummaryReport
		}
		status := financedomain.NormalizePayableStatus(payable.Status)
		if status == financedomain.PayableStatusPaid || status == financedomain.PayableStatusVoid {
			continue
		}
		if err := outstandingAmount.Add(payable.OutstandingAmount); err != nil {
			return FinanceSummaryPayable{}, ErrInvalidFinanceSummaryReport
		}
		if err := aging.Add(payable.DueDate, filters.BusinessDate, payable.OutstandingAmount); err != nil {
			return FinanceSummaryPayable{}, ErrInvalidFinanceSummaryReport
		}
		if status == financedomain.PayableStatusPaymentRequested {
			summary.PaymentRequestedCount++
		}
		if status == financedomain.PayableStatusPaymentApproved {
			summary.PaymentApprovedCount++
		}
		if isFinanceSummaryOpenPayableStatus(status) {
			summary.OpenCount++
			if err := openAmount.Add(payable.OutstandingAmount); err != nil {
				return FinanceSummaryPayable{}, ErrInvalidFinanceSummaryReport
			}
		}
		if !financeSummaryDateOnly(payable.DueDate).After(financeSummaryDateOnly(filters.BusinessDate)) {
			summary.DueCount++
			if err := dueAmount.Add(payable.OutstandingAmount); err != nil {
				return FinanceSummaryPayable{}, ErrInvalidFinanceSummaryReport
			}
		}
	}

	summary.OpenAmount = openAmount.String()
	summary.DueAmount = dueAmount.String()
	summary.OutstandingAmount = outstandingAmount.String()
	summary.AgingBuckets = financeSummaryAgingBucketReferences("supplier_payable", filters, aging.Rows())
	summary.SourceReferences = financeSummarySectionReferences(filters, []financeSummaryReferenceSeed{
		{entityType: "supplier_payable", label: "supplier_payables"},
		{entityType: "payment_approval", label: "payment_approvals"},
	})

	return summary, nil
}

func summarizeFinanceSummaryCOD(
	filters ReportFilters,
	remittances []financedomain.CODRemittance,
) (FinanceSummaryCOD, error) {
	pendingAmount := newFinanceSummaryMoneyTotal()
	discrepancyAmount := newFinanceSummaryMoneyTotal()
	discrepancies := newFinanceSummaryDiscrepancyBuckets()
	var summary FinanceSummaryCOD

	for _, remittance := range remittances {
		if remittance.CurrencyCode != decimal.CurrencyVND {
			return FinanceSummaryCOD{}, ErrInvalidFinanceSummaryReport
		}
		if !filters.IncludesBusinessDate(remittance.BusinessDate) {
			continue
		}
		status := financedomain.NormalizeCODRemittanceStatus(remittance.Status)
		if status != financedomain.CODRemittanceStatusClosed && status != financedomain.CODRemittanceStatusVoid {
			summary.PendingCount++
			if err := pendingAmount.Add(remittance.ExpectedAmount); err != nil {
				return FinanceSummaryCOD{}, ErrInvalidFinanceSummaryReport
			}
		}
		if !remittance.DiscrepancyAmount.IsZero() || len(remittance.Discrepancies) > 0 {
			summary.DiscrepancyCount++
			if err := discrepancyAmount.Add(remittance.DiscrepancyAmount); err != nil {
				return FinanceSummaryCOD{}, ErrInvalidFinanceSummaryReport
			}
		}
		tracedLineIDs := make(map[string]struct{}, len(remittance.Discrepancies))
		for _, discrepancy := range remittance.Discrepancies {
			if err := discrepancies.Add(discrepancy); err != nil {
				return FinanceSummaryCOD{}, ErrInvalidFinanceSummaryReport
			}
			tracedLineIDs[strings.TrimSpace(discrepancy.LineID)] = struct{}{}
		}
		for _, line := range remittance.Lines {
			if line.DiscrepancyAmount.IsZero() {
				continue
			}
			if _, ok := tracedLineIDs[strings.TrimSpace(line.ID)]; ok {
				continue
			}
			if err := discrepancies.AddUntracedLine(remittance, line); err != nil {
				return FinanceSummaryCOD{}, ErrInvalidFinanceSummaryReport
			}
		}
	}

	summary.PendingAmount = pendingAmount.String()
	summary.DiscrepancyAmount = discrepancyAmount.String()
	summary.DiscrepancyBuckets = financeSummaryDiscrepancyBucketReferences(filters, discrepancies.Rows())
	summary.SourceReferences = financeSummarySectionReferences(filters, []financeSummaryReferenceSeed{
		{entityType: "cod_remittance", label: "cod_remittances"},
		{entityType: "cod_discrepancy", label: "cod_discrepancies"},
	})

	return summary, nil
}

func summarizeFinanceSummaryCash(
	filters ReportFilters,
	transactions []financedomain.CashTransaction,
) (FinanceSummaryCash, error) {
	cashIn := newFinanceSummaryMoneyTotal()
	cashOut := newFinanceSummaryMoneyTotal()
	net := newFinanceSummaryMoneyTotal()
	var summary FinanceSummaryCash

	for _, transaction := range transactions {
		if transaction.CurrencyCode != decimal.CurrencyVND {
			return FinanceSummaryCash{}, ErrInvalidFinanceSummaryReport
		}
		if financedomain.NormalizeCashTransactionStatus(transaction.Status) != financedomain.CashTransactionStatusPosted ||
			!filters.IncludesBusinessDate(transaction.BusinessDate) {
			continue
		}
		summary.TransactionCount++
		switch financedomain.NormalizeCashTransactionDirection(transaction.Direction) {
		case financedomain.CashTransactionDirectionIn:
			if err := cashIn.Add(transaction.TotalAmount); err != nil {
				return FinanceSummaryCash{}, ErrInvalidFinanceSummaryReport
			}
			if err := net.Add(transaction.TotalAmount); err != nil {
				return FinanceSummaryCash{}, ErrInvalidFinanceSummaryReport
			}
		case financedomain.CashTransactionDirectionOut:
			if err := cashOut.Add(transaction.TotalAmount); err != nil {
				return FinanceSummaryCash{}, ErrInvalidFinanceSummaryReport
			}
			if err := net.Subtract(transaction.TotalAmount); err != nil {
				return FinanceSummaryCash{}, ErrInvalidFinanceSummaryReport
			}
		default:
			return FinanceSummaryCash{}, ErrInvalidFinanceSummaryReport
		}
	}

	summary.CashInAmount = cashIn.String()
	summary.CashOutAmount = cashOut.String()
	summary.NetCashAmount = net.String()
	summary.SourceReferences = financeSummarySectionReferences(filters, []financeSummaryReferenceSeed{
		{entityType: "cash_transaction", label: "cash_transactions"},
	})

	return summary, nil
}

type financeSummaryReferenceSeed struct {
	entityType string
	label      string
}

func financeSummarySectionReferences(filters ReportFilters, seeds []financeSummaryReferenceSeed) []ReportSourceReference {
	references := make([]ReportSourceReference, 0, len(seeds))
	for _, seed := range seeds {
		references = append(references, financeSummarySourceReference(
			seed.entityType,
			seed.entityType+":"+filters.FromDateString()+":"+filters.ToDateString()+":"+filters.BusinessDateString(),
			seed.label,
			financeSummarySourceHref(seed.entityType, filters, nil),
		))
	}

	return references
}

func financeSummaryAgingBucketReferences(
	entityType string,
	filters ReportFilters,
	buckets []FinanceSummaryAgingBucket,
) []FinanceSummaryAgingBucket {
	rows := make([]FinanceSummaryAgingBucket, 0, len(buckets))
	for _, bucket := range buckets {
		bucket.SourceReference = financeSummarySourceReference(
			entityType,
			entityType+":"+bucket.Bucket+":"+filters.FromDateString()+":"+filters.ToDateString()+":"+filters.BusinessDateString(),
			bucket.Bucket,
			financeSummarySourceHref(entityType, filters, map[string]string{"bucket": bucket.Bucket}),
		)
		rows = append(rows, bucket)
	}

	return rows
}

func financeSummaryDiscrepancyBucketReferences(
	filters ReportFilters,
	buckets []FinanceSummaryDiscrepancyBucket,
) []FinanceSummaryDiscrepancyBucket {
	rows := make([]FinanceSummaryDiscrepancyBucket, 0, len(buckets))
	for _, bucket := range buckets {
		bucket.SourceReference = financeSummarySourceReference(
			"cod_discrepancy",
			"cod_discrepancy:"+bucket.Type+":"+bucket.Status+":"+bucket.sourceKey()+":"+filters.FromDateString()+":"+filters.ToDateString(),
			bucket.Type+":"+bucket.Status,
			financeSummarySourceHref("cod_discrepancy", filters, map[string]string{
				"type":       bucket.Type,
				"status":     bucket.Status,
				"source_ids": bucket.sourceKey(),
			}),
		)
		rows = append(rows, bucket)
	}

	return rows
}

func financeSummarySourceReference(entityType string, id string, label string, href string) ReportSourceReference {
	reference, err := NewReportSourceReference(ReportSourceReferenceInput{
		EntityType: entityType,
		ID:         id,
		Label:      label,
		Href:       href,
	})
	if err != nil {
		panic(ErrInvalidFinanceSummaryReport)
	}

	return reference
}

func financeSummarySourceHref(entityType string, filters ReportFilters, extra map[string]string) string {
	params := url.Values{}
	params.Set("source_type", entityType)
	params.Set("from_date", filters.FromDateString())
	params.Set("to_date", filters.ToDateString())
	params.Set("business_date", filters.BusinessDateString())
	for key, value := range extra {
		params.Set(key, value)
	}

	return "/finance?" + params.Encode()
}

func isFinanceSummaryOpenPayableStatus(status financedomain.PayableStatus) bool {
	switch status {
	case financedomain.PayableStatusOpen,
		financedomain.PayableStatusPaymentRequested,
		financedomain.PayableStatusPaymentApproved,
		financedomain.PayableStatusPartiallyPaid:
		return true
	default:
		return false
	}
}

func isFinanceSummaryPastDue(dueDate time.Time, asOf time.Time) bool {
	if dueDate.IsZero() {
		return false
	}

	return financeSummaryDateOnly(dueDate).Before(financeSummaryDateOnly(asOf))
}

func financeSummaryAgingBucket(dueDate time.Time, asOf time.Time) string {
	days := int(financeSummaryDateOnly(asOf).Sub(financeSummaryDateOnly(dueDate)).Hours() / 24)
	switch {
	case days <= 0:
		return "current"
	case days <= 7:
		return "1_7"
	case days <= 30:
		return "8_30"
	default:
		return "31_plus"
	}
}

func financeSummaryDateOnly(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}
	local := value.In(HoChiMinhLocation())

	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, HoChiMinhLocation())
}

type financeSummaryMoneyTotal struct {
	cents *big.Int
}

func newFinanceSummaryMoneyTotal() financeSummaryMoneyTotal {
	return financeSummaryMoneyTotal{cents: big.NewInt(0)}
}

func (t *financeSummaryMoneyTotal) Add(amount decimal.Decimal) error {
	cents, err := financeSummaryMoneyCents(amount)
	if err != nil {
		return err
	}
	t.cents.Add(t.cents, cents)

	return nil
}

func (t *financeSummaryMoneyTotal) Subtract(amount decimal.Decimal) error {
	cents, err := financeSummaryMoneyCents(amount)
	if err != nil {
		return err
	}
	t.cents.Sub(t.cents, cents)

	return nil
}

func (t financeSummaryMoneyTotal) String() string {
	value := new(big.Int).Set(t.cents)
	negative := value.Sign() < 0
	if negative {
		value.Abs(value)
	}
	digits := value.String()
	if len(digits) <= decimal.MoneyScale {
		digits = strings.Repeat("0", decimal.MoneyScale-len(digits)+1) + digits
	}
	integer := digits[:len(digits)-decimal.MoneyScale]
	fraction := digits[len(digits)-decimal.MoneyScale:]
	if negative && value.Sign() != 0 {
		return "-" + integer + "." + fraction
	}

	return integer + "." + fraction
}

func financeSummaryMoneyCents(amount decimal.Decimal) (*big.Int, error) {
	normalized, err := decimal.ParseMoneyAmount(amount.String())
	if err != nil {
		return nil, err
	}
	digits := strings.ReplaceAll(normalized.String(), ".", "")
	negative := strings.HasPrefix(digits, "-")
	digits = strings.TrimPrefix(digits, "-")
	cents, ok := new(big.Int).SetString(digits, 10)
	if !ok {
		return nil, decimal.ErrInvalidDecimal
	}
	if negative {
		cents.Neg(cents)
	}

	return cents, nil
}

type financeSummaryAgingBuckets struct {
	values map[string]*financeSummaryAgingBucketTotal
}

type financeSummaryAgingBucketTotal struct {
	count  int
	amount financeSummaryMoneyTotal
}

func newFinanceSummaryAgingBuckets() financeSummaryAgingBuckets {
	values := make(map[string]*financeSummaryAgingBucketTotal)
	for _, bucket := range []string{"current", "1_7", "8_30", "31_plus"} {
		values[bucket] = &financeSummaryAgingBucketTotal{amount: newFinanceSummaryMoneyTotal()}
	}

	return financeSummaryAgingBuckets{values: values}
}

func (b financeSummaryAgingBuckets) Add(dueDate time.Time, asOf time.Time, amount decimal.Decimal) error {
	bucket := financeSummaryAgingBucket(dueDate, asOf)
	total := b.values[bucket]
	if total == nil {
		total = &financeSummaryAgingBucketTotal{amount: newFinanceSummaryMoneyTotal()}
		b.values[bucket] = total
	}
	total.count++

	return total.amount.Add(amount)
}

func (b financeSummaryAgingBuckets) Rows() []FinanceSummaryAgingBucket {
	order := map[string]int{
		"current": 1,
		"1_7":     2,
		"8_30":    3,
		"31_plus": 4,
	}
	rows := make([]FinanceSummaryAgingBucket, 0, len(b.values))
	for bucket, total := range b.values {
		rows = append(rows, FinanceSummaryAgingBucket{
			Bucket: bucket,
			Count:  total.count,
			Amount: total.amount.String(),
		})
	}
	sort.Slice(rows, func(i int, j int) bool {
		return order[rows[i].Bucket] < order[rows[j].Bucket]
	})

	return rows
}

type financeSummaryDiscrepancyBuckets struct {
	values map[string]*financeSummaryDiscrepancyBucketTotal
}

type financeSummaryDiscrepancyBucketTotal struct {
	discrepancyType string
	status          string
	count           int
	amount          financeSummaryMoneyTotal
	sourceIDs       map[string]struct{}
}

func newFinanceSummaryDiscrepancyBuckets() financeSummaryDiscrepancyBuckets {
	return financeSummaryDiscrepancyBuckets{values: make(map[string]*financeSummaryDiscrepancyBucketTotal)}
}

func (b financeSummaryDiscrepancyBuckets) Add(discrepancy financedomain.CODDiscrepancy) error {
	discrepancyType := string(financedomain.NormalizeCODDiscrepancyType(discrepancy.Type))
	status := string(financedomain.NormalizeCODDiscrepancyStatus(discrepancy.Status))
	if discrepancyType == "" || status == "" {
		return ErrInvalidFinanceSummaryReport
	}

	return b.add(discrepancyType, status, discrepancy.Amount, discrepancy.ID)
}

func (b financeSummaryDiscrepancyBuckets) AddUntracedLine(
	remittance financedomain.CODRemittance,
	line financedomain.CODRemittanceLine,
) error {
	discrepancyType := financeSummaryCODLineDiscrepancyType(line)
	if discrepancyType == "" || line.DiscrepancyAmount.IsZero() {
		return nil
	}

	return b.add(
		discrepancyType,
		string(financedomain.CODDiscrepancyStatusOpen),
		line.DiscrepancyAmount,
		strings.TrimSpace(remittance.ID)+":"+strings.TrimSpace(line.ID),
	)
}

func (b financeSummaryDiscrepancyBuckets) add(
	discrepancyType string,
	status string,
	amount decimal.Decimal,
	sourceID string,
) error {
	key := discrepancyType + ":" + status
	total := b.values[key]
	if total == nil {
		total = &financeSummaryDiscrepancyBucketTotal{
			discrepancyType: discrepancyType,
			status:          status,
			amount:          newFinanceSummaryMoneyTotal(),
			sourceIDs:       make(map[string]struct{}),
		}
		b.values[key] = total
	}
	total.count++
	if normalizedSourceID := strings.TrimSpace(sourceID); normalizedSourceID != "" {
		total.sourceIDs[normalizedSourceID] = struct{}{}
	}

	return total.amount.Add(amount)
}

func (b financeSummaryDiscrepancyBuckets) Rows() []FinanceSummaryDiscrepancyBucket {
	rows := make([]FinanceSummaryDiscrepancyBucket, 0, len(b.values))
	for _, total := range b.values {
		rows = append(rows, FinanceSummaryDiscrepancyBucket{
			Type:   total.discrepancyType,
			Status: total.status,
			Count:  total.count,
			Amount: total.amount.String(),
			sourceIDs: financeSummarySortedSourceIDs(
				total.sourceIDs,
				total.discrepancyType+":"+total.status,
			),
		})
	}
	sort.Slice(rows, func(i int, j int) bool {
		if rows[i].Type != rows[j].Type {
			return rows[i].Type < rows[j].Type
		}

		return rows[i].Status < rows[j].Status
	})

	return rows
}

func financeSummaryCODLineDiscrepancyType(line financedomain.CODRemittanceLine) string {
	switch financedomain.NormalizeCODLineMatchStatus(line.MatchStatus) {
	case financedomain.CODLineMatchStatusShortPaid:
		return string(financedomain.CODDiscrepancyTypeShortPaid)
	case financedomain.CODLineMatchStatusOverPaid:
		return string(financedomain.CODDiscrepancyTypeOverPaid)
	}
	if line.DiscrepancyAmount.IsNegative() {
		return string(financedomain.CODDiscrepancyTypeShortPaid)
	}
	if !line.DiscrepancyAmount.IsZero() {
		return string(financedomain.CODDiscrepancyTypeOverPaid)
	}

	return ""
}

func financeSummarySortedSourceIDs(values map[string]struct{}, fallback string) []string {
	sourceIDs := make([]string, 0, len(values))
	for sourceID := range values {
		sourceIDs = append(sourceIDs, sourceID)
	}
	sort.Strings(sourceIDs)
	if len(sourceIDs) == 0 {
		sourceIDs = append(sourceIDs, fallback)
	}

	return sourceIDs
}

func (bucket FinanceSummaryDiscrepancyBucket) sourceKey() string {
	return strings.Join(bucket.sourceIDs, ",")
}
