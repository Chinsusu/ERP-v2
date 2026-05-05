import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  createStockTransfer,
  createWarehouseIssue,
  getStockTransfers,
  getWarehouseIssues,
  resetPrototypeWarehouseDocumentsForTest,
  transitionStockTransfer,
  transitionWarehouseIssue
} from "./warehouseDocumentService";

describe("warehouseDocumentService", () => {
  beforeEach(() => {
    resetPrototypeWarehouseDocumentsForTest();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("creates and posts a prototype stock transfer when the API is not reachable", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("offline")));

    const created = await createStockTransfer({
      sourceWarehouseId: "wh-hcm",
      sourceWarehouseCode: "HCM",
      destinationWarehouseId: "wh-hn",
      destinationWarehouseCode: "HN",
      reasonCode: "replenishment",
      lines: [
        {
          sku: "SERUM-30ML",
          batchId: "batch-serum-2604a",
          batchNo: "LOT-2604A",
          sourceLocationId: "bin-hcm-a01",
          sourceLocationCode: "A-01",
          quantity: "2.000000",
          baseUomCode: "PCS"
        }
      ]
    });
    await transitionStockTransfer(created.id, "submit");
    await transitionStockTransfer(created.id, "approve");
    const posted = await transitionStockTransfer(created.id, "post");

    expect(posted).toMatchObject({
      status: "posted",
      postedBy: "local-dev"
    });
    await expect(getStockTransfers()).resolves.toEqual(expect.arrayContaining([posted]));
  });

  it("creates and posts a prototype warehouse issue when the API is not reachable", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("offline")));

    const created = await createWarehouseIssue({
      warehouseId: "wh-hcm",
      warehouseCode: "HCM",
      destinationType: "factory",
      destinationName: "Factory A",
      reasonCode: "production_plan_issue",
      lines: [
        {
          sku: "ACI_BHA",
          itemName: "ACID SALICYLIC",
          batchNo: "RM-2605-A",
          locationId: "bin-rm-a01",
          locationCode: "RM-A-01",
          quantity: "0.125000",
          baseUomCode: "KG",
          sourceDocumentType: "production_plan",
          sourceDocumentId: "PP-260505-000001",
          sourceDocumentLineId: "pp-line-001"
        }
      ]
    });
    await transitionWarehouseIssue(created.id, "submit");
    await transitionWarehouseIssue(created.id, "approve");
    const posted = await transitionWarehouseIssue(created.id, "post");

    expect(posted).toMatchObject({
      status: "posted",
      postedBy: "local-dev",
      lines: [expect.objectContaining({ sourceDocumentLineId: "pp-line-001" })]
    });
    await expect(getWarehouseIssues()).resolves.toEqual(expect.arrayContaining([posted]));
  });
});
