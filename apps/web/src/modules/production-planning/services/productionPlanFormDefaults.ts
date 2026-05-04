import type { FormulaMasterDataItem, ProductMasterDataItem } from "@/modules/masterdata/types";
import type { ProductionPlanDraftLine } from "../types";

const defaultDraftQty = "1";

export function formulaBelongsToProduct(formula: FormulaMasterDataItem, product?: ProductMasterDataItem) {
  if (!product) {
    return false;
  }

  return formula.finishedItemId === product.id || formula.finishedSku === product.skuCode;
}

export function findFormulaForProduct(formulas: FormulaMasterDataItem[], product?: ProductMasterDataItem) {
  return formulas.find((formula) => formulaBelongsToProduct(formula, product));
}

export function defaultProductionPlanUom(product?: ProductMasterDataItem, formula?: FormulaMasterDataItem) {
  return formula?.batchUomCode || product?.uomBase || "";
}

export function createProductionPlanDraftLine(
  rowId: string,
  product?: ProductMasterDataItem,
  activeFormulas: FormulaMasterDataItem[] = []
): ProductionPlanDraftLine {
  const formula = findFormulaForProduct(activeFormulas, product);

  return {
    rowId,
    outputItemId: product?.id ?? "",
    formulaId: formula?.id ?? "",
    plannedQty: defaultDraftQty,
    uomCode: defaultProductionPlanUom(product, formula) || "PCS",
    plannedStartDate: "",
    plannedEndDate: ""
  };
}

export function applyProductToProductionPlanDraftLine(
  line: ProductionPlanDraftLine,
  product: ProductMasterDataItem | undefined,
  activeFormulas: FormulaMasterDataItem[]
): ProductionPlanDraftLine {
  const formula = findFormulaForProduct(activeFormulas, product);

  return {
    ...line,
    outputItemId: product?.id ?? "",
    formulaId: formula?.id ?? "",
    uomCode: defaultProductionPlanUom(product, formula) || line.uomCode
  };
}

export function applyFormulaToProductionPlanDraftLine(
  line: ProductionPlanDraftLine,
  product: ProductMasterDataItem | undefined,
  formula: FormulaMasterDataItem | undefined
): ProductionPlanDraftLine {
  return {
    ...line,
    formulaId: formula?.id ?? "",
    uomCode: defaultProductionPlanUom(product, formula) || line.uomCode
  };
}
