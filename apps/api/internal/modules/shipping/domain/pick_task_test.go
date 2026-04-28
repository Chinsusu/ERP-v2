package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewPickTaskDefaultsCreatedWithReservedLines(t *testing.T) {
	task, err := NewPickTask(validNewPickTaskInput())
	if err != nil {
		t.Fatalf("new pick task: %v", err)
	}

	if task.Status != PickTaskStatusCreated || task.AssignedTo != "" {
		t.Fatalf("task state = %s assigned %q, want created unassigned", task.Status, task.AssignedTo)
	}
	if task.PickTaskNo != "PICK-SO-260428-0001" || task.SalesOrderID != "so-260428-0001" {
		t.Fatalf("task source = %+v, want source sales order", task)
	}
	if len(task.Lines) != 1 {
		t.Fatalf("lines = %d, want 1", len(task.Lines))
	}
	line := task.Lines[0]
	if line.Status != PickTaskLineStatusPending ||
		line.StockReservationID != "rsv-so-260428-0001-line-01" ||
		line.QtyToPick != "3.000000" ||
		line.BaseUOMCode != "EA" {
		t.Fatalf("line = %+v, want pending reservation line in base UOM", line)
	}
}

func TestPickTaskAssignStartPickAndComplete(t *testing.T) {
	task, err := NewPickTask(validNewPickTaskInput())
	if err != nil {
		t.Fatalf("new pick task: %v", err)
	}
	assignedAt := time.Date(2026, 4, 28, 14, 0, 0, 0, time.UTC)
	assigned, err := task.Assign("user-picker", assignedAt)
	if err != nil {
		t.Fatalf("assign task: %v", err)
	}
	if assigned.Status != PickTaskStatusAssigned || assigned.AssignedTo != "user-picker" || !assigned.AssignedAt.Equal(assignedAt) {
		t.Fatalf("assigned task = %+v, want assigned metadata", assigned)
	}

	started, err := assigned.Start("user-picker", assignedAt.Add(5*time.Minute))
	if err != nil {
		t.Fatalf("start task: %v", err)
	}
	picked, err := started.MarkLinePicked(started.Lines[0].ID, "3", "user-picker", assignedAt.Add(10*time.Minute))
	if err != nil {
		t.Fatalf("mark line picked: %v", err)
	}
	if picked.Lines[0].Status != PickTaskLineStatusPicked || picked.Lines[0].QtyPicked != "3.000000" {
		t.Fatalf("picked line = %+v, want picked full qty", picked.Lines[0])
	}

	completed, err := picked.Complete("user-picker", assignedAt.Add(15*time.Minute))
	if err != nil {
		t.Fatalf("complete task: %v", err)
	}
	if completed.Status != PickTaskStatusCompleted || completed.CompletedBy != "user-picker" {
		t.Fatalf("completed task = %+v, want completed metadata", completed)
	}
}

func TestPickTaskRejectsDuplicateReservationAndInvalidQuantity(t *testing.T) {
	input := validNewPickTaskInput()
	input.Lines = append(input.Lines, input.Lines[0])
	input.Lines[1].ID = "pick-so-260428-0001-line-02"
	input.Lines[1].LineNo = 2
	input.Lines[1].SalesOrderLineID = "so-line-02"
	if _, err := NewPickTask(input); !errors.Is(err, ErrPickTaskDuplicateLine) {
		t.Fatalf("duplicate reservation err = %v, want duplicate line", err)
	}

	input = validNewPickTaskInput()
	input.Lines[0].QtyToPick = "0"
	if _, err := NewPickTask(input); !errors.Is(err, ErrPickTaskInvalidQuantity) {
		t.Fatalf("invalid qty err = %v, want invalid quantity", err)
	}
}

func TestPickTaskCannotCompleteUntilAllLinesPicked(t *testing.T) {
	task, err := NewPickTask(validNewPickTaskInput())
	if err != nil {
		t.Fatalf("new pick task: %v", err)
	}
	assigned, err := task.Assign("user-picker", time.Date(2026, 4, 28, 14, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("assign task: %v", err)
	}
	started, err := assigned.Start("user-picker", time.Date(2026, 4, 28, 14, 5, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("start task: %v", err)
	}

	if _, err := started.Complete("user-picker", time.Date(2026, 4, 28, 14, 10, 0, 0, time.UTC)); !errors.Is(err, ErrPickTaskInvalidTransition) {
		t.Fatalf("complete err = %v, want invalid transition until line picked", err)
	}
}

func TestPickTaskLineExceptionBlocksPickingAndComplete(t *testing.T) {
	task, err := NewPickTask(validNewPickTaskInput())
	if err != nil {
		t.Fatalf("new pick task: %v", err)
	}
	assigned, err := task.Assign("user-picker", time.Date(2026, 4, 28, 14, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("assign task: %v", err)
	}
	started, err := assigned.Start("user-picker", time.Date(2026, 4, 28, 14, 5, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("start task: %v", err)
	}

	reported, err := started.ReportLineException(
		started.Lines[0].ID,
		PickTaskLineStatusWrongBatch,
		"user-picker",
		time.Date(2026, 4, 28, 14, 6, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("report line exception: %v", err)
	}
	if reported.Status != PickTaskStatusWrongBatch ||
		reported.Lines[0].Status != PickTaskLineStatusWrongBatch ||
		reported.Lines[0].QtyPicked != "0.000000" {
		t.Fatalf("reported task = %+v, want wrong batch task and line without picked qty", reported)
	}
	if _, err := reported.MarkLinePicked(reported.Lines[0].ID, "3", "user-picker", time.Date(2026, 4, 28, 14, 7, 0, 0, time.UTC)); !errors.Is(err, ErrPickTaskInvalidTransition) {
		t.Fatalf("mark exception line picked err = %v, want invalid transition", err)
	}
	if _, err := reported.Complete("user-picker", time.Date(2026, 4, 28, 14, 8, 0, 0, time.UTC)); !errors.Is(err, ErrPickTaskInvalidTransition) {
		t.Fatalf("complete exception task err = %v, want invalid transition", err)
	}
}

func TestPickTaskLineExceptionRejectsInvalidStatus(t *testing.T) {
	task, err := NewPickTask(validNewPickTaskInput())
	if err != nil {
		t.Fatalf("new pick task: %v", err)
	}

	if _, err := task.ReportLineException(task.Lines[0].ID, PickTaskLineStatusPending, "user-picker", time.Date(2026, 4, 28, 14, 0, 0, 0, time.UTC)); !errors.Is(err, ErrPickTaskInvalidStatus) {
		t.Fatalf("report pending line exception err = %v, want invalid status", err)
	}
}

func validNewPickTaskInput() NewPickTaskInput {
	createdAt := time.Date(2026, 4, 28, 13, 0, 0, 0, time.UTC)
	return NewPickTaskInput{
		ID:            "pick-so-260428-0001",
		OrgID:         "org-my-pham",
		PickTaskNo:    "pick-so-260428-0001",
		SalesOrderID:  "so-260428-0001",
		OrderNo:       "SO-260428-0001",
		WarehouseID:   "wh-hcm-fg",
		WarehouseCode: "WH-HCM-FG",
		CreatedAt:     createdAt,
		Lines: []NewPickTaskLineInput{
			{
				ID:                 "pick-so-260428-0001-line-01",
				LineNo:             1,
				SalesOrderLineID:   "so-line-01",
				StockReservationID: "rsv-so-260428-0001-line-01",
				ItemID:             "item-serum-30ml",
				SKUCode:            "serum-30ml",
				BatchID:            "batch-serum-2604a",
				BatchNo:            "lot-2604a",
				WarehouseID:        "wh-hcm-fg",
				BinID:              "bin-hcm-pick-a01",
				BinCode:            "pick-a-01",
				QtyToPick:          "3",
				BaseUOMCode:        "EA",
			},
		},
	}
}
