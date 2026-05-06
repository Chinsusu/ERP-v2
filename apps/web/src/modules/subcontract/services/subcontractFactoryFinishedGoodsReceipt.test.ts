import { describe, expect, it } from "vitest";
import {
  buildFactoryFinishedGoodsReceiptInput,
  buildSubcontractFactoryFinishedGoodsReceipt
} from "./subcontractFactoryFinishedGoodsReceipt";
import type { SubcontractOrder } from "../types";

describe("subcontractFactoryFinishedGoodsReceipt", () => {
  it("opens finished goods receipt after mass production starts", () => {
    const gate = buildSubcontractFactoryFinishedGoodsReceipt({
      ...baseOrder,
      status: "mass_production_started"
    });

    expect(gate).toMatchObject({
      status: "ready_to_receive",
      canReceive: true,
      plannedQty: "999.000000",
      receivedQty: "0.000000",
      remainingQty: "999.000000",
      uomCode: "PCS"
    });
  });

  it("keeps receipt blocked before mass production starts", () => {
    const gate = buildSubcontractFactoryFinishedGoodsReceipt({
      ...baseOrder,
      status: "sample_approved"
    });

    expect(gate).toMatchObject({
      status: "blocked",
      canReceive: false,
      blockedReason: expect.stringContaining("sản xuất hàng loạt")
    });
  });

  it("keeps partial receipt open until the planned quantity is fully received", () => {
    const gate = buildSubcontractFactoryFinishedGoodsReceipt({
      ...baseOrder,
      status: "finished_goods_received",
      receivedQty: "450.000000"
    });

    expect(gate).toMatchObject({
      status: "partial",
      canReceive: true,
      plannedQty: "999.000000",
      receivedQty: "450.000000",
      remainingQty: "549.000000"
    });
  });

  it("closes receipt when finished goods are fully received", () => {
    const gate = buildSubcontractFactoryFinishedGoodsReceipt({
      ...baseOrder,
      status: "finished_goods_received",
      receivedQty: "999.000000"
    });

    expect(gate).toMatchObject({
      status: "complete",
      canReceive: false,
      remainingQty: "0.000000"
    });
  });

  it("blocks extra receipt after the order moves into QC", () => {
    const gate = buildSubcontractFactoryFinishedGoodsReceipt({
      ...baseOrder,
      status: "qc_in_progress",
      receivedQty: "450.000000"
    });

    expect(gate).toMatchObject({
      status: "blocked",
      canReceive: false,
      remainingQty: "549.000000",
      blockedReason: expect.stringContaining("QC")
    });
  });

  it("builds a QC hold receipt payload for the production factory order", () => {
    const input = buildFactoryFinishedGoodsReceiptInput({
      order: {
        ...baseOrder,
        status: "mass_production_started"
      },
      warehouseId: "warehouse_main",
      warehouseCode: "WH-MAIN",
      locationId: "qc_hold",
      locationCode: "QC-HOLD",
      deliveryNoteNo: " DN-260506-01 ",
      receivedBy: " warehouse-user ",
      evidenceFileName: " packing-list.pdf ",
      note: " received from factory ",
      draft: {
        receiveQty: "999",
        batchNo: "BATCH-260506",
        lotNo: "LOT-01",
        expiryDate: "2028-05-06",
        packagingStatus: "intact",
        note: " du kien "
      }
    });

    expect(input).toMatchObject({
      order: expect.objectContaining({ id: "sco-001" }),
      warehouseId: "warehouse_main",
      warehouseCode: "WH-MAIN",
      locationId: "qc_hold",
      locationCode: "QC-HOLD",
      deliveryNoteNo: "DN-260506-01",
      receivedBy: "warehouse-user",
      note: "received from factory",
      lines: [
        {
          lineNo: 1,
          itemId: "item-aah",
          skuCode: "AAH",
          itemName: "Kem u phuc hoi AS A HABIT BIO 350GR",
          batchId: "BATCH-260506",
          batchNo: "BATCH-260506",
          lotNo: "LOT-01",
          expiryDate: "2028-05-06",
          receiveQty: "999",
          uomCode: "PCS",
          baseReceiveQty: "999",
          baseUOMCode: "PCS",
          conversionFactor: "1",
          packagingStatus: "intact",
          note: "du kien"
        }
      ],
      evidence: [
        {
          evidenceType: "delivery_note",
          fileName: "packing-list.pdf",
          objectKey: "subcontract-finished-goods/sco-001/packing-list.pdf"
        }
      ]
    });
    expect(input.receiptId).toMatch(/^sfgr-sco-001-/);
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
  quantity: 999,
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
  materialLines: [
    {
      id: "sco-line-001",
      itemId: "item-aci-bha",
      skuCode: "ACI_BHA",
      itemName: "ACID SALICYLIC",
      plannedQty: "0.099900",
      issuedQty: "0.099900",
      uomCode: "KG",
      unitCost: "0.000000",
      currencyCode: "VND",
      lineCostAmount: "0.00",
      lotTraceRequired: true
    }
  ],
  auditLogIds: []
};
