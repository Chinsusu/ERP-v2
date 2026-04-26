"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";
import type { MockUser } from "@/shared/auth/mockSession";
import { getVisibleMenuGroups } from "@/shared/permissions/menu";

type AppShellProps = {
  children: ReactNode;
  user: MockUser;
};

function isActivePath(pathname: string, href: string) {
  return pathname === href || pathname.startsWith(`${href}/`);
}

export function AppShell({ children, user }: AppShellProps) {
  const pathname = usePathname();
  const groups = getVisibleMenuGroups(user);

  return (
    <div className="erp-shell">
      <header className="erp-topbar">
        <Link className="erp-brand" href="/dashboard" aria-label="ERP dashboard">
          <span className="erp-brand-mark" aria-hidden="true">
            ERP
          </span>
          <span className="erp-brand-text">ERP Platform</span>
        </Link>

        <label className="erp-global-search">
          <span className="erp-sr-only">Global search</span>
          <input
            className="erp-global-search-input"
            type="search"
            placeholder="Search order, SKU, batch, receipt, tracking code..."
          />
        </label>

        <div className="erp-topbar-actions" aria-label="Quick actions">
          <button className="erp-button erp-button--primary" type="button">
            Quick create
          </button>
          <button className="erp-button erp-button--secondary" type="button">
            Alerts
          </button>
          <button className="erp-button erp-button--secondary" type="button">
            Docs
          </button>
          <div className="erp-user-badge" aria-label={`Signed in as ${user.name}`}>
            <span className="erp-user-avatar" aria-hidden="true">
              {user.name.slice(0, 2).toUpperCase()}
            </span>
            <span>
              <strong>{user.name}</strong>
              <small>{user.role}</small>
            </span>
          </div>
        </div>
      </header>

      <div className="erp-shell-body">
        <aside className="erp-sidebar" aria-label="ERP navigation">
          {groups.map((group) => (
            <nav
              className="erp-sidebar-group"
              key={group.label}
              aria-labelledby={`menu-${group.label}`}
            >
              <h2 className="erp-sidebar-heading" id={`menu-${group.label}`}>
                {group.label}
              </h2>
              <ul className="erp-sidebar-list">
                {group.items.map((item) => {
                  const active = isActivePath(pathname, item.href);

                  return (
                    <li key={item.href}>
                      <Link
                        className="erp-sidebar-link"
                        data-active={active ? "true" : "false"}
                        href={item.href}
                        aria-current={active ? "page" : undefined}
                      >
                        <span className="erp-sidebar-code" aria-hidden="true">
                          {item.code}
                        </span>
                        <span>{item.label}</span>
                      </Link>
                    </li>
                  );
                })}
              </ul>
            </nav>
          ))}
        </aside>

        <main className="erp-shell-content">{children}</main>
      </div>
    </div>
  );
}
