import { describe, expect, it } from "vitest";
import {
  getWarehouseDailyBoard,
  prototypeWarehouseDailyTasks,
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

  it("maps operational status to UI tones", () => {
    expect(warehouseTaskTone("packed")).toBe("success");
    expect(warehouseTaskTone("handover")).toBe("info");
    expect(warehouseTaskTone("mismatch")).toBe("danger");
  });
});
