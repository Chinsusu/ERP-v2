import { describe, expect, it } from "vitest";
import { setActiveLocale } from "../../../shared/i18n/runtime";
import { getMasterDataTabs, getMasterDataViewStatusLabel } from "./masterDataTabs";

describe("master data tab model", () => {
  it("splits finished products, materials, supplier, and customer master data into separate tabs", () => {
    setActiveLocale("en");

    const tabs = getMasterDataTabs();

    expect(tabs).toEqual([
      { label: "Finished products", value: "finishedProducts" },
      { label: "Materials / Packaging", value: "materials" },
      { label: "Warehouses / Locations", value: "warehouses" },
      { label: "Suppliers", value: "suppliers" },
      { label: "Customers", value: "customers" }
    ]);
    expect(tabs.map((tab) => tab.value)).not.toContain("parties");
    expect(tabs.map((tab) => tab.value)).not.toContain("formulas");
  });

  it("uses distinct status copy for the supplier and customer views", () => {
    setActiveLocale("en");

    expect(getMasterDataViewStatusLabel("suppliers")).toBe("Supplier setup");
    expect(getMasterDataViewStatusLabel("customers")).toBe("Customer setup");
  });

  it("uses distinct status copy for finished products and material views", () => {
    setActiveLocale("en");

    expect(getMasterDataViewStatusLabel("finishedProducts")).toBe("Finished product setup");
    expect(getMasterDataViewStatusLabel("materials")).toBe("Material setup");
  });
});
