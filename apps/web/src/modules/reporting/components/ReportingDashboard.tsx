"use client";

import { FinanceSummaryReportPanel } from "./FinanceSummaryReport";
import { InventorySnapshotReportPanel } from "./InventorySnapshotReport";
import { OperationsDailyReportPanel } from "./OperationsDailyReport";
import type { AuthenticatedUser } from "../../../shared/auth/session";
import { getVisibleReportingTabs } from "../services/reportingAccess";
import { urlOptionParam, useReportUrlState } from "../hooks/useReportUrlState";

type ReportingDashboardProps = {
  user: AuthenticatedUser;
};

export function ReportingDashboard({ user }: ReportingDashboardProps) {
  const { searchParams, replaceReportUrlParams } = useReportUrlState();
  const visibleTabs = getVisibleReportingTabs(user);
  const activeTab = urlOptionParam(
    searchParams,
    "report",
    visibleTabs.map((tab) => tab.id),
    "inventory"
  );
  const controls = (
    <div className="erp-reporting-tabs" role="tablist" aria-label="Reporting views">
      {visibleTabs.map((tab) => (
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
