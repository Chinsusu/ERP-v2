import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  createPrototypeInventorySnapshotReport,
  downloadInventorySnapshotCSV,
  getInventorySnapshotReport,
  inventorySnapshotQueryString
} from "./inventorySnapshotReportService";

describe("inventorySnapshotReportService", () => {
  beforeEach(() => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() => Promise.reject(new Error("offline")))
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("returns filtered prototype report when API is unavailable", async () => {
    const report = await getInventorySnapshotReport({
      businessDate: "2026-04-30",
      warehouseId: "wh-hcm",
      itemId: "item-serum-30ml",
      status: "quarantine"
    });

    expect(report.metadata.filters).toMatchObject({
      businessDate: "2026-04-30",
      warehouseId: "wh-hcm",
      itemId: "item-serum-30ml",
      status: "quarantine"
    });
    expect(report.summary).toMatchObject({
      rowCount: 1,
      expiryWarningRows: 1
    });
    expect(report.rows[0]).toMatchObject({
      sku: "SERUM-30ML",
      availableQty: "110.000000",
      sourceStockState: "quarantine"
    });
  });

  it("maps API report and sends inventory report filters", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          success: true,
          data: {
            metadata: {
              generated_at: "2026-04-30T06:30:00Z",
              timezone: "Asia/Ho_Chi_Minh",
              source_version: "reporting-v1",
              filters: {
                from_date: "2026-04-30",
                to_date: "2026-04-30",
                business_date: "2026-04-30",
                warehouse_id: "wh-hcm",
                status: "quarantine",
                item_id: "item-serum-30ml"
              }
            },
            summary: {
              row_count: 1,
              low_stock_row_count: 0,
              expiry_warning_rows: 1,
              expired_rows: 0,
              totals_by_uom: [
                {
                  base_uom_code: "PCS",
                  physical_qty: "128.000000",
                  reserved_qty: "10.000000",
                  quarantine_qty: "8.000000",
                  blocked_qty: "0.000000",
                  available_qty: "110.000000"
                }
              ]
            },
            rows: [
              {
                warehouse_id: "wh-hcm",
                warehouse_code: "HCM",
                location_id: "bin-hcm-a01",
                location_code: "A-01",
                item_id: "item-serum-30ml",
                sku: "SERUM-30ML",
                batch_id: "batch-serum-2604a",
                batch_no: "LOT-2604A",
                batch_expiry: "2026-05-20",
                base_uom_code: "PCS",
                physical_qty: "128.000000",
                reserved_qty: "10.000000",
                quarantine_qty: "8.000000",
                blocked_qty: "0.000000",
                available_qty: "110.000000",
                low_stock: false,
                expiry_warning: true,
                expired: false,
                batch_qc_status: "pass",
                batch_status: "active",
                source_stock_state: "quarantine",
                source_references: [
                  {
                    entity_type: "warehouse",
                    id: "wh-hcm",
                    label: "HCM",
                    href: "/master-data?source_id=wh-hcm&source_type=warehouse",
                    unavailable: false
                  },
                  {
                    entity_type: "item",
                    id: "item-serum-30ml",
                    label: "SERUM-30ML",
                    href: "/master-data?source_id=item-serum-30ml&source_type=item",
                    unavailable: false
                  }
                ]
              }
            ]
          },
          request_id: "req-report"
        }),
        { status: 200 }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    const report = await getInventorySnapshotReport({
      businessDate: "2026-04-30",
      warehouseId: "wh-hcm",
      status: "quarantine",
      itemId: "item-serum-30ml",
      lowStockThreshold: "10",
      expiryWarningDays: "45"
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/reports/inventory-snapshot?business_date=2026-04-30&warehouse_id=wh-hcm&status=quarantine&item_id=item-serum-30ml&low_stock_threshold=10&expiry_warning_days=45",
      {
        headers: {
          Authorization: "Bearer local-dev-access-token"
        }
      }
    );
    expect(report.summary.totalsByUom[0]).toMatchObject({
      baseUomCode: "PCS",
      availableQty: "110.000000"
    });
    expect(report.rows[0]).toMatchObject({
      itemId: "item-serum-30ml",
      expiryWarning: true
    });
    expect(report.rows[0].sourceReferences[0]).toMatchObject({
      entityType: "warehouse",
      id: "wh-hcm",
      href: "/master-data?source_id=wh-hcm&source_type=warehouse",
      unavailable: false
    });
    expect(report.rows[0].sourceReferences[1]).toMatchObject({
      entityType: "item",
      id: "item-serum-30ml",
      href: "/master-data?source_id=item-serum-30ml&source_type=item",
      unavailable: false
    });
  });

  it("downloads inventory CSV with current filters", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response("warehouse_id,sku,available_qty\nwh-hcm,SERUM-30ML,110.000000\n", {
        status: 200,
        headers: {
          "Content-Disposition": `attachment; filename="inventory-snapshot-2026-04-30.csv"`,
          "Content-Type": "text/csv; charset=utf-8"
        }
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    const download = await downloadInventorySnapshotCSV({
      businessDate: "2026-04-30",
      warehouseId: "wh-hcm",
      status: "quarantine",
      itemId: "item-serum-30ml",
      lowStockThreshold: "10",
      expiryWarningDays: "45"
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/reports/inventory-snapshot/export.csv?business_date=2026-04-30&warehouse_id=wh-hcm&status=quarantine&item_id=item-serum-30ml&low_stock_threshold=10&expiry_warning_days=45",
      {
        headers: {
          Authorization: "Bearer local-dev-access-token"
        }
      }
    );
    await expect(download.blob.text()).resolves.toContain("SERUM-30ML");
    expect(download.filename).toBe("inventory-snapshot-2026-04-30.csv");
  });

  it("does not hide API permission errors behind prototype fallback", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            error: {
              code: "FORBIDDEN",
              message: "Permission denied",
              details: { permission: "reports:view" },
              request_id: "req-denied"
            }
          }),
          { status: 403 }
        )
      )
    );

    await expect(getInventorySnapshotReport()).rejects.toMatchObject({
      name: "ApiError",
      status: 403,
      code: "FORBIDDEN",
      requestId: "req-denied"
    });
  });

  it("builds stable query strings without blank filters", () => {
    expect(
      inventorySnapshotQueryString({
        fromDate: "",
        toDate: "2026-04-30",
        businessDate: "2026-04-30",
        warehouseId: "wh-hcm",
        sku: "SERUM-30ML"
      })
    ).toBe("?to_date=2026-04-30&business_date=2026-04-30&warehouse_id=wh-hcm&sku=SERUM-30ML");
  });

  it("summarizes prototype totals by UOM", () => {
    const report = createPrototypeInventorySnapshotReport({ warehouseId: "wh-hcm" });

    expect(report.summary.totalsByUom).toHaveLength(1);
    expect(report.summary.totalsByUom[0]).toMatchObject({
      baseUomCode: "PCS",
      physicalQty: "174.000000",
      availableQty: "142.000000"
    });
    expect(report.rows[0].sourceReferences.map((reference) => reference.entityType)).toEqual([
      "warehouse",
      "item",
      "inventory_batch",
      "stock_state",
      "inventory_warning"
    ]);
  });
});
