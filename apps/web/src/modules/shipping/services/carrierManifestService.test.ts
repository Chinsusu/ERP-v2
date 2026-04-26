import { describe, expect, it } from "vitest";
import {
  addShipmentToManifest,
  carrierManifestScanSeverityTone,
  carrierManifestStatusTone,
  createCarrierManifest,
  getCarrierManifests,
  prototypeCarrierManifests,
  summarizeCarrierManifestLines,
  verifyCarrierManifestScan
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

  it("verifies a matching tracking scan and updates counts", async () => {
    await expect(
      verifyCarrierManifestScan({
        manifestId: "manifest-hcm-ghn-morning",
        code: "GHN260426003",
        stationId: "dock-a"
      })
    ).resolves.toMatchObject({
      resultCode: "MATCHED",
      severity: "success",
      manifest: {
        summary: {
          expectedCount: 3,
          scannedCount: 3,
          missingCount: 0
        }
      },
      scanEvent: {
        code: "GHN260426003",
        resultCode: "MATCHED",
        stationId: "dock-a"
      }
    });
  });

  it("returns clear warning codes for duplicate wrong manifest unpacked and unknown scans", async () => {
    await expect(verifyCarrierManifestScan({ manifestId: "manifest-hcm-ghn-morning", code: "GHN260426001" })).resolves.toMatchObject({
      resultCode: "DUPLICATE_SCAN"
    });
    await expect(verifyCarrierManifestScan({ manifestId: "manifest-hcm-ghn-morning", code: "VTP260426011" })).resolves.toMatchObject({
      resultCode: "MANIFEST_MISMATCH",
      expectedManifestId: "manifest-hcm-vtp-noon"
    });
    await expect(verifyCarrierManifestScan({ manifestId: "manifest-hcm-ghn-morning", code: "GHN260426099" })).resolves.toMatchObject({
      resultCode: "INVALID_STATE"
    });
    await expect(verifyCarrierManifestScan({ manifestId: "manifest-hcm-ghn-morning", code: "UNKNOWN-CODE" })).resolves.toMatchObject({
      resultCode: "NOT_FOUND"
    });
  });

  it("maps scan severity directly to UI tone", () => {
    expect(carrierManifestScanSeverityTone("success")).toBe("success");
    expect(carrierManifestScanSeverityTone("warning")).toBe("warning");
    expect(carrierManifestScanSeverityTone("danger")).toBe("danger");
  });
});
