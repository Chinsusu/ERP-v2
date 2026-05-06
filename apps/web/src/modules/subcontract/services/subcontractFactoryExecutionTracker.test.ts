import { describe, expect, it } from "vitest";
import { buildSubcontractFactoryExecutionTracker } from "./subcontractFactoryExecutionTracker";
import type { SubcontractOrder } from "../types";

describe("subcontractFactoryExecutionTracker", () => {
  it("points a factory-confirmed order to material handover", () => {
    const tracker = buildSubcontractFactoryExecutionTracker(
      {
        ...baseOrder,
        status: "factory_confirmed",
        depositStatus: "paid"
      },
      { dispatchStatus: "confirmed" }
    );

    expect(tracker.currentGate).toMatchObject({
      id: "material-handover",
      title: "Bàn giao vật tư cho nhà máy",
      status: "current",
      action: {
        label: "Mở xuất vật tư",
        href: "/production/factory-orders/sco-001#factory-material-handover",
        disabled: false
      }
    });
    expect(tracker.items.find((item) => item.id === "material-handover")).toMatchObject({
      metric: "0/1 dòng đủ"
    });
  });

  it("keeps material handover pending while required deposit is not recorded", () => {
    const tracker = buildSubcontractFactoryExecutionTracker(
      {
        ...baseOrder,
        status: "factory_confirmed",
        depositStatus: "pending"
      },
      { dispatchStatus: "confirmed" }
    );

    expect(tracker.currentGate).toMatchObject({
      id: "deposit",
      status: "current"
    });
    expect(tracker.items.find((item) => item.id === "material-handover")).toMatchObject({
      status: "pending"
    });
  });

  it("returns a revision-requested dispatch to the dispatch gate", () => {
    const tracker = buildSubcontractFactoryExecutionTracker(
      {
        ...baseOrder,
        status: "approved"
      },
      { dispatchStatus: "revision_requested" }
    );

    expect(tracker.currentGate).toMatchObject({
      id: "factory-dispatch",
      status: "current",
      metric: "Cần chỉnh"
    });
    expect(tracker.items.find((item) => item.id === "factory-confirmation")).toMatchObject({
      status: "pending"
    });
  });

  it("shows a terminal gate for cancelled factory orders", () => {
    const tracker = buildSubcontractFactoryExecutionTracker({
      ...baseOrder,
      status: "cancelled"
    });

    expect(tracker.currentGate).toMatchObject({
      id: "cancelled",
      title: "Lệnh đã hủy",
      status: "blocked",
      metric: "Đã hủy"
    });
    expect(tracker.items).toHaveLength(1);
  });

  it("skips the sample gate when the order does not require a sample", () => {
    const tracker = buildSubcontractFactoryExecutionTracker(
      {
        ...baseOrder,
        sampleRequired: false,
        status: "materials_issued_to_factory",
        materialLines: [
          {
            ...baseOrder.materialLines[0],
            issuedQty: "0.099900"
          }
        ]
      },
      { dispatchStatus: "confirmed" }
    );

    expect(tracker.items.find((item) => item.id === "sample-gate")).toMatchObject({
      status: "complete",
      metric: "Không yêu cầu mẫu"
    });
    expect(tracker.currentGate).toMatchObject({
      id: "mass-production",
      title: "Chạy sản xuất hàng loạt"
    });
  });

  it("blocks mass production while a submitted sample is rejected", () => {
    const tracker = buildSubcontractFactoryExecutionTracker(
      {
        ...baseOrder,
        status: "sample_rejected",
        materialLines: [
          {
            ...baseOrder.materialLines[0],
            issuedQty: "0.099900"
          }
        ]
      },
      { dispatchStatus: "confirmed" }
    );

    expect(tracker.currentGate).toMatchObject({
      id: "sample-gate",
      status: "blocked",
      action: {
        label: "Mở duyệt mẫu",
        href: "/production/factory-orders/sco-001#factory-sample-approval",
        disabled: false
      }
    });
    expect(tracker.items.find((item) => item.id === "mass-production")).toMatchObject({
      status: "blocked"
    });
  });

  it("links sample and mass production gates to the production factory order detail", () => {
    const tracker = buildSubcontractFactoryExecutionTracker(
      {
        ...baseOrder,
        status: "materials_issued_to_factory",
        materialLines: [
          {
            ...baseOrder.materialLines[0],
            issuedQty: "0.099900"
          }
        ]
      },
      { dispatchStatus: "confirmed" }
    );

    expect(tracker.items.find((item) => item.id === "sample-gate")).toMatchObject({
      status: "current",
      action: {
        href: "/production/factory-orders/sco-001#factory-sample-approval",
        disabled: false
      }
    });
    expect(tracker.items.find((item) => item.id === "mass-production")).toMatchObject({
      action: {
        href: "/production/factory-orders/sco-001#factory-mass-production"
      }
    });
  });

  it("links finished goods receipt to the production factory order detail", () => {
    const tracker = buildSubcontractFactoryExecutionTracker({
      ...baseOrder,
      status: "mass_production_started"
    });

    expect(tracker.items.find((item) => item.id === "finished-goods-receipt")).toMatchObject({
      status: "current",
      action: {
        href: "/production/factory-orders/sco-001#factory-finished-goods-receipt",
        disabled: false
      }
    });
  });

  it("links finished goods QC closeout to the production factory order detail", () => {
    const tracker = buildSubcontractFactoryExecutionTracker({
      ...baseOrder,
      status: "finished_goods_received",
      receivedQty: "80.000000",
      materialLines: [
        {
          ...baseOrder.materialLines[0],
          issuedQty: "0.099900"
        }
      ]
    });

    expect(tracker.currentGate).toMatchObject({
      id: "qc-closeout",
      status: "current",
      action: {
        label: "Má»Ÿ QC",
        href: "/production/factory-orders/sco-001#factory-finished-goods-qc-closeout",
        disabled: false
      }
    });
    expect(tracker.items.find((item) => item.id === "qc-closeout")).toMatchObject({
      action: {
        href: "/production/factory-orders/sco-001#factory-finished-goods-qc-closeout"
      }
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
  depositStatus: "pending",
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
      issuedQty: "0.000000",
      uomCode: "KG",
      unitCost: "0.000000",
      currencyCode: "VND",
      lineCostAmount: "0.00",
      lotTraceRequired: true
    }
  ],
  auditLogIds: []
};
