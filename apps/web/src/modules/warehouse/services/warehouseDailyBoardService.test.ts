import { describe, expect, it } from "vitest";
import {
  closeEndOfDayReconciliation,
  getEndOfDayReconciliations,
  getWarehouseDailyBoard,
  prototypeEndOfDayReconciliations,
  prototypeWarehouseDailyTasks,
  sortWarehouseTasksByRisk,
  summarizeReconciliationLines,
  reconciliationStatusTone,
  summarizeWarehouseDailyBoard,
  warehouseTaskTone
} from "./warehouseDailyBoardService";

describe("warehouseDailyBoardService", () => {
  it("summarizes the daily board counters", () => {
    expect(summarizeWarehouseDailyBoard(prototypeWarehouseDailyTasks)).toEqual({
      waiting: 2,
      picking: 1,
      packed: 1,
      handover: 1,
      returns: 1,
      reconciliationMismatch: 2,
      overdue: 2
    });
  });

  it("filters tasks by warehouse, date, and status", async () => {
    await expect(
      getWarehouseDailyBoard({
        warehouseId: "wh-hcm",
        date: "2026-04-26",
        status: "mismatch"
      })
    ).resolves.toMatchObject({
      warehouseCode: "HCM",
      tasks: [
        {
          reference: "VAR-260426-003",
          status: "mismatch"
        }
      ]
    });
  });

  it("prioritizes mismatch and P0 work queues before routine tasks", () => {
    expect(sortWarehouseTasksByRisk(prototypeWarehouseDailyTasks).slice(0, 2)).toMatchObject([
      {
        reference: "VAR-260426-003",
        status: "mismatch",
        priority: "P0"
      },
      {
        reference: "VAR-260426-HN-001",
        status: "mismatch",
        priority: "P0"
      }
    ]);
  });

  it("maps operational status to UI tones", () => {
    expect(warehouseTaskTone("packed")).toBe("success");
    expect(warehouseTaskTone("handover")).toBe("info");
    expect(warehouseTaskTone("mismatch")).toBe("danger");
  });

  it("filters end-of-day reconciliation sessions", async () => {
    await expect(
      getEndOfDayReconciliations({
        warehouseId: "wh-hcm",
        date: "2026-04-26",
        status: "in_review"
      })
    ).resolves.toMatchObject([
      {
        id: "rec-hcm-260426-day",
        summary: {
          systemQuantity: 164,
          countedQuantity: 162,
          varianceQuantity: -2,
          varianceCount: 1
        }
      }
    ]);
  });

  it("summarizes system quantity versus counted quantity", () => {
    expect(summarizeReconciliationLines(prototypeEndOfDayReconciliations[0].lines)).toMatchObject({
      systemQuantity: 164,
      countedQuantity: 162,
      varianceQuantity: -2,
      varianceCount: 1
    });
  });

  it("requires an exception note before closing with blocking checklist items", async () => {
    await expect(closeEndOfDayReconciliation("rec-hcm-260426-day", "")).rejects.toThrow(
      "Exception note is required before closing this shift"
    );

    await expect(closeEndOfDayReconciliation("rec-hcm-260426-day", "Variance accepted")).resolves.toMatchObject({
      status: "closed",
      auditLogId: "audit-close-rec-hcm-260426-day"
    });
  });

  it("maps reconciliation status to UI tones", () => {
    expect(reconciliationStatusTone("open")).toBe("info");
    expect(reconciliationStatusTone("in_review")).toBe("warning");
    expect(reconciliationStatusTone("closed")).toBe("success");
  });
});
