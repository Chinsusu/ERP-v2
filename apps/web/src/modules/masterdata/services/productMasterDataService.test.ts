import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
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

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("loads every API page when the product catalog is larger than one backend page", async () => {
    const firstPage = Array.from({ length: 100 }, (_, index) => apiProduct(`API-PAGE-1-${index + 1}`));
    const secondPage = [apiProduct("API-PAGE-2-1"), apiProduct("API-PAGE-2-2")];
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(apiPage(firstPage))
      .mockResolvedValueOnce(apiPage(secondPage));
    vi.stubGlobal("fetch", fetchMock);

    const items = await getProducts();

    expect(items).toHaveLength(102);
    expect(items.at(-1)?.skuCode).toBe("API-PAGE-2-2");
    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(String(fetchMock.mock.calls[0][0])).toContain("page=1");
    expect(String(fetchMock.mock.calls[1][0])).toContain("page=2");
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

    expect(items).toHaveLength(371);
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
    expect(items).toContainEqual(
      expect.objectContaining({
        skuCode: "GRN",
        name: "DẦU GỘI RETRO NANO 350ML",
        itemType: "finished_good",
        itemGroup: "retail_hair",
        uomBase: "BOTTLE",
        isSellable: true
      })
    );
    expect(items).toContainEqual(
      expect.objectContaining({
        skuCode: "BTMN",
        itemGroup: "gift_skin",
        uomBase: "PACK",
        isSellable: false
      })
    );
    expect(items).toContainEqual(
      expect.objectContaining({
        skuCode: "BONG",
        itemGroup: "accessory_skin",
        uomBase: "PCS"
      })
    );
    expect(items.some((item) => item.skuCode === "SU500" || item.skuCode === "BÔNG" || item.skuCode === "MCBĐ")).toBe(false);
    expect(summarizeProducts(items)).toEqual({
      total: 371,
      active: 371,
      draft: 0,
      controlled: 371
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
      "PACK",
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

function apiPage(data: ReturnType<typeof apiProduct>[]) {
  return new Response(
    JSON.stringify({
      success: true,
      data,
      pagination: {
        page: 1,
        page_size: 100,
        total_items: data.length,
        total_pages: 1
      },
      request_id: "req-product-page"
    }),
    { status: 200, headers: { "Content-Type": "application/json" } }
  );
}

function apiProduct(skuCode: string) {
  return {
    id: `item-${skuCode.toLowerCase()}`,
    item_code: skuCode,
    sku_code: skuCode,
    name: `Product ${skuCode}`,
    item_type: "raw_material",
    item_group: "api",
    brand_code: "MYH",
    uom_code: "KG",
    uom_base: "KG",
    uom_purchase: "KG",
    uom_issue: "KG",
    lot_controlled: true,
    expiry_controlled: true,
    shelf_life_days: 365,
    qc_required: true,
    status: "active",
    standard_cost: "0.000000",
    is_sellable: false,
    is_purchasable: true,
    is_producible: false,
    spec_version: "",
    created_at: "2026-05-03T00:00:00Z",
    updated_at: "2026-05-03T00:00:00Z"
  } as const;
}
