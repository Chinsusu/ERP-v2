package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type PackTaskStatus string

const (
	PackTaskStatusCreated       PackTaskStatus = "created"
	PackTaskStatusInProgress    PackTaskStatus = "in_progress"
	PackTaskStatusPacked        PackTaskStatus = "packed"
	PackTaskStatusPackException PackTaskStatus = "pack_exception"
	PackTaskStatusCancelled     PackTaskStatus = "cancelled"
)

type PackTaskLineStatus string

const (
	PackTaskLineStatusPending       PackTaskLineStatus = "pending"
	PackTaskLineStatusPacked        PackTaskLineStatus = "packed"
	PackTaskLineStatusPackException PackTaskLineStatus = "pack_exception"
	PackTaskLineStatusCancelled     PackTaskLineStatus = "cancelled"
)

var ErrPackTaskRequiredField = errors.New("pack task required field is missing")
var ErrPackTaskInvalidStatus = errors.New("pack task status is invalid")
var ErrPackTaskInvalidTransition = errors.New("pack task status transition is invalid")
var ErrPackTaskInvalidQuantity = errors.New("pack task quantity is invalid")
var ErrPackTaskDuplicateLine = errors.New("pack task line is duplicated")
var ErrPackTaskActorRequired = errors.New("pack task actor is required")

type PackTask struct {
	ID            string
	OrgID         string
	PackTaskNo    string
	SalesOrderID  string
	OrderNo       string
	PickTaskID    string
	PickTaskNo    string
	WarehouseID   string
	WarehouseCode string
	Status        PackTaskStatus
	AssignedTo    string
	AssignedAt    time.Time
	StartedAt     time.Time
	StartedBy     string
	PackedAt      time.Time
	PackedBy      string
	Lines         []PackTaskLine
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type PackTaskLine struct {
	ID               string
	PackTaskID       string
	LineNo           int
	PickTaskLineID   string
	SalesOrderLineID string
	ItemID           string
	SKUCode          string
	BatchID          string
	BatchNo          string
	WarehouseID      string
	QtyToPack        decimal.Decimal
	QtyPacked        decimal.Decimal
	BaseUOMCode      decimal.UOMCode
	Status           PackTaskLineStatus
	PackedAt         time.Time
	PackedBy         string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type NewPackTaskInput struct {
	ID            string
	OrgID         string
	PackTaskNo    string
	SalesOrderID  string
	OrderNo       string
	PickTaskID    string
	PickTaskNo    string
	WarehouseID   string
	WarehouseCode string
	Status        PackTaskStatus
	AssignedTo    string
	AssignedAt    time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Lines         []NewPackTaskLineInput
}

type NewPackTaskLineInput struct {
	ID               string
	LineNo           int
	PickTaskLineID   string
	SalesOrderLineID string
	ItemID           string
	SKUCode          string
	BatchID          string
	BatchNo          string
	WarehouseID      string
	QtyToPack        string
	BaseUOMCode      string
	Status           PackTaskLineStatus
}

func NewPackTask(input NewPackTaskInput) (PackTask, error) {
	now := time.Now().UTC()
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	status := NormalizePackTaskStatus(input.Status)
	if status == "" {
		status = PackTaskStatusCreated
	}
	task := PackTask{
		ID:            strings.TrimSpace(input.ID),
		OrgID:         strings.TrimSpace(input.OrgID),
		PackTaskNo:    strings.ToUpper(strings.TrimSpace(input.PackTaskNo)),
		SalesOrderID:  strings.TrimSpace(input.SalesOrderID),
		OrderNo:       strings.ToUpper(strings.TrimSpace(input.OrderNo)),
		PickTaskID:    strings.TrimSpace(input.PickTaskID),
		PickTaskNo:    strings.ToUpper(strings.TrimSpace(input.PickTaskNo)),
		WarehouseID:   strings.TrimSpace(input.WarehouseID),
		WarehouseCode: strings.ToUpper(strings.TrimSpace(input.WarehouseCode)),
		Status:        status,
		AssignedTo:    strings.TrimSpace(input.AssignedTo),
		AssignedAt:    input.AssignedAt.UTC(),
		CreatedAt:     createdAt.UTC(),
		UpdatedAt:     updatedAt.UTC(),
	}
	if task.PackTaskNo == "" && task.OrderNo != "" {
		task.PackTaskNo = fmt.Sprintf("PACK-%s", task.OrderNo)
	}
	if task.ID == "" {
		task.ID = strings.ToLower(task.PackTaskNo)
	}
	if task.WarehouseCode == "" {
		task.WarehouseCode = strings.ToUpper(task.WarehouseID)
	}
	if len(input.Lines) == 0 {
		return PackTask{}, ErrPackTaskRequiredField
	}

	lines := make([]PackTaskLine, 0, len(input.Lines))
	lineNumbers := make(map[int]struct{}, len(input.Lines))
	pickLines := make(map[string]struct{}, len(input.Lines))
	for _, lineInput := range input.Lines {
		line, err := newPackTaskLine(task, lineInput)
		if err != nil {
			return PackTask{}, err
		}
		if _, ok := lineNumbers[line.LineNo]; ok {
			return PackTask{}, ErrPackTaskDuplicateLine
		}
		lineNumbers[line.LineNo] = struct{}{}
		if _, ok := pickLines[line.PickTaskLineID]; ok {
			return PackTask{}, ErrPackTaskDuplicateLine
		}
		pickLines[line.PickTaskLineID] = struct{}{}
		lines = append(lines, line)
	}
	task.Lines = lines
	if err := task.Validate(); err != nil {
		return PackTask{}, err
	}

	return task, nil
}

func NormalizePackTaskStatus(status PackTaskStatus) PackTaskStatus {
	switch PackTaskStatus(strings.ToLower(strings.TrimSpace(string(status)))) {
	case PackTaskStatusCreated:
		return PackTaskStatusCreated
	case PackTaskStatusInProgress:
		return PackTaskStatusInProgress
	case PackTaskStatusPacked:
		return PackTaskStatusPacked
	case PackTaskStatusPackException:
		return PackTaskStatusPackException
	case PackTaskStatusCancelled:
		return PackTaskStatusCancelled
	default:
		return ""
	}
}

func NormalizePackTaskLineStatus(status PackTaskLineStatus) PackTaskLineStatus {
	switch PackTaskLineStatus(strings.ToLower(strings.TrimSpace(string(status)))) {
	case PackTaskLineStatusPending:
		return PackTaskLineStatusPending
	case PackTaskLineStatusPacked:
		return PackTaskLineStatusPacked
	case PackTaskLineStatusPackException:
		return PackTaskLineStatusPackException
	case PackTaskLineStatusCancelled:
		return PackTaskLineStatusCancelled
	default:
		return ""
	}
}

func (t PackTask) Validate() error {
	if strings.TrimSpace(t.ID) == "" ||
		strings.TrimSpace(t.OrgID) == "" ||
		strings.TrimSpace(t.PackTaskNo) == "" ||
		strings.TrimSpace(t.SalesOrderID) == "" ||
		strings.TrimSpace(t.PickTaskID) == "" ||
		strings.TrimSpace(t.WarehouseID) == "" ||
		t.CreatedAt.IsZero() ||
		t.UpdatedAt.IsZero() ||
		len(t.Lines) == 0 {
		return ErrPackTaskRequiredField
	}
	if NormalizePackTaskStatus(t.Status) == "" {
		return ErrPackTaskInvalidStatus
	}
	if t.Status == PackTaskStatusInProgress && (t.StartedAt.IsZero() || strings.TrimSpace(t.StartedBy) == "") {
		return ErrPackTaskRequiredField
	}
	if t.Status == PackTaskStatusPacked && (t.StartedAt.IsZero() || strings.TrimSpace(t.StartedBy) == "" || t.PackedAt.IsZero() || strings.TrimSpace(t.PackedBy) == "") {
		return ErrPackTaskRequiredField
	}
	for _, line := range t.Lines {
		if err := line.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (t PackTask) Start(actorID string, startedAt time.Time) (PackTask, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return PackTask{}, ErrPackTaskActorRequired
	}
	if t.Status != PackTaskStatusCreated {
		return PackTask{}, ErrPackTaskInvalidTransition
	}
	if startedAt.IsZero() {
		startedAt = time.Now().UTC()
	}
	next := t.Clone()
	next.Status = PackTaskStatusInProgress
	next.StartedAt = startedAt.UTC()
	next.StartedBy = actorID
	next.UpdatedAt = startedAt.UTC()

	return next, next.Validate()
}

func (t PackTask) MarkLinePacked(lineID string, packedQty string, actorID string, packedAt time.Time) (PackTask, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return PackTask{}, ErrPackTaskActorRequired
	}
	if t.Status != PackTaskStatusInProgress {
		return PackTask{}, ErrPackTaskInvalidTransition
	}
	qty, err := decimal.ParseQuantity(packedQty)
	if err != nil || qty.IsNegative() || qty.IsZero() {
		return PackTask{}, ErrPackTaskInvalidQuantity
	}
	if packedAt.IsZero() {
		packedAt = time.Now().UTC()
	}

	next := t.Clone()
	for index, line := range next.Lines {
		if strings.TrimSpace(line.ID) != strings.TrimSpace(lineID) {
			continue
		}
		if line.Status != PackTaskLineStatusPending {
			return PackTask{}, ErrPackTaskInvalidTransition
		}
		if !quantityEqual(qty, line.QtyToPack) {
			return PackTask{}, ErrPackTaskInvalidQuantity
		}
		line.QtyPacked = qty
		line.Status = PackTaskLineStatusPacked
		line.PackedAt = packedAt.UTC()
		line.PackedBy = actorID
		line.UpdatedAt = packedAt.UTC()
		next.Lines[index] = line
		next.UpdatedAt = packedAt.UTC()
		return next, next.Validate()
	}

	return PackTask{}, ErrPackTaskRequiredField
}

func (t PackTask) Complete(actorID string, packedAt time.Time) (PackTask, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return PackTask{}, ErrPackTaskActorRequired
	}
	if t.Status != PackTaskStatusInProgress {
		return PackTask{}, ErrPackTaskInvalidTransition
	}
	for _, line := range t.Lines {
		if line.Status != PackTaskLineStatusPacked {
			return PackTask{}, ErrPackTaskInvalidTransition
		}
	}
	if packedAt.IsZero() {
		packedAt = time.Now().UTC()
	}
	next := t.Clone()
	next.Status = PackTaskStatusPacked
	next.PackedAt = packedAt.UTC()
	next.PackedBy = actorID
	next.UpdatedAt = packedAt.UTC()

	return next, next.Validate()
}

func (t PackTask) Clone() PackTask {
	clone := t
	clone.Lines = append([]PackTaskLine(nil), t.Lines...)
	return clone
}

func (line PackTaskLine) Validate() error {
	if strings.TrimSpace(line.ID) == "" ||
		strings.TrimSpace(line.PackTaskID) == "" ||
		line.LineNo <= 0 ||
		strings.TrimSpace(line.PickTaskLineID) == "" ||
		strings.TrimSpace(line.SalesOrderLineID) == "" ||
		strings.TrimSpace(line.ItemID) == "" ||
		strings.TrimSpace(line.SKUCode) == "" ||
		strings.TrimSpace(line.BatchID) == "" ||
		strings.TrimSpace(line.WarehouseID) == "" ||
		line.CreatedAt.IsZero() ||
		line.UpdatedAt.IsZero() {
		return ErrPackTaskRequiredField
	}
	if _, err := decimal.NormalizeUOMCode(line.BaseUOMCode.String()); err != nil {
		return ErrPackTaskRequiredField
	}
	if NormalizePackTaskLineStatus(line.Status) == "" {
		return ErrPackTaskInvalidStatus
	}
	if line.QtyToPack.IsNegative() || line.QtyToPack.IsZero() || line.QtyPacked.IsNegative() {
		return ErrPackTaskInvalidQuantity
	}
	if !quantityAtMost(line.QtyPacked, line.QtyToPack) {
		return ErrPackTaskInvalidQuantity
	}
	if line.Status == PackTaskLineStatusPacked && (!quantityEqual(line.QtyPacked, line.QtyToPack) || line.PackedAt.IsZero() || strings.TrimSpace(line.PackedBy) == "") {
		return ErrPackTaskRequiredField
	}

	return nil
}

func SortPackTasks(tasks []PackTask) {
	sort.SliceStable(tasks, func(i int, j int) bool {
		left := tasks[i]
		right := tasks[j]
		if left.Status != right.Status {
			return left.Status < right.Status
		}
		if left.CreatedAt.Equal(right.CreatedAt) {
			return left.PackTaskNo < right.PackTaskNo
		}

		return left.CreatedAt.Before(right.CreatedAt)
	})
}

func newPackTaskLine(task PackTask, input NewPackTaskLineInput) (PackTaskLine, error) {
	status := NormalizePackTaskLineStatus(input.Status)
	if status == "" {
		status = PackTaskLineStatusPending
	}
	qtyToPack, err := decimal.ParseQuantity(input.QtyToPack)
	if err != nil || qtyToPack.IsNegative() || qtyToPack.IsZero() {
		return PackTaskLine{}, ErrPackTaskInvalidQuantity
	}
	baseUOMCode, err := decimal.NormalizeUOMCode(input.BaseUOMCode)
	if err != nil {
		return PackTaskLine{}, ErrPackTaskRequiredField
	}
	line := PackTaskLine{
		ID:               strings.TrimSpace(input.ID),
		PackTaskID:       task.ID,
		LineNo:           input.LineNo,
		PickTaskLineID:   strings.TrimSpace(input.PickTaskLineID),
		SalesOrderLineID: strings.TrimSpace(input.SalesOrderLineID),
		ItemID:           strings.TrimSpace(input.ItemID),
		SKUCode:          strings.ToUpper(strings.TrimSpace(input.SKUCode)),
		BatchID:          strings.TrimSpace(input.BatchID),
		BatchNo:          strings.ToUpper(strings.TrimSpace(input.BatchNo)),
		WarehouseID:      firstNonBlank(input.WarehouseID, task.WarehouseID),
		QtyToPack:        qtyToPack,
		QtyPacked:        decimal.MustQuantity("0"),
		BaseUOMCode:      baseUOMCode,
		Status:           status,
		CreatedAt:        task.CreatedAt,
		UpdatedAt:        task.UpdatedAt,
	}
	if line.ID == "" {
		line.ID = fmt.Sprintf("%s-line-%02d", task.ID, line.LineNo)
	}
	if err := line.Validate(); err != nil {
		return PackTaskLine{}, err
	}

	return line, nil
}
