import { describe, expect, it } from "vitest";
import type { MockUser } from "@/shared/auth/mockSession";
import {
  canAccessMenuItem,
  getMenuItemForModule,
  getVisibleActions,
  getVisibleMenuGroups,
  moduleActions,
  permissionCatalog,
  reportingActions,
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
    expect(rolePermissions.FINANCE_OPS).toContain("finance:manage");
    expect(rolePermissions.FINANCE_OPS).toContain("cod:reconcile");
    expect(rolePermissions.FINANCE_OPS).toContain("payment:approve");
    expect(rolePermissions.FINANCE_OPS).toContain("reports:export");
    expect(rolePermissions.FINANCE_OPS).toContain("reports:finance:view");
    expect(rolePermissions.FINANCE_OPS).not.toContain("record:create");
    expect(rolePermissions.WAREHOUSE_LEAD).toContain("reports:export");
    expect(rolePermissions.WAREHOUSE_LEAD).not.toContain("reports:finance:view");
    expect(rolePermissions.WAREHOUSE_LEAD).not.toContain("qc:decision");
    expect(rolePermissions.WAREHOUSE_STAFF).not.toContain("settings:view");
    expect(rolePermissions.WAREHOUSE_STAFF).not.toContain("reports:view");
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
    expect(permissionCatalog).toContainEqual({
      key: "finance:manage",
      label: "Finance Manage",
      group: "control"
    });
    expect(permissionCatalog).toContainEqual({
      key: "cod:reconcile",
      label: "COD Reconcile",
      group: "action"
    });
    expect(permissionCatalog).toContainEqual({
      key: "payment:approve",
      label: "Payment Approve",
      group: "action"
    });
  });

  it("defines Sprint 7 reporting permissions", () => {
    expect(permissionCatalog).toContainEqual({
      key: "reports:view",
      label: "Reports",
      group: "control"
    });
    expect(permissionCatalog).toContainEqual({
      key: "reports:export",
      label: "Reports Export",
      group: "action"
    });
    expect(permissionCatalog).toContainEqual({
      key: "reports:finance:view",
      label: "Finance Reports",
      group: "control"
    });
  });

  it("uses subcontract manufacturing as the Phase 1 production entrypoint", () => {
    const labels = getVisibleMenuGroups(productionUser).flatMap((group) => group.items.map((item) => item.label));
    const productionEntrypoints = getVisibleMenuGroups(productionUser)
      .flatMap((group) => group.items)
      .filter((item) => item.href === "/subcontract");

    expect(labels).toContain("Production / Subcontract");
    expect(labels).not.toContain("Production");
    expect(labels).not.toContain("Subcontract");
    expect(productionEntrypoints).toEqual([
      {
        label: "Production / Subcontract",
        href: "/subcontract",
        code: "PD",
        permission: "subcontract:view"
      }
    ]);
    expect(getMenuItemForModule("production")).toEqual(productionEntrypoints[0]);
    expect(getMenuItemForModule("subcontract")).toEqual(productionEntrypoints[0]);
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
    expect(financeLabels).toContain("Reporting");
    expect(financeLabels).not.toContain("QC");
  });

  it("shows Sprint 22 sales and stock availability surfaces to sales UAT users", () => {
    const salesUser: MockUser = {
      id: "sales-user",
      name: "Sales User",
      email: "sales_user@example.local",
      role: "SALES_OPS",
      permissions: rolePermissions.SALES_OPS
    };

    const labels = getVisibleMenuGroups(salesUser).flatMap((group) => group.items.map((item) => item.label));

    expect(labels).toContain("Sales Orders");
    expect(labels).toContain("Inventory");
    expect(labels).not.toContain("Finance");
    expect(labels).not.toContain("Settings");
  });

  it("shows reporting to leads and protects finance reporting visibility", () => {
    const warehouseLead: MockUser = {
      id: "warehouse-lead",
      name: "Warehouse Lead",
      email: "warehouse-lead@example.local",
      role: "WAREHOUSE_LEAD",
      permissions: rolePermissions.WAREHOUSE_LEAD
    };
    const staffLabels = getVisibleMenuGroups(warehouseUser).flatMap((group) => group.items.map((item) => item.label));
    const leadLabels = getVisibleMenuGroups(warehouseLead).flatMap((group) => group.items.map((item) => item.label));

    expect(staffLabels).not.toContain("Reporting");
    expect(leadLabels).toContain("Reporting");
    expect(rolePermissions.WAREHOUSE_LEAD).not.toContain("reports:finance:view");
  });

  it("shows Sprint 22 receiving and QC surfaces to QC UAT users", () => {
    const qcUser: MockUser = {
      id: "qc-user",
      name: "QC User",
      email: "qc_user@example.local",
      role: "QA",
      permissions: rolePermissions.QA
    };

    const labels = getVisibleMenuGroups(qcUser).flatMap((group) => group.items.map((item) => item.label));

    expect(labels).toContain("Receiving");
    expect(labels).toContain("Inventory");
    expect(labels).toContain("QC");
    expect(labels).not.toContain("Sales Orders");
    expect(labels).not.toContain("Finance");
    expect(labels).not.toContain("Settings");
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

  it("filters reporting export action by reporting export permission", () => {
    const warehouseLead: MockUser = {
      id: "warehouse-lead",
      name: "Warehouse Lead",
      email: "warehouse-lead@example.local",
      role: "WAREHOUSE_LEAD",
      permissions: rolePermissions.WAREHOUSE_LEAD
    };
    const staffActions = getVisibleActions(warehouseUser, reportingActions);
    const leadActions = getVisibleActions(warehouseLead, reportingActions);

    expect(staffActions).toEqual([]);
    expect(leadActions.map((action) => action.label)).toEqual(["Export CSV"]);
  });
});
