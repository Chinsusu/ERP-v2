import { beforeEach, describe, expect, it } from "vitest";
import {
  changeLocationStatus,
  changeWarehouseStatus,
  createLocation,
  createWarehouse,
  emptyLocationInput,
  emptyWarehouseInput,
  getLocations,
  getWarehouses,
  locationTypeLabel,
  resetPrototypeWarehouseMasterData,
  summarizeWarehouseLocations,
  updateLocation,
  updateWarehouse,
  warehouseStatusTone,
  warehouseTypeLabel
} from "./warehouseMasterDataService";

describe("warehouseMasterDataService", () => {
  beforeEach(() => {
    resetPrototypeWarehouseMasterData();
  });

  it("filters warehouse master data by search, status, and type", async () => {
    await expect(
      getWarehouses({
        search: "finished",
        status: "active",
        warehouseType: "finished_good"
      })
    ).resolves.toMatchObject([
      {
        warehouseCode: "WH-HCM-FG",
        status: "active",
        warehouseType: "finished_good"
      }
    ]);
  });

  it("creates, updates, and changes warehouse status in the local fallback store", async () => {
    const created = await createWarehouse({
      ...emptyWarehouseInput,
      warehouseCode: "wh-hn-fg",
      warehouseName: "Finished Goods Warehouse HN",
      siteCode: "hn",
      address: "Ha Noi DC"
    });

    expect(created).toMatchObject({
      warehouseCode: "WH-HN-FG",
      siteCode: "HN",
      status: "active"
    });
    expect(created.auditLogId).toContain("audit-local-warehouse-create");

    const updated = await updateWarehouse(created.id, {
      ...emptyWarehouseInput,
      warehouseCode: "WH-HN-FG",
      warehouseName: "Finished Goods Warehouse Ha Noi",
      siteCode: "HN"
    });
    expect(updated.warehouseName).toBe("Finished Goods Warehouse Ha Noi");

    const inactive = await changeWarehouseStatus(created.id, "inactive");
    expect(inactive.status).toBe("inactive");
  });

  it("blocks duplicate warehouse code and invalid location warehouse", async () => {
    await expect(
      createWarehouse({
        ...emptyWarehouseInput,
        warehouseCode: "WH-HCM-FG",
        warehouseName: "Duplicate FG"
      })
    ).rejects.toThrow("Warehouse code already exists");

    await expect(
      createLocation({
        ...emptyLocationInput,
        warehouseId: "missing-warehouse",
        locationCode: "FG-PACK-02",
        locationName: "Packing Bay 02",
        locationType: "pack"
      })
    ).rejects.toThrow("invalid warehouse");
  });

  it("creates, updates, filters, and inactivates warehouse locations", async () => {
    const created = await createLocation({
      ...emptyLocationInput,
      warehouseId: "wh-hcm-fg",
      locationCode: "fg-pack-02",
      locationName: "Packing Bay 02",
      locationType: "pack",
      zoneCode: "pack",
      allowStore: true
    });

    expect(created).toMatchObject({
      warehouseCode: "WH-HCM-FG",
      locationCode: "FG-PACK-02",
      zoneCode: "PACK",
      status: "active"
    });

    const updated = await updateLocation(created.id, {
      ...emptyLocationInput,
      warehouseId: "wh-hcm-fg",
      locationCode: "FG-PACK-02",
      locationName: "Packing Bay 02 Updated",
      locationType: "pack",
      zoneCode: "PACK",
      allowStore: true
    });
    expect(updated.locationName).toBe("Packing Bay 02 Updated");

    const inactive = await changeLocationStatus(created.id, "inactive");
    expect(inactive.status).toBe("inactive");
    const activeRows = await getLocations({ warehouseId: "wh-hcm-fg", status: "active" });
    expect(activeRows.find((location) => location.id === created.id)).toBeUndefined();
  });

  it("summarizes warehouse and location operating counts", async () => {
    const warehouses = await getWarehouses();
    const locations = await getLocations();

    expect(summarizeWarehouseLocations(warehouses, locations)).toEqual({
      warehouses: 4,
      activeWarehouses: 3,
      activeLocations: 4,
      receivingLocations: 3
    });
  });

  it("maps labels and tones", () => {
    expect(warehouseStatusTone("active")).toBe("success");
    expect(warehouseStatusTone("inactive")).toBe("warning");
    expect(warehouseTypeLabel("quarantine")).toBe("Quarantine");
    expect(locationTypeLabel("qc_hold")).toBe("QC hold");
  });
});
