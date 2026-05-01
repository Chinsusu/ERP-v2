import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  createPrototypeOperationsDailyReport,
  downloadOperationsDailyCSV,
  getOperationsDailyReport,
  operationsDailyCSVFilename,
  operationsDailyQueryString
} from "./operationsDailyReportService";

describe("operationsDailyReportService", () => {
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
    const report = await getOperationsDailyReport({
      businessDate: "2026-04-30",
      warehouseId: "wh-hcm",
      status: "blocked"
    });

    expect(report.metadata.filters).toMatchObject({
      businessDate: "2026-04-30",
      warehouseId: "wh-hcm",
      status: "blocked"
    });
    expect(report.summary).toMatchObject({
      signalCount: 2,
      blockedCount: 2
    });
    expect(report.areas.map((area) => area.area)).toEqual(["outbound", "stock_count"]);
    expect(report.rows.map((row) => row.exceptionCode)).toEqual(["MISSING_HANDOVER_SCAN", "VARIANCE_REVIEW"]);
  });

  it("maps API report and sends operations report filters", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          success: true,
          data: {
            metadata: {
              generated_at: "2026-04-30T07:00:52Z",
              timezone: "Asia/Ho_Chi_Minh",
              source_version: "reporting-v1",
              filters: {
                from_date: "2026-04-30",
                to_date: "2026-04-30",
                business_date: "2026-04-30",
                warehouse_id: "wh-hcm",
                status: "blocked"
              }
            },
            summary: {
              signal_count: 2,
              pending_count: 0,
              in_progress_count: 0,
              completed_count: 0,
              blocked_count: 2,
              exception_count: 0
            },
            areas: [
              {
                area: "outbound",
                signal_count: 1,
                pending_count: 0,
                in_progress_count: 0,
                completed_count: 0,
                blocked_count: 1,
                exception_count: 0
              }
            ],
            rows: [
              {
                id: "ops-outbound-hcm-260430-0002",
                area: "outbound",
                source_type: "carrier_manifest",
                source_id: "manifest-260430-ghn",
                source_reference: {
                  entity_type: "carrier_manifest",
                  id: "manifest-260430-ghn",
                  label: "MAN-260430-GHN",
                  href: "/shipping?source_id=manifest-260430-ghn&source_type=carrier_manifest",
                  unavailable: false
                },
                ref_no: "MAN-260430-GHN",
                title: "Carrier handover missing scan",
                warehouse_id: "wh-hcm",
                warehouse_code: "HCM",
                business_date: "2026-04-30",
                status: "blocked",
                severity: "danger",
                exception_code: "MISSING_HANDOVER_SCAN",
                owner: "shipping"
              }
            ]
          },
          request_id: "req-report"
        }),
        { status: 200 }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    const report = await getOperationsDailyReport({
      businessDate: "2026-04-30",
      warehouseId: "wh-hcm",
      status: "blocked"
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/reports/operations-daily?business_date=2026-04-30&warehouse_id=wh-hcm&status=blocked",
      {
        headers: {
          Authorization: "Bearer local-dev-access-token"
        }
      }
    );
    expect(report.summary.blockedCount).toBe(2);
    expect(report.rows[0]).toMatchObject({
      refNo: "MAN-260430-GHN",
      exceptionCode: "MISSING_HANDOVER_SCAN",
      sourceReference: {
        entityType: "carrier_manifest",
        id: "manifest-260430-ghn",
        label: "MAN-260430-GHN",
        href: "/shipping?source_id=manifest-260430-ghn&source_type=carrier_manifest",
        unavailable: false
      }
    });
  });

  it("downloads operations CSV with current filters", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response("ref_no,status,quantity\nGR-260430-0001,pending,12.000000\n", {
        status: 200,
        headers: {
          "Content-Disposition": `attachment; filename="operations-daily-2026-04-30-to-2026-04-30.csv"`,
          "Content-Type": "text/csv; charset=utf-8"
        }
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    const download = await downloadOperationsDailyCSV({
      fromDate: "2026-04-30",
      toDate: "2026-04-30",
      businessDate: "2026-04-30",
      warehouseId: "wh-hcm",
      status: "pending"
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/reports/operations-daily/export.csv?from_date=2026-04-30&to_date=2026-04-30&business_date=2026-04-30&warehouse_id=wh-hcm&status=pending",
      {
        headers: {
          Authorization: "Bearer local-dev-access-token"
        }
      }
    );
    await expect(download.blob.text()).resolves.toContain("GR-260430-0001");
    expect(download.filename).toBe("operations-daily-2026-04-30-to-2026-04-30.csv");
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

    await expect(getOperationsDailyReport()).rejects.toMatchObject({
      name: "ApiError",
      status: 403,
      code: "FORBIDDEN",
      requestId: "req-denied"
    });
  });

  it("builds stable query strings without blank filters", () => {
    expect(
      operationsDailyQueryString({
        fromDate: "",
        toDate: "2026-04-30",
        businessDate: "2026-04-30",
        warehouseId: "wh-hcm"
      })
    ).toBe("?to_date=2026-04-30&business_date=2026-04-30&warehouse_id=wh-hcm");
  });

  it("builds stable export filenames for the current date range", () => {
    expect(
      operationsDailyCSVFilename({
        fromDate: "2026-04-26",
        toDate: "2026-04-28",
        warehouseId: "wh-hcm",
        status: "blocked"
      })
    ).toBe("operations-daily-2026-04-26-to-2026-04-28.csv");
  });

  it("summarizes prototype areas by status", () => {
    const report = createPrototypeOperationsDailyReport({
      businessDate: "2026-04-30",
      warehouseId: "wh-hcm"
    });

    expect(report.summary).toMatchObject({
      signalCount: 7,
      pendingCount: 2,
      inProgressCount: 2,
      blockedCount: 2,
      exceptionCount: 1
    });
    expect(report.areas.map((area) => area.area)).toEqual([
      "inbound",
      "qc",
      "outbound",
      "returns",
      "stock_count",
      "subcontract"
    ]);
    expect(report.rows[0].sourceReference).toMatchObject({
      entityType: "goods_receipt",
      id: "gr-260430-0001",
      href: "/receiving?source_type=goods_receipt&source_id=gr-260430-0001",
      unavailable: false
    });
  });
});
