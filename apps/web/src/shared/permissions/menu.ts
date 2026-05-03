import type { AuthenticatedUser } from "@/shared/auth/session";

export type RoleKey =
  | "CEO"
  | "ERP_ADMIN"
  | "WAREHOUSE_STAFF"
  | "WAREHOUSE_LEAD"
  | "QA"
  | "PURCHASE_OPS"
  | "FINANCE_OPS"
  | "SALES_OPS"
  | "PRODUCTION_OPS";

export type PermissionKey =
  | "dashboard:view"
  | "warehouse:view"
  | "inventory:view"
  | "purchase:view"
  | "finance:view"
  | "finance:manage"
  | "cod:reconcile"
  | "payment:approve"
  | "qc:view"
  | "qc:decision"
  | "production:view"
  | "subcontract:view"
  | "sales:view"
  | "shipping:view"
  | "returns:view"
  | "master-data:view"
  | "approvals:view"
  | "audit-log:view"
  | "reports:view"
  | "reports:export"
  | "reports:finance:view"
  | "settings:view"
  | "record:create"
  | "record:export";

export type AppMenuItem = {
  label: string;
  href: string;
  code: string;
  permission: PermissionKey;
};

export type AppMenuGroup = {
  label: string;
  items: AppMenuItem[];
};

export type AppAction = {
  label: string;
  permission: PermissionKey;
  variant: "primary" | "secondary";
};

export type PermissionDefinition = {
  key: PermissionKey;
  label: string;
  group: "overview" | "operations" | "data" | "control" | "action";
};

export const roleKeys: RoleKey[] = [
  "CEO",
  "ERP_ADMIN",
  "WAREHOUSE_STAFF",
  "WAREHOUSE_LEAD",
  "QA",
  "PURCHASE_OPS",
  "FINANCE_OPS",
  "SALES_OPS",
  "PRODUCTION_OPS"
];

export const permissionCatalog: PermissionDefinition[] = [
  { key: "dashboard:view", label: "Dashboard", group: "overview" },
  { key: "warehouse:view", label: "Warehouse", group: "operations" },
  { key: "inventory:view", label: "Inventory", group: "operations" },
  { key: "purchase:view", label: "Purchase", group: "operations" },
  { key: "finance:view", label: "Finance", group: "control" },
  { key: "finance:manage", label: "Finance Manage", group: "control" },
  { key: "cod:reconcile", label: "COD Reconcile", group: "action" },
  { key: "payment:approve", label: "Payment Approve", group: "action" },
  { key: "qc:view", label: "QC", group: "operations" },
  { key: "qc:decision", label: "QC Decision", group: "action" },
  { key: "production:view", label: "Production", group: "operations" },
  { key: "subcontract:view", label: "Subcontract", group: "operations" },
  { key: "sales:view", label: "Sales", group: "operations" },
  { key: "shipping:view", label: "Shipping", group: "operations" },
  { key: "returns:view", label: "Returns", group: "operations" },
  { key: "master-data:view", label: "Master Data", group: "data" },
  { key: "approvals:view", label: "Approvals", group: "control" },
  { key: "audit-log:view", label: "Audit Log", group: "control" },
  { key: "reports:view", label: "Reports", group: "control" },
  { key: "reports:export", label: "Reports Export", group: "action" },
  { key: "reports:finance:view", label: "Finance Reports", group: "control" },
  { key: "settings:view", label: "Settings", group: "control" },
  { key: "record:create", label: "Create Record", group: "action" },
  { key: "record:export", label: "Export Record", group: "action" }
];

export const rolePermissions: Record<RoleKey, PermissionKey[]> = {
  CEO: [
    "dashboard:view",
    "approvals:view",
    "audit-log:view",
    "reports:view",
    "reports:export",
    "reports:finance:view",
    "record:export"
  ],
  ERP_ADMIN: [
    "dashboard:view",
    "warehouse:view",
    "inventory:view",
    "purchase:view",
    "finance:view",
    "finance:manage",
    "cod:reconcile",
    "payment:approve",
    "qc:view",
    "qc:decision",
    "production:view",
    "subcontract:view",
    "sales:view",
    "shipping:view",
    "returns:view",
    "master-data:view",
    "approvals:view",
    "audit-log:view",
    "reports:view",
    "reports:export",
    "reports:finance:view",
    "settings:view",
    "record:create",
    "record:export"
  ],
  WAREHOUSE_STAFF: [
    "dashboard:view",
    "warehouse:view",
    "inventory:view",
    "shipping:view",
    "returns:view"
  ],
  WAREHOUSE_LEAD: [
    "dashboard:view",
    "warehouse:view",
    "inventory:view",
    "shipping:view",
    "returns:view",
    "approvals:view",
    "reports:view",
    "reports:export",
    "record:create",
    "record:export"
  ],
  QA: [
    "dashboard:view",
    "inventory:view",
    "qc:view",
    "qc:decision",
    "production:view",
    "subcontract:view",
    "returns:view",
    "reports:view",
    "record:create"
  ],
  PURCHASE_OPS: [
    "dashboard:view",
    "purchase:view",
    "master-data:view",
    "reports:view",
    "reports:export",
    "record:create",
    "record:export"
  ],
  FINANCE_OPS: [
    "dashboard:view",
    "purchase:view",
    "finance:view",
    "finance:manage",
    "cod:reconcile",
    "payment:approve",
    "reports:view",
    "reports:export",
    "reports:finance:view",
    "audit-log:view",
    "record:export"
  ],
  SALES_OPS: [
    "dashboard:view",
    "sales:view",
    "shipping:view",
    "returns:view",
    "master-data:view",
    "reports:view",
    "record:create"
  ],
  PRODUCTION_OPS: [
    "dashboard:view",
    "inventory:view",
    "qc:view",
    "production:view",
    "subcontract:view",
    "reports:view",
    "record:create"
  ]
};

export const appMenuGroups: AppMenuGroup[] = [
  {
    label: "Overview",
    items: [
      { label: "Dashboard", href: "/dashboard", code: "DB", permission: "dashboard:view" },
      { label: "Alert Center", href: "/alerts", code: "AL", permission: "dashboard:view" }
    ]
  },
  {
    label: "Operations",
    items: [
      { label: "Warehouse Daily Board", href: "/warehouse", code: "WH", permission: "warehouse:view" },
      { label: "Receiving", href: "/receiving", code: "RC", permission: "warehouse:view" },
      { label: "Inventory", href: "/inventory", code: "IV", permission: "inventory:view" },
      { label: "Purchase", href: "/purchase", code: "PU", permission: "purchase:view" },
      { label: "QC", href: "/qc", code: "QC", permission: "qc:view" },
      { label: "Production", href: "/production", code: "PD", permission: "production:view" },
      { label: "Subcontract", href: "/subcontract", code: "SUB", permission: "subcontract:view" },
      { label: "Sales Orders", href: "/sales", code: "SO", permission: "sales:view" },
      { label: "Shipping", href: "/shipping", code: "SH", permission: "shipping:view" },
      { label: "Returns", href: "/returns", code: "RT", permission: "returns:view" }
    ]
  },
  {
    label: "Data",
    items: [
      { label: "Master Data", href: "/master-data", code: "MD", permission: "master-data:view" },
      { label: "SKU / Batch", href: "/sku-batch", code: "SK", permission: "master-data:view" },
      { label: "Supplier / Factory", href: "/suppliers", code: "SF", permission: "master-data:view" },
      { label: "Customer", href: "/customers", code: "CU", permission: "master-data:view" }
    ]
  },
  {
    label: "Control",
    items: [
      { label: "Approvals", href: "/approvals", code: "AP", permission: "approvals:view" },
      { label: "Finance", href: "/finance", code: "FI", permission: "finance:view" },
      { label: "Audit Log", href: "/audit-log", code: "AU", permission: "audit-log:view" },
      { label: "Reporting", href: "/reports", code: "RP", permission: "reports:view" },
      { label: "Settings", href: "/settings", code: "ST", permission: "settings:view" }
    ]
  }
];

export const topbarActions: AppAction[] = [
  { label: "Quick create", permission: "record:create", variant: "primary" },
  { label: "Alerts", permission: "dashboard:view", variant: "secondary" },
  { label: "Docs", permission: "reports:view", variant: "secondary" }
];

export const moduleActions: AppAction[] = [
  { label: "Export", permission: "record:export", variant: "secondary" },
  { label: "New record", permission: "record:create", variant: "primary" }
];

export const reportingActions: AppAction[] = [
  { label: "Export CSV", permission: "reports:export", variant: "secondary" }
];

export function getPermissionsForRole(role: RoleKey): PermissionKey[] {
  return [...rolePermissions[role]];
}

export function hasPermission(user: AuthenticatedUser, permission: PermissionKey) {
  return user.permissions.includes(permission);
}

export function canAccessMenuItem(user: AuthenticatedUser, item: AppMenuItem) {
  return hasPermission(user, item.permission);
}

export function getVisibleMenuGroups(user: AuthenticatedUser): AppMenuGroup[] {
  return appMenuGroups
    .map((group) => ({
      ...group,
      items: group.items.filter((item) => canAccessMenuItem(user, item))
    }))
    .filter((group) => group.items.length > 0);
}

export function canUseAction(user: AuthenticatedUser, action: AppAction) {
  return hasPermission(user, action.permission);
}

export function getVisibleActions(user: AuthenticatedUser, actions: AppAction[]) {
  return actions.filter((action) => canUseAction(user, action));
}
