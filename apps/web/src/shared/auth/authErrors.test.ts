import { describe, expect, it } from "vitest";

import { ApiError } from "@/shared/api/client";
import { loginErrorReasonFromUnknown } from "./authErrors";

function unauthorized(reason: string) {
  return new ApiError(401, {
    error: {
      code: "UNAUTHORIZED",
      message: "Authentication failed",
      details: { reason },
      request_id: "req-auth"
    }
  });
}

describe("auth error mapping", () => {
  it("maps backend auth failure reasons to safe login query reasons", () => {
    expect(loginErrorReasonFromUnknown(unauthorized("invalid_credentials"))).toBe("invalid_credentials");
    expect(loginErrorReasonFromUnknown(unauthorized("password_policy"))).toBe("password_policy");
    expect(loginErrorReasonFromUnknown(unauthorized("locked"))).toBe("locked");
  });

  it("uses a safe generic reason for unknown errors", () => {
    expect(loginErrorReasonFromUnknown(new Error("network down"))).toBe("auth_unavailable");
  });
});
