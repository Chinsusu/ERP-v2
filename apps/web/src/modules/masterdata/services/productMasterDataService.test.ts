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
        search: "serum",
        status: "active",
        itemType: "finished_good"
      })
    ).resolves.toMatchObject([
      {
        skuCode: "SERUM-30ML",
        status: "active",
        itemType: "finished_good"
      }
    ]);
  });

  it("summarizes active, draft, and controlled SKU counts", async () => {
    const items = await getProducts();

    expect(summarizeProducts(items)).toEqual({
      total: 3,
      active: 2,
      draft: 1,
      controlled: 3
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
        itemCode: "ITEM-SERUM-HYDRA",
        skuCode: "NEW-SERUM-30ML",
        name: "Duplicate Serum"
      })
    ).rejects.toThrow("Item code already exists");

    await expect(
      createProduct({
        ...emptyProductInput,
        itemCode: "ITEM-NEW-SERUM",
        skuCode: "SERUM-30ML",
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
