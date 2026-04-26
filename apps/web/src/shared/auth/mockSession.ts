import type { PermissionKey, RoleKey } from "@/shared/permissions/menu";

export type MockUser = {
  id: string;
  name: string;
  email: string;
  role: RoleKey;
  permissions: PermissionKey[];
};

export type MockSession = {
  isAuthenticated: boolean;
  user: MockUser;
};

export const mockSession: MockSession = {
  isAuthenticated: true,
  user: {
    id: "user-erp-admin",
    name: "ERP Admin",
    email: "admin@example.local",
    role: "ERP_ADMIN",
    permissions: [
      "dashboard:view",
      "warehouse:view",
      "inventory:view",
      "purchase:view",
      "qc:view",
      "production:view",
      "sales:view",
      "shipping:view",
      "returns:view",
      "master-data:view",
      "approvals:view",
      "audit-log:view",
      "reports:view",
      "settings:view"
    ]
  }
};
