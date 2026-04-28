package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type PickTaskStatus string

const (
	PickTaskStatusCreated       PickTaskStatus = "created"
	PickTaskStatusAssigned      PickTaskStatus = "assigned"
	PickTaskStatusInProgress    PickTaskStatus = "in_progress"
	PickTaskStatusCompleted     PickTaskStatus = "completed"
	PickTaskStatusMissingStock  PickTaskStatus = "missing_stock"
	PickTaskStatusWrongSKU      PickTaskStatus = "wrong_sku"
	PickTaskStatusWrongBatch    PickTaskStatus = "wrong_batch"
	PickTaskStatusWrongLocation PickTaskStatus = "wrong_location"
	PickTaskStatusCancelled     PickTaskStatus = "cancelled"
)

type PickTaskLineStatus string

const (
	PickTaskLineStatusPending       PickTaskLineStatus = "pending"
	PickTaskLineStatusPicked        PickTaskLineStatus = "picked"
	PickTaskLineStatusMissingStock  PickTaskLineStatus = "missing_stock"
	PickTaskLineStatusWrongSKU      PickTaskLineStatus = "wrong_sku"
	PickTaskLineStatusWrongBatch    PickTaskLineStatus = "wrong_batch"
	PickTaskLineStatusWrongLocation PickTaskLineStatus = "wrong_location"
	PickTaskLineStatusCancelled     PickTaskLineStatus = "cancelled"
)

var ErrPickTaskRequiredField = errors.New("pick task required field is missing")
var ErrPickTaskInvalidStatus = errors.New("pick task status is invalid")
var ErrPickTaskInvalidTransition = errors.New("pick task status transition is invalid")
var ErrPickTaskInvalidQuantity = errors.New("pick task quantity is invalid")
var ErrPickTaskDuplicateLine = errors.New("pick task line is duplicated")
var ErrPickTaskActorRequired = errors.New("pick task actor is required")

type PickTask struct {
	ID            string
	OrgID         string
	PickTaskNo    string
	SalesOrderID  string
	OrderNo       string
	WarehouseID   string
	WarehouseCode string
	Status        PickTaskStatus
	AssignedTo    string
	AssignedAt    time.Time
	StartedAt     time.Time
	StartedBy     string
	CompletedAt   time.Time
	CompletedBy   string
	Lines         []PickTaskLine
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type PickTaskLine struct {
	ID                 string
	PickTaskID         string
	LineNo             int
	SalesOrderLineID   string
	StockReservationID string
	ItemID             string
	SKUCode            string
	BatchID            string
	BatchNo            string
	WarehouseID        string
	BinID              string
	BinCode            string
	QtyToPick          decimal.Decimal
	QtyPicked          decimal.Decimal
	BaseUOMCode        decimal.UOMCode
	Status             PickTaskLineStatus
	PickedAt           time.Time
	PickedBy           string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type NewPickTaskInput struct {
	ID            string
	OrgID         string
	PickTaskNo    string
	SalesOrderID  string
	OrderNo       string
	WarehouseID   string
	WarehouseCode string
	Status        PickTaskStatus
	AssignedTo    string
	AssignedAt    time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Lines         []NewPickTaskLineInput
}

type NewPickTaskLineInput struct {
	ID                 string
	LineNo             int
	SalesOrderLineID   string
	StockReservationID string
	ItemID             string
	SKUCode            string
	BatchID            string
	BatchNo            string
	WarehouseID        string
	BinID              string
	BinCode            string
	QtyToPick          string
	BaseUOMCode        string
	Status             PickTaskLineStatus
}

func NewPickTask(input NewPickTaskInput) (PickTask, error) {
	now := time.Now().UTC()
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	status := NormalizePickTaskStatus(input.Status)
	if status == "" {
		status = PickTaskStatusCreated
	}
	task := PickTask{
		ID:            strings.TrimSpace(input.ID),
		OrgID:         strings.TrimSpace(input.OrgID),
		PickTaskNo:    strings.ToUpper(strings.TrimSpace(input.PickTaskNo)),
		SalesOrderID:  strings.TrimSpace(input.SalesOrderID),
		OrderNo:       strings.ToUpper(strings.TrimSpace(input.OrderNo)),
		WarehouseID:   strings.TrimSpace(input.WarehouseID),
		WarehouseCode: strings.ToUpper(strings.TrimSpace(input.WarehouseCode)),
		Status:        status,
		AssignedTo:    strings.TrimSpace(input.AssignedTo),
		AssignedAt:    input.AssignedAt.UTC(),
		CreatedAt:     createdAt.UTC(),
		UpdatedAt:     updatedAt.UTC(),
	}
	if task.PickTaskNo == "" && task.OrderNo != "" {
		task.PickTaskNo = fmt.Sprintf("PICK-%s", task.OrderNo)
	}
	if task.ID == "" {
		task.ID = strings.ToLower(task.PickTaskNo)
	}
	if task.WarehouseCode == "" {
		task.WarehouseCode = strings.ToUpper(task.WarehouseID)
	}
	if len(input.Lines) == 0 {
		return PickTask{}, ErrPickTaskRequiredField
	}
	lines := make([]PickTaskLine, 0, len(input.Lines))
	lineNumbers := make(map[int]struct{}, len(input.Lines))
	reservations := make(map[string]struct{}, len(input.Lines))
	for _, lineInput := range input.Lines {
		line, err := newPickTaskLine(task, lineInput)
		if err != nil {
			return PickTask{}, err
		}
		if _, ok := lineNumbers[line.LineNo]; ok {
			return PickTask{}, ErrPickTaskDuplicateLine
		}
		lineNumbers[line.LineNo] = struct{}{}
		if _, ok := reservations[line.StockReservationID]; ok {
			return PickTask{}, ErrPickTaskDuplicateLine
		}
		reservations[line.StockReservationID] = struct{}{}
		lines = append(lines, line)
	}
	task.Lines = lines
	if err := task.Validate(); err != nil {
		return PickTask{}, err
	}

	return task, nil
}

func NormalizePickTaskStatus(status PickTaskStatus) PickTaskStatus {
	switch PickTaskStatus(strings.ToLower(strings.TrimSpace(string(status)))) {
	case PickTaskStatusCreated:
		return PickTaskStatusCreated
	case PickTaskStatusAssigned:
		return PickTaskStatusAssigned
	case PickTaskStatusInProgress:
		return PickTaskStatusInProgress
	case PickTaskStatusCompleted:
		return PickTaskStatusCompleted
	case PickTaskStatusMissingStock:
		return PickTaskStatusMissingStock
	case PickTaskStatusWrongSKU:
		return PickTaskStatusWrongSKU
	case PickTaskStatusWrongBatch:
		return PickTaskStatusWrongBatch
	case PickTaskStatusWrongLocation:
		return PickTaskStatusWrongLocation
	case PickTaskStatusCancelled:
		return PickTaskStatusCancelled
	default:
		return ""
	}
}

func NormalizePickTaskLineStatus(status PickTaskLineStatus) PickTaskLineStatus {
	switch PickTaskLineStatus(strings.ToLower(strings.TrimSpace(string(status)))) {
	case PickTaskLineStatusPending:
		return PickTaskLineStatusPending
	case PickTaskLineStatusPicked:
		return PickTaskLineStatusPicked
	case PickTaskLineStatusMissingStock:
		return PickTaskLineStatusMissingStock
	case PickTaskLineStatusWrongSKU:
		return PickTaskLineStatusWrongSKU
	case PickTaskLineStatusWrongBatch:
		return PickTaskLineStatusWrongBatch
	case PickTaskLineStatusWrongLocation:
		return PickTaskLineStatusWrongLocation
	case PickTaskLineStatusCancelled:
		return PickTaskLineStatusCancelled
	default:
		return ""
	}
}

func (t PickTask) Validate() error {
	if strings.TrimSpace(t.ID) == "" ||
		strings.TrimSpace(t.OrgID) == "" ||
		strings.TrimSpace(t.PickTaskNo) == "" ||
		strings.TrimSpace(t.SalesOrderID) == "" ||
		strings.TrimSpace(t.WarehouseID) == "" ||
		t.CreatedAt.IsZero() ||
		t.UpdatedAt.IsZero() ||
		len(t.Lines) == 0 {
		return ErrPickTaskRequiredField
	}
	if NormalizePickTaskStatus(t.Status) == "" {
		return ErrPickTaskInvalidStatus
	}
	if t.Status == PickTaskStatusAssigned && (strings.TrimSpace(t.AssignedTo) == "" || t.AssignedAt.IsZero()) {
		return ErrPickTaskRequiredField
	}
	if t.Status == PickTaskStatusInProgress && (strings.TrimSpace(t.AssignedTo) == "" || t.AssignedAt.IsZero() || t.StartedAt.IsZero() || strings.TrimSpace(t.StartedBy) == "") {
		return ErrPickTaskRequiredField
	}
	if t.Status == PickTaskStatusCompleted && (strings.TrimSpace(t.AssignedTo) == "" || t.AssignedAt.IsZero() || t.StartedAt.IsZero() || strings.TrimSpace(t.StartedBy) == "" || t.CompletedAt.IsZero() || strings.TrimSpace(t.CompletedBy) == "") {
		return ErrPickTaskRequiredField
	}
	for _, line := range t.Lines {
		if err := line.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (t PickTask) Assign(actorID string, assignedAt time.Time) (PickTask, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return PickTask{}, ErrPickTaskActorRequired
	}
	if t.Status != PickTaskStatusCreated {
		return PickTask{}, ErrPickTaskInvalidTransition
	}
	if assignedAt.IsZero() {
		assignedAt = time.Now().UTC()
	}
	next := t.Clone()
	next.Status = PickTaskStatusAssigned
	next.AssignedTo = actorID
	next.AssignedAt = assignedAt.UTC()
	next.UpdatedAt = assignedAt.UTC()

	return next, next.Validate()
}

func (t PickTask) Start(actorID string, startedAt time.Time) (PickTask, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return PickTask{}, ErrPickTaskActorRequired
	}
	if t.Status != PickTaskStatusAssigned {
		return PickTask{}, ErrPickTaskInvalidTransition
	}
	if startedAt.IsZero() {
		startedAt = time.Now().UTC()
	}
	next := t.Clone()
	next.Status = PickTaskStatusInProgress
	next.StartedAt = startedAt.UTC()
	next.StartedBy = actorID
	next.UpdatedAt = startedAt.UTC()

	return next, next.Validate()
}

func (t PickTask) MarkLinePicked(lineID string, pickedQty string, actorID string, pickedAt time.Time) (PickTask, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return PickTask{}, ErrPickTaskActorRequired
	}
	if t.Status != PickTaskStatusInProgress {
		return PickTask{}, ErrPickTaskInvalidTransition
	}
	qty, err := decimal.ParseQuantity(pickedQty)
	if err != nil || qty.IsNegative() || qty.IsZero() {
		return PickTask{}, ErrPickTaskInvalidQuantity
	}
	if pickedAt.IsZero() {
		pickedAt = time.Now().UTC()
	}
	next := t.Clone()
	for index, line := range next.Lines {
		if strings.TrimSpace(line.ID) != strings.TrimSpace(lineID) {
			continue
		}
		if line.Status != PickTaskLineStatusPending {
			return PickTask{}, ErrPickTaskInvalidTransition
		}
		if !quantityEqual(qty, line.QtyToPick) {
			return PickTask{}, ErrPickTaskInvalidQuantity
		}
		line.QtyPicked = qty
		line.Status = PickTaskLineStatusPicked
		line.PickedAt = pickedAt.UTC()
		line.PickedBy = actorID
		line.UpdatedAt = pickedAt.UTC()
		next.Lines[index] = line
		next.UpdatedAt = pickedAt.UTC()
		return next, next.Validate()
	}

	return PickTask{}, ErrPickTaskRequiredField
}

func (t PickTask) ReportLineException(lineID string, status PickTaskLineStatus, actorID string, reportedAt time.Time) (PickTask, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return PickTask{}, ErrPickTaskActorRequired
	}
	lineID = strings.TrimSpace(lineID)
	if lineID == "" {
		return PickTask{}, ErrPickTaskRequiredField
	}
	lineStatus := NormalizePickTaskLineStatus(status)
	taskStatus := pickTaskStatusForLineException(lineStatus)
	if taskStatus == "" {
		return PickTask{}, ErrPickTaskInvalidStatus
	}
	if t.Status == PickTaskStatusCompleted {
		return PickTask{}, ErrPickTaskInvalidTransition
	}
	if isPickTaskExceptionStatus(t.Status) && t.Status != taskStatus {
		return PickTask{}, ErrPickTaskInvalidTransition
	}
	if reportedAt.IsZero() {
		reportedAt = time.Now().UTC()
	}

	next := t.Clone()
	for index, line := range next.Lines {
		if strings.TrimSpace(line.ID) != lineID {
			continue
		}
		if line.Status == PickTaskLineStatusPicked {
			return PickTask{}, ErrPickTaskInvalidTransition
		}
		if isPickTaskLineExceptionStatus(line.Status) && line.Status != lineStatus {
			return PickTask{}, ErrPickTaskInvalidTransition
		}
		if line.Status != PickTaskLineStatusPending && line.Status != lineStatus {
			return PickTask{}, ErrPickTaskInvalidTransition
		}
		line.Status = lineStatus
		line.UpdatedAt = reportedAt.UTC()
		next.Lines[index] = line
		next.Status = taskStatus
		next.UpdatedAt = reportedAt.UTC()
		return next, next.Validate()
	}

	return PickTask{}, ErrPickTaskRequiredField
}

func (t PickTask) Complete(actorID string, completedAt time.Time) (PickTask, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return PickTask{}, ErrPickTaskActorRequired
	}
	if t.Status != PickTaskStatusInProgress {
		return PickTask{}, ErrPickTaskInvalidTransition
	}
	for _, line := range t.Lines {
		if line.Status != PickTaskLineStatusPicked {
			return PickTask{}, ErrPickTaskInvalidTransition
		}
	}
	if completedAt.IsZero() {
		completedAt = time.Now().UTC()
	}
	next := t.Clone()
	next.Status = PickTaskStatusCompleted
	next.CompletedAt = completedAt.UTC()
	next.CompletedBy = actorID
	next.UpdatedAt = completedAt.UTC()

	return next, next.Validate()
}

func (t PickTask) Clone() PickTask {
	clone := t
	clone.Lines = append([]PickTaskLine(nil), t.Lines...)
	return clone
}

func (line PickTaskLine) Validate() error {
	if strings.TrimSpace(line.ID) == "" ||
		strings.TrimSpace(line.PickTaskID) == "" ||
		line.LineNo <= 0 ||
		strings.TrimSpace(line.SalesOrderLineID) == "" ||
		strings.TrimSpace(line.StockReservationID) == "" ||
		strings.TrimSpace(line.ItemID) == "" ||
		strings.TrimSpace(line.SKUCode) == "" ||
		strings.TrimSpace(line.BatchID) == "" ||
		strings.TrimSpace(line.WarehouseID) == "" ||
		strings.TrimSpace(line.BinID) == "" ||
		line.CreatedAt.IsZero() ||
		line.UpdatedAt.IsZero() {
		return ErrPickTaskRequiredField
	}
	if _, err := decimal.NormalizeUOMCode(line.BaseUOMCode.String()); err != nil {
		return ErrPickTaskRequiredField
	}
	if NormalizePickTaskLineStatus(line.Status) == "" {
		return ErrPickTaskInvalidStatus
	}
	if line.QtyToPick.IsNegative() || line.QtyToPick.IsZero() || line.QtyPicked.IsNegative() {
		return ErrPickTaskInvalidQuantity
	}
	if !quantityAtMost(line.QtyPicked, line.QtyToPick) {
		return ErrPickTaskInvalidQuantity
	}
	if line.Status == PickTaskLineStatusPicked && (!quantityEqual(line.QtyPicked, line.QtyToPick) || line.PickedAt.IsZero() || strings.TrimSpace(line.PickedBy) == "") {
		return ErrPickTaskRequiredField
	}

	return nil
}

func SortPickTasks(tasks []PickTask) {
	sort.SliceStable(tasks, func(i int, j int) bool {
		left := tasks[i]
		right := tasks[j]
		if left.Status != right.Status {
			return left.Status < right.Status
		}
		if left.CreatedAt.Equal(right.CreatedAt) {
			return left.PickTaskNo < right.PickTaskNo
		}

		return left.CreatedAt.Before(right.CreatedAt)
	})
}

func newPickTaskLine(task PickTask, input NewPickTaskLineInput) (PickTaskLine, error) {
	status := NormalizePickTaskLineStatus(input.Status)
	if status == "" {
		status = PickTaskLineStatusPending
	}
	qtyToPick, err := decimal.ParseQuantity(input.QtyToPick)
	if err != nil || qtyToPick.IsNegative() || qtyToPick.IsZero() {
		return PickTaskLine{}, ErrPickTaskInvalidQuantity
	}
	baseUOMCode, err := decimal.NormalizeUOMCode(input.BaseUOMCode)
	if err != nil {
		return PickTaskLine{}, ErrPickTaskRequiredField
	}
	line := PickTaskLine{
		ID:                 strings.TrimSpace(input.ID),
		PickTaskID:         task.ID,
		LineNo:             input.LineNo,
		SalesOrderLineID:   strings.TrimSpace(input.SalesOrderLineID),
		StockReservationID: strings.TrimSpace(input.StockReservationID),
		ItemID:             strings.TrimSpace(input.ItemID),
		SKUCode:            strings.ToUpper(strings.TrimSpace(input.SKUCode)),
		BatchID:            strings.TrimSpace(input.BatchID),
		BatchNo:            strings.ToUpper(strings.TrimSpace(input.BatchNo)),
		WarehouseID:        firstNonBlank(input.WarehouseID, task.WarehouseID),
		BinID:              strings.TrimSpace(input.BinID),
		BinCode:            strings.ToUpper(strings.TrimSpace(input.BinCode)),
		QtyToPick:          qtyToPick,
		QtyPicked:          decimal.MustQuantity("0"),
		BaseUOMCode:        baseUOMCode,
		Status:             status,
		CreatedAt:          task.CreatedAt,
		UpdatedAt:          task.UpdatedAt,
	}
	if line.ID == "" {
		line.ID = fmt.Sprintf("%s-line-%02d", task.ID, line.LineNo)
	}
	if err := line.Validate(); err != nil {
		return PickTaskLine{}, err
	}

	return line, nil
}

func quantityAtMost(left decimal.Decimal, right decimal.Decimal) bool {
	delta, err := decimal.SubtractQuantity(right, left)
	return err == nil && !delta.IsNegative()
}

func quantityEqual(left decimal.Decimal, right decimal.Decimal) bool {
	delta, err := decimal.SubtractQuantity(left, right)
	return err == nil && delta.IsZero()
}

func pickTaskStatusForLineException(status PickTaskLineStatus) PickTaskStatus {
	switch status {
	case PickTaskLineStatusMissingStock:
		return PickTaskStatusMissingStock
	case PickTaskLineStatusWrongSKU:
		return PickTaskStatusWrongSKU
	case PickTaskLineStatusWrongBatch:
		return PickTaskStatusWrongBatch
	case PickTaskLineStatusWrongLocation:
		return PickTaskStatusWrongLocation
	case PickTaskLineStatusCancelled:
		return PickTaskStatusCancelled
	default:
		return ""
	}
}

func isPickTaskExceptionStatus(status PickTaskStatus) bool {
	switch status {
	case PickTaskStatusMissingStock,
		PickTaskStatusWrongSKU,
		PickTaskStatusWrongBatch,
		PickTaskStatusWrongLocation,
		PickTaskStatusCancelled:
		return true
	default:
		return false
	}
}

func isPickTaskLineExceptionStatus(status PickTaskLineStatus) bool {
	switch status {
	case PickTaskLineStatusMissingStock,
		PickTaskLineStatusWrongSKU,
		PickTaskLineStatusWrongBatch,
		PickTaskLineStatusWrongLocation,
		PickTaskLineStatusCancelled:
		return true
	default:
		return false
	}
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}

	return ""
}
