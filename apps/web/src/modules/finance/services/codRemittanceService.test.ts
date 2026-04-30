import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  approveCODRemittance,
  closeCODRemittance,
  getCODRemittances,
  recordCODRemittanceDiscrepancy,
  resetPrototypeCODRemittancesForTest,
  submitCODRemittance
} from "./codRemittanceService";

describe("codRemittanceService", () => {
  beforeEach(() => {
    resetPrototypeCODRemittancesForTest();
    vi.stubGlobal(
      "fetch",
      vi.fn(() => Promise.reject(new Error("offline")))
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("filters prototype COD remittances by status and search", async () => {
    const remittances = await getCODRemittances({ status: "draft", search: "GHN" });

    expect(remittances).toHaveLength(1);
    expect(remittances[0]).toMatchObject({
      id: "cod-remit-260430-0001",
      status: "draft",
      carrierCode: "GHN",
      discrepancyAmount: "-50000.00"
    });
  });

  it("maps API list rows and sends COD query parameters", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          success: true,
          data: [
            {
              id: "cod-api-1",
              remittance_no: "COD-API-1",
              carrier_id: "carrier-api",
              carrier_code: "API",
              carrier_name: "API Carrier",
              status: "draft",
              business_date: "2026-04-30",
              expected_amount: "500000.00",
              remitted_amount: "500000.00",
              discrepancy_amount: "0.00",
              currency_code: "VND",
              line_count: 1,
              discrepancy_count: 0,
              created_at: "2026-04-30T08:00:00Z",
              updated_at: "2026-04-30T08:05:00Z",
              version: 2
            }
          ],
          request_id: "req-cod-list"
        }),
        { status: 200 }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    const remittances = await getCODRemittances({ search: "COD-API", status: "draft" });

    expect(fetchMock).toHaveBeenCalledWith("http://localhost:8080/api/v1/cod-remittances?q=COD-API&status=draft", {
      headers: {
        Authorization: "Bearer local-dev-access-token"
      }
    });
    expect(remittances[0]).toMatchObject({
      id: "cod-api-1",
      remittanceNo: "COD-API-1",
      lineCount: 1,
      discrepancyCount: 0
    });
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
              details: { permission: "cod:reconcile" },
              request_id: "req-denied"
            }
          }),
          { status: 403 }
        )
      )
    );

    await expect(getCODRemittances()).rejects.toMatchObject({
      name: "ApiError",
      status: 403,
      code: "FORBIDDEN",
      requestId: "req-denied"
    });
  });

  it("records a discrepancy trace before submitting discrepant COD remittances", async () => {
    const traced = await recordCODRemittanceDiscrepancy("cod-remit-260430-0001", {
      lineId: "cod-remit-260430-0001-line-1",
      reason: "carrier short remittance",
      ownerId: "finance-user"
    });

    expect(traced.previousStatus).toBe("draft");
    expect(traced.currentStatus).toBe("discrepancy");
    expect(traced.codRemittance.discrepancies[0]).toMatchObject({
      type: "short_paid",
      amount: "-50000.00",
      reason: "carrier short remittance"
    });

    const submitted = await submitCODRemittance("cod-remit-260430-0001");

    expect(submitted.previousStatus).toBe("discrepancy");
    expect(submitted.currentStatus).toBe("submitted");
  });

  it("moves clean COD remittances through submit, approve, and close", async () => {
    const submitted = await submitCODRemittance("cod-remit-260430-0002");
    const approved = await approveCODRemittance("cod-remit-260430-0002");
    const closed = await closeCODRemittance("cod-remit-260430-0002");

    expect(submitted.currentStatus).toBe("submitted");
    expect(approved.currentStatus).toBe("approved");
    expect(closed.currentStatus).toBe("closed");
    expect(closed.codRemittance.auditLogId).toBe("audit-cod-close-cod-remit-260430-0002");
  });
});
