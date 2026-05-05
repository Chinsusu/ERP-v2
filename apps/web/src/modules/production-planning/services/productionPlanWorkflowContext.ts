import { formatProductionPlanQuantity } from "./productionPlanService";
import type { ProductionPlan } from "../types";

export const productionPlanWorkflowSteps = [
  {
    number: 1,
    label: "Chọn kế hoạch sản xuất",
    description: "Chọn kế hoạch đang xử lý trước khi tính vật tư và tạo chứng từ."
  },
  {
    number: 2,
    label: "Tính nhu cầu vật tư",
    description: "Hệ thống tính nguyên liệu, bao bì cần dùng theo công thức."
  },
  {
    number: 3,
    label: "Tạo PO",
    description: "Tạo đơn mua từ các dòng vật tư còn thiếu."
  },
  {
    number: 4,
    label: "Tạo lệnh gia công",
    description: "Chuyển kế hoạch đủ vật tư sang lệnh sản xuất/gia công."
  }
] as const;

export type ProductionPlanWorkflowContext = {
  planLabel: string;
  outputLabel: string;
  quantityLabel: string;
  formulaLabel: string;
  shortageLineCount: number;
  purchaseLineCount: number;
  materialStatusLabel: string;
  materialStatusTone: "success" | "warning";
  purchaseTitle: string;
  purchaseSummary: string;
  purchaseButtonLabel: string;
  subcontractTitle: string;
  subcontractSummary: string;
  subcontractButtonLabel: string;
};

export function buildProductionPlanWorkflowContext(plan: ProductionPlan): ProductionPlanWorkflowContext {
  const quantityLabel = formatProductionPlanQuantity(plan.plannedQty, plan.uomCode);
  const shortageLineCount = plan.lines.filter((line) => line.needsPurchase || Number(line.shortageQty) > 0).length;
  const purchaseLineCount = plan.purchaseRequestDraft.lines.length;

  return {
    planLabel: `${plan.planNo} - ${plan.outputSku} - ${quantityLabel}`,
    outputLabel: `${plan.outputSku} - ${plan.outputItemName}`,
    quantityLabel,
    formulaLabel: `${plan.formulaCode} - ${plan.formulaVersion}`,
    shortageLineCount,
    purchaseLineCount,
    materialStatusLabel: shortageLineCount > 0 ? `Thiếu ${shortageLineCount} dòng vật tư` : "Đủ vật tư",
    materialStatusTone: shortageLineCount > 0 ? "warning" : "success",
    purchaseTitle: `Tạo PO từ ${plan.planNo}`,
    purchaseSummary:
      purchaseLineCount > 0
        ? `${purchaseLineCount} dòng vật tư cần mua cho ${plan.outputSku} - ${quantityLabel}.`
        : "Kế hoạch này không có dòng đề nghị mua.",
    purchaseButtonLabel: "Tạo PO từ kế hoạch này",
    subcontractTitle: `Tạo lệnh gia công từ ${plan.planNo}`,
    subcontractSummary:
      shortageLineCount === 0
        ? `Đủ vật tư để tạo lệnh gia công từ ${plan.planNo}.`
        : `Còn ${shortageLineCount} dòng thiếu vật tư, cần xử lý mua hàng trước.`,
    subcontractButtonLabel: "Tạo lệnh gia công từ kế hoạch này"
  };
}
