import type { AppMenuItem } from "../permissions/menu";
import { translate, type Locale } from "./index";

const navigationKeyByHref: Record<string, string> = {
  "/dashboard": "dashboard",
  "/alerts": "alerts",
  "/warehouse": "warehouse",
  "/receiving": "receiving",
  "/inventory": "inventory",
  "/purchase": "purchase",
  "/qc": "qc",
  "/production": "production",
  "/subcontract": "subcontract",
  "/sales": "sales",
  "/shipping": "shipping",
  "/returns": "returns",
  "/master-data": "masterData",
  "/sku-batch": "skuBatch",
  "/suppliers": "suppliers",
  "/customers": "customers",
  "/approvals": "approvals",
  "/finance": "finance",
  "/audit-log": "auditLog",
  "/reports": "reports",
  "/settings": "settings"
};

export function getNavigationGroupLabel(label: string, locale?: Locale) {
  return translate(`navigation.groups.${label}`, { locale, fallback: label });
}

export function getNavigationItemLabel(item: Pick<AppMenuItem, "href" | "label">, locale?: Locale) {
  const key = navigationKeyByHref[item.href];
  if (!key) {
    return item.label;
  }

  return translate(`navigation.items.${key}`, { locale, fallback: item.label });
}
