import { t } from "../../../shared/i18n";

export type MasterDataView = "finishedProducts" | "materials" | "warehouses" | "suppliers" | "customers";

type MasterDataTab = {
  label: string;
  value: MasterDataView;
};

export function getMasterDataTabs(): MasterDataTab[] {
  return [
    { label: masterDataCopy("views.finishedProducts"), value: "finishedProducts" },
    { label: masterDataCopy("views.materials"), value: "materials" },
    { label: masterDataCopy("views.warehouses"), value: "warehouses" },
    { label: masterDataCopy("views.suppliers"), value: "suppliers" },
    { label: masterDataCopy("views.customers"), value: "customers" }
  ];
}

export function getMasterDataViewStatusLabel(view: MasterDataView) {
  return masterDataCopy(`viewStatus.${view}`);
}

function masterDataCopy(key: string, values?: Record<string, string | number>, fallback?: string) {
  return t(`masterdata.${key}`, { values, fallback });
}
