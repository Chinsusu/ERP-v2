import { afterEach, describe, expect, it, vi } from "vitest";
import { ApiError, apiGet } from "./client";

describe("apiGet", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("unwraps successful API envelopes", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            success: true,
            data: { status: "ok" },
            request_id: "req-test"
          }),
          { status: 200 }
        )
      )
    );

    await expect(apiGet<{ status: string }>("/health")).resolves.toEqual({ status: "ok" });
  });

  it("throws structured API errors", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            error: {
              code: "VALIDATION_ERROR",
              message: "Invalid request",
              details: { field: "quantity" },
              request_id: "req-error"
            }
          }),
          { status: 400 }
        )
      )
    );

    await expect(apiGet("/inventory/stock-movements")).rejects.toMatchObject({
      name: "ApiError",
      status: 400,
      code: "VALIDATION_ERROR",
      message: "Invalid request",
      details: { field: "quantity" },
      requestId: "req-error"
    });
  });
});
