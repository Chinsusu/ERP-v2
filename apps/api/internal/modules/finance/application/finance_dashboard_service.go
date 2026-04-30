package application

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"time"

	financedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
	apperrors "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/errors"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

const ErrorCodeFinanceDashboardValidation response.ErrorCode = "FINANCE_DASHBOARD_VALIDATION_ERROR"

type FinanceDashboardService struct {
	customerReceivables CustomerReceivableStore
	supplierPayables    SupplierPayableStore
	codRemittances      CODRemittanceStore
	cashTransactions    CashTransactionStore
	clock               func() time.Time
}

type FinanceDashboardFilter struct {
	BusinessDate string
}

type FinanceDashboardMetrics struct {
	BusinessDate string
	GeneratedAt  time.Time
	CurrencyCode string
	AR           FinanceDashboardReceivableMetrics
	AP           FinanceDashboardPayableMetrics
	COD          FinanceDashboardCODMetrics
	Cash         FinanceDashboardCashMetrics
}

type FinanceDashboardReceivableMetrics struct {
	OpenCount         int
	OverdueCount      int
	DisputedCount     int
	OpenAmount        string
	OverdueAmount     string
	OutstandingAmount string
}

type FinanceDashboardPayableMetrics struct {
	OpenCount             int
	DueCount              int
	PaymentRequestedCount int
	PaymentApprovedCount  int
	OpenAmount            string
	DueAmount             string
	OutstandingAmount     string
}

type FinanceDashboardCODMetrics struct {
	PendingCount      int
	DiscrepancyCount  int
	PendingAmount     string
	DiscrepancyAmount string
}

type FinanceDashboardCashMetrics struct {
	TransactionCount int
	CashInToday      string
	CashOutToday     string
	NetCashToday     string
}

type moneyAccumulator struct {
	cents *big.Int
}

func NewFinanceDashboardService(
	customerReceivables CustomerReceivableStore,
	supplierPayables SupplierPayableStore,
	codRemittances CODRemittanceStore,
	cashTransactions CashTransactionStore,
) FinanceDashboardService {
	return FinanceDashboardService{
		customerReceivables: customerReceivables,
		supplierPayables:    supplierPayables,
		codRemittances:      codRemittances,
		cashTransactions:    cashTransactions,
		clock:               func() time.Time { return time.Now().UTC() },
	}
}

func (s FinanceDashboardService) WithClock(clock func() time.Time) FinanceDashboardService {
	if clock != nil {
		s.clock = clock
	}

	return s
}

func (s FinanceDashboardService) GetFinanceDashboardMetrics(
	ctx context.Context,
	filter FinanceDashboardFilter,
) (FinanceDashboardMetrics, error) {
	if s.customerReceivables == nil ||
		s.supplierPayables == nil ||
		s.codRemittances == nil ||
		s.cashTransactions == nil {
		return FinanceDashboardMetrics{}, errors.New("finance dashboard stores are required")
	}
	businessDate, err := normalizeFinanceDashboardBusinessDate(filter.BusinessDate, s.clock())
	if err != nil {
		return FinanceDashboardMetrics{}, financeDashboardValidationError(err, map[string]any{"field": "business_date"})
	}

	receivables, err := s.customerReceivables.List(ctx, CustomerReceivableFilter{})
	if err != nil {
		return FinanceDashboardMetrics{}, err
	}
	payables, err := s.supplierPayables.List(ctx, SupplierPayableFilter{})
	if err != nil {
		return FinanceDashboardMetrics{}, err
	}
	remittances, err := s.codRemittances.List(ctx, CODRemittanceFilter{})
	if err != nil {
		return FinanceDashboardMetrics{}, err
	}
	cashTransactions, err := s.cashTransactions.List(ctx, CashTransactionFilter{})
	if err != nil {
		return FinanceDashboardMetrics{}, err
	}

	ar, err := summarizeFinanceDashboardReceivables(receivables, businessDate)
	if err != nil {
		return FinanceDashboardMetrics{}, err
	}
	ap, err := summarizeFinanceDashboardPayables(payables, businessDate)
	if err != nil {
		return FinanceDashboardMetrics{}, err
	}
	cod, err := summarizeFinanceDashboardCOD(remittances)
	if err != nil {
		return FinanceDashboardMetrics{}, err
	}
	cash, err := summarizeFinanceDashboardCash(cashTransactions, businessDate)
	if err != nil {
		return FinanceDashboardMetrics{}, err
	}

	return FinanceDashboardMetrics{
		BusinessDate: businessDate,
		GeneratedAt:  s.clock().UTC(),
		CurrencyCode: decimal.CurrencyVND.String(),
		AR:           ar,
		AP:           ap,
		COD:          cod,
		Cash:         cash,
	}, nil
}

func summarizeFinanceDashboardReceivables(
	receivables []financedomain.CustomerReceivable,
	businessDate string,
) (FinanceDashboardReceivableMetrics, error) {
	openAmount := newMoneyAccumulator()
	overdueAmount := newMoneyAccumulator()
	outstandingAmount := newMoneyAccumulator()
	var metrics FinanceDashboardReceivableMetrics
	for _, receivable := range receivables {
		status := financedomain.NormalizeReceivableStatus(receivable.Status)
		if status == financedomain.ReceivableStatusPaid || status == financedomain.ReceivableStatusVoid {
			continue
		}
		if err := outstandingAmount.Add(receivable.OutstandingAmount); err != nil {
			return FinanceDashboardReceivableMetrics{}, err
		}
		if status == financedomain.ReceivableStatusOpen || status == financedomain.ReceivableStatusPartiallyPaid {
			metrics.OpenCount++
			if err := openAmount.Add(receivable.OutstandingAmount); err != nil {
				return FinanceDashboardReceivableMetrics{}, err
			}
		}
		if status == financedomain.ReceivableStatusDisputed {
			metrics.DisputedCount++
		}
		if isFinanceDashboardDueBefore(receivable.DueDate, businessDate) {
			metrics.OverdueCount++
			if err := overdueAmount.Add(receivable.OutstandingAmount); err != nil {
				return FinanceDashboardReceivableMetrics{}, err
			}
		}
	}
	metrics.OpenAmount = openAmount.String()
	metrics.OverdueAmount = overdueAmount.String()
	metrics.OutstandingAmount = outstandingAmount.String()

	return metrics, nil
}

func summarizeFinanceDashboardPayables(
	payables []financedomain.SupplierPayable,
	businessDate string,
) (FinanceDashboardPayableMetrics, error) {
	openAmount := newMoneyAccumulator()
	dueAmount := newMoneyAccumulator()
	outstandingAmount := newMoneyAccumulator()
	var metrics FinanceDashboardPayableMetrics
	for _, payable := range payables {
		status := financedomain.NormalizePayableStatus(payable.Status)
		if status == financedomain.PayableStatusPaid || status == financedomain.PayableStatusVoid {
			continue
		}
		if err := outstandingAmount.Add(payable.OutstandingAmount); err != nil {
			return FinanceDashboardPayableMetrics{}, err
		}
		if status == financedomain.PayableStatusPaymentRequested {
			metrics.PaymentRequestedCount++
		}
		if status == financedomain.PayableStatusPaymentApproved {
			metrics.PaymentApprovedCount++
		}
		if status == financedomain.PayableStatusOpen ||
			status == financedomain.PayableStatusPaymentRequested ||
			status == financedomain.PayableStatusPaymentApproved ||
			status == financedomain.PayableStatusPartiallyPaid {
			metrics.OpenCount++
			if err := openAmount.Add(payable.OutstandingAmount); err != nil {
				return FinanceDashboardPayableMetrics{}, err
			}
		}
		if isFinanceDashboardDueOnOrBefore(payable.DueDate, businessDate) {
			metrics.DueCount++
			if err := dueAmount.Add(payable.OutstandingAmount); err != nil {
				return FinanceDashboardPayableMetrics{}, err
			}
		}
	}
	metrics.OpenAmount = openAmount.String()
	metrics.DueAmount = dueAmount.String()
	metrics.OutstandingAmount = outstandingAmount.String()

	return metrics, nil
}

func summarizeFinanceDashboardCOD(
	remittances []financedomain.CODRemittance,
) (FinanceDashboardCODMetrics, error) {
	pendingAmount := newMoneyAccumulator()
	discrepancyAmount := newMoneyAccumulator()
	var metrics FinanceDashboardCODMetrics
	for _, remittance := range remittances {
		status := financedomain.NormalizeCODRemittanceStatus(remittance.Status)
		if status != financedomain.CODRemittanceStatusClosed && status != financedomain.CODRemittanceStatusVoid {
			metrics.PendingCount++
			if err := pendingAmount.Add(remittance.ExpectedAmount); err != nil {
				return FinanceDashboardCODMetrics{}, err
			}
		}
		if !remittance.DiscrepancyAmount.IsZero() || len(remittance.Discrepancies) > 0 {
			metrics.DiscrepancyCount++
			if err := discrepancyAmount.Add(remittance.DiscrepancyAmount); err != nil {
				return FinanceDashboardCODMetrics{}, err
			}
		}
	}
	metrics.PendingAmount = pendingAmount.String()
	metrics.DiscrepancyAmount = discrepancyAmount.String()

	return metrics, nil
}

func summarizeFinanceDashboardCash(
	transactions []financedomain.CashTransaction,
	businessDate string,
) (FinanceDashboardCashMetrics, error) {
	cashIn := newMoneyAccumulator()
	cashOut := newMoneyAccumulator()
	net := newMoneyAccumulator()
	var metrics FinanceDashboardCashMetrics
	for _, transaction := range transactions {
		if financedomain.NormalizeCashTransactionStatus(transaction.Status) != financedomain.CashTransactionStatusPosted ||
			dateStringUTC(transaction.BusinessDate) != businessDate {
			continue
		}
		metrics.TransactionCount++
		switch financedomain.NormalizeCashTransactionDirection(transaction.Direction) {
		case financedomain.CashTransactionDirectionIn:
			if err := cashIn.Add(transaction.TotalAmount); err != nil {
				return FinanceDashboardCashMetrics{}, err
			}
			if err := net.Add(transaction.TotalAmount); err != nil {
				return FinanceDashboardCashMetrics{}, err
			}
		case financedomain.CashTransactionDirectionOut:
			if err := cashOut.Add(transaction.TotalAmount); err != nil {
				return FinanceDashboardCashMetrics{}, err
			}
			if err := net.Subtract(transaction.TotalAmount); err != nil {
				return FinanceDashboardCashMetrics{}, err
			}
		}
	}
	metrics.CashInToday = cashIn.String()
	metrics.CashOutToday = cashOut.String()
	metrics.NetCashToday = net.String()

	return metrics, nil
}

func newMoneyAccumulator() moneyAccumulator {
	return moneyAccumulator{cents: big.NewInt(0)}
}

func (a *moneyAccumulator) Add(amount decimal.Decimal) error {
	cents, err := moneyDecimalCents(amount)
	if err != nil {
		return err
	}
	a.cents.Add(a.cents, cents)

	return nil
}

func (a *moneyAccumulator) Subtract(amount decimal.Decimal) error {
	cents, err := moneyDecimalCents(amount)
	if err != nil {
		return err
	}
	a.cents.Sub(a.cents, cents)

	return nil
}

func (a moneyAccumulator) String() string {
	value := new(big.Int).Set(a.cents)
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

func moneyDecimalCents(amount decimal.Decimal) (*big.Int, error) {
	normalized, err := decimal.ParseMoneyAmount(amount.String())
	if err != nil {
		return nil, err
	}
	digits := strings.ReplaceAll(normalized.String(), ".", "")
	cents, ok := new(big.Int).SetString(digits, 10)
	if !ok {
		return nil, decimal.ErrInvalidDecimal
	}

	return cents, nil
}

func normalizeFinanceDashboardBusinessDate(value string, now time.Time) (string, error) {
	value = strings.TrimSpace(value)
	if value != "" {
		if _, err := time.Parse(time.DateOnly, value); err != nil {
			return "", err
		}

		return value, nil
	}

	return financeDashboardBusinessDate(now), nil
}

func financeDashboardBusinessDate(value time.Time) string {
	loc, err := time.LoadLocation(decimal.TimezoneHoChiMinh)
	if err != nil {
		loc = time.FixedZone(decimal.TimezoneHoChiMinh, 7*60*60)
	}

	return value.In(loc).Format(time.DateOnly)
}

func isFinanceDashboardDueBefore(dueDate time.Time, businessDate string) bool {
	if dueDate.IsZero() {
		return false
	}

	return dateStringUTC(dueDate) < businessDate
}

func isFinanceDashboardDueOnOrBefore(dueDate time.Time, businessDate string) bool {
	if dueDate.IsZero() {
		return false
	}

	return dateStringUTC(dueDate) <= businessDate
}

func dateStringUTC(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.UTC().Format(time.DateOnly)
}

func financeDashboardValidationError(cause error, details map[string]any) error {
	return apperrors.BadRequest(
		ErrorCodeFinanceDashboardValidation,
		"Finance dashboard request is invalid",
		cause,
		details,
	)
}
