"use client";

import type { KeyboardEvent, Key, ReactNode } from "react";
import { useEffect, useMemo, useRef, useState } from "react";
import {
  decimalScales,
  formatMoney,
  formatQuantity,
  formatRate,
  normalizeCurrencyCode,
  normalizeDecimalInput,
  normalizeUOMCode,
  type DecimalScale
} from "../format/numberFormat";
import { t } from "../i18n";

export const coreComponentNames = [
  "DataTable",
  "FormSection",
  "StatusChip",
  "ConfirmDialog",
  "DetailDrawer",
  "ToastStack",
  "EmptyState",
  "LoadingState",
  "ErrorState",
  "ScanInput"
] as const;

export type StatusTone = "normal" | "success" | "warning" | "danger" | "info";

export type StatusChipProps = {
  children: ReactNode;
  tone?: StatusTone;
};

export function statusToneClassName(tone: StatusTone = "normal") {
  return `erp-ds-status-chip erp-ds-status-chip--${tone}`;
}

export function StatusChip({ children, tone = "normal" }: StatusChipProps) {
  return <span className={statusToneClassName(tone)}>{children}</span>;
}

export type DataTableColumn<T> = {
  key: string;
  header: ReactNode;
  render: (row: T) => ReactNode;
  align?: "left" | "center" | "right";
  width?: string;
  sticky?: boolean;
};

export type DataTablePageSize = 10 | 20 | 50 | "all";

export const dataTablePageSizeOptions = [10, 20, 50, "all"] as const satisfies readonly DataTablePageSize[];

export type DataTablePage<T> = {
  rows: T[];
  page: number;
  pageCount: number;
  start: number;
  end: number;
  total: number;
};

export function paginateDataTableRows<T>(rows: T[], requestedPage: number, pageSize: DataTablePageSize): DataTablePage<T> {
  const total = rows.length;
  const pageCount = pageSize === "all" ? 1 : Math.max(1, Math.ceil(total / pageSize));
  const requested = Number.isFinite(requestedPage) ? Math.trunc(requestedPage) : 1;
  const page = Math.min(Math.max(requested, 1), pageCount);

  if (total === 0) {
    return { rows: [], page: 1, pageCount: 1, start: 0, end: 0, total };
  }

  if (pageSize === "all") {
    return { rows, page: 1, pageCount: 1, start: 1, end: total, total };
  }

  const startIndex = (page - 1) * pageSize;
  const pageRows = rows.slice(startIndex, startIndex + pageSize);

  return {
    rows: pageRows,
    page,
    pageCount,
    start: startIndex + 1,
    end: startIndex + pageRows.length,
    total
  };
}

export type DataTableProps<T> = {
  columns: DataTableColumn<T>[];
  rows: T[];
  getRowKey: (row: T, index: number) => Key;
  toolbar?: ReactNode;
  bulkActions?: ReactNode;
  emptyState?: ReactNode;
  loading?: boolean;
  error?: ReactNode;
  pagination?: boolean;
  initialPageSize?: DataTablePageSize;
  pageSizeOptions?: readonly DataTablePageSize[];
  preserveColumnWidths?: boolean;
  rowClassName?: (row: T, index: number) => string | undefined;
};

export function DataTable<T>({
  columns,
  rows,
  getRowKey,
  toolbar,
  bulkActions,
  emptyState,
  loading = false,
  error,
  pagination = false,
  initialPageSize = 10,
  pageSizeOptions = dataTablePageSizeOptions,
  preserveColumnWidths = false,
  rowClassName
}: DataTableProps<T>) {
  const resolvedPageSizeOptions = pageSizeOptions.length > 0 ? pageSizeOptions : dataTablePageSizeOptions;
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState<DataTablePageSize>(initialPageSize);
  const activePageSize = resolvedPageSizeOptions.includes(pageSize) ? pageSize : resolvedPageSizeOptions[0];
  const tableMinWidth = useMemo(
    () => (preserveColumnWidths ? resolveTableMinWidth(columns) : undefined),
    [columns, preserveColumnWidths]
  );
  const pageInfo = useMemo(() => paginateDataTableRows(rows, page, activePageSize), [activePageSize, page, rows]);
  const visibleRows = pagination ? pageInfo.rows : rows;

  useEffect(() => {
    if (pagination && pageInfo.page !== page) {
      setPage(pageInfo.page);
    }
  }, [page, pageInfo.page, pagination]);

  useEffect(() => {
    if (pagination) {
      setPage(1);
    }
  }, [activePageSize, pagination]);

  if (loading) {
    return <LoadingState title={t("common.loadingRecords")} />;
  }

  if (error) {
    return typeof error === "string" ? <ErrorState title={error} /> : <>{error}</>;
  }

  if (rows.length === 0) {
    return <>{emptyState ?? <EmptyState title={t("common.noRecordsYet")} />}</>;
  }

  return (
    <section className="erp-ds-table-shell">
      {toolbar ? <div className="erp-ds-table-toolbar">{toolbar}</div> : null}
      {bulkActions ? <div className="erp-ds-table-bulk-actions">{bulkActions}</div> : null}
      <div className="erp-ds-table-scroll">
        <table className="erp-ds-table" style={tableMinWidth ? { minWidth: tableMinWidth } : undefined}>
          <colgroup>
            {columns.map((column) => (
              <col key={column.key} style={{ width: column.width }} />
            ))}
          </colgroup>
          <thead>
            <tr>
              {columns.map((column) => (
                <th
                  className={column.sticky ? "erp-ds-table-cell--sticky" : undefined}
                  key={column.key}
                  style={{ width: column.width, textAlign: column.align ?? "left" }}
                  scope="col"
                >
                  {column.header}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {visibleRows.map((row, rowIndex) => (
              <tr
                className={rowClassName?.(row, pagination ? pageInfo.start - 1 + rowIndex : rowIndex)}
                key={getRowKey(row, pagination ? pageInfo.start - 1 + rowIndex : rowIndex)}
              >
                {columns.map((column) => (
                  <td
                    className={column.sticky ? "erp-ds-table-cell--sticky" : undefined}
                    key={column.key}
                    style={{ textAlign: column.align ?? "left" }}
                  >
                    {column.render(row)}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {pagination ? (
        <footer className="erp-ds-table-pagination" aria-label={t("common.tablePagination")}>
          <span className="erp-ds-table-pagination-summary">
            {t("common.paginationRange", {
              values: { start: pageInfo.start, end: pageInfo.end, total: pageInfo.total }
            })}
          </span>
          <div className="erp-ds-table-pagination-controls">
            <label className="erp-ds-table-pagination-size">
              <span>{t("common.paginationRowsPerPage")}</span>
              <select
                className="erp-input"
                value={String(activePageSize)}
                onChange={(event) => {
                  const nextPageSize =
                    resolvedPageSizeOptions.find((option) => String(option) === event.target.value) ??
                    resolvedPageSizeOptions[0];
                  setPageSize(nextPageSize);
                }}
              >
                {resolvedPageSizeOptions.map((option) => (
                  <option key={String(option)} value={String(option)}>
                    {option === "all" ? t("common.paginationAll") : option}
                  </option>
                ))}
              </select>
            </label>
            <button
              className="erp-button erp-button--secondary erp-button--compact"
              type="button"
              onClick={() => setPage((currentPage) => Math.max(1, currentPage - 1))}
              disabled={pageInfo.page <= 1}
            >
              {t("common.paginationPrevious")}
            </button>
            <button
              className="erp-button erp-button--secondary erp-button--compact"
              type="button"
              onClick={() => setPage((currentPage) => Math.min(pageInfo.pageCount, currentPage + 1))}
              disabled={pageInfo.page >= pageInfo.pageCount}
            >
              {t("common.paginationNext")}
            </button>
          </div>
        </footer>
      ) : null}
    </section>
  );
}

function resolveTableMinWidth<T>(columns: DataTableColumn<T>[]) {
  const pixelWidth = columns.reduce((total, column) => {
    const match = column.width?.match(/^(\d+)px$/);
    return match ? total + Number(match[1]) : total;
  }, 0);

  return pixelWidth > 640 ? `${pixelWidth}px` : undefined;
}
export type FormSectionProps = {
  title: string;
  description?: string;
  children: ReactNode;
  footer?: ReactNode;
};

export function FormSection({ title, description, children, footer }: FormSectionProps) {
  return (
    <section className="erp-ds-form-section">
      <header className="erp-ds-form-section-header">
        <h2>{title}</h2>
        {description ? <p>{description}</p> : null}
      </header>
      <div className="erp-ds-form-section-body">{children}</div>
      {footer ? <footer className="erp-ds-form-section-footer">{footer}</footer> : null}
    </section>
  );
}

export type ConfirmDialogProps = {
  open: boolean;
  title: string;
  description: string;
  confirmLabel: string;
  cancelLabel?: string;
  tone?: "normal" | "danger";
  onConfirm: () => void;
  onCancel: () => void;
};

export function ConfirmDialog({
  open,
  title,
  description,
  confirmLabel,
  cancelLabel = t("actions.Cancel"),
  tone = "normal",
  onConfirm,
  onCancel
}: ConfirmDialogProps) {
  if (!open) {
    return null;
  }

  return (
    <div className="erp-ds-dialog-backdrop">
      <section className="erp-ds-confirm-dialog" role="dialog" aria-modal="true" aria-labelledby="erp-confirm-title">
        <h2 id="erp-confirm-title">{title}</h2>
        <p>{description}</p>
        <footer className="erp-ds-dialog-actions">
          <button className="erp-button erp-button--secondary" type="button" onClick={onCancel}>
            {cancelLabel}
          </button>
          <button
            className={`erp-button erp-button--${tone === "danger" ? "danger" : "primary"}`}
            type="button"
            onClick={onConfirm}
          >
            {confirmLabel}
          </button>
        </footer>
      </section>
    </div>
  );
}

export type DetailDrawerProps = {
  open: boolean;
  title: string;
  subtitle?: string;
  children: ReactNode;
  footer?: ReactNode;
  onClose: () => void;
};

export function DetailDrawer({ open, title, subtitle, children, footer, onClose }: DetailDrawerProps) {
  if (!open) {
    return null;
  }

  return (
    <div className="erp-ds-drawer-backdrop">
      <aside className="erp-ds-drawer" role="dialog" aria-modal="true" aria-labelledby="erp-drawer-title">
        <header className="erp-ds-drawer-header">
          <div>
            <h2 id="erp-drawer-title">{title}</h2>
            {subtitle ? <p>{subtitle}</p> : null}
          </div>
          <button className="erp-ds-icon-button" type="button" aria-label={t("common.closeDrawer")} onClick={onClose}>
            x
          </button>
        </header>
        <div className="erp-ds-drawer-body">{children}</div>
        {footer ? <footer className="erp-ds-drawer-footer">{footer}</footer> : null}
      </aside>
    </div>
  );
}

export type ToastMessage = {
  id: string;
  title: string;
  description?: string;
  tone?: StatusTone;
};

export type ToastStackProps = {
  messages: ToastMessage[];
};

export function ToastStack({ messages }: ToastStackProps) {
  if (messages.length === 0) {
    return null;
  }

  return (
    <ol className="erp-ds-toast-stack" aria-live="polite" aria-label={t("common.notifications")}>
      {messages.map((message) => (
        <li className={`erp-ds-toast erp-ds-toast--${message.tone ?? "normal"}`} key={message.id}>
          <strong>{message.title}</strong>
          {message.description ? <span>{message.description}</span> : null}
        </li>
      ))}
    </ol>
  );
}

export type StateBlockProps = {
  title: string;
  description?: string;
  action?: ReactNode;
};

export function EmptyState({ title, description, action }: StateBlockProps) {
  return <StateBlock tone="empty" title={title} description={description} action={action} />;
}

export function LoadingState({ title, description = t("common.loadingDescription") }: StateBlockProps) {
  return <StateBlock tone="loading" title={title} description={description} />;
}

export function ErrorState({ title, description, action }: StateBlockProps) {
  return <StateBlock tone="error" title={title} description={description} action={action} />;
}

function StateBlock({ tone, title, description, action }: StateBlockProps & { tone: "empty" | "loading" | "error" }) {
  return (
    <section className={`erp-ds-state erp-ds-state--${tone}`}>
      <span className="erp-ds-state-mark" aria-hidden="true" />
      <h2>{title}</h2>
      {description ? <p>{description}</p> : null}
      {action ? <div className="erp-ds-state-action">{action}</div> : null}
    </section>
  );
}

export type ScanInputProps = {
  label?: string;
  placeholder?: string;
  feedback?: ToastMessage;
  autoFocus?: boolean;
  onScan?: (value: string) => void;
};

export function ScanInput({
  label = t("common.scanCode"),
  placeholder = t("common.scanPlaceholder"),
  feedback,
  autoFocus = true,
  onScan
}: ScanInputProps) {
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (autoFocus) {
      inputRef.current?.focus();
    }
  }, [autoFocus]);

  function handleKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key !== "Enter") {
      return;
    }

    const value = event.currentTarget.value.trim();
    if (value === "") {
      return;
    }

    onScan?.(value);
    event.currentTarget.value = "";
  }

  return (
    <label className="erp-ds-scan-input">
      <span>{label}</span>
      <input ref={inputRef} type="text" placeholder={placeholder} onKeyDown={handleKeyDown} />
      {feedback ? (
        <small className={`erp-ds-scan-feedback erp-ds-scan-feedback--${feedback.tone ?? "normal"}`}>
          {feedback.title}
        </small>
      ) : null}
    </label>
  );
}

export type MoneyDisplayProps = {
  value: string;
  currencyCode?: string;
};

export function MoneyDisplay({ value, currencyCode = "VND" }: MoneyDisplayProps) {
  return <span className="erp-ds-decimal erp-ds-decimal--money">{formatMoney(value, currencyCode)}</span>;
}

export type QuantityDisplayProps = {
  value: string;
  uomCode?: string;
};

export function QuantityDisplay({ value, uomCode }: QuantityDisplayProps) {
  return <span className="erp-ds-decimal erp-ds-decimal--quantity">{formatQuantity(value, uomCode)}</span>;
}

export type RateDisplayProps = {
  value: string;
};

export function RateDisplay({ value }: RateDisplayProps) {
  return <span className="erp-ds-decimal erp-ds-decimal--rate">{formatRate(value)}</span>;
}

export type CurrencyCodeDisplayProps = {
  value?: string;
};

export function CurrencyCodeDisplay({ value = "VND" }: CurrencyCodeDisplayProps) {
  return <span className="erp-ds-code">{normalizeCurrencyCode(value)}</span>;
}

export type UOMCodeDisplayProps = {
  value: string;
};

export function UOMCodeDisplay({ value }: UOMCodeDisplayProps) {
  return <span className="erp-ds-code">{normalizeUOMCode(value)}</span>;
}

export type DecimalInputProps = {
  label: string;
  value: string;
  scale?: DecimalScale;
  suffix?: string;
  onChange: (value: string) => void;
};

export function DecimalInput({ label, value, scale = decimalScales.quantity, suffix, onChange }: DecimalInputProps) {
  function handleBlur() {
    try {
      onChange(normalizeDecimalInput(value, scale));
    } catch {
      // Keep the raw value so the owning form can show its existing validation error.
    }
  }

  return (
    <label className="erp-ds-decimal-input">
      <span>{label}</span>
      <span className="erp-ds-decimal-input-control">
        <input inputMode="decimal" type="text" value={value} onBlur={handleBlur} onChange={(event) => onChange(event.currentTarget.value)} />
        {suffix ? <small>{suffix}</small> : null}
      </span>
    </label>
  );
}
