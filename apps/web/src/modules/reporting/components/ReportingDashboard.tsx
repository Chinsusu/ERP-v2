"use client";

import { useState } from "react";
import { InventorySnapshotReportPanel } from "./InventorySnapshotReport";
import { OperationsDailyReportPanel } from "./OperationsDailyReport";

type ReportingTab = "inventory" | "operations";

const tabs: Array<{ id: ReportingTab; label: string }> = [
  { id: "inventory", label: "Inventory" },
  { id: "operations", label: "Operations" }
];

export function ReportingDashboard() {
  const [activeTab, setActiveTab] = useState<ReportingTab>("inventory");
  const controls = (
    <div className="erp-reporting-tabs" role="tablist" aria-label="Reporting views">
      {tabs.map((tab) => (
        <button
          key={tab.id}
          className={`erp-reporting-tab${activeTab === tab.id ? " is-active" : ""}`}
          type="button"
          role="tab"
          aria-selected={activeTab === tab.id}
          onClick={() => setActiveTab(tab.id)}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );

  if (activeTab === "operations") {
    return <OperationsDailyReportPanel controls={controls} />;
  }

  return <InventorySnapshotReportPanel controls={controls} />;
}
