import type { ReactNode } from "react";
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
          <nav className="erp-ds-breadcrumbs" aria-label="Breadcrumb">
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
        <div className="erp-ds-page-actions" aria-label="Page actions">
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

export function FilterBar({ children, actions, label = "Filters" }: FilterBarProps) {
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
    <footer className="erp-ds-sticky-footer" aria-label="Page actions">
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

export function AuditLogPanel({ entries, title = "Audit log" }: AuditLogPanelProps) {
  return (
    <TemplatePanel title={title} description="Latest traceable changes for this record">
      {entries.length > 0 ? (
        <ol className="erp-ds-audit-list">
          {entries.map((entry) => (
            <li className="erp-ds-audit-entry" key={entry.id}>
              <time dateTime={entry.occurredAt}>{entry.occurredAt}</time>
              <strong>{entry.action}</strong>
              <span>{entry.actor}</span>
              {entry.before || entry.after ? (
                <div className="erp-ds-audit-diff">
                  <span>{entry.before ?? "Before: n/a"}</span>
                  <span>{entry.after ?? "After: n/a"}</span>
                </div>
              ) : null}
              {entry.metadata ? <small>{entry.metadata}</small> : null}
            </li>
          ))}
        </ol>
      ) : (
        <p className="erp-ds-panel-empty">No audit events recorded.</p>
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
  action?: ReactNode;
};

export type AttachmentPanelProps = {
  items: readonly AttachmentPanelItem[];
  title?: ReactNode;
  action?: ReactNode;
};

export function AttachmentPanel({ items, title = "Attachments", action }: AttachmentPanelProps) {
  return (
    <TemplatePanel title={title} description="Files linked to this record" actions={action}>
      {items.length > 0 ? (
        <div className="erp-ds-attachment-list">
          {items.map((item) => (
            <article className="erp-ds-attachment-row" key={item.id}>
              <div>
                <strong>{item.name}</strong>
                <span>{item.kind}</span>
              </div>
              <div>
                <span>{item.uploadedBy}</span>
                <time dateTime={item.uploadedAt}>{item.uploadedAt}</time>
              </div>
              {item.action ? <div className="erp-ds-attachment-action">{item.action}</div> : null}
            </article>
          ))}
        </div>
      ) : (
        <p className="erp-ds-panel-empty">No attachments uploaded.</p>
      )}
    </TemplatePanel>
  );
}
