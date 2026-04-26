import { describe, expect, it } from "vitest";
import {
  changeSubcontractOrderStatus,
  createSubcontractOrder,
  formatSubcontractDepositStatus,
  formatSubcontractOrderStatus,
  getSubcontractOrders,
  prototypeSubcontractOrders,
  subcontractDepositStatusTone,
  subcontractOrderStatusOptions,
  subcontractOrderStatusTone,
  summarizeSubcontractOrders
} from "./subcontractOrderService";

describe("subcontractOrderService", () => {
  it("creates an external factory order with the required skeleton fields", () => {
    const order = createSubcontractOrder({
      factoryId: "factory-lotus",
      productId: "product-repair-cream-50ml",
      quantity: 1200,
      specVersion: "SPEC-CREAM-50ML-v4",
      sampleRequired: true,
      expectedDeliveryDate: "2026-05-20",
      depositStatus: "pending",
      depositAmount: 5000000
    });

    expect(order).toMatchObject({
      factoryName: "Lotus GMP Factory",
      productName: "Repair Cream 50ml",
      quantity: 1200,
      specVersion: "SPEC-CREAM-50ML-v4",
      sampleRequired: true,
      expectedDeliveryDate: "2026-05-20",
      depositStatus: "pending",
      status: "DRAFT"
    });
  });

  it("rejects invalid external factory orders before creating a draft", () => {
    expect(() =>
      createSubcontractOrder({
        factoryId: "",
        productId: "product-repair-cream-50ml",
        quantity: 0,
        specVersion: "SPEC-CREAM-50ML-v4",
        sampleRequired: true,
        expectedDeliveryDate: "2026-05-20",
        depositStatus: "pending"
      })
    ).toThrow("Factory is required");
  });

  it("defines the Sprint 0 subcontract status model", () => {
    expect(subcontractOrderStatusOptions.map((option) => option.value)).toEqual([
      "DRAFT",
      "CONFIRMED",
      "MATERIAL_TRANSFERRED",
      "SAMPLE_APPROVED",
      "IN_PRODUCTION",
      "DELIVERED",
      "QC_REVIEW",
      "ACCEPTED",
      "REJECTED",
      "CLOSED"
    ]);
  });

  it("writes an audit log when the subcontract order status changes", () => {
    const [order] = prototypeSubcontractOrders;
    const result = changeSubcontractOrderStatus({
      order,
      nextStatus: "MATERIAL_TRANSFERRED",
      actorId: "user-subcontract-coordinator",
      actorName: "Subcontract Coordinator",
      note: "Materials handover recorded"
    });

    expect(result.order).toMatchObject({
      id: order.id,
      status: "MATERIAL_TRANSFERRED"
    });
    expect(result.order.auditLogIds).toContain(result.auditLog.id);
    expect(result.auditLog).toMatchObject({
      action: "subcontract.order.status_changed",
      entityType: "subcontract_order",
      entityId: order.id,
      beforeData: {
        status: "CONFIRMED"
      },
      afterData: {
        status: "MATERIAL_TRANSFERRED"
      },
      metadata: {
        note: "Materials handover recorded"
      }
    });
  });

  it("filters and summarizes subcontract orders", async () => {
    await expect(getSubcontractOrders({ status: "CONFIRMED" })).resolves.toMatchObject([
      {
        orderNo: "SUB-260426-0001",
        status: "CONFIRMED"
      }
    ]);

    expect(summarizeSubcontractOrders(prototypeSubcontractOrders)).toMatchObject({
      total: 1,
      confirmed: 1,
      active: 1,
      nextDeliveryDate: "2026-05-12"
    });
  });

  it("maps subcontract status and deposit status to UI labels and tones", () => {
    expect(formatSubcontractOrderStatus("MATERIAL_TRANSFERRED")).toBe("Material transferred");
    expect(subcontractOrderStatusTone("REJECTED")).toBe("danger");
    expect(subcontractOrderStatusTone("ACCEPTED")).toBe("success");
    expect(formatSubcontractDepositStatus("not_required")).toBe("Not required");
    expect(subcontractDepositStatusTone("pending")).toBe("warning");
  });
});
