import type { MockUser } from "../../../shared/auth/mockSession";
import { hasPermission } from "../../../shared/permissions/menu";

export type ReportingTab = "inventory" | "operations" | "finance";

export const reportingTabs: Array<{ id: ReportingTab; label: string }> = [
  { id: "inventory", label: "Inventory" },
  { id: "operations", label: "Operations" },
  { id: "finance", label: "Finance" }
];

export function getVisibleReportingTabs(user: MockUser) {
  return reportingTabs.filter((tab) => tab.id !== "finance" || hasPermission(user, "reports:finance:view"));
}
