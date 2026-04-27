import { describe, expect, it } from "vitest";
import { validateLocalCredentials, validateLocalPasswordPolicy } from "./sessionPolicy";

describe("local auth session policy", () => {
  it("accepts the seeded local credentials", () => {
    expect(validateLocalCredentials("ADMIN@example.local", "local-only-mock-password")).toEqual({ ok: true });
  });

  it("rejects weak or common passwords", () => {
    expect(validateLocalPasswordPolicy("short-1")).toContain("minimum length");
    expect(validateLocalPasswordPolicy("onlyletterslong")).toContain("number or symbol");
    expect(validateLocalPasswordPolicy("password123")).toContain("common");
  });

  it("rejects invalid credentials after policy passes", () => {
    expect(validateLocalCredentials("admin@example.local", "wrong-password!")).toEqual({
      ok: false,
      reason: "invalid_credentials",
      message: "Invalid email or password."
    });
  });
});
