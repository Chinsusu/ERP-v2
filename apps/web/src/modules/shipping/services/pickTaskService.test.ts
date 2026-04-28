import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  completePickTask,
  confirmPickTaskLine,
  getPickTasks,
  pickTaskLineStatusTone,
  pickTaskStatusTone,
  reportPickTaskException,
  resetPrototypePickTasksForTest,
  startPickTask
} from "./pickTaskService";

describe("pickTaskService", () => {
  beforeEach(() => {
    resetPrototypePickTasksForTest();
    vi.stubGlobal(
      "fetch",
      vi.fn(() => Promise.reject(new Error("offline")))
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("filters prototype pick tasks by warehouse and status", async () => {
    const tasks = await getPickTasks({ warehouseId: "wh-hcm-fg", status: "created" });

    expect(tasks).toHaveLength(1);
    expect(tasks[0]).toMatchObject({
      id: "pick-so-260428-0001",
      status: "created",
      warehouseCode: "WH-HCM-FG"
    });
  });

  it("starts confirms and completes a scan-first pick task", async () => {
    const started = await startPickTask("pick-so-260428-0001");

    expect(started).toMatchObject({
      status: "in_progress",
      assignedTo: "user-picker"
    });

    const confirmed = await confirmPickTaskLine(
      "pick-so-260428-0001",
      "pick-so-260428-0001-line-01",
      "3.000000"
    );

    expect(confirmed.lines[0]).toMatchObject({
      status: "picked",
      qtyPicked: "3.000000"
    });

    const completed = await completePickTask("pick-so-260428-0001");

    expect(completed.status).toBe("completed");
  });

  it("reports a pick exception", async () => {
    const task = await reportPickTaskException("pick-so-260428-0001", "wrong_batch", "Scanned batch mismatch");

    expect(task.status).toBe("wrong_batch");
    expect(task.auditLogId).toBe("audit-pick-task-exception-prototype");
  });

  it("maps API pick task responses", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          success: true,
          data: [
            {
              id: "pick-api-1",
              org_id: "org-my-pham",
              pick_task_no: "PICK-API-1",
              sales_order_id: "so-api-1",
              order_no: "SO-API-1",
              warehouse_id: "wh-hcm-fg",
              warehouse_code: "WH-HCM-FG",
              status: "in_progress",
              assigned_to: "user-picker",
              lines: [
                {
                  id: "pick-api-1-line-01",
                  line_no: 1,
                  sales_order_line_id: "so-api-1-line-01",
                  stock_reservation_id: "rsv-api-1-line-01",
                  item_id: "item-api",
                  sku_code: "SKU-API",
                  batch_id: "batch-api",
                  batch_no: "LOT-API",
                  warehouse_id: "wh-hcm-fg",
                  bin_id: "bin-api",
                  bin_code: "PICK-API",
                  qty_to_pick: "2.000000",
                  qty_picked: "0.000000",
                  base_uom_code: "EA",
                  status: "pending",
                  created_at: "2026-04-28T09:00:00Z",
                  updated_at: "2026-04-28T09:00:00Z"
                }
              ],
              created_at: "2026-04-28T09:00:00Z",
              updated_at: "2026-04-28T09:00:00Z"
            }
          ],
          request_id: "req-pick-list"
        }),
        { status: 200 }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    const tasks = await getPickTasks({ warehouseId: "wh-hcm-fg", status: "in_progress" });

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/pick-tasks?warehouse_id=wh-hcm-fg&status=in_progress",
      {
        headers: {
          Authorization: "Bearer local-dev-access-token"
        }
      }
    );
    expect(tasks[0]).toMatchObject({
      id: "pick-api-1",
      assignedTo: "user-picker",
      lines: [{ skuCode: "SKU-API", baseUOMCode: "EA" }]
    });
  });

  it("maps pick task and line statuses to UI tones", () => {
    expect(pickTaskStatusTone("completed")).toBe("success");
    expect(pickTaskStatusTone("wrong_location")).toBe("danger");
    expect(pickTaskLineStatusTone("pending")).toBe("warning");
    expect(pickTaskLineStatusTone("picked")).toBe("success");
  });
});
