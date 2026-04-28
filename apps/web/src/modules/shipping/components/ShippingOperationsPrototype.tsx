"use client";

import { useState } from "react";
import { CarrierManifestPrototype } from "./CarrierManifestPrototype";
import { PickingPrototype } from "./PickingPrototype";

type ShippingView = "picking" | "handover";

export function ShippingOperationsPrototype() {
  const [view, setView] = useState<ShippingView>("picking");

  return (
    <section className="erp-shipping-operations">
      <nav className="erp-shipping-tabs" aria-label="Shipping operation views">
        <button
          className={`erp-shipping-tab ${view === "picking" ? "erp-shipping-tab--active" : ""}`}
          type="button"
          onClick={() => setView("picking")}
        >
          Picking
        </button>
        <button
          className={`erp-shipping-tab ${view === "handover" ? "erp-shipping-tab--active" : ""}`}
          type="button"
          onClick={() => setView("handover")}
        >
          Carrier handover
        </button>
      </nav>
      {view === "picking" ? <PickingPrototype /> : <CarrierManifestPrototype />}
    </section>
  );
}
