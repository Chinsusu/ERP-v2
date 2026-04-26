import { describe, expect, it } from "vitest";
import {
  addShipmentToManifest,
  carrierManifestStatusTone,
  createCarrierManifest,
  getCarrierManifests,
  prototypeCarrierManifests,
  summarizeCarrierManifestLines
} from "./carrierManifestService";

describe("carrierManifestService", () => {
  it("filters carrier manifests by warehouse, date, carrier, and status", async () => {
    await expect(
      getCarrierManifests({
        warehouseId: "wh-hcm",
        date: "2026-04-26",
        carrierCode: "GHN",
        status: "scanning"
      })
    ).resolves.toMatchObject([
      {
        id: "manifest-hcm-ghn-morning",
        summary: {
          expectedCount: 3,
          scannedCount: 2,
          missingCount: 1
        }
      }
    ]);
  });

  it("summarizes expected scanned and missing counts", () => {
    expect(summarizeCarrierManifestLines(prototypeCarrierManifests[0].lines)).toEqual({
      expectedCount: 3,
      scannedCount: 2,
      missingCount: 1
    });
  });

  it("creates a draft manifest by carrier date and warehouse", async () => {
    await expect(
      createCarrierManifest({
        carrierCode: "NJV",
        carrierName: "Ninja Van",
        warehouseId: "wh-hcm",
        warehouseCode: "HCM",
        date: "2026-04-26"
      })
    ).resolves.toMatchObject({
      carrierCode: "NJV",
      warehouseId: "wh-hcm",
      status: "draft",
      summary: {
        expectedCount: 0,
        scannedCount: 0,
        missingCount: 0
      }
    });
  });

  it("adds a shipment to a manifest and increases missing count", async () => {
    await expect(addShipmentToManifest("manifest-hcm-ghn-morning", "ship-hcm-260426-004")).resolves.toMatchObject({
      summary: {
        expectedCount: 4,
        scannedCount: 2,
        missingCount: 2
      }
    });
  });

  it("maps manifest status to UI tones", () => {
    expect(carrierManifestStatusTone("completed")).toBe("success");
    expect(carrierManifestStatusTone("exception")).toBe("danger");
    expect(carrierManifestStatusTone("scanning")).toBe("warning");
  });
});
