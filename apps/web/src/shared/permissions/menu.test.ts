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
      "PURCHASE_OPS",
      "FINANCE_OPS",
      "SALES_OPS",
      "PRODUCTION_OPS"
    ]);
    expect(rolePermissions.ERP_ADMIN).toContain("settings:view");
    expect(rolePermissions.QA).toContain("qc:decision");
    expect(rolePermissions.PURCHASE_OPS).toContain("purchase:view");
    expect(rolePermissions.PURCHASE_OPS).not.toContain("qc:decision");
    expect(rolePermissions.FINANCE_OPS).toContain("finance:view");
    expect(rolePermissions.FINANCE_OPS).not.toContain("record:create");
    expect(rolePermissions.WAREHOUSE_LEAD).not.toContain("qc:decision");
    expect(rolePermissions.WAREHOUSE_STAFF).not.toContain("settings:view");
  });

  it("filters menu items by the mock user's permissions", () => {
    const groups = getVisibleMenuGroups(warehouseUser);
    const labels = groups.flatMap((group) => group.items.map((item) => item.label));

    expect(labels).toContain("Dashboard");
    expect(labels).toContain("Warehouse Daily Board");
    expect(labels).toContain("Receiving");
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

  it("defines finance as a control permission", () => {
    expect(permissionCatalog).toContainEqual({
      key: "finance:view",
      label: "Finance",
      group: "control"
    });
  });

  it("shows subcontract operations only to users with subcontract access", () => {
    const labels = getVisibleMenuGroups(productionUser).flatMap((group) => group.items.map((item) => item.label));

    expect(labels).toContain("Production");
    expect(labels).toContain("Subcontract");
  });

  it("shows purchase and finance menus to their Sprint 4 roles", () => {
    const purchaseUser: MockUser = {
      id: "purchase-user",
      name: "Purchase User",
      email: "purchase@example.local",
      role: "PURCHASE_OPS",
      permissions: rolePermissions.PURCHASE_OPS
    };
    const financeUser: MockUser = {
      id: "finance-user",
      name: "Finance User",
      email: "finance@example.local",
      role: "FINANCE_OPS",
      permissions: rolePermissions.FINANCE_OPS
    };

    const purchaseLabels = getVisibleMenuGroups(purchaseUser).flatMap((group) => group.items.map((item) => item.label));
    const financeLabels = getVisibleMenuGroups(financeUser).flatMap((group) => group.items.map((item) => item.label));

    expect(purchaseLabels).toContain("Purchase");
    expect(purchaseLabels).not.toContain("QC");
    expect(financeLabels).toContain("Finance");
    expect(financeLabels).toContain("Purchase");
    expect(financeLabels).not.toContain("QC");
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
