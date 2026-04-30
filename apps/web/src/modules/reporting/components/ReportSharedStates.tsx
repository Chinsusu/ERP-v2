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
