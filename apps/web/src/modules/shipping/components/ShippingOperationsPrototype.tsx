"use client";

import { useState } from "react";
import { t } from "@/shared/i18n";
import { CarrierManifestPrototype } from "./CarrierManifestPrototype";
import { PackingPrototype } from "./PackingPrototype";
import { PickingPrototype } from "./PickingPrototype";

type ShippingView = "picking" | "packing" | "handover";

export function ShippingOperationsPrototype() {
  const [view, setView] = useState<ShippingView>("picking");

  return (
    <section className="erp-shipping-operations">
      <nav className="erp-shipping-tabs" aria-label={t("shipping.operations.label")}>
        <button
          className={`erp-shipping-tab ${view === "picking" ? "erp-shipping-tab--active" : ""}`}
          type="button"
          onClick={() => setView("picking")}
        >
          {t("shipping.operations.tabs.picking")}
        </button>
        <button
          className={`erp-shipping-tab ${view === "packing" ? "erp-shipping-tab--active" : ""}`}
          type="button"
          onClick={() => setView("packing")}
        >
          {t("shipping.operations.tabs.packing")}
        </button>
        <button
          className={`erp-shipping-tab ${view === "handover" ? "erp-shipping-tab--active" : ""}`}
          type="button"
          onClick={() => setView("handover")}
        >
          {t("shipping.operations.tabs.handover")}
        </button>
      </nav>
      {view === "picking" ? <PickingPrototype /> : null}
      {view === "packing" ? <PackingPrototype /> : null}
      {view === "handover" ? <CarrierManifestPrototype /> : null}
    </section>
  );
}
