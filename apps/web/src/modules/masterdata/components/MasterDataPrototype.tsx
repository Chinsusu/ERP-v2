"use client";

import { useState } from "react";
import { StatusChip } from "@/shared/design-system/components";
import { t } from "@/shared/i18n";
import { ProductMasterDataPrototype } from "./ProductMasterDataPrototype";
import { SupplierCustomerMasterDataPrototype } from "./SupplierCustomerMasterDataPrototype";
import { WarehouseLocationMasterDataPrototype } from "./WarehouseLocationMasterDataPrototype";
import { getMasterDataTabs, getMasterDataViewStatusLabel, type MasterDataView } from "./masterDataTabs";

export function MasterDataPrototype() {
  const [view, setView] = useState<MasterDataView>("products");
  const masterDataViews = getMasterDataTabs();

  return (
    <section className="erp-module-page erp-masterdata-page">
      <header className="erp-page-header">
        <div>
          <h1 className="erp-page-title">{masterDataCopy("title")}</h1>
          <p className="erp-page-description">{masterDataCopy("description")}</p>
        </div>
        <StatusChip tone="info">{getMasterDataViewStatusLabel(view)}</StatusChip>
      </header>

      <nav className="erp-masterdata-tabs" aria-label={masterDataCopy("sectionsLabel")}>
        {masterDataViews.map((item) => (
          <button
            className="erp-masterdata-tab"
            data-active={view === item.value ? "true" : "false"}
            key={item.value}
            type="button"
            onClick={() => setView(item.value)}
          >
            {item.label}
          </button>
        ))}
      </nav>

      {view === "products" ? <ProductMasterDataPrototype embedded /> : null}
      {view === "warehouses" ? <WarehouseLocationMasterDataPrototype embedded /> : null}
      {view === "suppliers" ? <SupplierCustomerMasterDataPrototype embedded mode="suppliers" /> : null}
      {view === "customers" ? <SupplierCustomerMasterDataPrototype embedded mode="customers" /> : null}
    </section>
  );
}

function masterDataCopy(key: string, values?: Record<string, string | number>, fallback?: string) {
  return t(`masterdata.${key}`, { values, fallback });
}
