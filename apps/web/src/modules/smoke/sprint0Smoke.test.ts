import { describe, expect, it } from "vitest";
import type { MockUser } from "../../shared/auth/mockSession";
import {
  appMenuGroups,
  canAccessMenuItem,
  getPermissionsForRole,
  type RoleKey
} from "../../shared/permissions/menu";
import { sprint0FrontendSmokeChecks } from "./sprint0SmokePlan";

const menuItems = appMenuGroups.flatMap((group) => group.items);

function mockUserForRole(role: RoleKey): MockUser {
  return {
    id: `smoke-${role.toLowerCase()}`,
    name: role,
    email: `${role.toLowerCase()}@example.local`,
    role,
    permissions: getPermissionsForRole(role)
  };
}

describe("Sprint 0 frontend smoke pack", () => {
  it("covers the required frontend smoke areas", () => {
    expect(sprint0FrontendSmokeChecks.map((check) => check.href)).toEqual(
      expect.arrayContaining(["/login", "/master-data", "/sku-batch", "/inventory", "/shipping"])
    );
  });

  it("keeps protected smoke routes visible to the expected roles", () => {
    for (const check of sprint0FrontendSmokeChecks.filter((item) => item.href !== "/login")) {
      const menuItem = menuItems.find((item) => item.href === check.href);

      expect(menuItem, `${check.label} menu item`).toBeDefined();
      expect(canAccessMenuItem(mockUserForRole(check.ownerRole), menuItem!)).toBe(true);
    }
  });

  it("blocks warehouse staff from master data smoke routes", () => {
    const warehouseStaff = mockUserForRole("WAREHOUSE_STAFF");
    const masterData = menuItems.find((item) => item.href === "/master-data");

    expect(masterData).toBeDefined();
    expect(canAccessMenuItem(warehouseStaff, masterData!)).toBe(false);
  });
});
