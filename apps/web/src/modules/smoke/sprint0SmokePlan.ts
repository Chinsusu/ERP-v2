export type Sprint0SmokeCheck = {
  id: string;
  label: string;
  href: string;
  ownerRole: "ERP_ADMIN" | "WAREHOUSE_STAFF" | "SALES_OPS";
};

export const sprint0FrontendSmokeChecks: Sprint0SmokeCheck[] = [
  {
    id: "S0SMK-002",
    label: "Login",
    href: "/login",
    ownerRole: "ERP_ADMIN"
  },
  {
    id: "S0SMK-003A",
    label: "Master Data",
    href: "/master-data",
    ownerRole: "ERP_ADMIN"
  },
  {
    id: "S0SMK-004",
    label: "Stock Movement",
    href: "/inventory",
    ownerRole: "ERP_ADMIN"
  },
  {
    id: "S0SMK-005",
    label: "Scan Handover",
    href: "/shipping",
    ownerRole: "WAREHOUSE_STAFF"
  }
];
