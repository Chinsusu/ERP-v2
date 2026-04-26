"use client";

import type { KeyboardEvent, Key, ReactNode } from "react";
import { useEffect, useRef } from "react";

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

export type DataTableProps<T> = {
  columns: DataTableColumn<T>[];
  rows: T[];
  getRowKey: (row: T, index: number) => Key;
  toolbar?: ReactNode;
  bulkActions?: ReactNode;
  emptyState?: ReactNode;
  loading?: boolean;
  error?: ReactNode;
};

export function DataTable<T>({
  columns,
  rows,
  getRowKey,
  toolbar,
  bulkActions,
  emptyState,
  loading = false,
  error
}: DataTableProps<T>) {
  if (loading) {
    return <LoadingState title="Loading records" />;
  }

  if (error) {
    return typeof error === "string" ? <ErrorState title={error} /> : <>{error}</>;
  }

  if (rows.length === 0) {
    return <>{emptyState ?? <EmptyState title="No records yet" />}</>;
  }

  return (
    <section className="erp-ds-table-shell">
      {toolbar ? <div className="erp-ds-table-toolbar">{toolbar}</div> : null}
      {bulkActions ? <div className="erp-ds-table-bulk-actions">{bulkActions}</div> : null}
      <div className="erp-ds-table-scroll">
        <table className="erp-ds-table">
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
            {rows.map((row, rowIndex) => (
              <tr key={getRowKey(row, rowIndex)}>
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
    </section>
  );
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
  cancelLabel = "Cancel",
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
          <button className="erp-ds-icon-button" type="button" aria-label="Close drawer" onClick={onClose}>
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
    <ol className="erp-ds-toast-stack" aria-live="polite" aria-label="Notifications">
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

export function LoadingState({ title, description = "Please wait while the latest data loads." }: StateBlockProps) {
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
  label = "Scan code",
  placeholder = "Scan order, batch, or tracking code",
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
