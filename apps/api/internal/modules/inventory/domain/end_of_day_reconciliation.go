package domain

import (
	"errors"
	"sort"
	"strings"
	"time"
)

type EndOfDayReconciliationStatus string

const ReconciliationStatusOpen EndOfDayReconciliationStatus = "open"
const ReconciliationStatusInReview EndOfDayReconciliationStatus = "in_review"
const ReconciliationStatusClosed EndOfDayReconciliationStatus = "closed"

var ErrReconciliationAlreadyClosed = errors.New("end-of-day reconciliation is already closed")
var ErrReconciliationNeedsExceptionNote = errors.New("end-of-day reconciliation requires exception note")
var ErrReconciliationUnresolvedIssue = errors.New("end-of-day reconciliation has unresolved operational issue")

type ReconciliationChecklistItem struct {
	Key      string
	Label    string
	Complete bool
	Blocking bool
	Note     string
}

type ReconciliationLine struct {
	ID              string
	SKU             string
	BatchNo         string
	BinCode         string
	SystemQuantity  int64
	CountedQuantity int64
	Reason          string
	Owner           string
}

type EndOfDayReconciliation struct {
	ID            string
	WarehouseID   string
	WarehouseCode string
	Date          string
	ShiftCode     string
	Status        EndOfDayReconciliationStatus
	Owner         string
	Operations    ReconciliationOperations
	Checklist     []ReconciliationChecklistItem
	Lines         []ReconciliationLine
	ClosedAt      time.Time
	ClosedBy      string
}

type ReconciliationOperations struct {
	OrderCount             int
	HandoverOrderCount     int
	ReturnOrderCount       int
	StockMovementCount     int
	StockCountSessionCount int
	PendingIssueCount      int
}

type EndOfDayReconciliationSummary struct {
	SystemQuantity     int64
	CountedQuantity    int64
	VarianceQuantity   int64
	VarianceCount      int
	ChecklistTotal     int
	ChecklistCompleted int
	ReadyToClose       bool
}

type EndOfDayReconciliationFilter struct {
	WarehouseID string
	Date        string
	ShiftCode   string
	Status      EndOfDayReconciliationStatus
}

func NewEndOfDayReconciliationFilter(
	warehouseID string,
	date string,
	shiftCode string,
	status EndOfDayReconciliationStatus,
) EndOfDayReconciliationFilter {
	return EndOfDayReconciliationFilter{
		WarehouseID: strings.TrimSpace(warehouseID),
		Date:        strings.TrimSpace(date),
		ShiftCode:   strings.ToLower(strings.TrimSpace(shiftCode)),
		Status:      NormalizeReconciliationStatus(status),
	}
}

func NormalizeReconciliationStatus(status EndOfDayReconciliationStatus) EndOfDayReconciliationStatus {
	switch status {
	case ReconciliationStatusOpen, ReconciliationStatusInReview, ReconciliationStatusClosed:
		return status
	default:
		return ""
	}
}

func (r EndOfDayReconciliation) Summary(exceptionNote string) EndOfDayReconciliationSummary {
	summary := EndOfDayReconciliationSummary{
		ChecklistTotal:     len(r.Checklist),
		ChecklistCompleted: completedChecklistCount(r.Checklist),
		ReadyToClose:       r.CanClose(exceptionNote),
	}
	for _, line := range r.Lines {
		variance := line.VarianceQuantity()
		summary.SystemQuantity += line.SystemQuantity
		summary.CountedQuantity += line.CountedQuantity
		summary.VarianceQuantity += variance
		if variance != 0 {
			summary.VarianceCount++
		}
	}

	return summary
}

func (r EndOfDayReconciliation) CanClose(exceptionNote string) bool {
	if r.Status == ReconciliationStatusClosed {
		return false
	}
	if len(r.UnresolvedOperationalIssues()) > 0 {
		return false
	}
	if len(r.OpenBlockingChecklistItems()) == 0 {
		return true
	}

	return strings.TrimSpace(exceptionNote) != ""
}

func (r EndOfDayReconciliation) Close(actorID string, exceptionNote string, closedAt time.Time) (EndOfDayReconciliation, error) {
	if r.Status == ReconciliationStatusClosed {
		return EndOfDayReconciliation{}, ErrReconciliationAlreadyClosed
	}
	if len(r.UnresolvedOperationalIssues()) > 0 {
		return EndOfDayReconciliation{}, ErrReconciliationUnresolvedIssue
	}
	if !r.CanClose(exceptionNote) {
		return EndOfDayReconciliation{}, ErrReconciliationNeedsExceptionNote
	}
	if closedAt.IsZero() {
		closedAt = time.Now().UTC()
	}

	closed := r.Clone()
	closed.Status = ReconciliationStatusClosed
	closed.ClosedAt = closedAt.UTC()
	closed.ClosedBy = strings.TrimSpace(actorID)

	return closed, nil
}

func (r EndOfDayReconciliation) OpenBlockingChecklistItems() []ReconciliationChecklistItem {
	items := make([]ReconciliationChecklistItem, 0)
	for _, item := range r.Checklist {
		if item.Blocking && !item.Complete {
			items = append(items, item)
		}
	}

	return items
}

func (r EndOfDayReconciliation) UnresolvedOperationalIssues() []ReconciliationChecklistItem {
	items := make([]ReconciliationChecklistItem, 0)
	for _, item := range r.OpenBlockingChecklistItems() {
		if item.BlocksShiftClosing() {
			items = append(items, item)
		}
	}

	return items
}

func (r EndOfDayReconciliation) Clone() EndOfDayReconciliation {
	clone := r
	clone.Checklist = append([]ReconciliationChecklistItem(nil), r.Checklist...)
	clone.Lines = append([]ReconciliationLine(nil), r.Lines...)
	return clone
}

func SortEndOfDayReconciliations(rows []EndOfDayReconciliation) {
	sort.SliceStable(rows, func(i int, j int) bool {
		left := rows[i]
		right := rows[j]
		if left.Date != right.Date {
			return left.Date > right.Date
		}
		if left.WarehouseCode != right.WarehouseCode {
			return left.WarehouseCode < right.WarehouseCode
		}

		return left.ShiftCode < right.ShiftCode
	})
}

func (line ReconciliationLine) VarianceQuantity() int64 {
	return line.CountedQuantity - line.SystemQuantity
}

func (item ReconciliationChecklistItem) BlocksShiftClosing() bool {
	switch strings.ToLower(strings.TrimSpace(item.Key)) {
	case "returns",
		"return",
		"return_inspection",
		"manifest",
		"manifests",
		"carrier_manifest",
		"handover",
		"shipments",
		"adjustment",
		"adjustments",
		"stock_adjustment",
		"stock_adjustments",
		"pending_tasks",
		"pending_issues":
		return true
	default:
		return false
	}
}

func completedChecklistCount(items []ReconciliationChecklistItem) int {
	count := 0
	for _, item := range items {
		if item.Complete {
			count++
		}
	}

	return count
}
