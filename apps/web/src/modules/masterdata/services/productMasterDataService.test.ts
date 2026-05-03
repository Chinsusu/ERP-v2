import { beforeEach, describe, expect, it } from "vitest";
import {
  changeProductStatus,
  createProduct,
  emptyProductInput,
  getProducts,
  productUomOptions,
  productStatusTone,
  productTypeLabel,
  resetPrototypeProductMasterData,
  summarizeProducts,
  updateProduct
} from "./productMasterDataService";

describe("productMasterDataService", () => {
  beforeEach(() => {
    resetPrototypeProductMasterData();
  });

  it("filters product master data by search, status, and type", async () => {
    await expect(
      getProducts({
        search: "citric",
        status: "active",
        itemType: "raw_material"
      })
    ).resolves.toMatchObject([
      {
        skuCode: "ACI_CITRIC",
        status: "active",
        itemType: "raw_material"
      }
    ]);
  });

  it("uses the normalized source sheet items in the local fallback store", async () => {
    const items = await getProducts();

    expect(items).toHaveLength(332);
    expect(items.some((item) => ["SERUM-30ML", "CREAM-50G", "TONER-100ML"].includes(item.skuCode))).toBe(false);
    expect(items).toContainEqual(
      expect.objectContaining({
        skuCode: "FRA_NTG",
        itemType: "raw_material",
        itemGroup: "fragrance",
        uomBase: "KG"
      })
    );
    expect(items).toContainEqual(
      expect.objectContaining({
        skuCode: "TP-100",
        itemType: "packaging",
        itemGroup: "tube",
        uomBase: "TUBE"
      })
    );
    expect(summarizeProducts(items)).toEqual({
      total: 332,
      active: 332,
      draft: 0,
      controlled: 332
    });
  });

  it("uses PCS as the clean default unit for new SKU forms", () => {
    expect(emptyProductInput.uomBase).toBe("PCS");
    expect(emptyProductInput.uomPurchase).toBe("PCS");
    expect(emptyProductInput.uomIssue).toBe("PCS");
  });

  it("offers the standard UOM catalog for SKU unit fields", () => {
    expect(productUomOptions.map((option) => option.value)).toEqual([
      "MG",
      "G",
      "KG",
      "ML",
      "L",
      "PCS",
      "BOTTLE",
      "JAR",
      "TUBE",
      "BOX",
      "CARTON",
      "SET",
      "BAG",
      "ROLL",
      "CM",
      "SERVICE"
    ]);
  });

  it("creates, updates, and changes status in the local fallback store", async () => {
    const created = await createProduct({
      ...emptyProductInput,
      itemCode: "item-mask-set",
      skuCode: "mask-set-05",
      name: "Sheet Mask Set",
      itemGroup: "mask",
      specVersion: "SPEC-MASK-2026.04"
    });

    expect(created).toMatchObject({
      itemCode: "ITEM-MASK-SET",
      skuCode: "MASK-SET-05",
      status: "draft"
    });
    expect(created.auditLogId).toContain("audit-local-create");

    const updated = await updateProduct(created.id, {
      ...emptyProductInput,
      itemCode: "ITEM-MASK-SET",
      skuCode: "MASK-SET-05",
      name: "Sheet Mask Set 5pcs",
      itemGroup: "mask"
    });
    expect(updated.name).toBe("Sheet Mask Set 5pcs");

    const active = await changeProductStatus(created.id, "active");
    expect(active.status).toBe("active");
  });

  it("blocks duplicate item and SKU codes in the local fallback store", async () => {
    await expect(
      createProduct({
        ...emptyProductInput,
        itemCode: "ACI_CITRIC",
        skuCode: "NEW-SERUM-30ML",
        name: "Duplicate Serum"
      })
    ).rejects.toThrow("Item code already exists");

    await expect(
      createProduct({
        ...emptyProductInput,
        itemCode: "ITEM-NEW-SERUM",
        skuCode: "ACI_CITRIC",
        name: "Duplicate Serum"
      })
    ).rejects.toThrow("SKU code already exists");
  });

  it("maps statuses and item types to display labels", () => {
    expect(productStatusTone("active")).toBe("success");
    expect(productStatusTone("obsolete")).toBe("danger");
    expect(productTypeLabel("semi_finished")).toBe("Semi finished");
  });
});
