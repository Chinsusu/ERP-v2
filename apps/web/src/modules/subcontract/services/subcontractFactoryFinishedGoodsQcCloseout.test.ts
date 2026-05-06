import { describe, expect, it } from "vitest";
import {
  buildFactoryFinishedGoodsQcAcceptInput,
  buildFactoryFinishedGoodsQcPartialAcceptInput,
  buildFactoryFinishedGoodsQcRejectInput,
  buildSubcontractFactoryFinishedGoodsQcCloseout
} from "./subcontractFactoryFinishedGoodsQcCloseout";
import type { SubcontractFinishedGoodsReceipt, SubcontractOrder } from "../types";

describe("subcontractFactoryFinishedGoodsQcCloseout", () => {
  it("opens QC closeout after finished goods are received into QC hold", () => {
    const gate = buildSubcontractFactoryFinishedGoodsQcCloseout(
      {
        ...baseOrder,
        status: "finished_goods_received",
        receivedQty: "100.000000",
        acceptedQty: "0.000000",
        rejectedQty: "0.000000"
      },
      baseReceipt
    );

    expect(gate).toMatchObject({
      status: "ready_for_qc",
      canCloseout: true,
      receivedQty: "100.000000",
      acceptedQty: "0.000000",
      rejectedQty: "0.000000",
      remainingQcQty: "100.000000",
      uomCode: "PCS"
    });
  });

  it("blocks QC closeout before finished goods receipt", () => {
    const gate = buildSubcontractFactoryFinishedGoodsQcCloseout(
      {
        ...baseOrder,
        status: "mass_production_started",
        receivedQty: "0.000000"
      },
      undefined
    );

    expect(gate).toMatchObject({
      status: "blocked",
      canCloseout: false,
      remainingQcQty: "0.000000",
      blockedReason: expect.stringContaining("QC hold")
    });
  });

  it("opens full-pass QC closeout from order quantities when receipt details are not loaded", () => {
    const gate = buildSubcontractFactoryFinishedGoodsQcCloseout({
      ...baseOrder,
      status: "finished_goods_received",
      receivedQty: "100.000000",
      acceptedQty: "0.000000",
      rejectedQty: "0.000000"
    });

    expect(gate).toMatchObject({
      status: "ready_for_qc",
      canCloseout: true,
      receivedQty: "100.000000",
      remainingQcQty: "100.000000",
      uomCode: "PCS"
    });
  });

  it("builds a full-pass accept payload for the latest QC hold receipt", () => {
    const input = buildFactoryFinishedGoodsQcAcceptInput({
      order: {
        ...baseOrder,
        status: "finished_goods_received",
        receivedQty: "100.000000"
      },
      latestReceipt: baseReceipt,
      acceptedBy: " qa-user ",
      acceptedAt: "2026-05-06T08:00:00Z",
      note: " dat qc "
    });

    expect(input).toMatchObject({
      order: expect.objectContaining({ id: "sco-001" }),
      acceptedBy: "qa-user",
      acceptedAt: "2026-05-06T08:00:00Z",
      note: "dat qc"
    });
  });

  it("builds a partial-pass payload with accepted quantity and factory claim evidence", () => {
    const input = buildFactoryFinishedGoodsQcPartialAcceptInput({
      order: {
        ...baseOrder,
        status: "qc_in_progress",
        receivedQty: "100.000000"
      },
      latestReceipt: baseReceipt,
      acceptedQty: "80",
      rejectedQty: "20",
      acceptedBy: "qa-user",
      openedBy: "qa-user",
      ownerId: "factory-owner",
      reasonCode: "QUALITY_FAIL",
      reason: "20 pcs fail viscosity check",
      severity: "P1",
      evidenceFileName: "qc-fail-photo.jpg",
      evidenceNote: "photo after inspection"
    });

    expect(input).toMatchObject({
      order: expect.objectContaining({ id: "sco-001" }),
      acceptedQty: "80",
      rejectedQty: "20",
      uomCode: "PCS",
      baseAcceptedQty: "80",
      baseRejectedQty: "20",
      receiptId: "sfgr-001",
      receiptNo: "SFGR-260506-001",
      reasonCode: "QUALITY_FAIL",
      reason: "20 pcs fail viscosity check",
      severity: "P1",
      ownerId: "factory-owner",
      acceptedBy: "qa-user",
      openedBy: "qa-user",
      evidence: [
        {
          evidenceType: "qc_photo",
          fileName: "qc-fail-photo.jpg",
          objectKey: "subcontract-finished-goods/sco-001/qc/qc-fail-photo.jpg",
          note: "photo after inspection"
        }
      ]
    });
    expect(input.claimId).toMatch(/^sfc-sco-001-/);
  });

  it("builds a reject payload for the full remaining QC quantity", () => {
    const input = buildFactoryFinishedGoodsQcRejectInput({
      order: {
        ...baseOrder,
        status: "finished_goods_received",
        receivedQty: "100.000000"
      },
      latestReceipt: baseReceipt,
      ownerId: "factory-owner",
      openedBy: "qa-user",
      reasonCode: "PACKAGING_DAMAGED",
      reason: "carton crushed",
      severity: "P2",
      evidenceFileName: "qc-fail-photo.jpg"
    });

    expect(input).toMatchObject({
      order: expect.objectContaining({ id: "sco-001" }),
      receiptId: "sfgr-001",
      receiptNo: "SFGR-260506-001",
      affectedQty: "100.000000",
      uomCode: "PCS",
      baseAffectedQty: "100.000000",
      baseUOMCode: "PCS",
      reasonCode: "PACKAGING_DAMAGED",
      reason: "carton crushed",
      severity: "P2",
      ownerId: "factory-owner",
      openedBy: "qa-user"
    });
  });
});

const baseOrder: SubcontractOrder = {
  id: "sco-001",
  orderNo: "SCO-260505-0001",
  factoryId: "factory-001",
  factoryCode: "FACTORY-001",
  factoryName: "Factory Partner",
  productId: "item-aah",
  sku: "AAH",
  productName: "Kem u phuc hoi AS A HABIT BIO 350GR",
  quantity: 100,
  uomCode: "PCS",
  receivedQty: "0.000000",
  acceptedQty: "0.000000",
  rejectedQty: "0.000000",
  sourceProductionPlanId: "plan-001",
  sourceProductionPlanNo: "PP-260505-0001",
  specVersion: "S23SMK260504200049",
  sampleRequired: true,
  expectedDeliveryDate: "2026-05-20",
  depositStatus: "paid",
  depositAmount: 1000000,
  finalPaymentStatus: "hold",
  status: "draft",
  createdBy: "Production Ops",
  createdAt: "2026-05-05T08:00:00Z",
  updatedAt: "2026-05-05T08:00:00Z",
  version: 1,
  estimatedCostAmount: "1000000.00",
  materialLines: [],
  auditLogIds: []
};

const baseReceipt: SubcontractFinishedGoodsReceipt = {
  id: "sfgr-001",
  receiptNo: "SFGR-260506-001",
  orderId: "sco-001",
  orderNo: "SCO-260505-0001",
  factoryId: "factory-001",
  factoryCode: "FACTORY-001",
  factoryName: "Factory Partner",
  warehouseId: "warehouse_main",
  warehouseCode: "WH-MAIN",
  locationId: "qc_hold",
  locationCode: "QC-HOLD",
  deliveryNoteNo: "DN-260506-01",
  status: "qc_hold",
  lines: [
    {
      id: "sfgr-line-001",
      lineNo: 1,
      itemId: "item-aah",
      skuCode: "AAH",
      itemName: "Kem u phuc hoi AS A HABIT BIO 350GR",
      batchId: "BATCH-260506",
      batchNo: "BATCH-260506",
      lotNo: "LOT-01",
      expiryDate: "2028-05-06",
      receiveQty: "100.000000",
      uomCode: "PCS",
      baseReceiveQty: "100.000000",
      baseUOMCode: "PCS",
      conversionFactor: "1",
      packagingStatus: "intact"
    }
  ],
  evidence: [
    {
      id: "sfgr-evidence-001",
      evidenceType: "delivery_note",
      fileName: "delivery-note.pdf",
      objectKey: "subcontract-finished-goods/sco-001/delivery-note.pdf"
    }
  ],
  receivedBy: "warehouse-user",
  receivedAt: "2026-05-06T07:30:00Z",
  createdAt: "2026-05-06T07:30:00Z",
  updatedAt: "2026-05-06T07:30:00Z",
  version: 1
};
