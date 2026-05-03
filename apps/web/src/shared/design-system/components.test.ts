import { describe, expect, it } from "vitest";
import {
  coreComponentNames,
  MoneyDisplay,
  paginateDataTableRows,
  statusToneClassName,
  UOMCodeDisplay
} from "./components";

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

  it("exports decimal display components", () => {
    expect(MoneyDisplay).toBeTypeOf("function");
    expect(UOMCodeDisplay).toBeTypeOf("function");
  });

  it("paginates table rows with a 10 row default page size", () => {
    const result = paginateDataTableRows(Array.from({ length: 23 }, (_, index) => index + 1), 1, 10);

    expect(result.rows).toEqual([1, 2, 3, 4, 5, 6, 7, 8, 9, 10]);
    expect(result.page).toBe(1);
    expect(result.pageCount).toBe(3);
    expect(result.start).toBe(1);
    expect(result.end).toBe(10);
    expect(result.total).toBe(23);
  });

  it("clamps table pagination to the available page range", () => {
    const result = paginateDataTableRows(Array.from({ length: 23 }, (_, index) => index + 1), 99, 20);

    expect(result.rows).toEqual([21, 22, 23]);
    expect(result.page).toBe(2);
    expect(result.pageCount).toBe(2);
    expect(result.start).toBe(21);
    expect(result.end).toBe(23);
  });

  it("shows all table rows when page size is all", () => {
    const result = paginateDataTableRows(Array.from({ length: 12 }, (_, index) => index + 1), 2, "all");

    expect(result.rows).toHaveLength(12);
    expect(result.page).toBe(1);
    expect(result.pageCount).toBe(1);
    expect(result.start).toBe(1);
    expect(result.end).toBe(12);
  });
});
