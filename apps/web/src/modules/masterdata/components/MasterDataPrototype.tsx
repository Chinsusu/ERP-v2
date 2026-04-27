"use client";

import { useState } from "react";
import { StatusChip } from "@/shared/design-system/components";
import { ProductMasterDataPrototype } from "./ProductMasterDataPrototype";
import { WarehouseLocationMasterDataPrototype } from "./WarehouseLocationMasterDataPrototype";

type MasterDataView = "products" | "warehouses";

const masterDataViews: { label: string; value: MasterDataView }[] = [
  { label: "Items / SKU", value: "products" },
  { label: "Warehouses / Locations", value: "warehouses" }
];

export function MasterDataPrototype() {
  const [view, setView] = useState<MasterDataView>("products");

  return (
    <section className="erp-module-page erp-masterdata-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">MD</p>
          <h1 className="erp-page-title">Master Data</h1>
          <p className="erp-page-description">Operational master data for SKU, warehouse, location, and inventory setup</p>
        </div>
        <StatusChip tone="info">{view === "products" ? "Item setup" : "Warehouse setup"}</StatusChip>
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

      {view === "products" ? <ProductMasterDataPrototype embedded /> : <WarehouseLocationMasterDataPrototype embedded />}
    </section>
  );
}
