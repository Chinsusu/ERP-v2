import { afterEach, describe, expect, it, vi } from "vitest";

import { ApiError } from "./client";
import { isPrototypeFallbackRuntimeEnabled, shouldUsePrototypeFallback } from "./prototypeFallback";

describe("prototype fallback policy", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it("allows local prototype fallback outside production for offline errors", () => {
    vi.stubEnv("NODE_ENV", "test");

    expect(isPrototypeFallbackRuntimeEnabled()).toBe(true);
    expect(shouldUsePrototypeFallback(new Error("offline"))).toBe(true);
  });

  it("blocks prototype fallback in production runtime", () => {
    vi.stubEnv("NODE_ENV", "production");

    expect(isPrototypeFallbackRuntimeEnabled()).toBe(false);
    expect(shouldUsePrototypeFallback(new Error("offline"))).toBe(false);
  });

  it("does not hide backend API errors", () => {
    vi.stubEnv("NODE_ENV", "test");

    expect(
      shouldUsePrototypeFallback(
        new ApiError(403, {
          error: {
            code: "FORBIDDEN",
            message: "Permission denied",
            request_id: "req-denied"
          }
        })
      )
    ).toBe(false);
    expect(shouldUsePrototypeFallback(new Error("API request failed: 500"))).toBe(false);
  });
});
