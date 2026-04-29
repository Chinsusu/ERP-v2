import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  prototypeReturnReceipts,
  resetPrototypeReturnReceiptsForTest
} from "../../returns/services/returnReceivingService";
import { prototypeCarrierManifests } from "../../shipping/services/carrierManifestService";
import type { GoodsReceipt } from "../../receiving/types";
import {
  createGoodsReceipt,
  markGoodsReceiptInspectReady,
  postGoodsReceipt,
  resetPrototypeGoodsReceiptsForTest,
  submitGoodsReceipt
} from "../../receiving/services/warehouseReceivingService";
import { resetPrototypeStockAdjustmentsForTest } from "../../inventory/services/stockAdjustmentService";
import {
  buildWarehouseFulfillmentDrillDownHref,
  buildWarehouseInboundDrillDownHref,
  buildWarehouseQueueDrillDownHref,
  buildWarehouseShiftClosingDrillDownHref,
  closeEndOfDayReconciliation,
  composeWarehouseDailyBoard,
  getEndOfDayReconciliations,
  getWarehouseFulfillmentMetrics,
  getWarehouseInboundMetrics,
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
    resetPrototypeGoodsReceiptsForTest();
    resetPrototypeReturnReceiptsForTest();
    resetPrototypeStockAdjustmentsForTest();
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
      returnPending: 0,
      qaHold: 0,
      adjustmentPending: 0,
      reconciliationMismatch: 0,
      stockCountVariance: 0,
      closingBlocked: 0,
      closingReady: 0,
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

  it("shows S3-06-01 return, QA hold, adjustment, variance, and closing widgets", async () => {
    await expect(
      getWarehouseDailyBoard({
        warehouseId: "wh-hcm",
        date: "2026-04-26",
        shiftCode: "day"
      })
    ).resolves.toMatchObject({
      warehouseCode: "HCM",
      summary: {
        returnPending: 1,
        qaHold: 1,
        adjustmentPending: 1,
        stockCountVariance: 1,
        closingBlocked: 1
      }
    });

    await expect(
      getWarehouseDailyBoard({
        warehouseId: "wh-hcm",
        date: "2026-04-26",
        shiftCode: "day",
        status: "adjustment"
      })
    ).resolves.toMatchObject({
      tasks: [
        {
          reference: "ADJ-260426-0001",
          source: "adjustment",
          status: "adjustment"
        }
      ]
    });
  });

  it("uses service-backed return, adjustment, reconciliation, and closing data for the board", async () => {
    const board = await getWarehouseDailyBoard({
      warehouseId: "wh-hcm",
      date: "2026-04-26",
      shiftCode: "day"
    });

    expect(board.summary).toMatchObject({
      returnPending: 1,
      qaHold: 1,
      adjustmentPending: 1,
      stockCountVariance: 1,
      closingBlocked: 1
    });
    expect(board.tasks).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          reference: "RET-260425-099",
          source: "returns",
          status: "returns",
          href: "/returns?warehouse_id=wh-hcm&date=2026-04-26&queue=returns#return-receipts"
        }),
        expect.objectContaining({
          reference: "RET-260425-120",
          source: "returns",
          status: "qa_hold",
          href: "/returns?warehouse_id=wh-hcm&date=2026-04-26&queue=qa_hold#return-receipts"
        }),
        expect.objectContaining({
          reference: "ADJ-260426-0001",
          source: "adjustment",
          status: "adjustment",
          href: "/inventory?warehouse_id=wh-hcm&date=2026-04-26&queue=adjustment#stock-adjustments"
        }),
        expect.objectContaining({
          reference: "VAR-20260426-SERUM-30ML",
          source: "reconciliation",
          status: "mismatch",
          href: "/inventory?warehouse_id=wh-hcm&date=2026-04-26&queue=mismatch#stock-counts"
        }),
        expect.objectContaining({
          reference: "HCM-2026-04-26-day",
          source: "closing",
          status: "closing",
          href: "/warehouse?warehouse_id=wh-hcm&date=2026-04-26&shift_code=day#shift-closing"
        })
      ])
    );
  });

  it("uses service-backed QC-hold stock movements for variance rows", async () => {
    const receipt = await createGoodsReceipt({
      id: "grn-s3-06-03-qc-hold",
      orgId: "org-my-pham",
      receiptNo: "GRN-S30603-QCHOLD",
      warehouseId: "wh-hcm-fg",
      locationId: "loc-hcm-fg-recv-01",
      referenceDocType: "purchase_order",
      referenceDocId: "PO-S30603-QCHOLD",
      supplierId: "sup-rm-bioactive",
      deliveryNoteNo: "DN-S30603-QCHOLD",
      lines: [
        {
          id: "line-s3-06-03-qc-hold",
          purchaseOrderLineId: "po-line-s3-06-03-qc-hold",
          batchId: "batch-serum-2604a",
          lotNo: "LOT-2604A",
          expiryDate: "2027-04-01",
          quantity: "3.000000",
          uomCode: "EA",
          baseUomCode: "EA",
          packagingStatus: "intact"
        }
      ]
    });

    await submitGoodsReceipt(receipt.id);
    await markGoodsReceiptInspectReady(receipt.id);
    await postGoodsReceipt(receipt.id);

    const board = await getWarehouseDailyBoard({
      warehouseId: "wh-hcm",
      date: "2026-04-27",
      shiftCode: "day",
      status: "mismatch"
    });

    expect(board.summary.stockCountVariance).toBe(1);
    expect(board.tasks).toEqual([
      expect.objectContaining({
        reference: "GRN-S30603-QCHOLD-MV-001",
        source: "stock_movement",
        sourceField: "stock_movements.stock_status",
        status: "mismatch",
        priority: "P0"
      })
    ]);
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

  it("loads inbound metrics from the daily board API using the stock warehouse scope", async () => {
    const fetchMock = vi.fn((input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes("/warehouse/daily-board/inbound-metrics")) {
        return Promise.resolve(
          new Response(
            JSON.stringify({
              success: true,
              request_id: "req-inbound-metrics",
              data: {
                warehouse_id: "wh-hcm-fg",
                date: "2026-04-29",
                shift_code: "day",
                purchase_orders_incoming: 4,
                receiving_pending: 3,
                receiving_draft: 1,
                receiving_submitted: 1,
                receiving_inspect_ready: 1,
                qc_hold: 1,
                qc_fail: 1,
                qc_pass: 2,
                qc_partial: 1,
                supplier_rejections: 1,
                supplier_rejection_draft: 0,
                supplier_rejection_submitted: 1,
                supplier_rejection_confirmed: 0,
                supplier_rejection_cancelled: 0,
                generated_at: "2026-04-29T10:00:00Z"
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
      getWarehouseInboundMetrics({
        warehouseId: "wh-hcm",
        date: "2026-04-29",
        shiftCode: "day"
      })
    ).resolves.toMatchObject({
      warehouseId: "wh-hcm-fg",
      purchaseOrdersIncoming: 4,
      receivingPending: 3,
      qcFail: 1,
      supplierRejections: 1
    });
    expect(fetchMock.mock.calls.some(([url]) => String(url).includes("warehouse_id=wh-hcm-fg"))).toBe(true);
  });

  it("composes daily board inbound metrics returned by the API", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn((input: RequestInfo | URL) => {
        const url = String(input);
        if (url.includes("/warehouse/daily-board/inbound-metrics")) {
          return Promise.resolve(
            new Response(
              JSON.stringify({
                success: true,
                request_id: "req-board-inbound",
                data: {
                  warehouse_id: "wh-hcm-fg",
                  date: "2026-04-29",
                  shift_code: "day",
                  purchase_orders_incoming: 2,
                  receiving_pending: 1,
                  receiving_draft: 0,
                  receiving_submitted: 1,
                  receiving_inspect_ready: 0,
                  qc_hold: 0,
                  qc_fail: 1,
                  qc_pass: 3,
                  qc_partial: 0,
                  supplier_rejections: 1,
                  supplier_rejection_draft: 1,
                  supplier_rejection_submitted: 0,
                  supplier_rejection_confirmed: 0,
                  supplier_rejection_cancelled: 0,
                  generated_at: "2026-04-29T10:00:00Z"
                }
              }),
              { status: 200 }
            )
          );
        }

        return Promise.reject(new Error("offline"));
      })
    );

    await expect(
      getWarehouseDailyBoard({
        warehouseId: "wh-hcm",
        date: "2026-04-29",
        shiftCode: "day"
      })
    ).resolves.toMatchObject({
      inbound: {
        warehouseId: "wh-hcm-fg",
        purchaseOrdersIncoming: 2,
        receivingPending: 1,
        qcFail: 1,
        supplierRejections: 1
      }
    });
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

  it("builds drill-down links for fulfillment metrics", () => {
    expect(
      buildWarehouseFulfillmentDrillDownHref("waiting_handover", {
        warehouseId: "wh-hcm",
        date: "2026-04-26",
        shiftCode: "day",
        carrierCode: "ghn"
      })
    ).toBe("/warehouse?warehouse_id=wh-hcm&date=2026-04-26&shift_code=day&carrier_code=GHN&queue=handover#task-board");
    expect(buildWarehouseFulfillmentDrillDownHref("reserved", { warehouseId: "wh-hcm", date: "2026-04-26" })).toBe(
      "/sales?warehouse_id=wh-hcm-fg&date=2026-04-26&status=reserved#sales-list"
    );
    expect(buildWarehouseFulfillmentDrillDownHref("missing", { carrierCode: "VTP" })).toBe(
      "/shipping?carrier_code=VTP&status=exception#carrier-manifest-list"
    );
  });

  it("builds drill-down links for inbound metrics", () => {
    expect(
      buildWarehouseInboundDrillDownHref("purchase_orders_incoming", {
        warehouseId: "wh-hcm",
        date: "2026-04-29",
        shiftCode: "day"
      })
    ).toBe("/purchase?warehouse_id=wh-hcm-fg&date=2026-04-29&status=approved#purchase-list");
    expect(buildWarehouseInboundDrillDownHref("receiving_pending", { warehouseId: "wh-hcm" })).toBe(
      "/receiving?warehouse_id=wh-hcm-fg&status=pending#receiving-list"
    );
    expect(buildWarehouseInboundDrillDownHref("qc_fail", { warehouseId: "wh-hcm", date: "2026-04-29" })).toBe(
      "/qc?warehouse_id=wh-hcm-fg&date=2026-04-29&result=fail#qc-inspections"
    );
    expect(buildWarehouseInboundDrillDownHref("supplier_rejections", { warehouseId: "wh-hcm" })).toBe(
      "/returns?warehouse_id=wh-hcm-fg&panel=supplier_rejections#supplier-rejections"
    );
  });

  it("builds drill-down links for daily board queue alerts", () => {
    expect(
      buildWarehouseQueueDrillDownHref("returns", {
        warehouseId: "wh-hcm",
        date: "2026-04-26",
        shiftCode: "day"
      })
    ).toBe("/returns?warehouse_id=wh-hcm&date=2026-04-26&queue=returns#return-receipts");
    expect(
      buildWarehouseQueueDrillDownHref("qa_hold", {
        warehouseId: "wh-hcm",
        date: "2026-04-26",
        shiftCode: "day"
      })
    ).toBe("/returns?warehouse_id=wh-hcm&date=2026-04-26&queue=qa_hold#return-receipts");
    expect(
      buildWarehouseQueueDrillDownHref("overdue", {
        warehouseId: "wh-hcm",
        date: "2026-04-26",
        shiftCode: "day",
        carrierCode: "ghn"
      })
    ).toBe("/warehouse?warehouse_id=wh-hcm&date=2026-04-26&shift_code=day&carrier_code=GHN&queue=overdue#task-board");
    expect(buildWarehouseQueueDrillDownHref("adjustment", { warehouseId: "wh-hcm", date: "2026-04-26" })).toBe(
      "/inventory?warehouse_id=wh-hcm&date=2026-04-26&queue=adjustment#stock-adjustments"
    );
    expect(buildWarehouseQueueDrillDownHref("mismatch", { warehouseId: "wh-hcm", date: "2026-04-26" })).toBe(
      "/inventory?warehouse_id=wh-hcm&date=2026-04-26&queue=mismatch#stock-counts"
    );
    expect(
      buildWarehouseShiftClosingDrillDownHref({
        warehouseId: "wh-hcm",
        date: "2026-04-26",
        shiftCode: "day"
      })
    ).toBe("/warehouse?warehouse_id=wh-hcm&date=2026-04-26&shift_code=day#shift-closing");
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
    expect(board.summary.qaHold).toBe(1);
    expect(board.summary.closingBlocked).toBe(1);
    expect(board.summary.overdue).toBe(3);
    expect(board.sourceFields.find((source) => source.counter === "reconciliationMismatch")?.fields).toContain(
      "reconciliation_lines.variance_quantity"
    );
    expect(board.tasks.find((task) => task.status === "returns")?.href).toContain("#return-receipts");
    expect(board.tasks.find((task) => task.status === "closing")?.href).toContain("#shift-closing");
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
      returnPending: 1,
      qaHold: 0,
      adjustmentPending: 0,
      reconciliationMismatch: 2,
      stockCountVariance: 2,
      closingBlocked: 1,
      closingReady: 0,
      overdue: 3
    });
    expect(board.tasks.map((task) => task.source)).toEqual([
      "stock_movement",
      "reconciliation",
      "closing",
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
        operations: {
          orderCount: 42,
          handoverOrderCount: 27,
          returnOrderCount: 3,
          pendingIssueCount: 2
        },
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

  it("blocks closing when operational checklist items are unresolved", async () => {
    await expect(closeEndOfDayReconciliation("rec-hcm-260426-day", "")).rejects.toThrow(
      "Resolve return, manifest, adjustment, or pending issue before closing this shift"
    );
    await expect(closeEndOfDayReconciliation("rec-hcm-260426-day", "Variance accepted")).rejects.toThrow(
      "Resolve return, manifest, adjustment, or pending issue before closing this shift"
    );
  });

  it("closes a clean end-of-day reconciliation session", async () => {
    await expect(closeEndOfDayReconciliation("rec-hn-260426-day", "")).resolves.toMatchObject({
      status: "closed",
      auditLogId: "audit-close-rec-hn-260426-day",
      operations: {
        pendingIssueCount: 0
      }
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
        purchaseOrderLineId: "po-line-test-draft",
        itemId: "item-serum-30ml",
        sku: "SERUM-30ML",
        batchId: "batch-serum-2604a",
        batchNo: "LOT-2604A",
        lotNo: "LOT-2604A",
        expiryDate: "2027-04-01",
        warehouseId: "wh-hcm-fg",
        locationId: "loc-hcm-fg-recv-01",
        quantity: "6.000000",
        uomCode: "EA",
        baseUomCode: "EA",
        packagingStatus: "intact"
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
