import { beforeEach, describe, expect, it } from "vitest";
import {
  approveSubcontractOrder,
  changeSubcontractOrderStatus,
  confirmFactorySubcontractOrder,
  createSubcontractOrder,
  formatSubcontractDepositStatus,
  formatSubcontractOrderStatus,
  getSubcontractOrders,
  issueSubcontractMaterials,
  prototypeSubcontractOrders,
  resetPrototypeSubcontractOrdersForTest,
  subcontractDepositStatusTone,
  subcontractOrderStatusOptions,
  subcontractOrderStatusTone,
  submitSubcontractOrder,
  summarizeSubcontractOrders
} from "./subcontractOrderService";

describe("subcontractOrderService", () => {
  beforeEach(() => {
    resetPrototypeSubcontractOrdersForTest();
  });

  it("creates an external factory order with the required API fields", async () => {
    const order = await createSubcontractOrder({
      factoryId: "sup-out-lotus",
      productId: "item-serum-30ml",
      quantity: 1200,
      specVersion: "SPEC-SERUM-2026.04",
      sampleRequired: true,
      expectedDeliveryDate: "2026-05-20",
      depositStatus: "pending",
      depositAmount: 5000000,
      materialItemId: "item-cream-50g",
      materialQty: "20",
      materialUnitCost: "58000"
    });

    expect(order).toMatchObject({
      factoryName: "Lotus Filling Partner",
      productName: "Hydrating Serum 30ml",
      quantity: 1200,
      specVersion: "SPEC-SERUM-2026.04",
      sampleRequired: true,
      expectedDeliveryDate: "2026-05-20",
      depositStatus: "pending",
      estimatedCostAmount: "1160000.00",
      status: "draft"
    });
  });

  it("rejects invalid external factory orders before creating a draft", async () => {
    await expect(
      createSubcontractOrder({
        factoryId: "",
        productId: "item-serum-30ml",
        quantity: 0,
        specVersion: "SPEC-SERUM-2026.04",
        sampleRequired: true,
        expectedDeliveryDate: "2026-05-20",
        depositStatus: "pending",
        materialItemId: "item-cream-50g",
        materialQty: "20",
        materialUnitCost: "58000"
      })
    ).rejects.toThrow("Factory is required");
  });

  it("defines the Sprint 5 subcontract status model", () => {
    expect(subcontractOrderStatusOptions.map((option) => option.value)).toEqual([
      "draft",
      "submitted",
      "approved",
      "factory_confirmed",
      "deposit_recorded",
      "materials_issued_to_factory",
      "sample_submitted",
      "sample_approved",
      "sample_rejected",
      "mass_production_started",
      "finished_goods_received",
      "qc_in_progress",
      "accepted",
      "rejected_with_factory_issue",
      "final_payment_ready",
      "closed",
      "cancelled"
    ]);
  });

  it("writes an audit log when the subcontract order status changes", () => {
    const [order] = prototypeSubcontractOrders;
    const result = changeSubcontractOrderStatus({
      order,
      nextStatus: "materials_issued_to_factory",
      actorId: "user-subcontract-coordinator",
      actorName: "Subcontract Coordinator",
      note: "Materials handover recorded"
    });

    expect(result.order).toMatchObject({
      id: order.id,
      status: "materials_issued_to_factory"
    });
    expect(result.order.auditLogIds).toContain(result.auditLog.id);
    expect(result.auditLog).toMatchObject({
      action: "subcontract.order.status_changed",
      entityType: "subcontract_order",
      entityId: order.id,
      beforeData: {
        status: "approved"
      },
      afterData: {
        status: "materials_issued_to_factory"
      },
      metadata: {
        note: "Materials handover recorded"
      }
    });
  });

  it("filters and summarizes subcontract orders", async () => {
    await expect(getSubcontractOrders({ status: "approved" })).resolves.toMatchObject([
      {
        orderNo: "SCO-260429-0001",
        status: "approved"
      }
    ]);

    expect(summarizeSubcontractOrders(prototypeSubcontractOrders)).toMatchObject({
      total: 1,
      confirmed: 1,
      active: 1,
      nextDeliveryDate: "2026-05-20"
    });
  });

  it("runs submit and approve actions against the prototype fallback", async () => {
    const draft = await createSubcontractOrder({
      factoryId: "sup-out-lotus",
      productId: "item-serum-30ml",
      quantity: 1200,
      specVersion: "SPEC-SERUM-2026.04",
      sampleRequired: true,
      expectedDeliveryDate: "2026-05-20",
      depositStatus: "pending",
      materialItemId: "item-cream-50g",
      materialQty: "20",
      materialUnitCost: "58000"
    });

    const submitted = await submitSubcontractOrder(draft.id, draft.version);
    const approved = await approveSubcontractOrder(submitted.order.id, submitted.order.version);

    expect(submitted.order.status).toBe("submitted");
    expect(approved.order.status).toBe("approved");
  });

  it("issues subcontract materials through the prototype fallback", async () => {
    const draft = await createSubcontractOrder({
      factoryId: "sup-out-lotus",
      productId: "item-serum-30ml",
      quantity: 1200,
      specVersion: "SPEC-SERUM-2026.04",
      sampleRequired: true,
      expectedDeliveryDate: "2026-05-20",
      depositStatus: "pending",
      materialItemId: "item-cream-50g",
      materialQty: "20",
      materialUnitCost: "58000"
    });
    const submitted = await submitSubcontractOrder(draft.id, draft.version);
    const approved = await approveSubcontractOrder(submitted.order.id, submitted.order.version);
    const confirmed = await confirmFactorySubcontractOrder(approved.order.id, approved.order.version);

    const result = await issueSubcontractMaterials({
      order: confirmed.order,
      sourceWarehouseId: "wh-hcm",
      sourceWarehouseCode: "HCM",
      handoverBy: "warehouse-user",
      receivedBy: "factory-receiver",
      lines: [
        {
          orderMaterialLineId: confirmed.order.materialLines[0].id,
          issueQty: "20",
          uomCode: "EA",
          batchNo: "CREAM-LOT-001"
        }
      ],
      evidence: [
        {
          evidenceType: "handover",
          fileName: "handover.pdf",
          objectKey: "subcontract/handover.pdf"
        }
      ]
    });

    expect(result.order.status).toBe("materials_issued_to_factory");
    expect(result.transfer).toMatchObject({
      orderId: confirmed.order.id,
      sourceWarehouseId: "wh-hcm",
      status: "SENT",
      signedHandover: true
    });
    expect(result.stockMovements).toHaveLength(1);
    expect(result.auditLog.action).toBe("subcontract.materials_issued");
  });

  it("accumulates partial subcontract material issues in the prototype fallback", async () => {
    const draft = await createSubcontractOrder({
      factoryId: "sup-out-lotus",
      productId: "item-serum-30ml",
      quantity: 1200,
      specVersion: "SPEC-SERUM-2026.04",
      sampleRequired: true,
      expectedDeliveryDate: "2026-05-20",
      depositStatus: "pending",
      materialItemId: "item-cream-50g",
      materialQty: "20",
      materialUnitCost: "58000"
    });
    const submitted = await submitSubcontractOrder(draft.id, draft.version);
    const approved = await approveSubcontractOrder(submitted.order.id, submitted.order.version);
    const confirmed = await confirmFactorySubcontractOrder(approved.order.id, approved.order.version);
    const firstIssue = await issueSubcontractMaterials({
      order: confirmed.order,
      sourceWarehouseId: "wh-hcm",
      sourceWarehouseCode: "HCM",
      handoverBy: "warehouse-user",
      receivedBy: "factory-receiver",
      lines: [
        {
          orderMaterialLineId: confirmed.order.materialLines[0].id,
          issueQty: "5",
          uomCode: "EA",
          batchNo: "CREAM-LOT-001"
        }
      ]
    });

    const finalIssue = await issueSubcontractMaterials({
      order: firstIssue.order,
      sourceWarehouseId: "wh-hcm",
      sourceWarehouseCode: "HCM",
      handoverBy: "warehouse-user",
      receivedBy: "factory-receiver",
      lines: [
        {
          orderMaterialLineId: firstIssue.order.materialLines[0].id,
          issueQty: "15",
          uomCode: "EA",
          batchNo: "CREAM-LOT-001"
        }
      ]
    });

    expect(firstIssue.order.status).toBe("factory_confirmed");
    expect(firstIssue.order.materialLines[0].issuedQty).toBe("5.000000");
    expect(finalIssue.order.status).toBe("materials_issued_to_factory");
    expect(finalIssue.order.materialLines[0].issuedQty).toBe("20.000000");
  });

  it("maps subcontract status and deposit status to UI labels and tones", () => {
    expect(formatSubcontractOrderStatus("materials_issued_to_factory")).toBe("Materials issued");
    expect(subcontractOrderStatusTone("rejected_with_factory_issue")).toBe("danger");
    expect(subcontractOrderStatusTone("accepted")).toBe("success");
    expect(formatSubcontractDepositStatus("not_required")).toBe("Not required");
    expect(subcontractDepositStatusTone("pending")).toBe("warning");
  });
});
