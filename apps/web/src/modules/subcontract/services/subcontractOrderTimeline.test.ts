import { describe, expect, it } from "vitest";
import {
  buildSubcontractOrderTimeline,
  productionFactoryOrderHref,
  productionFactoryOrderSourcePlanHref
} from "./subcontractOrderTimeline";
import type { SubcontractOrder } from "../types";

describe("subcontractOrderTimeline", () => {
  it("builds a factory-order timeline with the current factory confirmation gate", () => {
    const timeline = buildSubcontractOrderTimeline(
      {
        ...baseOrder,
        status: "approved"
      },
      { dispatchStatus: "sent" }
    );

    expect(timeline.map((item) => [item.id, item.status])).toEqual([
      ["created", "complete"],
      ["submitted", "complete"],
      ["approved", "complete"],
      ["factory-dispatch", "complete"],
      ["factory-confirmed", "current"],
      ["deposit", "pending"],
      ["materials-issued", "pending"],
      ["sample", "pending"],
      ["mass-production", "pending"],
      ["finished-goods-received", "pending"],
      ["qc", "pending"],
      ["final-payment", "pending"],
      ["closed", "pending"]
    ]);
    expect(timeline.find((item) => item.id === "factory-confirmed")).toMatchObject({
      label: "Nhà máy xác nhận",
      action: {
        label: "Mở xử lý lệnh",
        href: "/subcontract?source_production_plan_id=plan-001&search=PP-260505-0001#subcontract-orders",
        disabled: false
      }
    });
  });

  it("keeps factory confirmation pending until the dispatch is sent", () => {
    const timeline = buildSubcontractOrderTimeline(
      {
        ...baseOrder,
        status: "approved"
      },
      { dispatchStatus: "draft" }
    );

    expect(timeline.find((item) => item.id === "factory-dispatch")).toMatchObject({
      status: "current"
    });
    expect(timeline.find((item) => item.id === "factory-confirmed")).toMatchObject({
      status: "pending"
    });
  });

  it("marks receiving and QC as current after mass production starts", () => {
    const timeline = buildSubcontractOrderTimeline({
      ...baseOrder,
      status: "mass_production_started"
    });

    expect(timeline.find((item) => item.id === "materials-issued")).toMatchObject({ status: "complete" });
    expect(timeline.find((item) => item.id === "sample")).toMatchObject({ status: "complete" });
    expect(timeline.find((item) => item.id === "mass-production")).toMatchObject({ status: "complete" });
    expect(timeline.find((item) => item.id === "finished-goods-received")).toMatchObject({
      status: "current",
      action: {
        label: "Mở nhận thành phẩm",
        href: "/subcontract?source_production_plan_id=plan-001&search=PP-260505-0001#subcontract-inbound",
        disabled: false
      }
    });
  });

  it("keeps the sample gate current while the factory sample is only submitted", () => {
    const timeline = buildSubcontractOrderTimeline({
      ...baseOrder,
      status: "sample_submitted"
    });

    expect(timeline.find((item) => item.id === "sample")).toMatchObject({ status: "current" });
    expect(timeline.find((item) => item.id === "mass-production")).toMatchObject({ status: "pending" });
  });

  it("builds production-facing links for factory orders and their source plans", () => {
    expect(productionFactoryOrderHref(baseOrder)).toBe("/production/factory-orders/sco-001");
    expect(productionFactoryOrderSourcePlanHref(baseOrder)).toBe("/production/plans/plan-001");
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
