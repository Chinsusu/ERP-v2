"use client";

import { useMemo, useState } from "react";
import { EmptyState, StatusChip } from "@/shared/design-system/components";
import { t } from "@/shared/i18n";
import { FormulaMasterDataPrototype } from "./FormulaMasterDataPrototype";
import { ProductMasterDataPrototype } from "./ProductMasterDataPrototype";
import type { ProductMasterDataItem } from "../types";

export function FinishedProductMasterDataPrototype() {
  const [items, setItems] = useState<ProductMasterDataItem[]>([]);
  const [selectedItemId, setSelectedItemId] = useState<string>("");
  const selectedItem = useMemo(() => items.find((item) => item.id === selectedItemId), [items, selectedItemId]);
  const activeParents = useMemo(
    () => items.filter((item) => item.status === "active" && (item.itemType === "finished_good" || item.itemType === "semi_finished")),
    [items]
  );

  return (
    <section className="erp-masterdata-finished-flow">
      <ProductMasterDataPrototype
        embedded
        mode="finished"
        activeFormulaItemId={selectedItemId}
        onItemsChange={setItems}
        onOpenFormula={(item) => setSelectedItemId(item.id)}
      />

      <section className="erp-masterdata-formula-panel">
        <div className="erp-section-header">
          <div>
            <h2 className="erp-section-title">{copy("finished.formulaTitle")}</h2>
            <p className="erp-page-description">
              {selectedItem ? copy("finished.formulaDescription", { sku: selectedItem.skuCode }) : copy("finished.formulaEmptyDescription")}
            </p>
          </div>
          <StatusChip tone={selectedItem ? "info" : "warning"}>
            {selectedItem ? selectedItem.skuCode : copy("finished.noSelection")}
          </StatusChip>
        </div>

        {selectedItem ? (
          <FormulaMasterDataPrototype parentItems={activeParents} selectedParentItemId={selectedItem.id} />
        ) : (
          <EmptyState title={copy("finished.emptyTitle")} description={copy("finished.emptyDescription")} />
        )}
      </section>
    </section>
  );
}

function copy(key: string, values?: Record<string, string | number>) {
  return t(`masterdata.${key}`, { values });
}
