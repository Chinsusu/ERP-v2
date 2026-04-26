import type { MockUser } from "@/shared/auth/mockSession";

export type RoleKey =
  | "CEO"
  | "ERP_ADMIN"
  | "WAREHOUSE_STAFF"
  | "WAREHOUSE_LEAD"
  | "QA"
  | "SALES_OPS"
  | "PRODUCTION_OPS";

export type PermissionKey =
  | "dashboard:view"
  | "warehouse:view"
  | "inventory:view"
  | "purchase:view"
  | "qc:view"
  | "production:view"
  | "sales:view"
  | "shipping:view"
  | "returns:view"
  | "master-data:view"
  | "approvals:view"
  | "audit-log:view"
  | "reports:view"
  | "settings:view";

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
      { label: "Inventory", href: "/inventory", code: "IV", permission: "inventory:view" },
      { label: "Purchase", href: "/purchase", code: "PU", permission: "purchase:view" },
      { label: "QC", href: "/qc", code: "QC", permission: "qc:view" },
      { label: "Production", href: "/production", code: "PD", permission: "production:view" },
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
      { label: "Audit Log", href: "/audit-log", code: "AU", permission: "audit-log:view" },
      { label: "Reports", href: "/reports", code: "RP", permission: "reports:view" },
      { label: "Settings", href: "/settings", code: "ST", permission: "settings:view" }
    ]
  }
];

export function canAccessMenuItem(user: MockUser, item: AppMenuItem) {
  return user.permissions.includes(item.permission);
}

export function getVisibleMenuGroups(user: MockUser): AppMenuGroup[] {
  return appMenuGroups
    .map((group) => ({
      ...group,
      items: group.items.filter((item) => canAccessMenuItem(user, item))
    }))
    .filter((group) => group.items.length > 0);
}
