"use client";

import type { ReactNode } from "react";
import { StatusChip, type StatusTone } from "@/shared/design-system/components";
import type { ReportSourceReference } from "../types";

type ReportStateBannerProps = {
  loading: boolean;
  error: Error | null;
  empty: boolean;
  liveLabel: string;
  emptyLabel: string;
};

type ReportExportActionProps = {
  disabled: boolean;
  exporting: boolean;
  error: Error | null;
  filename: string;
  exportedFilename: string;
  reportLabel: string;
  onExport: () => void;
};

export function ReportStateBanner({ loading, error, empty, liveLabel, emptyLabel }: ReportStateBannerProps) {
  const state = reportState({ loading, error, empty });

  return (
    <section className={`erp-reporting-state-banner erp-reporting-state-banner--${state.tone}`} aria-live="polite">
      <StatusChip tone={state.tone}>{state.label}</StatusChip>
      <span className="erp-reporting-state-banner-message">
        {error?.message ?? (loading ? "Loading report data" : empty ? emptyLabel : liveLabel)}
      </span>
    </section>
  );
}

export function ReportExportAction({
  disabled,
  exporting,
  error,
  filename,
  exportedFilename,
  reportLabel,
  onExport
}: ReportExportActionProps) {
  const state = reportExportState({ exporting, error, exportedFilename });
  const visibleFilename = exportedFilename || filename;
  const title = error?.message ?? `Export ${reportLabel} CSV as ${filename}`;

  return (
    <div className="erp-reporting-export-action">
      <span className="erp-reporting-export-status" title={error?.message ?? visibleFilename} aria-live="polite">
        <StatusChip tone={state.tone}>{state.label}</StatusChip>
        <small>{visibleFilename}</small>
      </span>
      <button
        className="erp-button erp-button--secondary"
        type="button"
        disabled={disabled || exporting}
        title={title}
        onClick={onExport}
      >
        {exporting ? "Exporting" : "Export CSV"}
      </button>
    </div>
  );
}

export function ReportSourceReferenceLink({
  reference,
  children,
  label
}: {
  reference: ReportSourceReference;
  children: ReactNode;
  label?: string;
}) {
  if (reference.unavailable) {
    return (
      <span className="erp-reporting-source-unavailable" title="Source unavailable">
        {children}
        <StatusChip tone="warning">Unavailable</StatusChip>
      </span>
    );
  }
  if (reference.href) {
    return (
      <a className="erp-reporting-source-link" href={reference.href} aria-label={`Open ${label ?? reference.label}`}>
        {children}
      </a>
    );
  }

  return <>{children}</>;
}

function reportState({
  loading,
  error,
  empty
}: {
  loading: boolean;
  error: Error | null;
  empty: boolean;
}): { label: string; tone: StatusTone } {
  if (error) {
    return { label: "API error", tone: "danger" };
  }
  if (loading) {
    return { label: "Loading", tone: "warning" };
  }
  if (empty) {
    return { label: "Empty", tone: "warning" };
  }

  return { label: "Live", tone: "info" };
}

function reportExportState({
  exporting,
  error,
  exportedFilename
}: {
  exporting: boolean;
  error: Error | null;
  exportedFilename: string;
}): { label: string; tone: StatusTone } {
  if (error) {
    return { label: "Export failed", tone: "danger" };
  }
  if (exporting) {
    return { label: "Exporting", tone: "warning" };
  }
  if (exportedFilename) {
    return { label: "Exported", tone: "success" };
  }

  return { label: "Ready", tone: "info" };
}
