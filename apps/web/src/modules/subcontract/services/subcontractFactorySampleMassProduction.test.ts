import { describe, expect, it } from "vitest";
import {
  buildFactorySampleDecisionInput,
  buildFactorySampleSubmissionInput,
  buildSubcontractFactorySampleMassProduction
} from "./subcontractFactorySampleMassProduction";
import type { SubcontractOrder, SubcontractSampleApproval } from "../types";

describe("subcontractFactorySampleMassProduction", () => {
  it("makes sample submission current after materials are issued", () => {
    const gate = buildSubcontractFactorySampleMassProduction({
      ...baseOrder,
      status: "materials_issued_to_factory",
      materialLines: fullyIssuedLines()
    });

    expect(gate).toMatchObject({
      sampleStatus: "ready_to_submit",
      canSubmitSample: true,
      canApproveSample: false,
      canRejectSample: false,
      massStatus: "pending",
      canStartMassProduction: false
    });
  });

  it("allows sample decision while a submitted sample waits for QA decision", () => {
    const gate = buildSubcontractFactorySampleMassProduction(
      {
        ...baseOrder,
        status: "sample_submitted",
        materialLines: fullyIssuedLines()
      },
      submittedSample
    );

    expect(gate).toMatchObject({
      sampleStatus: "submitted",
      canSubmitSample: false,
      canApproveSample: true,
      canRejectSample: true,
      latestSampleCode: "SCO-260505-0001-SAMPLE-A",
      massStatus: "pending",
      canStartMassProduction: false
    });
  });

  it("opens mass production after sample approval", () => {
    const gate = buildSubcontractFactorySampleMassProduction(
      {
        ...baseOrder,
        status: "sample_approved",
        materialLines: fullyIssuedLines()
      },
      {
        ...submittedSample,
        status: "approved",
        storageStatus: "retained_in_qa_cabinet"
      }
    );

    expect(gate).toMatchObject({
      sampleStatus: "approved",
      canSubmitSample: false,
      canApproveSample: false,
      canRejectSample: false,
      massStatus: "ready_to_start",
      canStartMassProduction: true
    });
  });

  it("lets no-sample orders start mass production after material handover", () => {
    const gate = buildSubcontractFactorySampleMassProduction({
      ...baseOrder,
      sampleRequired: false,
      status: "materials_issued_to_factory",
      materialLines: fullyIssuedLines()
    });

    expect(gate).toMatchObject({
      sampleStatus: "not_required",
      canSubmitSample: false,
      massStatus: "ready_to_start",
      canStartMassProduction: true
    });
  });

  it("builds sample submission payload with evidence", () => {
    const input = buildFactorySampleSubmissionInput({
      order: {
        ...baseOrder,
        status: "materials_issued_to_factory"
      },
      sampleCode: " SAMPLE-A ",
      formulaVersion: " FORMULA-2026.05 ",
      evidenceFileName: " sample-front.jpg ",
      note: " factory sent retained sample "
    });

    expect(input).toMatchObject({
      order: expect.objectContaining({ id: "sco-001" }),
      sampleCode: "SAMPLE-A",
      formulaVersion: "FORMULA-2026.05",
      specVersion: "S23SMK260504200049",
      submittedBy: "factory-user",
      note: "factory sent retained sample",
      evidence: [
        {
          evidenceType: "photo",
          fileName: "sample-front.jpg",
          objectKey: "subcontract-samples/sco-001/sample-front.jpg"
        }
      ]
    });
    expect(input.sampleApprovalId).toMatch(/^sample-sco-001-/);
  });

  it("builds sample decision payloads for approval and rejection", () => {
    const approval = buildFactorySampleDecisionInput({
      order: {
        ...baseOrder,
        status: "sample_submitted"
      },
      sampleApproval: submittedSample,
      decision: "approve",
      reason: " shade and fill level approved ",
      storageStatus: " retained_in_qa_cabinet "
    });
    const rejection = buildFactorySampleDecisionInput({
      order: {
        ...baseOrder,
        status: "sample_submitted"
      },
      decision: "reject",
      reason: " label color is wrong ",
      storageStatus: " retained_in_qa_cabinet "
    });

    expect(approval).toMatchObject({
      order: expect.objectContaining({ id: "sco-001" }),
      sampleApprovalId: "sample-a",
      reason: "shade and fill level approved",
      storageStatus: "retained_in_qa_cabinet"
    });
    expect(rejection).toMatchObject({
      sampleApprovalId: undefined,
      reason: "label color is wrong",
      storageStatus: undefined
    });
  });
});

function fullyIssuedLines() {
  return baseOrder.materialLines.map((line) => ({
    ...line,
    issuedQty: line.plannedQty
  }));
}

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
      plannedQty: "1.000000",
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

const submittedSample: SubcontractSampleApproval = {
  id: "sample-a",
  orderId: "sco-001",
  orderNo: "SCO-260505-0001",
  sampleCode: "SCO-260505-0001-SAMPLE-A",
  formulaVersion: "FORMULA-2026.05",
  specVersion: "S23SMK260504200049",
  status: "submitted",
  evidence: [
    {
      id: "sample-a-evidence-01",
      evidenceType: "photo",
      fileName: "sample-front.jpg",
      objectKey: "subcontract-samples/sco-001/sample-front.jpg",
      createdAt: "2026-05-06T08:00:00Z",
      createdBy: "factory-user"
    }
  ],
  submittedBy: "factory-user",
  submittedAt: "2026-05-06T08:00:00Z",
  createdAt: "2026-05-06T08:00:00Z",
  updatedAt: "2026-05-06T08:00:00Z",
  version: 1
};
