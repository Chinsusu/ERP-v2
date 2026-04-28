package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type WarehouseShiftStatus string
type WarehouseShiftIssueStatus string
type WarehouseShiftIssueSeverity string

const (
	WarehouseShiftStatusOpen         WarehouseShiftStatus = "open"
	WarehouseShiftStatusReconciling  WarehouseShiftStatus = "reconciling"
	WarehouseShiftStatusBlocked      WarehouseShiftStatus = "blocked"
	WarehouseShiftStatusReadyToClose WarehouseShiftStatus = "ready_to_close"
	WarehouseShiftStatusClosed       WarehouseShiftStatus = "closed"
	WarehouseShiftStatusReopened     WarehouseShiftStatus = "reopened"
)

const (
	WarehouseShiftIssueStatusOpen     WarehouseShiftIssueStatus = "open"
	WarehouseShiftIssueStatusResolved WarehouseShiftIssueStatus = "resolved"
	WarehouseShiftIssueStatusWaived   WarehouseShiftIssueStatus = "waived"
)

const (
	WarehouseShiftIssueSeverityInfo     WarehouseShiftIssueSeverity = "info"
	WarehouseShiftIssueSeverityWarning  WarehouseShiftIssueSeverity = "warning"
	WarehouseShiftIssueSeverityCritical WarehouseShiftIssueSeverity = "critical"
)

var ErrWarehouseShiftRequiredField = errors.New("warehouse shift required field is missing")
var ErrWarehouseShiftInvalidStatus = errors.New("warehouse shift status is invalid")
var ErrWarehouseShiftInvalidQuantity = errors.New("warehouse shift quantity is invalid")
var ErrWarehouseShiftPendingIssue = errors.New("warehouse shift has unresolved blocking issue")

type WarehouseShift struct {
	ID            string
	OrgID         string
	WarehouseID   string
	WarehouseCode string
	BusinessDate  string
	ShiftCode     string
	Status        WarehouseShiftStatus
	OpenedBy      string
	ClosedBy      string
	Metrics       WarehouseShiftMetrics
	PendingIssues []WarehouseShiftPendingIssue
	OpenedAt      time.Time
	UpdatedAt     time.Time
	ClosedAt      time.Time
}

type WarehouseShiftMetrics struct {
	ReceivedOrders      int
	ProcessedOrders     int
	PackedOrders        int
	HandedOverOrders    int
	ReturnOrders        int
	MovementCount       int
	InboundMovementQty  decimal.Decimal
	OutboundMovementQty decimal.Decimal
	NetMovementQty      decimal.Decimal
}

type WarehouseShiftPendingIssue struct {
	ID          string
	IssueType   string
	Severity    WarehouseShiftIssueSeverity
	Status      WarehouseShiftIssueStatus
	ReferenceID string
	Description string
	Blocking    bool
	Owner       string
	ResolvedBy  string
	ResolvedAt  time.Time
}

type NewWarehouseShiftInput struct {
	ID            string
	OrgID         string
	WarehouseID   string
	WarehouseCode string
	BusinessDate  string
	ShiftCode     string
	OpenedBy      string
	Metrics       WarehouseShiftMetrics
	PendingIssues []WarehouseShiftPendingIssue
	OpenedAt      time.Time
}

func NewWarehouseShift(input NewWarehouseShiftInput) (WarehouseShift, error) {
	openedAt := input.OpenedAt
	if openedAt.IsZero() {
		openedAt = time.Now().UTC()
	}
	metrics, err := normalizeWarehouseShiftMetrics(input.Metrics)
	if err != nil {
		return WarehouseShift{}, err
	}
	shift := WarehouseShift{
		ID:            strings.TrimSpace(input.ID),
		OrgID:         strings.TrimSpace(input.OrgID),
		WarehouseID:   strings.TrimSpace(input.WarehouseID),
		WarehouseCode: strings.ToUpper(strings.TrimSpace(input.WarehouseCode)),
		BusinessDate:  strings.TrimSpace(input.BusinessDate),
		ShiftCode:     strings.ToLower(strings.TrimSpace(input.ShiftCode)),
		Status:        WarehouseShiftStatusOpen,
		OpenedBy:      strings.TrimSpace(input.OpenedBy),
		Metrics:       metrics,
		OpenedAt:      openedAt.UTC(),
		UpdatedAt:     openedAt.UTC(),
	}
	if shift.OrgID == "" {
		shift.OrgID = "org-my-pham"
	}
	if shift.WarehouseCode == "" {
		shift.WarehouseCode = strings.ToUpper(shift.WarehouseID)
	}
	if shift.ID == "" {
		shift.ID = fmt.Sprintf("shift_%d", openedAt.UTC().UnixNano())
	}
	shift.PendingIssues = normalizeWarehouseShiftIssues(input.PendingIssues)
	if len(shift.OpenBlockingIssues()) > 0 {
		shift.Status = WarehouseShiftStatusBlocked
	}

	if err := shift.Validate(); err != nil {
		return WarehouseShift{}, err
	}

	return shift, nil
}

func NormalizeWarehouseShiftStatus(status WarehouseShiftStatus) WarehouseShiftStatus {
	switch status {
	case WarehouseShiftStatusOpen,
		WarehouseShiftStatusReconciling,
		WarehouseShiftStatusBlocked,
		WarehouseShiftStatusReadyToClose,
		WarehouseShiftStatusClosed,
		WarehouseShiftStatusReopened:
		return status
	default:
		return ""
	}
}

func (s WarehouseShift) Validate() error {
	if strings.TrimSpace(s.ID) == "" ||
		strings.TrimSpace(s.OrgID) == "" ||
		strings.TrimSpace(s.WarehouseID) == "" ||
		strings.TrimSpace(s.BusinessDate) == "" ||
		strings.TrimSpace(s.ShiftCode) == "" ||
		strings.TrimSpace(s.OpenedBy) == "" {
		return ErrWarehouseShiftRequiredField
	}
	if NormalizeWarehouseShiftStatus(s.Status) == "" {
		return ErrWarehouseShiftInvalidStatus
	}

	return nil
}

func (s WarehouseShift) EvaluatedStatus() WarehouseShiftStatus {
	if s.Status == WarehouseShiftStatusClosed || s.Status == WarehouseShiftStatusReopened {
		return s.Status
	}
	if len(s.OpenBlockingIssues()) > 0 {
		return WarehouseShiftStatusBlocked
	}
	if s.Status == WarehouseShiftStatusBlocked {
		return WarehouseShiftStatusReadyToClose
	}

	return s.Status
}

func (s WarehouseShift) CanClose() bool {
	return s.Status != WarehouseShiftStatusClosed && len(s.OpenBlockingIssues()) == 0
}

func (s WarehouseShift) Close(actorID string, closedAt time.Time) (WarehouseShift, error) {
	if s.Status == WarehouseShiftStatusClosed {
		return WarehouseShift{}, ErrWarehouseShiftInvalidStatus
	}
	if strings.TrimSpace(actorID) == "" {
		return WarehouseShift{}, ErrWarehouseShiftRequiredField
	}
	if len(s.OpenBlockingIssues()) > 0 {
		return WarehouseShift{}, ErrWarehouseShiftPendingIssue
	}
	if closedAt.IsZero() {
		closedAt = time.Now().UTC()
	}

	closed := s.Clone()
	closed.Status = WarehouseShiftStatusClosed
	closed.ClosedBy = strings.TrimSpace(actorID)
	closed.ClosedAt = closedAt.UTC()
	closed.UpdatedAt = closedAt.UTC()

	return closed, nil
}

func (s WarehouseShift) OpenBlockingIssues() []WarehouseShiftPendingIssue {
	issues := make([]WarehouseShiftPendingIssue, 0)
	for _, issue := range s.PendingIssues {
		if issue.Blocking && issue.IsOpen() {
			issues = append(issues, issue)
		}
	}

	return issues
}

func (s WarehouseShift) Clone() WarehouseShift {
	clone := s
	clone.PendingIssues = append([]WarehouseShiftPendingIssue(nil), s.PendingIssues...)
	return clone
}

func (i WarehouseShiftPendingIssue) IsOpen() bool {
	return i.Status == WarehouseShiftIssueStatusOpen || i.Status == ""
}

func SortWarehouseShifts(rows []WarehouseShift) {
	sort.SliceStable(rows, func(i int, j int) bool {
		left := rows[i]
		right := rows[j]
		if left.BusinessDate != right.BusinessDate {
			return left.BusinessDate > right.BusinessDate
		}
		if left.WarehouseCode != right.WarehouseCode {
			return left.WarehouseCode < right.WarehouseCode
		}

		return left.ShiftCode < right.ShiftCode
	})
}

func normalizeWarehouseShiftMetrics(metrics WarehouseShiftMetrics) (WarehouseShiftMetrics, error) {
	var err error
	metrics.InboundMovementQty, err = normalizeShiftQuantity(metrics.InboundMovementQty)
	if err != nil {
		return WarehouseShiftMetrics{}, err
	}
	metrics.OutboundMovementQty, err = normalizeShiftQuantity(metrics.OutboundMovementQty)
	if err != nil {
		return WarehouseShiftMetrics{}, err
	}
	metrics.NetMovementQty, err = normalizeShiftQuantity(metrics.NetMovementQty)
	if err != nil {
		return WarehouseShiftMetrics{}, err
	}

	return metrics, nil
}

func normalizeShiftQuantity(value decimal.Decimal) (decimal.Decimal, error) {
	if strings.TrimSpace(value.String()) == "" {
		return decimal.MustQuantity("0"), nil
	}
	quantity, err := decimal.ParseQuantity(value.String())
	if err != nil {
		return "", ErrWarehouseShiftInvalidQuantity
	}

	return quantity, nil
}

func normalizeWarehouseShiftIssues(inputs []WarehouseShiftPendingIssue) []WarehouseShiftPendingIssue {
	issues := make([]WarehouseShiftPendingIssue, 0, len(inputs))
	for index, input := range inputs {
		issue := WarehouseShiftPendingIssue{
			ID:          strings.TrimSpace(input.ID),
			IssueType:   strings.TrimSpace(input.IssueType),
			Severity:    normalizeWarehouseShiftIssueSeverity(input.Severity),
			Status:      normalizeWarehouseShiftIssueStatus(input.Status),
			ReferenceID: strings.TrimSpace(input.ReferenceID),
			Description: strings.TrimSpace(input.Description),
			Blocking:    input.Blocking,
			Owner:       strings.TrimSpace(input.Owner),
			ResolvedBy:  strings.TrimSpace(input.ResolvedBy),
			ResolvedAt:  input.ResolvedAt,
		}
		if issue.ID == "" {
			issue.ID = fmt.Sprintf("shift-issue-%03d", index+1)
		}
		issues = append(issues, issue)
	}

	return issues
}

func normalizeWarehouseShiftIssueStatus(status WarehouseShiftIssueStatus) WarehouseShiftIssueStatus {
	switch status {
	case WarehouseShiftIssueStatusResolved, WarehouseShiftIssueStatusWaived:
		return status
	case WarehouseShiftIssueStatusOpen, "":
		return WarehouseShiftIssueStatusOpen
	default:
		return WarehouseShiftIssueStatusOpen
	}
}

func normalizeWarehouseShiftIssueSeverity(severity WarehouseShiftIssueSeverity) WarehouseShiftIssueSeverity {
	switch severity {
	case WarehouseShiftIssueSeverityInfo, WarehouseShiftIssueSeverityWarning, WarehouseShiftIssueSeverityCritical:
		return severity
	default:
		return WarehouseShiftIssueSeverityWarning
	}
}
