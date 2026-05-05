import { describe, expect, it } from "vitest";
import { buildSubcontractOrderFromProductionPlan } from "./productionPlanNextActions";
import type { ProductionPlan } from "../types";

describe("productionPlanNextActions", () => {
  it("builds a subcontract order from a material-ready production plan", () => {
    const input = buildSubcontractOrderFromProductionPlan(availablePlan, {
      factoryId: "sup-out-lotus",
      expectedDeliveryDate: "2026-05-20"
    });

    expect(input).toMatchObject({
      factoryId: "sup-out-lotus",
      productId: "item-xff-150",
      productName: "Tinh chat buoi Fast & Furious 150ML",
      quantity: 162,
      uomCode: "PCS",
      sampleRequired: true,
      expectedDeliveryDate: "2026-05-20",
      depositStatus: "pending",
      sourceProductionPlanId: "plan-002",
      sourceProductionPlanNo: "PP-260504-0002"
    });
    expect(input.specVersion).toContain("XFF-150ML");
    expect(input.specVersion).toContain("PP-260504-0002");
    expect(input.materialLines).toEqual([
      expect.objectContaining({
        itemId: "item-act-baicapil",
        skuCode: "ACT_BAICAPIL",
        itemName: "BAICAPIL",
        plannedQty: "0.162000",
        uomCode: "KG",
        unitCost: "0",
        lotTraceRequired: true
      }),
      expect.objectContaining({
        itemId: "item-pkg-bottle-150",
        skuCode: "CPGC-01",
        itemName: "Chai PET 150ML",
        plannedQty: "162.000000",
        uomCode: "PCS",
        unitCost: "0",
        lotTraceRequired: true
      })
    ]);
  });

  it("blocks subcontract order creation while any material shortage remains", () => {
    expect(() =>
      buildSubcontractOrderFromProductionPlan(shortagePlan, {
        factoryId: "sup-out-lotus",
        expectedDeliveryDate: "2026-05-20"
      })
    ).toThrow("Production plan still has material shortages");
  });

  it("blocks subcontract order creation until ready material has posted issue evidence", () => {
    const readyButNotIssuedPlan: ProductionPlan = {
      ...availablePlan,
      lines: availablePlan.lines.map((line) => ({
        ...line,
        issuedQty: "0.000000",
        remainingIssueQty: line.requiredStockBaseQty,
        issueStatus: "ready_to_issue",
        warehouseIssues: []
      }))
    };

    expect(() =>
      buildSubcontractOrderFromProductionPlan(readyButNotIssuedPlan, {
        factoryId: "sup-out-lotus",
        expectedDeliveryDate: "2026-05-20"
      })
    ).toThrow("Production plan materials must be issued before subcontract order creation");
  });
});

const basePlan = {
  id: "plan-001",
  orgId: "org-my-pham",
  planNo: "PP-260504-0001",
  outputItemId: "item-xff-150",
  outputSku: "XFF",
  outputItemName: "Tinh chat buoi Fast & Furious 150ML",
  outputItemType: "finished_good",
  plannedQty: "162.000000",
  uomCode: "PCS",
  formulaId: "formula-xff-v1",
  formulaCode: "XFF-150ML",
  formulaVersion: "v1",
  formulaBatchQty: "81.000000",
  formulaBatchUomCode: "PCS",
  plannedStartDate: "2026-05-10",
  plannedEndDate: "2026-05-12",
  status: "purchase_request_draft_created",
  auditLogId: "audit-production-plan-001",
  createdAt: "2026-05-04T03:00:00Z",
  createdBy: "user-production",
  updatedAt: "2026-05-04T03:00:00Z",
  updatedBy: "user-production",
  version: 1
} satisfies Omit<ProductionPlan, "lines" | "purchaseRequestDraft">;

const shortagePlan: ProductionPlan = {
  ...basePlan,
  lines: [
    {
      id: "pp-line-plan-001-001",
      formulaLineId: "formula-line-001",
      lineNo: 1,
      componentItemId: "item-act-baicapil",
      componentSku: "ACT_BAICAPIL",
      componentName: "BAICAPIL",
      componentType: "raw_material",
      formulaQty: "1.000000",
      formulaUomCode: "G",
      requiredQty: "162.000000",
      requiredUomCode: "G",
      requiredStockBaseQty: "0.162000",
      stockBaseUomCode: "KG",
      availableQty: "0.000500",
      shortageQty: "0.161500",
      purchaseDraftQty: "0.161500",
      purchaseDraftUomCode: "KG",
      issuedQty: "0.000000",
      remainingIssueQty: "0.162000",
      issueStatus: "shortage",
      warehouseIssues: [],
      isStockManaged: true,
      needsPurchase: true
    }
  ],
  purchaseRequestDraft: {
    id: "pr-draft-001",
    requestNo: "PR-DRAFT-260504-0001",
    sourceProductionPlanId: "plan-001",
    sourceProductionPlanNo: "PP-260504-0001",
    status: "draft",
    lines: [
      {
        id: "pr-line-001",
        lineNo: 1,
        sourceProductionPlanLineId: "pp-line-plan-001-001",
        itemId: "item-act-baicapil",
        sku: "ACT_BAICAPIL",
        itemName: "BAICAPIL",
        requestedQty: "0.161500",
        uomCode: "KG"
      }
    ]
  }
};

const availablePlan: ProductionPlan = {
  ...basePlan,
  id: "plan-002",
  planNo: "PP-260504-0002",
  status: "draft",
  lines: [
    {
      id: "pp-line-plan-002-001",
      formulaLineId: "formula-line-001",
      lineNo: 1,
      componentItemId: "item-act-baicapil",
      componentSku: "ACT_BAICAPIL",
      componentName: "BAICAPIL",
      componentType: "raw_material",
      formulaQty: "1.000000",
      formulaUomCode: "G",
      requiredQty: "162.000000",
      requiredUomCode: "G",
      requiredStockBaseQty: "0.162000",
      stockBaseUomCode: "KG",
      availableQty: "1.000000",
      shortageQty: "0.000000",
      purchaseDraftQty: "0.000000",
      purchaseDraftUomCode: "KG",
      issuedQty: "0.162000",
      remainingIssueQty: "0.000000",
      issueStatus: "issued",
      warehouseIssues: [
        {
          id: "issue-001",
          issueNo: "WI-260505-0001",
          lineId: "wi-line-001",
          status: "posted",
          quantity: "0.162000"
        }
      ],
      isStockManaged: true,
      needsPurchase: false
    },
    {
      id: "pp-line-plan-002-002",
      formulaLineId: "formula-line-002",
      lineNo: 2,
      componentItemId: "item-pkg-bottle-150",
      componentSku: "CPGC-01",
      componentName: "Chai PET 150ML",
      componentType: "packaging",
      formulaQty: "81.000000",
      formulaUomCode: "PCS",
      requiredQty: "162.000000",
      requiredUomCode: "PCS",
      requiredStockBaseQty: "162.000000",
      stockBaseUomCode: "PCS",
      availableQty: "1000.000000",
      shortageQty: "0.000000",
      purchaseDraftQty: "0.000000",
      purchaseDraftUomCode: "PCS",
      issuedQty: "162.000000",
      remainingIssueQty: "0.000000",
      issueStatus: "issued",
      warehouseIssues: [
        {
          id: "issue-002",
          issueNo: "WI-260505-0001",
          lineId: "wi-line-002",
          status: "posted",
          quantity: "162.000000"
        }
      ],
      isStockManaged: true,
      needsPurchase: false
    }
  ],
  purchaseRequestDraft: { lines: [] }
};
