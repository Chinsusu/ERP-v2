import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { prototypeReturnReceipts } from "../../returns/services/returnReceivingService";
import { prototypeCarrierManifests } from "../../shipping/services/carrierManifestService";
import type { GoodsReceipt } from "../../receiving/types";
import {
  closeEndOfDayReconciliation,
  composeWarehouseDailyBoard,
  getEndOfDayReconciliations,
  getWarehouseFulfillmentMetrics,
  getWarehouseDailyBoard,
  prototypeEndOfDayReconciliations,
  prototypeWarehouseDailyTasks,
  sortWarehouseTasksByRisk,
  summarizeWarehouseFulfillmentMetrics,
  summarizeReconciliationLines,
  reconciliationStatusTone,
  summarizeWarehouseDailyBoard,
  warehouseTaskTone
} from "./warehouseDailyBoardService";

describe("warehouseDailyBoardService", () => {
  beforeEach(() => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() => Promise.reject(new Error("offline")))
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("summarizes the daily board counters", () => {
    expect(summarizeWarehouseDailyBoard(prototypeWarehouseDailyTasks)).toEqual({
      waiting: 2,
      picking: 1,
      packed: 1,
      handover: 0,
      returns: 0,
      reconciliationMismatch: 0,
      overdue: 0
    });
  });

  it("returns an empty board when no source has work for the selected shift", () => {
    const board = composeWarehouseDailyBoard(
      { date: "2026-04-26", shiftCode: "night" },
      {
        orderTasks: prototypeWarehouseDailyTasks,
        carrierManifests: prototypeCarrierManifests,
        returnReceipts: prototypeReturnReceipts,
        reconciliations: prototypeEndOfDayReconciliations
      }
    );

    expect(board).toMatchObject({
      shiftCode: "night",
      shiftStatus: "closing",
      summary: {
        waiting: 0,
        picking: 0,
        packed: 0,
        handover: 0,
        returns: 0,
        reconciliationMismatch: 0,
        overdue: 0
      },
      tasks: []
    });
  });

  it("filters integrated tasks by warehouse, date, shift, and status", async () => {
    await expect(
      getWarehouseDailyBoard({
        warehouseId: "wh-hcm",
        date: "2026-04-26",
        shiftCode: "day",
        status: "mismatch"
      })
    ).resolves.toMatchObject({
      warehouseCode: "HCM",
      tasks: [
        {
          reference: "VAR-20260426-SERUM-30ML",
          status: "mismatch",
          source: "reconciliation"
        }
      ]
    });
  });

  it("loads fulfillment metrics from the daily board API with the selected carrier", async () => {
    const fetchMock = vi.fn((input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes("/warehouse/daily-board/fulfillment-metrics")) {
        return Promise.resolve(
          new Response(
            JSON.stringify({
              success: true,
              request_id: "req-fulfillment-metrics",
              data: {
                warehouse_id: "wh-hcm",
                date: "2026-04-26",
                shift_code: "day",
                carrier_code: "GHN",
                total_orders: 8,
                new_orders: 1,
                reserved_orders: 2,
                picking_orders: 1,
                packed_orders: 1,
                waiting_handover_orders: 2,
                missing_orders: 1,
                handover_orders: 0,
                generated_at: "2026-04-26T10:00:00Z"
              }
            }),
            { status: 200 }
          )
        );
      }

      return Promise.reject(new Error("offline"));
    });
    vi.stubGlobal("fetch", fetchMock);

    await expect(
      getWarehouseFulfillmentMetrics({
        warehouseId: "wh-hcm",
        date: "2026-04-26",
        shiftCode: "day",
        carrierCode: "ghn"
      })
    ).resolves.toMatchObject({
      carrierCode: "GHN",
      totalOrders: 8,
      reservedOrders: 2,
      missingOrders: 1
    });
    expect(fetchMock.mock.calls.some(([url]) => String(url).includes("carrier_code=GHN"))).toBe(true);
  });

  it("keeps the board carrier filter aligned with fulfillment and handover tasks", () => {
    const board = composeWarehouseDailyBoard(
      { warehouseId: "wh-hcm", date: "2026-04-26", shiftCode: "day", carrierCode: "GHN" },
      {
        carrierManifests: prototypeCarrierManifests,
        fulfillmentMetrics: {
          warehouseId: "wh-hcm",
          date: "2026-04-26",
          shiftCode: "day",
          carrierCode: "GHN",
          totalOrders: 3,
          newOrders: 0,
          reservedOrders: 0,
          pickingOrders: 0,
          packedOrders: 0,
          waitingHandoverOrders: 3,
          missingOrders: 1,
          handoverOrders: 0,
          generatedAt: "2026-04-26T10:00:00Z"
        }
      }
    );

    expect(board.fulfillment).toMatchObject({
      carrierCode: "GHN",
      totalOrders: 3,
      missingOrders: 1
    });
    expect(board.tasks.length).toBeGreaterThan(0);
    expect(board.tasks.every((task) => task.carrierCode === "GHN")).toBe(true);
  });

  it("derives fallback fulfillment metrics from available board tasks", () => {
    expect(
      summarizeWarehouseFulfillmentMetrics({ date: "2026-04-26", shiftCode: "day" }, prototypeWarehouseDailyTasks)
    ).toMatchObject({
      totalOrders: 4,
      newOrders: 2,
      reservedOrders: 0,
      pickingOrders: 1,
      packedOrders: 1,
      generatedAt: "2026-04-26T00:00:00Z"
    });
  });

  it("keeps active shift open when a P0 exception exists", () => {
    const board = composeWarehouseDailyBoard(
      { warehouseId: "wh-hcm", date: "2026-04-26", shiftCode: "day" },
      {
        orderTasks: prototypeWarehouseDailyTasks,
        carrierManifests: prototypeCarrierManifests,
        returnReceipts: prototypeReturnReceipts,
        reconciliations: prototypeEndOfDayReconciliations
      }
    );

    expect(board.shiftStatus).toBe("open");
    expect(board.summary.reconciliationMismatch).toBe(1);
    expect(board.summary.overdue).toBe(1);
    expect(board.sourceFields.find((source) => source.counter === "reconciliationMismatch")?.fields).toContain(
      "reconciliation_lines.variance_quantity"
    );
  });

  it("composes mixed receiving, stock movement, shipping, return, and exception workloads", () => {
    const board = composeWarehouseDailyBoard(
      { warehouseId: "wh-hcm", date: "2026-04-27", shiftCode: "day" },
      {
        orderTasks: [],
        goodsReceipts: [createDraftReceipt(), createPostedQCHoldReceipt()],
        carrierManifests: [
          {
            ...prototypeCarrierManifests[0],
            date: "2026-04-27",
            createdAt: "2026-04-27T08:00:00Z"
          }
        ],
        returnReceipts: [
          {
            ...prototypeReturnReceipts[0],
            receivedAt: "2026-04-27T09:00:00Z",
            createdAt: "2026-04-27T09:00:00Z"
          }
        ],
        reconciliations: [
          {
            ...prototypeEndOfDayReconciliations[0],
            date: "2026-04-27"
          }
        ]
      }
    );

    expect(board.summary).toEqual({
      waiting: 1,
      picking: 0,
      packed: 0,
      handover: 1,
      returns: 1,
      reconciliationMismatch: 2,
      overdue: 2
    });
    expect(board.tasks.map((task) => task.source)).toEqual([
      "stock_movement",
      "reconciliation",
      "shipping",
      "returns",
      "receiving"
    ]);
  });

  it("prioritizes mismatch and P0 work queues before routine tasks", () => {
    const board = composeWarehouseDailyBoard(
      { date: "2026-04-26", shiftCode: "day" },
      {
        orderTasks: prototypeWarehouseDailyTasks,
        carrierManifests: prototypeCarrierManifests,
        returnReceipts: prototypeReturnReceipts,
        reconciliations: prototypeEndOfDayReconciliations
      }
    );

    expect(sortWarehouseTasksByRisk(board.tasks).slice(0, 1)).toMatchObject([
      {
        reference: "VAR-20260426-SERUM-30ML",
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
        shiftCode: "day",
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

function createDraftReceipt(): GoodsReceipt {
  return {
    id: "grn-test-draft",
    orgId: "org-my-pham",
    receiptNo: "GRN-260427-TEST-001",
    warehouseId: "wh-hcm-fg",
    warehouseCode: "WH-HCM-FG",
    locationId: "loc-hcm-fg-recv-01",
    locationCode: "FG-RECV-01",
    referenceDocType: "purchase_order",
    referenceDocId: "PO-260427-TEST-001",
    status: "submitted",
    lines: [
      {
        id: "line-test-draft",
        itemId: "item-serum-30ml",
        sku: "SERUM-30ML",
        warehouseId: "wh-hcm-fg",
        locationId: "loc-hcm-fg-recv-01",
        quantity: "6.000000",
        baseUomCode: "EA"
      }
    ],
    createdBy: "user-warehouse-lead",
    createdAt: "2026-04-27T08:00:00Z",
    updatedAt: "2026-04-27T08:30:00Z",
    submittedAt: "2026-04-27T08:30:00Z"
  };
}

function createPostedQCHoldReceipt(): GoodsReceipt {
  return {
    ...createDraftReceipt(),
    id: "grn-test-posted",
    receiptNo: "GRN-260427-TEST-002",
    referenceDocId: "PO-260427-TEST-002",
    status: "posted",
    updatedAt: "2026-04-27T08:45:00Z",
    postedAt: "2026-04-27T08:45:00Z",
    postedBy: "user-warehouse-lead",
    stockMovements: [
      {
        movementNo: "MV-260427-TEST-001",
        movementType: "purchase_receipt",
        itemId: "item-serum-30ml",
        batchId: "batch-serum-2604a",
        warehouseId: "wh-hcm-fg",
        locationId: "loc-hcm-fg-recv-01",
        quantity: "6.000000",
        baseUomCode: "EA",
        stockStatus: "qc_hold",
        sourceDocId: "grn-test-posted",
        sourceDocLineId: "line-test-draft"
      }
    ]
  };
}
