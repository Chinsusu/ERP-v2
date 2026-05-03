"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, type ReactNode } from "react";
import {
  clearClientAccessToken,
  rememberClientAccessToken
} from "../auth/clientSessionToken";
import type { AuthenticatedUser } from "../auth/session";
import { supportedLocales, type Locale } from "@/shared/i18n/config";
import { getActionLabel } from "@/shared/i18n/action-labels";
import { t } from "@/shared/i18n";
import { getNavigationGroupLabel, getNavigationItemLabel } from "@/shared/i18n/navigation-labels";
import { useLocale } from "@/shared/i18n/useLocale";
import { getVisibleActions, getVisibleMenuGroups, topbarActions } from "@/shared/permissions/menu";

type AppShellProps = {
  accessToken: string;
  children: ReactNode;
  expiresAt: string;
  signOutAction: () => Promise<void>;
  user: AuthenticatedUser;
};

const logoutSignalKey = "erp_auth_logout_at";

function isActivePath(pathname: string, href: string) {
  return pathname === href || pathname.startsWith(`${href}/`);
}

type RefreshResponse = {
  data?: {
    access_token?: string;
    expires_at?: string;
  };
};

export function AppShell({ accessToken, children, expiresAt, signOutAction, user }: AppShellProps) {
  rememberClientAccessToken(accessToken);

  const pathname = usePathname();
  const [locale, setLocale] = useLocale();
  const groups = getVisibleMenuGroups(user);
  const actions = getVisibleActions(user, topbarActions);

  useEffect(() => {
    const handleStorage = (event: StorageEvent) => {
      if (event.key === logoutSignalKey) {
        clearClientAccessToken();
        window.location.assign("/login");
      }
    };

    window.addEventListener("storage", handleStorage);
    return () => window.removeEventListener("storage", handleStorage);
  }, []);

  useEffect(() => {
    let stopped = false;
    let refreshTimer: number | undefined;

    const scheduleRefresh = (targetExpiresAt: string) => {
      const expiresAtMs = Date.parse(targetExpiresAt);
      if (!Number.isFinite(expiresAtMs)) {
        return;
      }

      const refreshDelayMs = Math.max(expiresAtMs - Date.now() - 60_000, 0);
      refreshTimer = window.setTimeout(async () => {
        try {
          const response = await fetch("/api/auth/refresh", { method: "POST" });
          if (stopped) {
            return;
          }
          if (!response.ok) {
            throw new Error("session refresh failed");
          }

          const payload = (await response.json()) as RefreshResponse;
          const nextAccessToken = payload.data?.access_token;
          const nextExpiresAt = payload.data?.expires_at;
          if (!nextAccessToken || !nextExpiresAt) {
            throw new Error("session refresh response is incomplete");
          }

          rememberClientAccessToken(nextAccessToken);
          scheduleRefresh(nextExpiresAt);
        } catch {
          if (!stopped) {
            clearClientAccessToken();
            publishLogoutSignal();
            window.location.assign("/login?error=session_expired");
          }
        }
      }, refreshDelayMs);
    };

    rememberClientAccessToken(accessToken);
    scheduleRefresh(expiresAt);

    return () => {
      stopped = true;
      if (refreshTimer !== undefined) {
        window.clearTimeout(refreshTimer);
      }
    };
  }, [accessToken, expiresAt]);

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
          <form action={signOutAction} onSubmit={publishLogoutSignal}>
            <button className="erp-button erp-button--secondary" type="submit">
              {t("auth.signOut")}
            </button>
          </form>
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

function publishLogoutSignal() {
  clearClientAccessToken();
  try {
    window.localStorage.setItem(logoutSignalKey, String(Date.now()));
  } catch {
    // The server logout action still clears the httpOnly cookies for this tab.
  }
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
