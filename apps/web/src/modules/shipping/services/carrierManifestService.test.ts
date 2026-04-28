import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  addShipmentToManifest,
  cancelCarrierManifest,
  carrierManifestScanSeverityTone,
  carrierManifestStatusTone,
  confirmCarrierManifestHandover,
  createCarrierManifest,
  getCarrierManifests,
  markCarrierManifestReady,
  prototypeCarrierManifests,
  removeShipmentFromManifest,
  resetPrototypeCarrierManifestsForTest,
  summarizeCarrierManifestLines,
  verifyCarrierManifestScan
} from "./carrierManifestService";

describe("carrierManifestService", () => {
  beforeEach(() => {
    resetPrototypeCarrierManifestsForTest();
    vi.stubGlobal(
      "fetch",
      vi.fn(() => Promise.reject(new Error("offline")))
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

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

  it("adds a same-carrier shipment to a manifest and increases missing count", async () => {
    await expect(addShipmentToManifest("manifest-hcm-vtp-noon", "ship-hcm-vtp-260426-002")).resolves.toMatchObject({
      status: "ready",
      summary: {
        expectedCount: 2,
        scannedCount: 0,
        missingCount: 2
      }
    });
  });

  it("rejects wrong-carrier shipments in prototype fallback", async () => {
    await expect(addShipmentToManifest("manifest-hcm-vtp-noon", "ship-hcm-260426-004")).rejects.toThrow(
      "Shipment carrier does not match carrier manifest"
    );
  });

  it("marks draft manifests ready removes shipments and cancels manifests", async () => {
    const created = await createCarrierManifest({
      carrierCode: "GHN",
      carrierName: "GHN",
      warehouseId: "wh-hcm",
      warehouseCode: "HCM",
      date: "2026-04-26"
    });
    const added = await addShipmentToManifest(created.id, "ship-hcm-260426-004");

    expect(added.status).toBe("draft");
    expect(added.summary.expectedCount).toBe(1);

    const ready = await markCarrierManifestReady(added.id);
    expect(ready.status).toBe("ready");

    const removed = await removeShipmentFromManifest(ready.id, "ship-hcm-260426-004");
    expect(removed).toMatchObject({
      status: "draft",
      summary: { expectedCount: 0, scannedCount: 0, missingCount: 0 }
    });

    const cancelled = await cancelCarrierManifest(removed.id, "carrier pickup moved");
    expect(cancelled.status).toBe("cancelled");
  });

  it("maps manifest status to UI tones", () => {
    expect(carrierManifestStatusTone("completed")).toBe("success");
    expect(carrierManifestStatusTone("handed_over")).toBe("success");
    expect(carrierManifestStatusTone("exception")).toBe("danger");
    expect(carrierManifestStatusTone("cancelled")).toBe("danger");
    expect(carrierManifestStatusTone("scanning")).toBe("warning");
  });

  it("verifies a matching tracking scan and updates counts", async () => {
    await expect(
      verifyCarrierManifestScan({
        manifestId: "manifest-hcm-ghn-morning",
        code: "GHN260426003",
        stationId: "dock-a",
        deviceId: "scanner-01",
        source: "handheld_scanner"
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
        stationId: "dock-a",
        deviceId: "scanner-01",
        source: "handheld_scanner"
      }
    });
  });

  it("confirms handover only after all manifest lines are scanned", async () => {
    await expect(confirmCarrierManifestHandover("manifest-hcm-ghn-morning")).rejects.toThrow("missing orders");

    await verifyCarrierManifestScan({
      manifestId: "manifest-hcm-ghn-morning",
      code: "GHN260426003"
    });

    await expect(confirmCarrierManifestHandover("manifest-hcm-ghn-morning")).resolves.toMatchObject({
      status: "handed_over",
      auditLogId: "audit-manifest-handed-over-prototype",
      summary: { missingCount: 0 }
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
    await expect(verifyCarrierManifestScan({ manifestId: "manifest-hcm-ghn-morning", code: "VTP260426012" })).resolves.toMatchObject({
      resultCode: "MANIFEST_MISMATCH",
      message: "Shipment carrier does not match carrier manifest"
    });
    await expect(verifyCarrierManifestScan({ manifestId: "manifest-hcm-ghn-morning", code: "GHN260426004" })).resolves.toMatchObject({
      resultCode: "MANIFEST_MISMATCH",
      message: "Packed shipment is not expected on this manifest",
      line: { trackingNo: "GHN260426004" }
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

  it("maps API manifest responses and action endpoints", async () => {
    const manifestApi = {
      id: "manifest-api-1",
      carrier_code: "GHN",
      carrier_name: "GHN Express",
      warehouse_id: "wh-hcm",
      warehouse_code: "HCM",
      date: "2026-04-26",
      handover_batch: "morning",
      staging_zone: "handover-a",
      handover_zone_code: "handover-a",
      handover_bin_code: "A01",
      status: "draft",
      owner: "Warehouse Lead",
      audit_log_id: "audit-api",
      summary: {
        expected_count: 1,
        scanned_count: 0,
        missing_count: 1
      },
      lines: [
        {
          id: "line-api-1",
          shipment_id: "ship-api-1",
          order_no: "SO-API-1",
          tracking_no: "GHNAPI1",
          package_code: "TOTE-A01",
          staging_zone: "handover-a",
          handover_zone_code: "handover-a",
          handover_bin_code: "A01",
          scanned: false
        }
      ],
      created_at: "2026-04-26T07:45:00Z"
    };
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ success: true, data: [manifestApi], request_id: "req-list" }), { status: 200 })
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({ success: true, data: { ...manifestApi, status: "ready" }, request_id: "req-ready" }),
          { status: 200 }
        )
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ success: true, data: manifestApi, request_id: "req-remove" }), { status: 200 })
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            success: true,
            data: {
              result_code: "MATCHED",
              severity: "success",
              message: "Scan matched manifest line",
              line: manifestApi.lines[0],
              scan_event: {
                id: "scan-api-1",
                manifest_id: "manifest-api-1",
                code: "GHNAPI1",
                result_code: "MATCHED",
                severity: "success",
                message: "Scan matched manifest line",
                shipment_id: "ship-api-1",
                order_no: "SO-API-1",
                tracking_no: "GHNAPI1",
                actor_id: "user-handover-operator",
                station_id: "dock-a",
                device_id: "scanner-01",
                source: "handheld_scanner",
                warehouse_id: "wh-hcm",
                carrier_code: "GHN",
                created_at: "2026-04-26T08:10:00Z"
              },
              manifest: {
                ...manifestApi,
                status: "scanning",
                summary: { expected_count: 1, scanned_count: 1, missing_count: 0 },
                lines: [{ ...manifestApi.lines[0], scanned: true }]
              },
              audit_log_id: "audit-scan-api"
            },
            request_id: "req-scan"
          }),
          { status: 200 }
        )
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            success: true,
            data: {
              ...manifestApi,
              status: "handed_over",
              summary: { expected_count: 1, scanned_count: 1, missing_count: 0 },
              lines: [{ ...manifestApi.lines[0], scanned: true }],
              missing_lines: [],
              audit_log_id: "audit-confirm-api"
            },
            request_id: "req-confirm"
          }),
          { status: 200 }
        )
      );
    vi.stubGlobal("fetch", fetchMock);

    const manifests = await getCarrierManifests({ warehouseId: "wh-hcm", date: "2026-04-26" });
    const ready = await markCarrierManifestReady("manifest-api-1");
    await removeShipmentFromManifest("manifest-api-1", "ship-api-1");
    const scan = await verifyCarrierManifestScan({
      manifestId: "manifest-api-1",
      code: "GHNAPI1",
      stationId: "dock-a",
      deviceId: "scanner-01",
      source: "handheld_scanner"
    });
    const handedOver = await confirmCarrierManifestHandover("manifest-api-1");

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "http://localhost:8080/api/v1/shipping/manifests?warehouse_id=wh-hcm&date=2026-04-26",
      {
        headers: {
          Authorization: "Bearer local-dev-access-token"
        }
      }
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "http://localhost:8080/api/v1/shipping/manifests/manifest-api-1/ready",
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer local-dev-access-token"
        },
        body: "{}"
      }
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      3,
      "http://localhost:8080/api/v1/shipping/manifests/manifest-api-1/shipments/ship-api-1",
      {
        method: "DELETE",
        headers: {
          Authorization: "Bearer local-dev-access-token"
        }
      }
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      4,
      "http://localhost:8080/api/v1/shipping/manifests/manifest-api-1/scan",
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer local-dev-access-token"
        },
        body: JSON.stringify({
          code: "GHNAPI1",
          station_id: "dock-a",
          device_id: "scanner-01",
          source: "handheld_scanner"
        })
      }
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      5,
      "http://localhost:8080/api/v1/shipping/manifests/manifest-api-1/confirm-handover",
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer local-dev-access-token"
        },
        body: "{}"
      }
    );
    expect(manifests[0]).toMatchObject({
      id: "manifest-api-1",
      handoverZoneCode: "handover-a",
      handoverBinCode: "A01",
      missingLines: [{ trackingNo: "GHNAPI1" }],
      lines: [{ handoverZoneCode: "handover-a", handoverBinCode: "A01" }]
    });
    expect(ready.status).toBe("ready");
    expect(scan.scanEvent).toMatchObject({
      code: "GHNAPI1",
      deviceId: "scanner-01",
      source: "handheld_scanner"
    });
    expect(handedOver).toMatchObject({
      status: "handed_over",
      auditLogId: "audit-confirm-api",
      missingLines: []
    });
  });
});
