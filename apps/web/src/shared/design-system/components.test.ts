import { describe, expect, it } from "vitest";
import { coreComponentNames, statusToneClassName } from "./components";

describe("core UI components", () => {
  it("exports the S0-03-02 component set", () => {
    expect(coreComponentNames).toEqual([
      "DataTable",
      "FormSection",
      "StatusChip",
      "ConfirmDialog",
      "DetailDrawer",
      "ToastStack",
      "EmptyState",
      "LoadingState",
      "ErrorState",
      "ScanInput"
    ]);
  });

  it("maps status tone to stable class names", () => {
    expect(statusToneClassName("success")).toBe("erp-ds-status-chip erp-ds-status-chip--success");
    expect(statusToneClassName("danger")).toBe("erp-ds-status-chip erp-ds-status-chip--danger");
  });
});
