import { afterEach, describe, expect, it, vi } from "vitest";

import {
  clearClientAccessToken,
  rememberClientAccessToken,
  resolveApiAccessToken
} from "./clientSessionToken";

describe("client session access token", () => {
  afterEach(() => {
    clearClientAccessToken();
    vi.unstubAllEnvs();
  });

  it("uses the backend session token instead of the local static token", () => {
    rememberClientAccessToken("backend-access-token");

    expect(resolveApiAccessToken("local-dev-access-token")).toBe("backend-access-token");
  });

  it("keeps explicit non-mock tokens", () => {
    rememberClientAccessToken("backend-access-token");

    expect(resolveApiAccessToken("operation-token")).toBe("operation-token");
  });

  it("blocks the local static token in production-like runtime", () => {
    vi.stubEnv("NODE_ENV", "production");

    expect(resolveApiAccessToken("local-dev-access-token")).toBeUndefined();
  });

  it("allows local static tokens for explicit dev app env builds", () => {
    vi.stubEnv("NODE_ENV", "production");
    vi.stubEnv("NEXT_PUBLIC_APP_ENV", "dev");

    expect(resolveApiAccessToken("local-dev-access-token")).toBe("local-dev-access-token");
  });
});
