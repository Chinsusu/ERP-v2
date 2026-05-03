"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";
import type { MockUser } from "@/shared/auth/mockSession";
import { supportedLocales, type Locale } from "@/shared/i18n/config";
import { getActionLabel } from "@/shared/i18n/action-labels";
import { t } from "@/shared/i18n";
import { getNavigationGroupLabel, getNavigationItemLabel } from "@/shared/i18n/navigation-labels";
import { useLocale } from "@/shared/i18n/useLocale";
import { getVisibleActions, getVisibleMenuGroups, topbarActions } from "@/shared/permissions/menu";

type AppShellProps = {
  children: ReactNode;
  user: MockUser;
};

function isActivePath(pathname: string, href: string) {
  return pathname === href || pathname.startsWith(`${href}/`);
}

export function AppShell({ children, user }: AppShellProps) {
  const pathname = usePathname();
  const [locale, setLocale] = useLocale();
  const groups = getVisibleMenuGroups(user);
  const actions = getVisibleActions(user, topbarActions);

  return (
    <div className="erp-shell">
      <header className="erp-topbar">
        <Link className="erp-brand" href="/dashboard" aria-label={t("navigation.items.dashboard")}>
          <span className="erp-brand-mark" aria-hidden="true">
            ERP
          </span>
          <span className="erp-brand-text">{t("common.appName")}</span>
        </Link>

        <label className="erp-global-search">
          <span className="erp-sr-only">{t("common.globalSearch")}</span>
          <input
            className="erp-global-search-input"
            type="search"
            placeholder={t("common.searchPlaceholder")}
          />
        </label>

        <div className="erp-topbar-actions" aria-label={t("common.quickActions")}>
          {actions.map((action) => (
            <button
              className={`erp-button erp-button--${action.variant}`}
              key={action.label}
              type="button"
            >
              {getActionLabel(action.label)}
            </button>
          ))}
          <LanguageSwitch locale={locale} onLocaleChange={setLocale} />
          <div className="erp-user-badge" aria-label={t("common.signedInAs", { values: { name: user.name } })}>
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
        <aside className="erp-sidebar" aria-label={t("common.navigation")}>
          {groups.map((group) => (
            <nav
              className="erp-sidebar-group"
              key={group.label}
              aria-labelledby={`menu-${group.label}`}
            >
              <h2 className="erp-sidebar-heading" id={`menu-${group.label}`}>
                {getNavigationGroupLabel(group.label)}
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
                        <span>{getNavigationItemLabel(item)}</span>
                      </Link>
                    </li>
                  );
                })}
              </ul>
            </nav>
          ))}
        </aside>

        <main className="erp-shell-content" key={locale}>{children}</main>
      </div>
    </div>
  );
}

type LanguageSwitchProps = {
  locale: Locale;
  onLocaleChange: (locale: Locale) => void;
};

function LanguageSwitch({ locale, onLocaleChange }: LanguageSwitchProps) {
  return (
    <div className="erp-language-switch" role="group" aria-label={t("common.language")}>
      {supportedLocales.map((candidate) => {
        const active = candidate === locale;

        return (
          <button
            aria-pressed={active}
            className="erp-language-option"
            data-active={active ? "true" : "false"}
            key={candidate}
            onClick={() => onLocaleChange(candidate)}
            title={t(`common.locales.${candidate}`)}
            type="button"
          >
            {candidate.toUpperCase()}
          </button>
        );
      })}
    </div>
  );
}
