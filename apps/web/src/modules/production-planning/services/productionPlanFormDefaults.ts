import type { FormulaMasterDataItem, ProductMasterDataItem } from "@/modules/masterdata/types";

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
