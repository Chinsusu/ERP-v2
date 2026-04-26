import { describe, expect, it } from "vitest";
import type { MockUser } from "@/shared/auth/mockSession";
import {
  canAccessMenuItem,
  getVisibleActions,
  getVisibleMenuGroups,
  moduleActions,
  roleKeys,
  rolePermissions
} from "./menu";

const warehouseUser: MockUser = {
  id: "warehouse-user",
  name: "Warehouse User",
  email: "warehouse@example.local",
  role: "WAREHOUSE_STAFF",
  permissions: ["dashboard:view", "warehouse:view", "inventory:view"]
};

describe("permission menu", () => {
  it("defines the Phase 1 RBAC skeleton roles", () => {
    expect(roleKeys).toEqual([
      "CEO",
      "ERP_ADMIN",
      "WAREHOUSE_STAFF",
      "WAREHOUSE_LEAD",
      "QA",
      "SALES_OPS",
      "PRODUCTION_OPS"
    ]);
    expect(rolePermissions.ERP_ADMIN).toContain("settings:view");
    expect(rolePermissions.WAREHOUSE_STAFF).not.toContain("settings:view");
  });

  it("filters menu items by the mock user's permissions", () => {
    const groups = getVisibleMenuGroups(warehouseUser);
    const labels = groups.flatMap((group) => group.items.map((item) => item.label));

    expect(labels).toContain("Dashboard");
    expect(labels).toContain("Warehouse Daily Board");
    expect(labels).toContain("Inventory");
    expect(labels).not.toContain("Settings");
    expect(labels).not.toContain("Audit Log");
  });

  it("checks single menu item access", () => {
    const [overview] = getVisibleMenuGroups(warehouseUser);
    const dashboard = overview.items.find((item) => item.label === "Dashboard");

    expect(dashboard).toBeDefined();
    expect(canAccessMenuItem(warehouseUser, dashboard!)).toBe(true);
  });

  it("filters module actions by permission", () => {
    const actions = getVisibleActions(warehouseUser, moduleActions);
    const labels = actions.map((action) => action.label);

    expect(labels).not.toContain("Export");
    expect(labels).not.toContain("New record");
  });
});
