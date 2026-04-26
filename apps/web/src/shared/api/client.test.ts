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

    await expect(apiGet("/health")).resolves.toEqual({ status: "ok" });
  });

  it("passes bearer tokens when provided", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          success: true,
          data: { id: "user-erp-admin" },
          request_id: "req-test"
        }),
        { status: 200 }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    await apiGet("/me", { accessToken: "local-dev-access-token" });

    expect(fetchMock).toHaveBeenCalledWith("http://localhost:8080/api/v1/me", {
      headers: {
        Authorization: "Bearer local-dev-access-token"
      }
    });
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

    await expect(apiGet("/inventory/available-stock")).rejects.toMatchObject({
      name: "ApiError",
      status: 400,
      code: "VALIDATION_ERROR",
      message: "Invalid request",
      details: { field: "quantity" },
      requestId: "req-error"
    });
  });

  it("serializes typed query parameters from the generated OpenAPI contract", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          success: true,
          data: [],
          request_id: "req-test"
        }),
        { status: 200 }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    await apiGet("/inventory/available-stock", {
      accessToken: "local-dev-access-token",
      query: {
        warehouse_id: "wh-hcm",
        sku: "SERUM-30ML"
      }
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/inventory/available-stock?warehouse_id=wh-hcm&sku=SERUM-30ML",
      {
        headers: {
          Authorization: "Bearer local-dev-access-token"
        }
      }
    );
  });
});
