import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  createCashTransaction,
  getCashTransaction,
  getCashTransactions,
  resetPrototypeCashTransactionsForTest
} from "./cashTransactionService";

describe("cashTransactionService", () => {
  beforeEach(() => {
    resetPrototypeCashTransactionsForTest();
    vi.stubGlobal(
      "fetch",
      vi.fn(() => Promise.reject(new Error("offline")))
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("filters prototype cash transactions by direction, status, and search", async () => {
    const transactions = await getCashTransactions({ direction: "cash_out", status: "posted", search: "SUP" });

    expect(transactions).toHaveLength(1);
    expect(transactions[0]).toMatchObject({
      id: "cash-out-260430-0002",
      direction: "cash_out",
      status: "posted"
    });
  });

  it("maps API list rows and sends cash transaction query parameters", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          success: true,
          data: [
            {
              id: "cash-api-1",
              transaction_no: "CASH-IN-API-1",
              direction: "cash_in",
              status: "posted",
              business_date: "2026-04-30",
              counterparty_id: "carrier-api",
              counterparty_name: "API Carrier",
              payment_method: "bank_transfer",
              reference_no: "BANK-API-1",
              total_amount: "900000.00",
              currency_code: "VND",
              posted_by: "finance-user",
              posted_at: "2026-04-30T10:00:00Z",
              created_at: "2026-04-30T10:00:00Z",
              updated_at: "2026-04-30T10:05:00Z",
              version: 2
            }
          ],
          request_id: "req-cash-list"
        }),
        { status: 200 }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    const transactions = await getCashTransactions({ search: "API", direction: "cash_in", status: "posted" });

    expect(fetchMock).toHaveBeenCalledWith("http://localhost:8080/api/v1/cash-transactions?q=API&status=posted&direction=cash_in", {
      headers: {
        Authorization: "Bearer local-dev-access-token"
      }
    });
    expect(transactions[0]).toMatchObject({
      id: "cash-api-1",
      transactionNo: "CASH-IN-API-1",
      allocations: [],
      totalAmount: "900000.00"
    });
  });

  it("creates prototype posted cash transactions when API is unavailable", async () => {
    const transaction = await createCashTransaction({
      direction: "cash_in",
      businessDate: "2026-04-30",
      counterpartyId: "carrier-ghn",
      counterpartyName: "GHN COD",
      paymentMethod: "bank_transfer",
      referenceNo: "BANK-LOCAL-1",
      totalAmount: "1250000",
      currencyCode: "VND",
      allocations: [
        {
          id: "cash-local-line-1",
          targetType: "customer_receivable",
          targetId: "ar-cod-260430-0001",
          targetNo: "AR-COD-260430-0001",
          amount: "1250000"
        }
      ]
    });

    expect(transaction).toMatchObject({
      direction: "cash_in",
      status: "posted",
      totalAmount: "1250000.00",
      postedBy: "finance-user"
    });

    await expect(getCashTransaction(transaction.id)).resolves.toMatchObject({
      id: transaction.id,
      allocations: [{ amount: "1250000.00" }]
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
              details: { permission: "finance:manage" },
              request_id: "req-denied"
            }
          }),
          { status: 403 }
        )
      )
    );

    await expect(getCashTransactions()).rejects.toMatchObject({
      name: "ApiError",
      status: 403,
      code: "FORBIDDEN",
      requestId: "req-denied"
    });
  });
});
