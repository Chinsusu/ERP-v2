import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  confirmPackTask,
  getPackTasks,
  packTaskLineStatusTone,
  packTaskStatusTone,
  reportPackTaskException,
  resetPrototypePackTasksForTest,
  startPackTask
} from "./packTaskService";

describe("packTaskService", () => {
  beforeEach(() => {
    resetPrototypePackTasksForTest();
    vi.stubGlobal(
      "fetch",
      vi.fn(() => Promise.reject(new Error("offline")))
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("filters prototype pack tasks by warehouse and status", async () => {
    const tasks = await getPackTasks({ warehouseId: "wh-hcm-fg", status: "created" });

    expect(tasks).toHaveLength(1);
    expect(tasks[0]).toMatchObject({
      id: "pack-so-260428-0003",
      status: "created",
      warehouseCode: "WH-HCM-FG"
    });
  });

  it("starts and confirms a packing task", async () => {
    const started = await startPackTask("pack-so-260428-0003");

    expect(started.status).toBe("in_progress");

    const confirmed = await confirmPackTask("pack-so-260428-0003", [
      { lineId: "pack-so-260428-0003-line-01", packedQty: "3.000000" }
    ]);

    expect(confirmed.status).toBe("packed");
    expect(confirmed.salesOrderStatus).toBe("packed");
    expect(confirmed.lines[0]).toMatchObject({
      status: "packed",
      qtyPacked: "3.000000"
    });
  });

  it("reports a pack exception", async () => {
    const task = await reportPackTaskException(
      "pack-so-260428-0003",
      "pack_exception",
      "Packed quantity mismatch",
      "pack-so-260428-0003-line-01"
    );

    expect(task.status).toBe("pack_exception");
    expect(task.lines[0].status).toBe("pack_exception");
    expect(task.auditLogId).toBe("audit-pack-task-exception-prototype");
  });

  it("maps API pack task responses", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          success: true,
          data: [
            {
              id: "pack-api-1",
              org_id: "org-my-pham",
              pack_task_no: "PACK-API-1",
              sales_order_id: "so-api-1",
              sales_order_status: "packing",
              order_no: "SO-API-1",
              pick_task_id: "pick-api-1",
              pick_task_no: "PICK-API-1",
              warehouse_id: "wh-hcm-fg",
              warehouse_code: "WH-HCM-FG",
              status: "in_progress",
              assigned_to: "user-packer",
              lines: [
                {
                  id: "pack-api-1-line-01",
                  line_no: 1,
                  pick_task_line_id: "pick-api-1-line-01",
                  sales_order_line_id: "so-api-1-line-01",
                  item_id: "item-api",
                  sku_code: "SKU-API",
                  batch_id: "batch-api",
                  batch_no: "LOT-API",
                  warehouse_id: "wh-hcm-fg",
                  qty_to_pack: "2.000000",
                  qty_packed: "0.000000",
                  base_uom_code: "EA",
                  status: "pending",
                  created_at: "2026-04-28T10:00:00Z",
                  updated_at: "2026-04-28T10:00:00Z"
                }
              ],
              created_at: "2026-04-28T10:00:00Z",
              updated_at: "2026-04-28T10:00:00Z"
            }
          ],
          request_id: "req-pack-list"
        }),
        { status: 200 }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    const tasks = await getPackTasks({ warehouseId: "wh-hcm-fg", status: "in_progress" });

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/pack-tasks?warehouse_id=wh-hcm-fg&status=in_progress",
      {
        headers: {
          Authorization: "Bearer local-dev-access-token"
        }
      }
    );
    expect(tasks[0]).toMatchObject({
      id: "pack-api-1",
      salesOrderStatus: "packing",
      lines: [{ skuCode: "SKU-API", baseUOMCode: "EA" }]
    });
  });

  it("maps pack statuses to UI tones", () => {
    expect(packTaskStatusTone("packed")).toBe("success");
    expect(packTaskStatusTone("pack_exception")).toBe("danger");
    expect(packTaskLineStatusTone("pending")).toBe("warning");
    expect(packTaskLineStatusTone("packed")).toBe("success");
  });
});
