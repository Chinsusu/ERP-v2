import { describe, expect, it } from "vitest";
import type { MockUser } from "../../../shared/auth/mockSession";
import { buildFinanceSummaryReportHref } from "../../finance/services/financeDashboardService";
import { getVisibleReportingTabs } from "./reportingAccess";
import {
  buildWarehouseFinanceReportHref,
  buildWarehouseInventoryReportHref,
  buildWarehouseOperationsReportHref
} from "../../warehouse/services/warehouseDailyBoardService";
import { appMenuGroups, canAccessMenuItem, getVisibleMenuGroups, rolePermissions } from "../../../shared/permissions/menu";

const reportMenuItem = appMenuGroups.flatMap((group) => group.items).find((item) => item.href === "/reports");
const financeMenuItem = appMenuGroups.flatMap((group) => group.items).find((item) => item.href === "/finance");

describe("report access regression", () => {
  it("keeps dashboard report entry points behind reporting and finance permission gates", () => {
    expect(reportMenuItem).toBeDefined();
    expect(financeMenuItem).toBeDefined();

    const dashboardOnlyUser: MockUser = {
      id: "dashboard-only",
      name: "Dashboard Only",
      email: "dashboard-only@example.local",
      role: "WAREHOUSE_STAFF",
      permissions: ["dashboard:view"]
    };
    const warehouseLead: MockUser = {
      id: "warehouse-lead",
      name: "Warehouse Lead",
      email: "warehouse-lead@example.local",
      role: "WAREHOUSE_LEAD",
      permissions: rolePermissions.WAREHOUSE_LEAD
    };
    const financeUser: MockUser = {
      id: "finance-user",
      name: "Finance User",
      email: "finance@example.local",
      role: "FINANCE_OPS",
      permissions: rolePermissions.FINANCE_OPS
    };

    expect(buildWarehouseInventoryReportHref({ warehouseId: "wh-hcm", date: "2026-04-26" })).toBe(
      "/reports?report=inventory&business_date=2026-04-26&warehouse_id=wh-hcm&status=available"
    );
    expect(buildWarehouseOperationsReportHref({ warehouseId: "wh-hcm", date: "2026-04-26" })).toBe(
      "/reports?report=operations&from_date=2026-04-26&to_date=2026-04-26&business_date=2026-04-26&warehouse_id=wh-hcm"
    );
    expect(buildWarehouseFinanceReportHref({ warehouseId: "wh-hcm", date: "2026-04-26" })).toBe(
      "/reports?report=finance&from_date=2026-04-26&to_date=2026-04-26&business_date=2026-04-26"
    );
    expect(buildFinanceSummaryReportHref({ businessDate: "2026-04-30" })).toBe(
      "/reports?report=finance&from_date=2026-04-30&to_date=2026-04-30&business_date=2026-04-30"
    );

    expect(canAccessMenuItem(dashboardOnlyUser, reportMenuItem!)).toBe(false);
    expect(canAccessMenuItem(warehouseLead, reportMenuItem!)).toBe(true);
    expect(canAccessMenuItem(warehouseLead, financeMenuItem!)).toBe(false);
    expect(warehouseLead.permissions).not.toContain("reports:finance:view");
    expect(canAccessMenuItem(financeUser, reportMenuItem!)).toBe(true);
    expect(canAccessMenuItem(financeUser, financeMenuItem!)).toBe(true);
    expect(financeUser.permissions).toEqual(expect.arrayContaining(["reports:view", "reports:finance:view"]));
  });

  it("keeps finance report tabs hidden from non-finance reporting users", () => {
    const warehouseLead: MockUser = {
      id: "warehouse-lead",
      name: "Warehouse Lead",
      email: "warehouse-lead@example.local",
      role: "WAREHOUSE_LEAD",
      permissions: rolePermissions.WAREHOUSE_LEAD
    };
    const financeUser: MockUser = {
      id: "finance-user",
      name: "Finance User",
      email: "finance@example.local",
      role: "FINANCE_OPS",
      permissions: rolePermissions.FINANCE_OPS
    };

    expect(getVisibleReportingTabs(warehouseLead).map((tab) => tab.id)).toEqual(["inventory", "operations"]);
    expect(getVisibleReportingTabs(financeUser).map((tab) => tab.id)).toEqual(["inventory", "operations", "finance"]);
  });

  it("keeps module menu visibility aligned with route-level permissions", () => {
    const warehouseLead: MockUser = {
      id: "warehouse-lead",
      name: "Warehouse Lead",
      email: "warehouse-lead@example.local",
      role: "WAREHOUSE_LEAD",
      permissions: rolePermissions.WAREHOUSE_LEAD
    };
    const moduleHrefs = getVisibleMenuGroups(warehouseLead).flatMap((group) => group.items.map((item) => item.href));

    expect(moduleHrefs).toEqual(
      expect.arrayContaining(["/dashboard", "/warehouse", "/receiving", "/inventory", "/shipping", "/returns", "/reports"])
    );
    expect(moduleHrefs).not.toContain("/finance");
    expect(moduleHrefs).not.toContain("/audit-log");
    expect(moduleHrefs).not.toContain("/settings");
  });
});
