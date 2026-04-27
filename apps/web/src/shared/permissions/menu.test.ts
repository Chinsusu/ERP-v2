import { describe, expect, it } from "vitest";
import type { MockUser } from "@/shared/auth/mockSession";
import {
  canAccessMenuItem,
  getVisibleActions,
  getVisibleMenuGroups,
  moduleActions,
  permissionCatalog,
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

const productionUser: MockUser = {
  id: "production-user",
  name: "Production User",
  email: "production@example.local",
  role: "PRODUCTION_OPS",
  permissions: ["dashboard:view", "production:view", "subcontract:view"]
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
    expect(rolePermissions.QA).toContain("qc:decision");
    expect(rolePermissions.WAREHOUSE_LEAD).not.toContain("qc:decision");
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
    expect(labels).not.toContain("Subcontract");
  });

  it("keeps role permissions inside the shared permission catalog", () => {
    const knownPermissions = new Set(permissionCatalog.map((permission) => permission.key));

    for (const permissions of Object.values(rolePermissions)) {
      for (const permission of permissions) {
        expect(knownPermissions.has(permission), permission).toBe(true);
      }
    }
  });

  it("defines subcontract as an operations permission", () => {
    expect(permissionCatalog).toContainEqual({
      key: "subcontract:view",
      label: "Subcontract",
      group: "operations"
    });
  });

  it("shows subcontract operations only to users with subcontract access", () => {
    const labels = getVisibleMenuGroups(productionUser).flatMap((group) => group.items.map((item) => item.label));

    expect(labels).toContain("Production");
    expect(labels).toContain("Subcontract");
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
