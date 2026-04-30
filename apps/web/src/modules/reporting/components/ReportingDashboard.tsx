"use client";

import { FinanceSummaryReportPanel } from "./FinanceSummaryReport";
import { InventorySnapshotReportPanel } from "./InventorySnapshotReport";
import { OperationsDailyReportPanel } from "./OperationsDailyReport";
import { urlOptionParam, useReportUrlState } from "../hooks/useReportUrlState";

type ReportingTab = "inventory" | "operations" | "finance";

const tabs: Array<{ id: ReportingTab; label: string }> = [
  { id: "inventory", label: "Inventory" },
  { id: "operations", label: "Operations" },
  { id: "finance", label: "Finance" }
];

export function ReportingDashboard() {
  const { searchParams, replaceReportUrlParams } = useReportUrlState();
  const activeTab = urlOptionParam(
    searchParams,
    "report",
    tabs.map((tab) => tab.id),
    "inventory"
  );
  const controls = (
    <div className="erp-reporting-tabs" role="tablist" aria-label="Reporting views">
      {tabs.map((tab) => (
        <button
          key={tab.id}
          className={`erp-reporting-tab${activeTab === tab.id ? " is-active" : ""}`}
          type="button"
          role="tab"
          aria-selected={activeTab === tab.id}
          onClick={() => replaceReportUrlParams(tab.id, {})}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );

  if (activeTab === "operations") {
    return <OperationsDailyReportPanel controls={controls} />;
  }
  if (activeTab === "finance") {
    return <FinanceSummaryReportPanel controls={controls} />;
  }

  return <InventorySnapshotReportPanel controls={controls} />;
}
