import { describe, expect, it } from "vitest";
import type { FormulaMasterDataItem, ProductMasterDataItem } from "@/modules/masterdata/types";
import {
  applyFormulaToProductionPlanDraftLine,
  applyProductToProductionPlanDraftLine,
  createProductionPlanDraftLine,
  defaultProductionPlanUom,
  findFormulaForProduct,
  formulaBelongsToProduct
} from "./productionPlanFormDefaults";

const product: ProductMasterDataItem = {
  id: "item-aah",
  itemCode: "AAH",
  skuCode: "AAH",
  name: "Kem u phuc hoi As A Habit 150gr",
  itemType: "finished_good",
  itemGroup: "toc",
  brandCode: "MYH",
  uomBase: "JAR",
  uomPurchase: "JAR",
  uomIssue: "JAR",
  lotControlled: true,
  expiryControlled: true,
  shelfLifeDays: 365,
  qcRequired: true,
  status: "active",
  standardCost: "0.000000",
  isSellable: true,
  isPurchasable: false,
  isProducible: true,
  specVersion: "",
  createdAt: "2026-05-04T00:00:00Z",
  updatedAt: "2026-05-04T00:00:00Z"
};

const formula: FormulaMasterDataItem = {
  id: "formula-aah-v1",
  formulaCode: "AAH",
  finishedItemId: "2e2f71b4-a502-43e8-a448-04d875a04cb5",
  finishedSku: "AAH",
  finishedItemName: "Kem u phuc hoi As A Habit 150gr",
  finishedItemType: "finished_good",
  formulaVersion: "v1",
  batchQty: "81.000000",
  batchUomCode: "PCS",
  baseBatchQty: "81.000000",
  baseBatchUomCode: "PCS",
  status: "active",
  approvalStatus: "approved",
  lines: [],
  createdAt: "2026-05-04T00:00:00Z",
  updatedAt: "2026-05-04T00:00:00Z",
  version: 1
};

describe("productionPlanFormDefaults", () => {
  it("matches formulas by finished SKU when product and formula IDs use different references", () => {
    expect(formulaBelongsToProduct(formula, product)).toBe(true);
    expect(findFormulaForProduct([formula], product)?.id).toBe(formula.id);
  });

  it("uses the selected formula batch UOM instead of the product base UOM", () => {
    expect(defaultProductionPlanUom(product, formula)).toBe("PCS");
  });

  it("creates a draft line from the product and its active formula", () => {
    expect(createProductionPlanDraftLine("line-1", product, [formula])).toMatchObject({
      rowId: "line-1",
      outputItemId: "item-aah",
      formulaId: "formula-aah-v1",
      plannedQty: "1.000000",
      uomCode: "PCS"
    });
  });

  it("updates product and formula fields without losing quantity and dates", () => {
    const current = createProductionPlanDraftLine("line-1", undefined, []);
    const withProduct = applyProductToProductionPlanDraftLine(
      {
        ...current,
        plannedQty: "25.000000",
        plannedStartDate: "2026-05-05",
        plannedEndDate: "2026-05-06"
      },
      product,
      [formula]
    );

    expect(withProduct).toMatchObject({
      plannedQty: "25.000000",
      plannedStartDate: "2026-05-05",
      plannedEndDate: "2026-05-06",
      outputItemId: "item-aah",
      formulaId: "formula-aah-v1",
      uomCode: "PCS"
    });
    expect(applyFormulaToProductionPlanDraftLine(withProduct, product, undefined).uomCode).toBe("JAR");
  });
});
