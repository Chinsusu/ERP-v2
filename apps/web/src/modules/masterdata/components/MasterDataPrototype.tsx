"use client";

import { useState } from "react";
import { StatusChip } from "@/shared/design-system/components";
import { ProductMasterDataPrototype } from "./ProductMasterDataPrototype";
import { SupplierCustomerMasterDataPrototype } from "./SupplierCustomerMasterDataPrototype";
import { WarehouseLocationMasterDataPrototype } from "./WarehouseLocationMasterDataPrototype";

type MasterDataView = "products" | "warehouses" | "parties";

const masterDataViews: { label: string; value: MasterDataView }[] = [
  { label: "Items / SKU", value: "products" },
  { label: "Warehouses / Locations", value: "warehouses" },
  { label: "Suppliers / Customers", value: "parties" }
];

export function MasterDataPrototype() {
  const [view, setView] = useState<MasterDataView>("products");

  return (
    <section className="erp-module-page erp-masterdata-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">MD</p>
          <h1 className="erp-page-title">Master Data</h1>
          <p className="erp-page-description">Operational master data for SKU, warehouse, location, supplier, and customer setup</p>
        </div>
        <StatusChip tone="info">{viewLabel(view)}</StatusChip>
      </header>

      <nav className="erp-masterdata-tabs" aria-label="Master data sections">
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
      return "Item setup";
    case "warehouses":
      return "Warehouse setup";
    case "parties":
      return "Party setup";
    default:
      return "Master data";
  }
}
