import type { ReactNode } from "react";
import { formatDateTimeVI } from "../format/numberFormat";
import { t } from "../i18n";
import type { StatusTone } from "./components";

export const pageTemplateNames = [
  "AppShell",
  "PageHeader",
  "FilterBar",
  "TablePageTemplate",
  "FormPageTemplate",
  "DetailPageTemplate",
  "ModalTemplate",
  "DrawerTemplate",
  "PopoverTemplate",
  "EmptyState",
  "LoadingState",
  "ErrorState",
  "AuditLogPanel",
  "AttachmentPanel"
] as const;

export type PageTemplateVariant = "app-shell" | "dashboard" | "table" | "form" | "detail";
export type TemplatePanelTone = "plain" | StatusTone;
export type OverlayTemplateVariant = "modal" | "drawer" | "popover";

export function pageTemplateClassName(variant: PageTemplateVariant = "table") {
  return `erp-ds-page-template erp-ds-page-template--${variant}`;
}

export function templatePanelClassName(tone: TemplatePanelTone = "plain") {
  return `erp-ds-template-panel erp-ds-template-panel--${tone}`;
}

export function overlayTemplateClassName(variant: OverlayTemplateVariant) {
  return `erp-ds-overlay-template erp-ds-overlay-template--${variant}`;
}

export type PageHeaderProps = {
  eyebrow?: ReactNode;
  breadcrumbs?: readonly ReactNode[];
  title: ReactNode;
  description?: ReactNode;
  status?: ReactNode;
  actions?: ReactNode;
  meta?: ReactNode;
};

export function PageHeader({ eyebrow, breadcrumbs = [], title, description, status, actions, meta }: PageHeaderProps) {
  return (
    <header className="erp-ds-page-header">
      <div className="erp-ds-page-header-main">
        {breadcrumbs.length > 0 ? (
          <nav className="erp-ds-breadcrumbs" aria-label={t("common.breadcrumb")}>
            <ol>
              {breadcrumbs.map((breadcrumb, index) => (
                <li key={index}>{breadcrumb}</li>
              ))}
            </ol>
          </nav>
        ) : null}
        {eyebrow ? <p className="erp-ds-page-eyebrow">{eyebrow}</p> : null}
        <div className="erp-ds-page-title-row">
          <h1>{title}</h1>
          {status ? <div className="erp-ds-page-status">{status}</div> : null}
        </div>
        {description ? <p className="erp-ds-page-description">{description}</p> : null}
        {meta ? <div className="erp-ds-page-meta">{meta}</div> : null}
      </div>
      {actions ? (
        <div className="erp-ds-page-actions" aria-label={t("common.pageActions")}>
          {actions}
        </div>
      ) : null}
    </header>
  );
}

export type FilterBarProps = {
  children: ReactNode;
  actions?: ReactNode;
  label?: string;
};

export function FilterBar({ children, actions, label = t("common.filters") }: FilterBarProps) {
  return (
    <section className="erp-ds-filter-bar" aria-label={label}>
      <div className="erp-ds-filter-fields">{children}</div>
      {actions ? <div className="erp-ds-filter-actions">{actions}</div> : null}
    </section>
  );
}

export type PageTemplateProps = {
  variant?: PageTemplateVariant;
  header: ReactNode;
  filters?: ReactNode;
  children: ReactNode;
  footer?: ReactNode;
};

export function PageTemplate({ variant = "table", header, filters, children, footer }: PageTemplateProps) {
  return (
    <section className={pageTemplateClassName(variant)}>
      {header}
      {filters}
      <div className="erp-ds-page-template-body">{children}</div>
      {footer ? <StickyFooter>{footer}</StickyFooter> : null}
    </section>
  );
}

export type TablePageTemplateProps = {
  header: ReactNode;
  filters?: ReactNode;
  table: ReactNode;
  bulkActions?: ReactNode;
};

export function TablePageTemplate({ header, filters, table, bulkActions }: TablePageTemplateProps) {
  return (
    <PageTemplate variant="table" header={header} filters={filters}>
      {bulkActions ? <div className="erp-ds-template-bulk-actions">{bulkActions}</div> : null}
      {table}
    </PageTemplate>
  );
}

export type FormPageTemplateProps = {
  header: ReactNode;
  children: ReactNode;
  footer: ReactNode;
};

export function FormPageTemplate({ header, children, footer }: FormPageTemplateProps) {
  return (
    <PageTemplate variant="form" header={header} footer={footer}>
      <div className="erp-ds-form-template-body">{children}</div>
    </PageTemplate>
  );
}

export type DetailPageTemplateProps = {
  header: ReactNode;
  main: ReactNode;
  aside?: ReactNode;
  tabs?: ReactNode;
};

export function DetailPageTemplate({ header, main, aside, tabs }: DetailPageTemplateProps) {
  return (
    <PageTemplate variant="detail" header={header}>
      {tabs ? <div className="erp-ds-detail-tabs">{tabs}</div> : null}
      <div className="erp-ds-detail-layout">
        <div className="erp-ds-detail-main">{main}</div>
        {aside ? <aside className="erp-ds-detail-aside">{aside}</aside> : null}
      </div>
    </PageTemplate>
  );
}

export type StickyFooterProps = {
  children: ReactNode;
};

export function StickyFooter({ children }: StickyFooterProps) {
  return (
    <footer className="erp-ds-sticky-footer" aria-label={t("common.pageActions")}>
      {children}
    </footer>
  );
}

export type TemplatePanelProps = {
  title?: ReactNode;
  description?: ReactNode;
  actions?: ReactNode;
  children: ReactNode;
  footer?: ReactNode;
  tone?: TemplatePanelTone;
};

export function TemplatePanel({
  title,
  description,
  actions,
  children,
  footer,
  tone = "plain"
}: TemplatePanelProps) {
  return (
    <section className={templatePanelClassName(tone)}>
      {title || description || actions ? (
        <header className="erp-ds-template-panel-header">
          <div>
            {title ? <h2>{title}</h2> : null}
            {description ? <p>{description}</p> : null}
          </div>
          {actions ? <div className="erp-ds-template-panel-actions">{actions}</div> : null}
        </header>
      ) : null}
      <div className="erp-ds-template-panel-body">{children}</div>
      {footer ? <footer className="erp-ds-template-panel-footer">{footer}</footer> : null}
    </section>
  );
}

export type OverlayTemplateProps = {
  title: ReactNode;
  description?: ReactNode;
  children: ReactNode;
  footer?: ReactNode;
  variant: OverlayTemplateVariant;
};

function OverlayTemplate({ title, description, children, footer, variant }: OverlayTemplateProps) {
  return (
    <section
      className={overlayTemplateClassName(variant)}
      role={variant === "popover" ? "region" : "dialog"}
      aria-label={typeof title === "string" ? title : undefined}
      aria-modal={variant === "popover" ? undefined : "true"}
    >
      <header className="erp-ds-overlay-header">
        <h2>{title}</h2>
        {description ? <p>{description}</p> : null}
      </header>
      <div className="erp-ds-overlay-body">{children}</div>
      {footer ? <footer className="erp-ds-overlay-footer">{footer}</footer> : null}
    </section>
  );
}

export function ModalTemplate(props: Omit<OverlayTemplateProps, "variant">) {
  return <OverlayTemplate {...props} variant="modal" />;
}

export function DrawerTemplate(props: Omit<OverlayTemplateProps, "variant">) {
  return <OverlayTemplate {...props} variant="drawer" />;
}

export function PopoverTemplate(props: Omit<OverlayTemplateProps, "variant">) {
  return <OverlayTemplate {...props} variant="popover" />;
}

export type AuditLogPanelEntry = {
  id: string;
  occurredAt: string;
  actor: string;
  action: string;
  before?: ReactNode;
  after?: ReactNode;
  metadata?: ReactNode;
};

export type AuditLogPanelProps = {
  entries: readonly AuditLogPanelEntry[];
  title?: ReactNode;
};

export function AuditLogPanel({ entries, title = t("common.auditLog") }: AuditLogPanelProps) {
  return (
    <TemplatePanel title={title} description={t("common.auditDescription")}>
      {entries.length > 0 ? (
        <ol className="erp-ds-audit-list">
          {entries.map((entry) => (
            <li className="erp-ds-audit-entry" key={entry.id}>
              <time dateTime={entry.occurredAt}>{entry.occurredAt}</time>
              <strong>{entry.action}</strong>
              <span>{entry.actor}</span>
              {entry.before || entry.after ? (
                <div className="erp-ds-audit-diff">
                  <span>{entry.before ?? t("common.beforeNA")}</span>
                  <span>{entry.after ?? t("common.afterNA")}</span>
                </div>
              ) : null}
              {entry.metadata ? <small>{entry.metadata}</small> : null}
            </li>
          ))}
        </ol>
      ) : (
        <p className="erp-ds-panel-empty">{t("common.noAuditEvents")}</p>
      )}
    </TemplatePanel>
  );
}

export type AttachmentPanelItem = {
  id: string;
  name: string;
  kind: string;
  uploadedBy: string;
  uploadedAt: string;
  detail?: ReactNode;
  status?: ReactNode;
  storageKey?: ReactNode;
  canDownload?: boolean;
  canDelete?: boolean;
  downloadLabel?: string;
  deleteLabel?: string;
  onDownload?: () => void;
  onDelete?: () => void;
  action?: ReactNode;
};

export type AttachmentPanelProps = {
  items: readonly AttachmentPanelItem[];
  title?: ReactNode;
  action?: ReactNode;
  uploadAction?: ReactNode;
  emptyMessage?: ReactNode;
};

export function AttachmentPanel({
  items,
  title = t("common.attachments"),
  action,
  uploadAction,
  emptyMessage = t("common.noAttachments")
}: AttachmentPanelProps) {
  return (
    <TemplatePanel
      title={title}
      description={t("common.attachmentDescription")}
      actions={
        action || uploadAction ? (
          <div className="erp-ds-attachment-toolbar">
            {action}
            {uploadAction}
          </div>
        ) : null
      }
    >
      {items.length > 0 ? (
        <div className="erp-ds-attachment-list">
          {items.map((item) => (
            <article className="erp-ds-attachment-row" key={item.id}>
              <div>
                <strong>{item.name}</strong>
                <span>{item.kind}</span>
                {item.storageKey ? <small className="erp-ds-attachment-storage">{item.storageKey}</small> : null}
              </div>
              <div>
                <span>{item.uploadedBy}</span>
                <time dateTime={item.uploadedAt}>{formatDateTimeVI(item.uploadedAt)}</time>
                {item.detail ? <small>{item.detail}</small> : null}
              </div>
              {item.status ? <div className="erp-ds-attachment-status">{item.status}</div> : null}
              {item.action || item.canDownload || item.canDelete ? (
                <div className="erp-ds-attachment-action">
                  {item.action}
                  {item.canDownload ? (
                    <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={item.onDownload}>
                      {item.downloadLabel ?? t("common.downloadAttachment")}
                    </button>
                  ) : null}
                  {item.canDelete ? (
                    <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={item.onDelete}>
                      {item.deleteLabel ?? t("common.deleteAttachment")}
                    </button>
                  ) : null}
                </div>
              ) : null}
            </article>
          ))}
        </div>
      ) : (
        <p className="erp-ds-panel-empty">{emptyMessage}</p>
      )}
    </TemplatePanel>
  );
}
