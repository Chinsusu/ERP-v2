import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  calculateFormulaRequirement,
  createFormula,
  emptyFormulaInput,
  formulaInputForParentItem,
  formatFormulaQuantity,
  getFormulas,
  normalizeFormulaInput,
  resetPrototypeFormulaMasterData,
  summarizeFormulas
} from "./formulaMasterDataService";

describe("formulaMasterDataService", () => {
  beforeEach(() => {
    resetPrototypeFormulaMasterData();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("normalizes Vietnamese decimal quantities without losing six-decimal formula amounts", () => {
    const normalized = normalizeFormulaInput({
      ...emptyFormulaInput,
      formulaCode: "xff-150ml",
      finishedItemId: "item-xff-150",
      finishedSku: "xff",
      finishedItemName: "Tính chất bưởi Fast & Furious 150ML",
      lines: [
        {
          ...emptyFormulaInput.lines[0],
          componentSku: "ACT_TRACE",
          componentName: "Trace active",
          enteredQty: "0,000001",
          enteredUomCode: "KG",
          calcQty: "1",
          calcUomCode: "MG",
          stockBaseQty: "0,000001",
          stockBaseUomCode: "KG"
        }
      ]
    });

    expect(normalized.formulaCode).toBe("XFF-150ML");
    expect(normalized.finishedSku).toBe("XFF");
    expect(normalized.lines[0].enteredQty).toBe("0.000001");
    expect(normalized.lines[0].stockBaseQty).toBe("0.000001");
  });

  it("formats tiny mass quantities in mg, g, or kg for readable formula review", () => {
    expect(formatFormulaQuantity("0.000001", "KG")).toBe("1 mg");
    expect(formatFormulaQuantity("0.001000", "KG")).toBe("1 g");
    expect(formatFormulaQuantity("1.250000", "KG")).toBe("1,25 kg");
  });

  it("binds formula parent fields from a finished or semi-finished SKU instead of free text", () => {
    const input = formulaInputForParentItem(emptyFormulaInput, {
      id: "item-grn",
      skuCode: "GRN",
      name: "Dáº¦U Gá»˜I RETRO NANO 350ML",
      itemType: "finished_good",
      status: "active"
    });

    expect(input).toMatchObject({
      finishedItemId: "item-grn",
      finishedSku: "GRN",
      finishedItemName: "Dáº¦U Gá»˜I RETRO NANO 350ML",
      finishedItemType: "finished_good"
    });
  });

  it("creates formulas and calculates planned requirements in the local fallback store", async () => {
    const formula = await createFormula({
      ...emptyFormulaInput,
      formulaCode: "XFF-150ML",
      finishedItemId: "item-xff-150",
      finishedSku: "XFF",
      finishedItemName: "Tính chất bưởi Fast & Furious 150ML",
      batchQty: "81",
      baseBatchQty: "999",
      lines: [
        {
          ...emptyFormulaInput.lines[0],
          componentItemId: "item-act-baicapil",
          componentSku: "ACT_BAICAPIL",
          componentName: "BAICAPIL",
          enteredQty: "0,001",
          enteredUomCode: "KG",
          calcQty: "1",
          calcUomCode: "G",
          stockBaseQty: "0,001",
          stockBaseUomCode: "KG"
        }
      ]
    });

    await expect(getFormulas({ search: "xff" })).resolves.toHaveLength(1);
    await expect(calculateFormulaRequirement(formula.id, { plannedQty: "162", plannedUomCode: "PCS" })).resolves.toMatchObject({
      requirements: [
        {
          componentSku: "ACT_BAICAPIL",
          requiredStockBaseQty: "0.162000",
          stockBaseUomCode: "KG"
        }
      ]
    });
    expect(summarizeFormulas(await getFormulas())).toEqual({
      total: 1,
      active: 0,
      draft: 1,
      lines: 1
    });
  });
});
