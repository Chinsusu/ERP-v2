import { describe, expect, it } from "vitest";
import {
  buildFactoryFinalPaymentApHandoff,
  buildFactoryFinalPaymentApHandoffFromSource,
  buildFactoryFinalPaymentApHandoffHref
} from "./subcontractFactoryFinalPaymentApHandoff";
import type { SubcontractPaymentMilestoneResult } from "../types";

describe("subcontractFactoryFinalPaymentApHandoff", () => {
  it("builds a finance AP handoff when final payment readiness created a supplier payable", () => {
    const handoff = buildFactoryFinalPaymentApHandoff(baseFinalPaymentResult);

    expect(handoff.status).toBe("ready");
    expect(handoff.payableId).toBe("ap-spm-s34-handoff");
    expect(handoff.payableNo).toBe("AP-SPM-S34-HANDOFF");
    expect(handoff.financeHref).toBe("/finance?ap_q=AP-SPM-S34-HANDOFF#supplier-payables");
    expect(handoff.message).toContain("AP-SPM-S34-HANDOFF");
  });

  it("keeps AP handoff pending when final payment readiness has no supplier payable evidence yet", () => {
    const handoff = buildFactoryFinalPaymentApHandoff({
      ...baseFinalPaymentResult,
      supplierPayable: undefined
    });

    expect(handoff.status).toBe("pending");
    expect(handoff.financeHref).toBeUndefined();
    expect(handoff.message).toContain("chưa có AP");
  });

  it("encodes payable search query for Finance deep links", () => {
    expect(buildFactoryFinalPaymentApHandoffHref("AP SPM/34 001")).toBe(
      "/finance?ap_q=AP+SPM%2F34+001#supplier-payables"
    );
  });

  it("builds a Finance fallback link from the subcontract order source after page reload", () => {
    const handoff = buildFactoryFinalPaymentApHandoffFromSource("SCO-S34-HANDOFF");

    expect(handoff.status).toBe("pending");
    expect(handoff.financeHref).toBe("/finance?ap_q=SCO-S34-HANDOFF#supplier-payables");
    expect(handoff.message).toContain("SCO-S34-HANDOFF");
  });
});

const baseFinalPaymentResult: SubcontractPaymentMilestoneResult = {
  order: {
    id: "sco-s34-handoff",
    orderNo: "SCO-S34-HANDOFF",
    factoryId: "factory-bd-002",
    factoryCode: "FACT-BD-002",
    factoryName: "Binh Duong Gia Cong",
    productId: "item-aah",
    sku: "AAH",
    productName: "Kem u phuc hoi AS A HABIT BIO 350GR",
    quantity: 999,
    uomCode: "PCS",
    acceptedQty: "999.000000",
    rejectedQty: "0.000000",
    specVersion: "S23SMK260504200049",
    sampleRequired: true,
    expectedDeliveryDate: "2026-05-20",
    depositStatus: "paid",
    depositAmount: 1000000,
    finalPaymentStatus: "pending",
    status: "final_payment_ready",
    createdBy: "Production Ops",
    createdAt: "2026-05-07T08:00:00Z",
    updatedAt: "2026-05-07T09:00:00Z",
    version: 12,
    estimatedCostAmount: "6800000.00",
    materialLines: [],
    auditLogIds: []
  },
  milestone: {
    id: "spm-s34-handoff",
    milestoneNo: "SPM-S34-HANDOFF",
    orderId: "sco-s34-handoff",
    orderNo: "SCO-S34-HANDOFF",
    factoryId: "factory-bd-002",
    factoryCode: "FACT-BD-002",
    factoryName: "Binh Duong Gia Cong",
    kind: "final_payment",
    status: "ready",
    amount: "6800000.00",
    currencyCode: "VND",
    readyBy: "finance-user",
    readyAt: "2026-05-07T09:00:00Z",
    createdAt: "2026-05-07T09:00:00Z",
    updatedAt: "2026-05-07T09:00:00Z",
    version: 1
  },
  supplierPayable: {
    payableId: "ap-spm-s34-handoff",
    payableNo: "AP-SPM-S34-HANDOFF",
    auditLogId: "audit-ap-spm-s34-handoff"
  },
  auditLog: {
    id: "audit-s34-final-payment",
    actorId: "finance-user",
    action: "subcontract.final_payment_ready",
    entityType: "subcontract_order",
    entityId: "sco-s34-handoff",
    metadata: {},
    createdAt: "2026-05-07T09:00:00Z"
  },
  auditLogId: "audit-s34-final-payment"
};
