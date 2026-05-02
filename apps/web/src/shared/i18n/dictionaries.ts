import enActions from "./locales/en/actions.json";
import enAudit from "./locales/en/audit.json";
import enAuth from "./locales/en/auth.json";
import enCommon from "./locales/en/common.json";
import enErrors from "./locales/en/errors.json";
import enFinance from "./locales/en/finance.json";
import enInventory from "./locales/en/inventory.json";
import enMasterdata from "./locales/en/masterdata.json";
import enNavigation from "./locales/en/navigation.json";
import enPurchase from "./locales/en/purchase.json";
import enQc from "./locales/en/qc.json";
import enReturns from "./locales/en/returns.json";
import enSales from "./locales/en/sales.json";
import enSettings from "./locales/en/settings.json";
import enShipping from "./locales/en/shipping.json";
import enStatus from "./locales/en/status.json";
import enUnits from "./locales/en/units.json";
import enValidation from "./locales/en/validation.json";
import enWarehouse from "./locales/en/warehouse.json";
import viActions from "./locales/vi/actions.json";
import viAudit from "./locales/vi/audit.json";
import viAuth from "./locales/vi/auth.json";
import viCommon from "./locales/vi/common.json";
import viErrors from "./locales/vi/errors.json";
import viFinance from "./locales/vi/finance.json";
import viInventory from "./locales/vi/inventory.json";
import viMasterdata from "./locales/vi/masterdata.json";
import viNavigation from "./locales/vi/navigation.json";
import viPurchase from "./locales/vi/purchase.json";
import viQc from "./locales/vi/qc.json";
import viReturns from "./locales/vi/returns.json";
import viSales from "./locales/vi/sales.json";
import viSettings from "./locales/vi/settings.json";
import viShipping from "./locales/vi/shipping.json";
import viStatus from "./locales/vi/status.json";
import viUnits from "./locales/vi/units.json";
import viValidation from "./locales/vi/validation.json";
import viWarehouse from "./locales/vi/warehouse.json";
import type { Locale } from "./config";

export const dictionaryNamespaces = [
  "common",
  "navigation",
  "actions",
  "status",
  "errors",
  "validation",
  "masterdata",
  "inventory",
  "warehouse",
  "sales",
  "shipping",
  "returns",
  "purchase",
  "qc",
  "finance",
  "auth",
  "audit",
  "settings",
  "units"
] as const;

export type DictionaryNamespace = (typeof dictionaryNamespaces)[number];
export type DictionaryTree = Record<string, unknown>;
export type LocaleDictionary = Record<DictionaryNamespace, DictionaryTree>;

export const dictionaries: Record<Locale, LocaleDictionary> = {
  vi: {
    common: viCommon,
    navigation: viNavigation,
    actions: viActions,
    status: viStatus,
    errors: viErrors,
    validation: viValidation,
    masterdata: viMasterdata,
    inventory: viInventory,
    warehouse: viWarehouse,
    sales: viSales,
    shipping: viShipping,
    returns: viReturns,
    purchase: viPurchase,
    qc: viQc,
    finance: viFinance,
    auth: viAuth,
    audit: viAudit,
    settings: viSettings,
    units: viUnits
  },
  en: {
    common: enCommon,
    navigation: enNavigation,
    actions: enActions,
    status: enStatus,
    errors: enErrors,
    validation: enValidation,
    masterdata: enMasterdata,
    inventory: enInventory,
    warehouse: enWarehouse,
    sales: enSales,
    shipping: enShipping,
    returns: enReturns,
    purchase: enPurchase,
    qc: enQc,
    finance: enFinance,
    auth: enAuth,
    audit: enAudit,
    settings: enSettings,
    units: enUnits
  }
};
