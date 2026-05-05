import { describe, expect, it } from "vitest";
import { buildProductionPlanWorklist } from "./productionPlanWorklist";
import type { ProductionPlan } from "../types";

describe("productionPlanWorklist", () => {
  it("builds a plan-level task list with material shortage and purchase tracking", () => {
    const tasks = buildProductionPlanWorklist(shortagePlan);

    expect(tasks.map((task) => task.title)).toEqual([
      "Kế hoạch sản xuất",
      "Nhu cầu vật tư",
      "Đề nghị mua",
      "Duyệt đề nghị mua",
      "PO vật tư",
      "Nhập kho/QC vật tư",
      "Lệnh gia công"
    ]);
    expect(tasks[1]).toMatchObject({
      statusLabel: "Thiếu 1 dòng vật tư",
      statusTone: "warning",
      detail: "1/1 dòng vật tư cần mua thêm."
    });
    expect(tasks[2]).toMatchObject({
      statusLabel: "Đề nghị nháp",
      statusTone: "warning",
      detail: "PR-DRAFT-260504-0001 có 1 dòng vật tư thiếu từ PP-260504-0001.",
      action: { label: "Mở đề nghị", href: "/purchase/requests/pr-draft-001", disabled: false }
    });
    expect(tasks[3]).toMatchObject({
      statusLabel: "Chờ gửi duyệt",
      statusTone: "warning",
      action: { label: "Mở đề nghị", href: "/purchase/requests/pr-draft-001", disabled: false }
    });
    expect(tasks[4]).toMatchObject({
      statusLabel: "Chờ duyệt đề nghị",
      statusTone: "info",
      action: { label: "Mở đề nghị", href: "/purchase/requests/pr-draft-001", disabled: false }
    });
    expect(tasks[6]).toMatchObject({
      statusLabel: "Chờ đủ vật tư",
      statusTone: "warning",
      action: { label: "Chờ bước 6", disabled: true }
    });
  });

  it("marks purchase and material-gate tasks complete when the plan has no shortages", () => {
    const tasks = buildProductionPlanWorklist(availablePlan);

    expect(tasks[1]).toMatchObject({
      statusLabel: "Đủ vật tư",
      statusTone: "success",
      detail: "1 dòng vật tư đã đủ tồn khả dụng."
    });
    expect(tasks[2]).toMatchObject({
      statusLabel: "Không cần đề nghị mua",
      statusTone: "success"
    });
    expect(tasks[6]).toMatchObject({
      statusLabel: "Sẵn sàng tạo lệnh",
      statusTone: "success",
      action: { label: "Mở gia công", href: "/subcontract", disabled: false }
    });
  });
});

const basePlan = {
  id: "plan-001",
  orgId: "org-my-pham",
  planNo: "PP-260504-0001",
  outputItemId: "item-aah",
  outputSku: "AAH",
  outputItemName: "Kem u phuc hoi AS A HABIT BIO 350GR",
  outputItemType: "finished_good",
  plannedQty: "999.000000",
  uomCode: "PCS",
  formulaId: "formula-aah-v1",
  formulaCode: "S23SMK260504200049",
  formulaVersion: "v260504200049",
  formulaBatchQty: "1.000000",
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
      componentItemId: "item-aci-bha",
      componentSku: "ACI_BHA",
      componentName: "ACID SALICYLIC",
      componentType: "raw_material",
      formulaQty: "0.100000",
      formulaUomCode: "G",
      requiredQty: "99.900000",
      requiredUomCode: "G",
      requiredStockBaseQty: "0.099900",
      stockBaseUomCode: "KG",
      availableQty: "0.000000",
      shortageQty: "0.099900",
      purchaseDraftQty: "0.099900",
      purchaseDraftUomCode: "KG",
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
        itemId: "item-aci-bha",
        sku: "ACI_BHA",
        itemName: "ACID SALICYLIC",
        requestedQty: "0.099900",
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
      ...shortagePlan.lines[0],
      id: "pp-line-plan-002-001",
      availableQty: "1.000000",
      shortageQty: "0.000000",
      purchaseDraftQty: "0.000000",
      needsPurchase: false
    }
  ],
  purchaseRequestDraft: { lines: [] }
};
