import { afterEach, describe, expect, it, vi } from "vitest";
import {
  auditActionTone,
  compactAuditPayload,
  getAuditLogs,
  prototypeAuditLogs,
  summarizeAuditLogs
} from "./auditLogService";

describe("auditLogService", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("summarizes audit events by operational category", () => {
    expect(summarizeAuditLogs(prototypeAuditLogs)).toEqual({
      total: 3,
      adjustments: 1,
      securityEvents: 1,
      latestEventAt: "2026-04-26T08:30:00Z"
    });
  });

  it("maps sensitive actions to visible tones", () => {
    expect(auditActionTone("inventory.stock_movement.adjusted")).toBe("warning");
    expect(auditActionTone("security.role.assigned")).toBe("danger");
    expect(auditActionTone("qc.lot.released")).toBe("success");
  });

  it("compacts metadata into a scannable payload", () => {
    expect(compactAuditPayload({ reason: "cycle count", source: "inventory" })).toBe(
      "reason: cycle count, source: inventory"
    );
    expect(compactAuditPayload()).toBe("-");
  });

  it("falls back to prototype rows when the API is not reachable", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("offline")));

    await expect(getAuditLogs({ action: "security.role.assigned" })).resolves.toEqual([prototypeAuditLogs[1]]);
  });
});
