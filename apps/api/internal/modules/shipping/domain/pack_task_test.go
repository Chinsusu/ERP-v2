package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewPackTaskDefaultsCreatedWithPickedLines(t *testing.T) {
	task, err := NewPackTask(validNewPackTaskInput())
	if err != nil {
		t.Fatalf("new pack task: %v", err)
	}

	if task.Status != PackTaskStatusCreated || task.PackTaskNo != "PACK-SO-260428-0001" {
		t.Fatalf("task state = %s no %s, want created pack task", task.Status, task.PackTaskNo)
	}
	if task.SalesOrderID != "so-260428-0001" || task.PickTaskID != "pick-so-260428-0001" {
		t.Fatalf("task source = %+v, want linked sales order and pick task", task)
	}
	if len(task.Lines) != 1 {
		t.Fatalf("lines = %d, want 1", len(task.Lines))
	}
	line := task.Lines[0]
	if line.Status != PackTaskLineStatusPending ||
		line.PickTaskLineID != "pick-so-260428-0001-line-01" ||
		line.QtyToPack != "3.000000" ||
		line.BaseUOMCode != "EA" {
		t.Fatalf("line = %+v, want pending picked line in base UOM", line)
	}
}

func TestPackTaskStartPackAndComplete(t *testing.T) {
	task, err := NewPackTask(validNewPackTaskInput())
	if err != nil {
		t.Fatalf("new pack task: %v", err)
	}
	startedAt := time.Date(2026, 4, 28, 15, 0, 0, 0, time.UTC)
	started, err := task.Start("user-packer", startedAt)
	if err != nil {
		t.Fatalf("start task: %v", err)
	}
	if started.Status != PackTaskStatusInProgress || started.StartedBy != "user-packer" || !started.StartedAt.Equal(startedAt) {
		t.Fatalf("started task = %+v, want in-progress metadata", started)
	}

	packed, err := started.MarkLinePacked(started.Lines[0].ID, "3", "user-packer", startedAt.Add(5*time.Minute))
	if err != nil {
		t.Fatalf("mark line packed: %v", err)
	}
	if packed.Lines[0].Status != PackTaskLineStatusPacked || packed.Lines[0].QtyPacked != "3.000000" {
		t.Fatalf("packed line = %+v, want packed full qty", packed.Lines[0])
	}

	completed, err := packed.Complete("user-packer", startedAt.Add(10*time.Minute))
	if err != nil {
		t.Fatalf("complete task: %v", err)
	}
	if completed.Status != PackTaskStatusPacked || completed.PackedBy != "user-packer" {
		t.Fatalf("completed task = %+v, want packed metadata", completed)
	}
}

func TestPackTaskRejectsDuplicatePickLineAndInvalidQuantity(t *testing.T) {
	input := validNewPackTaskInput()
	input.Lines = append(input.Lines, input.Lines[0])
	input.Lines[1].ID = "pack-so-260428-0001-line-02"
	input.Lines[1].LineNo = 2
	input.Lines[1].SalesOrderLineID = "so-line-02"
	if _, err := NewPackTask(input); !errors.Is(err, ErrPackTaskDuplicateLine) {
		t.Fatalf("duplicate pick line err = %v, want duplicate line", err)
	}

	input = validNewPackTaskInput()
	input.Lines[0].QtyToPack = "0"
	if _, err := NewPackTask(input); !errors.Is(err, ErrPackTaskInvalidQuantity) {
		t.Fatalf("invalid qty err = %v, want invalid quantity", err)
	}
}

func TestPackTaskCannotCompleteUntilAllLinesPacked(t *testing.T) {
	task, err := NewPackTask(validNewPackTaskInput())
	if err != nil {
		t.Fatalf("new pack task: %v", err)
	}
	started, err := task.Start("user-packer", time.Date(2026, 4, 28, 15, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("start task: %v", err)
	}

	if _, err := started.Complete("user-packer", time.Date(2026, 4, 28, 15, 10, 0, 0, time.UTC)); !errors.Is(err, ErrPackTaskInvalidTransition) {
		t.Fatalf("complete err = %v, want invalid transition until line packed", err)
	}
}

func validNewPackTaskInput() NewPackTaskInput {
	createdAt := time.Date(2026, 4, 28, 14, 45, 0, 0, time.UTC)
	return NewPackTaskInput{
		ID:            "pack-so-260428-0001",
		OrgID:         "org-my-pham",
		PackTaskNo:    "pack-so-260428-0001",
		SalesOrderID:  "so-260428-0001",
		OrderNo:       "SO-260428-0001",
		PickTaskID:    "pick-so-260428-0001",
		PickTaskNo:    "PICK-SO-260428-0001",
		WarehouseID:   "wh-hcm-fg",
		WarehouseCode: "WH-HCM-FG",
		CreatedAt:     createdAt,
		Lines: []NewPackTaskLineInput{
			{
				ID:               "pack-so-260428-0001-line-01",
				LineNo:           1,
				PickTaskLineID:   "pick-so-260428-0001-line-01",
				SalesOrderLineID: "so-line-01",
				ItemID:           "item-serum-30ml",
				SKUCode:          "serum-30ml",
				BatchID:          "batch-serum-2604a",
				BatchNo:          "lot-2604a",
				WarehouseID:      "wh-hcm-fg",
				QtyToPack:        "3",
				BaseUOMCode:      "EA",
			},
		},
	}
}
