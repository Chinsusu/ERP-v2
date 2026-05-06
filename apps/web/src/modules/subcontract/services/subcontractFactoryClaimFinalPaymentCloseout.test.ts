import { describe, expect, it } from "vitest";
import { buildSubcontractFactoryClaimFinalPaymentCloseout } from "./subcontractFactoryClaimFinalPaymentCloseout";
import type { SubcontractFactoryClaim, SubcontractOrder } from "../types";

describe("subcontractFactoryClaimFinalPaymentCloseout", () => {
  it("blocks final payment while a partial QC claim is still open", () => {
    const closeout = buildSubcontractFactoryClaimFinalPaymentCloseout(
      {
        ...baseOrder,
        status: "accepted",
        acceptedQty: "994.000000",
        rejectedQty: "5.000000"
      },
      [{ ...baseClaim, status: "open", blocksFinalPayment: true }]
    );

    expect(closeout.blockingClaimCount).toBe(1);
    expect(closeout.claimStatus).toBe("current");
    expect(closeout.finalPaymentStatus).toBe("blocked");
    expect(closeout.canAcknowledgeClaim).toBe(true);
    expect(closeout.canResolveClaim).toBe(true);
    expect(closeout.canMarkFinalPaymentReady).toBe(false);
  });

  it("allows final payment after the factory claim is resolved", () => {
    const closeout = buildSubcontractFactoryClaimFinalPaymentCloseout(
      {
        ...baseOrder,
        status: "accepted",
        acceptedQty: "994.000000",
        rejectedQty: "5.000000"
      },
      [
        {
          ...baseClaim,
          status: "resolved",
          blocksFinalPayment: false,
          resolvedBy: "factory-owner",
          resolvedAt: "2026-05-07T09:00:00Z",
          resolutionNote: "Factory accepted credit memo before final payment"
        }
      ]
    );

    expect(closeout.blockingClaimCount).toBe(0);
    expect(closeout.claimStatus).toBe("complete");
    expect(closeout.finalPaymentStatus).toBe("current");
    expect(closeout.canAcknowledgeClaim).toBe(false);
    expect(closeout.canResolveClaim).toBe(false);
    expect(closeout.canMarkFinalPaymentReady).toBe(true);
  });

  it("does not allow final payment for a full factory rejection even after claim resolution", () => {
    const closeout = buildSubcontractFactoryClaimFinalPaymentCloseout(
      {
        ...baseOrder,
        status: "rejected_with_factory_issue",
        acceptedQty: "0.000000",
        rejectedQty: "999.000000"
      },
      [
        {
          ...baseClaim,
          status: "resolved",
          blocksFinalPayment: false,
          resolvedBy: "factory-owner",
          resolvedAt: "2026-05-07T09:00:00Z",
          resolutionNote: "Factory will remake this order"
        }
      ]
    );

    expect(closeout.claimStatus).toBe("complete");
    expect(closeout.finalPaymentStatus).toBe("blocked");
    expect(closeout.canMarkFinalPaymentReady).toBe(false);
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
  receivedQty: "999.000000",
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
  status: "finished_goods_received",
  createdBy: "Production Ops",
  createdAt: "2026-05-05T08:00:00Z",
  updatedAt: "2026-05-05T08:00:00Z",
  version: 1,
  estimatedCostAmount: "1000000.00",
  materialLines: [],
  auditLogIds: []
};

const baseClaim: SubcontractFactoryClaim = {
  id: "sfc-001",
  claimNo: "SFC-260507-0001",
  orderId: "sco-001",
  orderNo: "SCO-260505-0001",
  factoryId: "factory-001",
  factoryCode: "FACTORY-001",
  factoryName: "Factory Partner",
  receiptId: "sfgr-001",
  receiptNo: "SFGR-260507-0001",
  reasonCode: "QUALITY_FAIL",
  reason: "QC failed after receiving finished goods",
  severity: "P1",
  status: "open",
  affectedQty: "5.000000",
  uomCode: "PCS",
  baseAffectedQty: "5.000000",
  baseUOMCode: "PCS",
  evidence: [
    {
      id: "sfc-001-evidence-01",
      evidenceType: "qc_photo",
      objectKey: "claims/sfc-001/photo.jpg",
      createdAt: "2026-05-07T08:00:00Z",
      createdBy: "qa-user"
    }
  ],
  ownerId: "factory-owner",
  openedBy: "qa-user",
  openedAt: "2026-05-07T08:00:00Z",
  dueAt: "2026-05-14T08:00:00Z",
  blocksFinalPayment: true,
  createdAt: "2026-05-07T08:00:00Z",
  updatedAt: "2026-05-07T08:00:00Z",
  version: 1
};
