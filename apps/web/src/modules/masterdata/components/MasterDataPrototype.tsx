"use client";

import { useState } from "react";
import { StatusChip } from "@/shared/design-system/components";
import { t } from "@/shared/i18n";
import { ProductMasterDataPrototype } from "./ProductMasterDataPrototype";
import { SupplierCustomerMasterDataPrototype } from "./SupplierCustomerMasterDataPrototype";
import { WarehouseLocationMasterDataPrototype } from "./WarehouseLocationMasterDataPrototype";

type MasterDataView = "products" | "warehouses" | "parties";

const masterDataViews: { label: string; value: MasterDataView }[] = [
  { label: masterDataCopy("views.products"), value: "products" },
  { label: masterDataCopy("views.warehouses"), value: "warehouses" },
  { label: masterDataCopy("views.parties"), value: "parties" }
];

export function MasterDataPrototype() {
  const [view, setView] = useState<MasterDataView>("products");

  return (
    <section className="erp-module-page erp-masterdata-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">MD</p>
          <h1 className="erp-page-title">{masterDataCopy("title")}</h1>
          <p className="erp-page-description">{masterDataCopy("description")}</p>
        </div>
        <StatusChip tone="info">{viewLabel(view)}</StatusChip>
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
      {view === "parties" ? <SupplierCustomerMasterDataPrototype embedded /> : null}
    </section>
  );
}

function viewLabel(view: MasterDataView) {
  switch (view) {
    case "products":
      return masterDataCopy("viewStatus.products");
    case "warehouses":
      return masterDataCopy("viewStatus.warehouses");
    case "parties":
      return masterDataCopy("viewStatus.parties");
    default:
      return masterDataCopy("title");
  }
}

function masterDataCopy(key: string, values?: Record<string, string | number>, fallback?: string) {
  return t(`masterdata.${key}`, { values, fallback });
}
